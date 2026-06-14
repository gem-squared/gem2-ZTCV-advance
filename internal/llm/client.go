// Package llm provides a provider-agnostic interface for structured
// JSON generation calls used by AI-driven verification steps
// (Predictive Disclosure today; Layer 2 risk verdict in a follow-on WP).
//
// Discipline (per WP-ZTCV-06):
//   - Never log API keys.
//   - Never crash on provider failure — return an error and let the caller
//     fall back to a deterministic mock.
//   - Default to fast, cheap models (Claude Haiku class) for sub-4s SLAs.
package llm

import (
	"context"
	"errors"
)

// ErrNoProvider is returned by the no-op client when no provider is
// configured. Callers should treat this as a signal to use their
// deterministic fallback path.
var ErrNoProvider = errors.New("llm: no provider configured")

// ErrTimeout is returned when the provider call exceeds the configured
// budget. Callers should treat this identically to ErrNoProvider.
var ErrTimeout = errors.New("llm: provider timed out")

// ErrInvalidResponse is returned when the provider returns content that
// does not match the requested JSON schema. Callers fall back.
var ErrInvalidResponse = errors.New("llm: response did not satisfy schema")

// Client is the provider-agnostic interface every concrete LLM client
// implements. The contract is intentionally narrow — one synchronous
// structured-JSON generation call with a tight budget.
type Client interface {
	// GenerateJSON sends the prompt to the provider and returns raw
	// JSON bytes that the caller will unmarshal. The caller specifies
	// the expected schema name purely for telemetry; validation is the
	// caller's responsibility (we keep this package transport-only).
	GenerateJSON(ctx context.Context, prompt string, schemaName string) ([]byte, error)

	// Name returns a stable identifier of the provider (e.g.
	// "anthropic/claude-haiku-4-5" or "noop"). Used in passport
	// `intent_handshake.source` telemetry and structured logs.
	Name() string
}
