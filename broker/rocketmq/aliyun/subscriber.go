package aliyun

import (
	"sync"

	aliyun "github.com/aliyunmq/mq-http-go-sdk"

	"github.com/tx7do/go-wind-plugins/broker"
)

type Subscriber struct {
	sync.RWMutex
	r       *aliyunmqBroker
	topic   string
	options broker.SubscribeOptions
	handler broker.Handler
	binder  broker.Binder
	reader  aliyun.MQConsumer
	closed  bool
	done    chan struct{}
}

func (s *Subscriber) Options() broker.SubscribeOptions {
	return s.options
}

func (s *Subscriber) Topic() string {
	return s.topic
}

func (s *Subscriber) Unsubscribe(removeFromManager bool) error {
	s.Lock()
	defer s.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true

	// Signal doConsume goroutine to stop
	close(s.done)

	return nil
}
