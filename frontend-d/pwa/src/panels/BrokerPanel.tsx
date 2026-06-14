import { useEffect, useState, useRef } from 'react'
import type { RunScenarioResult } from '../api'
import type { ScenarioId } from '../scenarios'

export type BrokerPhase = 'idle' | 'receiving' | 'verifying' | 'complete'

interface Props {
  scenarioId: ScenarioId | null
  phase: BrokerPhase
  result: RunScenarioResult | null
}

// Full 9-step pipeline. Step 7 (Intent Handshake / Predictive
// Disclosure) is the AI-driven step — animated distinctly to make the
// "AI part" visually unmistakable for judges.
const CHECKLIST = [
  '발신 번호 조회',                      // 1
  'Caller DID binding 확인',             // 2
  '기관 등록 상태 확인',                  // 3
  'AI Agent VC 검증',                    // 4
  '통화 목적 권한 확인',                  // 5
  '신고/위험 패턴 조회',                  // 6
  '통화 의도 사전 공개 (AI 분석)',         // 7 ← Intent Handshake (distinctive)
  '수신자 Mobile ID 확인',                // 8
  'Chain Receipt 준비'                   // 9
]

// Index of the Intent Handshake step (0-based) — used to apply
// distinct visual treatment.
const INTENT_STEP_INDEX = 6

type OutcomeMark = 'OK' | 'FAIL' | 'NA'
type ItemState = 'hidden' | 'checking' | 'resolved'

// 9-element outcome vectors per scenario. The Intent Handshake step
// (index 6) ALWAYS runs — even for blocked scenarios it produces a
// safety-warning manifest, so judges always see the AI step animate.
const PER_SCENARIO_OUTCOMES: Record<ScenarioId, OutcomeMark[]> = {
  1: ['OK', 'OK',   'OK', 'OK', 'OK',   'OK', 'OK', 'OK', 'OK'], // SAFE — all 9 OK
  2: ['OK', 'FAIL', 'NA', 'NA', 'NA',   'NA', 'OK', 'NA', 'OK'], // BLOCK unknown_did — AI still issues warning
  3: ['OK', 'OK',   'OK', 'OK', 'FAIL', 'NA', 'OK', 'NA', 'OK']  // BLOCK unauthorized_purpose — AI still issues warning
}

const ITEM_INTERVAL_MS = 280
const CHECKING_DURATION_MS = 200
// Intent Handshake takes longer (visualizes the LLM call) so the AI
// step stands out from the rule-based steps.
const INTENT_CHECKING_DURATION_MS = 1100

const PHASE_LABEL: Record<BrokerPhase, string> = {
  idle: '대기',
  receiving: 'CallerProof 수신',
  verifying: '검증 진행',
  complete: '완료'
}

const PHASE_DOT: Record<BrokerPhase, string> = {
  idle: 'bg-zinc-700',
  receiving: 'bg-blue-400 animate-pulse',
  verifying: 'bg-blue-400 animate-pulse',
  complete: 'bg-safe'
}

interface LogEntry {
  ts: string
  text: string
  tone: 'info' | 'ok' | 'fail'
}

function tsNow(): string {
  const d = new Date()
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  const ss = String(d.getSeconds()).padStart(2, '0')
  const ms = String(d.getMilliseconds()).padStart(3, '0')
  return `${hh}:${mm}:${ss}.${ms}`
}

// Inline SVG ZTCV system / shield icon for the broker header
function BrokerSystemIcon({ className = 'w-7 h-7' }: { className?: string }) {
  return (
    <svg viewBox="0 0 32 32" fill="none" className={className} aria-label="ztcv broker">
      {/* Shield outline */}
      <path
        d="M16 3 L27 7 V15 Q27 22 16 29 Q5 22 5 15 V7 Z"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinejoin="round"
        fill="currentColor"
        fillOpacity="0.1"
      />
      {/* Inner checkmark */}
      <path
        d="M11 16 L14.5 19 L21 12"
        stroke="currentColor"
        strokeWidth="2.5"
        strokeLinecap="round"
        strokeLinejoin="round"
        fill="none"
      />
    </svg>
  )
}

export function BrokerPanel({ scenarioId, phase, result }: Props) {
  const outcomes = scenarioId !== null ? PER_SCENARIO_OUTCOMES[scenarioId] : null
  const [itemStates, setItemStates] = useState<ItemState[]>(() =>
    CHECKLIST.map(() => 'hidden')
  )
  const [log, setLog] = useState<LogEntry[]>([])
  const logRef = useRef<HTMLDivElement>(null)
  const scenarioRef = useRef<ScenarioId | null>(null)

  useEffect(() => {
    if (scenarioId !== scenarioRef.current) {
      scenarioRef.current = scenarioId
      setItemStates(CHECKLIST.map(() => 'hidden'))
      setLog([])
    }

    if (phase !== 'verifying' || scenarioId === null || !outcomes) return

    let cancelled = false
    const timeouts: ReturnType<typeof setTimeout>[] = []

    const pushLog = (text: string, tone: LogEntry['tone'] = 'info') =>
      setLog(prev => [...prev, { ts: tsNow(), text, tone }])

    const setItemAt = (i: number, s: ItemState) =>
      setItemStates(prev => {
        if (cancelled || prev[i] === s) return prev
        const next = [...prev]
        next[i] = s
        return next
      })

    pushLog(`scenario=${scenarioId} verification started`, 'info')

    // Track cumulative time so the Intent step (which takes longer)
    // shifts every subsequent step further out in time.
    let cumulative = 0
    for (let i = 0; i < outcomes.length; i++) {
      const startTime = cumulative
      const isNA = outcomes[i] === 'NA'
      const isIntent = i === INTENT_STEP_INDEX
      const checkDuration = isIntent ? INTENT_CHECKING_DURATION_MS : CHECKING_DURATION_MS

      if (isNA) {
        const t = setTimeout(() => {
          if (cancelled) return
          setItemAt(i, 'resolved')
          pushLog(`${CHECKLIST[i]} — skipped (short-circuit)`, 'info')
        }, startTime)
        timeouts.push(t)
        cumulative += ITEM_INTERVAL_MS
      } else {
        const t1 = setTimeout(() => {
          if (cancelled) return
          setItemAt(i, 'checking')
          if (isIntent) {
            pushLog(`${CHECKLIST[i]} → Anthropic Claude Haiku 호출 …`, 'info')
          } else {
            pushLog(`${CHECKLIST[i]} …`, 'info')
          }
        }, startTime)
        timeouts.push(t1)

        const t2 = setTimeout(() => {
          if (cancelled) return
          setItemAt(i, 'resolved')
          const tone: LogEntry['tone'] = outcomes[i] === 'OK' ? 'ok' : 'fail'
          const symbol = outcomes[i] === 'OK' ? '✓' : '✗'
          if (isIntent) {
            pushLog(`${CHECKLIST[i]} ${symbol}  (manifest 생성 완료)`, tone)
          } else {
            pushLog(`${CHECKLIST[i]} ${symbol}`, tone)
          }
        }, startTime + checkDuration)
        timeouts.push(t2)
        cumulative += ITEM_INTERVAL_MS + (isIntent ? INTENT_CHECKING_DURATION_MS - CHECKING_DURATION_MS : 0)
      }
    }

    return () => {
      cancelled = true
      for (const t of timeouts) clearTimeout(t)
    }
  }, [scenarioId, phase, outcomes])

  useEffect(() => {
    if (logRef.current) {
      logRef.current.scrollTop = logRef.current.scrollHeight
    }
  }, [log])

  const verdict = result?.data.passport.outcome
  const txHash = result?.data.passport.receiptTxHash
  const isMockHash = !!txHash && txHash.toUpperCase().startsWith('0XMOCK')

  return (
    <div className="bg-bg-card rounded-lg overflow-hidden flex flex-col h-full min-h-[520px] border border-blue-900/30 shadow-[0_2px_12px_rgba(59,130,246,0.08)]">
      {/* System dashboard header bar */}
      <div className="bg-gradient-to-r from-blue-950/40 via-zinc-900/80 to-blue-950/40 border-b border-blue-900/40 px-4 py-2.5 flex items-center justify-between">
        <div className="flex items-center gap-2.5">
          <span className="text-blue-400">
            <BrokerSystemIcon className="w-5 h-5" />
          </span>
          <div className="text-[10px] text-zinc-300 font-mono uppercase tracking-wider font-semibold">
            VOUCH · Verification Broker
          </div>
        </div>
        <div className="inline-flex items-center gap-1.5 text-[10px] text-zinc-400">
          <span className={`w-1.5 h-1.5 rounded-full ${PHASE_DOT[phase]}`} />
          <span>{PHASE_LABEL[phase]}</span>
        </div>
      </div>

      {/* Title */}
      <div className="px-5 pt-4 pb-2 border-b border-zinc-800/60">
        <div className="text-xs text-zinc-500 uppercase tracking-wider mb-1">Actor 2</div>
        <div className="text-lg font-semibold text-zinc-100">Verification Broker</div>
        <div className="text-[10px] text-zinc-500 mt-1">
          OmniOne CX + Open DID 기반 발신 검증 · signed pre-call trust layer
        </div>
      </div>

      {/* Checklist */}
      <div className="p-5">
        <ul className="space-y-2.5 text-sm">
          {CHECKLIST.map((label, i) => {
            const state = itemStates[i]
            const outcome = outcomes ? outcomes[i] : 'NA'
            const visible = state !== 'hidden'
            const checking = state === 'checking'
            const resolved = state === 'resolved'
            const isSkipped = outcome === 'NA'
            const isIntent = i === INTENT_STEP_INDEX

            const barWidth = state === 'hidden' ? '0%' : '100%'
            const barColorClass =
              state === 'hidden'
                ? 'bg-zinc-700'
                : checking
                ? (isIntent
                  ? 'bg-gradient-to-r from-fuchsia-500 via-purple-400 to-cyan-300 animate-pulse'
                  : 'bg-gradient-to-r from-blue-500 via-blue-400 to-cyan-400')
                : outcome === 'OK'
                ? (isIntent ? 'bg-gradient-to-r from-fuchsia-500 to-cyan-400' : 'bg-safe')
                : outcome === 'FAIL'
                ? 'bg-block'
                : 'bg-zinc-700'
            const barTransition = checking
              ? (isIntent
                ? 'width 1100ms linear, background-color 120ms'
                : 'width 200ms linear, background-color 120ms')
              : 'width 140ms ease-out, background-color 120ms'

            // Pull live AI summary for the Intent step once resolved
            const intentSummary =
              isIntent && resolved && !isSkipped && result?.data.passport.intent_handshake?.safety_summary
                ? result.data.passport.intent_handshake.safety_summary
                : null
            const intentSource = result?.data.passport.intent_handshake?.source

            return (
              <li
                key={label}
                className={`transition-opacity duration-200 ${
                  visible ? 'opacity-100' : 'opacity-0'
                }`}
              >
                <div className="flex items-center justify-between mb-1">
                  <span className="inline-flex items-center gap-1.5">
                    <span
                      className={
                        resolved && isSkipped
                          ? 'text-zinc-700 text-xs'
                          : checking
                          ? (isIntent ? 'text-fuchsia-200 font-semibold text-xs' : 'text-zinc-100 font-medium text-xs')
                          : resolved
                          ? (isIntent ? 'text-fuchsia-300 font-medium text-xs' : 'text-zinc-300 text-xs')
                          : 'text-zinc-500 text-xs'
                      }
                    >
                      {label}
                    </span>
                    {isIntent && visible && (
                      <span
                        className={`text-[8px] font-mono uppercase tracking-wider px-1 py-px rounded ${
                          checking
                            ? 'bg-fuchsia-500/20 text-fuchsia-200 animate-pulse'
                            : resolved && !isSkipped
                            ? 'bg-fuchsia-500/20 text-fuchsia-300'
                            : 'bg-zinc-800 text-zinc-600'
                        }`}
                      >
                        AI
                      </span>
                    )}
                  </span>
                  <span className="ml-2 shrink-0 inline-flex items-center justify-end w-4 h-4">
                    {!visible ? (
                      <span className="text-zinc-700">·</span>
                    ) : checking ? (
                      isIntent ? (
                        <span className="inline-block w-3 h-3 border-2 border-fuchsia-400 border-t-transparent rounded-full animate-spin" />
                      ) : (
                        <span className="inline-block w-3 h-3 border-2 border-blue-400 border-t-transparent rounded-full animate-spin" />
                      )
                    ) : outcome === 'OK' ? (
                      <span className={isIntent ? 'text-fuchsia-300 text-sm leading-none' : 'text-safe text-sm leading-none'}>✓</span>
                    ) : outcome === 'FAIL' ? (
                      <span className="text-block text-sm leading-none">✗</span>
                    ) : (
                      <span className="text-zinc-700">—</span>
                    )}
                  </span>
                </div>
                <div className={`h-0.5 w-full bg-zinc-800/60 rounded-full overflow-hidden ${isIntent && checking ? 'shadow-[0_0_8px_rgba(217,70,239,0.5)]' : ''}`}>
                  <div
                    className={`h-full ${barColorClass} rounded-full`}
                    style={{ width: barWidth, transition: barTransition }}
                  />
                </div>
                {intentSummary && (
                  <div className="mt-1.5 ml-0.5 text-[10px] text-fuchsia-200/90 leading-tight border-l-2 border-fuchsia-500/60 pl-2">
                    <span className="text-fuchsia-400 font-semibold">AI 안전 안내:</span>{' '}
                    {intentSummary}
                    {intentSource && (
                      <span className="ml-1 text-[8px] text-fuchsia-300/60 font-mono uppercase">
                        · {intentSource}
                      </span>
                    )}
                  </div>
                )}
              </li>
            )
          })}
        </ul>
      </div>

      {/* Log feed */}
      <div className="px-5 pb-3 flex-1 min-h-0">
        <div className="text-[10px] text-zinc-500 mb-1 uppercase tracking-wider">log</div>
        <div
          ref={logRef}
          className="bg-bg-base rounded-md p-2 text-[10px] font-mono text-zinc-400 overflow-y-auto max-h-28 border border-zinc-800"
        >
          {log.length === 0 ? (
            <div className="text-zinc-700">— 대기 중 —</div>
          ) : (
            log.map((entry, i) => (
              <div
                key={i}
                className={
                  entry.tone === 'ok'
                    ? 'text-safe/90'
                    : entry.tone === 'fail'
                    ? 'text-block/90'
                    : 'text-zinc-500'
                }
              >
                <span className="text-zinc-700">[{entry.ts}]</span> {entry.text}
              </div>
            ))
          )}
        </div>
      </div>

      {/* Verdict */}
      {phase === 'complete' && verdict && (
        <div className="px-5 pb-3">
          <div
            className={`text-xs font-semibold mb-1 ${
              verdict === 'SAFE' ? 'text-safe' : 'text-block'
            }`}
          >
            verdict: {verdict}
          </div>
          {txHash && (
            <div className="text-[10px] text-zinc-500 font-mono truncate">
              {isMockHash ? 'sim hash · ' : 'sepolia tx · '}
              {txHash.slice(0, 14)}…{txHash.slice(-8)}
            </div>
          )}
        </div>
      )}

      <div className="border-t border-blue-900/30 px-5 py-2.5 bg-gradient-to-r from-blue-950/20 to-transparent">
        <div className="text-center text-[10px] text-zinc-500">
          앱/Broker 레벨 Call Passport · 통신망 코어 연동 불필요
        </div>
      </div>
    </div>
  )
}
