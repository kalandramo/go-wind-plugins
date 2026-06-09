package temporal

import (
	"context"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"

	enumspb "go.temporal.io/api/enums/v1"
)

////////////////////////////////////////////////////////////////////////////////
/// Client Options
////////////////////////////////////////////////////////////////////////////////

type ClientOptions struct {
	// HostPort is the Temporal server host:port (default: "localhost:7233").
	HostPort string

	// Namespace is the Temporal namespace (default: "default").
	Namespace string

	// Context for passing additional values.
	Context context.Context
}

type clientHostPortKey struct{}
type clientNamespaceKey struct{}

// WithClientHostPort sets the Temporal server address.
func WithClientHostPort(hostPort string) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.HostPort = hostPort
	}
}

// WithClientNamespace sets the Temporal namespace.
func WithClientNamespace(namespace string) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.Namespace = namespace
	}
}

////////////////////////////////////////////////////////////////////////////////
/// Execute Workflow Options
////////////////////////////////////////////////////////////////////////////////

type ExecuteOptions struct {
	// TaskQueue is the task queue for the workflow (defaults to the topic).
	TaskQueue string

	// WorkflowID is a unique identifier for the workflow execution.
	WorkflowID string

	// WorkflowFn is the workflow function to execute.
	WorkflowFn any

	// RunTimeout is the maximum time a single workflow run is allowed to execute.
	RunTimeout time.Duration

	// ExecutionTimeout is the maximum total time including retries and continue-as-new.
	ExecutionTimeout time.Duration

	// TaskTimeout is the timeout for a single workflow task.
	TaskTimeout time.Duration

	// RetryPolicy is the retry policy for the workflow.
	RetryPolicy *temporal.RetryPolicy

	// CronSchedule schedules the workflow to run on a cron expression.
	CronSchedule string

	// IDReusePolicy controls behavior when WorkflowID already exists.
	IDReusePolicy enumspb.WorkflowIdReusePolicy

	// Context for passing additional values.
	Context context.Context
}

////////////////////////////////////////////////////////////////////////////////
/// Worker Options
////////////////////////////////////////////////////////////////////////////////

type WorkerOptions struct {
	// TaskQueue is the task queue the worker listens to.
	TaskQueue string

	// WorkerOptions are the native Temporal worker options.
	Options worker.Options

	// Workflows is a list of additional workflow functions to register.
	Workflows []any

	// Activities is a list of additional activity functions/structs to register.
	Activities []any

	// Context for passing additional values.
	Context context.Context
}

////////////////////////////////////////////////////////////////////////////////
/// Helper: apply ExecuteOptions to StartWorkflowOptions
////////////////////////////////////////////////////////////////////////////////

func toStartWorkflowOptions(opts ExecuteOptions) client.StartWorkflowOptions {
	swo := client.StartWorkflowOptions{
		TaskQueue: opts.TaskQueue,
	}

	if opts.WorkflowID != "" {
		swo.ID = opts.WorkflowID
	}
	if opts.RunTimeout > 0 {
		swo.WorkflowRunTimeout = opts.RunTimeout
	}
	if opts.ExecutionTimeout > 0 {
		swo.WorkflowExecutionTimeout = opts.ExecutionTimeout
	}
	if opts.TaskTimeout > 0 {
		swo.WorkflowTaskTimeout = opts.TaskTimeout
	}
	if opts.RetryPolicy != nil {
		swo.RetryPolicy = opts.RetryPolicy
	}
	if opts.CronSchedule != "" {
		swo.CronSchedule = opts.CronSchedule
	}
	swo.WorkflowIDReusePolicy = opts.IDReusePolicy

	return swo
}
