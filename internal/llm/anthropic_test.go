package llm

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// roundTripperFunc lets us mock http.Client.Transport without spinning
// up a real httptest server (faster, no port allocation).
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func buildResp(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

func TestAnthropic_SuccessReturnsJSON(t *testing.T) {
	c := NewAnthropic(AnthropicConfig{
		ApiKey: "test-key",
		RoundTripper: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			if r.Header.Get("x-api-key") == "" {
				t.Fatal("api key header missing")
			}
			return buildResp(200, `{"content":[{"type":"text","text":"{\"expected_requests\":[\"본인 명의 확인\"]}"}],"stop_reason":"end_turn"}`), nil
		}),
	})
	out, err := c.GenerateJSON(context.Background(), "test prompt", "schema")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(out), "expected_requests") {
		t.Fatalf("unexpected body: %s", string(out))
	}
}

func TestAnthropic_StripsJSONFence(t *testing.T) {
	c := NewAnthropic(AnthropicConfig{
		ApiKey: "test-key",
		RoundTripper: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return buildResp(200, "{\"content\":[{\"type\":\"text\",\"text\":\"```json\\n{\\\"x\\\":1}\\n```\"}]}"), nil
		}),
	})
	out, err := c.GenerateJSON(context.Background(), "p", "s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != `{"x":1}` {
		t.Fatalf("fence not stripped: %q", string(out))
	}
}

func TestAnthropic_Non200ReturnsError(t *testing.T) {
	c := NewAnthropic(AnthropicConfig{
		ApiKey: "test-key",
		RoundTripper: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return buildResp(429, `{"error":"rate limited"}`), nil
		}),
	})
	_, err := c.GenerateJSON(context.Background(), "p", "s")
	if err == nil {
		t.Fatal("expected error on 429")
	}
}

func TestAnthropic_TimeoutReturnsErrTimeout(t *testing.T) {
	c := NewAnthropic(AnthropicConfig{
		ApiKey:  "test-key",
		Timeout: 10 * time.Millisecond,
		RoundTripper: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			// Block until ctx fires.
			<-r.Context().Done()
			return nil, r.Context().Err()
		}),
	})
	_, err := c.GenerateJSON(context.Background(), "p", "s")
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("expected ErrTimeout, got %v", err)
	}
}

func TestAnthropic_EmptyContentReturnsInvalid(t *testing.T) {
	c := NewAnthropic(AnthropicConfig{
		ApiKey: "test-key",
		RoundTripper: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return buildResp(200, `{"content":[],"stop_reason":"end_turn"}`), nil
		}),
	})
	_, err := c.GenerateJSON(context.Background(), "p", "s")
	if !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("expected ErrInvalidResponse, got %v", err)
	}
}

func TestNoop_AlwaysReturnsNoProvider(t *testing.T) {
	c := NewNoop()
	_, err := c.GenerateJSON(context.Background(), "p", "s")
	if !errors.Is(err, ErrNoProvider) {
		t.Fatalf("expected ErrNoProvider, got %v", err)
	}
	if c.Name() != "noop" {
		t.Fatalf("unexpected name: %s", c.Name())
	}
}

func TestFactory_NoKeyReturnsNoop(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("LLM_PROVIDER", "")
	c := NewClient(LoadFactoryConfig())
	if c.Name() != "noop" {
		t.Fatalf("expected noop, got %s", c.Name())
	}
}

func TestFactory_WithKeyReturnsAnthropic(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-test")
	t.Setenv("LLM_PROVIDER", "")
	c := NewClient(LoadFactoryConfig())
	if !strings.HasPrefix(c.Name(), "anthropic/") {
		t.Fatalf("expected anthropic, got %s", c.Name())
	}
}
