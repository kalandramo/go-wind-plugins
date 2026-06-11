package dtm

import (
	"database/sql"

	"github.com/dtm-labs/client/dtmcli"
)

// Msg wraps a DTM 2-phase message transaction.
//
// The 2-phase message pattern guarantees "at-least-once" delivery of the
// action. Combined with idempotent handlers this achieves "exactly-once"
// semantics.
//
// Usage:
//
//	msg := client.NewMsg("msg-001")
//	msg.Add("/api/send-email", emailPayload)
//	msg.Prepare("/api/query-prepared")
//	// ... do local business logic ...
//	msg.Submit()
//
// Or use DoAndSubmit for prepare + business + submit in one call:
//
//	msg := client.NewMsg("msg-001")
//	msg.Add("/api/send-email", emailPayload)
//	msg.DoAndSubmit("/api/query-prepared", func(bb *dtmcli.BranchBarrier) error {
//	    // local business logic
//	    return nil
//	})
type Msg struct {
	msg *dtmcli.Msg
}

// Add adds a new step to the message transaction.
//   - action: URL that will be called when the message is committed.
//   - data:   request payload.
func (m *Msg) Add(action string, data interface{}) *Msg {
	m.msg.Add(action, data)
	return m
}

// AddTopic adds a new topic-based step for message queue integration.
func (m *Msg) AddTopic(topic string, data interface{}) *Msg {
	m.msg.AddTopic(topic, data)
	return m
}

// SetDelay sets a delay (in seconds) before the action is invoked.
func (m *Msg) SetDelay(delay uint64) *Msg {
	m.msg.SetDelay(delay)
	return m
}

// Prepare prepares the message transaction.
// queryPrepared is the URL DTM will call to check if the transaction should proceed.
func (m *Msg) Prepare(queryPrepared string) error {
	return m.msg.Prepare(queryPrepared)
}

// Submit submits the message transaction.
func (m *Msg) Submit() error {
	return m.msg.Submit()
}

// DoAndSubmit combines prepare + business logic + submit in one call.
// If busiCall returns ErrFailure the transaction is aborted.
// If busiCall returns any other non-nil error, DTM queries queryPrepared
// to determine the outcome.
func (m *Msg) DoAndSubmit(queryPrepared string, busiCall func(bb *dtmcli.BranchBarrier) error) error {
	return m.msg.DoAndSubmit(queryPrepared, busiCall)
}

// DoAndSubmitDB is the same as DoAndSubmit but accepts a *sql.DB for
// database operations wrapped in the branch barrier.
func (m *Msg) DoAndSubmitDB(queryPrepared string, db *sql.DB, busiCall dtmcli.BarrierBusiFunc) error {
	return m.msg.DoAndSubmitDB(queryPrepared, db, busiCall)
}
