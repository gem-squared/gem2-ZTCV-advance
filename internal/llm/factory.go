package llm

import (
	"os"
	"strconv"
	"time"
)

// FactoryConfig is read from env at process start. The caller passes
// the resulting Client to anything that needs structured JSON
// generation (Predictive Disclosure today; Layer 2 risk verdict in a
// follow-on WP).
type FactoryConfig struct {
	// Provider selects which client to construct. Empty / "noop"
	// returns a NoopClient so the caller's fallback is always used.
	Provider string

	// AnthropicAPIKey is read from ANTHROPIC_API_KEY env. Empty value
	// downgrades to noop regardless of Provider.
	AnthropicAPIKey string

	// Model overrides the default model (claude-haiku-4-5-20251001).
	Model string

	// Timeout in milliseconds (LLM_TIMEOUT_MS env).
	TimeoutMillis int
}

// LoadFactoryConfig reads env vars and returns a populated
// FactoryConfig. No I/O beyond env reads.
func LoadFactoryConfig() FactoryConfig {
	cfg := FactoryConfig{
		Provider:        os.Getenv("LLM_PROVIDER"),
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
		Model:           os.Getenv("LLM_MODEL"),
	}
	if v := os.Getenv("LLM_TIMEOUT_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.TimeoutMillis = n
		}
	}
	return cfg
}

// NewClient constructs the appropriate Client. Always returns a usable
// non-nil Client (NoopClient when no provider configured) so callers
// never need to nil-check.
func NewClient(cfg FactoryConfig) Client {
	switch cfg.Provider {
	case "", "noop":
		// Even when provider unset, an API key alone is a strong
		// signal of intent; honor it.
		if cfg.AnthropicAPIKey != "" {
			return newAnthropicOrNoop(cfg)
		}
		return NewNoop()
	case "anthropic":
		return newAnthropicOrNoop(cfg)
	default:
		return NewNoop()
	}
}

func newAnthropicOrNoop(cfg FactoryConfig) Client {
	a := NewAnthropic(AnthropicConfig{
		ApiKey:  cfg.AnthropicAPIKey,
		Model:   cfg.Model,
		Timeout: time.Duration(cfg.TimeoutMillis) * time.Millisecond,
	})
	if a == nil {
		return NewNoop()
	}
	return a
}
