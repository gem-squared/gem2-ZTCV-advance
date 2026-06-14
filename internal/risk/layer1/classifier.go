// Package layer1 is the deterministic rule-based risk classifier.
// It runs ALWAYS — Layer 2 LLM is advisory. The composer (parent
// package) combines L1 + L2 via conservative-block.
//
// Six rules:
//
//	Rule 1  송금요구             — transfer demand keywords (송금, 입금, 이체)
//	Rule 2  안전계좌 phrase       — fraud pattern
//	Rule 3  URL/account in script — URLs or naked account numbers in call_purpose
//	Rule 4  DID-purpose mismatch  — caller authorized purpose ≠ requested purpose
//	                               *** THIS IS THE SCENARIO 3 BLOCK TRIGGER ***
//	Rule 5  기관사칭 pattern       — caller_org claim conflicts with caller_did org
//	Rule 6  긴급성 협박            — urgency + threat phrasing
package layer1

import (
	"regexp"
	"strings"
	"time"

	"github.com/gem-squared/gem2-ZTCV/internal/types"
)

// Input is the bundle of session context the classifier evaluates.
// Optional fields can be empty without breaking any rule.
type Input struct {
	CallerDID         string
	OrgDIDClaim       string
	Purpose           string
	AuthorizedPurpose []string // from didregistry — empty if unknown_did
	IsAuthorized      bool     // did the auth lookup succeed for this purpose
	CallScript        string   // free text from the institution / agent
	RequestedAction   string   // optional structured action label
}

// Rule represents one classifier rule.
type Rule struct {
	ID       string
	Severity types.RiskVerdict
	Run      func(in Input) (triggered bool, reason string)
}

// allRules is exported via Classify; declared here so tests can target
// individual rules.
var allRules = []Rule{
	{
		ID:       "RULE_1_TRANSFER_DEMAND",
		Severity: types.RiskBLOCK,
		Run: func(in Input) (bool, string) {
			words := []string{"송금", "입금", "이체", "보내주세요", "송금해", "이체해"}
			lower := in.CallScript + " " + in.RequestedAction
			for _, w := range words {
				if strings.Contains(lower, w) {
					return true, "송금/이체 요구 키워드 감지: " + w
				}
			}
			return false, ""
		},
	},
	{
		ID:       "RULE_2_SAFE_ACCOUNT",
		Severity: types.RiskBLOCK,
		Run: func(in Input) (bool, string) {
			phrases := []string{"안전계좌", "안전 계좌", "지정계좌", "안전한 계좌"}
			text := in.CallScript + " " + in.RequestedAction
			for _, p := range phrases {
				if strings.Contains(text, p) {
					return true, "안전계좌 패턴 감지: " + p
				}
			}
			return false, ""
		},
	},
	{
		ID:       "RULE_3_URL_OR_ACCOUNT",
		Severity: types.RiskHIGH,
		Run: func(in Input) (bool, string) {
			urlRe := regexp.MustCompile(`https?://[^\s]+`)
			accountRe := regexp.MustCompile(`\b\d{3,}-?\d{2,}-?\d{4,}\b`) // loose Korean bank acct pattern
			if urlRe.MatchString(in.CallScript) {
				return true, "통화 스크립트 내 URL 감지"
			}
			if accountRe.MatchString(in.CallScript) {
				return true, "통화 스크립트 내 계좌번호 패턴 감지"
			}
			return false, ""
		},
	},
	{
		// THE SCENARIO 3 RULE — fires before any LLM call.
		ID:       "RULE_4_PURPOSE_MISMATCH",
		Severity: types.RiskBLOCK,
		Run: func(in Input) (bool, string) {
			if in.CallerDID == "" {
				return false, "" // unknown DID falls under Rule 1/2 or didregistry layer
			}
			if !in.IsAuthorized {
				return true, "권한 외 목적 — agent " + in.CallerDID + " 는 purpose=\"" + in.Purpose + "\" 권한 없음 (보유 권한: " + strings.Join(in.AuthorizedPurpose, ",") + ")"
			}
			return false, ""
		},
	},
	{
		ID:       "RULE_5_ORG_IMPERSONATION",
		Severity: types.RiskHIGH,
		Run: func(in Input) (bool, string) {
			// Heuristic: if caller_did contains a recognizable org slug
			// and the org_did claim doesn't match the registry org,
			// flag it. Phase 1 uses a tiny known-org map.
			knownOrgs := map[string]string{
				"kakaobank": "did:opendid:org:kakaobank",
			}
			for slug, expectedOrgDID := range knownOrgs {
				if strings.Contains(in.CallerDID, slug) && in.OrgDIDClaim != expectedOrgDID && in.OrgDIDClaim != "" {
					return true, "기관 사칭 의심: caller_did=\"" + in.CallerDID + "\" claims org=\"" + in.OrgDIDClaim + "\" but expected \"" + expectedOrgDID + "\""
				}
			}
			return false, ""
		},
	},
	{
		ID:       "RULE_6_URGENCY_THREAT",
		Severity: types.RiskHIGH,
		Run: func(in Input) (bool, string) {
			urgency := []string{"즉시", "긴급", "지금 당장", "바로"}
			threats := []string{"체포", "구속", "압류", "범죄 연루", "수사", "조사"}
			text := in.CallScript
			hasUrgency := false
			hasThreat := false
			for _, u := range urgency {
				if strings.Contains(text, u) {
					hasUrgency = true
					break
				}
			}
			for _, t := range threats {
				if strings.Contains(text, t) {
					hasThreat = true
					break
				}
			}
			if hasUrgency && hasThreat {
				return true, "긴급성 + 위협 패턴 동시 출현"
			}
			return false, ""
		},
	},
}

// Classify runs every rule and returns a Layer1Result with the highest
// severity attained. LOW = nothing fired.
func Classify(in Input) types.Layer1Result {
	now := time.Now().UTC()
	res := types.Layer1Result{
		Verdict:        types.RiskLOW,
		TriggeredRules: nil,
		Reasons:        nil,
		EvaluatedAt:    now,
	}
	for _, rule := range allRules {
		triggered, reason := rule.Run(in)
		if triggered {
			res.TriggeredRules = append(res.TriggeredRules, types.TriggeredRule{
				RuleID:   rule.ID,
				Reason:   reason,
				Severity: rule.Severity,
			})
			res.Reasons = append(res.Reasons, reason)
			if severityRank(rule.Severity) > severityRank(res.Verdict) {
				res.Verdict = rule.Severity
			}
		}
	}
	return res
}

func severityRank(v types.RiskVerdict) int {
	switch v {
	case types.RiskLOW:
		return 0
	case types.RiskMEDIUM:
		return 1
	case types.RiskHIGH:
		return 2
	case types.RiskBLOCK:
		return 3
	}
	return -1
}
