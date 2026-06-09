package conductor

import (
	"testing"

	"github.com/conductor-sdk/conductor-go/sdk/model"
)

func TestNewClient(t *testing.T) {
	// Test that NewClient returns an error-free client with default URL
	client, err := NewClient(ClientOptions{})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil client")
	}
	if client.options.ServerURL != defaultServerURL {
		t.Errorf("expected default server URL %s, got %s", defaultServerURL, client.options.ServerURL)
	}
	_ = client.Close()
}

func TestNewClientWithCustomURL(t *testing.T) {
	customURL := "http://custom-server:8080/api"
	client, err := NewClient(ClientOptions{
		ServerURL: customURL,
	})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil client")
	}
	if client.options.ServerURL != customURL {
		t.Errorf("expected server URL %s, got %s", customURL, client.options.ServerURL)
	}
	_ = client.Close()
}

func TestNewClientWithAuth(t *testing.T) {
	client, err := NewClient(ClientOptions{
		ServerURL:  "http://localhost:8080/api",
		AuthKey:    "test-key",
		AuthSecret: "test-secret",
	})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil client")
	}
	_ = client.Close()
}

func TestToStartWorkflowRequest(t *testing.T) {
	version := int32(1)
	priority := int32(10)

	opts := StartWorkflowOptions{
		Name:          "test_workflow",
		Version:       &version,
		Input:         map[string]interface{}{"key": "value"},
		CorrelationID: "corr-123",
		Priority:      &priority,
	}

	req := toStartWorkflowRequest(opts)

	if req.Name != "test_workflow" {
		t.Errorf("expected name test_workflow, got %s", req.Name)
	}
	if req.Version != 1 {
		t.Errorf("expected version 1, got %d", req.Version)
	}
	if req.CorrelationId != "corr-123" {
		t.Errorf("expected correlationId corr-123, got %s", req.CorrelationId)
	}
	if req.Priority != 10 {
		t.Errorf("expected priority 10, got %d", req.Priority)
	}
}

func TestToStartWorkflowRequestDefaults(t *testing.T) {
	opts := StartWorkflowOptions{
		Name:  "test_workflow",
		Input: map[string]interface{}{"key": "value"},
	}

	req := toStartWorkflowRequest(opts)

	if req.Name != "test_workflow" {
		t.Errorf("expected name test_workflow, got %s", req.Name)
	}
	if req.Version != 0 {
		t.Errorf("expected version 0 (unset), got %d", req.Version)
	}
	if req.CorrelationId != "" {
		t.Errorf("expected empty correlationId, got %s", req.CorrelationId)
	}
}

func TestTaskWorkerLifecycle(t *testing.T) {
	client, err := NewClient(ClientOptions{})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer func() { _ = client.Close() }()

	// StartWorker will fail without a real server, but we test the config
	worker, err := client.StartWorker("test_task", func(task *model.Task) (interface{}, error) {
		return nil, nil
	}, 1, defaultPollInterval)
	if err != nil {
		t.Fatalf("StartWorker returned error: %v", err)
	}
	if worker == nil {
		t.Fatal("StartWorker returned nil worker")
	}
	if worker.TaskType() != "test_task" {
		t.Errorf("expected task type test_task, got %s", worker.TaskType())
	}
	if !worker.IsRunning() {
		t.Error("expected worker to be running")
	}
	worker.Stop()
	if worker.IsRunning() {
		t.Error("expected worker to be stopped after Stop()")
	}
}

func TestWorkerConfigDefaults(t *testing.T) {
	config := WorkerConfig{
		TaskType: "my_task",
	}

	client, err := NewClient(ClientOptions{})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer func() { _ = client.Close() }()

	worker, err := client.StartWorkerWithConfig(config, func(task *model.Task) (interface{}, error) {
		return nil, nil
	})
	if err != nil {
		t.Fatalf("StartWorkerWithConfig returned error: %v", err)
	}
	if worker == nil {
		t.Fatal("StartWorkerWithConfig returned nil worker")
	}
	if worker.config.Concurrency != defaultConcurrency {
		t.Errorf("expected default concurrency %d, got %d", defaultConcurrency, worker.config.Concurrency)
	}
	if worker.config.PollInterval != defaultPollInterval {
		t.Errorf("expected default poll interval %v, got %v", defaultPollInterval, worker.config.PollInterval)
	}
}
