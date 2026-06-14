import type { RunScenarioResult } from '../api'
import { SCENARIOS, type ScenarioId } from '../scenarios'
import type { Stamp } from '../types/passport'

export type ReceiverPhase = 'idle' | 'ringing' | 'revealed'

interface Props {
  scenarioId: ScenarioId | null
  phase: ReceiverPhase
  result: RunScenarioResult | null
}

const PHASE_LABEL: Record<ReceiverPhase, string> = {
  idle: '대기',
  ringing: '수신 중',
  revealed: 'Call Passport'
}

const PHASE_DOT: Record<ReceiverPhase, string> = {
  idle: 'bg-zinc-700',
  ringing: 'bg-blue-400 animate-pulse',
  revealed: 'bg-safe'
}

const STAMP_LABEL_REMAP: Record<string, string> = {
  '송금 요구 감지': '고위험 사칭/송금 유도 패턴 감지',
  'AI Agent 권한 (권한 외 목적)': 'AI Agent 목적 권한'
}

const BLOCK_REASON_TEXTS: Record<string, string> = {
  unknown_did: '발신자 신원을 확인할 수 없습니다. 보이스피싱 의심.',
  unauthorized_purpose: '허가된 통화 목적이 아닙니다. 권한 외 목적 통화 차단.'
}

function displayLabel(stamp: Stamp): string {
  return STAMP_LABEL_REMAP[stamp.label] ?? stamp.label
}

function truncateHash(h: string): string {
  if (h.length <= 16) return h
  return `${h.slice(0, 10)}…${h.slice(-6)}`
}

export function ReceiverPanel({ scenarioId, phase, result }: Props) {
  const scenario = scenarioId !== null ? SCENARIOS[scenarioId] : null

  const passport = result?.data.passport
  const isSafe = passport?.outcome === 'SAFE'
  const txHash = passport?.receiptTxHash
  const isMockHash = !!txHash && txHash.toUpperCase().startsWith('0XMOCK')

  return (
    // Outer phone bezel
    <div className="mx-auto w-full max-w-[320px] flex flex-col">
      <div className="rounded-[2.5rem] border-[6px] border-zinc-800 bg-black shadow-[0_8px_24px_rgba(0,0,0,0.5)] p-1 relative">
        {/* Inner screen */}
        <div className="rounded-[2rem] bg-bg-card overflow-hidden flex flex-col min-h-[540px] relative">
          {/* Notch */}
          <div className="absolute top-1.5 left-1/2 -translate-x-1/2 w-16 h-4 bg-black rounded-b-2xl z-10" />

          {/* iOS-style status bar */}
          <div className="pt-3 pb-1.5 px-6 flex items-center justify-between text-[9px] text-zinc-300 font-medium relative z-0">
            <span>9:41</span>
            <span className="flex items-center gap-1">
              {/* signal dots */}
              <span className="flex items-end gap-px">
                <span className="w-0.5 h-1 bg-zinc-300 rounded-sm" />
                <span className="w-0.5 h-1.5 bg-zinc-300 rounded-sm" />
                <span className="w-0.5 h-2 bg-zinc-300 rounded-sm" />
                <span className="w-0.5 h-2.5 bg-zinc-300 rounded-sm" />
              </span>
              {/* battery */}
              <span className="ml-1 inline-flex items-center">
                <span className="w-5 h-2 border border-zinc-400 rounded-sm relative">
                  <span className="absolute inset-y-0 left-0 w-3.5 bg-zinc-300 rounded-sm" />
                </span>
                <span className="w-0.5 h-1 bg-zinc-400 ml-px rounded-sm" />
              </span>
            </span>
          </div>

          {/* Mini header (actor label) */}
          <div className="px-4 pb-2 flex items-center justify-between">
            <div>
              <div className="text-[10px] text-zinc-500 uppercase tracking-wider">Actor 3</div>
              <div className="text-base font-semibold text-zinc-100 mt-0.5">Receiver (시민)</div>
              <div className="text-[10px] text-zinc-500">VOUCH 앱 · Call Passport</div>
            </div>
            <div className="inline-flex items-center gap-1.5 text-[10px] text-zinc-400">
              <span className={`w-1.5 h-1.5 rounded-full ${PHASE_DOT[phase]}`} />
              <span>{PHASE_LABEL[phase]}</span>
            </div>
          </div>

          {/* Body */}
          <div className="flex-1 flex flex-col px-4 pb-2">
            {phase === 'idle' && (
              <div className="flex-1 flex items-center justify-center text-zinc-600 text-sm">
                수신 대기 중
              </div>
            )}

            {phase === 'ringing' && scenario && (
              <div className="flex-1 flex flex-col items-center justify-center">
                <div className="relative w-20 h-20 mb-4">
                  <div className="absolute inset-0 rounded-full bg-blue-500/30 animate-ping" />
                  <div className="relative w-20 h-20 rounded-full bg-blue-500 flex items-center justify-center text-2xl">
                    📞
                  </div>
                </div>
                <div className="text-sm font-semibold text-zinc-100 text-center">
                  {scenario.callerLabel}
                </div>
                <div className="text-xs text-zinc-400 mt-0.5">{scenario.callerOrg}</div>
                <div className="text-xs text-zinc-500 mt-3 inline-flex items-center gap-1.5">
                  <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
                  VOUCH 검증 중
                </div>
              </div>
            )}

            {phase === 'revealed' && scenario && passport && (
              <>
                <div className="text-center mb-3">
                  <div
                    className={`text-5xl mb-2 ${
                      isSafe
                        ? 'drop-shadow-[0_0_16px_rgba(34,197,94,0.5)]'
                        : 'drop-shadow-[0_0_16px_rgba(239,68,68,0.5)]'
                    }`}
                  >
                    {isSafe ? '✅' : '❌'}
                  </div>
                  <h2
                    className={`text-xl font-semibold ${
                      isSafe ? 'text-safe' : 'text-block'
                    }`}
                  >
                    {isSafe ? '안전한 통화' : '차단됨'}
                  </h2>
                </div>

                <div className="bg-bg-base rounded-lg p-3 text-xs space-y-1.5">
                  <div className="flex justify-between">
                    <span className="text-zinc-500">발신자</span>
                    <span className="text-zinc-300 text-right">
                      {scenario.callerOrg} / {scenario.callerLabel}
                    </span>
                  </div>
                  {passport.stamps.map(stamp => (
                    <div key={stamp.label} className="flex justify-between">
                      <span className="text-zinc-500">{displayLabel(stamp)}</span>
                      <span
                        className={stamp.status === 'OK' ? 'text-safe' : 'text-block'}
                      >
                        {stamp.status === 'OK' ? '✓' : '✗'}
                      </span>
                    </div>
                  ))}
                  <div className="flex justify-between border-t border-bg-elevated pt-1.5 mt-1.5">
                    <span className="text-zinc-500">위험도</span>
                    <span
                      className={`inline-flex items-center gap-1 ${
                        isSafe ? 'text-safe' : 'text-block'
                      }`}
                    >
                      <span
                        className={`w-1.5 h-1.5 rounded-full inline-block ${
                          isSafe ? 'bg-safe' : 'bg-block'
                        }`}
                      />
                      {isSafe ? '낮음' : '높음'}
                    </span>
                  </div>
                </div>

                <div className="text-[11px] text-zinc-400 mt-2.5">
                  {isSafe
                    ? '본인 확인 완료. 안전한 통화입니다.'
                    : BLOCK_REASON_TEXTS[passport.blockReason || ''] ||
                      '통화가 차단되었습니다.'}
                </div>

                {/* AI 안전 안내 — Intent Handshake / Predictive Disclosure */}
                {passport.intent_handshake?.safety_summary && (
                  <div className="bg-gradient-to-br from-fuchsia-950/30 to-purple-950/20 border border-fuchsia-500/30 rounded-lg p-2.5 mt-2.5">
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-[9px] uppercase tracking-wider text-fuchsia-300 font-semibold">
                        🤖 AI 안전 안내
                      </span>
                      <span className="text-[8px] font-mono text-fuchsia-300/70 uppercase">
                        {passport.intent_handshake.source === 'live' ? '● LIVE' : '○ FALLBACK'}
                      </span>
                    </div>
                    <div className="text-[11px] text-fuchsia-100/90 leading-snug">
                      {passport.intent_handshake.safety_summary}
                    </div>
                    {passport.intent_handshake.forbidden_requests && passport.intent_handshake.forbidden_requests.length > 0 && (
                      <div className="mt-1.5 pt-1.5 border-t border-fuchsia-500/20">
                        <div className="text-[9px] text-fuchsia-300/80 mb-0.5">요청 금지 항목:</div>
                        <div className="text-[10px] text-zinc-300 leading-snug">
                          {passport.intent_handshake.forbidden_requests.slice(0, 4).join(' · ')}
                        </div>
                      </div>
                    )}
                  </div>
                )}

                {txHash &&
                  (isMockHash ? (
                    <div className="bg-bg-base rounded-lg p-2.5 mt-2.5 text-[10px]">
                      <div className="text-zinc-500">
                        영수증 해시 · 시뮬레이션 (데모 환경)
                      </div>
                      <div className="text-zinc-300 font-mono truncate mt-0.5">
                        {truncateHash(txHash)}
                      </div>
                    </div>
                  ) : (
                    <a
                      href={passport.explorerUrl || '#'}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="bg-bg-base rounded-lg p-2.5 mt-2.5 text-[10px] block hover:bg-bg-elevated transition-colors"
                    >
                      <div className="text-zinc-500">영수증 해시 · Sepolia 테스트넷</div>
                      <div className="text-zinc-300 font-mono truncate mt-0.5">
                        {truncateHash(txHash)}
                      </div>
                      <div className="text-[9px] text-zinc-600 mt-0.5">
                        Etherscan에서 확인 →
                      </div>
                    </a>
                  ))}
              </>
            )}
          </div>

          {/* Footer label */}
          <div className="px-4 pb-3 text-center text-[9px] text-zinc-600 leading-tight">
            앱/Broker 레벨 Call Passport
            <br />
            OmniOne CX + Open DID 기반
          </div>

          {/* iOS-style home indicator */}
          <div className="pb-1.5 flex justify-center">
            <div className="w-24 h-1 bg-zinc-700 rounded-full" />
          </div>
        </div>
      </div>
    </div>
  )
}
