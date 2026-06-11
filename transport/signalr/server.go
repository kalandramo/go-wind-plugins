// Package signalr provides a SignalR server that implements the
// [transport.Server] interface.
//
// It wraps the philippseith/signalr library with CORS middleware support.
// The server lifecycle is managed via the standard Start/Stop pattern,
// making it compatible with [wind.App].
package signalr

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/tx7do/go-wind-plugins/encoding"
	"github.com/tx7do/go-wind/transport"

	"github.com/philippseith/signalr"
)

// KindSignalR 是 SignalR 传输类型标识。
const KindSignalR = "signalr"

// 确保 Server 实现了 wind transport.Server 接口。
var _ transport.Server = (*Server)(nil)

// signalrLogger 实现 signalr 库的 Logger 接口，使用标准库 log 输出。
type signalrLogger struct {
	debug bool
}

func (l *signalrLogger) Log(keyVals ...any) error {
	log.Printf("[signalr] %v", fmt.Sprint(keyVals...))
	return nil
}

type Server struct {
	signalr.Server

	lis     net.Listener
	tlsConf *tls.Config

	network string
	address string

	keepAliveInterval  time.Duration
	chanReceiveTimeout time.Duration

	streamBufferCapacity uint

	debug bool

	codec encoding.Codec

	hub signalr.HubInterface

	router      *http.ServeMux
	middlewares []Middleware
}

// Middleware 是标准 HTTP 中间件类型。
// 使用类型别名使得 transport/http/middleware 下的中间件可以直接复用。
type Middleware = func(http.Handler) http.Handler

func NewServer(opts ...Option) *Server {
	srv := &Server{
		network:              "tcp",
		address:              ":0",
		router:               http.NewServeMux(),
		keepAliveInterval:    2 * time.Second,
		chanReceiveTimeout:   200 * time.Millisecond,
		streamBufferCapacity: 5,
		debug:                false,
	}

	srv.init(opts...)

	return srv
}

// Start 启动 SignalR 服务器，阻塞直到 ctx 被取消。
func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen(s.network, s.address)
	if err != nil {
		return err
	}
	s.lis = lis

	log.Printf("[signalr] server listening on: %s", lis.Addr().String())

	// 应用中间件链
	handler := s.CORS(s.router)
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		handler = s.middlewares[i](handler)
	}

	go func() {
		if s.tlsConf != nil {
			_ = http.ServeTLS(s.lis, handler, "", "")
		} else {
			_ = http.Serve(s.lis, handler)
		}
	}()

	// 阻塞等待 ctx 取消
	<-ctx.Done()

	if s.lis != nil {
		_ = s.lis.Close()
	}

	log.Println("[signalr] server stopped")
	return nil
}

// Stop 优雅关闭 SignalR 服务器。
func (s *Server) Stop(_ context.Context) error {
	var err error
	if s.lis != nil {
		err = s.lis.Close()
	}
	log.Println("[signalr] server stopped")
	return err
}

// Endpoint 返回服务器的访问地址。
func (s *Server) Endpoint() string {
	var addr string
	if s.lis != nil {
		addr = s.lis.Addr().String()
	} else {
		addr = s.address
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil || port == "" {
		return KindSignalR + "://" + addr
	}
	if host == "" || host == "0.0.0.0" {
		host = "localhost"
	}
	return KindSignalR + "://" + net.JoinHostPort(host, port)
}

func (s *Server) MapHTTP(path string) {
	s.Server.MapHTTP(signalr.WithHTTPServeMux(s.router), path)
}

// Use 注册全局标准 HTTP 中间件，对所有路由生效。
// 支持直接使用 transport/http/middleware 下的中间件。
// 必须在 Start 之前调用。
func (s *Server) Use(middlewares ...Middleware) {
	s.middlewares = append(s.middlewares, middlewares...)
}

func (s *Server) init(opts ...Option) {
	for _, o := range opts {
		o(s)
	}

	server, err := signalr.NewServer(context.Background(),
		signalr.Logger(&signalrLogger{debug: s.debug}, s.debug),
		signalr.SimpleHubFactory(s.hub),
		signalr.KeepAliveInterval(s.keepAliveInterval),
		signalr.ChanReceiveTimeout(s.chanReceiveTimeout),
		signalr.StreamBufferCapacity(s.streamBufferCapacity),
	)
	if err != nil {
		log.Printf("[signalr] create server failed: %s", err)
		return
	}
	s.Server = server
}
