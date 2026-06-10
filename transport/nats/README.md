# NATS

基于 [nats-io/nats.go](https://github.com/nats-io/nats.go) 的 NATS 消息服务器，实现了 `transport.Server` 接口。支持核心 NATS 和 JetStream 两种模式。

## 核心特性

- 核心 NATS 和 JetStream 模式
- 泛型自动反序列化（`RegisterSubscriber[T]`）
- TLS 加密连接
- 自定义编解码
- 链路追踪（OpenTelemetry）


## 安装

```bash
go get github.com/tx7do/go-wind-plugins/transport/nats
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

    "github.com/tx7do/go-wind-plugins/transport/nats"
    "github.com/tx7do/go-wind-plugins/broker"
)

// MyMessage 示例消息。
type MyMessage struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

func main() {
    srv := nats.NewServer(
        nats.WithAddress([]string{"127.0.0.1"}),
        nats.WithCodec("json"),
    )

    // 注册订阅者（泛型，自动反序列化）
_ = nats.RegisterSubscriber(srv,
    context.Background(),
    "my-subject",
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
docker run -itd --name nats-server \
    -p 4222:4222 -p 8222:8222 \
    bitnami/nats:latest
```

管理后台：<http://localhost:8222>

## 配置选项

| 选项 | 类型 | 说明 |
|------|------|------|
| `WithAddress(addrs)` | []string | NATS 服务器地址列表 |
| `WithCodec(c)` | string | 编解码器名称（默认 json） |
| `WithTLSConfig(c)` | *tls.Config | TLS 配置 |
| `WithJetStream()` | - | 启用 JetStream 模式 |
| `WithBrokerOptions(opts)` | ...broker.Option | 直接传递 broker 选项 |


## 参考资料

- [NATS 官方文档](https://docs.nats.io/)
- [nats.go 客户端](https://github.com/nats-io/nats.go)

