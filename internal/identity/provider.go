// Package identity implements the IdentityProvider interface for
// customer-side Mobile ID verification. Two providers ship:
//
//   - MockMobileIDProvider — deterministic JWT for the 3 demo
//     customers (customer-001/002/003). Always active in Phase 1.
//
//   - OACXProvider — real-slot OmniOne CX REST integration. Compiles
//     and boots whenever OMNIONE_CX_MODE=real and OMNIONE_CX_LICENSE_KEY
//     is set. Inert in Phase 1 (we run with MODE=mock).
//
// The handler layer selects the provider at boot per cfg.OmniOneCX.Mode.
package identity

import (
	"github.com/gem-squared/gem2-ZTCV/internal/types"
)

// IdentityProvider abstracts Mobile ID verification.
type IdentityProvider interface {
	// StartVerification produces a QR/app-link the customer follows to
	// prove their Mobile ID. The returned VerificationRequest carries
	// the session binding so the callback can be correlated.
	StartVerification(sessionID string) (*types.VerificationRequest, error)

	// VerifyToken decodes the JWT/token returned by OACX (or the mock)
	// and yields the MobileIDClaims (hashed-PII only).
	VerifyToken(oacxToken string) (*types.MobileIDClaims, error)

	// Mode identifies the active provider for logging + UI labels.
	Mode() string
}
