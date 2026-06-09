package temporal

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	semConv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"go.temporal.io/sdk/client"
)

const (
	defaultHostPort  = "localhost:7233"
	defaultNamespace = "default"

	tracerMessageSystemKey = "temporal"
	tracerName             = "go-wind-plugins/workflow/temporal"
	spanNameProducer       = "temporal-producer"
	spanNameConsumer       = "temporal-consumer"
)

// WorkflowClient provides high-level access to Temporal workflow operations.
type WorkflowClient struct {
	client  client.Client
	options ClientOptions

	tracer trace.Tracer
}

// NewClient creates a new WorkflowClient and connects to the Temporal server.
func NewClient(opts ...func(*ClientOptions)) (*WorkflowClient, error) {
	options := ClientOptions{
		HostPort:  defaultHostPort,
		Namespace: defaultNamespace,
		Context:   context.Background(),
	}
	for _, o := range opts {
		o(&options)
	}

	c, err := client.NewClient(client.Options{
		HostPort:  options.HostPort,
		Namespace: options.Namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to temporal server at %s: %w", options.HostPort, err)
	}

	LogInfof("connected to temporal server at %s (namespace: %s)", options.HostPort, options.Namespace)

	return &WorkflowClient{
		client:  c,
		options: options,
	}, nil
}

// WithTracing enables OpenTelemetry tracing for the client.
func (wc *WorkflowClient) WithTracing() {
	wc.tracer = otel.Tracer(tracerName)
}

// Close closes the underlying Temporal client connection.
func (wc *WorkflowClient) Close() error {
	if wc.client != nil {
		wc.client.Close()
	}
	LogInfo("disconnected from temporal server")
	return nil
}

// TemporalClient returns the underlying Temporal SDK client for advanced operations.
func (wc *WorkflowClient) TemporalClient() client.Client {
	return wc.client
}

// Execute starts a workflow execution asynchronously (fire-and-forget).
// Returns the workflow run ID.
func (wc *WorkflowClient) Execute(ctx context.Context, args any, opts ExecuteOptions) (string, error) {
	if wc.client == nil {
		return "", fmt.Errorf("temporal client is not connected")
	}

	workflowFn := opts.WorkflowFn
	if workflowFn == nil {
		workflowFn = BrokerMessageWorkflow
	}

	swo := toStartWorkflowOptions(opts)

	ctx, span := wc.startProducerSpan(ctx, opts.TaskQueue)

	we, err := wc.client.ExecuteWorkflow(ctx, swo, workflowFn, args)

	wc.finishProducerSpan(span, err)

	if err != nil {
		return "", fmt.Errorf("execute workflow error: %w", err)
	}

	return we.GetRunID(), nil
}

// ExecuteSync starts a workflow and blocks until it completes, returning the result.
func (wc *WorkflowClient) ExecuteSync(ctx context.Context, args any, opts ExecuteOptions) ([]byte, error) {
	if wc.client == nil {
		return nil, fmt.Errorf("temporal client is not connected")
	}

	workflowFn := opts.WorkflowFn
	if workflowFn == nil {
		workflowFn = BrokerMessageWorkflow
	}

	swo := toStartWorkflowOptions(opts)

	ctx, span := wc.startProducerSpan(ctx, opts.TaskQueue)

	we, err := wc.client.ExecuteWorkflow(ctx, swo, workflowFn, args)

	if err != nil {
		wc.finishProducerSpan(span, err)
		return nil, fmt.Errorf("execute workflow error: %w", err)
	}

	var result []byte
	if err = we.Get(ctx, &result); err != nil {
		wc.finishProducerSpan(span, err)
		return nil, fmt.Errorf("get workflow result error: %w", err)
	}

	wc.finishProducerSpan(span, nil)

	return result, nil
}

// Signal sends a signal to a running workflow.
func (wc *WorkflowClient) Signal(ctx context.Context, workflowID, runID, signalName string, arg any) error {
	if wc.client == nil {
		return fmt.Errorf("temporal client is not connected")
	}
	return wc.client.SignalWorkflow(ctx, workflowID, runID, signalName, arg)
}

// Query queries the state of a running workflow.
func (wc *WorkflowClient) Query(ctx context.Context, workflowID, runID, queryType string, arg any) (any, error) {
	if wc.client == nil {
		return nil, fmt.Errorf("temporal client is not connected")
	}
	result, err := wc.client.QueryWorkflow(ctx, workflowID, runID, queryType, arg)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Cancel requests cancellation of a running workflow.
func (wc *WorkflowClient) Cancel(ctx context.Context, workflowID, runID string) error {
	if wc.client == nil {
		return fmt.Errorf("temporal client is not connected")
	}
	return wc.client.CancelWorkflow(ctx, workflowID, runID)
}

// Describe retrieves information about a workflow execution.
func (wc *WorkflowClient) Describe(ctx context.Context, workflowID, runID string) error {
	if wc.client == nil {
		return fmt.Errorf("temporal client is not connected")
	}
	_, err := wc.client.DescribeWorkflowExecution(ctx, workflowID, runID)
	return err
}

////////////////////////////////////////////////////////////////////////////////
/// Tracing helpers (native OpenTelemetry)
////////////////////////////////////////////////////////////////////////////////

func (wc *WorkflowClient) startProducerSpan(ctx context.Context, topic string) (context.Context, trace.Span) {
	if wc.tracer == nil {
		return ctx, nil
	}

	return wc.tracer.Start(ctx, spanNameProducer,
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			semConv.MessagingSystemKey.String(tracerMessageSystemKey),
			semConv.MessagingDestinationKindTopic,
			semConv.MessagingDestinationKey.String(topic),
		),
	)
}

func (wc *WorkflowClient) finishProducerSpan(span trace.Span, err error) {
	if span == nil {
		return
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}

func (wc *WorkflowClient) startConsumerSpan(ctx context.Context, topic string) (context.Context, trace.Span) {
	if wc.tracer == nil {
		return ctx, nil
	}

	return wc.tracer.Start(ctx, spanNameConsumer,
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			semConv.MessagingSystemKey.String(tracerMessageSystemKey),
			semConv.MessagingDestinationKindTopic,
			semConv.MessagingDestinationKey.String(topic),
			semConv.MessagingOperationReceive,
		),
	)
}

func (wc *WorkflowClient) finishConsumerSpan(span trace.Span, err error) {
	if span == nil {
		return
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}
