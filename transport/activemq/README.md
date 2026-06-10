# ActiveMQ

基于 [go-stomp](https://github.com/go-stomp/stomp) 的 ActiveMQ 消息队列服务器，实现了 `transport.Server` 接口。通过 STOMP 协议与 ActiveMQ 通信，支持发布/订阅和点对点消息模式。

## 核心特性

- 发布/订阅和点对点消息模式
- 泛型自动反序列化（`RegisterSubscriber[T]`）
- TLS 加密连接
- 自定义编解码（JSON / Proto / MsgPack 等）
- 链路追踪（OpenTelemetry）


## 安装

```bash
go get github.com/tx7do/go-wind-plugins/transport/activemq
```

## 快速开始

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/tx7do/go-wind-plugins/transport/activemq"
    "github.com/tx7do/go-wind-plugins/broker"
)

// MyMessage 示例消息。
type MyMessage struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

func main() {
    srv := activemq.NewServer(
        activemq.WithAddress([]string{"127.0.0.1"}),
        activemq.WithCodec("json"),
    )

    // 注册订阅者（泛型，自动反序列化）
_ = activemq.RegisterSubscriber(srv,
    context.Background(),
    "my-topic",
    func(ctx context.Context, topic string, headers broker.Headers, msg *MyMessage) error {
        log.Printf("received: %+v", msg)
        return nil
    },
)

    // 启动服务器（阻塞）
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    if err := srv.Start(ctx); err != nil {
        log.Fatal(err)
    }
}
```

## Docker 部署

```shell
docker run -d --name activemq \
    -p 61616:61616 -p 8161:8161 -p 61613:61613 \
    rmohr/activemq:latest
```

管理后台：<http://localhost:8161/admin/>（admin/admin）

| 端口  | 协议   |
|-------|--------|
| 61616 | JMS    |
| 8161  | Web UI |
| 61613 | STOMP  |

## 配置选项

| 选项 | 类型 | 说明 |
|------|------|------|
| `WithAddress(addrs)` | []string | STOMP 服务器地址 |
| `WithCodec(c)` | string | 编解码器名称（默认 json） |
| `WithTLSConfig(c)` | *tls.Config | TLS 配置 |
| `WithBrokerOptions(opts)` | ...broker.Option | 直接传递 broker 选项 |


## 参考资料

- [Apache ActiveMQ](https://activemq.apache.org/)
- [STOMP 协议](https://stomp.github.io/)
- [go-stomp](https://github.com/go-stomp/stomp)

