package client

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Option 定义 gRPC 客户端的配置选项。
type Option func(*Client)

// WithConn 通过选项直接注入已建立的 *grpc.ClientConn。
// 若调用方已有自定义连接，可由此注入，跳过 Dial。
func WithConn(conn *grpc.ClientConn) Option {
	return func(c *Client) { c.conn = conn }
}

// WithDialOption 追加原生 grpc.DialOption，透传给底层 grpc.NewClient。
func WithDialOption(opts ...grpc.DialOption) Option {
	return func(c *Client) { c.dialOptions = append(c.dialOptions, opts...) }
}

// WithTransportCredentials 设置传输层安全凭证（TLS）。
func WithTransportCredentials(creds credentials.TransportCredentials) Option {
	return func(c *Client) {
		c.hasCreds = true
		c.dialOptions = append(c.dialOptions, grpc.WithTransportCredentials(creds))
	}
}

// WithInsecure 设置不安全传输（明文）。与 WithTransportCredentials 互斥。
func WithInsecure() Option {
	return func(c *Client) {
		c.hasCreds = true
		c.dialOptions = append(c.dialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
}

// WithMiddleware 追加一元拦截器。
func WithMiddleware(middlewares ...Middleware) Option {
	return func(c *Client) { c.middlewares = append(c.middlewares, middlewares...) }
}

// WithStreamMiddleware 追加流拦截器。
func WithStreamMiddleware(middlewares ...StreamMiddleware) Option {
	return func(c *Client) { c.streamMiddlewares = append(c.streamMiddlewares, middlewares...) }
}
