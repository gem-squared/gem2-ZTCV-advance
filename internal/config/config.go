// Package config is the single source of truth for environment-driven
// configuration across all ZTCV binaries. Direct os.Getenv calls outside
// this package are a code smell — go through Load() instead.
//
// Defaults are chosen so an empty .env file still boots in simulation
// mode (mock LLM, mock OACX, local Hardhat chain). Anything that
// requires credentials (e.g., OMNIONE_CX_MODE=real) fails fast at Load
// with a clear error message.
package config

import (
	"fmt"
	"os"
	"strings"
)

// Config is the resolved, validated configuration. Created by Load.
type Config struct {
	Ports     Ports
	DBPaths   DBPaths
	LLM       LLM
	OmniOneCX OmniOneCX
	OpenDID   OpenDID
	Chain     Chain
	Crafter   Crafter
	LogLevel  string
	DevAdmin  DevAdmin
}

// Ports holds the 5 binary ports (4 backend + frontend).
type Ports struct {
	SessionSvc   string
	IdentitySvc  string
	RiskChainSvc string
	ChainAdapter string
	Frontend     string
}

// DBPaths holds per-binary SQLite file paths.
type DBPaths struct {
	SessionSvc   string
	IdentitySvc  string
	RiskChainSvc string
}

// LLMProvider enumerates the supported worker / auditor backends.
type LLMProvider string

const (
	LLMVultr     LLMProvider = "vultr"
	LLMOpenAI    LLMProvider = "openai"
	LLMGemini    LLMProvider = "gemini"
	LLMAnthropic LLMProvider = "anthropic"
	LLMMock      LLMProvider = "mock"
)

// LLM holds LLM provider selection + API keys.
type LLM struct {
	WorkerProvider  LLMProvider
	AuditorProvider LLMProvider
	VultrAPIKey     string
	OpenAIAPIKey    string
	GeminiAPIKey    string
	AnthropicAPIKey string
}

// OACXMode is the OmniOne CX integration mode.
type OACXMode string

const (
	OACXReal OACXMode = "real"
	OACXMock OACXMode = "mock"
)

// OmniOneCX holds Mobile ID verification configuration.
type OmniOneCX struct {
	Mode       OACXMode
	BaseURL    string
	LicenseKey string
}

// OpenDIDMode is the OmniOne Open DID adapter mode.
type OpenDIDMode string

const (
	OpenDIDReal OpenDIDMode = "real"
	OpenDIDMock OpenDIDMode = "mock"
)

// OpenDID holds OmniOne Open DID (Caller-side DID/VC) adapter configuration.
// Default mock = Go-native didregistry. Real mode (Java SDK sidecar)
// is deferred to the 결선 development phase.
type OpenDID struct {
	Mode OpenDIDMode
}

// ChainProvider enumerates supported chain backends.
type ChainProvider string

const (
	ChainLocal   ChainProvider = "local"
	ChainSepolia ChainProvider = "sepolia"
	ChainOmnione ChainProvider = "omnione"
)

// Chain holds chain provider configuration.
type Chain struct {
	Provider              ChainProvider
	RPCLocal              string
	RPCSepolia            string
	RPCOmnione            string
	DeployerPrivateKey    string
	ZTCVReceiptAnchorAddr string
}

// Crafter holds GEM²-Crafter reference (pre-gen artifact source).
type Crafter struct {
	BaseURL string
}

// DevAdmin holds development-only secrets (admin token + mock signing key).
type DevAdmin struct {
	AdminToken string
	MockSecret string
}

// Load reads environment variables into a Config and validates them.
// Returns an error explaining the first invalid field encountered.
// Defaults match .env.example.
func Load() (*Config, error) {
	cfg := &Config{
		Ports: Ports{
			SessionSvc:   getenv("PORT_SESSION_SVC", "8001"),
			IdentitySvc:  getenv("PORT_IDENTITY_SVC", "8002"),
			RiskChainSvc: getenv("PORT_RISK_CHAIN_SVC", "8003"),
			ChainAdapter: getenv("PORT_CHAIN_ADAPTER", "8004"),
			Frontend:     getenv("PORT_FRONTEND", "3000"),
		},
		DBPaths: DBPaths{
			SessionSvc:   getenv("DB_PATH_SESSION_SVC", "./data/session.db"),
			IdentitySvc:  getenv("DB_PATH_IDENTITY_SVC", "./data/identity.db"),
			RiskChainSvc: getenv("DB_PATH_RISK_CHAIN_SVC", "./data/risk-chain.db"),
		},
		LLM: LLM{
			WorkerProvider:  LLMProvider(getenv("LLM_PROVIDER_WORKER", "mock")),
			AuditorProvider: LLMProvider(getenv("LLM_PROVIDER_AUDITOR", "mock")),
			VultrAPIKey:     os.Getenv("LLM_API_VULTR"),
			OpenAIAPIKey:    os.Getenv("LLM_API_OPENAI"),
			GeminiAPIKey:    os.Getenv("LLM_API_GEMINI"),
			AnthropicAPIKey: os.Getenv("LLM_API_ANTHROPIC"),
		},
		OmniOneCX: OmniOneCX{
			Mode:       OACXMode(getenv("OMNIONE_CX_MODE", "mock")),
			BaseURL:    getenv("OMNIONE_CX_BASE_URL", "https://cx.raonsecure.co.kr:18543"),
			LicenseKey: os.Getenv("OMNIONE_CX_LICENSE_KEY"),
		},
		OpenDID: OpenDID{
			Mode: OpenDIDMode(getenv("OMNIONE_OPENDID_MODE", "mock")),
		},
		Chain: Chain{
			Provider:              ChainProvider(getenv("CHAIN_PROVIDER", "local")),
			RPCLocal:              getenv("CHAIN_RPC_LOCAL", "http://chain-sim:8545"),
			RPCSepolia:            os.Getenv("CHAIN_RPC_SEPOLIA"),
			RPCOmnione:            os.Getenv("CHAIN_RPC_OMNIONE"),
			DeployerPrivateKey:    os.Getenv("DEPLOYER_PK"),
			ZTCVReceiptAnchorAddr: os.Getenv("ZTCV_RECEIPT_ANCHOR_ADDRESS"),
		},
		Crafter: Crafter{
			BaseURL: getenv("CRAFTER_BASE_URL", "https://ai-olympic.gemsquared.ai"),
		},
		LogLevel: getenv("LOG_LEVEL", "info"),
		DevAdmin: DevAdmin{
			AdminToken: getenv("DEV_ADMIN_TOKEN", "change-me-in-local"),
			MockSecret: getenv("DEV_MOCK_SECRET", "mock-jwt-hs256-secret"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate runs all field-level checks. Fail-fast: returns the first
// problem encountered.
func (c *Config) Validate() error {
	// LLM provider enum check.
	if !validLLMProvider(c.LLM.WorkerProvider) {
		return fmt.Errorf("invalid LLM_PROVIDER_WORKER %q: must be one of vultr|openai|mock", c.LLM.WorkerProvider)
	}
	if !validLLMProvider(c.LLM.AuditorProvider) {
		return fmt.Errorf("invalid LLM_PROVIDER_AUDITOR %q: must be one of gemini|anthropic|mock", c.LLM.AuditorProvider)
	}

	// OACX mode + license check.
	if c.OmniOneCX.Mode != OACXReal && c.OmniOneCX.Mode != OACXMock {
		return fmt.Errorf("invalid OMNIONE_CX_MODE %q: must be real|mock", c.OmniOneCX.Mode)
	}
	if c.OmniOneCX.Mode == OACXReal && c.OmniOneCX.LicenseKey == "" {
		return fmt.Errorf("OMNIONE_CX_MODE=real requires OMNIONE_CX_LICENSE_KEY to be set")
	}

	// Open DID mode check.
	if c.OpenDID.Mode != OpenDIDReal && c.OpenDID.Mode != OpenDIDMock {
		return fmt.Errorf("invalid OMNIONE_OPENDID_MODE %q: must be real|mock", c.OpenDID.Mode)
	}

	// Chain provider enum check.
	switch c.Chain.Provider {
	case ChainLocal, ChainSepolia, ChainOmnione:
	default:
		return fmt.Errorf("invalid CHAIN_PROVIDER %q: must be local|sepolia|omnione", c.Chain.Provider)
	}
	// If provider needs a private key (sepolia / omnione), enforce it.
	if (c.Chain.Provider == ChainSepolia || c.Chain.Provider == ChainOmnione) && c.Chain.DeployerPrivateKey == "" {
		return fmt.Errorf("CHAIN_PROVIDER=%s requires DEPLOYER_PK to be set", c.Chain.Provider)
	}

	// Log level enum check.
	switch strings.ToLower(c.LogLevel) {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("invalid LOG_LEVEL %q: must be debug|info|warn|error", c.LogLevel)
	}

	return nil
}

// Test returns a Config populated with simulation-mode defaults, suitable
// for unit tests. No env vars consulted.
func Test() *Config {
	return &Config{
		Ports: Ports{
			SessionSvc:   "18001",
			IdentitySvc:  "18002",
			RiskChainSvc: "18003",
			ChainAdapter: "18004",
			Frontend:     "13000",
		},
		DBPaths: DBPaths{
			SessionSvc:   ":memory:",
			IdentitySvc:  ":memory:",
			RiskChainSvc: ":memory:",
		},
		LLM: LLM{
			WorkerProvider:  LLMMock,
			AuditorProvider: LLMMock,
		},
		OmniOneCX: OmniOneCX{
			Mode:    OACXMock,
			BaseURL: "https://test.invalid",
		},
		OpenDID: OpenDID{
			Mode: OpenDIDMock,
		},
		Chain: Chain{
			Provider: ChainLocal,
			RPCLocal: "http://localhost:8545",
		},
		Crafter:  Crafter{BaseURL: "https://test.invalid"},
		LogLevel: "debug",
		DevAdmin: DevAdmin{AdminToken: "test-admin", MockSecret: "test-mock"},
	}
}

// getenv returns the env value or a default if unset/empty.
func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func validLLMProvider(p LLMProvider) bool {
	switch p {
	case LLMVultr, LLMOpenAI, LLMGemini, LLMAnthropic, LLMMock:
		return true
	}
	return false
}
