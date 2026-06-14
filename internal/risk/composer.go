// Package risk owns the composer + bridges Layer 1 and Layer 2.
package risk

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/gem-squared/gem2-ZTCV/internal/types"
)

// PolicyVersion is the running policy hash recorded in every receipt.
const PolicyName = "ztcv-policy-v0.1.0"

// PolicyVersionHex is keccak-equivalent hex of "ztcv-policy-v0.1.0"
// (we use sha256 for Phase 1 simulation — Solidity uses keccak256 but
// we don't need cryptographic equivalence here, just a stable hex).
func PolicyVersionHex() string {
	h := sha256.Sum256([]byte(PolicyName))
	return "0x" + hex.EncodeToString(h[:])
}

// Compose returns the ComposedVerdict using conservative-block:
//
//	BLOCK  if EITHER layer says BLOCK
//	HIGH   if EITHER layer says HIGH (no BLOCK)
//	MEDIUM if EITHER says MEDIUM (no HIGH/BLOCK)
//	LOW    only if BOTH say LOW
//
// Disagreement is recorded when L1.verdict != L2.verdict.
func Compose(l1 types.Layer1Result, l2 types.Layer2Result) *types.ComposedVerdict {
	final := types.RiskLOW
	switch {
	case l1.Verdict == types.RiskBLOCK || l2.Verdict == types.RiskBLOCK:
		final = types.RiskBLOCK
	case l1.Verdict == types.RiskHIGH || l2.Verdict == types.RiskHIGH:
		final = types.RiskHIGH
	case l1.Verdict == types.RiskMEDIUM || l2.Verdict == types.RiskMEDIUM:
		final = types.RiskMEDIUM
	}
	return &types.ComposedVerdict{
		Layer1:        l1,
		Layer2:        l2,
		Final:         final,
		Disagreement:  l1.Verdict != l2.Verdict,
		PolicyVersion: PolicyVersionHex(),
		ComposedAt:    time.Now().UTC(),
	}
}
