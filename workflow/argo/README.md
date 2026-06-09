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
client.RetryWorkflow(ctx, workflowName, "")

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

## Configuration

### ClientOptions

| Field               | Description                          | Default                    |
|---------------------|--------------------------------------|----------------------------|
| ServerURL           | Argo Server API URL                  | `https://localhost:2746`   |
| Namespace           | Default Kubernetes namespace         | `default`                  |
| Token               | Bearer token for authentication      | -                          |
| InsecureSkipVerify  | Skip TLS certificate verification    | `false`                    |

## Why REST API instead of Go SDK?

The official Argo Workflows Go SDK (`github.com/argoproj/argo-workflows/v4`) pulls in a massive dependency tree including `k8s.io/client-go`, `k8s.io/apimachinery`, `google.golang.org/grpc`, protobuf, etc. This adds hundreds of indirect dependencies to your project.

By using the REST API directly, this module has **zero external dependencies** — it only uses the Go standard library (`log/slog` for logging). This makes it lightweight and suitable for any Go project.

## Architecture

Argo Workflows is a **Kubernetes-native workflow orchestration engine**. It is not suitable for the `broker.Broker` interface. This module lives under `workflow/argo` alongside `workflow/temporal`, `workflow/conductor`, and `workflow/goworkflows`. All four implement the `workflow.Client` interface (`Close() error`).

## References

- [Argo Workflows Documentation](https://argoproj.github.io/argo-workflows/)
- [Argo Workflows REST API](https://argo-workflows.readthedocs.io/en/latest/rest-api/)
- [Argo Workflows GitHub](https://github.com/argoproj/argo-workflows)
