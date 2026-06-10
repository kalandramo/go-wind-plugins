# Socket.IO Server

Socket.IO 是一个面向实时 Web 应用的双向通信库。它主要使用 WebSocket 协议，同时支持 HTTP 长轮询作为降级方案，使得服务器和客户端之间可以实时双向通信。

本模块封装了 [go-socket.io](https://github.com/googollee/go-socket.io) 库，实现了 `go-wind/transport.Server` 接口，内置 CORS 中间件和 gorilla/mux 路由，可以与 `go-wind` 应用框架无缝集成。

## 核心特性

- **实时双向通信**：基于 WebSocket，自动降级到长轮询
- **命名空间与事件**：支持 namespace / event 级别的消息路由
- **CORS 支持**：内置 CORS 中间件
- **自定义编解码**：基于 `encoding.Codec` 支持 JSON / Proto / MsgPack 等多种格式
- **阻塞式生命周期**：`Start` 阻塞直到 context 取消，兼容 `go-wind` App

## 安装

```bash
go get github.com/tx7do/go-wind-plugins/transport/socketio
```

## 快速开始

```go
package main

import (
    "context"
    "fmt"
    "log"

    socketio "github.com/tx7do/go-wind-plugins/transport/socketio"
    sio "github.com/googollee/go-socket.io"
)

func main() {
    srv := socketio.NewServer(
        socketio.WithAddress(":8000"),
        socketio.WithCodec("json"),
        socketio.WithPath("/socket.io/"),
    )

    // 连接事件
    srv.RegisterConnectHandler("/", func(s sio.Conn) error {
        s.SetContext("")
        fmt.Println("connected:", s.ID())
        return nil
    })

    // 消息事件
    srv.RegisterEventHandler("/", "notice", func(s sio.Conn, msg string) {
        fmt.Println("notice:", msg)
        s.Emit("reply", "have "+msg)
    })

    // 断开事件
    srv.RegisterDisconnectHandler("/", func(s sio.Conn, reason string) {
        fmt.Println("closed:", reason)
    })

    // 错误事件
    srv.RegisterErrorHandler("/", func(s sio.Conn, e error) {
        log.Printf("error: %v", e)
    })

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
| `WithPath(path)` | Socket.IO 路由路径 | `/socket.io/` |
| `WithTLSConfig(cfg)` | TLS 配置 | - |
| `WithConnectHandler(ns, fn)` | 连接事件处理器 | - |
| `WithDisconnectHandler(ns, fn)` | 断开事件处理器 | - |
| `WithErrorHandler(ns, fn)` | 错误事件处理器 | - |
| `WithEventHandler(ns, event, fn)` | 自定义事件处理器 | - |

## 事件注册

Socket.IO 通过命名空间（namespace）和事件名（event）来路由消息：

```go
// 命名空间 "/" 下的 "chat" 事件
srv.RegisterEventHandler("/", "chat", func(s sio.Conn, msg string) string {
    s.SetContext(msg)
    return "recv " + msg
})

// 命名空间 "/room" 下的 "join" 事件
srv.RegisterEventHandler("/room", "join", func(s sio.Conn, room string) {
    // 处理加入房间逻辑
})
```

## 参考资料

- [Socket.IO 官方文档](https://socket.io/zh-CN/docs/v4/)
- [go-socket.io](https://github.com/googollee/go-socket.io)
- [gorilla/mux](https://github.com/gorilla/mux)
