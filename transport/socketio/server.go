// Package socketio provides a Socket.IO server that implements the
// [transport.Server] interface.
//
// It wraps the go-socket.io library with gorilla/mux routing and CORS support.
// The server lifecycle is managed via the standard Start/Stop pattern,
// making it compatible with [wind.App].
package socketio

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"

	"github.com/tx7do/go-wind-plugins/encoding"
	"github.com/tx7do/go-wind/transport"

	socketIo "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	socketIoTransport "github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// KindSocketIo 是 Socket.IO 传输类型标识。
const KindSocketIo = "socket.io"

// 确保 Server 实现了 wind transport.Server 接口。
var _ transport.Server = (*Server)(nil)

type Server struct {
	*socketIo.Server

	lis     net.Listener
	tlsConf *tls.Config

	network string
	address string
	path    string

	codec encoding.Codec

	router *mux.Router
}

func NewServer(opts ...Option) *Server {
	srv := &Server{
		network: "tcp",
		address: ":0",
		router:  mux.NewRouter(),
		path:    "/socket.io/",
	}

	srv.init(opts...)

	return srv
}

// Start 启动 Socket.IO 服务器，阻塞直到 ctx 被取消。
func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen(s.network, s.address)
	if err != nil {
		return err
	}
	s.lis = lis

	log.Printf("[socket.io] server listening on: %s", lis.Addr().String())

	go func() {
		if err := s.Server.Serve(); err != nil {
			log.Printf("[socket.io] serve error: %s", err)
		}
	}()

	handler := handlers.CORS()(s.router)

	go func() {
		if s.tlsConf != nil {
			_ = http.ServeTLS(s.lis, handler, "", "")
		} else {
			_ = http.Serve(s.lis, handler)
		}
	}()

	// 阻塞等待 ctx 取消
	<-ctx.Done()

	_ = s.Server.Close()
	if s.lis != nil {
		_ = s.lis.Close()
	}

	log.Println("[socket.io] server stopped")
	return nil
}

// Stop 优雅关闭 Socket.IO 服务器。
func (s *Server) Stop(_ context.Context) error {
	err := s.Server.Close()
	if s.lis != nil {
		_ = s.lis.Close()
	}
	log.Println("[socket.io] server stopped")
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
		return KindSocketIo + "://" + addr
	}
	if host == "" || host == "0.0.0.0" {
		host = "localhost"
	}
	return KindSocketIo + "://" + net.JoinHostPort(host, port)
}

func (s *Server) RegisterConnectHandler(namespace string, f func(socketIo.Conn) error) {
	s.Server.OnConnect(namespace, f)
}

func (s *Server) RegisterDisconnectHandler(namespace string, f func(socketIo.Conn, string)) {
	s.Server.OnDisconnect(namespace, f)
}

func (s *Server) RegisterErrorHandler(namespace string, f func(socketIo.Conn, error)) {
	s.Server.OnError(namespace, f)
}

func (s *Server) RegisterEventHandler(namespace, event string, f any) {
	s.Server.OnEvent(namespace, event, f)
}

func (s *Server) init(opts ...Option) {
	server := socketIo.NewServer(&engineio.Options{
		Transports: []socketIoTransport.Transport{
			&polling.Transport{
				CheckOrigin: func(r *http.Request) bool { return true },
			},
			&websocket.Transport{
				CheckOrigin: func(r *http.Request) bool { return true },
			},
		},
	})
	if server == nil {
		log.Printf("[socket.io] create server failed")
		return
	}
	s.Server = server

	for _, o := range opts {
		o(s)
	}

	s.router.Use(mux.CORSMethodMiddleware(s.router))

	s.router.Handle(s.path, server)
}
