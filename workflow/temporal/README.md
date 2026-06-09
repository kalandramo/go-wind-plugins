# Temporal Workflow

[Temporal](https://temporal.io/) 工作流编排引擎的 Go 封装，提供原生的 Temporal 抽象。

> **注意**：本模块独立于 `broker` 包，使用 Temporal 自身的概念（Workflow/Activity/Signal/Query）而非强行套用消息代理模型。

## 核心概念

| 概念 | 说明 |
|------|------|
| **WorkflowClient** | 工作流客户端，负责连接、执行工作流、Signal、Query、Cancel |
| **WorkflowWorker** | 工作流 Worker，轮询 Task Queue，执行 Workflow 和 Activity |
| **BrokerMessageWorkflow** | 内置的简单工作流：接收消息体 → 调用 Activity 处理 |
| **ExecuteOptions** | 执行工作流的选项（TaskQueue、超时、重试、Cron 等） |
| **WorkerOptions** | Worker 配置（TaskQueue、注册的 Workflow/Activity） |

## Docker 部署

```shell
temporal server start-dev
```

- gRPC 端口：`7233`
- Web UI：`http://localhost:8233`
- 默认命名空间：`default`

## 使用方式

### 创建客户端

```go
import (
    temporal "github.com/tx7do/go-wind-plugins/workflow/temporal"
)

wc, err := temporal.NewClient(
    temporal.WithClientHostPort("localhost:7233"),
    temporal.WithClientNamespace("default"),
)
if err != nil {
    log.Fatal(err)
}
defer func() { _ = wc.Close() }()

// 可选：启用 OpenTelemetry 链路追踪
wc.WithTracing()
```

### 执行工作流（异步）

```go
runID, err := wc.Execute(ctx, []byte(`{"hello":"world"}`), temporal.ExecuteOptions{
    TaskQueue: "my-task-queue",
    WorkflowID: "my-workflow-1",
})
```

### 执行工作流（同步等待结果）

```go
result, err := wc.ExecuteSync(ctx, []byte(`{"hello":"world"}`), temporal.ExecuteOptions{
    TaskQueue:  "my-task-queue",
    WorkflowID: "sync-workflow-1",
    RunTimeout: time.Minute,
})
```

### 启动 Worker（最简模式）

```go
ww, err := wc.StartSimpleWorker(ctx, "my-task-queue",
    func(ctx context.Context, body []byte) error {
        fmt.Printf("received: %s\n", string(body))
        return nil
    },
)
if err != nil {
    log.Fatal(err)
}
// ww.Stop() 停止 Worker

// StartSimpleWorker 还支持传入额外的 WorkerOptions 函数选项：
ww2, err := wc.StartSimpleWorker(ctx, "another-queue",
    handler,
    func(o *temporal.WorkerOptions) {
        o.Workflows = []any{MyCustomWorkflow}
    },
)
```

### 启动 Worker（高级模式 — 自定义 Workflow/Activity）

```go
// 定义自定义 Workflow
func OrderWorkflow(ctx workflow.Context, order *Order) error {
    if err := workflow.ExecuteActivity(ctx, "DeductStock", order).Get(ctx, nil); err != nil {
        return err
    }
    return workflow.ExecuteActivity(ctx, "ProcessPayment", order).Get(ctx, nil)
}

// 定义 Activity
func DeductStock(ctx context.Context, order *Order) error { ... }
func ProcessPayment(ctx context.Context, order *Order) error { ... }

// 创建并启动 Worker
ww, err := wc.NewWorker(temporal.WorkerOptions{
    TaskQueue:  "order-task-queue",
    Workflows:  []any{OrderWorkflow},
    Activities: []any{DeductStock, ProcessPayment},
})
if err != nil {
    log.Fatal(err)
}
ww.Start()
```

### 高级选项（超时/重试/Cron）

```go
runID, err := wc.Execute(ctx, orderData, temporal.ExecuteOptions{
    TaskQueue:        "order-task-queue",
    WorkflowID:       "order-12345",
    WorkflowFn:       OrderWorkflow,
    RunTimeout:       10 * time.Minute,
    ExecutionTimeout: time.Hour,
    RetryPolicy: &temporal.RetryPolicy{
        InitialInterval:    5 * time.Second,
        BackoffCoefficient: 2.0,
        MaximumInterval:    time.Minute,
        MaximumAttempts:    5,
    },
    CronSchedule: "0 8 * * *",
})
```

### Signal / Query（通过底层 Client）

```go
tc := wc.TemporalClient()

// 发送 Signal
err := tc.SignalWorkflow(ctx, "order-12345", "", "cancel-signal", nil)

// 查询状态
result, err := tc.QueryWorkflow(ctx, "order-12345", "", "order-status")

// 取消工作流
err = wc.Cancel(ctx, "order-12345", "")
```

## API 一览

### WorkflowClient 方法

| 方法 | 说明 |
|------|------|
| `NewClient(opts...)` | 创建客户端连接 Temporal 服务器 |
| `WithTracing()` | 启用 OpenTelemetry 链路追踪（Producer/Consumer Span） |
| `Execute(ctx, args, opts)` | 异步启动工作流，返回 RunID |
| `ExecuteSync(ctx, args, opts)` | 同步执行工作流，阻塞等待结果 |
| `Signal(ctx, workflowID, runID, signal, arg)` | 发送 Signal |
| `Query(ctx, workflowID, runID, queryType, arg)` | 查询工作流状态 |
| `Cancel(ctx, workflowID, runID)` | 取消工作流 |
| `Describe(ctx, workflowID, runID)` | 获取工作流描述信息 |
| `TemporalClient()` | 获取底层 SDK Client |
| `NewWorker(opts)` | 创建 Worker |
| `StartSimpleWorker(ctx, taskQueue, handler, opts...)` | 最简 Worker，支持额外 WorkerOptions |
| `Close()` | 关闭客户端连接 |

### WorkflowWorker 方法

| 方法 | 说明 |
|------|------|
| `Start()` | 启动 Worker |
| `Stop()` | 停止 Worker |
| `RegisterWorkflow(fn)` | 注册 Workflow（须在 Start 前调用） |
| `RegisterActivity(fn)` | 注册 Activity（须在 Start 前调用） |
| `TaskQueue()` | 获取 Worker 监听的任务队列名 |
| `IsRunning()` | 检查 Worker 是否仍在运行 |

## 配置选项

### ClientOptions

| 选项 | 类型 | 说明 |
|------|------|------|
| `HostPort` | `string` | Temporal 服务器地址（默认 `localhost:7233`） |
| `Namespace` | `string` | 命名空间（默认 `default`） |
| `Context` | `context.Context` | 传递额外上下文值 |

### ExecuteOptions

| 选项 | 类型 | 说明 |
|------|------|------|
| `TaskQueue` | `string` | 任务队列名 |
| `WorkflowID` | `string` | 工作流唯一 ID |
| `WorkflowFn` | `any` | 自定义 Workflow 函数（默认 `BrokerMessageWorkflow`） |
| `RunTimeout` | `time.Duration` | 单次运行超时 |
| `ExecutionTimeout` | `time.Duration` | 总执行超时（含重试） |
| `TaskTimeout` | `time.Duration` | 单个任务超时 |
| `RetryPolicy` | `*temporal.RetryPolicy` | 重试策略 |
| `CronSchedule` | `string` | Cron 定时表达式 |
| `IDReusePolicy` | `WorkflowIdReusePolicy` | ID 复用策略 |
| `Context` | `context.Context` | 传递额外上下文值 |

### WorkerOptions

| 选项 | 类型 | 说明 |
|------|------|------|
| `TaskQueue` | `string` | Worker 监听的任务队列 |
| `Options` | `worker.Options` | 原生 Temporal Worker 选项 |
| `Workflows` | `[]any` | 额外注册的 Workflow 函数列表 |
| `Activities` | `[]any` | 额外注册的 Activity 函数/结构体列表 |
| `Context` | `context.Context` | 传递额外上下文值 |

## OpenTelemetry 链路追踪

调用 `WithTracing()` 后，模块会自动为以下操作创建 Span：

| 操作 | Span 名称 | Span Kind |
|------|-----------|-----------|
| `Execute` / `ExecuteSync` | `temporal-producer` | Producer |
| `ProcessMessage` Activity | `temporal-consumer` | Consumer |

Span 属性包含 `messaging.system=temporal` 和 `messaging.destination=<taskQueue>`。
