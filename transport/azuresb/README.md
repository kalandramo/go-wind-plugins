# Azure Service Bus

基于 [Azure Service Bus Go SDK](https://github.com/Azure/azure-sdk-for-go) 的消息队列服务器，实现了 `transport.Server` 接口。支持 Queue 和 Topic/Subscription 两种消息模型。

## 核心特性

- Queue（点对点）和 Topic/Subscription（发布/订阅）两种模型
- 泛型自动反序列化（`RegisterSubscriber[T]`）
- TLS 加密连接
- 自定义编解码
- 链路追踪（OpenTelemetry）


## 安装

```bash
go get github.com/tx7do/go-wind-plugins/transport/azuresb
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

    "github.com/tx7do/go-wind-plugins/transport/azuresb"
    "github.com/tx7do/go-wind-plugins/broker"
)

// MyMessage 示例消息。
type MyMessage struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

func main() {
    srv := azuresb.NewServer(
        azuresb.WithAddress([]string{"127.0.0.1"}),
        azuresb.WithCodec("json"),
    )

    // 注册订阅者（泛型，自动反序列化）
_ = azuresb.RegisterSubscriber(srv,
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
docker run -itd --name servicebus-emulator \
    -p 5672:5672 -p 5300:5300 \
    -e ACCEPT_EULA=Y \
    mcr.microsoft.com/azure-messaging/servicebus-emulator:latest
```

使用 [Azure Service Bus Emulator](https://learn.microsoft.com/en-us/azure/service-bus-messaging/test-locally-with-service-bus-emulator) 进行本地开发。

## 配置选项

| 选项 | 类型 | 说明 |
|------|------|------|
| `WithConnectionString(connStr)` | string | Azure Service Bus 连接字符串 |
| `WithCodec(c)` | string | 编解码器名称（默认 json） |
| `WithTLSConfig(c)` | *tls.Config | TLS 配置 |
| `WithBrokerOptions(opts)` | ...broker.Option | 直接传递 broker 选项 |


## 参考资料

- [Azure Service Bus 文档](https://learn.microsoft.com/zh-cn/azure/service-bus-messaging/)
- [Azure Service Bus Emulator](https://learn.microsoft.com/en-us/azure/service-bus-messaging/test-locally-with-service-bus-emulator)

