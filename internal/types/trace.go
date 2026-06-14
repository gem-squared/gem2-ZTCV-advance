package types

import "time"

// GateVerdict is the per-gate result for a single TrustTrace step. PASS lets
// the next gate run; FAIL short-circuits the trace to BLOCK; SKIPPED means
// the gate was not reached (typical for Scenario 3 where L1 fails early and
// L2/L3 never run).
type GateVerdict string

const (
	GatePASS    GateVerdict = "PASS"
	GateFAIL    GateVerdict = "FAIL"
	GateSKIPPED GateVerdict = "SKIPPED"
)

// GateResult captures a single gate's outcome with reasoning. The Reasons
// list feeds the frontend TrustTraceTimeline drawer detail.
type GateResult struct {
	Gate        string      `json:"gate"` // "L0"|"L1"|"F"|"L2"|"L3"
	Verdict     GateVerdict `json:"verdict"`
	Reasons     []string    `json:"reasons"`
	StartedAt   time.Time   `json:"started_at"`
	CompletedAt time.Time   `json:"completed_at,omitempty"`
}

// TrustTrace is the full orchestration log for one verification session. It
// is what the SSE stream emits (one GateResult event per gate completion)
// and what the admin dashboard reads to render the timeline.
//
// Gate semantics:
//
//	L0  Request Safety       — Lobster-Trap regex + LLM intent check on input
//	L1  Identity Grounding   — didregistry resolve + caller proof + customer proof
//	F   Risk Verdict         — risk-chain-svc composed Layer 1 + Layer 2
//	L2  Policy & Risk Check  — composed verdict vs policy thresholds
//	L3  Explanation Guard    — scrub outgoing explanation text for prompt-injection echo / overclaim
type TrustTrace struct {
	SessionID string       `json:"session_id"`
	Gates     []GateResult `json:"gates"`
	Final     GateVerdict  `json:"final"` // derived: any FAIL → FAIL, else PASS
	StartedAt time.Time    `json:"started_at"`
	EndedAt   time.Time    `json:"ended_at,omitempty"`
}
