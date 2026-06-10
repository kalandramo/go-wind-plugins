# GCP Pub/Sub

基于 [Google Cloud Pub/Sub Go SDK](https://cloud.google.com/pubsub/docs/reference/libraries) 的消息队列服务器，实现了 `transport.Server` 接口。使用 Pull 订阅模式消费消息。

## 核心特性

- Pull 订阅模式
- 泛型自动反序列化（`RegisterSubscriber[T]`）
- 服务账号认证（Credentials File）
- 本地 Emulator 支持
- 自定义编解码
- 链路追踪（OpenTelemetry）


## 安装

```bash
go get github.com/tx7do/go-wind-plugins/transport/gcpubsub
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

    "github.com/tx7do/go-wind-plugins/transport/gcpubsub"
    "github.com/tx7do/go-wind-plugins/broker"
)

// MyMessage 示例消息。
type MyMessage struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

func main() {
    srv := gcpubsub.NewServer(
        gcpubsub.WithAddress([]string{"127.0.0.1"}),
        gcpubsub.WithCodec("json"),
    )

    // 注册订阅者（泛型，自动反序列化）
_ = gcpubsub.RegisterSubscriber(srv,
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
docker run -itd --name pubsub-emulator -p 8085:8085 \
    google/cloud-sdk:latest \
    gcloud beta emulators pubsub start --project=test-project --host-port=0.0.0.0:8085
```

使用 [Pub/Sub Emulator](https://cloud.google.com/pubsub/docs/emulator) 进行本地开发。

## 配置选项

| 选项 | 类型 | 说明 |
|------|------|------|
| `WithProjectID(id)` | string | GCP 项目 ID |
| `WithCredentialsFile(path)` | string | 服务账号 JSON 凭据文件路径 |
| `WithEndpoint(endpoint)` | string | 自定义端点（用于本地模拟器） |
| `WithCodec(c)` | string | 编解码器名称（默认 json） |
| `WithBrokerOptions(opts)` | ...broker.Option | 直接传递 broker 选项 |


## 参考资料

- [Google Cloud Pub/Sub 文档](https://cloud.google.com/pubsub/docs)
- [Pub/Sub Emulator](https://cloud.google.com/pubsub/docs/emulator)

