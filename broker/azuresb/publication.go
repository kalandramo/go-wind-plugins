package azuresb

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"

	"github.com/tx7do/go-wind-plugins/broker"
)

type publication struct {
	topic    string
	msg      *broker.Message
	sbMsg    *azservicebus.ReceivedMessage
	receiver *azservicebus.Receiver
	err      error
	acked    bool
}

func (p *publication) Topic() string {
	return p.topic
}

func (p *publication) Message() *broker.Message {
	return p.msg
}

func (p *publication) RawMessage() any {
	return p.sbMsg
}

func (p *publication) Ack() error {
	if p.acked {
		return nil
	}
	if p.receiver == nil {
		p.err = fmt.Errorf("receiver is nil")
		return p.err
	}
	if err := p.receiver.CompleteMessage(context.Background(), p.sbMsg, nil); err != nil {
		p.err = err
		return err
	}
	p.acked = true
	return nil
}

func (p *publication) Error() error {
	return p.err
}
