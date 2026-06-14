// Open DID adapter — Caller-side institution DID + AI Agent VC resolution.
//
// Default: Go-native didregistry (W3C-aligned DID/VC data model + 해소 인터페이스
// per Dev-related/04-open-did.md).
// Real-slot (future): Java sidecar with did-issuer-server / did-verifier-server
// integration per github.com/OmniOneID (deferred to 결선 dev phase).
package omnione

import (
	"fmt"

	"github.com/gem-squared/gem2-ZTCV/internal/config"
	"github.com/gem-squared/gem2-ZTCV/internal/didregistry"
)

// OpenDIDAdapter is the OmniOne Open DID-shaped resolution interface used by
// the Verification Broker for Caller DID binding, AgentAuthorization
// verification, and DID document resolution.
//
// The current Go-native adapter is canonical for Stage 1. A Real adapter
// (Java SDK sidecar — did-issuer-server + did-verifier-server) is reserved
// for the 결선 dev phase.
type OpenDIDAdapter interface {
	// Mode returns the human-readable mode label for boot-log clarity.
	Mode() string
	// Repo returns the underlying registry so existing handlers in the
	// session-svc can continue to use it directly. This is a transitional
	// surface — future refactors will move all DID operations behind the
	// adapter interface.
	Repo() *didregistry.Repo
}

type goNativeOpenDID struct {
	repo *didregistry.Repo
	mode string
}

func (g *goNativeOpenDID) Mode() string            { return g.mode }
func (g *goNativeOpenDID) Repo() *didregistry.Repo { return g.repo }

// NewOpenDIDAdapter selects the Open DID adapter per env config.
//
// MODE switches:
//
//	OMNIONE_OPENDID_MODE=mock  → Go-native didregistry (default)
//	OMNIONE_OPENDID_MODE=real  → external Java SDK sidecar (NOT YET WIRED;
//	                            returns error to make the missing dependency
//	                            visible at boot)
func NewOpenDIDAdapter(mode config.OpenDIDMode, repo *didregistry.Repo) (OpenDIDAdapter, string, error) {
	switch mode {
	case config.OpenDIDMock, config.OpenDIDMode(""):
		return &goNativeOpenDID{
				repo: repo,
				mode: "mock (Go-native DID/VC data model + 해소 인터페이스)",
			},
			"mock (Go-native DID/VC data model + 해소 인터페이스)",
			nil
	case config.OpenDIDReal:
		return nil, "", fmt.Errorf(
			"OMNIONE_OPENDID_MODE=real not yet wired — Java sidecar integration " +
				"deferred to 결선 dev phase. Use OMNIONE_OPENDID_MODE=mock for now.",
		)
	default:
		return nil, "", fmt.Errorf(
			"invalid OMNIONE_OPENDID_MODE %q (expected: mock | real)", mode,
		)
	}
}
