# GCP Pub/Sub

## 什么是 GCP Pub/Sub？

Google Cloud Pub/Sub 是一种可扩展、灵活的消息传递服务，支持异步通信，可处理数百万条消息。适用于流式分析、事件驱动架构和无服务器应用。

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
