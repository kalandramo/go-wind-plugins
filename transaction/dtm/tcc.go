package dtm

import "github.com/dtm-labs/client/dtmcli"

// Tcc wraps a DTM TCC (Try-Confirm-Cancel) transaction branch.
//
// TCC is a strong-consistency pattern where each branch has three phases:
//   - Try:     reserve resources (e.g. hold inventory).
//   - Confirm: commit the reserved resources.
//   - Cancel:  release the reserved resources.
type Tcc struct {
	tcc *dtmcli.Tcc
}

// CallBranch registers and invokes a TCC branch transaction.
//   - body:       request payload.
//   - tryURL:     URL for the Try phase.
//   - confirmURL: URL for the Confirm phase.
//   - cancelURL:  URL for the Cancel phase.
func (t *Tcc) CallBranch(body interface{}, tryURL, confirmURL, cancelURL string) error {
	_, err := t.tcc.CallBranch(body, tryURL, confirmURL, cancelURL)
	return err
}
