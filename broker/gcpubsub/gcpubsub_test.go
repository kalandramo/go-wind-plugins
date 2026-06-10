package gcpubsub

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tx7do/go-wind-plugins/broker"
)

func TestNewBroker(t *testing.T) {
	b := NewBroker(
		WithProjectID("test-project"),
	)
	assert.NotNil(t, b)
	assert.Equal(t, "gcpubsub", b.Name())
}

func TestPublishNotConnected(t *testing.T) {
	b := NewBroker(
		WithProjectID("test-project"),
	)
	err := b.Publish(context.Background(), "test-topic", &broker.Message{})
	assert.Error(t, err)
}

func TestSubscribeNotConnected(t *testing.T) {
	b := NewBroker(
		WithProjectID("test-project"),
	)
	_, err := b.Subscribe("test-topic", func(ctx context.Context, event broker.Event) error {
		return nil
	}, nil)
	assert.Error(t, err)
}

func TestConnectWithoutProjectID(t *testing.T) {
	b := NewBroker()
	err := b.Connect()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project id")
}

func TestOptions(t *testing.T) {
	b := NewBroker(
		WithProjectID("my-project"),
		WithCredentialsFile("/path/to/key.json"),
		WithEndpoint("localhost:8085"),
	)
	assert.NotNil(t, b)
	opts := b.Options()
	assert.NotNil(t, opts)
}

func TestSubscribeOptions(t *testing.T) {
	_ = broker.SubscribeContextWithValue(subscriptionNameKey{}, "my-subscription")
}
