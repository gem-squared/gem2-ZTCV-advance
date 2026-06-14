package session

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gem-squared/gem2-ZTCV/internal/types"

	_ "modernc.org/sqlite" // pure-Go sqlite driver, no CGO
)

// ErrNotFound is returned when a session ID has no row.
var ErrNotFound = errors.New("session not found")

// Repo persists CallSession + TrustTrace rows. Sessions store the
// whole CallSession JSON in one blob column for forward-compat; the
// indexed columns (state, expires_at) enable sweepers.
type Repo struct {
	db *sql.DB
}

// NewRepo opens the SQLite file (or :memory:) and runs schema init.
func NewRepo(dsn string) (*Repo, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite open %q: %w", dsn, err)
	}
	r := &Repo{db: db}
	if err := r.InitSchema(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return r, nil
}

// Close releases the underlying handle.
func (r *Repo) Close() error { return r.db.Close() }

// DB exposes the raw handle for sibling packages (didregistry, trace)
// to share the same SQLite file.
func (r *Repo) DB() *sql.DB { return r.db }

// InitSchema creates tables idempotently.
func (r *Repo) InitSchema() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS call_sessions (
			id          TEXT PRIMARY KEY,
			nonce       TEXT NOT NULL,
			state       TEXT NOT NULL,
			expires_at  TIMESTAMP NOT NULL,
			payload     TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_call_sessions_state ON call_sessions(state)`,
		`CREATE INDEX IF NOT EXISTS idx_call_sessions_expires ON call_sessions(expires_at)`,
		`CREATE TABLE IF NOT EXISTS trace_logs (
			session_id  TEXT PRIMARY KEY,
			payload     TEXT NOT NULL,
			created_at  TIMESTAMP NOT NULL
		)`,
	}
	for _, s := range stmts {
		if _, err := r.db.Exec(s); err != nil {
			return fmt.Errorf("init schema: %w", err)
		}
	}
	return nil
}

// Save upserts the session row.
func (r *Repo) Save(s *types.CallSession) error {
	raw, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}
	_, err = r.db.Exec(`
		INSERT INTO call_sessions(id, nonce, state, expires_at, payload)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			nonce      = excluded.nonce,
			state      = excluded.state,
			expires_at = excluded.expires_at,
			payload    = excluded.payload
	`, s.ID, s.Nonce, string(s.State), s.ExpiresAt, string(raw))
	if err != nil {
		return fmt.Errorf("save session: %w", err)
	}
	return nil
}

// Load returns the session by ID or ErrNotFound.
func (r *Repo) Load(id string) (*types.CallSession, error) {
	var raw string
	err := r.db.QueryRow(`SELECT payload FROM call_sessions WHERE id = ?`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("load session %q: %w", id, err)
	}
	var s types.CallSession
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return nil, fmt.Errorf("unmarshal session %q: %w", id, err)
	}
	return &s, nil
}

// SaveTrace upserts a TrustTrace row keyed by session ID.
func (r *Repo) SaveTrace(t *types.TrustTrace) error {
	raw, err := json.Marshal(t)
	if err != nil {
		return fmt.Errorf("marshal trace: %w", err)
	}
	_, err = r.db.Exec(`
		INSERT INTO trace_logs(session_id, payload, created_at)
		VALUES (?, ?, ?)
		ON CONFLICT(session_id) DO UPDATE SET
			payload    = excluded.payload,
			created_at = excluded.created_at
	`, t.SessionID, string(raw), time.Now().UTC())
	if err != nil {
		return fmt.Errorf("save trace: %w", err)
	}
	return nil
}

// LoadTrace returns the trace for a session or nil if absent.
func (r *Repo) LoadTrace(sessionID string) (*types.TrustTrace, error) {
	var raw string
	err := r.db.QueryRow(`SELECT payload FROM trace_logs WHERE session_id = ?`, sessionID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("load trace %q: %w", sessionID, err)
	}
	var t types.TrustTrace
	if err := json.Unmarshal([]byte(raw), &t); err != nil {
		return nil, fmt.Errorf("unmarshal trace %q: %w", sessionID, err)
	}
	return &t, nil
}

// DeleteExpired marks sessions past expires_at as state=expired. Used
// by a background sweeper if added later. Returns count updated.
func (r *Repo) DeleteExpired(now time.Time) (int64, error) {
	res, err := r.db.Exec(`
		UPDATE call_sessions
		SET state = 'expired',
		    payload = json_set(payload, '$.state', 'expired')
		WHERE expires_at < ?
		  AND state NOT IN ('verified', 'blocked', 'expired')
	`, now)
	if err != nil {
		return 0, fmt.Errorf("delete expired: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}
