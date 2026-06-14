// risk-chain-svc owns Layer 1 (rule classifier, incl. Rule 4 DID-purpose
// mismatch for Scenario 3), Layer 2 (gem2-epistemic-engine wrapper +
// mock), the conservative composer, receipt generation, and the call to
// chain-adapter for anchoring.
//
// In Phase 1 it boots as an empty skeleton with /healthz + /metrics.
// Layers + composer land in WP-03.U1/U2/U3; receipt + anchor in WP-03.U6.
package main

import (
	"log"
	"net/http"

	"github.com/gem-squared/gem2-ZTCV/internal/config"
	"github.com/gem-squared/gem2-ZTCV/internal/server"
)

const svcName = "risk-chain-svc"

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[%s] config error: %v", svcName, err)
	}
	log.Printf("[%s] LLM worker=%s auditor=%s", svcName, cfg.LLM.WorkerProvider, cfg.LLM.AuditorProvider)
	b := server.New(svcName, cfg.Ports.RiskChainSvc)

	b.Route("/", func(w http.ResponseWriter, r *http.Request) {
		server.WriteJSON(w, http.StatusOK, map[string]string{
			"svc":  svcName,
			"note": "risk-chain-svc skeleton — Layer1 rules + Layer2 LLM + composer + receipt anchor land in WP-03",
		})
	})

	if err := b.Run(); err != nil {
		log.Fatalf("[%s] fatal: %v", svcName, err)
	}
}
