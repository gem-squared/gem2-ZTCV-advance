// Phase orchestration timeline for /flow/ cinematic mode.
// All offsets are cumulative milliseconds from scenario pick.
//
// The 7-phase narrative:
//   1. make-call           — caller dials, receiver phone rings
//   2. caller-proof        — CallerProof envelope flies to Broker
//   3. receiver-did        — receiver DID flies to Broker
//   4. (broker internal)   — 9-step checklist runs inside BrokerPanel
//   5. verification-mirror — broker verdict mirrors back to receiver
//   6. chain-anchor        — receipt anchor drops into the Ledger element
//   7. passport-reveal     — Receiver renders the final Call Passport

export const PHASE_TIMING_MS = {
  callerPreparing: 0,        // Caller card highlights
  makeCall: 400,             // Phase 1 — line draws Caller→Receiver
  receiverRinging: 900,      // Receiver shows incoming call UI
  callerSubmitting: 1200,    // Phase 2 — CallerProof line draws Caller→Broker
  receiverDID: 2000,         // Phase 3 — Receiver DID line draws Receiver→Broker
  callerSubmitted: 2400,     // Caller card shows ✓ + Broker.receiving
  brokerVerifying: 2800,     // Phase 4 — BrokerPanel 9-step animation starts (~4500ms)
  verificationMirror: 7300,  // Phase 5 — Broker→Receiver verdict line
  chainAnchor: 7700,         // Phase 6 — Broker→Ledger gold drop
  brokerCompleteReceiverRevealed: 8200, // Phase 7 — Passport renders on Receiver
} as const

// Per-line draw-on animation duration. Picked to feel deliberate
// without dragging.
export const LINE_DRAW_MS = 500
// How long a flow line stays visible after its phase fires (before
// fading). Set generously so the judge can trace the full sequence.
export const LINE_HOLD_MS = 6000
// Gold-drop animation duration when phase 6 fires.
export const CHAIN_DROP_MS = 700

// Hard ceiling on total flow before idle reset. Used to throttle
// auto-replay mode.
export const TOTAL_FLOW_MS = 9000
