package kcp

import (
	"context"
	"errors"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/tx7do/go-wind-plugins/encoding"
	"github.com/tx7do/go-wind-plugins/metrics"

	"go.opentelemetry.io/otel/trace"

	"github.com/xtaci/kcp-go/v5"
)

type ClientMessageHandler func(NetMessagePayload) error

type ClientRawMessageHandler func([]byte) error

type ClientHandlerData struct {
	Handler ClientMessageHandler
	Creator Creator
}
type ClientMessageHandlerMap map[NetMessageType]ClientHandlerData

type Client struct {
	conn *kcp.UDPSession

	url      string
	endpoint *url.URL

	codec             encoding.Codec
	messageHandlers   ClientMessageHandlerMap
	rawMessageHandler ClientRawMessageHandler

	tracer trace.Tracer
	m      metrics.Metrics

	timeout time.Duration

	blockCryptPassword, blockCryptSalt string
	dataShards, parityShards           int
}

func NewClient(opts ...ClientOption) *Client {
	cli := &Client{
		url:             "",
		timeout:         1 * time.Second,
		codec:           encoding.GetCodec("json"),
		messageHandlers: make(ClientMessageHandlerMap),
		dataShards:      10,
		parityShards:    3,
	}

	cli.init(opts...)

	return cli
}

func (c *Client) init(opts ...ClientOption) {
	for _, o := range opts {
		o(c)
	}

	addr := c.url

	prefix := "udp://"
	if !strings.HasPrefix(addr, "udp://") {
		prefix = "udp://"
	}
	addr = prefix + addr

	c.endpoint, _ = url.Parse(addr)
}

func (c *Client) Connect() error {
	if c.endpoint == nil {
		return errors.New("endpoint is nil")
	}

	log.Printf("[kcp] connecting to %s", c.endpoint.String())

	block := NewBlockCryptFromPassword(c.blockCryptPassword, c.blockCryptSalt)
	conn, err := kcp.DialWithOptions(c.url, block, c.dataShards, c.parityShards)
	if err != nil {
		log.Printf("[kcp] cant connect to server: %s", err)
		return err
	}

	c.conn = conn

	go c.run()

	return nil
}

func (c *Client) Disconnect() {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Printf("[kcp] disconnect error: %s", err)
		}
		c.conn = nil
	}
}

func (c *Client) RegisterMessageHandler(messageType NetMessageType, handler ClientMessageHandler, binder Creator) {
	if _, ok := c.messageHandlers[messageType]; ok {
		return
	}

	c.messageHandlers[messageType] = ClientHandlerData{handler, binder}
}

func RegisterClientMessageHandler[T any](cli *Client, messageType NetMessageType, handler func(*T) error) {
	cli.RegisterMessageHandler(messageType,
		func(payload NetMessagePayload) error {
			switch t := payload.(type) {
			case *T:
				return handler(t)
			default:
				log.Printf("[kcp] invalid payload struct type: %T", t)
				return errors.New("invalid payload struct type")
			}
		},
		func() any {
			var t T
			return &t
		},
	)
}

func (c *Client) DeregisterMessageHandler(messageType NetMessageType) {
	delete(c.messageHandlers, messageType)
}

func (c *Client) SendRawData(message []byte) error {
	if c.conn == nil {
		return errors.New("client is not connected")
	}

	startTime := time.Now()
	labels := map[string]string{
		"rpc.system": "kcp",
	}

	if c.m != nil {
		c.m.Counter(context.Background(), "kcp.client.messages.sent", 1, labels)
	}

	if _, err := c.conn.Write(message); err != nil {
		if c.m != nil {
			c.m.Counter(context.Background(), "kcp.client.messages.errors", 1, map[string]string{
				"rpc.system": "kcp",
				"error":      "true",
			})
		}
		return err
	}

	if c.m != nil {
		c.m.Histogram(context.Background(), "kcp.client.message.send_duration", time.Since(startTime).Seconds(), labels)
	}

	return nil
}

func (c *Client) SendMessage(messageType int, message any) error {
	var msg NetPacket
	msg.Type = NetMessageType(messageType)
	var err error
	msg.Payload, err = c.codec.Marshal(message)
	if err != nil {
		return err
	}

	var buff []byte
	if buff, err = msg.Marshal(); err != nil {
		return err
	}

	return c.SendRawData(buff)
}

func (c *Client) run() {
	defer c.Disconnect()

	buf := make([]byte, 102400)

	for {
		readLen, err := c.conn.Read(buf)
		if err != nil {
			log.Printf("[kcp] read message error: %v", err)
			return
		}

		if c.rawMessageHandler != nil {
			if err := c.rawMessageHandler(buf[:readLen]); err != nil {
				log.Printf("[kcp] raw data handler exception: %s", err)
				continue
			}
			continue
		}

		if err = c.messageHandler(buf[:readLen]); err != nil {
			log.Printf("[kcp] process message error: %v", err)
		}

		if c.m != nil {
			c.m.Counter(context.Background(), "kcp.client.messages.received", 1, map[string]string{
				"rpc.system": "kcp",
			})
		}
	}
}

func (c *Client) messageHandler(buf []byte) error {
	var msg NetPacket
	if err := msg.Unmarshal(buf); err != nil {
		log.Printf("[kcp] decode message exception: %s", err)
		return err
	}

	handlerData, ok := c.messageHandlers[msg.Type]
	if !ok {
		log.Printf("[kcp] message type not found: %d", msg.Type)
		return errors.New("message handler not found")
	}

	var payload NetMessagePayload

	if handlerData.Creator != nil {
		payload = handlerData.Creator()

		if err := c.codec.Unmarshal(msg.Payload, payload); err != nil {
			log.Printf("[kcp] unmarshal message exception: %s", err)
			return err
		}
	} else {
		payload = msg.Payload
	}

	if err := handlerData.Handler(payload); err != nil {
		log.Printf("[kcp] message handler exception: %s", err)
		return err
	}

	return nil
}
