package temporal

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

const (
	defaultActivityName    = "ProcessMessage"
	defaultActivityTimeout = time.Minute * 5
)

// BrokerMessageWorkflow is a simple Temporal workflow that receives a message body
// and delegates processing to the registered "ProcessMessage" activity.
// For complex multi-step orchestration, register a custom workflow via WorkerOptions.Workflows.
func BrokerMessageWorkflow(ctx workflow.Context, body []byte) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: defaultActivityTimeout,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	return workflow.ExecuteActivity(ctx, defaultActivityName, body).Get(ctx, nil)
}
