// /flow/ route — cinematic Protocol Flow Visualization.
//
// One PWA, three "device" panels separated by whitespace, animated
// SVG flow lines + labels connecting them. Ledger element below
// Broker. Orchestrated through 7 phases so judges can watch the
// protocol move through the verification pipeline in real time.
//
// Reuses the existing CallerPanel / BrokerPanel / ReceiverPanel
// components and the runScenario() API — no backend coupling change.

import { useCallback, useEffect, useRef, useState } from 'react'
import { CallerPanel, type CallerPhase } from '../panels/CallerPanel'
import { BrokerPanel, type BrokerPhase } from '../panels/BrokerPanel'
import { ReceiverPanel, type ReceiverPhase } from '../panels/ReceiverPanel'
import { SCENARIO_LIST, type ScenarioId, type BadgeTone } from '../scenarios'
import { runScenario, type RunScenarioResult } from '../api'
import { SvgOverlay } from '../flow/SvgOverlay'
import { FlowLine } from '../flow/FlowLine'
import { FlowLabel } from '../flow/FlowLabel'
import { LedgerElement } from '../flow/LedgerElement'
import { AuditDrawer } from '../flow/AuditDrawer'
import { BlockchainAnchor } from '../flow/anchor/BlockchainAnchor'
import { PreCallVerification } from '../flow/anchor/PreCallVerification'
import { useAnchors, type AnchorSpec } from '../flow/anchors'
import { PHASE_TIMING_MS } from '../flow/timing'

const TONE_BORDER: Record<BadgeTone, string> = {
  safe: 'border-safe/50',
  block: 'border-block/50',
  warn: 'border-warn/50',
}
const TONE_BADGE: Record<BadgeTone, string> = {
  safe: 'bg-safe/15 text-safe',
  block: 'bg-block/15 text-block',
  warn: 'bg-warn/15 text-warn',
}

// Active-line flags driven by phase progression.
interface FlowState {
  makeCall: boolean
  callerProof: boolean
  receiverDID: boolean
  verificationMirror: boolean
  chainAnchor: boolean
  ledgerAnchored: boolean
}

const FLOW_INIT: FlowState = {
  makeCall: false,
  callerProof: false,
  receiverDID: false,
  verificationMirror: false,
  chainAnchor: false,
  ledgerAnchored: false,
}

export function FlowPage() {
  const [scenarioId, setScenarioId] = useState<ScenarioId | null>(null)
  const [callerPhase, setCallerPhase] = useState<CallerPhase>('idle')
  const [brokerPhase, setBrokerPhase] = useState<BrokerPhase>('idle')
  const [receiverPhase, setReceiverPhase] = useState<ReceiverPhase>('idle')
  const [result, setResult] = useState<RunScenarioResult | null>(null)
  const [flow, setFlow] = useState<FlowState>(FLOW_INIT)
  const [anchorVersion, setAnchorVersion] = useState(0)
  const [auditOpen, setAuditOpen] = useState(false)

  const callerRef   = useRef<HTMLDivElement | null>(null)
  const brokerRef   = useRef<HTMLDivElement | null>(null)
  const receiverRef = useRef<HTMLDivElement | null>(null)
  const ledgerRef   = useRef<HTMLDivElement | null>(null)

  const timeoutsRef = useRef<ReturnType<typeof setTimeout>[]>([])

  const clearTimeouts = () => {
    for (const t of timeoutsRef.current) clearTimeout(t)
    timeoutsRef.current = []
  }
  useEffect(() => () => clearTimeouts(), [])

  // Force re-anchor whenever the visual layout shifts due to phase
  // progression. The Receiver panel grows when its 'revealed' phase
  // fires (Passport card appears), which pushes the page height and
  // shifts the Ledger row downward. Without this, the chain-anchor
  // line keeps pointing at the OLD ledger position (mid-page).
  useEffect(() => {
    if (flow.chainAnchor || receiverPhase === 'revealed') {
      setAnchorVersion((v) => v + 1)
    }
  }, [flow.chainAnchor, receiverPhase])

  // Anchor spec — all the connection endpoints we care about. Anchor
  // computation re-runs on resize + scroll automatically, plus when
  // we bump anchorVersion (used after scenario picks).
  const anchorSpecs: AnchorSpec[] = [
    { name: 'caller.right',   ref: callerRef,   side: 'right' },
    { name: 'broker.left',    ref: brokerRef,   side: 'left' },
    { name: 'broker.right',   ref: brokerRef,   side: 'right' },
    { name: 'broker.bottom',  ref: brokerRef,   side: 'bottom' },
    { name: 'receiver.left',  ref: receiverRef, side: 'left' },
    { name: 'ledger.top',     ref: ledgerRef,   side: 'top' },
  ]
  const anchors = useAnchors(anchorSpecs, anchorVersion)

  const verdict = result?.data.passport.outcome
  const isFailed = verdict === 'FAILED'
  // Mirror line color reflects outcome — green for SAFE, red for FAILED.
  const mirrorColor = isFailed ? 'block' as const : 'safe' as const

  const handlePick = useCallback(async (id: ScenarioId) => {
    clearTimeouts()
    setScenarioId(id)
    setResult(null)
    setFlow(FLOW_INIT)
    setCallerPhase('idle')
    setBrokerPhase('idle')
    setReceiverPhase('idle')
    // Bump anchor version so coordinates re-snap after the panels
    // possibly re-render with scenario-specific content.
    setAnchorVersion((v) => v + 1)

    // Fire backend fetch in parallel with the animation.
    const resultPromise = runScenario(id)

    // Phase orchestration timeline (see PHASE_TIMING_MS).
    const at = (ms: number, fn: () => void) =>
      timeoutsRef.current.push(setTimeout(fn, ms))

    at(PHASE_TIMING_MS.callerPreparing, () => setCallerPhase('preparing'))
    at(PHASE_TIMING_MS.makeCall, () => setFlow((f) => ({ ...f, makeCall: true })))
    at(PHASE_TIMING_MS.receiverRinging, () => setReceiverPhase('ringing'))
    at(PHASE_TIMING_MS.callerSubmitting, () => {
      setCallerPhase('submitting')
      setFlow((f) => ({ ...f, callerProof: true }))
    })
    at(PHASE_TIMING_MS.receiverDID, () =>
      setFlow((f) => ({ ...f, receiverDID: true }))
    )
    at(PHASE_TIMING_MS.callerSubmitted, () => {
      setCallerPhase('submitted')
      setBrokerPhase('receiving')
    })
    at(PHASE_TIMING_MS.brokerVerifying, () => setBrokerPhase('verifying'))
    at(PHASE_TIMING_MS.verificationMirror, () =>
      setFlow((f) => ({ ...f, verificationMirror: true }))
    )
    at(PHASE_TIMING_MS.chainAnchor, () =>
      setFlow((f) => ({ ...f, chainAnchor: true, ledgerAnchored: true }))
    )
    at(PHASE_TIMING_MS.brokerCompleteReceiverRevealed, () => {
      setBrokerPhase('complete')
      setReceiverPhase('revealed')
    })

    // Await the backend so the Receiver panel + Ledger element have
    // the final passport data when phase 7 fires.
    try {
      const r = await resultPromise
      setResult(r)
    } catch {
      // runScenario never rejects (silent fixture fallback) — but be
      // defensive.
    }
  }, [])

  const replay = () => {
    if (scenarioId) void handlePick(scenarioId)
  }

  const offlineMode = result?.source === 'fixture'
  const intentSource = result?.data.passport.intent_handshake?.source

  return (
    <div className="min-h-screen w-full bg-black text-zinc-100 relative overflow-x-hidden">
      {/* Header strip — brand + status pills + Replay only (scenario picker moved
          to left column below Caller per WP-07.1) */}
      <header className="px-6 py-4 border-b border-zinc-900/80 relative z-20 bg-black/80 backdrop-blur-sm">
        <div className="flex items-center gap-4 flex-wrap">
          <div className="flex items-center gap-3">
            <div className="text-2xl">🛡️</div>
            <div>
              <h1 className="text-lg font-semibold tracking-tight">VOUCH</h1>
              <div className="text-[10px] text-zinc-500 uppercase tracking-wider">
                Protocol Flow Visualization
              </div>
            </div>
          </div>
          <div className="flex items-center gap-2 ml-3">
            {offlineMode && (
              <span className="px-2 py-0.5 rounded-full bg-amber-950/60 border border-amber-900/80 text-amber-300 text-[10px]">
                ⚠ 오프라인 모드
              </span>
            )}
            {intentSource && (
              <span className={`px-2 py-0.5 rounded-full text-[10px] ${intentSource === 'live' ? 'bg-fuchsia-950/40 text-fuchsia-300 border border-fuchsia-900/60' : 'bg-zinc-900 text-zinc-500 border border-zinc-800'}`}>
                {intentSource === 'live' ? '● LIVE LLM' : '○ FALLBACK'}
              </span>
            )}
          </div>
          <div className="flex-1 min-w-0" />
          {/* Audit button — sized 130% of base; when a fresh result lands
              the button pulses (scale + amber glow), expanding amber rings
              radiate from its bounds, and a "● 데이터 준비됨" badge pops in,
              so judges immediately see that real audit data is ready.
              Re-keyed on sessionId so it re-fires on every new scenario play. */}
          <div
            className="relative"
            key={`audit-cta-${result?.data.passport.sessionId ?? 'idle'}`}
          >
            {result && !auditOpen && (
              <>
                <span
                  className="absolute inset-0 rounded-lg border-2 border-amber-400/70 pointer-events-none"
                  style={{ animation: 'audit-ready-ring 1500ms ease-out 1 both' }}
                />
                <span
                  className="absolute inset-0 rounded-lg border-2 border-amber-400/50 pointer-events-none"
                  style={{ animation: 'audit-ready-ring 1500ms ease-out 600ms 1 both' }}
                />
              </>
            )}
            <button
              onClick={() => setAuditOpen(true)}
              disabled={!result}
              className="relative px-4 py-2.5 text-[15px] bg-amber-950/40 hover:bg-amber-900/40 border border-amber-700/40 text-amber-200 rounded-lg disabled:opacity-30 disabled:cursor-not-allowed transition-colors inline-flex items-center gap-2 font-medium"
              title="원시 감사 기록 열람 — 백엔드 ↔ on-chain 정합성 검증"
              style={result && !auditOpen
                ? { animation: 'audit-ready-pulse 2200ms ease-in-out 1 both' }
                : undefined}
            >
              🔍 감사 추적
              {result && !auditOpen && (
                <span
                  className="ml-1 text-[10px] px-1.5 py-0.5 rounded-full bg-emerald-500/30 border border-emerald-400/50 text-emerald-200 font-semibold"
                  style={{ animation: 'data-ready-pop 320ms ease-out 200ms 1 both', opacity: 0 }}
                >
                  ● 데이터 준비됨
                </span>
              )}
            </button>
          </div>
          <button
            onClick={replay}
            disabled={!scenarioId}
            className="px-3 py-1.5 text-xs bg-bg-card hover:bg-bg-elevated rounded-lg disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
          >
            ↻ 재생
          </button>
        </div>
      </header>

      {/* Single 3-column grid where each column owns its own vertical flow.
          Column 1 (Caller): panel → small gap → scenario picker
          Column 2 (Broker): panel → small gap → Ledger
          Column 3 (Receiver): panel only
          This eliminates the dead whitespace caused by a separate bottom row
          inheriting the tallest panel's height. */}
      <div className="px-6 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-[1fr_1.3fr_0.85fr] gap-[140px] items-start">
          {/* — Column 1: Caller + Scenario picker — */}
          <div className="flex flex-col gap-6">
            <div ref={callerRef}>
              <CallerPanel scenarioId={scenarioId} phase={callerPhase} />
            </div>
            <div className="flex flex-col gap-3">
              <div className="text-[10px] text-zinc-500 uppercase tracking-wider mb-1 pl-1">
                시나리오 선택
              </div>
              {SCENARIO_LIST.map((s) => (
                <button
                  key={s.id}
                  onClick={() => handlePick(s.id)}
                  className={`bg-bg-card hover:bg-bg-elevated active:bg-bg-elevated transition-colors rounded-lg px-4 py-3 border-l-4 ${TONE_BORDER[s.badgeTone]} text-left flex items-center gap-3 w-full ${scenarioId === s.id ? 'ring-2 ring-zinc-700 shadow-[0_0_16px_rgba(255,255,255,0.06)]' : ''}`}
                >
                  <span className="text-sm text-zinc-500 font-mono shrink-0">#{s.id}</span>
                  <span className="text-sm text-zinc-100 flex-1 min-w-0">{s.title}</span>
                  <span className={`text-[10px] px-2 py-0.5 rounded-full shrink-0 ${TONE_BADGE[s.badgeTone]}`}>
                    {s.badge}
                  </span>
                </button>
              ))}
            </div>
          </div>

          {/* — Column 2: Broker + Ledger + BlockchainAnchor status card — */}
          <div className="flex flex-col gap-10 items-center">
            <div ref={brokerRef} className="w-full">
              <BrokerPanel scenarioId={scenarioId} phase={brokerPhase} result={result} />
            </div>
            <LedgerElement
              ref={ledgerRef}
              receiptTxHash={result?.data.passport.receiptTxHash}
              explorerUrl={result?.data.passport.explorerUrl}
              anchored={flow.ledgerAnchored}
            />
            <BlockchainAnchor
              active={flow.chainAnchor && !!result?.data.passport.receiptTxHash}
              txHash={result?.data.passport.receiptTxHash}
            />
          </div>

          {/* — Column 3: Receiver + PreCallVerification (broker-verifying micro-storytelling) — */}
          <div className="flex flex-col gap-4">
            <div ref={receiverRef}>
              <ReceiverPanel scenarioId={scenarioId} phase={receiverPhase} result={result} />
            </div>
            <PreCallVerification
              active={brokerPhase === 'verifying' || brokerPhase === 'receiving'}
            />
          </div>
        </div>

        <div className="text-center text-[10px] text-zinc-600 mt-12 leading-relaxed">
          통화 연결 전 발신 신원·권한·목적 검증 · OmniOne CX + Open DID 기반 ·
          영수증은 Sepolia 테스트넷 앵커링 · Intent Handshake (Predictive
          Disclosure) by Anthropic Claude Haiku 4.5
        </div>
      </div>

      {/* SVG flow overlay — sits above panels visually but
          pointer-events: none so clicks pass through. */}
      <SvgOverlay>
        {/* Phase 1: make-call (Caller → Receiver, white dashed) */}
        <FlowLine
          from={anchors['caller.right']}
          to={anchors['receiver.left']}
          color="white"
          style="dashed"
          active={flow.makeCall}
          curve={0.4}
        />
        <FlowLabel
          from={anchors['caller.right']}
          to={anchors['receiver.left']}
          text="통화 발신"
          step={1}
          color="white"
          active={flow.makeCall}
          offsetY={-28}
        />

        {/* Phase 2: caller-proof (Caller → Broker, blue, moving dot) */}
        <FlowLine
          from={anchors['caller.right']}
          to={anchors['broker.left']}
          color="blue"
          active={flow.callerProof}
          movingDot
          curve={0.6}
        />
        <FlowLabel
          from={anchors['caller.right']}
          to={anchors['broker.left']}
          text="CallerProof · DID · 권한 · 목적"
          step={2}
          color="blue"
          active={flow.callerProof}
          offsetY={-18}
        />

        {/* Phase 3: receiver-did (Receiver → Broker, blue, moving dot) */}
        <FlowLine
          from={anchors['receiver.left']}
          to={anchors['broker.right']}
          color="blue"
          active={flow.receiverDID}
          movingDot
          curve={0.6}
        />
        <FlowLabel
          from={anchors['receiver.left']}
          to={anchors['broker.right']}
          text="수신자 모바일 신분증 제시"
          step={3}
          color="blue"
          active={flow.receiverDID}
          offsetY={-18}
        />

        {/* Phase 5: verification-mirror (Broker → Receiver, color by outcome) */}
        <FlowLine
          from={anchors['broker.right']}
          to={anchors['receiver.left']}
          color={mirrorColor}
          active={flow.verificationMirror}
          movingDot
          curve={0.4}
          width={2.5}
        />
        <FlowLabel
          from={anchors['broker.right']}
          to={anchors['receiver.left']}
          text={isFailed ? '⑤ 검증 결과: BLOCK 동기화' : '⑤ 검증 결과: SAFE 동기화'}
          color={mirrorColor}
          active={flow.verificationMirror}
          offsetY={18}
        />

        {/* Phase 6: chain-anchor (Broker → Ledger, gold, moving dot drop).
            Use curve=0 so the line is dead-straight vertical — the
            broker.bottom and ledger.top anchors are both centered in
            column 2, so any curvature only introduces visual drift. */}
        <FlowLine
          from={anchors['broker.bottom']}
          to={anchors['ledger.top']}
          color="gold"
          active={flow.chainAnchor}
          movingDot
          curve={0}
          width={2.5}
        />
        <FlowLabel
          from={anchors['broker.bottom']}
          to={anchors['ledger.top']}
          text="Chain Receipt 앵커링"
          step={6}
          color="gold"
          active={flow.chainAnchor}
          offsetX={110}
          offsetY={0}
        />
      </SvgOverlay>

      {/* Audit Trail drawer — slide-out, exposes the raw audit data
          (CallerProof, Predictive Disclosure manifest with client-side
          hash verification, Risk Verdict, Receiver claims, Chain Receipt
          off↔on-chain side-by-side + Etherscan link). Materializes
          v3-KR §Ⅳ.2 "Audit Defensibility" as a clickable proof page. */}
      <AuditDrawer
        open={auditOpen}
        result={result}
        scenarioId={scenarioId}
        onClose={() => setAuditOpen(false)}
      />
    </div>
  )
}
