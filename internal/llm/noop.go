package llm

import "context"

// NoopClient always returns ErrNoProvider. It is wired when no
// provider env vars are set, so the caller's fallback path is always
// exercised.
type NoopClient struct{}

// NewNoop constructs a Noop client.
func NewNoop() *NoopClient { return &NoopClient{} }

// GenerateJSON always errors with ErrNoProvider. The caller MUST handle
// this by falling back to a deterministic generator.
func (NoopClient) GenerateJSON(ctx context.Context, prompt string, schemaName string) ([]byte, error) {
	return nil, ErrNoProvider
}

// Name returns the stable identifier of the no-op provider.
func (NoopClient) Name() string { return "noop" }
