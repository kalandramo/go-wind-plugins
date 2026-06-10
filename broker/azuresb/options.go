package azuresb

import (
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/tx7do/go-wind-plugins/broker"
)

///
/// Option
///

type connectionStringKey struct{}

// WithConnectionString sets the Azure Service Bus connection string.
// Example: "Endpoint=sb://<namespace>.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=<key>"
func WithConnectionString(connStr string) broker.Option {
	return broker.OptionContextWithValue(connectionStringKey{}, connStr)
}

///
/// PublishOption
///

type publishContentTypeKey struct{}
type publishSessionIDKey struct{}
type publishMessageIDKey struct{}

// WithPublishContentType sets the content type for the message.
func WithPublishContentType(contentType string) broker.PublishOption {
	return broker.PublishContextWithValue(publishContentTypeKey{}, contentType)
}

// WithPublishSessionID sets the session ID for the message (requires session-enabled queue/topic).
func WithPublishSessionID(sessionID string) broker.PublishOption {
	return broker.PublishContextWithValue(publishSessionIDKey{}, sessionID)
}

// WithPublishMessageID sets a custom message ID.
func WithPublishMessageID(messageID string) broker.PublishOption {
	return broker.PublishContextWithValue(publishMessageIDKey{}, messageID)
}

///
/// SubscribeOption
///

type subscriptionNameKey struct{}
type receiveModeKey struct{}

// WithSubscriptionName sets the subscription name for topic subscriptions.
func WithSubscriptionName(name string) broker.SubscribeOption {
	return broker.SubscribeContextWithValue(subscriptionNameKey{}, name)
}

// WithReceiveMode sets the receive mode (PeekLock or ReceiveAndDelete).
func WithReceiveMode(mode azservicebus.ReceiveMode) broker.SubscribeOption {
	return broker.SubscribeContextWithValue(receiveModeKey{}, mode)
}
