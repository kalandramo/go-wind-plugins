package nats

import (
	natsGo "github.com/nats-io/nats.go"

	"github.com/tx7do/go-wind-plugins/broker"
)

type publication struct {
	t   string
	err error
	m   *broker.Message
}

func (p *publication) Topic() string {
	return p.t
}

func (p *publication) Message() *broker.Message {
	return p.m
}

func (p *publication) RawMessage() any {
	return p.m
}

// Ack acknowledges the message.
// For JetStream messages (identified by having a Reply subject), this calls msg.Ack().
// For core NATS messages, this is a no-op since core NATS does not support acknowledgments.
func (p *publication) Ack() error {
	if p.m != nil {
		if msg, ok := p.m.Msg.(*natsGo.Msg); ok {
			// Only Ack JetStream messages (they have a reply subject for acking)
			if msg.Reply != "" {
				return msg.Ack()
			}
		}
	}
	return nil
}

func (p *publication) Error() error {
	return p.err
}
