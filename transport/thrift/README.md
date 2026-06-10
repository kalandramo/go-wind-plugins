# Thrift Transport

基于 [Apache Thrift](https://thrift.apache.org/) 的 RPC 服务器，实现了 `go-wind/transport.Server` 接口。支持多种协议（binary / compact / json）和传输模式（buffered / framed），可以与 `go-wind` 应用框架无缝集成。

## 核心特性

- **多协议支持**：binary（默认）、compact、json、simplejson
- **多传输模式**：plain、buffered、framed、buffered+framed
- **TLS 加密**：支持 SSL/TLS 加密传输
- **Processor 注册**：通过 thrift IDL 编译器生成的 Processor 直接注册
- **阻塞式生命周期**：`Start` 阻塞直到 context 取消，兼容 `go-wind` App

## 安装

```bash
go get github.com/tx7do/go-wind-plugins/transport/thrift
```

## 准备工作

### 安装 Thrift 编译器

```bash
# Linux
sudo apt install thrift-compiler

# macOS
brew install thrift

# Windows
# 从 https://thrift.apache.org/download 下载编译器，放入 PATH 目录
```

### 定义 IDL 并生成代码

编写 `.thrift` 文件：

```thrift
namespace go api

struct Hygrothermograph {
  1: optional double Humidity,
  2: optional double Temperature,
}

service HygrothermographService {
  Hygrothermograph getHygrothermograph()
}
```

生成 Go 代码：

```bash
thrift -r -gen go hygrothermograph.thrift
```

## 快速开始

### 服务端

```go
package main

import (
    "context"
    "log"
    "math/rand"

    "github.com/apache/thrift/lib/go/thrift"
    thriftServer "github.com/tx7do/go-wind-plugins/transport/thrift"
    api "your-module/gen-go/hygrothermograph"
)

// 实现 IDL 中定义的 Service Handler
type HygrothermographHandler struct{}

func (h *HygrothermographHandler) GetHygrothermograph(ctx context.Context) (*api.Hygrothermograph, error) {
    humidity := float64(rand.Intn(100))
    temperature := float64(rand.Intn(100))
    return &api.Hygrothermograph{
        Humidity:    &humidity,
        Temperature: &temperature,
    }, nil
}

func main() {
    // 创建 Processor
    processor := api.NewHygrothermographServiceProcessor(&HygrothermographHandler{})

    // 创建服务器
    srv := thriftServer.NewServer(":7700",
        thriftServer.WithProcessor(processor),
    )

    // 启动服务器（阻塞）
    ctx := context.Background()
    if err := srv.Start(ctx); err != nil {
        panic(err)
    }
}
```

### 客户端

客户端使用 Apache Thrift 原生客户端连接：

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/apache/thrift/lib/go/thrift"
    api "your-module/gen-go/hygrothermograph"
)

func main() {
    // 创建 Socket 传输
    trans, err := thrift.NewTSocket("localhost:7700")
    if err != nil {
        log.Fatal(err)
    }
    defer trans.Close()

    // 创建协议
    protoFactory := thrift.NewTBinaryProtocolFactoryConf(nil)
    _, err = thrift.NewTStandardClient(protoFactory.GetProtocol(trans), protoFactory.GetProtocol(trans))
    if err != nil {
        log.Fatal(err)
    }

    // 创建 Thrift 客户端
    client := api.NewHygrothermographServiceClientFactory(trans, protoFactory)

    // 调用 RPC
    reply, err := client.GetHygrothermograph(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Humidity: %f, Temperature: %f\n", *reply.Humidity, *reply.Temperature)
}
```

## 配置选项

| 选项                                            | 说明                                                | 默认值                  |
|-----------------------------------------------|---------------------------------------------------|----------------------|
| `WithProcessor(p)`                            | 设置 Thrift Processor（**必需**）                       | -                    |
| `WithProtocol(proto)`                         | 协议类型：`binary` / `compact` / `json` / `simplejson` | `binary`             |
| `WithTransportConfig(buffered, framed, size)` | 传输层配置                                             | `false, false, 8192` |
| `WithTLSConfig(cfg)`                          | TLS 加密配置                                          | -                    |

## 协议与传输模式

### 支持的协议

| 协议           | 说明               |
|--------------|------------------|
| `binary`     | 二进制格式（默认，最高性能）   |
| `compact`    | 压缩格式（更小的数据体积）    |
| `json`       | JSON 格式（跨语言调试友好） |
| `simplejson` | 只写 JSON 协议       |

### 传输模式组合

```go
// 默认：普通传输
srv := thriftServer.NewServer(":7700",
    thriftServer.WithProcessor(processor),
)

// 缓冲传输（减少系统调用，提升吞吐）
srv := thriftServer.NewServer(":7700",
    thriftServer.WithProcessor(processor),
    thriftServer.WithTransportConfig(true, false, 8192),
)

// 帧传输（非阻塞服务必需）
srv := thriftServer.NewServer(":7700",
    thriftServer.WithProcessor(processor),
    thriftServer.WithTransportConfig(false, true, 0),
)

// 使用 compact 协议
srv := thriftServer.NewServer(":7700",
    thriftServer.WithProcessor(processor),
    thriftServer.WithProtocol("compact"),
)

// TLS 加密传输
srv := thriftServer.NewServer(":7700",
    thriftServer.WithProcessor(processor),
    thriftServer.WithTLSConfig(tlsConfig),
)
```

## Thrift IDL 数据类型速查

| 类型                                                            | 说明                 |
|---------------------------------------------------------------|--------------------|
| `bool` / `byte` / `i16` / `i32` / `i64` / `double` / `string` | 基本类型               |
| `list<T>`                                                     | 有序列表（可重复）          |
| `set<T>`                                                      | 无序集合（不可重复）         |
| `map<K,V>`                                                    | 字典                 |
| `struct`                                                      | 结构体（类似 C struct）   |
| `enum`                                                        | 枚举（32 位整数）         |
| `union`                                                       | 联合类型（同一时间只有一个字段有效） |
| `exception`                                                   | 异常（映射到语言原生异常）      |
| `service`                                                     | 服务接口定义             |

## 参考资料

- [Apache Thrift 官网](https://thrift.apache.org/)
- [Thrift 各平台安装方法](https://thrift.apache.org/docs/install/)
- [Thrift IDL 规范](https://thrift.apache.org/docs/idl)
