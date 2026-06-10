# AWS SQS

## 什么是 AWS SQS？

Amazon Simple Queue Service (SQS) 是一种完全托管的消息队列服务，支持标准队列和 FIFO 队列，提供高可扩展性和可靠性的消息传递。无需自行部署和维护消息中间件，AWS 全托管。

## 队列类型对比

| 特性 | 标准队列 | FIFO 队列 |
|------|----------|----------|
| 消息顺序 | 尽力排序（最佳努力） | 严格先进先出 |
| 投递语义 | 至少一次投递（可能有重复） | 精确一次投递（不重复） |
| 吞吐量 | 几乎无限 | 300 TPS/队列，3000 TPS/消息组 |
| 队列名称 | 任意名称 | 必须以 `.fifo` 结尾 |
| 适用场景 | 高吞吐、可容忍少量重复的场景 | 订单处理、金融交易等严格要求顺序和去重的场景 |

## 基本概念

- **Queue**：消息队列，存储消息的缓冲区。生产者发送消息到队列，消费者从队列接收消息。
- **Message**：消息体，最大 256KB（文本）或使用 S3 扩展大消息。
- **Visibility Timeout**：消息被消费者接收后对其他消费者隐藏的时间窗口，防止重复消费。
- **Dead Letter Queue (DLQ)**：死信队列，接收超过最大接收次数的消息，用于异常处理。
- **Long Polling**：`WaitTimeSeconds` 控制长轮询等待时间，减少空响应和 API 调用。
- **Delay Queue**：队列级别的延迟投递（0-15 分钟）。

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

## 注意事项

- 标准队列不保证消息顺序，也可能会重复投递
- FIFO 队列名称必须以 `.fifo` 结尾
- 消息体最大 256KB，超过需配合 S3 使用
- `Visibility Timeout` 期间消息对其他消费者不可见
- 本地开发推荐使用 [ElasticMQ](https://github.com/softwaremill/elasticmq) 或 [LocalStack](https://localstack.cloud/)

## 参考资料

- [AWS SQS 官方文档](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/welcome.html)
- [AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/)
- [ElasticMQ — SQS 兼容的本地模拟器](https://github.com/softwaremill/elasticmq)
