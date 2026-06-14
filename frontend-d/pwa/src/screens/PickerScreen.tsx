import { SCENARIO_LIST, type ScenarioId, type BadgeTone } from '../scenarios'

interface Props {
  onPick: (id: ScenarioId) => void
}

const TONE_BORDER: Record<BadgeTone, string> = {
  safe: 'border-l-safe',
  block: 'border-l-block',
  warn: 'border-l-warn'
}

const TONE_BADGE: Record<BadgeTone, string> = {
  safe: 'bg-safe/15 text-safe',
  block: 'bg-block/15 text-block',
  warn: 'bg-warn/15 text-warn'
}

const TONE_ICON: Record<BadgeTone, string> = {
  safe: '🏛️',
  block: '⚠️',
  warn: '🏛️'
}

export function PickerScreen({ onPick }: Props) {
  return (
    <div className="min-h-screen bg-bg-base text-zinc-100 px-6 py-10">
      <header className="text-center mb-10">
        <h1 className="text-2xl font-semibold mb-2">VOUCH 통화 인증</h1>
        <p className="text-zinc-400 text-xs">제로트러스트 통화 검증 프로토콜 데모</p>
      </header>
      <div className="space-y-3 max-w-md mx-auto">
        {SCENARIO_LIST.map((s) => (
          <button
            key={s.id}
            onClick={() => onPick(s.id)}
            className={`w-full bg-bg-card hover:bg-bg-elevated active:bg-bg-elevated transition-colors rounded-xl p-5 border-l-4 ${TONE_BORDER[s.badgeTone]} text-left flex items-center justify-between gap-3`}
          >
            <div className="flex items-center gap-3 min-w-0">
              <div className="text-xl shrink-0">{TONE_ICON[s.badgeTone]}</div>
              <div className="min-w-0">
                <div className="font-semibold truncate">{s.title}</div>
                <div className="text-xs text-zinc-400 mt-1 truncate">{s.subtitle}</div>
              </div>
            </div>
            <div
              className={`text-xs px-3 py-1 rounded-full shrink-0 ${TONE_BADGE[s.badgeTone]}`}
            >
              {s.badge}
            </div>
          </button>
        ))}
      </div>
      <p className="text-center text-xs text-zinc-600 mt-8">체험할 시나리오를 선택하세요</p>
    </div>
  )
}
