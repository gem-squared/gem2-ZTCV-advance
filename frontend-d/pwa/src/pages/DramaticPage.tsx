import { useCallback, useEffect, useRef, useState } from 'react'
import { CallerPanel, type CallerPhase } from '../panels/CallerPanel'
import { BrokerPanel, type BrokerPhase } from '../panels/BrokerPanel'
import { ReceiverPanel, type ReceiverPhase } from '../panels/ReceiverPanel'
import { SCENARIO_LIST, type ScenarioId, type BadgeTone } from '../scenarios'
import { runScenario, type RunScenarioResult } from '../api'

const TONE_BORDER: Record<BadgeTone, string> = {
  safe: 'border-safe/50',
  block: 'border-block/50',
  warn: 'border-warn/50'
}

const TONE_BADGE: Record<BadgeTone, string> = {
  safe: 'bg-safe/15 text-safe',
  block: 'bg-block/15 text-block',
  warn: 'bg-warn/15 text-warn'
}

const CALLER_PREPARE_MS = 600    // Caller "preparing" phase
const CALLER_SUBMIT_MS = 600     // Caller "submitting" phase before Broker takes over
// 9-item pipeline + Intent step's extra checking duration (1100-200=900ms extra) + hold = ~4.5s
const BROKER_ANIMATION_MS = 9 * 280 + 900 + 700
const RECEIVER_RING_MS = 500     // brief ring before Passport reveal

export function DramaticPage() {
  const [scenarioId, setScenarioId] = useState<ScenarioId | null>(null)
  const [callerPhase, setCallerPhase] = useState<CallerPhase>('idle')
  const [brokerPhase, setBrokerPhase] = useState<BrokerPhase>('idle')
  const [receiverPhase, setReceiverPhase] = useState<ReceiverPhase>('idle')
  const [result, setResult] = useState<RunScenarioResult | null>(null)
  const [offline, setOffline] = useState(false)
  const timeoutsRef = useRef<ReturnType<typeof setTimeout>[]>([])

  const clearTimeouts = () => {
    for (const t of timeoutsRef.current) clearTimeout(t)
    timeoutsRef.current = []
  }

  useEffect(() => () => clearTimeouts(), [])

  const handlePick = useCallback(async (id: ScenarioId) => {
    clearTimeouts()
    setScenarioId(id)
    setResult(null)

    // Stage 1 — Caller preparing
    setCallerPhase('preparing')
    setBrokerPhase('idle')
    setReceiverPhase('idle')

    // Start backend fetch in parallel
    const resultPromise = runScenario(id)

    // Stage 2 — Caller submitting
    timeoutsRef.current.push(
      setTimeout(() => setCallerPhase('submitting'), CALLER_PREPARE_MS)
    )

    // Stage 3 — Caller submitted + Broker receiving
    timeoutsRef.current.push(
      setTimeout(() => {
        setCallerPhase('submitted')
        setBrokerPhase('receiving')
      }, CALLER_PREPARE_MS + CALLER_SUBMIT_MS)
    )

    // Stage 4 — Broker starts verifying (the checklist animation runs)
    const brokerVerifyStart = CALLER_PREPARE_MS + CALLER_SUBMIT_MS + 300
    timeoutsRef.current.push(
      setTimeout(() => setBrokerPhase('verifying'), brokerVerifyStart)
    )

    // Stage 5 — Broker complete + Receiver ringing
    const brokerComplete = brokerVerifyStart + BROKER_ANIMATION_MS
    timeoutsRef.current.push(
      setTimeout(() => {
        setBrokerPhase('complete')
        setReceiverPhase('ringing')
      }, brokerComplete)
    )

    // Stage 6 — Receiver Passport revealed
    timeoutsRef.current.push(
      setTimeout(() => setReceiverPhase('revealed'), brokerComplete + RECEIVER_RING_MS)
    )

    // Await result for the Receiver Passport display
    try {
      const r = await resultPromise
      setResult(r)
      setOffline(r.source === 'fixture')
    } catch {
      setOffline(true)
    }
  }, [])

  return (
    <div className="min-h-screen bg-bg-base text-zinc-100">
      <header className="px-6 py-5 border-b border-bg-elevated relative">
        {offline && (
          <div className="absolute top-3 right-6 text-[10px] font-medium text-amber-300 bg-amber-950/60 px-2.5 py-1 rounded-full border border-amber-900/80">
            ⚠ 오프라인 모드
          </div>
        )}
        <div className="flex items-center gap-3 mb-2">
          <div className="text-2xl">🛡️</div>
          <h1 className="text-xl font-semibold">VOUCH — Verified Outgoing-call Universal Compliance Hub</h1>
          <span className="ml-2 text-[10px] text-zinc-500 uppercase tracking-wider">
            Dramatic Mode
          </span>
        </div>
        <p className="text-xs text-zinc-400">
          Caller → Verification Broker → Receiver · App/Broker-level Call Passport · 통신망 코어 연동 불필요
        </p>
      </header>

      {/* Scenario picker */}
      <div className="px-6 py-4 border-b border-bg-elevated">
        <div className="text-[10px] text-zinc-500 uppercase tracking-wider mb-2">시나리오</div>
        <div className="flex flex-wrap gap-2">
          {SCENARIO_LIST.map(s => (
            <button
              key={s.id}
              onClick={() => handlePick(s.id)}
              className={`bg-bg-card hover:bg-bg-elevated active:bg-bg-elevated transition-colors rounded-lg px-4 py-2 border-l-2 ${TONE_BORDER[s.badgeTone]} text-left flex items-center gap-3`}
            >
              <span className="text-xs text-zinc-500">#{s.id}</span>
              <span className="text-sm text-zinc-200">{s.title}</span>
              <span className="text-[10px] text-zinc-500">{s.subtitle}</span>
              <span className={`text-[10px] px-2 py-0.5 rounded-full ${TONE_BADGE[s.badgeTone]}`}>
                {s.badge}
              </span>
            </button>
          ))}
        </div>
      </div>

      {/* 3-panel layout — distinct visual frames per Actor:
          Caller  = server-rack frame (rectangular, SDK feel)
          Broker  = system dashboard frame (terminal/admin feel)
          Receiver = phone bezel (citizen device) */}
      <div className="grid grid-cols-1 lg:grid-cols-[1fr_1.3fr_0.85fr] gap-6 p-6 items-start">
        <CallerPanel scenarioId={scenarioId} phase={callerPhase} />
        <BrokerPanel scenarioId={scenarioId} phase={brokerPhase} result={result} />
        <ReceiverPanel scenarioId={scenarioId} phase={receiverPhase} result={result} />
      </div>

      <div className="px-6 py-4 text-center text-[10px] text-zinc-600">
        Dramatic mode · 같은 backend 상태를 공유하는 3-role 시각화 · OmniOne CX + Open DID 기반 발신 검증
      </div>
    </div>
  )
}
