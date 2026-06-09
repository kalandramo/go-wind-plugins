package goworkflows

import (
	"context"
	"testing"
	"time"

	"github.com/cschleiden/go-workflows/worker"
)

// mockBackend is a minimal backend.Backend implementation for testing.
// We only test the wrapper logic, not actual workflow execution.

func TestNewClient_NilBackend(t *testing.T) {
	_, err := NewClient(nil)
	if err == nil {
		t.Fatal("expected error for nil backend")
	}
}

func TestNewWorker_NilBackend(t *testing.T) {
	_, err := NewWorker(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil backend")
	}
}

func TestNewWorkflowOnlyWorker_NilBackend(t *testing.T) {
	_, err := NewWorkflowOnlyWorker(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil backend")
	}
}

func TestNewActivityOnlyWorker_NilBackend(t *testing.T) {
	_, err := NewActivityOnlyWorker(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil backend")
	}
}

func TestWorkerOptions_Defaults(t *testing.T) {
	var opts *WorkerOptions
	result := opts.toWorkerOptions()
	defaults := worker.DefaultOptions

	if result.WorkflowPollers != defaults.WorkflowPollers {
		t.Errorf("expected WorkflowPollers %d, got %d", defaults.WorkflowPollers, result.WorkflowPollers)
	}
	if result.ActivityPollers != defaults.ActivityPollers {
		t.Errorf("expected ActivityPollers %d, got %d", defaults.ActivityPollers, result.ActivityPollers)
	}
}

func TestWorkerOptions_Custom(t *testing.T) {
	opts := &WorkerOptions{
		WorkflowPollers:           4,
		MaxParallelWorkflowTasks:  10,
		WorkflowHeartbeatInterval: 30 * time.Second,
		ActivityPollers:           3,
		SingleWorkerMode:          true,
	}

	result := opts.toWorkerOptions()

	if result.WorkflowPollers != 4 {
		t.Errorf("expected WorkflowPollers 4, got %d", result.WorkflowPollers)
	}
	if result.MaxParallelWorkflowTasks != 10 {
		t.Errorf("expected MaxParallelWorkflowTasks 10, got %d", result.MaxParallelWorkflowTasks)
	}
	if result.WorkflowHeartbeatInterval != 30*time.Second {
		t.Errorf("expected WorkflowHeartbeatInterval 30s, got %v", result.WorkflowHeartbeatInterval)
	}
	if result.ActivityPollers != 3 {
		t.Errorf("expected ActivityPollers 3, got %d", result.ActivityPollers)
	}
	if !result.SingleWorkerMode {
		t.Error("expected SingleWorkerMode to be true")
	}
}

func TestCreateWorkflowOptions_Validation(t *testing.T) {
	wc := &WorkflowClient{}

	_, err := wc.CreateWorkflowInstance(context.Background(), CreateWorkflowOptions{}, nil)
	if err == nil {
		t.Fatal("expected error for empty InstanceID")
	}
}

func TestWorkerIsRunning_Initially(t *testing.T) {
	ww := &WorkflowWorker{}
	if ww.IsRunning() {
		t.Error("expected worker to not be running initially")
	}
}

func TestWorkerStop_WhenNotRunning(t *testing.T) {
	ww := &WorkflowWorker{}
	ww.Stop() // Should not panic
}

func TestWorkerLifecycle(t *testing.T) {
	ww := &WorkflowWorker{}

	// Should not be running
	if ww.IsRunning() {
		t.Error("expected worker to not be running")
	}

	// Stop when not running should be safe
	ww.Stop()

	if ww.IsRunning() {
		t.Error("expected worker to not be running after stop")
	}
}
