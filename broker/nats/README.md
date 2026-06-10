# NATS

[NATS](https://nats.io/) 是一个高性能、轻量级的云原生分布式消息系统，最初由 Apcera（CloudFoundry 架构师 Derek Collison）使用 Go 语言开发。

NATS 以其 **极致性能**（单节点可达 1800 万 msg/s）、**极小体积**（Docker 镜像仅 3MB）和 **简洁协议**（基于文本的 Publish/Subscribe 协议）著称，
广泛用于云基础设施通信、IoT 设备消息、微服务架构等场景。

## 两种模式

本模块同时支持 **Core NATS** 和 **NATS JetStream** 两种运行模式：

| 特性 | Core NATS | JetStream |
|------|-----------|-----------|
| 消息模型 | Publish/Subscribe、Request/Reply、Queue | 同 Core NATS |
| 持久化 | 不支持（发送后不管） | 支持（Stream 存储） |
| 消息确认 | 无 | ACK / NAK / Term / InProgress |
| 至少一次投递 | 不支持 | 支持 |
| 消息去重 | 不支持 | `MsgId` 去重 |
| 消息回放 | 不支持 | DeliverAll / DeliverLast / DeliverNew |
| 消费者类型 | 无 | Push / Pull、Durable |
| 流管理 | 无 | AddStream / DeleteStream |
| Broker 构造 | `NewBroker()` | `NewJetStreamBroker()` |

## 核心概念映射

### Core NATS

| Broker 概念 | NATS 概念 | 说明 |
|------------|----------|------|
| Topic | Subject | 消息主题，支持 `.` 分隔和 `*` / `>` 通配符 |
| Publish | `conn.Publish()` | 即发即弃 |
| Subscribe | `conn.Subscribe()` | 内存推送，无持久化 |
| Request | `conn.Request()` | 请求-响应模式 |

### JetStream

| Broker 概念 | JetStream 概念 | 说明 |
|------------|---------------|------|
| Topic | Stream Subject | 消息路由目标，Stream 绑定多个 Subject |
| Publish | `js.PublishMsg()` | 持久化存储，支持去重和预期验证 |
| Subscribe | Push / Pull Consumer | Consumer 从 Stream 消费消息 |
| Message.Ack() | `msg.Ack()` | 消费者确认消息 |
| — | Stream | 持久化消息存储 |

## Docker 部署开发环境

Core NATS（无持久化）：

```shell
docker run -d --name nats-server \
    -p 4222:4222 \
    -p 8222:8222 \
    nats:latest
```

JetStream（带持久化，需启用 JS）：

```shell
docker run -d --name nats-js \
    -p 4222:4222 \
    -p 8222:8222 \
    nats:latest -js
```

- 客户端端口：`4222`
- 监控面板：`http://localhost:8222`

## 使用方式 — Core NATS

### 基础：发布/订阅

```go
b := nats.NewBroker(
    broker.WithAddress("nats://127.0.0.1:4222"),
)
b.Init()
b.Connect()

// 发布
b.Publish(ctx, "test.subject", broker.NewMessage([]byte(`{"hello":"world"}`)))

// 订阅
_, err := b.Subscribe("test.subject", handler, binder)
```

### 基础：请求-响应

```go
// 发送请求（阻塞等待响应）
reply, err := b.Request(ctx, "test.subject", broker.NewMessage(msg),
    nats.WithRequestTimeout(2*time.Second),
)
```

### 基础：队列组

```go
// 同一队列组的订阅者中只有一个会收到消息
_, err := b.Subscribe("test.subject", handler, binder,
    broker.WithSubscribeQueueName("order-group"),
)
```

## 使用方式 — JetStream

### 基础：创建 Stream + 发布 + 订阅

```go
b := nats.NewJetStreamBroker(
    broker.WithAddress("nats://127.0.0.1:4222"),
    broker.WithCodec("json"),
)
b.Init()
b.Connect()

// 创建 Stream
js := nats.GetJetStreamContext(b)
js.AddStream(&natsGo.StreamConfig{
    Name:     "ORDERS",
    Subjects: []string{"orders.*"},
})

// 发布到 JetStream（消息被持久化存储）
b.Publish(ctx, "orders.create", broker.NewMessage(orderData))

// 订阅（Push 模式，自动 Ack）
_, err := b.Subscribe("orders.*", handler, binder,
    nats.WithDurable("order-consumer"),
    nats.WithDeliverNew(),
)
```

### 高级：消息去重

```go
err := b.Publish(ctx, "orders.create", broker.NewMessage(orderData),
    nats.WithMsgId("order-12345"),              // 幂等去重
    nats.WithExpectStream("ORDERS"),            // 预期目标 Stream
    nats.WithExpectLastSequence(42),            // 乐观并发控制
)
```

### 高级：Pull 订阅

```go
// Pull 模式：消费者主动拉取消息，适合流控场景
_, err := b.Subscribe("orders.*", handler, binder,
    nats.WithPullSubscribe(),
    nats.WithPullBatchSize(10),
    nats.WithDurable("pull-consumer"),
    nats.WithDeliverAll(),
)
```

### 高级：手动确认 + NAK

```go
// 禁用自动 Ack，在 handler 中手动确认或拒绝
_, err := b.Subscribe("orders.*", handler, binder,
    nats.WithManualAck(),
    nats.WithSubscribeAckWait(30*time.Second),
    nats.WithSubscribeMaxAckPending(100),
)

// handler 中：
func handler(ctx context.Context, event broker.Event) error {
    msg, ok := nats.JetStreamMsgFromEvent(event)
    if !ok {
        return nil
    }

    // 处理失败 → NAK（触发重投）
    if err := process(event); err != nil {
        _ = msg.Nak()
        return err
    }

    // 处理成功 → ACK
    return msg.Ack()
}
```

### 高级：流管理

```go
js := nats.GetJetStreamContext(b)

// 创建 Stream
js.AddStream(&natsGo.StreamConfig{
    Name:     "ORDERS",
    Subjects: []string{"orders.*"},
    Retention: natsGo.LimitsPolicy,
    MaxMsgs:  10000,
})

// 更新 Stream
js.UpdateStream(&natsGo.StreamConfig{
    Name:     "ORDERS",
    Subjects: []string{"orders.*", "returns.*"},
})

// 删除 Stream
js.DeleteStream("ORDERS")

// 查看消费者信息
info, _ := js.ConsumerInfo("ORDERS", "order-consumer")
```

### 高级：消息高级操作

```go
msg, ok := nats.JetStreamMsgFromEvent(event)
if ok {
    msg.InProgress()  // 标记处理中（延长 Ack 超时）
    msg.Nak()         // 否定确认（触发重投）
    msg.Term()        // 终止消息（不再重投）
    msg.Ack()         // 确认消息
}
```

## 支持的 JetStream 选项

### Broker 选项

| 选项 | 说明 |
|------|------|
| `JetStreamContextOptions(opts ...natsGo.JSOpt)` | JetStream 上下文选项 |

### Publish 选项

| 选项 | 说明 |
|------|------|
| `WithMsgId(id)` | 消息 ID（去重） |
| `WithExpectStream(stream)` | 预期目标 Stream |
| `WithExpectLastSequence(seq)` | 预期最后序列号 |
| `WithExpectLastSequencePerSubject(seq)` | 预期 Subject 最后序列号 |
| `WithExpectLastMsgId(id)` | 预期最后消息 ID |
| `WithPublishRawOpts(opts ...)` | 传递原生 PubOpt |

### Subscribe 选项

| 选项 | 说明 |
|------|------|
| `WithDurable(name)` | Durable Consumer 名称 |
| `WithDeliverAll()` | 从头投递所有消息 |
| `WithDeliverLast()` | 仅投递最后一条消息 |
| `WithDeliverNew()` | 仅投递新消息 |
| `WithStartSequence(seq)` | 从指定序列号开始 |
| `WithStartTime(t)` | 从指定时间开始 |
| `WithSubscribeAckWait(d)` | ACK 等待超时 |
| `WithSubscribeMaxAckPending(n)` | 最大未确认消息数 |
| `WithBindStream(stream)` | 绑定到指定 Stream |
| `WithReplayInstant()` | 快速回放模式 |
| `WithManualAck()` | 手动确认模式 |
| `WithPullSubscribe()` | Pull 模式 |
| `WithPullBatchSize(n)` | Pull 批量大小 |
| `WithSubscribeRawOpts(opts ...)` | 传递原生 SubOpt |

## 工具函数

| 函数 | 说明 |
|------|------|
| `GetJetStreamContext(b)` | 获取底层 `JetStreamContext`（流管理） |
| `JetStreamMsgFromEvent(evt)` | 从 Event 提取底层 `*natsGo.Msg`（NAK/Term/InProgress） |
| `GetConn(b)` | 获取底层 `*natsGo.Conn` |
