package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gem-squared/gem2-ZTCV/internal/types"
)

// MockMobileIDProvider returns deterministic mock JWTs + Claims for the
// three demo customers. The "JWT" is a plaintext sentinel of the form
// "mock-jwt:<customer-id>:<sessionID>" — no signing is performed (we
// keep secret-key handling out of Phase 1). VerifyToken trusts any
// sentinel matching a seeded customer.
type MockMobileIDProvider struct {
	secret string
}

// NewMockProvider returns a mock provider. The secret is unused in
// Phase 1 (would be the HS256 signing key if we cared about signature
// integrity in mock mode).
func NewMockProvider(secret string) *MockMobileIDProvider {
	return &MockMobileIDProvider{secret: secret}
}

// Mode reports the provider mode.
func (m *MockMobileIDProvider) Mode() string { return "mock" }

// StartVerification returns a fake QR/app-link the frontend can render
// as a placeholder. The session ID is preserved in both fields so the
// callback can correlate.
func (m *MockMobileIDProvider) StartVerification(sessionID string) (*types.VerificationRequest, error) {
	expires := time.Now().UTC().Add(5 * time.Minute).Format(time.RFC3339)
	return &types.VerificationRequest{
		SessionID: sessionID,
		QRPayload: fmt.Sprintf("mock-oacx-qr:%s", sessionID),
		AppLink:   fmt.Sprintf("mock-oacx-app-link://verify?session=%s", sessionID),
		ExpiresAt: expires,
	}, nil
}

// MockToken builds the sentinel token for a given seeded customer +
// session. Use this in scripts/fixtures to feed VerifyToken.
func MockToken(customerID, sessionID string) string {
	return fmt.Sprintf("mock-jwt:%s:%s", customerID, sessionID)
}

// VerifyToken parses the sentinel and returns deterministic Claims.
// Any seeded customer id (customer-001/002/003) is accepted; anything
// else returns an error.
func (m *MockMobileIDProvider) VerifyToken(token string) (*types.MobileIDClaims, error) {
	parts := strings.SplitN(token, ":", 3)
	if len(parts) != 3 || parts[0] != "mock-jwt" {
		return nil, fmt.Errorf("mock provider: invalid token shape (want mock-jwt:<customer>:<session>)")
	}
	customerID := parts[1]
	switch customerID {
	case "customer-001", "customer-002", "customer-003":
		// seeded — fall through
	default:
		return nil, fmt.Errorf("mock provider: unseeded customer %q", customerID)
	}
	return &types.MobileIDClaims{
		VCType:     types.VCTypeMobileResidentCard,
		NameHash:   sha256Hex("mock-name-" + customerID),
		BirthDate:  "19900101",
		IssuingOrg: "행정안전부 (mock)",
		DocNoHash:  sha256Hex("mock-residentno-" + customerID),
		VerifiedAt: time.Now().UTC(),
	}, nil
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
