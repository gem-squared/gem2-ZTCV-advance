package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Anthropic constants — keep small, fast, predictable for sub-4s SLA.
const (
	anthropicEndpoint = "https://api.anthropic.com/v1/messages"
	anthropicVersion  = "2023-06-01"
	defaultModel      = "claude-haiku-4-5-20251001"
	defaultMaxTokens  = 800
	defaultTimeout    = 4 * time.Second
)

// AnthropicClient calls Anthropic's Messages API and returns the raw
// text block, which is expected to be a JSON object the caller will
// unmarshal. Wrapper enforces a tight timeout and a structured error
// taxonomy (ErrTimeout / ErrInvalidResponse) so the caller's fallback
// path is always reachable.
type AnthropicClient struct {
	apiKey  string
	model   string
	timeout time.Duration
	http    *http.Client
}

// AnthropicConfig captures the env-driven configuration. Only ApiKey
// is required.
type AnthropicConfig struct {
	ApiKey       string
	Model        string        // optional override; default Claude Haiku 4.5
	Timeout      time.Duration // optional override; default 4s
	RoundTripper http.RoundTripper // optional injection for tests
}

// NewAnthropic constructs a client. Returns nil if no API key is set
// so the caller can fall back to NewNoop().
func NewAnthropic(cfg AnthropicConfig) *AnthropicClient {
	if strings.TrimSpace(cfg.ApiKey) == "" {
		return nil
	}
	model := cfg.Model
	if model == "" {
		model = defaultModel
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	rt := cfg.RoundTripper
	if rt == nil {
		rt = http.DefaultTransport
	}
	return &AnthropicClient{
		apiKey:  cfg.ApiKey,
		model:   model,
		timeout: timeout,
		http:    &http.Client{Timeout: timeout, Transport: rt},
	}
}

// Name returns "anthropic/<model>" for telemetry.
func (a *AnthropicClient) Name() string {
	return "anthropic/" + a.model
}

// anthropicRequest matches the Messages API request shape we need.
type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicResponse matches the relevant subset of the response.
type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
}

// GenerateJSON sends the prompt and returns the first text block as
// raw bytes. The caller validates against its own schema; we only
// guarantee the bytes are non-empty.
func (a *AnthropicClient) GenerateJSON(ctx context.Context, prompt string, schemaName string) ([]byte, error) {
	// Apply a context timeout in addition to http.Client.Timeout so
	// callers passing their own context still respect our budget.
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	body, err := json.Marshal(anthropicRequest{
		Model:     a.model,
		MaxTokens: defaultMaxTokens,
		System:    "You return ONLY a single JSON object matching the requested schema. No prose, no markdown fences, no commentary.",
		Messages: []anthropicMessage{
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("llm: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("llm: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)

	resp, err := a.http.Do(req)
	if err != nil {
		// Distinguish timeout from generic network error so callers
		// can log telemetry — semantically both trigger fallback.
		if ctx.Err() == context.DeadlineExceeded {
			return nil, ErrTimeout
		}
		return nil, fmt.Errorf("llm: transport: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Drain a bounded slice so 4xx/5xx don't leak verbose vendor
		// error bodies into our logs.
		const maxErr = 512
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, maxErr))
		return nil, fmt.Errorf("llm: provider status %d: %s", resp.StatusCode, string(bytes.TrimSpace(errBody)))
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("llm: read body: %w", err)
	}

	var parsed anthropicResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("llm: parse envelope: %w", err)
	}
	if len(parsed.Content) == 0 {
		return nil, ErrInvalidResponse
	}
	// Concatenate all text-type blocks. In practice Haiku returns one
	// block per call when System message constrains it; we tolerate
	// multiple for robustness.
	var out bytes.Buffer
	for _, c := range parsed.Content {
		if c.Type == "text" {
			out.WriteString(c.Text)
		}
	}
	trimmed := bytes.TrimSpace(out.Bytes())
	if len(trimmed) == 0 {
		return nil, ErrInvalidResponse
	}
	// Strip optional ```json ... ``` fences in case the model adds them
	// despite the System instruction.
	trimmed = stripJSONFence(trimmed)
	return trimmed, nil
}

// stripJSONFence removes leading ```json (or ```) and trailing ``` if
// present, returning the inner JSON. Idempotent on already-clean
// bytes.
func stripJSONFence(b []byte) []byte {
	s := string(b)
	if strings.HasPrefix(s, "```") {
		// Drop the first line (the opening fence).
		if nl := strings.IndexByte(s, '\n'); nl >= 0 {
			s = s[nl+1:]
		}
	}
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}
	return []byte(s)
}
