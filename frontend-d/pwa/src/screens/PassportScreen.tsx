import type { RunScenarioResult } from '../api'
import { SCENARIOS, type ScenarioId } from '../scenarios'
import type { Stamp } from '../types/passport'

interface Props {
  scenarioId: ScenarioId
  result: RunScenarioResult
  onReset: () => void
}

const BLOCK_REASON_TEXTS: Record<string, string> = {
  unknown_did: '발신자 신원을 확인할 수 없습니다. 보이스피싱 의심.',
  unauthorized_purpose: '허가된 통화 목적이 아닙니다. 권한 외 목적 통화 차단.'
}

// v2 stamp label remap — frontend-only display override.
// Backend label remains the canonical API contract; UI tightens phrasing
// to align with the Verification Broker thesis (per 09-demo-screen-spec.md).
const STAMP_LABEL_REMAP: Record<string, string> = {
  '송금 요구 감지': '고위험 사칭/송금 유도 패턴 감지',
  'AI Agent 권한 (권한 외 목적)': 'AI Agent 목적 권한'
}

function truncateHash(hash: string): string {
  if (hash.length <= 16) return hash
  return `${hash.slice(0, 10)}…${hash.slice(-6)}`
}

function displayLabel(stamp: Stamp): string {
  return STAMP_LABEL_REMAP[stamp.label] ?? stamp.label
}

export function PassportScreen({ scenarioId, result, onReset }: Props) {
  const { passport } = result.data
  const scenario = SCENARIOS[scenarioId]
  const isSafe = passport.outcome === 'SAFE'
  const offline = result.source === 'fixture'

  const hash = passport.receiptTxHash
  const isMockHash = !!hash && hash.toUpperCase().startsWith('0XMOCK')
  const hashLabel = isMockHash
    ? '영수증 해시 · 시뮬레이션 (데모 환경)'
    : '영수증 해시 · Sepolia 테스트넷'

  const HashInner = hash ? (
    <>
      <div className="flex items-center gap-2 text-xs text-zinc-500 mb-1">
        <span>🔗</span>
        <span>{hashLabel}</span>
      </div>
      <div className="text-xs font-mono text-zinc-300 truncate">
        {truncateHash(hash)}
      </div>
      {!isMockHash && (
        <div className="text-[10px] text-zinc-600 mt-1">
          Etherscan에서 확인 →
        </div>
      )}
    </>
  ) : null

  return (
    <div className="min-h-screen bg-bg-base text-zinc-100 px-6 py-10 flex flex-col relative">
      {offline && (
        <div className="absolute top-4 right-4 text-[10px] font-medium text-amber-300 bg-amber-950/60 px-2.5 py-1 rounded-full border border-amber-900/80">
          ⚠ 오프라인 모드
        </div>
      )}

      <div className="text-center mb-8">
        <div
          className={`text-7xl mb-3 ${
            isSafe
              ? 'drop-shadow-[0_0_24px_rgba(34,197,94,0.5)]'
              : 'drop-shadow-[0_0_24px_rgba(239,68,68,0.5)]'
          }`}
        >
          {isSafe ? '✅' : '❌'}
        </div>
        <h1 className={`text-3xl font-semibold ${isSafe ? 'text-safe' : 'text-block'}`}>
          {isSafe ? '안전한 통화' : '차단됨'}
        </h1>
      </div>

      <div className="bg-bg-card rounded-xl p-5 mb-4 max-w-md w-full mx-auto">
        <table className="w-full text-sm">
          <tbody>
            <tr>
              <td className="text-zinc-500 py-2 pr-2 align-top whitespace-nowrap">발신자</td>
              <td className="text-right align-top">
                <span className="text-zinc-200">{scenario.callerOrg}</span>
                <span className="text-zinc-500"> / </span>
                <span className="text-zinc-300">{scenario.callerLabel}</span>
              </td>
            </tr>
            {passport.stamps.map((stamp) => (
              <tr key={stamp.label}>
                <td className="text-zinc-500 py-2 pr-2 align-top">{displayLabel(stamp)}</td>
                <td className="text-right align-top">
                  {stamp.status === 'OK' ? (
                    <span className="text-safe">✓ 확인됨</span>
                  ) : (
                    <span className="text-block">✗ 미확인</span>
                  )}
                </td>
              </tr>
            ))}
            <tr>
              <td className="text-zinc-500 py-2 pr-2 align-top">위험도</td>
              <td className="text-right align-top">
                {isSafe ? (
                  <span className="inline-flex items-center gap-1.5 text-safe">
                    <span className="w-2 h-2 rounded-full bg-safe inline-block" />
                    낮음
                  </span>
                ) : (
                  <span className="inline-flex items-center gap-1.5 text-block">
                    <span className="w-2 h-2 rounded-full bg-block inline-block" />
                    높음
                  </span>
                )}
              </td>
            </tr>
          </tbody>
        </table>

        <div className="mt-4 pt-4 border-t border-bg-elevated text-sm text-zinc-300">
          {isSafe
            ? '본인 확인 완료. 안전한 통화입니다.'
            : BLOCK_REASON_TEXTS[passport.blockReason || ''] ||
              '통화가 차단되었습니다.'}
        </div>
      </div>

      {hash &&
        (isMockHash ? (
          <div className="bg-bg-card rounded-xl p-4 mb-4 max-w-md w-full mx-auto">
            {HashInner}
          </div>
        ) : (
          <a
            href={passport.explorerUrl || '#'}
            target="_blank"
            rel="noopener noreferrer"
            className="bg-bg-card rounded-xl p-4 mb-4 max-w-md w-full mx-auto block hover:bg-bg-elevated active:bg-bg-elevated transition-colors"
          >
            {HashInner}
          </a>
        ))}

      <div className="flex-1" />

      <button
        onClick={onReset}
        className="bg-bg-card hover:bg-bg-elevated active:bg-bg-elevated transition-colors rounded-xl p-4 max-w-md w-full mx-auto text-sm font-medium"
      >
        다른 시나리오 체험
      </button>

      <div className="text-center text-[10px] text-zinc-500 mt-4 max-w-md mx-auto">
        앱/Broker 레벨 Call Passport · OmniOne CX + Open DID 기반
      </div>
    </div>
  )
}
