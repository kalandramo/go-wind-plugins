# Azure Service Bus

## 什么是 Azure Service Bus？

Microsoft Azure Service Bus 是一个企业级消息代理，支持队列、主题/订阅模型，提供消息会话、事务、重复消息检测等高级特性。

## 使用方式

### 基础：发布/订阅

```go
b := azuresb.NewBroker(
    broker.WithCodec("json"),
    azuresb.WithConnectionString(
        "Endpoint=sb://<namespace>.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=<key>",
    ),
)
b.Init()
b.Connect()
defer b.Disconnect()

// 发布到队列
b.Publish(ctx, "my-queue", broker.NewMessage(msg))

// 订阅队列
b.Subscribe("my-queue", handler, binder)

// 订阅主题（需指定 subscription 名称）
b.Subscribe("my-topic", handler, binder,
    azuresb.WithSubscriptionName("my-subscription"),
)
```

### 高级：消息属性

```go
b.Publish(ctx, "my-queue", broker.NewMessage(msg),
    azuresb.WithPublishContentType("application/json"),
    azuresb.WithPublishSessionID("session-1"),
    azuresb.WithPublishMessageID("msg-123"),
)
```

### 高级：资源管理

```go
// 在启动前确保 Queue/Topic/Subscription 存在
b.EnsureQueue(ctx, "my-queue", nil)
b.EnsureTopic(ctx, "my-topic", nil)
b.EnsureSubscription(ctx, "my-topic", "my-sub", nil)
```

## 配置选项

### Broker 选项

| 选项 | 说明 |
|------|------|
| `azuresb.WithConnectionString(connStr)` | Azure Service Bus 连接字符串（必填） |

### Publish 选项

| 选项 | 说明 |
|------|------|
| `azuresb.WithPublishContentType(ct)` | 消息 Content-Type |
| `azuresb.WithPublishSessionID(id)` | 会话 ID（需启用 Session 的 Queue/Topic） |
| `azuresb.WithPublishMessageID(id)` | 自定义消息 ID |

### Subscribe 选项

| 选项 | 说明 |
|------|------|
| `azuresb.WithSubscriptionName(name)` | 主题订阅名称（订阅 Topic 时必填） |
| `azuresb.WithReceiveMode(mode)` | 接收模式（`PeekLock` 或 `ReceiveAndDelete`） |

## Docker 部署开发环境

```shell
docker run -d \
    --name azuresb-emulator \
    -p 5672:5672 \
    -e ACCEPT_EULA=Y \
    mcr.microsoft.com/azure-messaging-servicebus-emulator:latest
```
