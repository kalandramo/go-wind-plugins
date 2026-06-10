# KCP Server

KCP 是一个基于 UDP 的快速可靠传输协议，相比 TCP 牺牲约 10%~20% 的带宽，但能显著降低延迟（平均延迟降低 30%~40%，最大延迟降低 3 倍），非常适合实时对战游戏、音视频通话等对延迟敏感的场景。

本模块提供了一个基于 KCP 的 Socket 服务器，实现了 `go-wind/transport.Server` 接口，支持会话管理、消息类型路由、BlockCrypt 加密和广播，可以与 `go-wind` 应用框架无缝集成。

## 核心特性

- **低延迟传输**：基于 UDP + KCP 协议，快速重传和选择性重传
- **前向纠错（FEC）**：支持 Reed-Solomon 编码的数据冗余，抵抗丢包
- **BlockCrypt 加密**：基于 AES + PBKDF2 的对称加密
- **会话管理**：每个客户端连接对应一个 Session，包含唯一的 SessionID
- **消息类型路由**：通过 `RegisterMessageHandler` 注册不同消息类型的处理器
- **自定义编解码**：基于 `encoding.Codec` 支持 JSON / Proto / MsgPack 等多种格式
- **广播 / 定向发送**：支持向所有客户端广播或向指定 Session 发送消息
- **阻塞式生命周期**：`Start` 阻塞直到 context 取消，兼容 `go-wind` App

## 安装

```bash
go get github.com/tx7do/go-wind-plugins/transport/kcp
```

## 快速开始

```go
package main

import (
    "context"
    "log"

    kcp "github.com/tx7do/go-wind-plugins/transport/kcp"
)

type ChatMessage struct {
    Sender  string `json:"sender"`
    Message string `json:"message"`
}

func main() {
    srv := kcp.NewServer(
        kcp.WithAddress(":9090"),
        kcp.WithCodec("json"),
        kcp.WithBlockCrypt(kcp.DefaultBlockCryptPassword, kcp.DefaultBlockCryptSalt),
    )

    // 注册消息处理器
    kcp.RegisterServerMessageHandler(srv, 1, func(sid kcp.SessionID, msg *ChatMessage) error {
        log.Printf("[%s] %s: %s", sid, msg.Sender, msg.Message)
        srv.Broadcast(1, *msg)
        return nil
    })

    // 启动服务器（阻塞）
    ctx := context.Background()
    if err := srv.Start(ctx); err != nil {
        panic(err)
    }
}
```

## 配置选项

### 服务器

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `WithAddress(addr)` | 监听地址 | `:0` |
| `WithTimeout(d)` | 读写超时 | `1s` |
| `WithCodec(name)` | 编解码器名称 | `json` |
| `WithBlockCrypt(pwd, salt)` | 加密密码和盐 | `go-wind-kcp-password` / `go-wind-kcp-salt` |
| `WithDataShards(n)` | FEC 数据分片数 | `10` |
| `WithParityShards(n)` | FEC 校验分片数 | `3` |
| `WithSocketConnectHandler(fn)` | 连接 / 断开回调 | - |
| `WithSocketRawDataHandler(fn)` | 原始数据处理回调 | - |
| `WithMessageMarshaler(fn)` | 自定义封包函数 | - |
| `WithMessageUnmarshaler(fn)` | 自定义拆包函数 | - |

### 客户端

```go
cli := kcp.NewClient(
    kcp.WithEndpoint("127.0.0.1:9090"),
    kcp.WithClientCodec("json"),
    kcp.WithClientBlockCrypt(kcp.DefaultBlockCryptPassword, kcp.DefaultBlockCryptSalt),
)
defer cli.Disconnect()

kcp.RegisterClientMessageHandler(cli, 1, func(msg *ChatMessage) error {
    log.Printf("received: %+v", msg)
    return nil
})

if err := cli.Connect(); err != nil {
    panic(err)
}

_ = cli.SendMessage(1, &ChatMessage{Sender: "alice", Message: "hello"})
```

## 协议格式

KCP 模块与 TCP 模块使用相同的 `NetPacket` 应用层协议包格式：

```
+-------------------+-------------------+
| Type (uint32, 4B) | Payload (变长)     |
+-------------------+-------------------+
```

- `Type`：消息类型标识，用于路由到对应的 handler
- `Payload`：消息体，由 `encoding.Codec` 进行序列化/反序列化

字节序默认为小端序，可通过 `WithBigEndian()` 切换为大端序。

## 参考资料

- [KCP - A Fast and Reliable ARQ Protocol](https://github.com/skywind3000/kcp)
- [kcp-go - Golang Implementation of KCP](https://github.com/xtaci/kcp-go)
