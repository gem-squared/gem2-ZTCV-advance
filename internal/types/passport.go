package types

import "time"

// StampStatus is the per-stamp result on a Call Passport. Either OK (passed)
// or FAIL (this verification step did not satisfy its predicate).
//
// CONTRACT BOUNDARY: the JSON encoding ("OK" / "FAIL") MUST match the
// Moafik directive §7 TypeScript shape exactly. The frontend's
// CallPassportCard + CallPassportFailedCard components depend on these
// literal string values.
type StampStatus string

const (
	StampOK   StampStatus = "OK"
	StampFAIL StampStatus = "FAIL"
)

// Stamp is one entry on a Call Passport. Label uses Korean stamp names for
// demo readability (e.g., "기관 DID", "AI Agent 권한", "수신자 Mobile ID",
// "Chain Receipt" for SAFE; "발신자 DID 없음", "AI Agent 권한 (권한 외 목적)",
// "송금 요구 감지", "차단 영수증 기록" for FAILED variants).
//
// CONTRACT BOUNDARY: JSON field names match Moafik directive §7 Stamp type
// verbatim. Do not rename without updating the frontend in lock-step.
type Stamp struct {
	Label  string      `json:"label"`
	Status StampStatus `json:"status"`
	Detail string      `json:"detail,omitempty"`
}

// CallPassportOutcome is the top-level verification outcome.
type CallPassportOutcome string

const (
	OutcomeSAFE   CallPassportOutcome = "SAFE"
	OutcomeFAILED CallPassportOutcome = "FAILED"
)

// CallPassport is the user-facing artifact issued by ZTCV before a call
// connects. Locked Korean definition:
//
//	"통화 연결 전에 발급되는 1회용 통화 신원 여권"
//	(a single-use call-identity passport issued before the call connects)
//
// It is a UX metaphor, not a legal travel document.
//
// CONTRACT BOUNDARY: this struct's JSON shape MUST match the Moafik
// directive §7 CallPassport TypeScript type exactly. Field names, types,
// and optionality (omitempty mapping to TypeScript ?) are part of the
// frontend ↔ backend interface contract.
//
// Composition: built by internal/passport/builder.go from a CallSession +
// TrustTrace + Receipt. The 4-stamp model is preserved for SAFE outcomes;
// FAILED outcomes typically show 3 stamps including "차단 영수증 기록" ✓
// because the BLOCK decision itself is anchored to chain.
type CallPassport struct {
	SessionID     string              `json:"sessionId"`
	IssuedAt      time.Time           `json:"issuedAt"`
	ExpiresAt     time.Time           `json:"expiresAt"`
	Outcome       CallPassportOutcome `json:"outcome"`
	Stamps        []Stamp             `json:"stamps"`
	BlockReason   string              `json:"blockReason,omitempty"`   // FAILED only — Korean explanation
	ReceiptTxHash string              `json:"receiptTxHash,omitempty"` // mock "0xMOCK…" in Phase 1; real tx hash in Phase 2
	ExplorerURL   string              `json:"explorerUrl,omitempty"`
	CallerDID     string              `json:"callerDid,omitempty"`
	CallerOrg     string              `json:"callerOrg,omitempty"`
	CallerPurpose string              `json:"callerPurpose,omitempty"`

	// Intent Handshake — Step 7 of the 9-step verification pipeline.
	// Added in WP-ZTCV-06. omitempty preserves S1/S2/S3 backwards
	// compatibility for any client that does not yet read these fields.
	IntentHandshake    *IntentManifest `json:"intent_handshake,omitempty"`
	IntentManifestHash string          `json:"intent_manifest_hash,omitempty"`
}

// IntentManifest is the Predictive Disclosure output from the Broker
// AI Agent. Tells the receiver, before the phone rings, what the
// caller is likely to request and what a legitimate caller of the
// declared purpose will NEVER ask. Source = "live" when an LLM
// generated it; "fallback" when the deterministic per-scenario script
// produced it (LLM unreachable / timeout / parse failure).
//
// CONTRACT BOUNDARY: this struct's JSON shape is part of the
// frontend ↔ backend interface contract for any client that opts in
// to render Intent Handshake content.
type IntentManifest struct {
	ExpectedRequests  []string  `json:"expected_requests"`
	ForbiddenRequests []string  `json:"forbidden_requests"`
	SafetySummary     string    `json:"safety_summary"`
	Source            string    `json:"source"`             // "live" | "fallback"
	Provider          string    `json:"provider,omitempty"` // telemetry only; e.g. "anthropic/claude-haiku-4-5"
	GeneratedAt       time.Time `json:"generated_at"`
}

// Canonical stamp labels per scenario (locked across project).
const (
	// SAFE stamps — all 4 ✓
	StampLabelOrgDID       = "기관 DID"
	StampLabelAgentAuth    = "AI Agent 권한"
	StampLabelMobileID     = "수신자 Mobile ID"
	StampLabelChainReceipt = "Chain Receipt"

	// FAILED stamps (Scenario 2 — unknown caller DID + transfer demand)
	StampLabelMissingDID     = "발신자 DID 없음"
	StampLabelTransferDemand = "송금 요구 감지"
	StampLabelBlockReceipt   = "차단 영수증 기록"

	// FAILED stamp (Scenario 3 — unauthorized purpose)
	StampLabelUnauthorizedPurpose = "AI Agent 권한 (권한 외 목적)"
)
