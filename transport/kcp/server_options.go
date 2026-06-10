package kcp

import (
	"time"

	"github.com/tx7do/go-wind-plugins/encoding"
)

////////////////////////////////////////////////////////////////////////////////

type Option func(o *Server)

func WithAddress(addr string) Option {
	return func(s *Server) {
		s.address = addr
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.timeout = timeout
	}
}

func WithCodec(c string) Option {
	return func(s *Server) {
		s.codec = encoding.GetCodec(c)
	}
}

func WithBlockCrypt(password, salt string) Option {
	return func(s *Server) {
		s.blockCryptPassword = password
		s.blockCryptSalt = salt
	}
}

func WithDataShards(dataShards int) Option {
	return func(s *Server) {
		s.dataShards = dataShards
	}
}

func WithParityShards(parityShards int) Option {
	return func(s *Server) {
		s.parityShards = parityShards
	}
}

func WithMessageMarshaler(m NetPacketMarshaler) Option {
	return func(s *Server) {
		s.netPacketMarshaler = m
	}
}

func WithMessageUnmarshaler(m NetPacketUnmarshaler) Option {
	return func(s *Server) {
		s.netPacketUnmarshaler = m
	}
}

func WithSocketConnectHandler(h SocketConnectHandler) Option {
	return func(s *Server) {
		s.socketConnectHandler = h
	}
}

func WithSocketRawDataHandler(h SocketRawDataHandler) Option {
	return func(s *Server) {
		if h != nil {
			s.socketRawDataHandler = h
		}
	}
}
