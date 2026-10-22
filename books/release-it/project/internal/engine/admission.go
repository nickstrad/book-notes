package engine

import (
	"errors"
)

// ErrStaleEpoch means the caller's route predates the database's current
// placement; it must re-resolve.
var ErrStaleEpoch = errors.New("stale epoch")

// admit decides whether a request may run on this handle. The node only
// serves requests routed with the database's current epoch.
func (h *Handle) admit(epoch int) error {
	if epoch != h.Epoch {
		return ErrStaleEpoch
	}
	return nil
}
