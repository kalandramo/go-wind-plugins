package dtm

import "github.com/dtm-labs/client/dtmcli"

// Saga wraps a DTM Saga transaction.
//
// Saga is a forward-compensate pattern: each step has a forward action and a
// compensate action. If any step fails, all previously completed steps are
// compensated in reverse order.
//
// Steps are executed sequentially by default. Use SetConcurrent for parallel
// execution, and AddBranchOrder to define ordering constraints.
type Saga struct {
	saga *dtmcli.Saga
}

// Add adds a new step to the saga transaction.
//   - action:     URL for the forward (happy-path) operation.
//   - compensate: URL for the compensation (rollback) operation.
//   - data:       request payload sent to both action and compensate endpoints.
func (s *Saga) Add(action, compensate string, data interface{}) *Saga {
	s.saga.Add(action, compensate, data)
	return s
}

// AddBranchOrder specifies that branch should execute after preBranches.
// Branch indices are 0-based, matching the order of Add calls.
func (s *Saga) AddBranchOrder(branch int, preBranches []int) *Saga {
	s.saga.AddBranchOrder(branch, preBranches)
	return s
}

// SetConcurrent enables concurrent execution of saga branches.
// Use AddBranchOrder to impose ordering constraints when needed.
func (s *Saga) SetConcurrent() *Saga {
	s.saga.SetConcurrent()
	return s
}

// Submit submits the saga transaction to the DTM server.
func (s *Saga) Submit() error {
	return s.saga.Submit()
}
