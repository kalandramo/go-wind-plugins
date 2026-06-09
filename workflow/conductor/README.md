# Conductor Workflow

Conductor workflow module for [go-wind-plugins](https://github.com/tx7do/go-wind-plugins), based on [conductor-oss/conductor](https://github.com/conductor-oss/conductor) (Netflix Conductor).

## Overview

[Conductor](https://conductor-oss.github.io/conductor/) is an event-driven orchestration platform providing durable and highly resilient execution for microservices.

This module provides a Go wrapper around the [Conductor Go SDK](https://github.com/conductor-oss/go-sdk), exposing native Conductor concepts (Workflow, Task, Worker) without any broker abstraction.

## Installation

```bash
go get github.com/tx7do/go-wind-plugins/workflow/conductor
```

## Quick Start

### 1. Create a Client

```go
import conductor "github.com/tx7do/go-wind-plugins/workflow/conductor"

// Connect to a local Conductor server
client, err := conductor.NewClient(conductor.ClientOptions{
    ServerURL: "http://localhost:8080/api",
})
if err != nil {
    log.Fatal(err)
}
defer func() { _ = client.Close() }()

// Or from environment variables: CONDUCTOR_SERVER_URL, CONDUCTOR_AUTH_KEY, CONDUCTOR_AUTH_SECRET
client, err = conductor.NewClientFromEnv()
```

### 2. Start a Workflow

```go
import "context"

// Async start
id, err := client.StartWorkflow(context.Background(), conductor.StartWorkflowOptions{
    Name: "my_workflow",
    Input: map[string]interface{}{
        "name": "Gopher",
    },
})

// Sync start (wait until a specific task completes)
run, err := client.StartWorkflowSync(context.Background(), conductor.StartWorkflowOptions{
    Name: "my_workflow",
    Input: map[string]interface{}{
        "name": "Gopher",
    },
}, "")
```

### 3. Define a Task Worker

```go
import "github.com/conductor-sdk/conductor-go/sdk/model"

// Define your task handler
func GreetTask(task *model.Task) (interface{}, error) {
    name := task.InputData["name"]
    return map[string]interface{}{
        "greeting": "Hello, " + name.(string),
    }, nil
}

// Start the worker
worker, err := client.StartWorker("greet", GreetTask, 1, 100*time.Millisecond)
```

### 4. Manage Workflows

```go
// Get workflow status
workflow, err := client.GetWorkflow(ctx, workflowID, true)

// Pause / Resume
client.Pause(ctx, workflowID)
client.Resume(ctx, workflowID)

// Terminate
client.Terminate(ctx, workflowID, "no longer needed")

// Retry failed workflow
client.Retry(ctx, workflowID, false)

// Restart completed workflow
client.Restart(ctx, workflowID, true)
```

## Configuration

### ClientOptions

| Field       | Description                              | Default                      |
|-------------|------------------------------------------|------------------------------|
| ServerURL   | Conductor server API URL                 | `http://localhost:8080/api`  |
| AuthKey     | Authentication key (for Orkes Cloud)     | -                            |
| AuthSecret  | Authentication secret (for Orkes Cloud)  | -                            |

### WorkerConfig

| Field        | Description                        | Default    |
|--------------|------------------------------------|------------|
| TaskType     | Task definition name to poll for   | -          |
| Concurrency  | Number of concurrent worker threads| `1`        |
| PollInterval | Interval between poll requests     | `100ms`    |
| Domain       | Task domain for isolation           | -          |

## Architecture

Unlike message brokers (Kafka, RabbitMQ, etc.), Conductor is a **workflow orchestration engine**. It is not suitable for the `broker.Broker` interface. This module lives under `workflow/conductor` alongside `workflow/temporal`, `workflow/argo`, and `workflow/goworkflows`. All four implement the `workflow.Client` interface (`Close() error`).

## References

- [Conductor Documentation](https://conductor-oss.github.io/conductor/)
- [Conductor Go SDK](https://github.com/conductor-oss/go-sdk)
- [Conductor Go SDK Examples](https://github.com/conductor-oss/go-sdk/tree/main/examples)
