import { useEffect } from 'react'
import { SCENARIOS, type ScenarioId } from '../scenarios'
import { runScenario, type RunScenarioResult } from '../api'

interface Props {
  scenarioId: ScenarioId
  onComplete: (scenarioId: ScenarioId, result: RunScenarioResult) => void
}

const RING_MIN_DURATION_MS = 1800

export function RingScreen({ scenarioId, onComplete }: Props) {
  const scenario = SCENARIOS[scenarioId]

  useEffect(() => {
    let cancelled = false
    const startedAt = Date.now()
    runScenario(scenarioId).then((result) => {
      if (cancelled) return
      const elapsed = Date.now() - startedAt
      const remain = Math.max(0, RING_MIN_DURATION_MS - elapsed)
      setTimeout(() => {
        if (!cancelled) onComplete(scenarioId, result)
      }, remain)
    })
    return () => {
      cancelled = true
    }
  }, [scenarioId, onComplete])

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#1a1145] to-[#0a0a1f] text-zinc-100 flex flex-col px-6 py-10">
      <div className="text-center mt-4 mb-6">
        <div className="inline-block bg-accent/40 text-zinc-100 text-xs px-4 py-1 rounded-full">
          수신 전화
        </div>
      </div>
      <div className="flex-1 flex flex-col items-center justify-center">
        <div className="relative w-28 h-28 mb-6">
          <div className="absolute inset-0 rounded-full bg-blue-500/30 animate-ping"></div>
          <div className="relative w-28 h-28 rounded-full bg-blue-500 flex items-center justify-center text-4xl">
            📞
          </div>
        </div>
        <h2 className="text-2xl font-semibold mb-2 text-center">{scenario.callerLabel}</h2>
        <div className="text-sm text-zinc-300 mb-1">{scenario.callerOrg}</div>
        <div className="text-sm text-zinc-400 mb-4">{scenario.callerPhone}</div>
        <div className="inline-flex items-center gap-2 bg-zinc-800/60 px-3 py-1 rounded text-xs text-zinc-300 mb-12">
          <span>{scenario.callerBadge === 'AI' ? '🤖' : scenario.callerBadge === '수사관' ? '🏛️' : '💼'}</span>
          <span>{scenario.callerBadge}</span>
        </div>
        <div className="inline-flex items-center gap-2 bg-emerald-500/15 text-emerald-300 text-xs px-4 py-2 rounded-full">
          <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
          VOUCH 인증 중
        </div>
      </div>
      <div className="flex justify-around pb-2 opacity-50 pointer-events-none">
        <div className="flex flex-col items-center gap-1">
          <div className="w-14 h-14 rounded-full bg-red-500 flex items-center justify-center text-2xl">📞</div>
          <span className="text-xs text-zinc-500">거절</span>
        </div>
        <div className="flex flex-col items-center gap-1">
          <div className="w-14 h-14 rounded-full bg-green-500 flex items-center justify-center text-2xl">📞</div>
          <span className="text-xs text-zinc-500">수락</span>
        </div>
      </div>
    </div>
  )
}
