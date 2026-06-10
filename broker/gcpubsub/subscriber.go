package gcpubsub

import (
	"context"
	"sync"

	"github.com/tx7do/go-wind-plugins/broker"
)

type subscriber struct {
	sync.RWMutex

	topic   string
	options broker.SubscribeOptions

	b      *gcpBroker
	cancel context.CancelFunc
	closed bool
}

func (s *subscriber) Options() broker.SubscribeOptions {
	s.RLock()
	defer s.RUnlock()
	return s.options
}

func (s *subscriber) Topic() string {
	s.RLock()
	defer s.RUnlock()
	return s.topic
}

func (s *subscriber) Unsubscribe(removeFromManager bool) error {
	s.Lock()
	defer s.Unlock()

	if s.cancel != nil {
		s.cancel()
	}

	s.closed = true

	if s.b != nil && s.b.subscribers != nil && removeFromManager {
		_ = s.b.subscribers.RemoveOnly(s.topic)
	}

	return nil
}
