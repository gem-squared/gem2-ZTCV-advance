// Package didregistry is the in-process DID registry module owned by
// session-svc. It implements W3C VC/VP-aligned DID Documents +
// AgentAuthorization + PhoneBinding storage and resolution.
//
// Phase 1 simulation uses an in-process SQLite repo (shared with
// session-svc DB). Production split into a separate didregistry-svc
// is preserved in docs/architecture-extended.md.
package didregistry

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gem-squared/gem2-ZTCV/internal/types"
)

// ErrNotFound is returned when a DID or authorization is missing.
var ErrNotFound = errors.New("did registry: not found")

// Repo persists DIDs + agent authorizations + phone bindings.
type Repo struct {
	db *sql.DB
}

// NewRepo wraps an open *sql.DB (typically the session-svc handle).
func NewRepo(db *sql.DB) (*Repo, error) {
	r := &Repo{db: db}
	if err := r.InitSchema(); err != nil {
		return nil, err
	}
	return r, nil
}

// InitSchema is idempotent.
func (r *Repo) InitSchema() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS did_registry (
			did            TEXT PRIMARY KEY,
			subject_type   TEXT NOT NULL,
			display_name   TEXT NOT NULL,
			document_json  TEXT NOT NULL,
			created_at     TIMESTAMP NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS agent_authorizations (
			agent_did          TEXT NOT NULL,
			org_did            TEXT NOT NULL,
			allowed_purposes   TEXT NOT NULL,
			valid_from         TIMESTAMP NOT NULL,
			valid_until        TIMESTAMP NOT NULL,
			status             TEXT NOT NULL,
			PRIMARY KEY (agent_did)
		)`,
		`CREATE TABLE IF NOT EXISTS phone_bindings (
			phone_hash      TEXT PRIMARY KEY,
			did             TEXT NOT NULL,
			binding_status  TEXT NOT NULL,
			created_at      TIMESTAMP NOT NULL
		)`,
	}
	for _, s := range stmts {
		if _, err := r.db.Exec(s); err != nil {
			return fmt.Errorf("didregistry init: %w", err)
		}
	}
	return nil
}

// UpsertDID inserts or updates a DIDDocument.
func (r *Repo) UpsertDID(rec types.DIDRecord) error {
	raw, _ := json.Marshal(rec.Document)
	_, err := r.db.Exec(`
		INSERT INTO did_registry(did, subject_type, display_name, document_json, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(did) DO UPDATE SET
			subject_type   = excluded.subject_type,
			display_name   = excluded.display_name,
			document_json  = excluded.document_json
	`, rec.DID, rec.SubjectType, rec.DisplayName, string(raw), time.Now().UTC())
	if err != nil {
		return fmt.Errorf("upsert did: %w", err)
	}
	return nil
}

// Resolve returns the DIDDocument or ErrNotFound.
func (r *Repo) Resolve(did string) (*types.DIDDocument, error) {
	var raw string
	err := r.db.QueryRow(`SELECT document_json FROM did_registry WHERE did = ?`, did).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("resolve %q: %w", did, err)
	}
	var doc types.DIDDocument
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		return nil, fmt.Errorf("decode did doc %q: %w", did, err)
	}
	return &doc, nil
}

// UpsertAuthorization stores an AgentAuthorization (one per agent).
func (r *Repo) UpsertAuthorization(a types.AgentAuthorization) error {
	purposes := strings.Join(a.AllowedPurposes, ",")
	_, err := r.db.Exec(`
		INSERT INTO agent_authorizations(agent_did, org_did, allowed_purposes, valid_from, valid_until, status)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(agent_did) DO UPDATE SET
			org_did          = excluded.org_did,
			allowed_purposes = excluded.allowed_purposes,
			valid_from       = excluded.valid_from,
			valid_until      = excluded.valid_until,
			status           = excluded.status
	`, a.AgentDID, a.OrgDID, purposes, a.ValidFrom, a.ValidUntil, string(a.Status))
	if err != nil {
		return fmt.Errorf("upsert auth: %w", err)
	}
	return nil
}

// LoadAuthorization returns the AgentAuthorization for an agent DID.
func (r *Repo) LoadAuthorization(agentDID string) (*types.AgentAuthorization, error) {
	var (
		orgDID, purposesRaw, status string
		validFrom, validUntil       time.Time
	)
	err := r.db.QueryRow(`
		SELECT org_did, allowed_purposes, valid_from, valid_until, status
		FROM agent_authorizations
		WHERE agent_did = ?
	`, agentDID).Scan(&orgDID, &purposesRaw, &validFrom, &validUntil, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("load auth %q: %w", agentDID, err)
	}
	return &types.AgentAuthorization{
		AgentDID:        agentDID,
		OrgDID:          orgDID,
		AllowedPurposes: splitNonEmpty(purposesRaw, ","),
		ValidFrom:       validFrom,
		ValidUntil:      validUntil,
		Status:          types.AuthorizationStatus(status),
	}, nil
}

// IsAuthorized is a convenience: does the agent have purpose-scope?
// Returns (true, nil) on match, (false, ErrNotFound) if no auth row
// at all, or (false, nil) if auth exists but purpose mismatch /
// out-of-time-window / non-active status.
//
// This is the function that powers the Scenario 3 BLOCK trigger.
func (r *Repo) IsAuthorized(agentDID, purpose string, now time.Time) (bool, *types.AgentAuthorization, error) {
	auth, err := r.LoadAuthorization(agentDID)
	if err != nil {
		return false, nil, err
	}
	if auth.Status != types.AuthActive {
		return false, auth, nil
	}
	if now.Before(auth.ValidFrom) || now.After(auth.ValidUntil) {
		return false, auth, nil
	}
	for _, p := range auth.AllowedPurposes {
		if p == purpose {
			return true, auth, nil
		}
	}
	return false, auth, nil
}

// UpsertPhoneBinding stores a phone-hash → DID mapping.
func (r *Repo) UpsertPhoneBinding(pb types.PhoneBinding) error {
	_, err := r.db.Exec(`
		INSERT INTO phone_bindings(phone_hash, did, binding_status, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(phone_hash) DO UPDATE SET
			did             = excluded.did,
			binding_status  = excluded.binding_status
	`, pb.PhoneHash, pb.DID, pb.BindingStatus, pb.CreatedAt)
	if err != nil {
		return fmt.Errorf("upsert phone: %w", err)
	}
	return nil
}

// LoadPhoneBinding returns the PhoneBinding or ErrNotFound.
func (r *Repo) LoadPhoneBinding(phoneHash string) (*types.PhoneBinding, error) {
	var (
		did, status string
		createdAt   time.Time
	)
	err := r.db.QueryRow(`SELECT did, binding_status, created_at FROM phone_bindings WHERE phone_hash = ?`, phoneHash).
		Scan(&did, &status, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("load phone %q: %w", phoneHash, err)
	}
	return &types.PhoneBinding{PhoneHash: phoneHash, DID: did, BindingStatus: status, CreatedAt: createdAt}, nil
}

func splitNonEmpty(s, sep string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, sep)
}
