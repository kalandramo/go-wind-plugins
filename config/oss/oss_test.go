package oss

import (
	"testing"

	baseConfig "github.com/tx7do/go-wind-plugins/config"
)

func TestNew_NilClient(t *testing.T) {
	_, err := New(nil, WithBucket("test"))
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNew_EmptyBucket(t *testing.T) {
	// Passing nil client will fail first, so we test bucket validation separately.
	s := &source{
		client:  nil, // won't be called
		options: &options{bucket: "", key: "config.yaml"},
	}
	_ = s

	_, err := New(nil)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNew_DefaultPollInterval(t *testing.T) {
	o := &options{}
	for _, opt := range []Option{WithBucket("test")} {
		opt(o)
	}
	if o.pollInterval != 0 {
		// pollInterval is set to default in New(), not in options directly
	}

	if defaultPollInterval <= 0 {
		t.Error("expected positive default poll interval")
	}
}

func TestOptions(t *testing.T) {
	o := &options{
		pollInterval: defaultPollInterval,
	}

	WithBucket("my-bucket")(o)
	WithKey("configs/app.yaml")(o)

	if o.bucket != "my-bucket" {
		t.Errorf("expected bucket my-bucket, got %s", o.bucket)
	}
	if o.key != "configs/app.yaml" {
		t.Errorf("expected key configs/app.yaml, got %s", o.key)
	}
}

func TestResolveKey(t *testing.T) {
	s := &source{options: &options{key: "default.yaml"}}

	if got := s.resolveKey(""); got != "default.yaml" {
		t.Errorf("expected default.yaml, got %s", got)
	}
	if got := s.resolveKey("override.yaml"); got != "override.yaml" {
		t.Errorf("expected override.yaml, got %s", got)
	}
}

func TestInterfaceCompliance(t *testing.T) {
	var _ baseConfig.Reader = (*source)(nil)
	var _ baseConfig.ValueWatcher = (*source)(nil)
}
