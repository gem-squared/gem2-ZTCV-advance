import { useEffect, useState } from 'react'
import type { RunScenarioResult } from '../api'
import type { ScenarioId } from '../scenarios'

interface Props {
  scenarioId: ScenarioId
  result: RunScenarioResult
  onComplete: () => void
}

const CHECKLIST = [
  '발신 번호 조회',
  'Caller DID binding 확인',
  '기관 등록 상태 확인',
  'AI Agent VC 검증',
  '통화 목적 권한 확인',
  '신고/위험 패턴 조회',
  'Chain Receipt 준비'
]

type OutcomeMark = 'OK' | 'FAIL' | 'NA'
type ItemState = 'hidden' | 'checking' | 'resolved'

const PER_SCENARIO_OUTCOMES: Record<ScenarioId, OutcomeMark[]> = {
  1: ['OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK'],
  2: ['OK', 'FAIL', 'NA', 'NA', 'NA', 'NA', 'OK'],
  3: ['OK', 'OK', 'OK', 'OK', 'FAIL', 'OK', 'OK']
}

const ITEM_INTERVAL_MS = 280        // gap between items starting
const CHECKING_DURATION_MS = 200    // spinner + bar-fill duration for OK/FAIL items
const HOLD_MS = 700                 // pause after last item resolves

export function BrokerVerificationScreen({ scenarioId, onComplete }: Props) {
  const outcomes = PER_SCENARIO_OUTCOMES[scenarioId]
  const [itemStates, setItemStates] = useState<ItemState[]>(
    () => outcomes.map(() => 'hidden')
  )

  useEffect(() => {
    let cancelled = false
    const timeouts: ReturnType<typeof setTimeout>[] = []

    const setItemAt = (i: number, s: ItemState) =>
      setItemStates(prev => {
        if (cancelled || prev[i] === s) return prev
        const next = [...prev]
        next[i] = s
        return next
      })

    for (let i = 0; i < outcomes.length; i++) {
      const startTime = i * ITEM_INTERVAL_MS
      const isNA = outcomes[i] === 'NA'

      if (isNA) {
        // Broker short-circuit: NA items resolve immediately at startTime
        const t = setTimeout(() => {
          if (!cancelled) setItemAt(i, 'resolved')
        }, startTime)
        timeouts.push(t)
      } else {
        const t1 = setTimeout(() => {
          if (!cancelled) setItemAt(i, 'checking')
        }, startTime)
        timeouts.push(t1)

        const t2 = setTimeout(() => {
          if (!cancelled) setItemAt(i, 'resolved')
        }, startTime + CHECKING_DURATION_MS)
        timeouts.push(t2)
      }
    }

    const total = outcomes.length * ITEM_INTERVAL_MS + HOLD_MS
    const completeT = setTimeout(() => {
      if (!cancelled) onComplete()
    }, total)
    timeouts.push(completeT)

    return () => {
      cancelled = true
      for (const t of timeouts) clearTimeout(t)
    }
  }, [scenarioId, onComplete, outcomes])

  return (
    <div className="min-h-screen bg-bg-base text-zinc-100 px-6 py-10 flex flex-col items-center">
      <div className="text-center mt-4 mb-6">
        <div className="text-3xl mb-3">🛡️</div>
        <h1 className="text-xl font-semibold mb-2">VOUCH Broker 검증 중</h1>
        <div className="inline-flex items-center gap-2 text-xs text-zinc-400">
          <span className="w-1.5 h-1.5 rounded-full bg-blue-400 animate-pulse"></span>
          <span>pre-call verification</span>
        </div>
      </div>

      <div className="bg-bg-card rounded-xl p-5 mb-6 max-w-md w-full">
        <ul className="space-y-3 text-sm">
          {CHECKLIST.map((label, i) => {
            const state = itemStates[i]
            const outcome = outcomes[i]
            const visible = state !== 'hidden'
            const checking = state === 'checking'
            const resolved = state === 'resolved'
            const isSkipped = outcome === 'NA'

            // Per-item progress bar config
            const barWidth = state === 'hidden' ? '0%' : '100%'
            const barColorClass =
              state === 'hidden'
                ? 'bg-zinc-700'
                : checking
                ? 'bg-gradient-to-r from-blue-500 via-blue-400 to-cyan-400'
                : outcome === 'OK'
                ? 'bg-safe'
                : outcome === 'FAIL'
                ? 'bg-block'
                : 'bg-zinc-700' // NA resolved
            const barTransition = checking
              ? 'width 200ms linear, background-color 120ms'
              : 'width 140ms ease-out, background-color 120ms'

            return (
              <li
                key={label}
                className={`transition-all duration-200 ${
                  visible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-1'
                }`}
              >
                <div className="flex items-center justify-between mb-1.5">
                  <span
                    className={
                      resolved && isSkipped
                        ? 'text-zinc-700'
                        : checking
                        ? 'text-zinc-100 font-medium'
                        : resolved
                        ? 'text-zinc-300'
                        : 'text-zinc-500'
                    }
                  >
                    {label}
                  </span>
                  <span className="ml-3 shrink-0 inline-flex items-center justify-end w-5 h-5">
                    {!visible ? (
                      <span className="text-zinc-700">·</span>
                    ) : checking ? (
                      <span
                        className="inline-block w-3.5 h-3.5 border-2 border-blue-400 border-t-transparent rounded-full animate-spin"
                        aria-label="checking"
                      />
                    ) : outcome === 'OK' ? (
                      <span className="text-safe text-base leading-none">✓</span>
                    ) : outcome === 'FAIL' ? (
                      <span className="text-block text-base leading-none">✗</span>
                    ) : (
                      <span className="text-zinc-700">—</span>
                    )}
                  </span>
                </div>
                {/* Per-item progress bar */}
                <div className="h-0.5 w-full bg-zinc-800/60 rounded-full overflow-hidden">
                  <div
                    className={`h-full ${barColorClass} rounded-full`}
                    style={{ width: barWidth, transition: barTransition }}
                  />
                </div>
              </li>
            )
          })}
        </ul>
      </div>

      <div className="flex-1" />

      <div className="text-center text-[11px] text-zinc-500 leading-relaxed max-w-md pb-2">
        <div>OmniOne CX + Open DID 기반 발신 검증</div>
        <div className="mt-0.5">앱/Broker 레벨 Call Passport · 통신망 코어 연동 불필요</div>
      </div>
    </div>
  )
}
