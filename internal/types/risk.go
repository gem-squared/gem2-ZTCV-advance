package types

import "time"

// RiskVerdict is the verdict ladder. Conservative composition rule (see
// ComposedVerdict): final = max(L1, L2) on this ladder ordering.
type RiskVerdict string

const (
	RiskLOW    RiskVerdict = "LOW"
	RiskMEDIUM RiskVerdict = "MEDIUM"
	RiskHIGH   RiskVerdict = "HIGH"
	RiskBLOCK  RiskVerdict = "BLOCK"
)

// TriggeredRule is a single rule firing from the Layer 1 classifier with the
// reason (Korean phrase or pattern label) and severity. Rule 4 is the
// DID-purpose mismatch — it fires for Scenario 3 before any LLM call.
type TriggeredRule struct {
	RuleID   string      `json:"rule_id"`  // e.g., "RULE_1_TRANSFER_DEMAND" | "RULE_4_PURPOSE_MISMATCH"
	Reason   string      `json:"reason"`   // human-readable trigger description
	Severity RiskVerdict `json:"severity"` // contribution to verdict
}

// Layer1Result is the deterministic rule-classifier output. 100% reproducible
// given identical input — this is the demo-safety floor. Empty TriggeredRules
// with verdict LOW means no Korean phishing pattern matched and purpose was
// authorized.
type Layer1Result struct {
	Verdict        RiskVerdict     `json:"verdict"`
	TriggeredRules []TriggeredRule `json:"triggered_rules"`
	Reasons        []string        `json:"reasons"`
	EvaluatedAt    time.Time       `json:"evaluated_at"`
}

// Layer2Result is the LLM-evaluation output (gem2-epistemic-engine wrapper
// or mock). KoreanExplanation is always populated — either by the LLM or by
// a deterministic template on timeout/fallback.
type Layer2Result struct {
	Verdict           RiskVerdict `json:"verdict"`
	RiskScore         float64     `json:"risk_score"` // 0.0..1.0
	Reasons           []string    `json:"reasons"`
	KoreanExplanation string      `json:"korean_explanation"`
	TimedOut          bool        `json:"timed_out"`
	UsedMockProvider  bool        `json:"used_mock_provider"`
	EvaluatedAt       time.Time   `json:"evaluated_at"`
}

// ComposedVerdict is the conservative-block composition: BLOCK if EITHER
// layer says BLOCK; SAFE-equivalent (LOW final) only if BOTH say LOW.
// Disagreement is recorded for audit (admin dashboard exposes /api/risk/
// disagreement-stats endpoint).
type ComposedVerdict struct {
	Layer1        Layer1Result `json:"layer1"`
	Layer2        Layer2Result `json:"layer2"`
	Final         RiskVerdict  `json:"final"`
	Disagreement  bool         `json:"disagreement"`
	PolicyVersion string       `json:"policy_version"` // hex of keccak256("ztcv-policy-v0.1.0")
	ComposedAt    time.Time    `json:"composed_at"`
}
