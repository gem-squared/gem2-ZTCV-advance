// identity-svc owns OmniOne CX integration (real-slot + mock-fallback)
// for customer-side Mobile ID verification.
//
// In Phase 1 it boots as an empty skeleton with /healthz + /metrics.
// IdentityProvider interface, MockMobileIDProvider, and OACXProvider land
// in WP-02.
package main

import (
	"log"
	"net/http"

	"github.com/gem-squared/gem2-ZTCV/internal/config"
	"github.com/gem-squared/gem2-ZTCV/internal/server"
)

const svcName = "identity-svc"

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[%s] config error: %v", svcName, err)
	}
	log.Printf("[%s] OmniOne CX mode=%s (license=%v)", svcName, cfg.OmniOneCX.Mode, cfg.OmniOneCX.LicenseKey != "")
	b := server.New(svcName, cfg.Ports.IdentitySvc)

	b.Route("/", func(w http.ResponseWriter, r *http.Request) {
		server.WriteJSON(w, http.StatusOK, map[string]string{
			"svc":  svcName,
			"note": "identity-svc skeleton — OACX real-slot + mock fallback land in WP-02.U3/U4",
		})
	})

	if err := b.Run(); err != nil {
		log.Fatalf("[%s] fatal: %v", svcName, err)
	}
}
