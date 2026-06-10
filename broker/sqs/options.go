package sqs

import (
	"time"

	"github.com/tx7do/go-wind-plugins/broker"
)

///
/// Option
///

type regionKey struct{}
type endpointKey struct{}
type queueUrlKey struct{}

// WithRegion sets the AWS region.
func WithRegion(region string) broker.Option {
	return broker.OptionContextWithValue(regionKey{}, region)
}

// WithEndpoint sets a custom endpoint URL (for local testing with ElasticMQ/LocalStack).
func WithEndpoint(endpoint string) broker.Option {
	return broker.OptionContextWithValue(endpointKey{}, endpoint)
}

// WithQueueUrl sets the default queue URL.
func WithQueueUrl(url string) broker.Option {
	return broker.OptionContextWithValue(queueUrlKey{}, url)
}

///
/// PublishOption
///

type delaySecondsKey struct{}
type messageGroupIdKey struct{}
type messageDeduplicationIdKey struct{}

// WithDelaySeconds sets the delay for message delivery (0-900 seconds).
func WithDelaySeconds(seconds int32) broker.PublishOption {
	return broker.PublishContextWithValue(delaySecondsKey{}, seconds)
}

// WithMessageGroupId sets the MessageGroupId for FIFO queues.
func WithMessageGroupId(groupId string) broker.PublishOption {
	return broker.PublishContextWithValue(messageGroupIdKey{}, groupId)
}

// WithMessageDeduplicationId sets the MessageDeduplicationId for FIFO queues.
func WithMessageDeduplicationId(dedupId string) broker.PublishOption {
	return broker.PublishContextWithValue(messageDeduplicationIdKey{}, dedupId)
}

///
/// SubscribeOption
///

type visibilityTimeoutKey struct{}
type waitTimeSecondsKey struct{}
type maxMessagesKey struct{}

const (
	DefaultVisibilityTimeout = 30 // seconds
	DefaultWaitTimeSeconds   = 20 // seconds (long polling)
	DefaultMaxMessages       = 10 // messages per receive
)

// WithVisibilityTimeout sets the visibility timeout for messages (seconds).
func WithVisibilityTimeout(seconds int32) broker.SubscribeOption {
	return broker.SubscribeContextWithValue(visibilityTimeoutKey{}, seconds)
}

// WithWaitTimeSeconds sets the long polling wait time (seconds).
func WithWaitTimeSeconds(seconds int32) broker.SubscribeOption {
	return broker.SubscribeContextWithValue(waitTimeSecondsKey{}, seconds)
}

// WithMaxMessages sets the maximum number of messages to retrieve per poll.
func WithMaxMessages(n int32) broker.SubscribeOption {
	return broker.SubscribeContextWithValue(maxMessagesKey{}, n)
}

///
/// SubscribeOption (internal)
///

type pollIntervalKey struct{}

// WithPollInterval sets the interval between polls when no messages are received.
func WithPollInterval(d time.Duration) broker.SubscribeOption {
	return broker.SubscribeContextWithValue(pollIntervalKey{}, d)
}

///
/// helpers
///
