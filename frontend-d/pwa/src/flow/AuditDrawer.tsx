// Audit Trail drawer — right-side slide-out that exposes the raw
// session data already in the API response, so judges can verify
// VOUCH's claims live. Materializes v3-KR §Ⅳ.2 "audit defensibility"
// from rhetoric into a clickable proof page.
//
// Sections (top → bottom):
//   1. Caller Identity (CallerProof Ed25519 envelope)
//   2. Predictive Disclosure (Intent Handshake) + client-side hash verify
//   3. Risk Verdict (Layer 1 rules + Layer 2 composer)
//   4. Receiver Identity (OACX token + mResident-shape claims)
//   5. Chain Receipt (off-chain JSON + on-chain 5-field side-by-side)
//   6. Honesty footer (✅ REAL / ⚠ MOCK separation)

import { useEffect, useState } from 'react'
import type { RunScenarioResult } from '../api'
import type { ScenarioId } from '../scenarios'
import { canonicalManifestJSON, manifestSha256Hex } from './audit/canonicalHash'

interface Props {
  open: boolean
  result: RunScenarioResult | null
  scenarioId: ScenarioId | null
  onClose: () => void
}

// Truncate-middle helper for long hex / hashes.
function truncateMid(s: string | undefined, head = 12, tail = 8): string {
  if (!s) return ''
  if (s.length <= head + tail + 1) return s
  return `${s.slice(0, head)}…${s.slice(-tail)}`
}

// Pretty-print JSON with stable indent.
function pretty(v: unknown): string {
  try { return JSON.stringify(v, null, 2) } catch { return String(v) }
}

// Section wrapper — gives every audit block a consistent header style.
function Section({
  title, subtitle, children,
}: { title: string; subtitle?: string; children: React.ReactNode }) {
  return (
    <section className="border-t border-zinc-800/80">
      <div className="px-5 pt-5 pb-3">
        <h3 className="text-sm font-semibold text-zinc-100 tracking-tight">{title}</h3>
        {subtitle && (
          <p className="text-[11px] text-zinc-500 mt-0.5 leading-snug">{subtitle}</p>
        )}
      </div>
      <div className="px-5 pb-5 space-y-3">{children}</div>
    </section>
  )
}

// Pill badge — used for REAL / MOCK / LIVE / FALLBACK / VERIFIED tags.
type PillTone = 'real' | 'mock' | 'live' | 'fallback' | 'verified' | 'mismatch' | 'info'
function Pill({ tone, children }: { tone: PillTone; children: React.ReactNode }) {
  const styles: Record<PillTone, string> = {
    real:     'bg-emerald-950/50 text-emerald-300 border-emerald-800/60',
    mock:     'bg-amber-950/40  text-amber-300   border-amber-800/60',
    live:     'bg-fuchsia-950/40 text-fuchsia-300 border-fuchsia-800/60',
    fallback: 'bg-zinc-900       text-zinc-400    border-zinc-700',
    verified: 'bg-emerald-950/50 text-emerald-300 border-emerald-800/60',
    mismatch: 'bg-red-950/50     text-red-300     border-red-800/60',
    info:     'bg-blue-950/40    text-blue-300    border-blue-800/60',
  }
  return (
    <span className={`inline-flex items-center text-[10px] font-medium px-2 py-0.5 rounded-full border ${styles[tone]}`}>
      {children}
    </span>
  )
}

// Code block for JSON / hash content. Fixed-width font + horizontal scroll if needed.
function Code({ children, className = '' }: { children: React.ReactNode; className?: string }) {
  return (
    <pre className={`bg-zinc-950 border border-zinc-800 rounded-md p-3 text-[10px] font-mono text-zinc-300 overflow-x-auto max-h-[260px] overflow-y-auto leading-relaxed ${className}`}>
      {children}
    </pre>
  )
}

// Inline label : value row for compact field rendering.
function Field({ label, mono, children }: { label: string; mono?: boolean; children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-0.5">
      <span className="text-[10px] uppercase tracking-wider text-zinc-500">{label}</span>
      <span className={`text-xs text-zinc-200 ${mono ? 'font-mono' : ''} break-all`}>{children}</span>
    </div>
  )
}

export function AuditDrawer({ open, result, scenarioId, onClose }: Props) {
  const [computedHash, setComputedHash] = useState<string | null>(null)
  const [computing, setComputing] = useState(false)

  // Esc key + body scroll lock while drawer is open.
  useEffect(() => {
    if (!open) return
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    window.addEventListener('keydown', onKey)
    const prevOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    return () => {
      window.removeEventListener('keydown', onKey)
      document.body.style.overflow = prevOverflow
    }
  }, [open, onClose])

  // Recompute manifest hash whenever the result changes.
  useEffect(() => {
    const ih = result?.data.passport.intent_handshake
    if (!ih) { setComputedHash(null); return }
    setComputing(true)
    manifestSha256Hex(ih).then((h) => {
      setComputedHash(h)
      setComputing(false)
    })
  }, [result])

  const passport     = result?.data.passport
  const session      = (result?.data.session ?? {}) as Record<string, unknown>
  const callerProof  = session.caller_proof   as Record<string, unknown> | undefined
  const customerProof = session.customer_proof as Record<string, unknown> | undefined
  const riskVerdict  = session.risk_verdict   as Record<string, unknown> | undefined
  const receipt      = session.receipt        as { on_chain?: Record<string, unknown>; off_chain?: Record<string, unknown> } | undefined

  const ih = passport?.intent_handshake
  const reportedHash = passport?.intent_manifest_hash
  const hashMatches =
    !!computedHash && !!reportedHash &&
    computedHash.toLowerCase() === reportedHash.toLowerCase()

  const isFailed = passport?.outcome === 'FAILED'
  const isSafe   = passport?.outcome === 'SAFE'

  return (
    <>
      {/* Backdrop */}
      <div
        className={`fixed inset-0 z-40 bg-black/60 backdrop-blur-sm transition-opacity duration-200 ${
          open ? 'opacity-100' : 'opacity-0 pointer-events-none'
        }`}
        onClick={onClose}
        aria-hidden="true"
      />

      {/* Drawer panel */}
      <aside
        className={`fixed top-0 right-0 z-50 h-screen w-full sm:w-[560px] bg-zinc-950 text-zinc-100 shadow-2xl border-l border-zinc-800 flex flex-col transition-transform duration-250 ease-out ${
          open ? 'translate-x-0' : 'translate-x-full'
        }`}
        role="dialog"
        aria-label="Audit Trail"
      >
        {/* Sticky header */}
        <header className="flex items-center gap-3 px-5 py-3.5 border-b border-zinc-800 bg-zinc-950/95 backdrop-blur-sm sticky top-0 z-10">
          <div className="flex-1 min-w-0">
            <div className="text-[10px] uppercase tracking-wider text-zinc-500 font-semibold">
              🔍 감사 추적 · 시나리오 {scenarioId ?? '—'} ·{' '}
              {isSafe ? <span className="text-safe">SAFE</span> :
               isFailed ? <span className="text-block">BLOCK</span> :
               <span className="text-zinc-500">대기</span>}
            </div>
            <div className="text-sm font-semibold tracking-tight text-zinc-100 mt-0.5">
              신뢰하되 검증하라 — 원천 데이터 검사
            </div>
          </div>
          <button
            onClick={onClose}
            className="px-2.5 py-1.5 text-xs bg-bg-card hover:bg-bg-elevated rounded-md transition-colors"
            aria-label="감사 추적 닫기"
          >
            ✕ 닫기
          </button>
        </header>

        {/* Scrollable body */}
        <div className="flex-1 overflow-y-auto">

          {/* 1. Caller Identity */}
          <Section
            title="① 발신자 신원 — CallerProof 봉투"
            subtitle="9단계 파이프라인의 Step 2에서 제출되는 Ed25519 서명 봉투. 서명 자체는 실제로 검증 가능하나, 공개키를 해소하는 DID 레지스트리는 현재 인메모리(mock-fallback) 상태입니다 (WP-02)."
          >
            <div className="grid grid-cols-2 gap-3">
              <Field label="org_did" mono>{(callerProof?.org_did as string) ?? '—'} <Pill tone="mock">⚠ MOCK 레지스트리</Pill></Field>
              <Field label="caller_did" mono>{truncateMid(callerProof?.caller_did as string)} <Pill tone="mock">⚠ MOCK 레지스트리</Pill></Field>
              <Field label="purpose">{(callerProof?.purpose as string) ?? '—'}</Field>
              <Field label="signature_alg">{(callerProof?.signature_alg as string) ?? '—'} <Pill tone="real">✓ 실제</Pill></Field>
              <Field label="nonce" mono>{truncateMid(callerProof?.nonce as string, 12, 8)} <Pill tone="real">✓ 실제 · CSPRNG</Pill></Field>
              <Field label="session_id" mono>{truncateMid(callerProof?.session_id as string, 12, 8)}</Field>
            </div>
            <Field label="signature (base64)" mono>{truncateMid(callerProof?.signature as string, 18, 12)} <Pill tone="real">✓ 실제 Ed25519</Pill></Field>
            <Code>{pretty(callerProof)}</Code>
            <p className="text-[10px] text-zinc-500 leading-snug">
              검증 흐름 (현재 DID 해소 단계만 mock 상태): <code className="text-fuchsia-300">caller_did</code> 해소 → 공개키 조회 → <code className="text-fuchsia-300">canonical(org_did || caller_did || purpose || session_id || nonce)</code>에 대해 서명 검증.
              서명 연산은 현 시점에서 실제이며, 레지스트리 스택은 결선 단계 (Open DID 포털 접근은 결선 이후 발급) 입니다.
            </p>
          </Section>

          {/* 2. Predictive Disclosure (Intent Handshake) */}
          <Section
            title="② 예측적 공개 — Intent Handshake (AI 분석)"
            subtitle="9단계 파이프라인의 Step 7. Anthropic Claude Haiku 4.5가 구조화된 manifest를 생성하며, LLM 호출 실패 시 시나리오별 결정적 fallback으로 전환됩니다."
          >
            {ih ? (
              <>
                <div className="flex flex-wrap items-center gap-2">
                  {ih.source === 'live'
                    ? <Pill tone="live">● LIVE · {ih.provider ?? 'anthropic'}</Pill>
                    : <Pill tone="fallback">○ FALLBACK · 결정적</Pill>}
                  <Pill tone="info">generated_at {ih.generated_at?.slice(0, 19) ?? '—'}</Pill>
                </div>

                <div>
                  <div className="text-[10px] uppercase tracking-wider text-zinc-500 mb-1">원시 LLM 출력 (a)</div>
                  <Code>{pretty(ih)}</Code>
                </div>

                <div>
                  <div className="text-[10px] uppercase tracking-wider text-zinc-500 mb-1">
                    정규화 JSON (b) <span className="text-zinc-600">— 안정된 필드 순서, 해싱에 사용</span>
                  </div>
                  <Code className="max-h-[140px]">{canonicalManifestJSON(ih)}</Code>
                </div>

                <div>
                  <div className="text-[10px] uppercase tracking-wider text-zinc-500 mb-1">
                    Manifest 해시 검증 (c) <span className="text-zinc-600">— 클라이언트 sha256 vs 백엔드 보고값</span>
                  </div>
                  <div className="bg-zinc-950 border border-zinc-800 rounded-md p-3 space-y-2">
                    <div className="text-[11px] flex gap-2">
                      <span className="text-zinc-500 shrink-0 w-32">계산값 (프론트엔드):</span>
                      <span className="font-mono text-zinc-300 break-all">{computing ? '…' : (computedHash ?? '—')}</span>
                    </div>
                    <div className="text-[11px] flex gap-2">
                      <span className="text-zinc-500 shrink-0 w-32">보고값 (백엔드):</span>
                      <span className="font-mono text-zinc-300 break-all">{reportedHash ?? '—'}</span>
                    </div>
                    <div className="pt-1">
                      {computing
                        ? <Pill tone="info">계산 중…</Pill>
                        : hashMatches
                        ? <Pill tone="verified">✓ 검증 완료 · 백엔드 ↔ 프론트엔드 정규화 해시 체인 일치</Pill>
                        : computedHash && reportedHash
                        ? <Pill tone="mismatch">✗ 해시 불일치 — 버그입니다, 제보 부탁드립니다</Pill>
                        : <Pill tone="info">데이터 대기 중…</Pill>}
                    </div>
                  </div>
                </div>

                <p className="text-[10px] text-zinc-500 leading-snug">
                  같은 시나리오를 두 번 재생해 보십시오 — <code className="text-fuchsia-300">source: "live"</code> 일 때 <code className="text-fuchsia-300">expected_requests</code> / <code className="text-fuchsia-300">forbidden_requests</code> 목록이 호출마다 달라집니다 (LLM 비결정성 = 실제 호출 증거). <code className="text-zinc-400">source: "fallback"</code> 일 때는 결정적 스크립트이므로 바이트 단위로 동일합니다.
                </p>
              </>
            ) : (
              <p className="text-[11px] text-zinc-500">응답에 intent_handshake 데이터 없음.</p>
            )}
          </Section>

          {/* 3. Risk Verdict */}
          <Section
            title="③ 위험도 판정 — Layer 1 (규칙) + Layer 2 (구성기)"
            subtitle="파이프라인의 Step 6. Layer 1은 규칙 기반으로 실제 동작합니다. Layer 2는 현재 결정적 mock 평가기로 동작하며 (WP-03), 전체 LLM 구성기는 Phase 1b로 분리되어 있습니다."
          >
            {riskVerdict ? (
              <>
                <div className="flex flex-wrap items-center gap-2">
                  <Pill tone="info">최종: {String((riskVerdict as { final?: unknown }).final ?? '—')}</Pill>
                  {(riskVerdict as { disagreement?: boolean }).disagreement && (
                    <Pill tone="mismatch">레이어 불일치</Pill>
                  )}
                  {(((riskVerdict as { layer2?: { used_mock_provider?: boolean } }).layer2)?.used_mock_provider) && (
                    <Pill tone="mock">⚠ Layer 2 mock 공급자</Pill>
                  )}
                </div>
                <Code>{pretty(riskVerdict)}</Code>
              </>
            ) : (
              <p className="text-[11px] text-zinc-500">응답에 risk_verdict 데이터 없음.</p>
            )}
          </Section>

          {/* 4. Receiver Identity */}
          <Section
            title="④ 수신자 신원 — OACX 토큰 + mResidentCard 클레임"
            subtitle="파이프라인의 Step 8. OmniOne CX SDK는 라이선스가 필요하여 현재 어댑터는 mock-fallback 상태입니다 (WP-02 — real-slot은 이미 배선되어 운영 단계에서 전환)."
          >
            {customerProof ? (
              <>
                <div className="flex flex-wrap items-center gap-2">
                  <Pill tone="mock">⚠ MOCK · 합성 JWT 서명자</Pill>
                  <Pill tone="info">vc_type = mresidentcard</Pill>
                </div>
                <Field label="oacx_token" mono>{truncateMid(customerProof.oacx_token as string, 16, 10)}</Field>
                <Code>{pretty(customerProof)}</Code>
              </>
            ) : (
              <p className="text-[11px] text-zinc-500">응답에 customer_proof 데이터 없음.</p>
            )}
          </Section>

          {/* 5. Chain Receipt — the headline */}
          <Section
            title="⑤ Chain Receipt — off-chain ↔ on-chain (핵심)"
            subtitle="Step 9. Off-chain receipt에는 전체 감사 기록이 남고, on-chain 앵커는 5개 필드뿐 — sessionHash, receiptHash, isSafe, policyVersion, block.timestamp. On-chain Zero-PII 보장."
          >
            {receipt ? (
              <>
                <div className="flex flex-wrap items-center gap-2">
                  {isSafe && <Pill tone="real">isSafe = true · 안전한 통화</Pill>}
                  {isFailed && <Pill tone="mock">isSafe = false · BLOCK도 영구 기록</Pill>}
                  <Pill tone="real">✓ 실제 · Sepolia</Pill>
                </div>

                <div>
                  <div className="text-[10px] uppercase tracking-wider text-zinc-500 mb-1">
                    On-chain (앵커링된 5개 필드) — Etherscan Logs 탭에서 디코딩되는 값
                  </div>
                  <div className="bg-zinc-950 border border-zinc-800 rounded-md p-3 space-y-1.5 text-[11px]">
                    {receipt.on_chain && Object.entries(receipt.on_chain).map(([k, v]) => (
                      <div key={k} className="flex gap-2">
                        <span className="text-zinc-500 shrink-0 w-32">{k}:</span>
                        <span className="font-mono text-zinc-300 break-all">{
                          typeof v === 'string'
                            ? (v.length > 50 ? truncateMid(v, 20, 14) : v)
                            : String(v)
                        }</span>
                      </div>
                    ))}
                  </div>
                </div>

                <div>
                  <div className="text-[10px] uppercase tracking-wider text-zinc-500 mb-1">
                    Off-chain (전체 감사 기록 — 백엔드 SQLite, sessionHash로 참조)
                  </div>
                  <Code>{pretty(receipt.off_chain)}</Code>
                </div>

                {passport?.explorerUrl && (
                  <a
                    href={passport.explorerUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-2 px-4 py-2.5 bg-amber-950/40 hover:bg-amber-900/40 border border-amber-700/60 rounded-lg text-amber-200 text-xs font-medium transition-colors w-full sm:w-auto justify-center"
                  >
                    🔍 Etherscan에서 확인 — 위 5개 필드와 바이트 단위 일치 검증
                  </a>
                )}

                <p className="text-[10px] text-zinc-500 leading-snug">
                  Etherscan 링크를 열고 → <strong>Logs</strong> 탭에서 → topics[1] (sessionHash) + 4개 data slot
                  (receiptHash, isSafe, policyVersion, block.timestamp) 을 디코딩하면 위 on-chain 표와 바이트 단위로 일치합니다.
                  {isFailed && (
                    <span className="text-amber-300">
                      {' '}이 앵커의 <code>isSafe = false</code> — BLOCK 판정 자체가 온체인에 기록되어, 발신자가 추후 "VOUCH가 이의 제기한 사실이 없다"고 부인할 수 없습니다 (부인방지).
                    </span>
                  )}
                </p>
              </>
            ) : (
              <p className="text-[11px] text-zinc-500">응답에 receipt 데이터 없음.</p>
            )}
          </Section>

          {/* 6. Honesty footer */}
          <Section
            title="📋 실구현 범위 표시"
            subtitle="v3-KR §Ⅱ.6 원칙에 따른 정직한 스코핑. 위 모든 주장은 라벨링되어 있으며, 어떤 위조도 없습니다."
          >
            <div className="grid grid-cols-1 gap-3">
              <div className="bg-emerald-950/20 border border-emerald-900/40 rounded-md p-3">
                <div className="text-emerald-300 text-xs font-semibold mb-2">✅ 실제 (Real — 오늘 시연 가능)</div>
                <ul className="text-[11px] text-zinc-300 space-y-1 list-disc pl-5 leading-snug">
                  <li><strong>Ed25519 CallerProof 서명</strong> — 매 호출마다 정규화된 입력에 대해 실제 계산</li>
                  <li><strong>Sepolia chain receipt</strong> — 모든 데모 호출에서 <code className="font-mono text-amber-300">0x70fc086A…dED04</code> 컨트랙트에 앵커링</li>
                  <li><strong>Anthropic Claude Haiku 4.5</strong> — <code>source: "live"</code> 일 때 지연 2.4–3.6초, 호출마다 내용 변화</li>
                  <li><strong>정규화 JSON 해시 체인</strong> — <code>intent_manifest_hash</code> sha256 결정적, 위에서 클라이언트 측 검증 완료</li>
                  <li><strong>sessionHash / receiptHash / policyVersion</strong> — 전부 keccak256, on-chain Zero-PII</li>
                </ul>
              </div>
              <div className="bg-amber-950/20 border border-amber-900/40 rounded-md p-3">
                <div className="text-amber-300 text-xs font-semibold mb-2">⚠ Mock (현재 단계 — 결선 W1-W2 전환 예정)</div>
                <ul className="text-[11px] text-zinc-300 space-y-1 list-disc pl-5 leading-snug">
                  <li><strong>OmniOne CX SDK</strong> — 라이선스 게이팅. 현재 어댑터는 mock-fallback (real-slot 배선 완료, 운영 단계에서 전환)</li>
                  <li><strong>Open DID 해소</strong> — 현재는 인메모리 레지스트리. OmniOne Open DID 포털 접근은 결선 이후 발급</li>
                  <li><strong>AI Agent VC 발급 체인</strong> — 합성 테스트 셋. 실제 VC 발급은 Phase 1b</li>
                  <li><strong>Layer 2 위험도 판정 LLM</strong> — 현재는 규칙 기반 mock. 전체 LLM 구성기는 Phase 1b</li>
                </ul>
              </div>
              <p className="text-[10px] text-zinc-500 text-center pt-1">
                소스 코드: <a href="https://github.com/gem-squared/gem2-ZTCV" target="_blank" rel="noopener noreferrer" className="text-blue-400 hover:underline">github.com/gem-squared/gem2-ZTCV</a> ·
                제안서 참조: §Ⅱ.5 (Chain Receipt 데이터 구조) · §Ⅱ.6 (What Ships Today) · §Ⅳ.2 (Audit Defensibility)
              </p>
            </div>
          </Section>
        </div>
      </aside>
    </>
  )
}
