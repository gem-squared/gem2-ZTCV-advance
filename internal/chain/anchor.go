// Package chain provides the ChainAnchor interface + a local in-memory
// simulator (Phase 1) and a real-slot EVM client stub (Phase 2+).
package chain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"github.com/gem-squared/gem2-ZTCV/internal/types"
)

// ChainAnchor is the abstraction over "send a receipt to a chain".
type ChainAnchor interface {
	AnchorReceipt(sessionHash, receiptHash string, isSafe bool, policyVersion string) (txHash, explorerURL string, err error)
	GetReceipt(sessionHash string) (*types.ReceiptOnChain, error)
}

// LocalSim is the in-memory chain simulator used by Phase 1 demos.
// No actual Hardhat node required — produces deterministic 0xMOCK…
// tx hashes derived from sha256(sessionHash || blockN).
type LocalSim struct {
	mu       sync.Mutex
	block    uint64
	store    map[string]*types.ReceiptOnChain // keyed by sessionHash
	explorer string                           // base URL for fake explorer links
}

// NewLocalSim returns a fresh in-memory chain.
func NewLocalSim() *LocalSim {
	return &LocalSim{
		block:    1,
		store:    map[string]*types.ReceiptOnChain{},
		explorer: "https://sepolia.etherscan.io/tx/", // visually familiar URL pattern
	}
}

// AnchorReceipt records the receipt + returns a synthesized tx hash.
func (s *LocalSim) AnchorReceipt(sessionHash, receiptHash string, isSafe bool, policyVersion string) (string, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	rec := &types.ReceiptOnChain{
		SessionHash:   sessionHash,
		ReceiptHash:   receiptHash,
		Timestamp:     now.Unix(),
		IsSafe:        isSafe,
		PolicyVersion: policyVersion,
	}
	s.store[sessionHash] = rec
	// Derived deterministic mock tx hash. Length matches real
	// keccak256/Etherscan hex (64 hex chars + 0x prefix).
	seed := sessionHash + ":" + receiptHash + ":blk:" + uintToStr(s.block)
	h := sha256.Sum256([]byte(seed))
	txHash := "0xMOCK" + hex.EncodeToString(h[:])[5:] // "0xMOCK" + 59 chars = 64-char hex equivalent
	s.block++
	return txHash, s.explorer + txHash, nil
}

// GetReceipt fetches an anchored receipt.
func (s *LocalSim) GetReceipt(sessionHash string) (*types.ReceiptOnChain, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r, ok := s.store[sessionHash]; ok {
		return r, nil
	}
	return nil, nil
}

// EVMStub is the real-slot Sepolia / OmniOne client placeholder. Not
// wired in Phase 1 — Activate when WP-03.U4/U5 brings up Hardhat +
// real contract address.
type EVMStub struct {
	RPCURL          string
	ContractAddress string
}

// NewEVMStub returns the inactive stub. AnchorReceipt errors until
// the real implementation lands.
func NewEVMStub(rpcURL, contractAddr string) *EVMStub {
	return &EVMStub{RPCURL: rpcURL, ContractAddress: contractAddr}
}

// AnchorReceipt — not implemented in Phase 1.
func (e *EVMStub) AnchorReceipt(sessionHash, receiptHash string, isSafe bool, policyVersion string) (string, string, error) {
	return "", "", chainNotImplementedErr
}

// GetReceipt — not implemented in Phase 1.
func (e *EVMStub) GetReceipt(sessionHash string) (*types.ReceiptOnChain, error) {
	return nil, chainNotImplementedErr
}

var chainNotImplementedErr = errChainNotImplemented{}

type errChainNotImplemented struct{}

func (errChainNotImplemented) Error() string {
	return "EVMStub: not implemented in Phase 1 (use CHAIN_PROVIDER=local)"
}

// Receipt is a tiny helper to compute hashes deterministically.
func MakeHashes(sessionID, nonce string, onChain types.ReceiptOnChain) (string, string) {
	sh := sha256.Sum256([]byte(sessionID + "|" + nonce))
	sessionHash := "0x" + hex.EncodeToString(sh[:])

	// receiptHash = sha256 over the canonical JSON of the on-chain
	// record's non-hash fields. We exclude SessionHash + ReceiptHash
	// so the hash is independent of itself.
	canonical := struct {
		Timestamp     int64
		IsSafe        bool
		PolicyVersion string
	}{
		Timestamp:     onChain.Timestamp,
		IsSafe:        onChain.IsSafe,
		PolicyVersion: onChain.PolicyVersion,
	}
	raw, _ := json.Marshal(canonical)
	rh := sha256.Sum256(raw)
	receiptHash := "0x" + hex.EncodeToString(rh[:])
	return sessionHash, receiptHash
}

func uintToStr(u uint64) string {
	// minimal stdlib-free conversion
	if u == 0 {
		return "0"
	}
	var b []byte
	for u > 0 {
		b = append([]byte{byte('0' + u%10)}, b...)
		u /= 10
	}
	return string(b)
}
