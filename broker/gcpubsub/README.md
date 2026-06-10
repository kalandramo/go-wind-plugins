# GCP Pub/Sub

## 什么是 GCP Pub/Sub？

Google Cloud Pub/Sub 是一种可扩展、灵活的全托管消息传递服务，支持异步通信，可处理数百万条消息。适用于流式分析、事件驱动架构和无服务器应用。

Pub/Sub 是 "Publish/Subscribe"（发布/订阅）的缩写，其核心概念包括：

- **Topic**：消息命名的资源，发布者向 Topic 发送消息。
- **Subscription**：订阅者从 Topic 接收消息的方式，支持 Pull 和 Push 两种投递模式。
- **Message**：消息载荷 + 属性键值对，最大 10MB。
- **Publisher**：创建并发送消息到 Topic 的应用。
- **Subscriber**：从 Subscription 消费消息的应用。
- ** Ordering Key**：消息排序键，保证同一键的消息按顺序投递。

## 投递模式

| 模式 | 说明 | 适用场景 |
|------|------|----------|
| Pull | 订阅者主动拉取消息，支持批量拉取和流量控制 | 高吞吐、需要精确流控的场景 |
| Push | Pub/Sub 将消息推送到 HTTPS Endpoint | 无服务器架构、Cloud Functions / Cloud Run |

## 使用方式

### 基础：发布/订阅

```go
b := gcpubsub.NewBroker(
    broker.WithCodec("json"),
    gcpubsub.WithProjectID("my-gcp-project"),
    gcpubsub.WithCredentialsFile("/path/to/service-account.json"),
)
b.Init()
b.Connect()
defer b.Disconnect()

// 发布
b.Publish(ctx, "my-topic", broker.NewMessage(msg))

// 订阅
b.Subscribe("my-topic", handler, binder,
    gcpubsub.WithSubscriptionName("my-subscription"),
)
```

### 高级：消息排序

```go
b.Publish(ctx, "my-topic", broker.NewMessage(msg),
    gcpubsub.WithPublishOrderingKey("order-key-1"),
)
```

### 高级：订阅配置

```go
b.Subscribe("my-topic", handler, binder,
    gcpubsub.WithSubscriptionName("my-sub"),
    gcpubsub.WithReceiveSettings(pubsub.ReceiveSettings{
        MaxOutstandingMessages: 1000,
        MaxOutstandingBytes:    1e9,
        NumGoroutines:          10,
    }),
)
```

## 配置选项

### Broker 选项

| 选项 | 说明 |
|------|------|
| `gcpubsub.WithProjectID(id)` | GCP 项目 ID（必填） |
| `gcpubsub.WithCredentialsFile(path)` | 服务账号 JSON 密钥文件路径 |
| `gcpubsub.WithEndpoint(url)` | 自定义端点（用于模拟器） |

### Publish 选项

| 选项 | 说明 |
|------|------|
| `gcpubsub.WithPublishTimeout(d)` | 发布超时时间 |
| `gcpubsub.WithPublishOrderingKey(key)` | 消息排序键 |

### Subscribe 选项

| 选项 | 说明 |
|------|------|
| `gcpubsub.WithSubscriptionName(name)` | 订阅名称（默认使用 topic 名称） |
| `gcpubsub.WithReceiveSettings(s)` | 接收配置（并发数、流控等） |

## Docker 部署开发环境

```shell
# 使用 Pub/Sub 模拟器
docker run -d \
    --name pubsub-emulator \
    -p 8085:8085 \
    gcr.io/google.com/cloudsdktool/google-cloud-cli:latest \
    gcloud beta emulators pubsub start --host-port=0.0.0.0:8085
```

## 注意事项

- 需要创建 GCP 项目并启用 Pub/Sub API
- 本地开发可使用 [Pub/Sub 模拟器](https://cloud.google.com/pubsub/docs/emulator)
- 消息最大 10MB，消息属性 key/value 均有长度限制
- 使用 Ordering Key 时必须启用消息排序（`EnableMessageOrdering`）
- 至少一次投递语义，业务侧需做幂等处理

## 参考资料

- [GCP Pub/Sub 官方文档](https://cloud.google.com/pubsub/docs)
- [Pub/Sub 客户端库 — Go](https://cloud.google.com/pubsub/docs/reference/libraries)
- [Pub/Sub 本地模拟器](https://cloud.google.com/pubsub/docs/emulator)
