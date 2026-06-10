package sqs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	"github.com/tx7do/go-wind-plugins/broker"
)

type publication struct {
	topic  string
	msg    *broker.Message
	sqsMsg *types.Message
	err    error

	client   *sqs.Client
	queueUrl string
	acked    bool
}

func (p *publication) Topic() string {
	return p.topic
}

func (p *publication) Message() *broker.Message {
	return p.msg
}

func (p *publication) RawMessage() any {
	return p.sqsMsg
}

func (p *publication) Ack() error {
	if p.acked {
		return nil
	}
	if p.client == nil {
		p.err = fmt.Errorf("SQS client is nil")
		return p.err
	}
	if p.sqsMsg == nil || p.sqsMsg.ReceiptHandle == nil {
		p.err = fmt.Errorf("SQS message or receipt handle is nil")
		return p.err
	}

	_, err := p.client.DeleteMessage(context.Background(), &sqs.DeleteMessageInput{
		QueueUrl:      &p.queueUrl,
		ReceiptHandle: p.sqsMsg.ReceiptHandle,
	})
	if err != nil {
		p.err = fmt.Errorf("delete message failed: %w", err)
		return p.err
	}

	p.acked = true
	return nil
}

func (p *publication) Error() error {
	return p.err
}
