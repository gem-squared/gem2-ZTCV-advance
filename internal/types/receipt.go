package types

import "time"

// Receipt is the verification artifact assembled at the end of the 4-stage
// flow. The OnChain sub-record is what actually gets anchored to
// ZTCVReceiptAnchor.sol — hashes only, NO PII. The OffChain sub-record
// retains the full evidence trail in session-svc SQLite for audit lookup.
type Receipt struct {
	OnChain  ReceiptOnChain  `json:"on_chain"`
	OffChain ReceiptOffChain `json:"off_chain"`
}

// ReceiptOnChain is the 5-field tuple that the Solidity contract stores.
// Every field is a bytes32 or primitive — the structure intentionally
// forbids PII leakage via type discipline.
type ReceiptOnChain struct {
	SessionHash   string `json:"session_hash"`   // 0x-prefixed hex of keccak256(session_id || nonce)
	ReceiptHash   string `json:"receipt_hash"`   // 0x-prefixed hex of keccak256(canonical_json(OnChain fields except this))
	Timestamp     int64  `json:"timestamp"`      // unix seconds
	IsSafe        bool   `json:"is_safe"`        // final_decision == SAFE
	PolicyVersion string `json:"policy_version"` // 0x-prefixed hex of keccak256("ztcv-policy-v0.1.0")
}

// ReceiptOffChain holds the verbose verification context kept in session-svc
// SQLite. Referenced by sessionHash but never anchored.
type ReceiptOffChain struct {
	CallerDID        string           `json:"caller_did"`
	OrgDID           string           `json:"org_did"`
	RecipientDID     string           `json:"recipient_did,omitempty"`
	MobileIDVerified bool             `json:"mobile_id_verified"`
	CallerVerified   bool             `json:"caller_verified"`
	AIRiskVerdict    RiskVerdict      `json:"ai_risk_verdict"`
	FinalDecision    string           `json:"final_decision"` // SAFE|BLOCK
	BlockReason      string           `json:"block_reason,omitempty"`
	ComposedVerdict  *ComposedVerdict `json:"composed_verdict,omitempty"`
	AnchoredTxHash   string           `json:"anchored_tx_hash,omitempty"`
	ExplorerURL      string           `json:"explorer_url,omitempty"`
	// IntentManifestHash — sha256 hex of the canonical Intent
	// Handshake manifest (WP-ZTCV-06). Persisted off-chain only; the
	// on-chain ReceiptOnChain tuple is intentionally NOT changed in
	// this WP so the deployed Solidity contract requires no migration.
	IntentManifestHash string    `json:"intent_manifest_hash,omitempty"`
	GeneratedAt        time.Time `json:"generated_at"`
}
