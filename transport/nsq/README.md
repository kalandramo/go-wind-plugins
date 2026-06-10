# NSQ

基于 [nsqio/go-nsq](https://github.com/nsqio/go-nsq) 的 NSQ 消息服务器，实现了 `transport.Server` 接口。支持 nsqlookupd 服务发现。

## 核心特性

- Topic / Channel 消费模型
- nsqlookupd 服务发现
- 泛型自动反序列化（`RegisterSubscriber[T]`）
- TLS 加密连接
- 自定义编解码


## 安装

```bash
go get github.com/tx7do/go-wind-plugins/transport/nsq
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

    "github.com/tx7do/go-wind-plugins/transport/nsq"
    "github.com/tx7do/go-wind-plugins/broker"
)

// MyMessage 示例消息。
type MyMessage struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

func main() {
    srv := nsq.NewServer(
        nsq.WithAddress([]string{"127.0.0.1"}),
        nsq.WithCodec("json"),
    )

    // 注册订阅者（泛型，自动反序列化）
_ = nsq.RegisterSubscriber(srv,
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
# nsqlookupd
docker run -d --name nsqlookupd -p 4160:4160 -p 4161:4161 nsqio/nsq /nsqlookupd

# nsqd
docker run -itd --name nsqd -p 4150:4150 -p 4151:4151 --link nsqlookupd \
    nsqio/nsq /nsqd --lookupd-tcp-address=nsqlookupd:4160 --broadcast-address=host.docker.internal

# nsqadmin
docker run -itd --name nsqadmin -p 4171:4171 --link nsqlookupd \
    nsqio/nsq /nsqadmin --lookupd-http-address=nsqlookupd:4161
```

管理后台：<http://localhost:4171>

## 配置选项

| 选项 | 类型 | 说明 |
|------|------|------|
| `WithAddress(addrs)` | []string | NSQd 地址列表 |
| `WithLookupdAddress(addrs)` | []string | nsqlookupd 地址列表 |
| `WithCodec(c)` | string | 编解码器名称（默认 json） |
| `WithTLSConfig(c)` | *tls.Config | TLS 配置 |
| `WithConsumerOptions(opts)` | []string | NSQ 消费者选项 |
| `WithBrokerOptions(opts)` | ...broker.Option | 直接传递 broker 选项 |


## 参考资料

- [NSQ 官方文档](https://nsq.io/)
- [go-nsq 客户端](https://github.com/nsqio/go-nsq)

