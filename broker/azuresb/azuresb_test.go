package azuresb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tx7do/go-wind-plugins/broker"
)

func TestNewBroker(t *testing.T) {
	b := NewBroker(
		WithConnectionString("Endpoint=sb://test.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=testkey"),
	)
	assert.NotNil(t, b)
	assert.Equal(t, "azuresb", b.Name())
}

func TestPublishNotConnected(t *testing.T) {
	b := NewBroker(
		WithConnectionString("Endpoint=sb://test.servicebus.windows.net/"),
	)
	err := b.Publish(context.Background(), "test-queue", &broker.Message{})
	assert.Error(t, err)
}

func TestSubscribeNotConnected(t *testing.T) {
	b := NewBroker(
		WithConnectionString("Endpoint=sb://test.servicebus.windows.net/"),
	)
	_, err := b.Subscribe("test-queue", func(ctx context.Context, event broker.Event) error {
		return nil
	}, nil)
	assert.Error(t, err)
}

func TestConnectWithoutConnectionString(t *testing.T) {
	b := NewBroker()
	err := b.Connect()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection string")
}

func TestOptions(t *testing.T) {
	b := NewBroker(
		WithConnectionString("Endpoint=sb://test.servicebus.windows.net/;SharedAccessKeyName=test;SharedAccessKey=test"),
	)
	assert.NotNil(t, b)
	opts := b.Options()
	assert.NotNil(t, opts)
}

func TestPublishOptions(t *testing.T) {
	_ = broker.PublishContextWithValue(publishSessionIDKey{}, "session-1")
	_ = broker.PublishContextWithValue(publishContentTypeKey{}, "application/json")
	_ = broker.PublishContextWithValue(publishMessageIDKey{}, "msg-123")
}
