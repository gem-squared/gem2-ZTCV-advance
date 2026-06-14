// Chain adapter — receipt anchor for Verification Broker.
//
// Default: local in-memory simulator (no external dependency).
// Sepolia: real EVM testnet via go-ethereum + Alchemy RPC (currently used in
//          production at https://ztcv-demo.gemsquared.ai).
// OmniOne: BESU-based OmniOne Chain via dev portal (deferred until portal
//          access; stub returns "not implemented" today).
package omnione

import (
	"fmt"

	"github.com/gem-squared/gem2-ZTCV/internal/chain"
	"github.com/gem-squared/gem2-ZTCV/internal/config"
)

// ChainAdapter aliases the existing ChainAnchor iface so callers depend on
// the omnione/ facade rather than internal/chain/ directly.
type ChainAdapter = chain.ChainAnchor

// NewChainAdapter selects the chain anchor implementation per env config.
//
// PROVIDER switches:
//
//	CHAIN_PROVIDER=local    → in-memory LocalSim (deterministic mock hashes)
//	CHAIN_PROVIDER=sepolia  → real Sepolia testnet via go-ethereum
//	                          (requires DEPLOYER_PK + CHAIN_RPC_SEPOLIA +
//	                          ZTCV_RECEIPT_ANCHOR_ADDRESS)
//	CHAIN_PROVIDER=omnione  → OmniOne Chain (BESU) via dev portal REST API
//	                          (DEFERRED — dev portal access pending)
func NewChainAdapter(cfg config.Chain) (ChainAdapter, string, error) {
	switch cfg.Provider {
	case config.ChainLocal:
		return chain.NewLocalSim(),
			"local sim (no real chain dependency)",
			nil

	case config.ChainSepolia:
		evm, err := chain.NewEVMTestnetProvider(
			cfg.RPCSepolia,
			cfg.DeployerPrivateKey,
			cfg.ZTCVReceiptAnchorAddr,
			"https://sepolia.etherscan.io/tx/",
		)
		if err != nil {
			return nil, "", fmt.Errorf("sepolia EVM provider init: %w", err)
		}
		label := fmt.Sprintf("sepolia EVMTestnetProvider (contract %s)", cfg.ZTCVReceiptAnchorAddr)
		return evm, label, nil

	case config.ChainOmnione:
		// Deferred until OmniOne Chain dev portal access.
		// Stub satisfies the interface and surfaces the missing wiring at
		// boot rather than silently substituting local sim.
		return chain.NewEVMStub(cfg.RPCOmnione, cfg.ZTCVReceiptAnchorAddr),
			"omnione stub (dev portal access pending — defer until credentials arrive)",
			nil

	default:
		return nil, "", fmt.Errorf(
			"invalid CHAIN_PROVIDER %q (expected: local | sepolia | omnione)", cfg.Provider,
		)
	}
}
