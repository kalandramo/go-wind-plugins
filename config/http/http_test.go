package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	baseConfig "github.com/tx7do/go-wind-plugins/config"
)

func TestNew_EmptyURL(t *testing.T) {
	_, err := New()
	if err == nil {
		t.Fatal("expected error for empty url")
	}
}

func TestLoad(t *testing.T) {
	body := `{"key":"value"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	defer srv.Close()

	src, err := New(WithURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()

	data, err := src.Load(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != body {
		t.Errorf("expected %q, got %q", body, data)
	}
}

func TestLoad_KeyOverride(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/a", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("aaa"))
	})
	mux.HandleFunc("/b", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("bbb"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	src, err := New(WithURL(srv.URL + "/a"))
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()

	data, _ := src.Load(context.Background(), "")
	if string(data) != "aaa" {
		t.Errorf("expected aaa, got %s", data)
	}

	data, _ = src.Load(context.Background(), srv.URL+"/b")
	if string(data) != "bbb" {
		t.Errorf("expected bbb, got %s", data)
	}
}

func TestLoad_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	src, err := New(WithURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()

	_, err = src.Load(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for 500")
	}
}

func TestLoad_WithHeaders(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	src, err := New(WithURL(srv.URL), WithHeader("Authorization", "Bearer token123"))
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()

	_, _ = src.Load(context.Background(), "")
	if gotAuth != "Bearer token123" {
		t.Errorf("expected auth header 'Bearer token123', got %q", gotAuth)
	}
}

func TestWatchValue_ETag(t *testing.T) {
	var counter int32
	currentBody := atomic.Value{}
	currentBody.Store("v1")
	currentETag := atomic.Value{}
	currentETag.Store("etag-v1")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&counter, 1)
		ifNoneMatch := r.Header.Get("If-None-Match")
		etag := currentETag.Load().(string)

		if ifNoneMatch == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", etag)
		w.Write([]byte(currentBody.Load().(string)))
	}))
	defer srv.Close()

	src, err := New(WithURL(srv.URL), WithPollInterval(100*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ch, err := src.WatchValue(ctx, "")
	if err != nil {
		t.Fatal(err)
	}

	// Should receive initial value.
	select {
	case data := <-ch:
		if string(data) != "v1" {
			t.Errorf("expected v1, got %q", data)
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for initial value")
	}

	// Update the value and ETag.
	currentBody.Store("v2")
	currentETag.Store("etag-v2")

	select {
	case data := <-ch:
		if string(data) != "v2" {
			t.Errorf("expected v2, got %q", data)
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for updated value")
	}
}

func TestClose(t *testing.T) {
	src := &source{
		options: &options{url: "http://localhost"},
		client:  &http.Client{},
	}
	if err := src.Close(); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestInterfaceCompliance(t *testing.T) {
	var _ baseConfig.Reader = (*source)(nil)
	var _ baseConfig.ValueWatcher = (*source)(nil)
	var _ baseConfig.Closer = (*source)(nil)
}
