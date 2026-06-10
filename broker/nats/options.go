package nats

import (
	"time"

	natsGo "github.com/nats-io/nats.go"

	"github.com/tx7do/go-wind-plugins/broker"
)

///////////////////////////////////////////////////////////////////////////////
/// Core NATS Broker Options
///////////////////////////////////////////////////////////////////////////////

type optionsKey struct{}
type drainConnectionKey struct{}

func Options(opts natsGo.Options) broker.Option {
	return broker.OptionContextWithValue(optionsKey{}, opts)
}

func DrainConnection() broker.Option {
	return broker.OptionContextWithValue(drainConnectionKey{}, struct{}{})
}

///////////////////////////////////////////////////////////////////////////////
/// Core NATS Publish Options
///////////////////////////////////////////////////////////////////////////////

type headersKey struct{}

func WithHeaders(h map[string][]string) broker.PublishOption {
	return broker.PublishContextWithValue(headersKey{}, h)
}

///////////////////////////////////////////////////////////////////////////////
/// Core NATS Request Options
///////////////////////////////////////////////////////////////////////////////

type requestTimeoutKey struct{}

func WithRequestTimeout(timeout time.Duration) broker.RequestOption {
	return broker.RequestContextWithValue(requestTimeoutKey{}, timeout)
}

func WithRequestHeaders(h map[string][]string) broker.RequestOption {
	return broker.RequestContextWithValue(headersKey{}, h)
}

///////////////////////////////////////////////////////////////////////////////
/// JetStream Broker Options
///////////////////////////////////////////////////////////////////////////////

type jetStreamContextOptsKey struct{}

// JetStreamContextOptions sets NATS JetStream context options.
func JetStreamContextOptions(opts ...natsGo.JSOpt) broker.Option {
	return broker.OptionContextWithValue(jetStreamContextOptsKey{}, opts)
}

///////////////////////////////////////////////////////////////////////////////
/// JetStream Publish Options
///////////////////////////////////////////////////////////////////////////////

type pubMsgIdKey struct{}
type pubExpectStreamKey struct{}
type pubExpectLastSequenceKey struct{}
type pubExpectLastSequencePerSubjectKey struct{}
type pubExpectLastMsgIdKey struct{}
type pubRawOptsKey struct{}

// WithMsgId sets a message ID for deduplication.
func WithMsgId(id string) broker.PublishOption {
	return broker.PublishContextWithValue(pubMsgIdKey{}, id)
}

// WithExpectStream sets the expected stream for publish.
func WithExpectStream(stream string) broker.PublishOption {
	return broker.PublishContextWithValue(pubExpectStreamKey{}, stream)
}

// WithExpectLastSequence sets the expected last sequence for publish.
func WithExpectLastSequence(seq uint64) broker.PublishOption {
	return broker.PublishContextWithValue(pubExpectLastSequenceKey{}, seq)
}

// WithExpectLastSequencePerSubject sets the expected last sequence per subject for publish.
func WithExpectLastSequencePerSubject(seq uint64) broker.PublishOption {
	return broker.PublishContextWithValue(pubExpectLastSequencePerSubjectKey{}, seq)
}

// WithExpectLastMsgId sets the expected last message ID for publish.
func WithExpectLastMsgId(id string) broker.PublishOption {
	return broker.PublishContextWithValue(pubExpectLastMsgIdKey{}, id)
}

// WithPublishRawOpts allows passing raw NATS PubOpt options directly.
func WithPublishRawOpts(opts ...natsGo.PubOpt) broker.PublishOption {
	return broker.PublishContextWithValue(pubRawOptsKey{}, opts)
}

///////////////////////////////////////////////////////////////////////////////
/// JetStream Subscribe Options
///////////////////////////////////////////////////////////////////////////////

type subDurableKey struct{}
type subDeliverAllKey struct{}
type subDeliverLastKey struct{}
type subDeliverNewKey struct{}
type subStartSequenceKey struct{}
type subStartTimeKey struct{}
type subAckWaitKey struct{}
type subMaxAckPendingKey struct{}
type subBindStreamKey struct{}
type subReplayInstantKey struct{}
type subDescriptionKey struct{}
type subManualAckKey struct{}
type subPullKey struct{}
type subPullBatchSizeKey struct{}
type subRawOptsKey struct{}

// WithDurable sets the durable consumer name for the subscription.
func WithDurable(name string) broker.SubscribeOption {
	return broker.SubscribeContextWithValue(subDurableKey{}, name)
}

// WithDeliverAll delivers all messages from the beginning of the stream.
func WithDeliverAll() broker.SubscribeOption {
	return broker.SubscribeContextWithValue(subDeliverAllKey{}, true)
}

// WithDeliverLast delivers only the last message from the stream.
func WithDeliverLast() broker.SubscribeOption {
	return broker.SubscribeContextWithValue(subDeliverLastKey{}, true)
}

// WithDeliverNew delivers only new messages.
func WithDeliverNew() broker.SubscribeOption {
	return broker.SubscribeContextWithValue(subDeliverNewKey{}, true)
}

// WithStartSequence sets the starting sequence for the subscription.
func WithStartSequence(seq uint64) broker.SubscribeOption {
	return broker.SubscribeContextWithValue(subStartSequenceKey{}, seq)
}

// WithStartTime sets the starting time for the subscription.
func WithStartTime(t time.Time) broker.SubscribeOption {
	return broker.SubscribeContextWithValue(subStartTimeKey{}, t)
}

// WithSubscribeAckWait sets how long to wait for an ACK before redelivering.
func WithSubscribeAckWait(d time.Duration) broker.SubscribeOption {
	return broker.SubscribeContextWithValue(subAckWaitKey{}, d)
}

// WithSubscribeMaxAckPending sets the maximum number of unacknowledged messages.
func WithSubscribeMaxAckPending(n int) broker.SubscribeOption {
	return broker.SubscribeContextWithValue(subMaxAckPendingKey{}, n)
}

// WithBindStream binds the subscription to a specific stream.
func WithBindStream(stream string) broker.SubscribeOption {
	return broker.SubscribeContextWithValue(subBindStreamKey{}, stream)
}

// WithReplayInstant replays messages as fast as possible.
func WithReplayInstant() broker.SubscribeOption {
	return broker.SubscribeContextWithValue(subReplayInstantKey{}, true)
}

// WithConsumerDescription sets the consumer description.
func WithConsumerDescription(desc string) broker.SubscribeOption {
	return broker.SubscribeContextWithValue(subDescriptionKey{}, desc)
}

// WithManualAck disables auto-acknowledgment; the handler must call Ack() manually.
func WithManualAck() broker.SubscribeOption {
	return broker.SubscribeContextWithValue(subManualAckKey{}, true)
}

// WithPullSubscribe enables pull-based subscription instead of push.
func WithPullSubscribe() broker.SubscribeOption {
	return broker.SubscribeContextWithValue(subPullKey{}, true)
}

// WithPullBatchSize sets the batch size for pull subscriptions (default: 10).
func WithPullBatchSize(n int) broker.SubscribeOption {
	return broker.SubscribeContextWithValue(subPullBatchSizeKey{}, n)
}

// WithSubscribeRawOpts allows passing raw NATS SubOpt options directly.
func WithSubscribeRawOpts(opts ...natsGo.SubOpt) broker.SubscribeOption {
	return broker.SubscribeContextWithValue(subRawOptsKey{}, opts)
}

///////////////////////////////////////////////////////////////////////////////
/// JetStream Utility Functions
///////////////////////////////////////////////////////////////////////////////

// GetJetStreamContext extracts the underlying NATS JetStream context from a broker.Broker.
// This allows advanced operations like stream management:
//
//	js := nats.GetJetStreamContext(b)
//	js.AddStream(&natsGo.StreamConfig{Name: "ORDERS", Subjects: []string{"orders.*"}})
func GetJetStreamContext(b broker.Broker) natsGo.JetStreamContext {
	type jsAccessor interface {
		JetStreamContext() natsGo.JetStreamContext
	}
	if jsa, ok := b.(jsAccessor); ok {
		return jsa.JetStreamContext()
	}
	return nil
}

// JetStreamMsgFromEvent extracts the underlying NATS message from a broker Event.
// This allows JetStream-specific operations like Nak, Term, InProgress:
//
//	msg, ok := nats.JetStreamMsgFromEvent(event)
//	if ok {
//	    _ = msg.Nak()
//	}
func JetStreamMsgFromEvent(evt broker.Event) (*natsGo.Msg, bool) {
	if evt == nil || evt.Message() == nil {
		return nil, false
	}
	if m, ok := evt.Message().Msg.(*natsGo.Msg); ok {
		return m, true
	}
	return nil, false
}

// GetConn extracts the underlying NATS connection from a broker.Broker.
func GetConn(b broker.Broker) *natsGo.Conn {
	type connAccessor interface {
		Conn() *natsGo.Conn
	}
	if ca, ok := b.(connAccessor); ok {
		return ca.Conn()
	}
	return nil
}
