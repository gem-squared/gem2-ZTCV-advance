// Package passport assembles the CallPassport artifact from a session
// + trust trace + receipt. The output JSON matches the Moafik directive
// §7 TypeScript shape exactly (frontend ↔ backend contract).
package passport

import (
	"time"

	"github.com/gem-squared/gem2-ZTCV/internal/types"
)

// Input bundles everything Build needs.
type Input struct {
	Session       *types.CallSession
	Trace         *types.TrustTrace
	ReceiptTxHash string
	ExplorerURL   string
	CallerOrgName string
	BlockReason   string // optional; if empty + session.State=blocked we derive

	// Intent Handshake (WP-ZTCV-06). Optional — when nil the Build
	// emits a passport with the IntentHandshake / IntentManifestHash
	// fields omitted, byte-identical to the pre-WP-06 response shape.
	IntentManifest     *types.IntentManifest
	IntentManifestHash string
}

// Build produces a CallPassport. The 4-stamp model:
//
//	SAFE outcome → 4 OK stamps in fixed order:
//	  ✓ 기관 DID
//	  ✓ AI Agent 권한
//	  ✓ 수신자 Mobile ID
//	  ✓ Chain Receipt
//
//	FAILED outcome → mix of ✗ + always-OK "차단 영수증 기록" at the end.
//	  Stamps fire per failure reason from session.BlockReason:
//	    unknown_did         → ✗ 발신자 DID 없음
//	    unauthorized_purpose → ✓ 기관 DID + ✗ AI Agent 권한 (권한 외 목적)
//	    transfer-demand risk → ✗ 송금 요구 감지
//	  Always ends with ✓ 차단 영수증 기록 (the BLOCK decision is anchored).
func Build(in Input) *types.CallPassport {
	if in.Session == nil {
		return nil
	}
	now := time.Now().UTC()
	cp := &types.CallPassport{
		SessionID:          in.Session.ID,
		IssuedAt:           now,
		ExpiresAt:          in.Session.ExpiresAt,
		ReceiptTxHash:      in.ReceiptTxHash,
		ExplorerURL:        in.ExplorerURL,
		CallerOrg:          in.CallerOrgName,
		IntentHandshake:    in.IntentManifest,
		IntentManifestHash: in.IntentManifestHash,
	}
	if cp.IssuedAt.Before(in.Session.CreatedAt) {
		cp.IssuedAt = in.Session.CreatedAt
	}
	if in.Session.CallerProof != nil {
		cp.CallerDID = in.Session.CallerProof.CallerDID
		cp.CallerPurpose = in.Session.CallerProof.Purpose
	}

	switch in.Session.State {
	case types.StateVerified:
		cp.Outcome = types.OutcomeSAFE
		cp.Stamps = []types.Stamp{
			{Label: types.StampLabelOrgDID, Status: types.StampOK, Detail: cp.CallerDID},
			{Label: types.StampLabelAgentAuth, Status: types.StampOK, Detail: "purpose=" + cp.CallerPurpose + " 권한 확인"},
			{Label: types.StampLabelMobileID, Status: types.StampOK, Detail: "OmniOne CX (mock) — 수신자 Mobile ID 검증 완료"},
			{Label: types.StampLabelChainReceipt, Status: types.StampOK, Detail: in.ReceiptTxHash},
		}

	case types.StateBlocked:
		cp.Outcome = types.OutcomeFAILED
		cp.BlockReason = pickReason(in.BlockReason, in.Session.BlockReason)
		cp.Stamps = buildFailedStamps(cp.BlockReason, in.Session, in.ReceiptTxHash)

	default:
		// Session still in progress — return a partial passport (frontend
		// can render "verifying…").
		cp.Outcome = types.OutcomeFAILED
		cp.BlockReason = "verification still in progress (state=" + string(in.Session.State) + ")"
		cp.Stamps = []types.Stamp{{Label: types.StampLabelChainReceipt, Status: types.StampFAIL, Detail: "anchor not yet recorded"}}
	}
	return cp
}

func pickReason(explicit, sessionReason string) string {
	if explicit != "" {
		return explicit
	}
	if sessionReason != "" {
		return sessionReason
	}
	return "통화가 차단되었습니다"
}

// buildFailedStamps returns the FAILED-form stamp list for the
// observed BlockReason. The "차단 영수증 기록" stamp is always last and
// always OK — the BLOCK decision itself was anchored.
func buildFailedStamps(blockReason string, s *types.CallSession, txHash string) []types.Stamp {
	stamps := []types.Stamp{}

	switch {
	case contains(blockReason, "unknown_did") || contains(blockReason, "발신자 DID 없음"):
		// Scenario 2 shape: ✗ 발신자 DID 없음 + ✗ 송금 요구 감지 + ✓ 차단 영수증 기록
		stamps = append(stamps,
			types.Stamp{Label: types.StampLabelMissingDID, Status: types.StampFAIL, Detail: "발신자 DID 미등록 — 신원 확인 불가"},
		)
		if s.RiskVerdict != nil && hasTransferDemand(s.RiskVerdict.Layer1) {
			stamps = append(stamps,
				types.Stamp{Label: types.StampLabelTransferDemand, Status: types.StampFAIL, Detail: "송금/안전계좌 패턴 감지"},
			)
		}

	case contains(blockReason, "unauthorized_purpose") || contains(blockReason, "권한 외 목적"):
		// Scenario 3 shape: ✓ 기관 DID + ✗ AI Agent 권한 (권한 외 목적) + ✓ 차단 영수증 기록
		detail := "기관 DID 확인"
		if s.CallerProof != nil {
			detail = s.CallerProof.OrgDID
		}
		stamps = append(stamps,
			types.Stamp{Label: types.StampLabelOrgDID, Status: types.StampOK, Detail: detail},
			types.Stamp{Label: types.StampLabelUnauthorizedPurpose, Status: types.StampFAIL,
				Detail: extractAuthScopeDetail(s)},
		)

	default:
		// Generic failure shape (e.g., Layer 2 BLOCK without specific
		// upstream reason). Show what we did succeed at + a generic risk fail.
		stamps = append(stamps,
			types.Stamp{Label: types.StampLabelOrgDID, Status: types.StampOK, Detail: ""},
			types.Stamp{Label: types.StampLabelTransferDemand, Status: types.StampFAIL, Detail: blockReason},
		)
	}

	// Always end with the always-OK BlockReceipt stamp.
	stamps = append(stamps, types.Stamp{Label: types.StampLabelBlockReceipt, Status: types.StampOK, Detail: txHash})
	return stamps
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func hasTransferDemand(l types.Layer1Result) bool {
	for _, t := range l.TriggeredRules {
		if t.RuleID == "RULE_1_TRANSFER_DEMAND" || t.RuleID == "RULE_2_SAFE_ACCOUNT" {
			return true
		}
	}
	return false
}

func extractAuthScopeDetail(s *types.CallSession) string {
	if s.CallerProof == nil {
		return "권한 외 목적"
	}
	return "보유 권한과 다른 목적 (요청 purpose=" + s.CallerProof.Purpose + ")"
}
