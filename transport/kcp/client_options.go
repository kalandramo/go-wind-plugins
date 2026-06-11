package kcp

import (
	"github.com/tx7do/go-wind-plugins/encoding"
	"github.com/tx7do/go-wind-plugins/metrics"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type ClientOption func(o *Client)

func WithClientCodec(codec string) ClientOption {
	return func(c *Client) {
		c.codec = encoding.GetCodec(codec)
	}
}

func WithEndpoint(uri string) ClientOption {
	return func(c *Client) {
		c.url = uri
	}
}

func WithClientRawDataHandler(h ClientRawMessageHandler) ClientOption {
	return func(c *Client) {
		c.rawMessageHandler = h
	}
}

func WithClientBlockCrypt(password, salt string) ClientOption {
	return func(c *Client) {
		c.blockCryptPassword = password
		c.blockCryptSalt = salt
	}
}

func WithClientDataShards(dataShards int) ClientOption {
	return func(c *Client) {
		c.dataShards = dataShards
	}
}

func WithClientParityShards(parityShards int) ClientOption {
	return func(c *Client) {
		c.parityShards = parityShards
	}
}

// WithClientTracer 注入自定义链路追踪器。
func WithClientTracer(t trace.Tracer) ClientOption {
	return func(c *Client) {
		c.tracer = t
	}
}

// WithClientTracerProvider 使用全局 TracerProvider 创建 tracer。
func WithClientTracerProvider() ClientOption {
	return func(c *Client) {
		c.tracer = otel.GetTracerProvider().Tracer("kcp-client")
	}
}

// WithClientMetrics 注入指标监控。
func WithClientMetrics(m metrics.Metrics) ClientOption {
	return func(c *Client) {
		c.m = m
	}
}
