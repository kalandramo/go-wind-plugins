# SignalR Server

SignalR 是一个面向实时 Web 应用的通信框架，简化了服务器向客户端推送内容的过程。它使得服务器端代码能够在数据可用时立即推送到连接的客户端，适用于聊天、仪表板、协作编辑、实时游戏等场景。

本模块封装了 [signalr](https://github.com/philippseith/signalr) 库，实现了 `go-wind/transport.Server` 接口，内置 CORS 中间件，支持 Hub 模式的 RPC 调用和流式传输，可以与 `go-wind` 应用框架无缝集成。

## 核心特性

- **Hub 模式**：通过 Hub 暴露 RPC 方法，客户端可远程调用服务器方法
- **流式传输**：支持 Server-to-Client 和 Client-to-Server 的流式数据
- **分组广播**：通过 Groups 将消息广播给指定分组的客户端
- **自动重连**：内置连接管理和心跳保活
- **自定义编解码**：基于 `encoding.Codec` 支持 JSON / Proto / MsgPack 等多种格式
- **CORS 支持**：内置 CORS 中间件
- **阻塞式生命周期**：`Start` 阻塞直到 context 取消，兼容 `go-wind` App

## 安装

```bash
go get github.com/tx7do/go-wind-plugins/transport/signalr
```

## 快速开始

```go
package main

import (
    "context"
    "fmt"

    signalr "github.com/tx7do/go-wind-plugins/transport/signalr"
    "github.com/philippseith/signalr"
)

// 定义 Hub
type ChatHub struct {
    signalr.Hub
}

func (h *ChatHub) OnConnected(connectionID string) {
    fmt.Printf("%s connected\n", connectionID)
    h.Groups().AddToGroup("lobby", connectionID)
}

func (h *ChatHub) OnDisconnected(connectionID string) {
    fmt.Printf("%s disconnected\n", connectionID)
    h.Groups().RemoveFromGroup("lobby", connectionID)
}

// Broadcast 是客户端可调用的 RPC 方法
func (h *ChatHub) Broadcast(message string) {
    h.Clients().Group("lobby").Send("receive", message)
}

// Echo 是客户端可调用的 RPC 方法，有返回值
func (h *ChatHub) Echo(message string) string {
    return "echo: " + message
}

func main() {
    srv := signalr.NewServer(
        signalr.WithAddress(":8100"),
        signalr.WithCodec("json"),
        signalr.WithHub(&ChatHub{}),
    )

    // 将 Hub 映射到 HTTP 路径
    srv.MapHTTP("/chat")

    // 启动服务器（阻塞）
    ctx := context.Background()
    if err := srv.Start(ctx); err != nil {
        panic(err)
    }
}
```

## 配置选项

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `WithAddress(addr)` | 监听地址 | `:0` |
| `WithNetwork(net)` | 网络类型 | `tcp` |
| `WithCodec(name)` | 编解码器名称 | `json` |
| `WithTLSConfig(cfg)` | TLS 配置 | - |
| `WithHub(hub)` | Hub 实例 | - |
| `WithKeepAliveInterval(d)` | 心跳间隔 | `2s` |
| `WithChanReceiveTimeout(d)` | 通道接收超时 | `200ms` |
| `WithStreamBufferCapacity(n)` | 流缓冲容量 | `5` |
| `WithDebug(enable)` | 调试模式 | `false` |

## Hub 开发

Hub 是 SignalR 的核心概念。通过在 Hub 上定义导出方法，客户端即可远程调用：

```go
type ChatHub struct {
    signalr.Hub
}

// 返回普通值
func (h *ChatHub) RequestTuple(msg string) (string, string, int) {
    return strings.ToUpper(msg), strings.ToLower(msg), len(msg)
}

// 返回 Channel（流式传输）
func (h *ChatHub) DateStream() <-chan string {
    r := make(chan string)
    go func() {
        defer close(r)
        for i := 0; i < 50; i++ {
            r <- fmt.Sprint(time.Now().Clock())
            time.Sleep(time.Second)
        }
    }()
    return r
}

// 接收客户端上传的流
func (h *ChatHub) UploadStream(upload1 <-chan int, factor float64, upload2 <-chan float64) {
    for u1 := range upload1 {
        h.Echo(fmt.Sprintf("u1: %v", u1))
    }
}
```

## 参考资料

- [Introduction to SignalR](https://learn.microsoft.com/en-us/aspnet/signalr/overview/getting-started/introduction-to-signalr)
- [go-signalr](https://github.com/philippseith/signalr)
- [SignalR vs. Socket.IO: which one is best for you?](https://ably.com/topic/signalr-vs-socketio)
