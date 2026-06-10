package redis

import (
	"github.com/tx7do/go-wind-plugins/broker"

	"github.com/tx7do/go-wind-plugins/broker/redis/option"
	"github.com/tx7do/go-wind-plugins/broker/redis/pubsub"
	"github.com/tx7do/go-wind-plugins/broker/redis/stream"
)

func NewBroker(driverType option.DriverType, opts ...broker.Option) broker.Broker {
	switch driverType {
	case option.DriverTypeStream:
		return stream.NewBroker(opts...)
	case option.DriverTypePubSub:
		return pubsub.NewBroker(opts...)
	default:
		return pubsub.NewBroker(opts...)
	}
}
