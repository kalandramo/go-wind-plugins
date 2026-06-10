# Azure Service Bus

## 什么是 Azure Service Bus？

Microsoft Azure Service Bus 是一个企业级全托管消息代理，支持队列、主题/订阅模型，提供消息会话、事务、重复消息检测等高级特性。

## 核心概念

- **Queue**：点对点消息队列，一条消息被一个消费者消费。
- **Topic**：发布/订阅模型的消息路由器，一条消息可被多个订阅者消费。
- **Subscription**：Topic 下的虚拟队列，每个 Subscription 独立接收 Topic 消息。
- **Session**：消息会话，保证同一 Session ID 的消息按序投递。
- **Dead-letter Queue (DLQ)**：死信队列，存储无法投递或过期的消息。
- **Scheduled Messages**：定时消息，支持延迟投递。
- **Transaction**：事务支持，可将多个发送操作原子化。

## 队列 vs 主题

| 特性 | Queue | Topic + Subscription |
|------|-------|---------------------|
| 消费模型 | 竞争消费（一条消息只给一个消费者） | 发布/订阅（一条消息给所有订阅者） |
| 路由规则 | 无 | 支持过滤规则（SQL/Correlation） |
| 适用场景 | 任务队列、命令分发 | 事件通知、数据广播 |

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

## 注意事项

- 需要在 Azure Portal 创建 Service Bus 命名空间
- 消息最大 256KB（标准层）或 100MB（高级层）
- 使用 Session 的 Queue/Topic 需要在创建时启用 Session 支持
- `ReceiveMode` 默认为 `PeekLock`（需要手动 ACK），`ReceiveAndDelete` 模式消息自动删除

## 参考资料

- [Azure Service Bus 官方文档](https://learn.microsoft.com/azure/service-bus-messaging/)
- [Azure Service Bus Go SDK](https://github.com/Azure/azure-sdk-for-go/tree/main/sdk/messaging/azservicebus)
- [Service Bus 模拟器](https://learn.microsoft.com/azure/service-bus-messaging/test-locally-with-service-bus-emulator)
