package didregistry

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/gem-squared/gem2-ZTCV/internal/types"
)

// SeedDemoFixtures inserts the canonical 3-scenario DID/auth/phone
// data. Idempotent — UpsertDID etc. handle re-runs.
//
// Scenarios these seeds power:
//   - Scenario 1 (SAFE):   did:opendid:agent:kakaobank-ai-loan-counselor-001
//     authorized for purposes=[loan_consultation]
//   - Scenario 2 (BLOCK):  no seed needed — caller DID is unknown,
//     so Resolve returns ErrNotFound = unknown_did
//   - Scenario 3 (BLOCK):  did:opendid:agent:kakaobank-ai-security-alert-007
//     authorized for purposes=[security_alert] ONLY
//     → triggers unauthorized_purpose when called with
//     loan_consultation or insurance_sales (the thesis)
func SeedDemoFixtures(r *Repo) error {
	now := time.Now().UTC()
	farFuture := now.AddDate(2, 0, 0) // 2 years from now

	orgs := []types.DIDRecord{
		{
			DID:         "did:opendid:org:kakaobank",
			SubjectType: "org",
			DisplayName: "카카오뱅크",
			Document: &types.DIDDocument{
				Context:    []string{"https://www.w3.org/ns/did/v1"},
				ID:         "did:opendid:org:kakaobank",
				Controller: []string{"did:opendid:org:kakaobank"},
				VerificationMethod: []types.VerificationMethod{{
					ID:              "did:opendid:org:kakaobank#keys-1",
					Type:            "Ed25519VerificationKey2020",
					Controller:      "did:opendid:org:kakaobank",
					PublicKeyBase58: "kakao-org-public-key-mock-base58-encoded",
				}},
				Authentication: []string{"did:opendid:org:kakaobank#keys-1"},
				Status:         "active",
				CreatedAt:      now,
			},
		},
	}

	agents := []types.DIDRecord{
		{
			DID:         "did:opendid:agent:kakaobank-ai-loan-counselor-001",
			SubjectType: "agent",
			DisplayName: "KakaoBank AI 대출 상담사",
			Document: &types.DIDDocument{
				Context:        []string{"https://www.w3.org/ns/did/v1"},
				ID:             "did:opendid:agent:kakaobank-ai-loan-counselor-001",
				Controller:     []string{"did:opendid:org:kakaobank"},
				Authentication: []string{"did:opendid:org:kakaobank#keys-1"},
				Status:         "active",
				CreatedAt:      now,
			},
		},
		{
			DID:         "did:opendid:agent:kakaobank-ai-security-alert-007",
			SubjectType: "agent",
			DisplayName: "KakaoBank AI 보안 알림 상담사",
			Document: &types.DIDDocument{
				Context:        []string{"https://www.w3.org/ns/did/v1"},
				ID:             "did:opendid:agent:kakaobank-ai-security-alert-007",
				Controller:     []string{"did:opendid:org:kakaobank"},
				Authentication: []string{"did:opendid:org:kakaobank#keys-1"},
				Status:         "active",
				CreatedAt:      now,
			},
		},
	}

	users := []types.DIDRecord{
		newUserDID("did:opendid:user:customer-001", "Customer 001", now),
		newUserDID("did:opendid:user:customer-002", "Customer 002", now),
		newUserDID("did:opendid:user:customer-003", "Customer 003", now),
	}

	// Insert DIDs
	for _, batch := range [][]types.DIDRecord{orgs, agents, users} {
		for _, rec := range batch {
			if err := r.UpsertDID(rec); err != nil {
				return err
			}
		}
	}

	// Authorizations — this is the Scenario 3 trigger fixture
	auths := []types.AgentAuthorization{
		{
			AgentDID:        "did:opendid:agent:kakaobank-ai-loan-counselor-001",
			OrgDID:          "did:opendid:org:kakaobank",
			AllowedPurposes: []string{"loan_consultation"},
			ValidFrom:       now.AddDate(0, -1, 0),
			ValidUntil:      farFuture,
			Status:          types.AuthActive,
		},
		{
			// THE SCENARIO 3 SEED — authorized for security_alert ONLY.
			// Calling with purpose=loan_consultation or insurance_sales
			// triggers unauthorized_purpose at the caller-proof step.
			AgentDID:        "did:opendid:agent:kakaobank-ai-security-alert-007",
			OrgDID:          "did:opendid:org:kakaobank",
			AllowedPurposes: []string{"security_alert"},
			ValidFrom:       now.AddDate(0, -1, 0),
			ValidUntil:      farFuture,
			Status:          types.AuthActive,
		},
	}
	for _, a := range auths {
		if err := r.UpsertAuthorization(a); err != nil {
			return err
		}
	}

	// Phone bindings
	phones := []types.PhoneBinding{
		newPhoneBinding("+82-10-1234-5678", "did:opendid:user:customer-001", now),
		newPhoneBinding("+82-10-2345-6789", "did:opendid:user:customer-002", now),
		newPhoneBinding("+82-10-3456-7890", "did:opendid:user:customer-003", now),
	}
	for _, pb := range phones {
		if err := r.UpsertPhoneBinding(pb); err != nil {
			return err
		}
	}

	return nil
}

func newUserDID(id, displayName string, now time.Time) types.DIDRecord {
	return types.DIDRecord{
		DID:         id,
		SubjectType: "user",
		DisplayName: displayName,
		Document: &types.DIDDocument{
			Context:   []string{"https://www.w3.org/ns/did/v1"},
			ID:        id,
			Status:    "active",
			CreatedAt: now,
		},
	}
}

func newPhoneBinding(phoneE164, did string, now time.Time) types.PhoneBinding {
	h := sha256.Sum256([]byte(phoneE164))
	return types.PhoneBinding{
		PhoneHash:     hex.EncodeToString(h[:]),
		DID:           did,
		BindingStatus: "active",
		CreatedAt:     now,
	}
}
