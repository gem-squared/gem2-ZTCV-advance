package intent

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gem-squared/gem2-ZTCV/internal/llm"
)

// fakeClient lets us drive the LLM client path without HTTP.
type fakeClient struct {
	name string
	out  []byte
	err  error
}

func (f *fakeClient) GenerateJSON(ctx context.Context, prompt string, schemaName string) ([]byte, error) {
	return f.out, f.err
}
func (f *fakeClient) Name() string { return f.name }

func fixedNow() time.Time { return time.Date(2026, 5, 29, 14, 0, 0, 0, time.UTC) }

func TestFallback_Safe(t *testing.T) {
	g := &generator{client: llm.NewNoop(), now: fixedNow}
	m := g.Generate(context.Background(), Input{Outcome: OutcomeSafe})
	if m.Source != "fallback" {
		t.Fatalf("expected fallback, got %s", m.Source)
	}
	if !strings.Contains(m.SafetySummary, "본인 명의 확인") {
		t.Fatalf("unexpected summary: %s", m.SafetySummary)
	}
	if len(m.ForbiddenRequests) == 0 {
		t.Fatal("safe scenario must have forbidden_requests populated")
	}
}

func TestFallback_UnknownDID(t *testing.T) {
	g := &generator{client: llm.NewNoop(), now: fixedNow}
	m := g.Generate(context.Background(), Input{Outcome: OutcomeUnknownDID})
	if m.Source != "fallback" {
		t.Fatalf("expected fallback, got %s", m.Source)
	}
	if len(m.ExpectedRequests) != 0 {
		t.Fatalf("FAILED scenario must have empty expected_requests, got %v", m.ExpectedRequests)
	}
	if !strings.Contains(m.SafetySummary, "권장하지 않습니다") {
		t.Fatalf("unexpected summary: %s", m.SafetySummary)
	}
}

func TestFallback_UnauthorizedScope(t *testing.T) {
	g := &generator{client: llm.NewNoop(), now: fixedNow}
	m := g.Generate(context.Background(), Input{Outcome: OutcomeUnauthorizedScope})
	if m.Source != "fallback" {
		t.Fatalf("expected fallback, got %s", m.Source)
	}
	if !strings.Contains(m.SafetySummary, "응답하지 마십시오") {
		t.Fatalf("unexpected summary: %s", m.SafetySummary)
	}
}

func TestLive_ValidJSON(t *testing.T) {
	g := &generator{
		client: &fakeClient{
			name: "anthropic/claude-haiku-4-5-test",
			out: []byte(`{
				"expected_requests": ["본인 명의 확인", "상담 동의 확인"],
				"forbidden_requests": ["계좌번호", "비밀번호"],
				"safety_summary": "상담 통화입니다. 계좌번호와 비밀번호는 요청되지 않습니다."
			}`),
		},
		now: fixedNow,
	}
	m := g.Generate(context.Background(), Input{Outcome: OutcomeSafe})
	if m.Source != "live" {
		t.Fatalf("expected live, got %s", m.Source)
	}
	if !strings.HasPrefix(m.Provider, "anthropic/") {
		t.Fatalf("expected provider anthropic/*, got %s", m.Provider)
	}
	if len(m.ExpectedRequests) != 2 {
		t.Fatalf("expected 2 expected_requests, got %d", len(m.ExpectedRequests))
	}
}

func TestLive_TimeoutFallsBack(t *testing.T) {
	g := &generator{
		client: &fakeClient{name: "anthropic/test", err: llm.ErrTimeout},
		now:    fixedNow,
	}
	m := g.Generate(context.Background(), Input{Outcome: OutcomeSafe})
	if m.Source != "fallback" {
		t.Fatalf("timeout MUST fall back, got %s", m.Source)
	}
}

func TestLive_InvalidJSONFallsBack(t *testing.T) {
	g := &generator{
		client: &fakeClient{name: "anthropic/test", out: []byte(`not json at all`)},
		now:    fixedNow,
	}
	m := g.Generate(context.Background(), Input{Outcome: OutcomeSafe})
	if m.Source != "fallback" {
		t.Fatalf("parse failure MUST fall back, got %s", m.Source)
	}
}

func TestLive_RunawayOutputRejected(t *testing.T) {
	// 13 items > maxItems(12) → must reject and fall back.
	big := `["a","a","a","a","a","a","a","a","a","a","a","a","a"]`
	g := &generator{
		client: &fakeClient{name: "anthropic/test", out: []byte(`{
			"expected_requests": ` + big + `,
			"forbidden_requests": [],
			"safety_summary": "x"
		}`)},
		now: fixedNow,
	}
	m := g.Generate(context.Background(), Input{Outcome: OutcomeSafe})
	if m.Source != "fallback" {
		t.Fatalf("runaway output MUST fall back, got %s", m.Source)
	}
}

func TestHash_DeterministicAcrossReplays(t *testing.T) {
	m := scenarioFallback(Input{Outcome: OutcomeSafe})
	m.Source = "fallback"
	h1 := Hash(m)
	h2 := Hash(m)
	if h1 == "" || h2 == "" {
		t.Fatal("hash empty")
	}
	if h1 != h2 {
		t.Fatalf("hash not deterministic: %s vs %s", h1, h2)
	}
	if !strings.HasPrefix(h1, "0x") || len(h1) != 66 {
		t.Fatalf("unexpected hash shape: %s", h1)
	}
}
