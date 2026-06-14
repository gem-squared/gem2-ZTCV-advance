// /signup/ — VOUCH 회원가입 시뮬레이션 페이지
//
// Visualizes the canonical OmniOne CX 활용시 6-step flow per 필수과제
// slide 21: (1) 회원가입 (2) Redirection (3) 신분증 요청 (4) 신분증 제출
// (5) Response (6) 서비스 제공. Honesty-labeled as simulation — actual
// OmniOne CX integration code lives in internal/identity/omnione.go and
// activates after 결선 라이선스 발급.
//
// MANUAL progression — user clicks "다음 →" to advance each content phase
// (1, 3, 4, 5). Phase 2 is a transitional loading state that AUTO-advances
// to Phase 3 (표준인증창 활성화) after 3s, mirroring real redirect UX.
// Skip button always visible.
import { useEffect, useState } from 'react'

type Phase = 1 | 2 | 3 | 4 | 5 | 6

const PHASE_LABELS: Record<Phase, string> = {
  1: '회원가입 신청',
  2: '표준인증창 이동',
  3: '신분증 요청',
  4: '신분증 제출',
  5: '검증 응답',
  6: '가입 완료',
}

const NEXT_LABEL: Record<Phase, string> = {
  1: '정부 모바일 신분증으로 시작 →',
  2: '표준인증창 띄우기 →',
  3: 'QR 스캔 시뮬레이션 →',
  4: '신분증 제출 →',
  5: '검증 결과 확인 →',
  6: '',
}

export function SignupPage() {
  const [phase, setPhase] = useState<Phase>(1)

  // Phase 2 is a transitional "redirect to OmniOne CX" loading state — it
  // auto-advances to Phase 3 (표준인증창 활성화) after 3s. All other phases
  // require explicit user click so the presenter controls the demo pacing.
  useEffect(() => {
    if (phase !== 2) return
    const t = setTimeout(() => {
      setPhase((p) => (p === 2 ? 3 : p))
    }, 3000)
    return () => clearTimeout(t)
  }, [phase])

  const nextPhase = () => {
    setPhase((p) => (p < 6 ? ((p + 1) as Phase) : p))
  }

  const prevPhase = () => {
    setPhase((p) => (p > 1 ? ((p - 1) as Phase) : p))
  }

  const replay = () => {
    setPhase(1)
  }

  const goToFlow = () => {
    window.location.href = '/flow/'
  }

  return (
    <div className="min-h-screen w-full bg-black text-zinc-100 relative overflow-x-hidden">
      {/* Header strip */}
      <header className="px-6 py-4 border-b border-zinc-900/80 relative z-20 bg-black/80 backdrop-blur-sm">
        <div className="flex items-center gap-4 flex-wrap">
          <div className="flex items-center gap-3">
            <div className="text-2xl">🛡️</div>
            <div>
              <h1 className="text-lg font-semibold tracking-tight">VOUCH</h1>
              <div className="text-[10px] text-zinc-500 uppercase tracking-wider">
                Sign-up · OmniOne CX 정부 모바일 신분증 연동
              </div>
            </div>
          </div>
          <div className="flex-1 min-w-0" />
          <button
            onClick={goToFlow}
            className="px-3 py-1.5 text-xs bg-bg-card hover:bg-bg-elevated rounded-lg transition-colors text-zinc-400"
          >
            건너뛰기 → 데모 보기 →
          </button>
        </div>
      </header>

      {/* Phase progress bar */}
      <div className="px-6 py-4 border-b border-zinc-900/50">
        <div className="max-w-5xl mx-auto flex items-center gap-2">
          {([1, 2, 3, 4, 5, 6] as Phase[]).map((p) => {
            const state = p < phase ? 'done' : p === phase ? 'active' : 'pending'
            const color =
              state === 'done'
                ? 'bg-emerald-500/80'
                : state === 'active'
                ? 'bg-blue-400'
                : 'bg-zinc-700'
            return (
              <div key={p} className="flex items-center gap-2 flex-1">
                <div className="flex items-center gap-2">
                  <span
                    className={`w-6 h-6 rounded-full flex items-center justify-center text-[10px] font-mono font-bold ${
                      state === 'done'
                        ? 'bg-emerald-950 border border-emerald-700 text-emerald-300'
                        : state === 'active'
                        ? 'bg-blue-950 border border-blue-600 text-blue-300'
                        : 'bg-zinc-900 border border-zinc-800 text-zinc-600'
                    } ${state === 'active' ? 'animate-pulse' : ''}`}
                  >
                    {state === 'done' ? '✓' : p}
                  </span>
                  <span
                    className={`text-[10px] uppercase tracking-wider ${
                      state === 'done'
                        ? 'text-emerald-400'
                        : state === 'active'
                        ? 'text-blue-300'
                        : 'text-zinc-600'
                    }`}
                  >
                    {PHASE_LABELS[p]}
                  </span>
                </div>
                {p < 6 && (
                  <div className="flex-1 h-px bg-zinc-800 mx-1">
                    <div
                      className={`h-px ${color} transition-all duration-300`}
                      style={{ width: state === 'pending' ? '0%' : '100%' }}
                    />
                  </div>
                )}
              </div>
            )
          })}
        </div>
      </div>

      {/* Main content */}
      <main className="px-6 py-10">
        <div className="max-w-5xl mx-auto">
          {phase === 1 && <Phase1Hero />}
          {phase === 2 && <Phase2Redirect />}
          {phase === 3 && <Phase3QrRequest />}
          {phase === 4 && <Phase4MobileSubmit />}
          {phase === 5 && <Phase5VerifyResponse />}
          {phase === 6 && <Phase6Complete onReplay={replay} onContinue={goToFlow} />}
        </div>

        {/* Phase navigation — manual click controls (visible on content phases 1, 3, 4, 5).
            Phase 2 is a redirect/loading state that auto-advances after 3s — show
            an auto-progression hint instead of buttons. */}
        {phase === 2 && (
          <div className="max-w-5xl mx-auto mt-8 text-center">
            <p className="text-[11px] text-zinc-500">
              <span className="inline-block w-1.5 h-1.5 rounded-full bg-blue-400 animate-pulse mr-1.5 align-middle" />
              표준인증창이 곧 활성화됩니다 (자동 진행)
            </p>
          </div>
        )}
        {(phase === 1 || phase === 3 || phase === 4 || phase === 5) && (
          <div className="max-w-5xl mx-auto mt-10 flex items-center justify-between gap-3">
            <button
              onClick={prevPhase}
              disabled={phase === 1}
              className="px-4 py-2.5 text-xs bg-bg-card hover:bg-bg-elevated rounded-lg disabled:opacity-30 disabled:cursor-not-allowed transition-colors text-zinc-400"
            >
              ← 이전
            </button>
            <button
              onClick={nextPhase}
              className="px-6 py-3 text-sm bg-blue-600 hover:bg-blue-500 text-white rounded-lg font-semibold transition-colors shadow-lg shadow-blue-900/40 animate-pulse"
            >
              {NEXT_LABEL[phase]}
            </button>
            <button
              onClick={replay}
              className="px-4 py-2.5 text-xs bg-bg-card hover:bg-bg-elevated rounded-lg transition-colors text-zinc-400"
            >
              ↻ 처음으로
            </button>
          </div>
        )}
      </main>

      {/* Honesty footer */}
      <footer className="px-6 py-6 border-t border-zinc-900/60 mt-10">
        <div className="max-w-5xl mx-auto text-center space-y-1">
          <p className="text-[11px] text-zinc-500 leading-relaxed">
            <span className="text-amber-400">⚠ 데모 시뮬레이션</span> — 본 페이지는 OmniOne CX 활용시 정식 6-step 흐름 (회원가입 → Redirection → 신분증 요청 → 신분증 제출 → Response → 서비스 제공) 의 시각적 재현입니다.
          </p>
          <p className="text-[10px] text-zinc-600 leading-relaxed">
            실제 OmniOne CX 연동 코드는 <code className="font-mono text-zinc-400">internal/identity/omnione.go</code> 에 보유 (REST client + JWT 파싱 구현) — 결선 라이선스 발급 후 <code className="font-mono text-zinc-400">OMNIONE_CX_MODE=real</code> 환경변수 1줄 전환으로 활성됩니다. 필수과제 §4.2 (모바일 신분증 연동) 가능성 demonstration.
          </p>
        </div>
      </footer>
    </div>
  )
}

/* ─── Phase 1 ─── Hero ────────────────────────────────────────────── */
function Phase1Hero() {
  return (
    <div
      className="text-center py-16"
      style={{ animation: 'precall-item-in 320ms ease-out 1' }}
    >
      <div className="text-6xl mb-6">🛡️</div>
      <h2 className="text-3xl font-bold mb-3 tracking-tight">VOUCH 시작하기</h2>
      <p className="text-zinc-400 text-base mb-10 leading-relaxed">
        정부 모바일 신분증으로 <span className="text-emerald-400 font-semibold">간편하게 가입</span>합니다.
        <br />
        별도 회원 정보 입력은 필요하지 않습니다.
      </p>
      <div className="inline-flex flex-col items-center gap-4 bg-zinc-950/60 border border-zinc-800/60 rounded-xl px-10 py-8">
        <div className="text-[10px] uppercase tracking-wider text-zinc-500">Powered by</div>
        <div className="text-sm font-semibold text-zinc-200">OmniOne CX · 표준인증창</div>
        <div className="text-[11px] text-zinc-400 mt-1">아래 "정부 모바일 신분증으로 시작 →" 버튼을 클릭해주세요</div>
      </div>
    </div>
  )
}

/* ─── Phase 2 ─── Redirect ────────────────────────────────────────── */
function Phase2Redirect() {
  return (
    <div
      className="text-center py-20"
      style={{ animation: 'precall-item-in 320ms ease-out 1' }}
    >
      <div className="inline-flex items-center gap-4 bg-zinc-950/60 border border-zinc-800/60 rounded-xl px-10 py-8">
        <span className="text-blue-400 text-2xl animate-spin inline-block">⟳</span>
        <div className="text-left">
          <div className="text-sm font-semibold text-zinc-100">OmniOne CX 표준인증창으로 이동 중…</div>
          <div className="text-[11px] text-zinc-500 mt-1">
            VOUCH 서비스 → <span className="font-mono text-zinc-400">cx.raonsecure.co.kr</span>
          </div>
        </div>
      </div>
    </div>
  )
}

/* ─── Phase 3 ─── QR Request (표준인증창) ──────────────────────────── */
function Phase3QrRequest() {
  return (
    <div
      className="grid grid-cols-1 md:grid-cols-2 gap-8 items-center max-w-4xl mx-auto py-8"
      style={{ animation: 'precall-item-in 320ms ease-out 1' }}
    >
      {/* 표준인증창 panel */}
      <div className="bg-zinc-950/80 border border-zinc-700/60 rounded-2xl p-6 shadow-2xl">
        <div className="flex items-center justify-between mb-4 pb-3 border-b border-zinc-800">
          <div>
            <div className="text-[10px] uppercase tracking-wider text-zinc-500">OmniOne CX</div>
            <div className="text-sm font-semibold text-zinc-100">표준인증창</div>
          </div>
          <span className="text-[10px] text-zinc-500 font-mono">id=&quot;oacxDiv&quot;</span>
        </div>
        <h3 className="text-base font-semibold text-zinc-100 mb-2 text-center">모바일 신분증 인증</h3>
        <p className="text-[11px] text-zinc-400 text-center mb-5 leading-relaxed">
          모바일 신분증 앱에서<br />
          QR 코드를 스캔해주세요.
        </p>
        {/* QR code visual */}
        <div className="flex justify-center mb-4">
          <DecorativeQR />
        </div>
        <div className="bg-blue-950/30 border border-blue-800/40 rounded-md p-3 text-center">
          <div className="text-[10px] text-blue-300 mb-1 uppercase tracking-wider">시간 제한</div>
          <div className="text-sm font-mono text-blue-200">02:47</div>
        </div>
        <div className="mt-4 grid grid-cols-4 gap-2">
          {['주민등록증', '운전면허증', '국가보훈증', '청소년증'].map((t, i) => (
            <div key={i} className="text-center">
              <div className="w-full aspect-square bg-zinc-900 border border-zinc-800 rounded flex items-center justify-center mb-1">
                <span className="text-xl">📇</span>
              </div>
              <div className="text-[9px] text-zinc-500">{t}</div>
            </div>
          ))}
        </div>
      </div>

      {/* 흐름 설명 */}
      <div className="space-y-4">
        <div className="text-[10px] uppercase tracking-wider text-zinc-500">현재 진행 단계</div>
        <h3 className="text-2xl font-bold text-zinc-100">정부 모바일 신분증 요청</h3>
        <p className="text-zinc-400 text-sm leading-relaxed">
          OmniOne CX가 표준인증창을 띄우고, 시민의 휴대폰에서 정부 모바일 신분증 앱을 통한 QR 스캔을 기다립니다.
        </p>
        <div className="bg-zinc-950/40 border border-zinc-800/40 rounded-md p-4 space-y-2 text-[11px]">
          <div className="flex gap-2">
            <span className="text-zinc-500 w-24 shrink-0">호출 endpoint:</span>
            <code className="font-mono text-zinc-300 break-all">/oacx/api/v1.0/trans</code>
          </div>
          <div className="flex gap-2">
            <span className="text-zinc-500 w-24 shrink-0">signType:</span>
            <code className="font-mono text-zinc-300">ENT_MID</code>
          </div>
          <div className="flex gap-2">
            <span className="text-zinc-500 w-24 shrink-0">발행자:</span>
            <span className="text-zinc-300">행정안전부</span>
          </div>
        </div>
      </div>
    </div>
  )
}

/* ─── Phase 4 ─── Mobile Submit ───────────────────────────────────── */
function Phase4MobileSubmit() {
  return (
    <div
      className="grid grid-cols-1 md:grid-cols-2 gap-8 items-center max-w-4xl mx-auto py-8"
      style={{ animation: 'precall-item-in 320ms ease-out 1' }}
    >
      {/* Phone bezel mock */}
      <div className="flex justify-center">
        <div className="w-64 bg-zinc-950 border-[3px] border-zinc-700 rounded-[2.5rem] p-6 shadow-2xl">
          <div className="flex justify-between items-center mb-6">
            <span className="text-[10px] text-zinc-500 font-mono">9:41</span>
            <div className="flex items-center gap-1 text-[10px] text-zinc-500">
              <span>📶</span>
              <span>🔋</span>
            </div>
          </div>
          <div className="text-center mb-4">
            <div className="text-[10px] uppercase tracking-wider text-zinc-500 mb-1">모바일 신분증 앱</div>
            <div className="text-sm font-semibold text-zinc-100">행정안전부</div>
          </div>
          <div
            className="bg-gradient-to-br from-blue-900/40 to-emerald-900/40 border border-blue-700/40 rounded-xl p-4 mb-4"
            style={{ animation: 'anchor-card-in 480ms ease-out 1' }}
          >
            <div className="text-[10px] uppercase tracking-wider text-blue-300/80 mb-2">모바일 주민등록증</div>
            <div className="text-base font-semibold text-zinc-100 mb-1">
              김민수 <span className="text-[10px] text-amber-400 ml-1">⚠ MOCK</span>
            </div>
            <div className="text-[11px] text-zinc-400 font-mono">880101-1******</div>
            <div className="mt-3 pt-3 border-t border-zinc-700/50">
              <div className="text-[10px] text-zinc-500">발급일</div>
              <div className="text-[11px] text-zinc-300">2024.03.15</div>
            </div>
          </div>
          <div className="bg-emerald-950/30 border border-emerald-700/40 rounded-lg py-2.5 text-center">
            <div className="text-xs text-emerald-300 font-semibold">신분증 제출 중...</div>
            <div className="text-[10px] text-emerald-400/70 mt-0.5 font-mono">eVP 전송 →</div>
          </div>
        </div>
      </div>

      <div className="space-y-4">
        <div className="text-[10px] uppercase tracking-wider text-zinc-500">현재 진행 단계</div>
        <h3 className="text-2xl font-bold text-zinc-100">신분증 제출</h3>
        <p className="text-zinc-400 text-sm leading-relaxed">
          시민이 모바일 신분증 앱에서 정부 발행 신분증 (모바일 주민등록증 / mDL 등) 을 OmniOne CX에 안전하게 제출합니다.
        </p>
        <div className="bg-zinc-950/40 border border-zinc-800/40 rounded-md p-4 space-y-2 text-[11px]">
          <div className="flex gap-2">
            <span className="text-zinc-500 w-24 shrink-0">전송 데이터:</span>
            <span className="text-zinc-300">eVP (encrypted VP)</span>
          </div>
          <div className="flex gap-2">
            <span className="text-zinc-500 w-24 shrink-0">신분증 종류:</span>
            <span className="text-zinc-300">모바일 주민등록증</span>
          </div>
          <div className="flex gap-2">
            <span className="text-zinc-500 w-24 shrink-0">표준:</span>
            <span className="text-zinc-300">W3C VC Data Model</span>
          </div>
        </div>
      </div>
    </div>
  )
}

/* ─── Phase 5 ─── Verify Response (JWT 파싱) ──────────────────────── */
function Phase5VerifyResponse() {
  return (
    <div
      className="max-w-3xl mx-auto py-8 space-y-6"
      style={{ animation: 'precall-item-in 320ms ease-out 1' }}
    >
      <div className="text-center">
        <div className="text-[10px] uppercase tracking-wider text-zinc-500 mb-2">현재 진행 단계</div>
        <h3 className="text-2xl font-bold text-zinc-100">검증 응답 · JWT 파싱</h3>
        <p className="text-zinc-400 text-sm mt-3 leading-relaxed">
          VC-Verifier가 모바일 신분증 VP를 검증한 후 VOUCH 백엔드로 JWT를 전달합니다. 백엔드의 <code className="font-mono text-zinc-300">OACXProvider.VerifyToken()</code> 가 응답 envelope을 파싱합니다.
        </p>
      </div>

      <div className="bg-zinc-950 border border-zinc-800 rounded-lg p-4">
        <div className="text-[10px] uppercase tracking-wider text-zinc-500 mb-2">
          OmniOne CX 응답 (parsed JSON envelope)
        </div>
        <pre className="text-[10px] font-mono text-zinc-300 overflow-x-auto leading-relaxed">
{`{
  "resultCode": "200",
  "oacxCode": "OK",
  "data": {
    "vcTypeCode": "mresidentcard",
    "name": "김민수",         // → sha256 → name_hash
    "birthDate": "19880101",
    "licNo": "880101-1******", // → sha256 → doc_no_hash
    "locpanm": "행정안전부"
  }
}`}
        </pre>
      </div>

      <div className="bg-emerald-950/20 border border-emerald-800/40 rounded-lg p-4">
        <div className="flex items-center gap-2 mb-3">
          <span className="text-emerald-400 animate-spin inline-block">⟳</span>
          <div className="text-sm font-semibold text-emerald-300">검증 중…</div>
        </div>
        <div className="space-y-1.5 text-[11px]">
          <div className="flex gap-2">
            <span className="text-emerald-400">✓</span>
            <span className="text-zinc-300">resultCode = 200 — 검증 성공</span>
          </div>
          <div className="flex gap-2">
            <span className="text-emerald-400">✓</span>
            <span className="text-zinc-300">vcTypeCode = mresidentcard — 모바일 주민등록증 확인</span>
          </div>
          <div className="flex gap-2">
            <span className="text-emerald-400">✓</span>
            <span className="text-zinc-300">name → sha256 hash 변환 (Zero-PII 보존)</span>
          </div>
          <div className="flex gap-2">
            <span className="text-emerald-400">✓</span>
            <span className="text-zinc-300">licNo → sha256 hash 변환 (Zero-PII 보존)</span>
          </div>
          <div className="flex gap-2">
            <span className="text-emerald-400">✓</span>
            <span className="text-zinc-300">MobileIDClaims 객체 구성 → VOUCH 시민 프로필 생성</span>
          </div>
        </div>
      </div>
    </div>
  )
}

/* ─── Phase 6 ─── Complete ────────────────────────────────────────── */
function Phase6Complete({ onReplay, onContinue }: { onReplay: () => void; onContinue: () => void }) {
  return (
    <div
      className="max-w-3xl mx-auto py-8 space-y-6 text-center"
      style={{ animation: 'anchor-card-in 480ms ease-out 1' }}
    >
      <div className="text-6xl">✅</div>
      <h2 className="text-3xl font-bold text-zinc-100">가입 완료</h2>
      <p className="text-zinc-400 text-base leading-relaxed">
        정부 모바일 신분증으로 VOUCH 회원가입이 완료되었습니다.<br />
        이제 <span className="text-emerald-400 font-semibold">통화 검증 데모</span>를 시작할 수 있습니다.
      </p>

      <div className="bg-zinc-950/60 border border-emerald-700/30 rounded-2xl px-8 py-6 max-w-md mx-auto">
        <div className="text-[10px] uppercase tracking-wider text-emerald-400/70 mb-3">
          시민 프로필 — VOUCH 가입자
        </div>
        <div className="space-y-2 text-left">
          <div className="flex justify-between text-sm">
            <span className="text-zinc-500">이름</span>
            <span className="text-zinc-100 font-medium">김민수 <span className="text-[10px] text-amber-400 ml-1">⚠ MOCK</span></span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-zinc-500">신분증 종류</span>
            <span className="text-zinc-100">모바일 주민등록증</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-zinc-500">발행자</span>
            <span className="text-zinc-100">행정안전부</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-zinc-500">VC Type</span>
            <span className="text-zinc-100 font-mono text-xs">mresidentcard</span>
          </div>
          <div className="flex justify-between text-sm pt-2 border-t border-zinc-800/60">
            <span className="text-zinc-500">name_hash</span>
            <span className="text-zinc-100 font-mono text-[10px]">0x4b7c2a…</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-zinc-500">검증 완료</span>
            <span className="text-emerald-400 text-xs">✓ Zero-PII 보존</span>
          </div>
        </div>
      </div>

      <div className="flex flex-col sm:flex-row gap-3 justify-center pt-4">
        <button
          onClick={onContinue}
          className="px-8 py-3 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-sm font-semibold transition-colors shadow-lg shadow-blue-900/40"
        >
          VOUCH 데모 보기 →
        </button>
        <button
          onClick={onReplay}
          className="px-6 py-3 bg-bg-card hover:bg-bg-elevated rounded-lg text-sm text-zinc-300 transition-colors"
        >
          ↻ 다시 보기
        </button>
      </div>
    </div>
  )
}

/* ─── Decorative QR (NOT a scannable QR) ──────────────────────────── */
function DecorativeQR() {
  const cells = 21
  const pattern: boolean[][] = Array.from({ length: cells }, (_, r) =>
    Array.from({ length: cells }, (_, c) => {
      // Finder patterns (3 corners)
      const inTopLeft = r < 7 && c < 7
      const inTopRight = r < 7 && c >= cells - 7
      const inBotLeft = r >= cells - 7 && c < 7
      if (inTopLeft || inTopRight || inBotLeft) {
        const dr = inTopLeft ? r : inTopRight ? r : r - (cells - 7)
        const dc = inTopLeft ? c : inTopRight ? c - (cells - 7) : c
        const onRing = dr === 0 || dr === 6 || dc === 0 || dc === 6
        const onInner = dr >= 2 && dr <= 4 && dc >= 2 && dc <= 4
        return onRing || onInner
      }
      // pseudo-random body (deterministic by r,c)
      const h = (r * 31 + c * 17 + 13) % 7
      return h < 3
    })
  )

  return (
    <div className="bg-white rounded-lg p-3 inline-block">
      <div
        className="grid"
        style={{ gridTemplateColumns: `repeat(${cells}, 6px)`, gap: 0 }}
      >
        {pattern.flatMap((row, r) =>
          row.map((on, c) => (
            <div
              key={`${r}-${c}`}
              className={on ? 'bg-zinc-900' : 'bg-white'}
              style={{ width: 6, height: 6 }}
            />
          ))
        )}
      </div>
      <div className="text-center text-[8px] text-zinc-500 mt-1 font-mono">demo QR</div>
    </div>
  )
}
