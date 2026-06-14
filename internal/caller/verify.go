// Package caller contains caller-proof verification — the gate that
// enforces purpose-scope authorization (the Scenario 3 trigger).
//
// Signature verification is intentionally permissive in Phase 1
// simulation (any non-empty signature passes) so the demo doesn't
// require real key management. In Phase 2 / production this swaps to
// Ed25519 over canonical JSON.
package caller

import (
	"errors"
	"time"

	"github.com/gem-squared/gem2-ZTCV/internal/didregistry"
	"github.com/gem-squared/gem2-ZTCV/internal/types"
)

// Reason is the machine-readable code surfaced to the frontend.
type Reason string

const (
	ReasonOK                  Reason = ""
	ReasonUnknownDID          Reason = "unknown_did"
	ReasonInvalidSignature    Reason = "invalid_signature"
	ReasonExpiredNonce        Reason = "expired_nonce"
	ReasonUnauthorizedPurpose Reason = "unauthorized_purpose"
)

// Result is the outcome of caller-proof verification.
type Result struct {
	OK        bool
	Reason    Reason
	OrgDID    string
	AgentDID  string
	Purpose   string
	OrgName   string
	AuthScope []string
}

// Verify checks the caller proof against the didregistry.
//
// Order matters — first failure short-circuits:
//  1. ResolveAgent(caller_did) — unknown_did
//  2. ResolveOrg(org_did) — unknown_did
//  3. nonce + session match (verified at handler layer before this)
//  4. signature non-empty (permissive Phase 1)
//  5. agent authorization exists for purpose (unauthorized_purpose)
func Verify(r *didregistry.Repo, proof *types.CallerProof, expectedSessionID, expectedNonce string, now time.Time) Result {
	// 1. Resolve agent
	agentDoc, err := r.Resolve(proof.CallerDID)
	if errors.Is(err, didregistry.ErrNotFound) {
		return Result{Reason: ReasonUnknownDID, AgentDID: proof.CallerDID}
	}
	if err != nil {
		return Result{Reason: ReasonUnknownDID, AgentDID: proof.CallerDID}
	}
	_ = agentDoc

	// 2. Resolve org
	orgDoc, err := r.Resolve(proof.OrgDID)
	if errors.Is(err, didregistry.ErrNotFound) {
		return Result{Reason: ReasonUnknownDID, OrgDID: proof.OrgDID}
	}
	if err != nil {
		return Result{Reason: ReasonUnknownDID, OrgDID: proof.OrgDID}
	}

	// 3. Nonce match
	if proof.SessionID != expectedSessionID || proof.Nonce != expectedNonce {
		return Result{Reason: ReasonExpiredNonce}
	}

	// 4. Signature non-empty (Phase 1 permissive — Phase 2 = real Ed25519)
	if len(proof.Signature) == 0 {
		return Result{Reason: ReasonInvalidSignature}
	}

	// 5. Purpose-scope (THE Scenario 3 trigger)
	ok, auth, err := r.IsAuthorized(proof.CallerDID, proof.Purpose, now)
	if err != nil || auth == nil {
		return Result{Reason: ReasonUnauthorizedPurpose, AgentDID: proof.CallerDID, Purpose: proof.Purpose}
	}
	if !ok {
		return Result{
			Reason:    ReasonUnauthorizedPurpose,
			AgentDID:  proof.CallerDID,
			OrgDID:    proof.OrgDID,
			Purpose:   proof.Purpose,
			OrgName:   orgDoc.ID, // org display name is up to handler; we surface the DID id
			AuthScope: auth.AllowedPurposes,
		}
	}

	return Result{
		OK:        true,
		Reason:    ReasonOK,
		OrgDID:    proof.OrgDID,
		AgentDID:  proof.CallerDID,
		Purpose:   proof.Purpose,
		OrgName:   orgDoc.ID,
		AuthScope: auth.AllowedPurposes,
	}
}

// MakeProof is a Phase 1 helper to construct a CallerProof with a
// non-empty mock signature so tests + scenario fixtures don't need
// real Ed25519 key material yet.
func MakeProof(callerDID, orgDID, purpose, sessionID, nonce string) *types.CallerProof {
	return &types.CallerProof{
		CallerDID:    callerDID,
		OrgDID:       orgDID,
		Purpose:      purpose,
		SessionID:    sessionID,
		Nonce:        nonce,
		Signature:    []byte("mock-signature-phase-1"),
		SignatureAlg: "Ed25519",
	}
}
