# VOUCH — Verified Outgoing-call Universal Compliance Hub

> **Pre-call mutual identity verification** — proves the caller's identity, authority, and purpose *before the phone rings.*

**Pitch:** *Existing solutions doubt the call after you receive it. VOUCH proves the call before you receive it.*

[![License: CC BY 4.0](https://img.shields.io/badge/License-CC%20BY%204.0-lightgrey.svg)](https://creativecommons.org/licenses/by/4.0/)

---

## Origin & Authorship

| | |
|---|---|
| **Concept & Architecture** | David Seo (서인석) |
| **Organization** | GEM² · gemsquared.ai / Gineers.ai |
| **Conceived** | 2026-05-20 |
| **First hackathon submission** | 2026-05-31 — 2026 Blockchain & AI Hackathon, Track 2 |
| **First public release** | 2026-06-14 |
| **Contact** | david@gemsquared.ai |

This repository is the authoritative public origin of the VOUCH protocol concept and its Go reference implementation. Licensed under [CC BY 4.0](./LICENSE) — attribution required on reuse.

---

## The Problem

Voice phishing losses in Korea reached **₩1.133 trillion in 2025** (+56.1% YoY — a 5-year record). Institution-impersonation fraud (prosecutors, financial supervisors, banks) accounts for **76–77% of all cases**. The 2030 generation accounts for 52% of victims; the share suffering losses over ₩100M doubled from 17% to 34%.

Current defenses are reactive — they block or flag *after* the call connects. VOUCH intervenes *before*.

Two forces make 2026 the right moment:
- **Korea's AI Basic Act (2026-01-22)** mandates disclosure of AI involvement in calls — per-call "is this a human or AI, and what is its authority?" requires a standard verification mechanism.
- **Mobile ID national rollout (2025-03~)** + OmniOne Digital ID at 45M-person scale — for the first time, both caller-side and receiver-side identity infrastructure exist simultaneously.

---

## What VOUCH Does

VOUCH sits between caller and receiver as a **Verification Broker**. Before the call connects:

1. The calling institution submits a signed **CallerProof** envelope (`org_did · caller_did · purpose · intent`).
2. The Broker runs a **9-step verification pipeline** (see below).
3. The receiver's phone displays a one-time **VOUCH Passport** — a call permit showing who is calling, why, and whether it passed all checks.
4. Every outcome (SAFE or BLOCK) is permanently recorded as a **Chain Receipt** on blockchain.

```
Caller (institution/AI agent)          VOUCH Broker              Receiver (citizen)
        │                                   │                            │
        │──── signed CallerProof ──────────►│                            │
        │                                   │◄─── Mobile ID verify ─────►│
        │                                   │                            │
        │                          9-step pipeline                       │
        │                                   │                            │
        │                          ┌────────▼────────┐                   │
        │                          │  VOUCH Passport  │──────────────────►│
        │                          │  (call permit)   │                   │
        │                          └─────────────────┘                   │
        │                                   │                            │
        │                     Chain Receipt (on-chain, permanent)        │
```

### 9-Step Verification Pipeline

| Step | What | Live status |
|------|------|-------------|
| 1 | Receive CallerProof (Ed25519-signed envelope) | ✅ Live |
| 2 | Phone number ↔ org DID binding check | △ Partial |
| 3 | DID · VC verification (OmniOne Open DID) | △ Partial |
| 4 | AI agent purpose · authority check (VC `allowed_purposes`) | △ Partial |
| 5 | Risk pattern lookup (Layer 1 rule classifier) | ✅ Live |
| 6 | **Intent Handshake** — AI pre-discloses call intent (Claude Haiku live) | ✅ Live + LLM |
| 7 | Receiver Mobile ID verification (OmniOne CX) | △ Mock-fallback |
| 8 | Generate VOUCH Passport | ✅ Live |
| 9 | Chain Receipt on Sepolia testnet (`ZTCVReceiptAnchor.sol`) | ✅ Live |

**Deployed contract:** [`0x70fc086A3f4e91dA3f7e3Aeb4D2C806DdF3dED04`](https://sepolia.etherscan.io/address/0x70fc086A3f4e91dA3f7e3Aeb4D2C806DdF3dED04) — Sepolia Testnet

### Three Demo Scenarios

| # | Caller | Outcome |
|---|---|---|
| 1 | Kakao Bank AI counselor (registered DID, valid purpose) | ✅ **SAFE** |
| 2 | Fake prosecutor (no registered DID) | ✗ **BLOCK** |
| 3 | Kakao Bank security AI (unauthorized scope) | ✗ **BLOCK** |

Both SAFE and BLOCK outcomes are recorded on-chain — **non-repudiation applies to all outcomes.**

---

## Business Model

VOUCH operates as both a **protocol** and an **operator**:

- **Protocol layer** — the CallerProof data model, 9-step verification procedure, and Chain Receipt anchoring spec are open and publishable as a compliance standard.
- **Operator layer** — the Broker pipeline + Intent Handshake service are a standalone SaaS business.

**Revenue model:** Per-call verification fee charged to outgoing-call operators.

| Customer segment | Why they pay |
|---|---|
| Banks · Insurance · Telecom | Regulatory compliance; reduce impersonation fraud exposure |
| Public agencies | AI Basic Act disclosure mandate (2026-01-22) |
| AI Agent operators | Call centre AI market: $60.5M → $275.8M; no identification standard exists today |

**Market entry:** Korea (₩1.133T damage, single dominant fraud type — institution impersonation). Expansion: any jurisdiction deploying national Mobile ID + DID infrastructure.

**OmniOne ecosystem effect:** Every verified call extends OmniOne from identity *issuance* into identity *verification infrastructure* — DID·VC resolution + Chain Receipt management as a per-call trust layer.

---

## Architecture

Single Go module, four microservice binaries:

```
┌─────────────────────────────────────────────────────┐
│  frontend-d/pwa  (React + TypeScript + Tailwind PWA) │
│  Caller Admin console · Receiver mobile view        │
└───────────────────────┬─────────────────────────────┘
                        │ REST + SSE
┌───────────────────────▼─────────────────────────────┐
│  session-svc  (orchestrator + 9-step pipeline)       │
└──────┬──────────┬──────────┬──────────┬─────────────┘
       │          │          │          │
  identity-svc  did-svc  risk-chain-svc  chain-adapter
  OmniOne CX   Open DID   Layer1 rules   ZTCVReceiptAnchor
  Mobile ID    resolver   + Claude LLM   Sepolia / OmniOne
  verification + AI-agent              Chain
               reputation
```

**Key design choices:**
- `ZTCVReceiptAnchor.sol` stores only 5 fields on-chain: `sessionHash · receiptHash · isSafe · policyVersion · timestamp` — **Zero PII on-chain.**
- LLM adapter is pluggable: currently Claude Haiku; designed to swap to sovereign/self-hosted LLM post-hackathon.
- Chain target: currently Sepolia testnet; config flip to OmniOne Chain for production.

---

## Live Demo

> The demo runs against Sepolia testnet with Claude Haiku live LLM (some steps mock-fallback pending full OmniOne integration).

- **Signup flow** (OmniOne CX mockup): https://vouch.gemsquared.ai/signup/
- **Verification flow** (main surface): https://vouch.gemsquared.ai/flow/

---

## Quick Start

```bash
# 1. Clone
git clone https://github.com/gem-squared/gem2-ZTCV-advance.git
cd gem2-ZTCV-advance

# 2. Configure environment
cp .env.example .env   # then fill in real values
# Required: LLM_API_ANTHROPIC (or set LLM_PROVIDER_WORKER=mock)

# 3. Boot stack (Docker — 5 containers)
make dev

# 4. Open demo UI
open http://localhost:3000

# 5. Run tests
make test
```

See `Makefile` for all targets (`make help`).

**Environment variables** (`.env.example` lists all keys):
- `LLM_PROVIDER_WORKER=mock` — run without a real LLM key
- `CHAIN_PROVIDER=local` — use local Hardhat node instead of Sepolia
- `OMNIONE_CX_MODE=mock` — run without OmniOne CX credentials

---

## Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.25, SQLite (per-service), SSE |
| Smart contract | Solidity 0.8.27, Hardhat, Sepolia testnet |
| Frontend | React 18, TypeScript, Vite, Tailwind CSS, PWA |
| Identity | OmniOne Open DID (Ed25519 VC), OmniOne CX, Mobile ID |
| AI | Anthropic Claude Haiku (Intent Handshake) |
| Chain | Ethereum Sepolia → OmniOne Chain (target) |

---

## License

[Creative Commons Attribution 4.0 International (CC BY 4.0)](./LICENSE)

Copyright 2026 David Seo · GEM² (gemsquared.ai / Gineers.ai)

Attribution required on reuse. See [NOTICE](./NOTICE) for full attribution requirements and third-party component credits.
