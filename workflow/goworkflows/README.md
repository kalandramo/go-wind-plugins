# Go Workflows

[go-workflows](https://github.com/cschleiden/go-workflows) workflow engine module for [go-wind-plugins](https://github.com/tx7do/go-wind-plugins).

## Overview

[go-workflows](https://github.com/cschleiden/go-workflows) is a lightweight, embeddable durable workflow engine for Go. It provides durable execution, activities, signals, and timers with pluggable backends (SQLite, Redis, MySQL, etc.).

This module wraps the go-workflows SDK, providing a high-level client and worker with sensible defaults while exposing the underlying SDK for advanced use cases.

## Installation

```bash
go get github.com/tx7do/go-wind-plugins/workflow/goworkflows
```

## Quick Start

### 1. Create a Client

```go
import (
    "github.com/cschleiden/go-workflows/backend"
    "github.com/cschleiden/go-workflows/backend/sqlite"
    gowf "github.com/tx7do/go-wind-plugins/workflow/goworkflows"
)

// Choose a backend (SQLite shown here; Redis, MySQL also available)
b := sqlite.NewSqliteBackend("workflows.db")

client, err := gowf.NewClient(b)
if err != nil {
    log.Fatal(err)
}
defer func() { _ = client.Close() }()
```

### 2. Create a Workflow Instance

```go
import (
    "context"
    "github.com/cschleiden/go-workflows/workflow"
)

// Define a workflow function
func MyWorkflow(ctx workflow.Context, input string) (string, error) {
    var result string
    if err := workflow.ExecuteActivity(ctx, "MyActivity", input).Get(ctx, &result); err != nil {
        return "", err
    }
    return result, nil
}

// Create an instance
instance, err := client.CreateWorkflowInstance(context.Background(),
    gowf.CreateWorkflowOptions{
        InstanceID: "my-workflow-1",
    },
    MyWorkflow,
    "hello world",
)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("started: %s (execution: %s)\n", instance.InstanceID, instance.ExecutionID)
```

### 3. Start a Worker

```go
// Define an activity
func MyActivity(ctx context.Context, input string) (string, error) {
    return "processed: " + input, nil
}

// Create a worker (processes both workflows and activities)
worker, err := gowf.NewWorker(b, &gowf.WorkerOptions{
    WorkflowQueues: []workflow.Queue{"default"},
    ActivityQueues: []workflow.Queue{"default"},
})
if err != nil {
    log.Fatal(err)
}

// Register workflow and activity
worker.RegisterWorkflow(MyWorkflow)
worker.RegisterActivity(MyActivity)

// Start the worker (blocks until context is cancelled)
if err := worker.Start(context.Background()); err != nil {
    log.Fatal(err)
}
// worker.Stop() to stop gracefully
// worker.WaitForCompletion() to drain in-flight tasks
```

### 4. Signal and Query

```go
// Send a signal to a running workflow
err := client.SignalWorkflow(ctx, "my-workflow-1", "cancel", nil)

// Get workflow state
state, err := client.GetWorkflowInstanceState(ctx, instance)

// Wait for workflow to finish
err = client.WaitForWorkflowInstance(ctx, instance, 30*time.Second)

// Cancel a running workflow
err = client.CancelWorkflowInstance(ctx, instance)
```

## API Reference

### WorkflowClient Methods

| Method | Description |
|--------|-------------|
| `NewClient(backend)` | Create a client with the given backend |
| `CreateWorkflowInstance(ctx, opts, wf, args...)` | Create and start a workflow instance |
| `CancelWorkflowInstance(ctx, instance)` | Cancel a running instance |
| `SignalWorkflow(ctx, instanceID, name, arg)` | Send a signal to a running instance |
| `GetWorkflowInstanceState(ctx, instance)` | Get the current state of an instance |
| `WaitForWorkflowInstance(ctx, instance, timeout)` | Block until the instance finishes (default 20s) |
| `RemoveWorkflowInstance(ctx, instance)` | Remove a completed instance from the backend |
| `RemoveWorkflowInstances(ctx, opts...)` | Remove multiple completed instances |
| `Backend()` | Get the underlying `backend.Backend` |
| `InnerClient()` | Get the underlying go-workflows `client.Client` |
| `Close()` | Close the client and backend |

### WorkflowWorker Methods

| Method | Description |
|--------|-------------|
| `NewWorker(backend, opts)` | Create a worker (workflows + activities) |
| `NewWorkflowOnlyWorker(backend, opts)` | Create a workflow-only worker |
| `NewActivityOnlyWorker(backend, opts)` | Create an activity-only worker |
| `RegisterWorkflow(wf)` | Register a workflow function |
| `RegisterActivity(a)` | Register an activity function or struct |
| `Start(ctx)` | Start the worker (blocks until ctx is cancelled) |
| `Stop()` | Signal the worker to stop |
| `WaitForCompletion()` | Wait for in-flight tasks to complete |
| `IsRunning()` | Check if the worker is running |

### Specialized Workers

| Constructor | Processes Workflows | Processes Activities |
|-------------|:-------------------:|:--------------------:|
| `NewWorker` | Yes | Yes |
| `NewWorkflowOnlyWorker` | Yes | No |
| `NewActivityOnlyWorker` | No | Yes |

This is useful for scaling workflows and activities independently across different processes.

## Configuration

### CreateWorkflowOptions

| Field | Type | Description |
|-------|------|-------------|
| `InstanceID` | `string` | Unique identifier for the workflow instance (**required**) |
| `Queue` | `workflow.Queue` | Queue for the instance (defaults to `QueueDefault`) |

### WorkerOptions

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `WorkflowPollers` | `int` | `2` | Number of workflow pollers |
| `MaxParallelWorkflowTasks` | `int` | `0` (unlimited) | Max concurrent workflow tasks |
| `WorkflowHeartbeatInterval` | `time.Duration` | `25s` | Heartbeat interval for workflow tasks |
| `WorkflowPollingInterval` | `time.Duration` | `200ms` | Interval between workflow task polls |
| `WorkflowExecutorCacheSize` | `int` | `128` | Max workflow executor cache entries |
| `WorkflowExecutorCacheTTL` | `time.Duration` | `10s` | TTL of workflow executor cache |
| `WorkflowQueues` | `[]workflow.Queue` | - | Queues to listen on for workflows |
| `ActivityPollers` | `int` | `2` | Number of activity pollers |
| `MaxParallelActivityTasks` | `int` | `0` (unlimited) | Max concurrent activity tasks |
| `ActivityHeartbeatInterval` | `time.Duration` | `25s` | Heartbeat interval for activity tasks |
| `ActivityPollingInterval` | `time.Duration` | `200ms` | Interval between activity task polls |
| `ActivityQueues` | `[]workflow.Queue` | - | Queues to listen on for activities |
| `SingleWorkerMode` | `bool` | `false` | Enable single-worker optimizations |

## Architecture

go-workflows is an embeddable durable workflow engine with pluggable backends. Unlike Temporal (which requires a separate server), go-workflows runs in-process and stores state in a database backend (SQLite, Redis, MySQL, etc.).

This module lives under `workflow/goworkflows` alongside `workflow/temporal`, `workflow/argo`, and `workflow/conductor`. All four implement the `workflow.Client` interface (`Close() error`). The `WorkflowWorker` in this module implements the `workflow.Worker` interface (`Stop()` + `IsRunning()`).

## References

- [go-workflows GitHub](https://github.com/cschleiden/go-workflows)
- [go-workflows Examples](https://github.com/cschleiden/go-workflows/tree/main/examples)
- [Supported Backends](https://github.com/cschleiden/go-workflows#backends)
