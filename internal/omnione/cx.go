// Package omnione provides adapter interfaces and factory functions for the
// three OmniOne integration points (CX, Open DID, Chain). Default impls are
// deterministic mocks so the demo runs locally without any external credentials.
// Real-slot impls activate when corresponding env vars are set.
//
// Design discipline (David, 2026-05-28):
//   - Adapter pattern: real-slot ready, mock-fallback default
//   - Never claim production OmniOne integration unless env actually points to it
//   - Mock impls satisfy the same iface — drop-in replacement when credentials arrive
//
// CX = OmniOne CX (Receiver-side Mobile ID verification — 표준인증창 + VC-Verifier)
package omnione

import (
	"fmt"

	"github.com/gem-squared/gem2-ZTCV/internal/config"
	"github.com/gem-squared/gem2-ZTCV/internal/identity"
)

// CXAdapter is the OmniOne CX-shaped Mobile ID verification interface.
// Currently aliased to identity.IdentityProvider — we keep the existing
// production iface as canonical, and use this name in main.go for clarity
// at the orchestration layer.
type CXAdapter = identity.IdentityProvider

// CXMode mirrors config.OACXMode but lives in this package so callers don't
// need to import config for the switch.
type CXMode = config.OACXMode

const (
	CXModeMock CXMode = config.OACXMock
	CXModeReal CXMode = config.OACXReal
)

// NewCXAdapter selects the OmniOne CX adapter implementation per env config.
//
// MODE switches:
//
//	OMNIONE_CX_MODE=mock  → deterministic MockMobileIDProvider (default)
//	OMNIONE_CX_MODE=real  → OACXProvider against cx.raonsecure.co.kr:18543
//	                       (requires OMNIONE_CX_LICENSE_KEY)
//
// Returns an error (rather than falling back silently) when mode=real but the
// license is absent — fail-fast at boot is safer than silent mock substitution.
func NewCXAdapter(cfg config.OmniOneCX, devMockSecret string) (CXAdapter, string, error) {
	switch cfg.Mode {
	case CXModeMock:
		return identity.NewMockProvider(devMockSecret),
			"mock (OmniOne CX real-slot inactive)",
			nil
	case CXModeReal:
		if cfg.LicenseKey == "" {
			return nil, "", fmt.Errorf(
				"OMNIONE_CX_MODE=real requires OMNIONE_CX_LICENSE_KEY to be set",
			)
		}
		return identity.NewOACXProvider(cfg.BaseURL, cfg.LicenseKey),
			"real (OmniOne CX live — license configured)",
			nil
	default:
		return nil, "", fmt.Errorf(
			"invalid OMNIONE_CX_MODE %q (expected: mock | real)", cfg.Mode,
		)
	}
}
