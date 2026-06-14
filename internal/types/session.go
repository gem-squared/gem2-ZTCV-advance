// Package types contains shared data types for ZTCV — the Zero-Trust Call
// Verification Protocol. These types are pure data (no side-effect methods)
// and use json struct tags for REST serialization across the 5 binaries.
package types

import "time"

// SessionState represents the discrete states a CallSession can occupy
// during the 4-stage verification flow:
//
//	created → caller_proved → customer_proved → risk_checked → anchored → verified
//	                       └─ blocked (terminal failure)
//	                       └─ expired (terminal nonce timeout)
type SessionState string

const (
	StateCreated        SessionState = "created"
	StateCallerProved   SessionState = "caller_proved"
	StateCustomerProved SessionState = "customer_proved"
	StateRiskChecked    SessionState = "risk_checked"
	StateAnchored       SessionState = "anchored"
	StateVerified       SessionState = "verified"
	StateBlocked        SessionState = "blocked"
	StateExpired        SessionState = "expired"
)

// CallSession is the central verification record threaded through every layer
// (session-svc state machine → identity-svc proof → didregistry purpose-scope
// check → risk-chain-svc verdict → chain-adapter anchor). The Passport is
// assembled from this session + TrustTrace + Receipt at the end.
type CallSession struct {
	ID        string       `json:"id"`
	Nonce     string       `json:"nonce"`
	State     SessionState `json:"state"`
	CreatedAt time.Time    `json:"created_at"`
	ExpiresAt time.Time    `json:"expires_at"`
	UpdatedAt time.Time    `json:"updated_at"`

	CallerProof   *CallerProof     `json:"caller_proof,omitempty"`
	CustomerProof *CustomerProof   `json:"customer_proof,omitempty"`
	RiskVerdict   *ComposedVerdict `json:"risk_verdict,omitempty"`
	Receipt       *Receipt         `json:"receipt,omitempty"`
	TxHash        string           `json:"tx_hash,omitempty"`

	BlockReason string `json:"block_reason,omitempty"`
}
