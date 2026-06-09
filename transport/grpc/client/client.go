package client

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Middleware 定义 gRPC 客户端一元拦截器类型。
type Middleware = grpc.UnaryClientInterceptor

// StreamMiddleware 定义 gRPC 客户端流拦截器类型。
type StreamMiddleware = grpc.StreamClientInterceptor

// Client 封装 gRPC 客户端连接，提供连接管理和拦截器支持。
// 通过 Conn() 获取底层 *grpc.ClientConn 后，即可使用 protobuf 生成的
// stub 创建具体服务的客户端。
type Client struct {
	target            string
	conn              *grpc.ClientConn
	dialOptions       []grpc.DialOption
	middlewares       []Middleware
	streamMiddlewares []StreamMiddleware
	hasCreds          bool // 是否已显式设置安全凭证
}

// NewClient 创建一个 gRPC 客户端实例。
// target 是 gRPC 目标地址，例如 "localhost:50051" 或 "dns:///svc:50051"。
func NewClient(target string, opts ...Option) *Client {
	c := &Client{target: target}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Dial 建立到目标地址的 gRPC 连接。
// 若已通过 WithConn 注入连接，直接返回。
// 若未通过 WithTransportCredentials 设置安全凭证，默认使用 insecure。
func (c *Client) Dial(ctx context.Context) error {
	if c.conn != nil {
		return nil
	}

	dialOpts := c.buildDialOptions()

	conn, err := grpc.NewClient(c.target, dialOpts...)
	if err != nil {
		return fmt.Errorf("grpc dial %s: %w", c.target, err)
	}
	c.conn = conn
	return nil
}

// Close 关闭 gRPC 连接。
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	err := c.conn.Close()
	c.conn = nil
	return err
}

// Conn 返回底层的 *grpc.ClientConn。
// 若尚未建立连接，返回 nil。
func (c *Client) Conn() *grpc.ClientConn {
	return c.conn
}

// Target 返回目标地址。
func (c *Client) Target() string {
	return c.target
}

// buildDialOptions 组装最终的 DialOption 列表。
func (c *Client) buildDialOptions() []grpc.DialOption {
	opts := make([]grpc.DialOption, 0, len(c.dialOptions)+2)

	// 用户自定义 DialOption 优先
	opts = append(opts, c.dialOptions...)

	// 拦截器链
	if len(c.middlewares) > 0 {
		opts = append(opts, grpc.WithChainUnaryInterceptor(c.middlewares...))
	}
	if len(c.streamMiddlewares) > 0 {
		opts = append(opts, grpc.WithChainStreamInterceptor(c.streamMiddlewares...))
	}

	// 若用户未设置任何安全凭证，默认使用 insecure
	if !c.hasCreds {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	return opts
}
