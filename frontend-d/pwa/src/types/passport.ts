// CallPassport TypeScript types — MUST match the Go CallPassport JSON shape
// at internal/types/passport.go. Backend response shape from
// POST /api/scenarios/run?n=1|2|3 wraps the passport in a {passport, ...}
// envelope; we type the passport strictly and leave wider session/risk
// data loose for now.

export type StampStatus = 'OK' | 'FAIL'

export interface Stamp {
  label: string
  status: StampStatus
  detail?: string
}

export type CallPassportOutcome = 'SAFE' | 'FAILED'

export interface IntentManifest {
  expected_requests: string[]
  forbidden_requests: string[]
  safety_summary: string
  source: 'live' | 'fallback'
  provider?: string
  generated_at: string
}

export interface CallPassport {
  sessionId: string
  issuedAt: string
  expiresAt: string
  outcome: CallPassportOutcome
  stamps: Stamp[]
  blockReason?: string
  receiptTxHash?: string
  explorerUrl?: string
  callerDid?: string
  callerOrg?: string
  callerPurpose?: string
  // Intent Handshake (Step 7 of 9-step pipeline). Present when backend
  // ran the step; absent in fixture-only fallback paths.
  intent_handshake?: IntentManifest
  intent_manifest_hash?: string
}

export interface ScenariosRunResponse {
  passport: CallPassport
  session?: Record<string, unknown>
  tx_hash?: string
  block_reason?: string
}

export const STAMP_LABELS = {
  ORG_DID: '기관 DID',
  AGENT_AUTH: 'AI Agent 권한',
  MOBILE_ID: '수신자 Mobile ID',
  CHAIN_RECEIPT: 'Chain Receipt',
  MISSING_DID: '발신자 DID 없음',
  TRANSFER_DEMAND: '송금 요구 감지',
  BLOCK_RECEIPT: '차단 영수증 기록',
  UNAUTHORIZED_PURPOSE: 'AI Agent 권한 (권한 외 목적)'
} as const
