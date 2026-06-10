// Package main demonstrates a TCP server using go-wind.
//
// This example shows:
//   - Registering typed message handlers with JSON deserialization
//   - Echoing messages back to the sender
//   - Broadcasting messages to all connected clients
//   - Connection/disconnection callbacks
//
// The TCP wire format is [4-byte type (little-endian)][payload bytes].
//
// Run:
//
//	go run ./_examples/tcp-basic
//
// Test with netcat:
//
//	# Connect (binary mode)
//	nc localhost 9000
//
//	# Or use a script to send a properly framed message:
//	#   4-byte type (LE) + JSON payload
//	#   type=1: echo  type=2: chat
package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/tx7do/go-wind-plugins/encoding/json" // side-effect: register JSON codec
	tcpServer "github.com/tx7do/go-wind-plugins/transport/tcp"
)

// Message types.
const (
	MsgTypeEcho tcpServer.NetMessageType = 1 // echo back to sender
	MsgTypeChat tcpServer.NetMessageType = 2 // broadcast to all
)

// chatMessage is the JSON payload for chat messages.
type chatMessage struct {
	Text     string `json:"text"`
	Username string `json:"username,omitempty"`
}

func main() {
	var srv *tcpServer.Server
	srv = tcpServer.NewServer(
		tcpServer.WithAddress(":9000"),
		tcpServer.WithCodec("json"),
		tcpServer.WithSocketConnectHandler(func(sid tcpServer.SessionID, connected bool) {
			if connected {
				log.Printf("[connect] %s connected (%d total)", sid, srv.SessionCount())
			} else {
				log.Printf("[disconnect] %s disconnected (%d total)", sid, srv.SessionCount())
			}
		}),
	)

	// Echo handler: deserializes chatMessage, logs it, sends it back.
	tcpServer.RegisterServerMessageHandler[chatMessage](srv, MsgTypeEcho,
		func(sessionId tcpServer.SessionID, msg *chatMessage) error {
			log.Printf("[echo] %s: %s", sessionId, msg.Text)
			return srv.SendMessage(sessionId, MsgTypeEcho, &chatMessage{
				Text:     "echo: " + msg.Text,
				Username: "server",
			})
		},
	)

	// Chat handler: broadcasts to all connected clients.
	tcpServer.RegisterServerMessageHandler[chatMessage](srv, MsgTypeChat,
		func(sessionId tcpServer.SessionID, msg *chatMessage) error {
			log.Printf("[chat] %s (%s): %s", sessionId, msg.Username, msg.Text)
			srv.Broadcast(MsgTypeChat, &chatMessage{
				Text:     msg.Text,
				Username: msg.Username,
			})
			return nil
		},
	)

	// Graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Printf("TCP server listening on %s\n", srv.Endpoint())
	fmt.Println()
	fmt.Println("Wire format: [4-byte type (LE)] [JSON payload]")
	fmt.Println()
	fmt.Println("Test with a script:")
	fmt.Printf("  type=%d (echo)  → {\"text\":\"hello\"}\n", MsgTypeEcho)
	fmt.Printf("  type=%d (chat)  → {\"text\":\"hi\",\"username\":\"alice\"}\n", MsgTypeChat)
	fmt.Println()

	// Print the raw bytes for type=1 echo "hello":
	helloPayload := []byte(`{"text":"hello"}`)
	buf := make([]byte, 4+len(helloPayload))
	binary.LittleEndian.PutUint32(buf[:4], uint32(MsgTypeEcho))
	copy(buf[4:], helloPayload)
	fmt.Printf("Raw bytes for echo: %x\n", buf)
	fmt.Println()

	if err := srv.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("server stopped")
}
