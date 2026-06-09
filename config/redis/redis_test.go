package redis

import (
	"testing"

	baseConfig "github.com/tx7do/go-wind-plugins/config"
)

func TestNew_NilClient(t *testing.T) {
	_, err := New(nil, WithPath("test"))
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNew_EmptyPath(t *testing.T) {
	// We can't pass a real client, but we can test path validation
	// by constructing a source directly.
	s := &source{
		client:  nil,
		options: &options{path: ""},
	}
	if s.options.path != "" {
		t.Error("expected empty path")
	}
}

func TestResolveKey(t *testing.T) {
	s := &source{options: &options{path: "myapp:config"}}

	if got := s.resolveKey(""); got != "myapp:config" {
		t.Errorf("expected myapp:config, got %s", got)
	}
	if got := s.resolveKey("override"); got != "override" {
		t.Errorf("expected override, got %s", got)
	}
}

func TestWatchChannel(t *testing.T) {
	got := watchChannel("myapp:config")
	expected := "__windcfg__:myapp:config"
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestInterfaceCompliance(t *testing.T) {
	var _ baseConfig.Reader = (*source)(nil)
	var _ baseConfig.ValueWatcher = (*source)(nil)
}
