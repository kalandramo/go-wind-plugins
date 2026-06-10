package transport

import "github.com/tx7do/go-wind-plugins/broker"

// SubscribeOption holds deferred subscription parameters.
type SubscribeOption struct {
	Handler          broker.Handler
	Binder           broker.Binder
	SubscribeOptions []broker.SubscribeOption
}

// SubscribeOptionMap maps topic names to their SubscribeOption.
type SubscribeOptionMap map[string]*SubscribeOption
