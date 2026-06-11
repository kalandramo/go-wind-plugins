// Package transaction defines the common interfaces for distributed
// transaction manager plugins.
//
// Each engine plugin (dtm, etc.) wraps its native SDK client and implements
// the Client interface. Engine-specific operations that cannot be abstracted
// (saga steps, tcc branches, etc.) are exposed as methods on the concrete
// client type.
//
// Example usage:
//
//	c := dtm.NewClient(dtm.WithServer("http://localhost:36789/api/dtmsvr"))
//	defer c.Close()
//
//	saga := c.NewSaga("order-123")
//	saga.Add("/order/create", "/order/create-compensate", orderData)
//	saga.Add("/stock/deduct", "/stock/deduct-compensate", stockData)
//	if err := saga.Submit(); err != nil {
//	    log.Fatal(err)
//	}
package transaction

// Client is the common interface for all distributed transaction manager clients.
// The only universally shared operation is Close; transaction pattern operations
// (saga, tcc, msg, xa) have incompatible signatures across engines and are
// therefore left to concrete types.
type Client interface {
	// Close releases the client connection and cleans up resources.
	Close() error
}
