# Argo Workflows

Argo Workflows module for [go-wind-plugins](https://github.com/tx7do/go-wind-plugins), based on [argoproj/argo-workflows](https://github.com/argoproj/argo-workflows).

## Overview

[Argo Workflows](https://argoproj.github.io/argo-workflows/) is an open-source container-native workflow engine for orchestrating parallel jobs on Kubernetes.

This module provides a lightweight Go wrapper around the **Argo Server REST API**, with zero dependency on the heavy Argo Workflows Go SDK or Kubernetes client-go. All you need is an Argo Server endpoint and an auth token.

## Installation

```bash
go get github.com/tx7do/go-wind-plugins/workflow/argo
```

## Quick Start

### 1. Create a Client

```go
import argowf "github.com/tx7do/go-wind-plugins/workflow/argo"

client, err := argowf.NewClient(argowf.ClientOptions{
    ServerURL:         "https://localhost:2746",
    Namespace:         "default",
    Token:             "Bearer token",
    InsecureSkipVerify: true, // for development only
})
if err != nil {
    log.Fatal(err)
}
defer func() { _ = client.Close() }()
```

### 2. Submit a Workflow

```go
wf, err := client.SubmitWorkflow(context.Background(), &argowf.Workflow{
    APIVersion: "argoproj.io/v1alpha1",
    Kind:       "Workflow",
    Metadata: argowf.ObjectMeta{
        GenerateName: "hello-world-",
    },
    Spec: argowf.WorkflowSpec{
        Entrypoint: "whalesay",
        Templates: []argowf.Template{
            {
                Name: "whalesay",
                Container: &argowf.Container{
                    Image:   "docker/whalesay:latest",
                    Command: []string{"cowsay"},
                    Args:    []string{"Hello World"},
                },
            },
        },
    },
}, nil)
```

### 3. Get Workflow Status

```go
wf, err := client.GetWorkflow(ctx, "hello-world-abc123", "")
if wf.Status != nil {
    fmt.Printf("Phase: %s\n", wf.Status.Phase)
}
```

### 4. Manage Workflows

```go
// Suspend
client.SuspendWorkflow(ctx, workflowName, "")

// Resume
client.ResumeWorkflow(ctx, workflowName, "")

// Terminate
client.TerminateWorkflow(ctx, workflowName, "")

// Retry failed
retryedWf, err := client.RetryWorkflow(ctx, workflowName, "")

// Resubmit
resubmittedWf, err := client.ResubmitWorkflow(ctx, workflowName, "")

// Stop
client.StopWorkflow(ctx, workflowName, "", "no longer needed")

// Delete
client.DeleteWorkflow(ctx, workflowName, "")

// List
list, err := client.ListWorkflows(ctx, &argowf.ListOptions{
    LabelSelector: "workflows.argoproj.io/workflow-template=my-template",
    Limit: 10,
})
```

### 5. Workflow Logs

```go
// Get logs for an entire workflow
logs, err := client.GetWorkflowLogs(ctx, workflowName, "", "")
fmt.Println(logs)

// Get logs for a specific pod
logs, err := client.GetWorkflowLogs(ctx, workflowName, "", "hello-world-abc123")
```

## Configuration

### ClientOptions

| Field               | Description                          | Default                    |
|---------------------|--------------------------------------|----------------------------|
| ServerURL           | Argo Server API URL                  | `https://localhost:2746`   |
| Namespace           | Default Kubernetes namespace         | `default`                  |
| Token               | Bearer token for authentication      | -                          |
| InsecureSkipVerify  | Skip TLS certificate verification    | `false`                    |

### SubmitOptions

| Field        | Description                                          | Default |
|--------------|------------------------------------------------------|---------|
| Namespace    | Target namespace (overrides client default)          | -       |
| ServerDryRun | Dry run without creating the workflow                | `false` |
| Parameters   | Workflow parameters in `"key=value"` format         | -       |

### ListOptions

| Field          | Description                                | Default |
|----------------|--------------------------------------------|---------|
| Namespace      | Target namespace (overrides client default)| -       |
| LabelSelector  | Filter workflows by Kubernetes label       | -       |
| FieldSelector  | Filter workflows by Kubernetes field       | -       |
| Limit          | Maximum number of results                  | -       |
| Offset         | Pagination offset                          | -       |

## Why REST API instead of Go SDK?

The official Argo Workflows Go SDK (`github.com/argoproj/argo-workflows/v4`) pulls in a massive dependency tree including `k8s.io/client-go`, `k8s.io/apimachinery`, `google.golang.org/grpc`, protobuf, etc. This adds hundreds of indirect dependencies to your project.

By using the REST API directly, this module has **zero external dependencies** — it only uses the Go standard library (`log/slog` for logging). This makes it lightweight and suitable for any Go project.

## API Reference

### WorkflowClient Methods

| Method | Description |
|--------|-------------|
| `NewClient(opts)` | Create a client connected to Argo Server |
| `SubmitWorkflow(ctx, wf, opts)` | Submit a new workflow, returns created Workflow |
| `GetWorkflow(ctx, name, ns)` | Retrieve a workflow by name |
| `ListWorkflows(ctx, opts)` | List workflows with filtering and pagination |
| `DeleteWorkflow(ctx, name, ns)` | Delete a workflow |
| `SuspendWorkflow(ctx, name, ns)` | Suspend a running workflow |
| `ResumeWorkflow(ctx, name, ns)` | Resume a suspended workflow |
| `TerminateWorkflow(ctx, name, ns)` | Terminate a running workflow |
| `RetryWorkflow(ctx, name, ns)` | Retry a failed workflow, returns updated Workflow |
| `ResubmitWorkflow(ctx, name, ns)` | Resubmit a workflow, returns new Workflow |
| `StopWorkflow(ctx, name, ns, msg)` | Stop a workflow with an optional message |
| `GetWorkflowLogs(ctx, name, ns, podName)` | Retrieve logs for a workflow or specific pod |
| `Close()` | Close the client and release resources |

### Workflow Phase

| Phase | Terminal? | Description |
|-------|-----------|-------------|
| `PhasePending` | No | Workflow is queued/pending |
| `PhaseRunning` | No | Workflow is executing |
| `PhaseSucceeded` | Yes | Workflow completed successfully |
| `PhaseFailed` | Yes | Workflow completed with failure |
| `PhaseError` | Yes | Workflow encountered an error |

Use `phase.IsTerminal()` to check if a workflow has reached a final state.

## Architecture

Argo Workflows is a **Kubernetes-native workflow orchestration engine**. It is not suitable for the `broker.Broker` interface. This module lives under `workflow/argo` alongside `workflow/temporal`, `workflow/conductor`, and `workflow/goworkflows`. All four implement the `workflow.Client` interface (`Close() error`).

## References

- [Argo Workflows Documentation](https://argoproj.github.io/argo-workflows/)
- [Argo Workflows REST API](https://argo-workflows.readthedocs.io/en/latest/rest-api/)
- [Argo Workflows GitHub](https://github.com/argoproj/argo-workflows)
