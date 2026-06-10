package gcpubsub

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"cloud.google.com/go/pubsub/v2"
	"google.golang.org/api/option"

	"github.com/tx7do/go-wind-plugins/broker"
)

type gcpBroker struct {
	sync.RWMutex

	options broker.Options

	client *pubsub.Client

	running bool

	subscribers *broker.SubscriberSyncMap
}

func NewBroker(opts ...broker.Option) broker.Broker {
	options := broker.NewOptionsAndApply(opts...)

	b := &gcpBroker{
		options:     options,
		subscribers: broker.NewSubscriberSyncMap(),
	}

	return b
}

func (b *gcpBroker) Name() string {
	return "gcpubsub"
}

func (b *gcpBroker) Options() broker.Options {
	return b.options
}

func (b *gcpBroker) Address() string {
	if len(b.options.Addrs) > 0 {
		return b.options.Addrs[0]
	}
	return ""
}

func (b *gcpBroker) Init(opts ...broker.Option) error {
	for _, o := range opts {
		o(&b.options)
	}
	return nil
}

func (b *gcpBroker) Connect() error {
	b.Lock()
	defer b.Unlock()

	if b.running {
		return nil
	}

	ctx := context.Background()

	projectID := ""
	if b.options.Context != nil {
		if v, ok := b.options.Context.Value(projectIDKey{}).(string); ok && v != "" {
			projectID = v
		}
	}

	if projectID == "" {
		return errors.New("gcp project id is required, use WithProjectID() to set it")
	}

	var clientOpts []option.ClientOption

	if b.options.Context != nil {
		if v, ok := b.options.Context.Value(credentialsFileKey{}).(string); ok && v != "" {
			clientOpts = append(clientOpts, option.WithCredentialsFile(v))
		}
		if v, ok := b.options.Context.Value(endpointKey{}).(string); ok && v != "" {
			clientOpts = append(clientOpts, option.WithEndpoint(v))
		}
	}

	client, err := pubsub.NewClient(ctx, projectID, clientOpts...)
	if err != nil {
		return fmt.Errorf("create pubsub client error: %w", err)
	}

	b.client = client
	b.running = true

	LogInfof("connected to GCP Pub/Sub, project: %s", projectID)

	return nil
}

func (b *gcpBroker) Disconnect() error {
	b.Lock()
	defer b.Unlock()

	if !b.running {
		return nil
	}

	b.subscribers.Clear()

	if b.client != nil {
		_ = b.client.Close()
	}

	b.client = nil
	b.running = false

	LogInfo("disconnected from GCP Pub/Sub")

	return nil
}

func (b *gcpBroker) Request(ctx context.Context, topic string, msg *broker.Message, opts ...broker.RequestOption) (*broker.Message, error) {
	return nil, errors.New("not implemented")
}

func (b *gcpBroker) Publish(ctx context.Context, topic string, msg *broker.Message, opts ...broker.PublishOption) error {
	var finalTask broker.PublishHandler = b.internalPublish

	if len(b.options.PublishMiddlewares) > 0 {
		finalTask = broker.ChainPublishMiddleware(finalTask, b.options.PublishMiddlewares)
	}

	return finalTask(ctx, topic, msg, opts...)
}

func (b *gcpBroker) internalPublish(ctx context.Context, topic string, msg *broker.Message, opts ...broker.PublishOption) error {
	buf, err := broker.Marshal(b.options.Codec, msg.Body)
	if err != nil {
		return err
	}

	sendMsg := msg.Clone()
	sendMsg.Body = buf

	return b.publish(ctx, topic, sendMsg, opts...)
}

func (b *gcpBroker) publish(ctx context.Context, topic string, msg *broker.Message, opts ...broker.PublishOption) error {
	b.RLock()
	client := b.client
	b.RUnlock()

	if client == nil {
		return errors.New("GCP Pub/Sub client is nil")
	}

	options := broker.PublishOptions{
		Context: ctx,
	}
	for _, o := range opts {
		o(&options)
	}

	t := client.Publisher(topic)

	pubsubMsg := &pubsub.Message{
		Data: msg.BodyBytes(),
	}

	if msg.Headers != nil {
		attrs := make(map[string]string, len(msg.Headers))
		for k, v := range msg.Headers {
			attrs[k] = v
		}
		pubsubMsg.Attributes = attrs
	}

	if options.Context != nil {
		if v, ok := options.Context.Value(publishOrderingKey{}).(string); ok && v != "" {
			pubsubMsg.OrderingKey = v
		}
	}

	result := t.Publish(ctx, pubsubMsg)

	_, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("publish message error: %w", err)
	}

	return nil
}

func (b *gcpBroker) Subscribe(topic string, handler broker.Handler, binder broker.Binder, opts ...broker.SubscribeOption) (broker.Subscriber, error) {
	b.RLock()
	client := b.client
	b.RUnlock()

	if client == nil {
		return nil, errors.New("GCP Pub/Sub client is nil, call Connect() first")
	}

	options := broker.SubscribeOptions{
		Context: context.Background(),
		AutoAck: true,
	}
	for _, o := range opts {
		o(&options)
	}

	if len(b.options.SubscriberMiddlewares) > 0 {
		handler = broker.ChainSubscriberMiddleware(handler, b.options.SubscriberMiddlewares)
	}

	// Resolve subscription name: subscribe context → topic name as default
	subscriptionName := topic
	if options.Context != nil {
		if v, ok := options.Context.Value(subscriptionNameKey{}).(string); ok && v != "" {
			subscriptionName = v
		}
	}

	var receiveSettings pubsub.ReceiveSettings
	if options.Context != nil {
		if v, ok := options.Context.Value(receiveSettingsKey{}).(pubsub.ReceiveSettings); ok {
			receiveSettings = v
		}
	}

	sub := &subscriber{
		topic:   topic,
		options: options,
		b:       b,
	}

	go b.receive(options.Context, client, subscriptionName, receiveSettings, handler, binder, options, sub)

	b.subscribers.Add(topic, sub)

	return sub, nil
}

func (b *gcpBroker) receive(ctx context.Context, client *pubsub.Client, subscriptionName string,
	receiveSettings pubsub.ReceiveSettings, handler broker.Handler, binder broker.Binder,
	options broker.SubscribeOptions, sub *subscriber) {

	subClient := client.Subscriber(subscriptionName)
	subClient.ReceiveSettings = receiveSettings

	subCtx, cancel := context.WithCancel(ctx)
	sub.Lock()
	sub.cancel = cancel
	sub.Unlock()

	defer func() {
		LogInfof("subscriber stopped, topic: %s, subscription: %s", sub.topic, subscriptionName)
	}()

	err := subClient.Receive(subCtx, func(receiveCtx context.Context, msg *pubsub.Message) {
		var m broker.Message

		// Extract headers from message attributes
		if msg.Attributes != nil {
			m.Headers = make(broker.Headers)
			for k, v := range msg.Attributes {
				m.Headers[k] = v
			}
		}

		// Extract body
		if len(msg.Data) > 0 {
			if binder != nil {
				m.Body = binder()
				if err := broker.Unmarshal(b.options.Codec, msg.Data, &m.Body); err != nil {
					LogErrorf("unmarshal message failed: %v", err)
					msg.Nack()
					return
				}
			} else {
				m.Body = msg.Data
			}
		}

		p := &publication{
			topic:  sub.topic,
			msg:    &m,
			gcpMsg: msg,
			ack:    msg.Ack,
		}

		if err := handler(receiveCtx, p); err != nil {
			p.err = err
			LogErrorf("handle message failed: %v", err)
			return
		}

		if options.AutoAck {
			if err := p.Ack(); err != nil {
				LogErrorf("unable to ack msg: %v", err)
			}
		}
	})

	if err != nil {
		if subCtx.Err() != nil {
			// context cancelled, normal exit
			return
		}
		LogErrorf("receive message error: %v", err)
	}
}
