package gcpubsub

import (
	"time"

	"cloud.google.com/go/pubsub/v2"
	"github.com/tx7do/go-wind-plugins/broker"
)

///
/// Option
///

type projectIDKey struct{}
type credentialsFileKey struct{}
type endpointKey struct{}

// WithProjectID sets the GCP project ID.
func WithProjectID(projectID string) broker.Option {
	return broker.OptionContextWithValue(projectIDKey{}, projectID)
}

// WithCredentialsFile sets the path to a service account credentials JSON file.
func WithCredentialsFile(path string) broker.Option {
	return broker.OptionContextWithValue(credentialsFileKey{}, path)
}

// WithEndpoint sets a custom endpoint (for testing with emulators).
func WithEndpoint(endpoint string) broker.Option {
	return broker.OptionContextWithValue(endpointKey{}, endpoint)
}

///
/// PublishOption
///

type publishTimeoutKey struct{}
type publishOrderingKey struct{}

// WithPublishTimeout sets the timeout for publishing a single message.
func WithPublishTimeout(d time.Duration) broker.PublishOption {
	return broker.PublishContextWithValue(publishTimeoutKey{}, d)
}

// WithPublishOrderingKey sets the ordering key for message ordering.
func WithPublishOrderingKey(key string) broker.PublishOption {
	return broker.PublishContextWithValue(publishOrderingKey{}, key)
}

///
/// SubscribeOption
///

type subscriptionNameKey struct{}
type receiveSettingsKey struct{}

// WithSubscriptionName sets the Pub/Sub subscription name.
func WithSubscriptionName(name string) broker.SubscribeOption {
	return broker.SubscribeContextWithValue(subscriptionNameKey{}, name)
}

// WithReceiveSettings sets the receive settings for the subscriber.
func WithReceiveSettings(settings pubsub.ReceiveSettings) broker.SubscribeOption {
	return broker.SubscribeContextWithValue(receiveSettingsKey{}, settings)
}
