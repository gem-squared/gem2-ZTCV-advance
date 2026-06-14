package types

import "time"

// AuthorizationStatus is the active/revoked/suspended state of an agent
// authorization. Only "active" plus a valid time window grants purpose-scope
// permission.
type AuthorizationStatus string

const (
	AuthActive    AuthorizationStatus = "active"
	AuthRevoked   AuthorizationStatus = "revoked"
	AuthSuspended AuthorizationStatus = "suspended"
)

// DIDDocument is a W3C VC/VP-aligned DID Document, minimal subset Go-native.
// Stored in session-svc SQLite did_registry table; resolved by the didregistry
// module's Resolver interface.
type DIDDocument struct {
	Context            []string             `json:"@context"`
	ID                 string               `json:"id"` // e.g., did:opendid:org:kakaobank
	Controller         []string             `json:"controller,omitempty"`
	VerificationMethod []VerificationMethod `json:"verificationMethod"`
	Authentication     []string             `json:"authentication"`
	Service            []ServiceEndpoint    `json:"service,omitempty"`
	Status             string               `json:"-"` // active|revoked|suspended (internal, not serialized)
	CreatedAt          time.Time            `json:"created"`
}

// VerificationMethod carries an Ed25519 public key used to verify CallerProof
// signatures. Ed25519 is the Open DID standard for this method type.
type VerificationMethod struct {
	ID              string `json:"id"`
	Type            string `json:"type"` // "Ed25519VerificationKey2020"
	Controller      string `json:"controller"`
	PublicKeyBase58 string `json:"publicKeyBase58"`
}

// ServiceEndpoint is the W3C DID Document service descriptor — not required
// for ZTCV MVP but included for forward compatibility.
type ServiceEndpoint struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	ServiceEndpoint string `json:"serviceEndpoint"`
}

// AgentAuthorization is the purpose-scope grant linking an AI agent DID to an
// organization, with an explicit allowlist of permitted call purposes. This
// is the data structure that powers the Scenario 3 thesis: even a verified
// agent is blocked when its proof.Purpose ∉ AllowedPurposes.
type AgentAuthorization struct {
	AgentDID        string              `json:"agent_did"`
	OrgDID          string              `json:"org_did"`
	AllowedPurposes []string            `json:"allowed_purposes"` // e.g., ["security_alert"]
	ValidFrom       time.Time           `json:"valid_from"`
	ValidUntil      time.Time           `json:"valid_until"`
	Status          AuthorizationStatus `json:"status"`
}

// PhoneBinding maps a hashed phone number to its bound DID, enabling reverse
// lookups during the verification flow without ever storing raw numbers.
type PhoneBinding struct {
	PhoneHash     string    `json:"phone_hash"` // sha256 hex of E.164 phone number
	DID           string    `json:"did"`
	BindingStatus string    `json:"binding_status"` // active|revoked
	CreatedAt     time.Time `json:"created_at"`
}

// DIDRecord is a denormalized row used by the session-svc DID admin API
// (registration + read). It mirrors the DIDDocument plus a display name and
// the parsed subject type for fast filtering.
type DIDRecord struct {
	DID         string       `json:"did"`
	SubjectType string       `json:"subject_type"` // org|agent|user|phone
	DisplayName string       `json:"display_name"`
	Document    *DIDDocument `json:"document,omitempty"`
}
