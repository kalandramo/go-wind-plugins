package conductor

import (
	"time"

	"github.com/conductor-sdk/conductor-go/sdk/model"
)

////////////////////////////////////////////////////////////////////////////////
/// Client Options
////////////////////////////////////////////////////////////////////////////////

// ClientOptions configures the Conductor API client.
type ClientOptions struct {
	// ServerURL is the Conductor server API URL (e.g. "http://localhost:8080/api").
	ServerURL string

	// AuthKey is the authentication key (optional, for Orkes Cloud).
	AuthKey string

	// AuthSecret is the authentication secret (optional, for Orkes Cloud).
	AuthSecret string
}

// StartWorkflowOptions holds options for starting a workflow execution.
type StartWorkflowOptions struct {
	// Name is the workflow definition name.
	Name string

	// Version is the workflow definition version (default: latest).
	Version *int32

	// Input is the workflow input data.
	Input map[string]interface{}

	// CorrelationID is used for message correlation.
	CorrelationID string

	// Priority is the workflow priority.
	Priority *int32
}

// WorkerConfig holds configuration for a task worker.
type WorkerConfig struct {
	// TaskType is the task definition name this worker polls for.
	TaskType string

	// Concurrency is the number of concurrent worker threads (default: 1).
	Concurrency int

	// PollInterval is the interval between poll requests (default: 100ms).
	PollInterval time.Duration

	// Domain is the task domain for isolation (optional).
	Domain string
}

////////////////////////////////////////////////////////////////////////////////
/// Defaults
////////////////////////////////////////////////////////////////////////////////

const (
	defaultServerURL    = "http://localhost:8080/api"
	defaultConcurrency  = 1
	defaultPollInterval = 100 * time.Millisecond
)

////////////////////////////////////////////////////////////////////////////////
/// Helper: convert StartWorkflowOptions to StartWorkflowRequest
////////////////////////////////////////////////////////////////////////////////

func toStartWorkflowRequest(opts StartWorkflowOptions) *model.StartWorkflowRequest {
	req := &model.StartWorkflowRequest{
		Name:  opts.Name,
		Input: opts.Input,
	}
	if opts.Version != nil {
		req.Version = *opts.Version
	}
	if opts.CorrelationID != "" {
		req.CorrelationId = opts.CorrelationID
	}
	if opts.Priority != nil {
		req.Priority = *opts.Priority
	}
	return req
}
