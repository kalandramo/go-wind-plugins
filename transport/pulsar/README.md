# Apache Pulsar

基于 [apache/pulsar-client-go](https://github.com/apache/pulsar-client-go) 的 Pulsar 消息服务器，实现了 `transport.Server` 接口。支持多种订阅模式（独占、共享、故障转移）。

## 核心特性

- 多种订阅模式（Exclusive / Shared / Failover）
- 泛型自动反序列化（`RegisterSubscriber[T]`）
- TLS 加密连接
- 自定义编解码
- 链路追踪（OpenTelemetry）


## 安装

```bash
go get github.com/tx7do/go-wind-plugins/transport/pulsar
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

    "github.com/tx7do/go-wind-plugins/transport/pulsar"
    "github.com/tx7do/go-wind-plugins/broker"
)

// MyMessage 示例消息。
type MyMessage struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

func main() {
    srv := pulsar.NewServer(
        pulsar.WithAddress([]string{"127.0.0.1"}),
        pulsar.WithCodec("json"),
    )

    // 注册订阅者（泛型，自动反序列化）
_ = pulsar.RegisterSubscriber(srv,
    context.Background(),
    "my-topic", "my-subscription",
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
docker run -itd --name pulsar-standalone \
    -p 6650:6650 -p 8080:8080 \
    apachepulsar/pulsar:latest bin/pulsar standalone
```

管理后台（Pulsar Manager）：

```shell
docker run -itd --name pulsar-manager -p 9527:9527 -p 7750:7750 \
    apachepulsar/pulsar-manager:latest
```

访问 <http://localhost:9527>，默认账号 admin / apachepulsar

## 配置选项

| 选项 | 类型 | 说明 |
|------|------|------|
| `WithAddress(addrs)` | []string | Pulsar Broker 地址列表 |
| `WithCodec(c)` | string | 编解码器名称（默认 json） |
| `WithTLSConfig(c)` | *tls.Config | TLS 配置 |
| `WithBrokerOptions(opts)` | ...broker.Option | 直接传递 broker 选项 |


## 参考资料

- [Apache Pulsar 文档](https://pulsar.apache.org/docs/)
- [pulsar-client-go](https://github.com/apache/pulsar-client-go)

