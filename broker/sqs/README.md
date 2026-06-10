# AWS SQS

## 什么是 AWS SQS？

Amazon Simple Queue Service (SQS) 是一种完全托管的消息队列服务，支持标准队列和 FIFO 队列，提供高可扩展性和可靠性的消息传递。

## 使用方式

### 基础：发布/订阅

```go
b := sqs.NewBroker(
    broker.WithCodec("json"),
    sqs.WithRegion("us-east-1"),
    sqs.WithEndpoint("http://localhost:9324"), // 本地 ElasticMQ
)
b.Init()
b.Connect()
defer b.Disconnect()

// 发布（topic 即队列名）
b.Publish(ctx, "my-queue", broker.NewMessage(msg))

// 订阅
b.Subscribe("my-queue", handler, binder)
```

### 高级：FIFO 队列

```go
b.Publish(ctx, "my-queue.fifo", broker.NewMessage(msg),
    sqs.WithMessageGroupId("group-1"),
    sqs.WithMessageDeduplicationId("dedup-123"),
)
```

### 高级：延迟消息

```go
b.Publish(ctx, "my-queue", broker.NewMessage(msg),
    sqs.WithDelaySeconds(60), // 延迟 60 秒投递
)
```

### 高级：长轮询配置

```go
b.Subscribe("my-queue", handler, binder,
    sqs.WithVisibilityTimeout(60),
    sqs.WithWaitTimeSeconds(20),
    sqs.WithMaxMessages(10),
)
```

## 配置选项

### Broker 选项

| 选项 | 说明 |
|------|------|
| `sqs.WithRegion(region)` | AWS 区域（默认 `us-east-1`） |
| `sqs.WithEndpoint(url)` | 自定义端点（用于 ElasticMQ/LocalStack） |
| `sqs.WithQueueUrl(url)` | 默认队列 URL |

### Publish 选项

| 选项 | 说明 |
|------|------|
| `sqs.WithDelaySeconds(n)` | 延迟投递秒数（0-900） |
| `sqs.WithMessageGroupId(id)` | FIFO 队列消息组 ID |
| `sqs.WithMessageDeduplicationId(id)` | FIFO 队列去重 ID |

### Subscribe 选项

| 选项 | 说明 |
|------|------|
| `sqs.WithVisibilityTimeout(n)` | 可见性超时秒数（默认 30） |
| `sqs.WithWaitTimeSeconds(n)` | 长轮询等待秒数（默认 20） |
| `sqs.WithMaxMessages(n)` | 每次拉取最大消息数（默认 10） |

## Docker 部署开发环境

```shell
# 使用 ElasticMQ（SQS 兼容）
docker run -d \
    --name elasticmq \
    -p 9324:9324 \
    -p 9325:9325 \
    softwaremill/elasticmq-native
```

管理后台: <http://localhost:9325>
