package goworkflows

import (
	"time"

	"github.com/cschleiden/go-workflows/worker"
	"github.com/cschleiden/go-workflows/workflow"
)

////////////////////////////////////////////////////////////////////////////////
/// Create Workflow Options
////////////////////////////////////////////////////////////////////////////////

// CreateWorkflowOptions holds options for creating a new workflow instance.
type CreateWorkflowOptions struct {
	// InstanceID is the unique identifier for the workflow instance (required).
	InstanceID string

	// Queue is the queue the workflow instance will be created in.
	// Defaults to workflow.QueueDefault if empty.
	Queue workflow.Queue
}

////////////////////////////////////////////////////////////////////////////////
/// Worker Options
////////////////////////////////////////////////////////////////////////////////

// WorkerOptions holds options for configuring a worker.
type WorkerOptions struct {
	// WorkflowPollers is the number of workflow pollers to start (default: 2).
	WorkflowPollers int

	// MaxParallelWorkflowTasks is the maximum number of concurrent workflow tasks (default: 0 = unlimited).
	MaxParallelWorkflowTasks int

	// WorkflowHeartbeatInterval is the interval between heartbeat attempts on workflow tasks (default: 25s).
	WorkflowHeartbeatInterval time.Duration

	// WorkflowPollingInterval is the interval between polling for new workflow tasks (default: 200ms).
	WorkflowPollingInterval time.Duration

	// WorkflowExecutorCacheSize is the max size of the workflow executor cache (default: 128).
	WorkflowExecutorCacheSize int

	// WorkflowExecutorCacheTTL is the max TTL of the workflow executor cache (default: 10s).
	WorkflowExecutorCacheTTL time.Duration

	// WorkflowQueues are the queues the worker listens to for workflow tasks.
	WorkflowQueues []workflow.Queue

	// ActivityPollers is the number of activity pollers to start (default: 2).
	ActivityPollers int

	// MaxParallelActivityTasks is the maximum number of concurrent activity tasks (default: 0 = unlimited).
	MaxParallelActivityTasks int

	// ActivityHeartbeatInterval is the interval between heartbeat attempts for activity tasks (default: 25s).
	ActivityHeartbeatInterval time.Duration

	// ActivityPollingInterval is the interval between polling for new activity tasks (default: 200ms).
	ActivityPollingInterval time.Duration

	// ActivityQueues are the queues the worker listens to for activity tasks.
	ActivityQueues []workflow.Queue

	// SingleWorkerMode enables optimizations for single worker scenarios.
	SingleWorkerMode bool
}

////////////////////////////////////////////////////////////////////////////////
/// Defaults
////////////////////////////////////////////////////////////////////////////////

const (
	defaultWaitTimeout = 20 * time.Second
)

// toWorkerOptions converts WorkerOptions to the upstream worker.Options.
func (wo *WorkerOptions) toWorkerOptions() worker.Options {
	opts := worker.DefaultOptions

	if wo == nil {
		return opts
	}

	if wo.WorkflowPollers > 0 {
		opts.WorkflowPollers = wo.WorkflowPollers
	}
	if wo.MaxParallelWorkflowTasks > 0 {
		opts.MaxParallelWorkflowTasks = wo.MaxParallelWorkflowTasks
	}
	if wo.WorkflowHeartbeatInterval > 0 {
		opts.WorkflowHeartbeatInterval = wo.WorkflowHeartbeatInterval
	}
	if wo.WorkflowPollingInterval > 0 {
		opts.WorkflowPollingInterval = wo.WorkflowPollingInterval
	}
	if wo.WorkflowExecutorCacheSize > 0 {
		opts.WorkflowExecutorCacheSize = wo.WorkflowExecutorCacheSize
	}
	if wo.WorkflowExecutorCacheTTL > 0 {
		opts.WorkflowExecutorCacheTTL = wo.WorkflowExecutorCacheTTL
	}
	if len(wo.WorkflowQueues) > 0 {
		opts.WorkflowQueues = wo.WorkflowQueues
	}

	if wo.ActivityPollers > 0 {
		opts.ActivityPollers = wo.ActivityPollers
	}
	if wo.MaxParallelActivityTasks > 0 {
		opts.MaxParallelActivityTasks = wo.MaxParallelActivityTasks
	}
	if wo.ActivityHeartbeatInterval > 0 {
		opts.ActivityHeartbeatInterval = wo.ActivityHeartbeatInterval
	}
	if wo.ActivityPollingInterval > 0 {
		opts.ActivityPollingInterval = wo.ActivityPollingInterval
	}
	if len(wo.ActivityQueues) > 0 {
		opts.ActivityQueues = wo.ActivityQueues
	}

	opts.SingleWorkerMode = wo.SingleWorkerMode

	return opts
}
