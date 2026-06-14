package types

import (
	"encoding/json"
	"testing"
	"time"
)

// TestCallPassport_ContractRoundtrip locks the JSON interface contract
// shared with the frontend PWA (Moafik directive §7). If this test breaks,
// the frontend will break too — coordinate any field change before editing.
func TestCallPassport_ContractRoundtrip(t *testing.T) {
	issued, _ := time.Parse(time.RFC3339, "2026-05-30T12:00:00Z")
	expires, _ := time.Parse(time.RFC3339, "2026-05-30T12:05:00Z")

	// Scenario 3 fixture from Moafik directive §7 — verbatim shape.
	p := CallPassport{
		SessionID: "sess_demo_scenario_3",
		IssuedAt:  issued,
		ExpiresAt: expires,
		Outcome:   OutcomeFAILED,
		Stamps: []Stamp{
			{Label: StampLabelOrgDID, Status: StampOK, Detail: "did:opendid:org:kakaobank"},
			{Label: StampLabelUnauthorizedPurpose, Status: StampFAIL, Detail: "보안 알림 권한만 보유 — 대출 영업 시도로 차단"},
			{Label: StampLabelBlockReceipt, Status: StampOK, Detail: "OmniOne Chain-compatible testnet"},
		},
		BlockReason:   "보안 알림 권한만 가진 은행 AI 상담사가 대출 영업을 시도했습니다. ZTCV는 권한·목적 외 통화를 사전 차단합니다.",
		ReceiptTxHash: "0xMOCK00000000000000000000000000000000000000000000000000000003",
		ExplorerURL:   "https://sepolia.etherscan.io/tx/0xMOCK...3",
		CallerDID:     "did:opendid:agent:kakaobank-ai-security-alert-007",
		CallerOrg:     "카카오뱅크",
		CallerPurpose: "loan_consultation",
	}

	// Field-name contract check: the JSON keys MUST match the TypeScript
	// type in the Moafik directive §7 exactly (camelCase, omitempty maps
	// to TypeScript optional `?`).
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	requiredKeys := []string{
		`"sessionId":`,
		`"issuedAt":`,
		`"expiresAt":`,
		`"outcome":`,
		`"stamps":`,
		`"blockReason":`,
		`"receiptTxHash":`,
		`"explorerUrl":`,
		`"callerDid":`,
		`"callerOrg":`,
		`"callerPurpose":`,
		`"label":`,
		`"status":`,
		`"detail":`,
	}
	js := string(raw)
	for _, k := range requiredKeys {
		if !contains(js, k) {
			t.Errorf("required JSON key missing from contract: %s\nfull JSON: %s", k, js)
		}
	}

	// Roundtrip back to struct — fields must survive.
	var back CallPassport
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if back.SessionID != p.SessionID || back.Outcome != p.Outcome || len(back.Stamps) != 3 {
		t.Fatalf("roundtrip lost data: got %+v", back)
	}
	if back.Stamps[1].Status != StampFAIL {
		t.Errorf("scenario-3 thesis stamp lost FAIL status: got %v", back.Stamps[1].Status)
	}
}

// TestCallPassport_ScenarioSafe_4Stamps verifies the canonical Scenario 1
// shape (SAFE outcome, 4 ✓ stamps in the locked order).
func TestCallPassport_ScenarioSafe_4Stamps(t *testing.T) {
	stamps := []Stamp{
		{Label: StampLabelOrgDID, Status: StampOK},
		{Label: StampLabelAgentAuth, Status: StampOK},
		{Label: StampLabelMobileID, Status: StampOK},
		{Label: StampLabelChainReceipt, Status: StampOK},
	}
	for i, s := range stamps {
		if s.Status != StampOK {
			t.Errorf("SAFE stamp %d should be OK: got %v", i, s.Status)
		}
	}
	if len(stamps) != 4 {
		t.Fatalf("SAFE outcome must have 4 stamps, got %d", len(stamps))
	}
}

// TestCallPassport_ScenarioBlock_BlockReceiptAlwaysOK verifies the BLOCK-
// receipt stamp invariant: every FAILED Call Passport carries a ✓ on the
// "차단 영수증 기록" stamp because the BLOCK decision itself is anchored to
// chain. The other failure stamps mark what actually failed.
func TestCallPassport_ScenarioBlock_BlockReceiptAlwaysOK(t *testing.T) {
	stampsScenario2 := []Stamp{
		{Label: StampLabelMissingDID, Status: StampFAIL},
		{Label: StampLabelTransferDemand, Status: StampFAIL},
		{Label: StampLabelBlockReceipt, Status: StampOK},
	}
	stampsScenario3 := []Stamp{
		{Label: StampLabelOrgDID, Status: StampOK},
		{Label: StampLabelUnauthorizedPurpose, Status: StampFAIL},
		{Label: StampLabelBlockReceipt, Status: StampOK},
	}
	for name, set := range map[string][]Stamp{"S2": stampsScenario2, "S3": stampsScenario3} {
		last := set[len(set)-1]
		if last.Label != StampLabelBlockReceipt || last.Status != StampOK {
			t.Errorf("%s: BlockReceipt stamp not ✓ as required (got label=%q status=%q)",
				name, last.Label, last.Status)
		}
	}
}

// contains is a tiny dependency-free substring check.
func contains(s, sub string) bool {
	return len(s) >= len(sub) && indexOf(s, sub) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
