package socketio

import (
	"crypto/tls"

	socketIo "github.com/googollee/go-socket.io"
	"github.com/tx7do/go-wind-plugins/encoding"
)

type Option func(o *Server)

func WithNetwork(network string) Option {
	return func(s *Server) {
		s.network = network
	}
}

func WithAddress(addr string) Option {
	return func(s *Server) {
		s.address = addr
	}
}

func WithTLSConfig(c *tls.Config) Option {
	return func(o *Server) {
		o.tlsConf = c
	}
}

func WithCodec(c string) Option {
	return func(s *Server) {
		s.codec = encoding.GetCodec(c)
	}
}

func WithPath(path string) Option {
	return func(s *Server) {
		s.path = path
	}
}

func WithConnectHandler(namespace string, f func(socketIo.Conn) error) Option {
	return func(s *Server) {
		s.Server.OnConnect(namespace, f)
	}
}

func WithDisconnectHandler(namespace string, f func(socketIo.Conn, string)) Option {
	return func(s *Server) {
		s.Server.OnDisconnect(namespace, f)
	}
}

func WithErrorHandler(namespace string, f func(socketIo.Conn, error)) Option {
	return func(s *Server) {
		s.Server.OnError(namespace, f)
	}
}

func WithEventHandler(namespace, event string, f any) Option {
	return func(s *Server) {
		s.Server.OnEvent(namespace, event, f)
	}
}
