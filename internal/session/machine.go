// Package session contains the CallSession state machine and SQLite
// repository. State transitions are pure functions; the repository
// handles persistence.
package session

import (
	"errors"
	"time"

	"github.com/gem-squared/gem2-ZTCV/internal/types"
)

// ErrInvalidTransition is returned when a state-machine call would
// move the session into an undefined state from its current one.
var ErrInvalidTransition = errors.New("invalid state transition")

// DefaultExpiry is the lifetime of a session nonce. After this window
// any proof submission is rejected with state=expired.
const DefaultExpiry = 5 * time.Minute

// Create returns a fresh CallSession in state=created.
func Create(id, nonce string, now time.Time) *types.CallSession {
	return &types.CallSession{
		ID:        id,
		Nonce:     nonce,
		State:     types.StateCreated,
		CreatedAt: now,
		ExpiresAt: now.Add(DefaultExpiry),
		UpdatedAt: now,
	}
}

// SubmitCallerProof transitions the session to caller_proved if the
// current state is created. The actual proof verification happens at
// the handler layer; this only enforces the transition rule.
func SubmitCallerProof(s *types.CallSession, proof *types.CallerProof, now time.Time) (*types.CallSession, error) {
	if s.State != types.StateCreated {
		return nil, ErrInvalidTransition
	}
	out := *s
	out.CallerProof = proof
	out.State = types.StateCallerProved
	out.UpdatedAt = now
	return &out, nil
}

// SubmitCustomerProof transitions the session to customer_proved.
// Allowed from caller_proved (canonical) or created (out-of-order
// customer-first flow — also valid in our design).
func SubmitCustomerProof(s *types.CallSession, proof *types.CustomerProof, now time.Time) (*types.CallSession, error) {
	if s.State != types.StateCallerProved && s.State != types.StateCreated {
		return nil, ErrInvalidTransition
	}
	out := *s
	out.CustomerProof = proof
	out.State = types.StateCustomerProved
	out.UpdatedAt = now
	return &out, nil
}

// MarkRiskChecked attaches the composed verdict and transitions to
// risk_checked. Allowed from customer_proved.
func MarkRiskChecked(s *types.CallSession, verdict *types.ComposedVerdict, now time.Time) (*types.CallSession, error) {
	if s.State != types.StateCustomerProved {
		return nil, ErrInvalidTransition
	}
	out := *s
	out.RiskVerdict = verdict
	out.State = types.StateRiskChecked
	out.UpdatedAt = now
	return &out, nil
}

// MarkAnchored attaches the receipt + tx hash and transitions to
// anchored. The final terminal state (verified or blocked) is set by
// MarkVerified or MarkBlocked.
func MarkAnchored(s *types.CallSession, receipt *types.Receipt, txHash string, now time.Time) (*types.CallSession, error) {
	if s.State != types.StateRiskChecked {
		return nil, ErrInvalidTransition
	}
	out := *s
	out.Receipt = receipt
	out.TxHash = txHash
	out.State = types.StateAnchored
	out.UpdatedAt = now
	return &out, nil
}

// MarkVerified marks a session as fully verified (Safe Call). Allowed
// from anchored.
func MarkVerified(s *types.CallSession, now time.Time) (*types.CallSession, error) {
	if s.State != types.StateAnchored {
		return nil, ErrInvalidTransition
	}
	out := *s
	out.State = types.StateVerified
	out.UpdatedAt = now
	return &out, nil
}

// MarkBlocked is a terminal transition from any non-terminal state.
// The reason is stored on the session for the CallPassport block-reason
// field. The receipt of the BLOCK decision is still anchored (the
// "차단 영수증 기록" stamp).
func MarkBlocked(s *types.CallSession, reason string, receipt *types.Receipt, txHash string, now time.Time) (*types.CallSession, error) {
	if s.State == types.StateVerified || s.State == types.StateBlocked || s.State == types.StateExpired {
		return nil, ErrInvalidTransition
	}
	out := *s
	out.BlockReason = reason
	out.Receipt = receipt
	out.TxHash = txHash
	out.State = types.StateBlocked
	out.UpdatedAt = now
	return &out, nil
}

// Expire transitions to terminal expired.
func Expire(s *types.CallSession, now time.Time) *types.CallSession {
	out := *s
	out.State = types.StateExpired
	out.UpdatedAt = now
	return &out
}

// IsExpired reports whether the session nonce window has passed.
func IsExpired(s *types.CallSession, now time.Time) bool {
	return now.After(s.ExpiresAt) && s.State != types.StateVerified && s.State != types.StateBlocked
}
