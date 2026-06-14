// Package intent implements the Intent Handshake step (Step 7 of the
// 9-step verification pipeline). Today this package ships AI-driven
// Predictive Disclosure with a deterministic rule-based fallback so
// the demo never breaks when the LLM is unreachable. Counter-Honeypot
// Challenge remains a finals-stage deliverable and is intentionally
// NOT implemented here.
package intent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gem-squared/gem2-ZTCV/internal/llm"
	"github.com/gem-squared/gem2-ZTCV/internal/types"
)

// Manifest is an alias for types.IntentManifest — the canonical wire
// type lives in package types alongside CallPassport so the
// frontend ↔ backend contract is owned by one package. We keep the
// short name here for ergonomic use inside the intent package and the
// LLM prompt code.
type Manifest = types.IntentManifest

// Outcome identifies which scenario branch the upstream pipeline took
// before reaching Step 7. The fallback uses this to choose its
// per-scenario script.
type Outcome string

const (
	OutcomeSafe              Outcome = "SAFE"
	OutcomeUnknownDID        Outcome = "FAILED_UNKNOWN_DID"
	OutcomeUnauthorizedScope Outcome = "FAILED_UNAUTHORIZED_PURPOSE"
)

// Input is the upstream signal the generator needs. Kept narrow so the
// LLM prompt stays short and the deterministic fallback stays simple.
type Input struct {
	CallerDID          string
	OrgDisplayName     string   // e.g. "카카오뱅크"
	Purpose            string   // declared purpose, e.g. "loan_consultation"
	VCAllowedPurposes  []string // for prompt context
	RiskScore          float64  // 0..1, from Layer 2
	Outcome            Outcome
}

// Generator produces a Manifest. Implementations MUST always return a
// usable Manifest (live or fallback) and MUST NOT propagate provider
// errors to the caller. Errors trigger fallback internally.
type Generator interface {
	Generate(ctx context.Context, in Input) Manifest
}

// generator is the canonical implementation. It calls the LLM client
// and falls back to deterministic per-scenario output on any error.
type generator struct {
	client llm.Client
	now    func() time.Time
}

// New constructs a Generator wired to the given LLM client. Pass
// llm.NewNoop() to force the deterministic fallback path (used by
// tests and by deployments without a key).
func New(client llm.Client) Generator {
	if client == nil {
		client = llm.NewNoop()
	}
	return &generator{client: client, now: time.Now}
}

// Generate runs the Intent Handshake step. Discipline:
//   - Always return a non-empty Manifest.
//   - Set Source="live" only on a successful, schema-valid LLM call.
//   - Set Source="fallback" on any provider error, timeout, or
//     schema violation.
func (g *generator) Generate(ctx context.Context, in Input) Manifest {
	now := g.now().UTC()
	if raw, err := g.client.GenerateJSON(ctx, buildPrompt(in), "intent_manifest"); err == nil {
		if m, ok := parseAndValidate(raw); ok {
			m.Source = "live"
			m.Provider = g.client.Name()
			m.GeneratedAt = now
			return m
		}
	}
	// Deterministic fallback path.
	m := scenarioFallback(in)
	m.Source = "fallback"
	m.Provider = g.client.Name()
	m.GeneratedAt = now
	return m
}

// buildPrompt produces the user-message body for the LLM. The System
// message (set by the provider client) already enforces JSON-only.
func buildPrompt(in Input) string {
	allowed := strings.Join(in.VCAllowedPurposes, ", ")
	if allowed == "" {
		allowed = "(none recorded)"
	}
	return fmt.Sprintf(`Generate an Intent Handshake Manifest for a pre-call verification system in South Korea.
The receiver is a private citizen; the caller is the institution below. Output STRICT JSON matching this schema:

{
  "expected_requests":  string[]  // Korean phrases of what the caller is LIKELY to ask the receiver
  "forbidden_requests": string[]  // Korean phrases of what a legitimate caller of this purpose will NEVER ask (sensitive items)
  "safety_summary":     string    // ONE Korean sentence the receiver can read before answering, in formal honorific style (합니다체)
}

Constraints:
- Korean only. Honorific 합니다체 endings.
- 4-7 items per list, concise nouns/short phrases.
- safety_summary must be one sentence (≤80 chars) advising the receiver.
- If outcome is FAILED, expected_requests should be empty and safety_summary must warn the receiver clearly.
- Do not invent institution-specific knowledge; ground only in caller_did, org, purpose.

Caller context:
- caller_did:           %s
- organization:         %s
- declared_purpose:     %s
- vc_allowed_purposes:  %s
- risk_score (0..1):    %.2f
- outcome:              %s
`,
		safe(in.CallerDID), safe(in.OrgDisplayName),
		safe(in.Purpose), safe(allowed),
		in.RiskScore, safe(string(in.Outcome)))
}

func safe(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(unknown)"
	}
	return s
}

// parseAndValidate accepts raw LLM bytes and returns a Manifest only
// when all required fields are present and within sanity bounds. The
// fallback is triggered on every other case.
func parseAndValidate(raw []byte) (Manifest, bool) {
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return Manifest{}, false
	}
	// Sanity bounds — drop hostile / runaway outputs.
	const maxItems = 12
	const maxLen = 120
	if len(m.ExpectedRequests) > maxItems || len(m.ForbiddenRequests) > maxItems {
		return Manifest{}, false
	}
	if len(m.SafetySummary) > 200 {
		return Manifest{}, false
	}
	for _, s := range m.ExpectedRequests {
		if len(s) == 0 || len(s) > maxLen {
			return Manifest{}, false
		}
	}
	for _, s := range m.ForbiddenRequests {
		if len(s) == 0 || len(s) > maxLen {
			return Manifest{}, false
		}
	}
	if strings.TrimSpace(m.SafetySummary) == "" {
		return Manifest{}, false
	}
	// Trim each item; sort for canonical output before hashing.
	m.ExpectedRequests = trimAndSort(m.ExpectedRequests)
	m.ForbiddenRequests = trimAndSort(m.ForbiddenRequests)
	m.SafetySummary = strings.TrimSpace(m.SafetySummary)
	return m, true
}

func trimAndSort(in []string) []string {
	if len(in) == 0 {
		return in
	}
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

// scenarioFallback returns the documented per-scenario manifest from
// WP-ZTCV-06 §Unit 2. Output is deterministic byte-for-byte so chain
// receipts derived from it are reproducible across replays.
func scenarioFallback(in Input) Manifest {
	switch in.Outcome {
	case OutcomeSafe:
		return Manifest{
			ExpectedRequests:  []string{"본인 명의 확인", "상담 의향 확인"},
			ForbiddenRequests: []string{"계좌번호", "비밀번호", "안전계좌 송금", "타인 계좌 이체"},
			SafetySummary:     "본인 명의 확인을 위한 상담 통화입니다. 계좌번호 및 비밀번호는 요청하지 않습니다.",
		}
	case OutcomeUnknownDID:
		return Manifest{
			ExpectedRequests:  []string{},
			ForbiddenRequests: []string{"계좌 정보", "송금 안내", "수사 협조 요청"},
			SafetySummary:     "발신자의 신원이 확인되지 않습니다. 통화 응답을 권장하지 않습니다.",
		}
	case OutcomeUnauthorizedScope:
		return Manifest{
			ExpectedRequests:  []string{"등록 목적 외 요청 가능성"},
			ForbiddenRequests: []string{"대출 가입 유도", "금융상품 권유", "추가 계좌 개설 요청"},
			SafetySummary:     "발신자의 등록 목적과 다른 요청이 예상됩니다. 응답하지 마십시오.",
		}
	default:
		// Conservative default — treat as block-suspect.
		return Manifest{
			ExpectedRequests:  []string{},
			ForbiddenRequests: []string{"금융·신원 정보 일체"},
			SafetySummary:     "발신자 검증이 완전하지 않습니다. 응답에 주의하십시오.",
		}
	}
}

// CanonicalJSON returns a deterministic JSON encoding of the manifest
// suitable for hashing. Field order is the struct order; slice
// contents are pre-sorted by parseAndValidate or scenarioFallback.
func CanonicalJSON(m Manifest) ([]byte, error) {
	// Re-marshal via an intermediate map so we control key order
	// independently from struct tags. Using a single struct keeps Go's
	// emitter deterministic for our fields.
	type stable struct {
		ExpectedRequests  []string `json:"expected_requests"`
		ForbiddenRequests []string `json:"forbidden_requests"`
		SafetySummary     string   `json:"safety_summary"`
		Source            string   `json:"source"`
	}
	return json.Marshal(stable{
		ExpectedRequests:  m.ExpectedRequests,
		ForbiddenRequests: m.ForbiddenRequests,
		SafetySummary:     m.SafetySummary,
		Source:            m.Source,
	})
}

// Hash returns the sha256 hex of the canonical JSON, prefixed "0x".
// Used as the on-chain intent_manifest_hash. Returns "" on error so
// the caller can omit the field cleanly.
func Hash(m Manifest) string {
	raw, err := CanonicalJSON(m)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return "0x" + hex.EncodeToString(sum[:])
}
