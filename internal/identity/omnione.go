package identity

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gem-squared/gem2-ZTCV/internal/types"
)

// OACXProvider is the real-slot OmniOne CX client. Inert in Phase 1
// simulation (we run with OMNIONE_CX_MODE=mock). The code path
// compiles and boots; license + base URL are gated at config-load.
type OACXProvider struct {
	baseURL    string
	licenseKey string
	client     *http.Client
}

// NewOACXProvider returns a real-slot provider. baseURL typically
// "https://cx.raonsecure.co.kr:18543".
func NewOACXProvider(baseURL, licenseKey string) *OACXProvider {
	return &OACXProvider{
		baseURL:    baseURL,
		licenseKey: licenseKey,
		client:     &http.Client{Timeout: 5 * time.Second},
	}
}

// Mode reports the provider mode.
func (p *OACXProvider) Mode() string { return "real" }

// StartVerification — placeholder. Real implementation calls
// POST /oacx/api/v1.0/trans then POST /oacx/api/v1.0/authen/qr/request
// or /authen/app/request based on User-Agent. Phase 1 returns ErrNotImplemented
// since the provider is never selected when MODE=mock.
func (p *OACXProvider) StartVerification(sessionID string) (*types.VerificationRequest, error) {
	return nil, errors.New("OACXProvider: StartVerification not implemented in Phase 1 (real-slot stub)")
}

// VerifyToken calls GET /oacx/api/v1.0/trans/{token} and parses the
// {data: {...}} payload into MobileIDClaims. The full mapping (VC type
// branching on locpanm + licNo vs name, etc.) lands when sandbox is
// licensed; for now we return a structural error so the handler can
// surface the gap.
func (p *OACXProvider) VerifyToken(oacxToken string) (*types.MobileIDClaims, error) {
	url := fmt.Sprintf("%s/oacx/api/v1.0/trans/%s", p.baseURL, oacxToken)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("oacx request build: %w", err)
	}
	if p.licenseKey != "" {
		req.Header.Set("X-OACX-License-Key", p.licenseKey)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("oacx request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("oacx body read: %w", err)
	}
	var envelope struct {
		ResultCode string `json:"resultCode"`
		OACXCode   string `json:"oacxCode"`
		Data       struct {
			VCType    string `json:"vcTypeCode"`
			Name      string `json:"name"`
			BirthDate string `json:"birthDate"`
			LicNo     string `json:"licNo"`
			Locpanm   string `json:"locpanm"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("oacx decode: %w (body=%s)", err, string(body))
	}
	if envelope.ResultCode != "200" {
		return nil, fmt.Errorf("oacx non-200 resultCode=%s oacxCode=%s", envelope.ResultCode, envelope.OACXCode)
	}
	return &types.MobileIDClaims{
		VCType:     types.VCType(envelope.Data.VCType),
		NameHash:   sha256Hex(envelope.Data.Name),
		BirthDate:  envelope.Data.BirthDate,
		IssuingOrg: envelope.Data.Locpanm,
		DocNoHash:  sha256Hex(envelope.Data.LicNo),
		VerifiedAt: time.Now().UTC(),
	}, nil
}
