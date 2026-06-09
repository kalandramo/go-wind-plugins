package fs

import (
	"context"
	"embed"
	"testing"

	baseConfig "github.com/tx7do/go-wind-plugins/config"
)

//go:embed testdata/*
var testFS embed.FS

func TestNew_NilFS(t *testing.T) {
	_, err := New()
	if err == nil {
		t.Fatal("expected error for nil fsys")
	}
}

func TestLoad(t *testing.T) {
	s, err := New(WithFS(testFS), WithPath("testdata/config.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	data, err := s.Load(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	expected := "name: test\nport: 8080\n"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, data)
	}
}

func TestLoad_KeyOverride(t *testing.T) {
	s, err := New(WithFS(testFS), WithPath("testdata/config.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	data, err := s.Load(context.Background(), "testdata/other.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "key: value\n" {
		t.Errorf("unexpected content: %q", data)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	s, err := New(WithFS(testFS), WithPath("testdata/config.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Load(context.Background(), "nonexistent.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_EmptyPath(t *testing.T) {
	s, err := New(WithFS(testFS))
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Load(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty path with no default")
	}
}

func TestInterfaceCompliance(t *testing.T) {
	var _ baseConfig.Reader = (*source)(nil)
}
