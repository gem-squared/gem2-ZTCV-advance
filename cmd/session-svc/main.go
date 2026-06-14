// session-svc owns the verification state machine, the in-process
// didregistry module (purpose-scope enforcement — Scenario 3 trigger),
// the caller-proof flow, and the trust-trace orchestration.
//
// All risk + chain logic runs in-process for Phase 1 simulation; the
// risk-chain-svc and chain-adapter binaries expose the same operations
// for direct frontend access if needed.
//
// OmniOne integration is mediated through the internal/omnione/ facade —
// real-slot ready, mock-fallback default. Activate via env vars when
// credentials arrive. See internal/omnione/README.md.
package main

import (
	"log"
	"net/http"

	"github.com/gem-squared/gem2-ZTCV/internal/config"
	"github.com/gem-squared/gem2-ZTCV/internal/didregistry"
	"github.com/gem-squared/gem2-ZTCV/internal/omnione"
	"github.com/gem-squared/gem2-ZTCV/internal/server"
	"github.com/gem-squared/gem2-ZTCV/internal/session"
)

const svcName = "session-svc"

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[%s] config error: %v", svcName, err)
	}

	// Storage
	repo, err := session.NewRepo(cfg.DBPaths.SessionSvc)
	if err != nil {
		log.Fatalf("[%s] sqlite open: %v", svcName, err)
	}
	defer func() { _ = repo.Close() }()

	// DID registry shares the same SQLite handle
	dids, err := didregistry.NewRepo(repo.DB())
	if err != nil {
		log.Fatalf("[%s] didregistry init: %v", svcName, err)
	}
	if err := didregistry.SeedDemoFixtures(dids); err != nil {
		log.Fatalf("[%s] didregistry seed: %v", svcName, err)
	}
	log.Printf("[%s] didregistry seeded — 3 demo scenarios ready (incl. security_alert agent for Scenario 3)", svcName)

	// ─── OmniOne adapter selection (facade — see internal/omnione/) ───

	// CX adapter — Receiver-side Mobile ID (OmniOne CX 표준인증창 + VC-Verifier)
	idp, cxLabel, err := omnione.NewCXAdapter(cfg.OmniOneCX, cfg.DevAdmin.MockSecret)
	if err != nil {
		log.Fatalf("[%s] OmniOne CX adapter: %v", svcName, err)
	}
	log.Printf("[%s] OmniOne CX adapter        = %s", svcName, cxLabel)

	// Open DID adapter — Caller-side institution DID + AgentAuthorization
	opendidAdapter, openDIDLabel, err := omnione.NewOpenDIDAdapter(cfg.OpenDID.Mode, dids)
	if err != nil {
		log.Fatalf("[%s] OmniOne Open DID adapter: %v", svcName, err)
	}
	log.Printf("[%s] OmniOne Open DID adapter   = %s", svcName, openDIDLabel)
	_ = opendidAdapter // currently logging-only; handlers continue to use dids directly until full refactor

	// Chain adapter — receipt anchor (local sim / Sepolia testnet / OmniOne Chain stub)
	anchor, chainLabel, err := omnione.NewChainAdapter(cfg.Chain)
	if err != nil {
		log.Fatalf("[%s] OmniOne Chain adapter: %v", svcName, err)
	}
	log.Printf("[%s] OmniOne Chain adapter      = %s", svcName, chainLabel)

	// HTTP
	b := server.New(svcName, cfg.Ports.SessionSvc)
	svc := session.NewService(repo, dids, idp, anchor)
	svc.RegisterRoutes(b)

	// Root placeholder
	b.Route("/", func(w http.ResponseWriter, r *http.Request) {
		server.WriteJSON(w, http.StatusOK, map[string]any{
			"svc": svcName,
			"endpoints": []string{
				"POST /api/session/create",
				"POST /api/session/{id}/caller-proof",
				"POST /api/session/{id}/customer-proof",
				"GET  /api/session/{id}",
				"GET  /api/session/{id}/passport",
				"GET  /api/session/{id}/events  (SSE)",
				"GET  /api/did/{did}",
				"POST /api/scenarios/run?n=1|2|3",
			},
		})
	})

	if err := b.Run(); err != nil {
		log.Fatalf("[%s] fatal: %v", svcName, err)
	}
}
