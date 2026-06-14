# internal/omnione — OmniOne integration adapter facade

**Status:** scaffolding · 2026-05-28 · matches Dev-related/00-strategy-reframe.md v2 (Verification Broker thesis) + 08-omnione-platform-mapping.md.

This package provides three adapter interfaces that the session-svc orchestrator uses to talk to OmniOne primitives. Each adapter has a mock default and a real-slot path. The mock path satisfies the demo without external dependencies; the real path activates when credentials are present.

## Three adapters

| Adapter | Purpose | Modes |
|---|---|---|
| `CXAdapter` | Receiver-side Mobile ID verification (OmniOne CX 표준인증창 + VC-Verifier) | `mock` (default) · `real` (requires `OMNIONE_CX_LICENSE_KEY`) |
| `OpenDIDAdapter` | Caller-side institution DID + AgentAuthorization resolution (Open DID / OmniOne Enterprise) | `mock` (default — Go-native didregistry) · `real` (Java SDK sidecar — deferred to 결선) |
| `ChainAdapter` | Receipt anchoring (ZTCVReceiptAnchor.sol) | `local` (in-memory sim) · `sepolia` (production EVM testnet) · `omnione` (BESU via dev portal — pending) |

## Env switches

```
OMNIONE_CX_MODE         = mock | real         # default: mock
OMNIONE_CX_LICENSE_KEY  = <key>               # required if MODE=real

OMNIONE_OPENDID_MODE    = mock | real         # default: mock
                                              # NOTE: real mode not yet wired

CHAIN_PROVIDER          = local | sepolia | omnione  # default: local
CHAIN_RPC_SEPOLIA       = <url>               # required if PROVIDER=sepolia
DEPLOYER_PK             = <pk>                # required if PROVIDER=sepolia
ZTCV_RECEIPT_ANCHOR_ADDRESS = <addr>          # required if PROVIDER=sepolia
CHAIN_RPC_OMNIONE       = <url>               # required if PROVIDER=omnione
```

## Honest framing (per v2 thesis)

- ⊢ **HACK** — substrate exists per hackathon guidebook
- ⊢ **IMPL** — adapter package + Go-native default impls implemented today
- ⊨ **PROP** — real-slot activation when credentials arrive
- ⊥ **RSCH** — Java sidecar (OpenDID real) + OmniOne Chain dev portal access still to acquire

Do **NOT** claim production OmniOne integration in pitch unless the env vars actually point to a real backend and a request has been successfully verified end-to-end.

## Activation path

1. Acquire OACX dev license from 라온시큐어 / 한국디지털인증협회 → set `OMNIONE_CX_MODE=real` + `OMNIONE_CX_LICENSE_KEY=…`
2. Acquire OmniOne Chain dev portal access (didalliance.org / hackathon@opendid.org) → upload `ZTCVReceiptAnchor.sol` → set `CHAIN_PROVIDER=omnione` + `CHAIN_RPC_OMNIONE=…`
3. 결선 dev phase: stand up did-issuer-server + did-verifier-server Java sidecar → wire OpenDIDAdapter real impl → set `OMNIONE_OPENDID_MODE=real`

## References

- `docs-internal/Dev-related/00-strategy-reframe.md` v2 — Verification Broker thesis
- `docs-internal/Dev-related/02-omnione-cx.md` — CX integration plan
- `docs-internal/Dev-related/04-open-did.md` — Open DID data model
- `docs-internal/Dev-related/08-omnione-platform-mapping.md` — verified facts + ZTCV mapping
