package config

import (
	"strings"
	"testing"
)

// TestLoad_EmptyEnvBootsClean verifies the "sensible defaults so empty
// .env still boots" promise from the WP-01.U4 contract.
func TestLoad_EmptyEnvBootsClean(t *testing.T) {
	clearAllZTCVEnv(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("empty env should load cleanly in simulation mode, got: %v", err)
	}
	if cfg.OmniOneCX.Mode != OACXMock {
		t.Errorf("default OACX mode should be mock, got %q", cfg.OmniOneCX.Mode)
	}
	if cfg.LLM.WorkerProvider != LLMMock || cfg.LLM.AuditorProvider != LLMMock {
		t.Errorf("default LLM providers should be mock; got worker=%q auditor=%q",
			cfg.LLM.WorkerProvider, cfg.LLM.AuditorProvider)
	}
	if cfg.Chain.Provider != ChainLocal {
		t.Errorf("default chain provider should be local, got %q", cfg.Chain.Provider)
	}
	if cfg.Ports.SessionSvc != "8001" {
		t.Errorf("default session-svc port should be 8001, got %q", cfg.Ports.SessionSvc)
	}
}

// TestLoad_OACXRealRequiresLicense enforces the .env.example invariant:
// OMNIONE_CX_MODE=real without a license key is a fail-fast error.
func TestLoad_OACXRealRequiresLicense(t *testing.T) {
	clearAllZTCVEnv(t)
	t.Setenv("OMNIONE_CX_MODE", "real")
	// no OMNIONE_CX_LICENSE_KEY set
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when OACX mode=real without license key")
	}
	if !strings.Contains(err.Error(), "OMNIONE_CX_LICENSE_KEY") {
		t.Errorf("error should mention OMNIONE_CX_LICENSE_KEY, got: %v", err)
	}
}

// TestLoad_OACXRealWithLicenseOK confirms the upgrade path: license
// arriving is purely a config flip (no code change).
func TestLoad_OACXRealWithLicenseOK(t *testing.T) {
	clearAllZTCVEnv(t)
	t.Setenv("OMNIONE_CX_MODE", "real")
	t.Setenv("OMNIONE_CX_LICENSE_KEY", "fake-license-for-test")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("real mode with license should load: %v", err)
	}
	if cfg.OmniOneCX.Mode != OACXReal {
		t.Errorf("expected OACX mode=real, got %q", cfg.OmniOneCX.Mode)
	}
}

// TestLoad_InvalidLLMProvider verifies enum validation surfaces a clear
// error rather than silently using an unknown provider.
func TestLoad_InvalidLLMProvider(t *testing.T) {
	clearAllZTCVEnv(t)
	t.Setenv("LLM_PROVIDER_WORKER", "totally-fake-provider")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid LLM_PROVIDER_WORKER")
	}
	if !strings.Contains(err.Error(), "LLM_PROVIDER_WORKER") {
		t.Errorf("error should mention LLM_PROVIDER_WORKER, got: %v", err)
	}
}

// TestLoad_SepoliaRequiresDeployerKey verifies chain-provider enforcement
// (sepolia and omnione are key-required; local is not).
func TestLoad_SepoliaRequiresDeployerKey(t *testing.T) {
	clearAllZTCVEnv(t)
	t.Setenv("CHAIN_PROVIDER", "sepolia")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for sepolia without DEPLOYER_PK")
	}
	if !strings.Contains(err.Error(), "DEPLOYER_PK") {
		t.Errorf("error should mention DEPLOYER_PK, got: %v", err)
	}
}

// TestLoad_InvalidLogLevel ensures log-level enum is checked.
func TestLoad_InvalidLogLevel(t *testing.T) {
	clearAllZTCVEnv(t)
	t.Setenv("LOG_LEVEL", "verbose-please")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid LOG_LEVEL")
	}
	if !strings.Contains(err.Error(), "LOG_LEVEL") {
		t.Errorf("error should mention LOG_LEVEL, got: %v", err)
	}
}

// TestTest_ReturnsValidConfig confirms the Test() helper produces a
// validating Config for unit tests.
func TestTest_ReturnsValidConfig(t *testing.T) {
	cfg := Test()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Test() config failed validation: %v", err)
	}
	if cfg.OmniOneCX.Mode != OACXMock {
		t.Errorf("Test() should use OACX mock, got %q", cfg.OmniOneCX.Mode)
	}
	if cfg.DBPaths.SessionSvc != ":memory:" {
		t.Errorf("Test() should use in-memory SQLite, got %q", cfg.DBPaths.SessionSvc)
	}
}

// clearAllZTCVEnv unsets every ZTCV-relevant env var via t.Setenv so
// tests are insulated from the developer's shell state.
func clearAllZTCVEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"PORT_SESSION_SVC", "PORT_IDENTITY_SVC", "PORT_RISK_CHAIN_SVC", "PORT_CHAIN_ADAPTER", "PORT_FRONTEND",
		"DB_PATH_SESSION_SVC", "DB_PATH_IDENTITY_SVC", "DB_PATH_RISK_CHAIN_SVC",
		"LLM_PROVIDER_WORKER", "LLM_PROVIDER_AUDITOR",
		"LLM_API_VULTR", "LLM_API_OPENAI", "LLM_API_GEMINI", "LLM_API_ANTHROPIC",
		"OMNIONE_CX_MODE", "OMNIONE_CX_BASE_URL", "OMNIONE_CX_LICENSE_KEY",
		"CHAIN_PROVIDER", "CHAIN_RPC_LOCAL", "CHAIN_RPC_SEPOLIA", "CHAIN_RPC_OMNIONE",
		"DEPLOYER_PK", "ZTCV_RECEIPT_ANCHOR_ADDRESS",
		"CRAFTER_BASE_URL", "LOG_LEVEL",
		"DEV_ADMIN_TOKEN", "DEV_MOCK_SECRET",
	}
	for _, k := range keys {
		t.Setenv(k, "")
	}
}
