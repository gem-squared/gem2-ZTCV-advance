#!/usr/bin/env bash
# Extract the compiled ABI for embedding into the Go backend via //go:embed.
# Run after `npm run compile`.

set -euo pipefail

CONTRACT="ZTCVReceiptAnchor"
ARTIFACT="artifacts/contracts/${CONTRACT}.sol/${CONTRACT}.json"
OUT="../internal/chain/embedded/ztcv_receipt_anchor.json"

if [ ! -f "$ARTIFACT" ]; then
    echo "Artifact not found: $ARTIFACT"
    echo "Run 'npm run compile' first."
    exit 1
fi

mkdir -p "$(dirname "$OUT")"
cp "$ARTIFACT" "$OUT"
echo "✓ ABI extracted → $OUT"
