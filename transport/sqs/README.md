# Amazon SQS

基于 [AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/) 的 Amazon SQS 消息服务器，实现了 `transport.Server` 接口。支持标准队列和 FIFO 队列。

## 核心特性

- 标准队列和 FIFO 队列
- 泛型自动反序列化（`RegisterSubscriber[T]`）
- 自定义端点（本地模拟器）
- TLS 加密连接
- 自定义编解码
- 链路追踪（OpenTelemetry）


## 安装

```bash
go get github.com/tx7do/go-wind-plugins/transport/sqs
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

    "github.com/tx7do/go-wind-plugins/transport/sqs"
    "github.com/tx7do/go-wind-plugins/broker"
)

// MyMessage 示例消息。
type MyMessage struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

func main() {
    srv := sqs.NewServer(
        sqs.WithAddress([]string{"127.0.0.1"}),
        sqs.WithCodec("json"),
    )

    // 注册订阅者（泛型，自动反序列化）
_ = sqs.RegisterSubscriber(srv,
    "my-queue",
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
docker run -itd --name elasticmq \
    -p 9324:9324 -p 9325:9325 \
    softwaremill/elasticmq:latest
```

SQS 兼容端点：<http://localhost:9324>
管理界面：<http://localhost:9325>

或使用 [LocalStack](https://localstack.cloud/)：

```shell
docker run -itd --name localstack -p 4566:4566 -e SERVICES=sqs localstack/localstack:latest
```

## 配置选项

| 选项 | 类型 | 说明 |
|------|------|------|
| `WithRegion(region)` | string | AWS 区域 |
| `WithEndpoint(url)` | string | 自定义端点（ElasticMQ / LocalStack） |
| `WithQueueUrl(url)` | string | 队列 URL |
| `WithCodec(c)` | string | 编解码器名称（默认 json） |
| `WithTLSConfig(c)` | *tls.Config | TLS 配置 |
| `WithBrokerOptions(opts)` | ...broker.Option | 直接传递 broker 选项 |


## 参考资料

- [Amazon SQS 文档](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/welcome.html)
- [ElasticMQ](https://github.com/softwaremill/elasticmq)
- [AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/)

