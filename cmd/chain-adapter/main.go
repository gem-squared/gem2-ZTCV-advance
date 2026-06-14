// chain-adapter owns the EVM client (go-ethereum), the embedded
// ZTCVReceiptAnchor ABI, and the local Hardhat / Sepolia / OmniOne RPC
// switching. Risk-chain-svc calls into it via REST when a session needs
// to anchor a Call Passport.
//
// In Phase 1 it boots as an empty skeleton with /healthz + /metrics.
// ChainAnchor interface + LocalChainSimulator + EVMTestnetProvider land
// in WP-03.U5; receipt anchor flow in WP-03.U6.
package main

import (
	"log"
	"net/http"

	"github.com/gem-squared/gem2-ZTCV/internal/config"
	"github.com/gem-squared/gem2-ZTCV/internal/server"
)

const svcName = "chain-adapter"

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[%s] config error: %v", svcName, err)
	}
	log.Printf("[%s] chain provider=%s", svcName, cfg.Chain.Provider)
	b := server.New(svcName, cfg.Ports.ChainAdapter)

	b.Route("/", func(w http.ResponseWriter, r *http.Request) {
		server.WriteJSON(w, http.StatusOK, map[string]string{
			"svc":  svcName,
			"note": "chain-adapter skeleton — ChainAnchor interface + Hardhat/Sepolia clients land in WP-03.U5",
		})
	})

	if err := b.Run(); err != nil {
		log.Fatalf("[%s] fatal: %v", svcName, err)
	}
}
