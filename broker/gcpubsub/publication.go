package gcpubsub

import (
	"fmt"

	"cloud.google.com/go/pubsub/v2"

	"github.com/tx7do/go-wind-plugins/broker"
)

type publication struct {
	topic  string
	msg    *broker.Message
	gcpMsg *pubsub.Message
	err    error
	ack    func()
	acked  bool
}

func (p *publication) Topic() string {
	return p.topic
}

func (p *publication) Message() *broker.Message {
	return p.msg
}

func (p *publication) RawMessage() any {
	return p.gcpMsg
}

func (p *publication) Ack() error {
	if p.acked {
		return nil
	}
	if p.ack == nil {
		p.err = fmt.Errorf("ack function is nil")
		return p.err
	}
	p.ack()
	p.acked = true
	return nil
}

func (p *publication) Error() error {
	return p.err
}
