// Package workflow defines the common interfaces for workflow engine plugins.
//
// Each engine plugin (argo, conductor, goworkflows, temporal) wraps its native
// SDK client and implements the Client interface. Engine-specific operations
// that cannot be abstracted (start workflow, cancel, signal, etc.) are exposed
// as methods on the concrete client type.
//
// Example usage:
//
//	// Create an engine-agnostic cleanup helper
//	func newEngine(engine string) (workflow.Client, func()) {
//	    var c workflow.Client
//	    switch engine {
//	    case "temporal":
//	        tc, _ := temporal.NewClient(...)
//	        c = tc
//	    case "argo":
//	        ac, _ := argo.NewClient(...)
//	        c = ac
//	    }
//	    return c, func() { c.Close() }
//	}
package workflow

// Client is the common interface for all workflow engine clients.
// The only universally shared operation is Close; workflow start, cancel,
// and query operations have incompatible signatures across engines and
// are therefore left to concrete types.
type Client interface {
	// Close releases the client connection and cleans up resources.
	Close() error
}

// Worker is the common interface for workflow engine workers.
// Implemented by the conductor, goworkflows, and temporal plugins.
// Argo does not have a worker concept.
type Worker interface {
	// Stop signals the worker to stop polling for tasks.
	Stop()
	// IsRunning returns whether the worker is still active.
	IsRunning() bool
}
