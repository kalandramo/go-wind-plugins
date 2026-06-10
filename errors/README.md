# Errors

协议无关的错误类型定义，为 `go-wind` 框架的所有传输层（HTTP / gRPC / GraphQL / Thrift 等）提供统一的错误模型。

与从 `.proto` 文件生成错误的方式不同，本包使用纯 Go 代码定义错误，无需 protobuf 工具链，可直接在所有传输协议中使用。

## 核心设计

`Error` 以 **HTTP 状态码** 作为语义锚点，各传输层将其映射为各自的表示形式：

- **HTTP**：直接作为 HTTP 状态码
- **gRPC**：通过 `transport/grpc/middleware/errors` 映射为 `codes.Code`

## 安装

```bash
go get github.com/tx7do/go-wind-plugins/errors
```

## 快速开始

### 定义领域错误

```go
package service

import (
    "net/http"

    errs "github.com/tx7do/go-wind-plugins/errors"
)

// 定义业务错误（包级变量）
var (
    ErrUserNotFound  = errs.New(http.StatusNotFound, "USER_NOT_FOUND", "user not found")
    ErrUnauthorized  = errs.New(http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
    ErrInvalidInput  = errs.Newf(http.StatusBadRequest, "INVALID_INPUT", "field %s is required", "email")
    ErrRateLimited   = errs.New(http.StatusTooManyRequests, "RATE_LIMITED", "too many requests")
)
```

### 在业务逻辑中使用

```go
func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    user, err := s.repo.Find(id)
    if err != nil {
        return nil, ErrUserNotFound
    }
    return user, nil
}
```

### 添加元数据

```go
// WithMetadata 返回浅拷贝，原始错误不会被修改
err := ErrInvalidInput.WithMetadata(map[string]string{
    "field": "email",
    "doc":   "https://docs.example.com/errors/INVALID_INPUT",
})
```

### 解析错误

```go
// 从标准 error 提取 *Error（支持 %w 包装）
wrapped := fmt.Errorf("service layer: %w", ErrUserNotFound)
e := errs.FromError(wrapped)
// e.Code == 404, e.Reason == "USER_NOT_FOUND"

// 获取 HTTP 状态码
code := errs.Code(wrapped) // 404
code := errs.Code(nil)     // 200（无错误）
code := errs.Code(io.EOF)  // 500（未知错误，默认 500）
```

## API 参考

### 创建错误

| 函数 | 说明 |
|------|------|
| `New(code int, reason, message string) *Error` | 创建错误 |
| `Newf(code int, reason, format string, args... any) *Error` | 创建格式化消息的错误 |

### 解析错误

| 函数 | 说明 |
|------|------|
| `FromError(err error) *Error` | 从 `error` 提取 `*Error`（支持 `%w` 包装） |
| `Code(err error) int` | 获取 HTTP 状态码（`nil` → 200，非 `*Error` → 500） |

### Error 方法

| 方法 | 说明 |
|------|------|
| `Error() string` | 实现 `error` 接口，返回 `"REASON: message"` |
| `WithMetadata(kv map[string]string) *Error` | 返回添加元数据的浅拷贝 |

## Error 结构体

```go
type Error struct {
    Code     int               // HTTP 状态码（语义锚点）
    Reason   string            // 机器可读的稳定标识符，如 "USER_NOT_FOUND"
    Message  string            // 人类可读的错误描述
    Metadata map[string]string // 可选的键值对上下文（trace ID、字段详情等）
}
```

## HTTP 状态码常量

包内预定义了常用状态码常量，无需引入 `net/http`：

| 常量 | 值 |
|------|----|
| `StatusOK` | 200 |
| `StatusBadRequest` | 400 |
| `StatusUnauthorized` | 401 |
| `StatusForbidden` | 403 |
| `StatusNotFound` | 404 |
| `StatusConflict` | 409 |
| `StatusPreconditionFailed` | 412 |
| `StatusUnprocessableEntity` | 422 |
| `StatusTooManyRequests` | 429 |
| `StatusInternalServerError` | 500 |
| `StatusNotImplemented` | 501 |
| `StatusBadGateway` | 502 |
| `StatusServiceUnavailable` | 503 |
| `StatusGatewayTimeout` | 504 |

## 与传输层集成

### HTTP 错误中间件

`transport/http/middleware/errors` 提供了 HTTP 错误响应中间件：

```go
import (
    httpErrors "github.com/tx7do/go-wind-plugins/transport/http/middleware/errors"
)

// 在 handler 中返回错误时自动转换为 JSON 响应
httpErrors.Respond(w, r, ErrUserNotFound)
// → HTTP 404, {"code":404,"reason":"USER_NOT_FOUND","message":"user not found"}
```

### gRPC 错误中间件

`transport/grpc/middleware/errors` 提供了 gRPC 双向错误转换拦截器：

```go
import (
    grpcErrors "github.com/tx7do/go-wind-plugins/transport/grpc/middleware/errors"
)

// 服务端：*Error → gRPC status
s := grpc.NewServer(
    grpc.UnaryInterceptor(grpcErrors.UnaryServerInterceptor()),
)

// 客户端：gRPC status → *Error
conn, _ := grpc.Dial(addr,
    grpc.WithUnaryInterceptor(grpcErrors.UnaryClientInterceptor()),
)
```

### gRPC ↔ HTTP 状态码映射

| HTTP | gRPC |
|------|------|
| 400 | `InvalidArgument` |
| 401 | `Unauthenticated` |
| 403 | `PermissionDenied` |
| 404 | `NotFound` |
| 409 | `AlreadyExists` / `Aborted` |
| 429 | `ResourceExhausted` |
| 500 | `Internal` |
| 501 | `Unimplemented` |
| 503 | `Unavailable` |
| 504 | `DeadlineExceeded` |

## 设计原则

- **纯 Go 定义**：不依赖 protobuf 工具链，错误类型在标准 Go 代码中定义
- **协议无关**：以 HTTP 状态码为锚点，所有传输层共用同一错误模型
- **零依赖**：仅依赖标准库 `errors` 和 `fmt`
- **可包装**：完全兼容 Go 1.13+ 的 `%w` 错误包装机制
