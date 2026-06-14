// Package layer2 wraps the LLM-based semantic risk evaluator. In
// simulation mode we run the deterministic Mock. The real engine
// (gem2-epistemic-engine via REST) lands when an API key is available.
package layer2

import (
	"strings"
	"time"

	"github.com/gem-squared/gem2-ZTCV/internal/types"
)

// Mock returns scenario-keyed responses so the demo is deterministic.
// Keyed on substring matches against caller_did + purpose + script.
type Mock struct{}

// NewMock returns the deterministic mock engine.
func NewMock() *Mock { return &Mock{} }

// Evaluate produces a Layer2Result based on coarse heuristics. The
// goal is to produce a verdict + Korean explanation that visually
// agrees with what a human would conclude, not to be a real LLM.
func (m *Mock) Evaluate(in Input) types.Layer2Result {
	now := time.Now().UTC()
	score := 0.0
	reasons := []string{}

	if in.CallerDID == "" {
		score += 0.7
		reasons = append(reasons, "발신자 DID가 등록되지 않음")
	}
	if strings.Contains(in.CallScript, "송금") || strings.Contains(in.CallScript, "안전계좌") {
		score += 0.6
		reasons = append(reasons, "송금/안전계좌 요청은 보이스피싱 전형 패턴")
	}
	if !in.IsAuthorized && in.CallerDID != "" {
		score += 0.5
		reasons = append(reasons, "권한 외 목적 통화 — 사용자 보호를 위해 보수적으로 BLOCK 권고")
	}
	if strings.Contains(in.CallScript, "체포") || strings.Contains(in.CallScript, "수사") {
		score += 0.4
		reasons = append(reasons, "기관 사칭 + 위협 표현 결합")
	}

	verdict := types.RiskLOW
	switch {
	case score >= 0.7:
		verdict = types.RiskBLOCK
	case score >= 0.5:
		verdict = types.RiskHIGH
	case score >= 0.3:
		verdict = types.RiskMEDIUM
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "통화 의도 및 발신자 권한이 정상 범위로 평가됨")
	}

	korean := strings.Join(reasons, " · ")
	if verdict == types.RiskBLOCK {
		korean = "[BLOCK] " + korean
	} else if verdict == types.RiskLOW {
		korean = "[SAFE] " + korean
	}

	return types.Layer2Result{
		Verdict:           verdict,
		RiskScore:         clamp01(score),
		Reasons:           reasons,
		KoreanExplanation: korean,
		TimedOut:          false,
		UsedMockProvider:  true,
		EvaluatedAt:       now,
	}
}

// Input is what the wrapper feeds the engine (mock or real).
type Input struct {
	CallerDID         string
	OrgDID            string
	Purpose           string
	IsAuthorized      bool
	AuthorizedPurpose []string
	CallScript        string
	MobileIDVerified  bool
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}
