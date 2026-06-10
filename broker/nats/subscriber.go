package nats

import (
	"sync"

	natsGo "github.com/nats-io/nats.go"
	"github.com/tx7do/go-wind-plugins/broker"
)

// subscriberRemover is an interface for removing subscribers from the broker's map.
type subscriberRemover interface {
	removeSubscriber(topic string) bool
}

// Ensure both brokers implement subscriberRemover
var _ subscriberRemover = (*natsBroker)(nil)
var _ subscriberRemover = (*jetStreamBroker)(nil)

func (b *natsBroker) removeSubscriber(topic string) bool {
	if b.subscribers != nil {
		return b.subscribers.RemoveOnly(topic)
	}
	return false
}

func (b *jetStreamBroker) removeSubscriber(topic string) bool {
	if b.subscribers != nil {
		return b.subscribers.RemoveOnly(topic)
	}
	return false
}

type subscriber struct {
	sync.RWMutex

	remover subscriberRemover
	s       *natsGo.Subscription
	options broker.SubscribeOptions
	closed  bool
}

func (s *subscriber) Options() broker.SubscribeOptions {
	s.RLock()
	defer s.RUnlock()

	return s.options
}

func (s *subscriber) Topic() string {
	s.RLock()
	defer s.RUnlock()

	if s.s == nil {
		return ""
	}

	return s.s.Subject
}

func (s *subscriber) Unsubscribe(removeFromManager bool) error {
	s.Lock()
	defer s.Unlock()

	s.closed = true

	var err error
	if s.s != nil {
		err = s.s.Unsubscribe()

		if s.remover != nil && removeFromManager {
			_ = s.remover.removeSubscriber(s.s.Subject)
		}
	}

	return err
}

func (s *subscriber) IsClosed() bool {
	s.RLock()
	defer s.RUnlock()

	return s.closed
}
