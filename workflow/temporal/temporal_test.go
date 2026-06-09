package temporal

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.temporal.io/sdk/temporal"

	enumspb "go.temporal.io/api/enums/v1"
)

func TestNewClient(t *testing.T) {
	wc, err := NewClient(
		WithClientHostPort("localhost:7233"),
		WithClientNamespace("default"),
	)
	// Will fail without a running server, but should not panic
	if err != nil {
		t.Logf("cant connect to temporal server, skip: %v", err)
		t.Skip()
	}
	defer func() { _ = wc.Close() }()
	assert.NotNil(t, wc)
}

func TestExecuteOptions(t *testing.T) {
	opts := ExecuteOptions{
		TaskQueue:        "my-task-queue",
		WorkflowID:       "test-workflow-1",
		RunTimeout:       time.Minute * 10,
		ExecutionTimeout: time.Hour,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second * 5,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    5,
		},
		CronSchedule:  "0 8 * * *",
		IDReusePolicy: enumspb.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}

	swo := toStartWorkflowOptions(opts)
	assert.Equal(t, "my-task-queue", swo.TaskQueue)
	assert.Equal(t, "test-workflow-1", swo.ID)
	assert.Equal(t, time.Minute*10, swo.WorkflowRunTimeout)
	assert.Equal(t, time.Hour, swo.WorkflowExecutionTimeout)
	assert.NotNil(t, swo.RetryPolicy)
	assert.Equal(t, "0 8 * * *", swo.CronSchedule)
}

func TestClientOptions(t *testing.T) {
	opts := ClientOptions{
		HostPort:  "localhost:7233",
		Namespace: "test-ns",
	}
	assert.Equal(t, "localhost:7233", opts.HostPort)
	assert.Equal(t, "test-ns", opts.Namespace)
}

func TestWorkerOptions(t *testing.T) {
	opts := WorkerOptions{
		TaskQueue: "my-task-queue",
		Workflows: []any{BrokerMessageWorkflow},
		Activities: []any{func(ctx context.Context, body []byte) error {
			return nil
		}},
	}
	assert.Equal(t, "my-task-queue", opts.TaskQueue)
	assert.Len(t, opts.Workflows, 1)
	assert.Len(t, opts.Activities, 1)
}
