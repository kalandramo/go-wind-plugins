# MQTT

基于 [eclipse/paho.mqtt.golang](https://github.com/eclipse/paho.mqtt.golang) 的 MQTT 消息服务器，实现了 `transport.Server` 接口。适用于物联网（IoT）和实时消息推送场景。

## 核心特性

- 发布/订阅模式
- 泛型自动反序列化（`RegisterSubscriber[T]`）
- 用户名/密码认证
- Clean Session 控制
- 客户端 ID 自定义
- 连接/断开回调
- 自定义编解码


## 安装

```bash
go get github.com/tx7do/go-wind-plugins/transport/mqtt
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

    "github.com/tx7do/go-wind-plugins/transport/mqtt"
    "github.com/tx7do/go-wind-plugins/broker"
)

// MyMessage 示例消息。
type MyMessage struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

func main() {
    srv := mqtt.NewServer(
        mqtt.WithAddress([]string{"127.0.0.1"}),
        mqtt.WithCodec("json"),
    )

    // 注册订阅者（泛型，自动反序列化）
_ = mqtt.RegisterSubscriber(srv,
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
docker run -itd --name mosquitto \
    -p 1883:1883 -p 9001:9001 \
    eclipse-mosquitto:latest
```

其他可用 Broker：[EMQX](https://www.emqx.io/)（端口 1883/18083）、[HiveMQ](https://www.hivemq.com/)（端口 1883/8000）

## 配置选项

| 选项 | 类型 | 说明 |
|------|------|------|
| `WithAddress(addrs)` | []string | MQTT Broker 地址列表 |
| `WithCodec(c)` | string | 编解码器名称（默认 json） |
| `WithAuth(user, pass)` | string, string | 用户名/密码认证 |
| `WithClientId(id)` | string | 客户端 ID |
| `WithCleanSession(enable)` | bool | Clean Session 标志 |
| `WithOnConnect(cb)` | func() | 连接成功回调 |
| `WithOnDisconnect(cb)` | func(error) | 断开连接回调 |
| `WithBrokerOptions(opts)` | ...broker.Option | 直接传递 broker 选项 |


## 参考资料

- [MQTT 协议](https://mqtt.org/)
- [Eclipse Paho MQTT Go](https://github.com/eclipse/paho.mqtt.golang)
- [Mosquitto](https://mosquitto.org/)

