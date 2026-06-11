package dtm

import (
	"database/sql"
	"net/url"

	"github.com/dtm-labs/client/dtmcli"
)

// XA wraps a DTM XA transaction branch.
//
// XA is a database-level 2-phase commit protocol. It provides the strongest
// consistency guarantee but requires XA-capable databases (MySQL, PostgreSQL, etc.).
type XA struct {
	xa *dtmcli.Xa
}

// CallBranch calls an XA branch transaction.
//   - body:      request payload.
//   - branchURL: URL for the branch operation.
func (x *XA) CallBranch(body interface{}, branchURL string) error {
	_, err := x.xa.CallBranch(body, branchURL)
	return err
}

// XaLocalTransaction starts an XA local transaction on the branch handler side.
//
// This function is used in HTTP handlers that serve as XA branch endpoints.
// It handles the phase-2 commit/rollback automatically based on the
// operation type indicated in the query parameters.
//
//   - qs:      URL query parameters from the DTM callback.
//   - dbConf:  database configuration for the XA resource.
//   - xaFunc:  the business function to execute within the XA transaction.
func XaLocalTransaction(qs url.Values, dbConf dtmcli.DBConf, xaFunc func(db *sql.DB, xa *dtmcli.Xa) error) error {
	return dtmcli.XaLocalTransaction(qs, dbConf, xaFunc)
}
