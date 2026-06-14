import { SCENARIOS, type ScenarioId } from '../scenarios'

export type CallerPhase = 'idle' | 'preparing' | 'submitting' | 'submitted'

interface Props {
  scenarioId: ScenarioId | null
  phase: CallerPhase
}

const PHASE_LABEL: Record<CallerPhase, string> = {
  idle: '대기',
  preparing: 'CallerProof 준비 중',
  submitting: 'Broker로 송신 중',
  submitted: '송신 완료'
}

const PHASE_DOT: Record<CallerPhase, string> = {
  idle: 'bg-zinc-700',
  preparing: 'bg-blue-400 animate-pulse',
  submitting: 'bg-blue-400 animate-pulse',
  submitted: 'bg-safe'
}

const SCENARIO_CALLER_DATA: Record<ScenarioId, {
  org_did: string
  caller_did: string
  purpose: string
  number: string
  intent_summary: string
}> = {
  1: {
    org_did: 'did:opendid:org:kakaobank',
    caller_did: 'did:opendid:agent:kakaobank-ai-loan-counselor-001',
    purpose: 'loan_consultation',
    number: '02-3456-7890',
    intent_summary: '대출 상담 안내'
  },
  2: {
    org_did: '(unregistered)',
    caller_did: '(unregistered)',
    purpose: 'investigation',
    number: '02-0000-0000',
    intent_summary: '송금 요구'
  },
  3: {
    org_did: 'did:opendid:org:kakaobank',
    caller_did: 'did:opendid:agent:kakaobank-ai-security-alert-007',
    purpose: 'loan_consultation',
    number: '02-3456-7891',
    intent_summary: '대출 가입 유도'
  }
}

function truncateMid(s: string, head = 22, tail = 8): string {
  if (s.length <= head + tail + 1) return s
  return `${s.slice(0, head)}…${s.slice(-tail)}`
}

// Inline SVG call-center agent icon (silhouette with headset + mic boom)
function CallCenterAgentIcon({ className = 'w-16 h-16' }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 64 64"
      fill="currentColor"
      className={className}
      aria-label="call-center agent"
    >
      {/* Body / shoulders */}
      <path d="M14 60 Q14 36 32 36 Q50 36 50 60 Z" />
      {/* Head */}
      <circle cx="32" cy="22" r="11" />
      {/* Headset band (arc over head) */}
      <path
        d="M14 26 Q14 6 32 6 Q50 6 50 26"
        stroke="currentColor"
        strokeWidth="4"
        fill="none"
        strokeLinecap="round"
      />
      {/* Ear cushions */}
      <rect x="11" y="20" width="5" height="10" rx="1.5" />
      <rect x="48" y="20" width="5" height="10" rx="1.5" />
      {/* Mic boom curving from left ear down */}
      <path
        d="M13 28 Q11 42 24 42"
        stroke="currentColor"
        strokeWidth="2.5"
        fill="none"
        strokeLinecap="round"
      />
      {/* Mic bulb */}
      <circle cx="25" cy="42" r="2.5" />
    </svg>
  )
}

export function CallerPanel({ scenarioId, phase }: Props) {
  const scenario = scenarioId !== null ? SCENARIOS[scenarioId] : null
  const data = scenarioId !== null ? SCENARIO_CALLER_DATA[scenarioId] : null

  return (
    <div className="bg-bg-card rounded-md overflow-hidden flex flex-col h-full min-h-[520px] border border-zinc-800/80 shadow-[0_2px_8px_rgba(0,0,0,0.3)]">
      {/* Server-rack header bar */}
      <div className="bg-zinc-900/80 border-b border-zinc-800 px-4 py-2.5 flex items-center justify-between">
        <div className="flex items-center gap-2 text-[10px] text-zinc-400 font-mono uppercase tracking-wider">
          <span>📡</span>
          <span>Caller · VOUCH SDK</span>
        </div>
        <div className="inline-flex items-center gap-1.5 text-[10px] text-zinc-400">
          <span className={`w-1.5 h-1.5 rounded-full ${PHASE_DOT[phase]}`} />
          <span>{PHASE_LABEL[phase]}</span>
        </div>
      </div>

      {/* Call-center agent icon section */}
      <div className="flex flex-col items-center pt-6 pb-5 border-b border-zinc-800/60 bg-gradient-to-b from-zinc-900/20 to-transparent">
        <div className="text-zinc-300">
          <CallCenterAgentIcon className="w-14 h-14" />
        </div>
        <div className="text-xs text-zinc-300 mt-2 font-medium">Call Center Agent</div>
        <div className="text-[10px] text-zinc-500 mt-0.5">Actor 1 · 기관 발신 측</div>
      </div>

      {/* Actor label + scenario info */}
      <div className="px-5 pt-4 pb-2">
        <div className="text-lg font-semibold text-zinc-100">Caller (기관)</div>
        <div className="text-[10px] text-zinc-500 mt-0.5">
          기관 관리자 콘솔 · 제안 UI / Proposed admin view
        </div>
      </div>

      {/* SDK system box (technical content) */}
      <div className="px-5 pb-3 flex-1 flex flex-col gap-3">
        {scenario && data ? (
          <>
            <div className="bg-bg-base rounded-md p-3 border border-zinc-800 text-sm">
              <div className="text-zinc-500 text-[10px] uppercase tracking-wider mb-1.5">
                기관 / 발신자
              </div>
              <div className="text-zinc-100 font-medium">{scenario.callerOrg}</div>
              <div className="text-zinc-300 text-xs mt-0.5">{scenario.callerLabel}</div>
              <div className="text-zinc-500 text-xs mt-1 font-mono">{data.number}</div>
            </div>

            <div className="bg-bg-base rounded-md p-3 border border-zinc-800 text-xs font-mono text-zinc-400 space-y-1">
              <div>
                <span className="text-zinc-600">org_did:    </span>
                <span className="text-zinc-300">{truncateMid(data.org_did)}</span>
              </div>
              <div>
                <span className="text-zinc-600">caller_did: </span>
                <span className="text-zinc-300">{truncateMid(data.caller_did)}</span>
              </div>
              <div>
                <span className="text-zinc-600">purpose:    </span>
                <span className="text-zinc-300">{data.purpose}</span>
              </div>
              <div>
                <span className="text-zinc-600">intent:     </span>
                <span className="text-zinc-300">{data.intent_summary}</span>
              </div>
            </div>

            <div className="bg-bg-base rounded-md p-3 border border-zinc-800 text-xs">
              <div className="text-zinc-500 text-[10px] uppercase tracking-wider mb-1.5">
                CallerProof
              </div>
              {phase === 'idle' ? (
                <div className="text-zinc-600">— 시나리오 선택 대기 —</div>
              ) : phase === 'preparing' ? (
                <div className="text-blue-400 flex items-center gap-2">
                  <span className="inline-block w-3 h-3 border-2 border-blue-400 border-t-transparent rounded-full animate-spin" />
                  Ed25519 서명 준비
                </div>
              ) : phase === 'submitting' ? (
                <div className="text-blue-400 flex items-center gap-2">
                  <span className="inline-block w-3 h-3 border-2 border-blue-400 border-t-transparent rounded-full animate-spin" />
                  Broker로 전송 중
                </div>
              ) : (
                <div className="text-safe">✓ CallerProof 송신 완료</div>
              )}
            </div>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center text-zinc-600 text-sm">
            시나리오를 선택하세요
          </div>
        )}
      </div>

      <div className="flex-1" />

      <div className="border-t border-zinc-800/60 px-5 py-2.5 bg-zinc-900/40">
        <div className="text-[10px] text-zinc-600 font-mono text-center">
          VOUCH Caller Admin · 제안 UI (Proposed)
        </div>
      </div>
    </div>
  )
}
