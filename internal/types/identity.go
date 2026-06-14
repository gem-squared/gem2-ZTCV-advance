package types

import "time"

// VCType enumerates Mobile ID credential variants supported by OmniOne CX.
type VCType string

const (
	VCTypeMobileResidentCard VCType = "mresidentcard"
	VCTypeMobileDriverLic    VCType = "mdriverlic"
	VCTypeMobileVeteransCard VCType = "mveteranscard"
)

// MobileIDClaims is the customer-side identity payload extracted from an
// OmniOne CX JWT (real provider) or generated deterministically by the mock
// provider. Raw PII (name, doc number) is hashed before persistence — only
// the hashed forms live in this struct beyond in-memory verification.
type MobileIDClaims struct {
	VCType     VCType    `json:"vc_type"`
	NameHash   string    `json:"name_hash"`   // sha256 hex
	BirthDate  string    `json:"birth_date"`  // YYYYMMDD
	IssuingOrg string    `json:"issuing_org"` // e.g., "서울특별시경찰청장"
	DocNoHash  string    `json:"doc_no_hash"` // sha256 hex of licNo / residentNo / veteranNo
	VerifiedAt time.Time `json:"verified_at"`
}

// CallerProof is the institution-side caller payload signed Ed25519 by the
// organization's controller key. The session-svc caller-proof endpoint
// validates this against the didregistry (DID resolution + agent authorization
// + purpose-scope enforcement).
type CallerProof struct {
	CallerDID    string `json:"caller_did"` // e.g., did:opendid:agent:kakaobank-ai-loan-counselor-001
	OrgDID       string `json:"org_did"`    // e.g., did:opendid:org:kakaobank
	Purpose      string `json:"purpose"`    // e.g., loan_consultation | security_alert | insurance_sales
	SessionID    string `json:"session_id"`
	Nonce        string `json:"nonce"`
	Signature    []byte `json:"signature"`     // Ed25519 over canonical JSON of above fields
	SignatureAlg string `json:"signature_alg"` // always "Ed25519" in MVP
}

// CustomerProof wraps an OmniOne CX token with the session binding so the
// customer-proof REST endpoint can validate freshness + nonce.
type CustomerProof struct {
	OACXToken  string          `json:"oacx_token"`
	Claims     *MobileIDClaims `json:"claims,omitempty"`
	SessionID  string          `json:"session_id"`
	ReceivedAt time.Time       `json:"received_at"`
}

// VerificationRequest is what identity-svc returns from StartVerification —
// either a QR code payload (desktop initiation) or an app-link (mobile init).
type VerificationRequest struct {
	SessionID string `json:"session_id"`
	QRPayload string `json:"qr_payload,omitempty"` // PC web flow
	AppLink   string `json:"app_link,omitempty"`   // mobile web/app flow
	ExpiresAt string `json:"expires_at"`           // ISO8601
}
