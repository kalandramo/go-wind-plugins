package thrift

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
)

// ---------------------------------------------------------------------------
// NewServer + defaults
// ---------------------------------------------------------------------------

func TestNewServer_Defaults(t *testing.T) {
	srv := NewServer(":0", WithProcessor(thrift.NewTMultiplexedProcessor()))
	if srv.Addr() != ":0" {
		t.Errorf("Addr() = %q, want %q", srv.Addr(), ":0")
	}
	if srv.protocol != "binary" {
		t.Errorf("protocol = %q, want %q", srv.protocol, "binary")
	}
}

// ---------------------------------------------------------------------------
// Endpoint
// ---------------------------------------------------------------------------

func TestServer_Endpoint(t *testing.T) {
	srv := NewServer(":7700", WithProcessor(thrift.NewTMultiplexedProcessor()))
	ep := srv.Endpoint()
	if !strings.HasPrefix(ep, KindThrift+"://") {
		t.Errorf("expected %s:// prefix, got %q", KindThrift, ep)
	}
}

func TestServer_Endpoint_NormalizeWildcard(t *testing.T) {
	srv := NewServer("0.0.0.0:7700", WithProcessor(thrift.NewTMultiplexedProcessor()))
	ep := srv.Endpoint()
	if !strings.Contains(ep, "localhost") {
		t.Errorf("expected localhost in endpoint, got %q", ep)
	}
}

// ---------------------------------------------------------------------------
// Start without processor (should return ErrNoProcessor)
// ---------------------------------------------------------------------------

func TestServer_StartNoProcessor(t *testing.T) {
	srv := NewServer(":0")
	err := srv.Start(context.Background())
	if err != ErrNoProcessor {
		t.Errorf("expected ErrNoProcessor, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Factory functions
// ---------------------------------------------------------------------------

func TestCreateProtocolFactory(t *testing.T) {
	tests := []struct {
		protocol string
		wantNil  bool
	}{
		{"binary", false},
		{"compact", false},
		{"json", false},
		{"simplejson", false},
		{"", false}, // default to binary
		{"invalid", true},
	}
	for _, tc := range tests {
		f := createProtocolFactory(tc.protocol)
		if (f == nil) != tc.wantNil {
			t.Errorf("createProtocolFactory(%q) nil=%v, want %v", tc.protocol, f == nil, tc.wantNil)
		}
	}
}

func TestCreateTransportFactory(t *testing.T) {
	cfg := &thrift.TConfiguration{}

	// Plain
	f := createTransportFactory(cfg, false, false, 0)
	if f == nil {
		t.Error("expected non-nil transport factory (plain)")
	}

	// Buffered
	f = createTransportFactory(cfg, true, false, 1024)
	if f == nil {
		t.Error("expected non-nil transport factory (buffered)")
	}

	// Framed
	f = createTransportFactory(cfg, false, true, 0)
	if f == nil {
		t.Error("expected non-nil transport factory (framed)")
	}

	// Buffered + Framed
	f = createTransportFactory(cfg, true, true, 1024)
	if f == nil {
		t.Error("expected non-nil transport factory (buffered+framed)")
	}
}

// ---------------------------------------------------------------------------
// Start + Stop lifecycle
// ---------------------------------------------------------------------------

func TestServer_StartStop(t *testing.T) {
	srv := NewServer("127.0.0.1:0", WithProcessor(thrift.NewTMultiplexedProcessor()))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start(ctx)
	}()

	time.Sleep(200 * time.Millisecond)

	if srv.server == nil {
		t.Fatal("expected server to be initialized after Start")
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Start returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server did not stop within 5s")
	}
}

// ---------------------------------------------------------------------------
// Stop before Start (should not panic)
// ---------------------------------------------------------------------------

func TestServer_StopBeforeStart(t *testing.T) {
	srv := NewServer(":0", WithProcessor(thrift.NewTMultiplexedProcessor()))
	if err := srv.Stop(context.Background()); err != nil {
		t.Errorf("Stop before Start returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// WithProtocol option
// ---------------------------------------------------------------------------

func TestWithProtocol(t *testing.T) {
	srv := NewServer(":0", WithProcessor(thrift.NewTMultiplexedProcessor()), WithProtocol("compact"))
	if srv.protocol != "compact" {
		t.Errorf("protocol = %q, want %q", srv.protocol, "compact")
	}
}

// ---------------------------------------------------------------------------
// WithTransportConfig option
// ---------------------------------------------------------------------------

func TestWithTransportConfig(t *testing.T) {
	srv := NewServer(":0",
		WithProcessor(thrift.NewTMultiplexedProcessor()),
		WithTransportConfig(true, true, 4096),
	)
	if !srv.buffered {
		t.Error("expected buffered=true")
	}
	if !srv.framed {
		t.Error("expected framed=true")
	}
	if srv.bufferSize != 4096 {
		t.Errorf("bufferSize = %d, want 4096", srv.bufferSize)
	}
}
