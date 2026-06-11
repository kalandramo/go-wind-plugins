// Package dtm implements the DTM (Distributed Transaction Manager) plugin
// for the go-wind framework.
//
// It wraps the DTM Go client SDK (github.com/dtm-labs/client) and provides
// support for all major transaction patterns: Saga, TCC, 2-Phase Message,
// and XA.
//
// Example:
//
//	c := dtm.NewClient(dtm.WithServer("http://localhost:36789/api/dtmsvr"))
//	defer c.Close()
//
//	// Saga pattern
//	saga := c.NewSaga("transfer-001")
//	saga.Add("/api/transfer/out", "/api/transfer/out/compensate", body)
//	saga.Add("/api/transfer/in",  "/api/transfer/in/compensate",  body)
//	if err := saga.Submit(); err != nil {
//	    log.Fatal(err)
//	}
package dtm

import (
	"net/url"

	"github.com/dtm-labs/client/dtmcli"
	"github.com/go-resty/resty/v2"

	"github.com/tx7do/go-wind-plugins/transaction"
)

var _ transaction.Client = (*Client)(nil)

// Client wraps the DTM distributed transaction client.
type Client struct {
	server string
}

// NewClient creates a new DTM client with the given options.
func NewClient(opts ...Option) *Client {
	cfg := config{Server: defaultServer}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Client{
		server: cfg.Server,
	}
}

// Close implements [transaction.Client].
func (c *Client) Close() error {
	return nil
}

// Server returns the DTM server address.
func (c *Client) Server() string {
	return c.server
}

// NewSaga creates a new Saga transaction builder.
// gid is the global transaction ID, which must be unique.
func (c *Client) NewSaga(gid string) *Saga {
	return &Saga{
		saga: dtmcli.NewSaga(c.server, gid),
	}
}

// TccGlobalTransaction begins a TCC (Try-Confirm-Cancel) global transaction.
//
// The tccFunc callback receives a Tcc wrapper for registering branches via
// CallBranch. If tccFunc returns a non-nil error the transaction is aborted
// (Cancel); otherwise it is submitted (Confirm).
func (c *Client) TccGlobalTransaction(gid string, tccFunc func(*Tcc) error) error {
	return dtmcli.TccGlobalTransaction2(c.server, gid, func(*dtmcli.Tcc) {},
		func(tcc *dtmcli.Tcc) (*resty.Response, error) {
			err := tccFunc(&Tcc{tcc: tcc})
			return nil, err
		},
	)
}

// NewMsg creates a new 2-phase message transaction builder.
// gid is the global transaction ID, which must be unique.
func (c *Client) NewMsg(gid string) *Msg {
	return &Msg{
		msg: dtmcli.NewMsg(c.server, gid),
	}
}

// XaGlobalTransaction begins an XA global transaction.
//
// The xaFunc callback receives an XA wrapper for registering branches via
// CallBranch. If xaFunc returns a non-nil error the transaction is rolled
// back; otherwise it is committed.
func (c *Client) XaGlobalTransaction(gid string, xaFunc func(*XA) error) error {
	return dtmcli.XaGlobalTransaction2(c.server, gid, func(*dtmcli.Xa) {},
		func(xa *dtmcli.Xa) (*resty.Response, error) {
			err := xaFunc(&XA{xa: xa})
			return nil, err
		},
	)
}

// BarrierFromQuery constructs a BranchBarrier from URL query parameters.
// This is a convenience wrapper for use in HTTP handler endpoints that
// participate in distributed transactions.
func BarrierFromQuery(qs url.Values) (*dtmcli.BranchBarrier, error) {
	return dtmcli.BarrierFromQuery(qs)
}
