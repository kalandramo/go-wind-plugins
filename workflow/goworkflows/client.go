package goworkflows

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cschleiden/go-workflows/backend"
	"github.com/cschleiden/go-workflows/client"
	"github.com/cschleiden/go-workflows/core"
	"github.com/cschleiden/go-workflows/workflow"
)

// WorkflowClient provides high-level access to go-workflows workflow instance management.
type WorkflowClient struct {
	mu      sync.RWMutex
	backend backend.Backend
	client  *client.Client
}

// NewClient creates a new WorkflowClient with the given backend.
func NewClient(b backend.Backend) (*WorkflowClient, error) {
	if b == nil {
		return nil, fmt.Errorf("backend is nil")
	}

	c := client.New(b)

	LogInfo("created workflow client")

	return &WorkflowClient{
		backend: b,
		client:  c,
	}, nil
}

// Backend returns the underlying backend.
func (wc *WorkflowClient) Backend() backend.Backend {
	return wc.backend
}

// InnerClient returns the underlying go-workflows client for advanced operations.
func (wc *WorkflowClient) InnerClient() *client.Client {
	return wc.client
}

// Close closes the underlying backend resources.
func (wc *WorkflowClient) Close() error {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	if wc.backend != nil {
		LogInfo("closing workflow client and backend")
		return wc.backend.Close()
	}
	return nil
}

// CreateWorkflowInstance creates a new workflow instance.
// The workflow function and arguments are passed directly.
// Returns the workflow instance with its InstanceID and ExecutionID.
func (wc *WorkflowClient) CreateWorkflowInstance(ctx context.Context, opts CreateWorkflowOptions, wf workflow.Workflow, args ...any) (*workflow.Instance, error) {
	if wc.client == nil {
		return nil, fmt.Errorf("client is nil")
	}

	if opts.InstanceID == "" {
		return nil, fmt.Errorf("InstanceID must be set in CreateWorkflowOptions")
	}

	clientOpts := client.WorkflowInstanceOptions{
		InstanceID: opts.InstanceID,
		Queue:      opts.Queue,
	}

	instance, err := wc.client.CreateWorkflowInstance(ctx, clientOpts, wf, args...)
	if err != nil {
		return nil, fmt.Errorf("create workflow instance error: %w", err)
	}

	LogInfof("created workflow instance: %s (execution: %s)", instance.InstanceID, instance.ExecutionID)
	return instance, nil
}

// CancelWorkflowInstance cancels a running workflow instance.
func (wc *WorkflowClient) CancelWorkflowInstance(ctx context.Context, instance *workflow.Instance) error {
	if wc.client == nil {
		return fmt.Errorf("client is nil")
	}

	if err := wc.client.CancelWorkflowInstance(ctx, instance); err != nil {
		return fmt.Errorf("cancel workflow instance error: %w", err)
	}

	LogInfof("cancelled workflow instance: %s", instance.InstanceID)
	return nil
}

// SignalWorkflow sends a signal to a running workflow instance.
func (wc *WorkflowClient) SignalWorkflow(ctx context.Context, instanceID string, name string, arg any) error {
	if wc.client == nil {
		return fmt.Errorf("client is nil")
	}

	if err := wc.client.SignalWorkflow(ctx, instanceID, name, arg); err != nil {
		return fmt.Errorf("signal workflow error: %w", err)
	}

	LogInfof("signaled workflow instance %s with signal %s", instanceID, name)
	return nil
}

// GetWorkflowInstanceState returns the current state of the given workflow instance.
func (wc *WorkflowClient) GetWorkflowInstanceState(ctx context.Context, instance *workflow.Instance) (core.WorkflowInstanceState, error) {
	if wc.client == nil {
		return core.WorkflowInstanceStateActive, fmt.Errorf("client is nil")
	}

	state, err := wc.client.GetWorkflowInstanceState(ctx, instance)
	if err != nil {
		return core.WorkflowInstanceStateActive, fmt.Errorf("get workflow state error: %w", err)
	}

	return state, nil
}

// WaitForWorkflowInstance waits for the given workflow instance to finish.
// Uses the default timeout of 20 seconds if timeout is 0.
func (wc *WorkflowClient) WaitForWorkflowInstance(ctx context.Context, instance *workflow.Instance, timeout time.Duration) error {
	if wc.client == nil {
		return fmt.Errorf("client is nil")
	}

	if timeout <= 0 {
		timeout = defaultWaitTimeout
	}

	if err := wc.client.WaitForWorkflowInstance(ctx, instance, timeout); err != nil {
		return fmt.Errorf("wait for workflow instance error: %w", err)
	}

	LogInfof("workflow instance %s finished", instance.InstanceID)
	return nil
}

// RemoveWorkflowInstance removes a completed workflow instance from the backend.
func (wc *WorkflowClient) RemoveWorkflowInstance(ctx context.Context, instance *core.WorkflowInstance) error {
	if wc.client == nil {
		return fmt.Errorf("client is nil")
	}

	if err := wc.client.RemoveWorkflowInstance(ctx, instance); err != nil {
		return fmt.Errorf("remove workflow instance error: %w", err)
	}

	LogInfof("removed workflow instance: %s", instance.InstanceID)
	return nil
}

// RemoveWorkflowInstances removes completed workflow instances from the backend.
func (wc *WorkflowClient) RemoveWorkflowInstances(ctx context.Context, opts ...backend.RemovalOption) error {
	if wc.client == nil {
		return fmt.Errorf("client is nil")
	}

	if err := wc.client.RemoveWorkflowInstances(ctx, opts...); err != nil {
		return fmt.Errorf("remove workflow instances error: %w", err)
	}

	LogInfo("removed workflow instances")
	return nil
}
