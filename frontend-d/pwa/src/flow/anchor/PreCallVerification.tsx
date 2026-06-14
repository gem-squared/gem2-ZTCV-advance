// PreCallVerification — 4-step Korean checklist that animates while the
// Broker is in 'verifying' phase. Each item staggers in, spins briefly,
// then settles as a check. Fills receiver-column dead air during the
// 3–4s verification window with micro-storytelling.
//
// Ported from feature/dramatic-desktop-upgrade branch; adapted for Vite
// (framer-motion → CSS keyframes from index.css).
import { useEffect, useState } from 'react'

const ITEMS = [
  '발신자 신원 확인',
  '통화 목적 검증',
  'AI 안전 분석',
  '영수증 발행',
]
const STAGGER = 350
const CHECK_MS = 200

type State = 'hidden' | 'checking' | 'done'

interface Props {
  active: boolean
}

export function PreCallVerification({ active }: Props) {
  const [states, setStates] = useState<State[]>(ITEMS.map(() => 'hidden'))

  useEffect(() => {
    if (!active) {
      setStates(ITEMS.map(() => 'hidden'))
      return
    }
    const timers: ReturnType<typeof setTimeout>[] = []
    ITEMS.forEach((_, i) => {
      timers.push(setTimeout(() => setStates((p) => {
        const n = [...p]; n[i] = 'checking'; return n
      }), i * STAGGER))
      timers.push(setTimeout(() => setStates((p) => {
        const n = [...p]; n[i] = 'done'; return n
      }), i * STAGGER + CHECK_MS))
    })
    return () => { timers.forEach(clearTimeout) }
  }, [active])

  if (!active) return null

  return (
    <div className="w-full bg-zinc-950/60 border border-zinc-800/60 rounded-lg px-4 py-3.5">
      <div className="flex items-center gap-2 mb-3">
        <span className="w-2 h-2 rounded-full bg-blue-400 animate-pulse" />
        <span className="text-zinc-100 text-xs font-semibold">VOUCH 검증 중</span>
      </div>
      <div className="w-full space-y-1.5">
        {ITEMS.map((label, i) => {
          const visible = states[i] !== 'hidden'
          return (
            <div
              key={i}
              className="flex items-center gap-2 text-[11px]"
              style={visible
                ? { animation: 'precall-item-in 280ms ease-out 1', opacity: 1 }
                : { opacity: 0 }}
            >
              <span className="w-4 text-center">
                {states[i] === 'checking' && (
                  <span className="text-blue-400 animate-spin inline-block">⟳</span>
                )}
                {states[i] === 'done' && (
                  <span className="text-emerald-400">✓</span>
                )}
              </span>
              <span className={states[i] === 'done' ? 'text-zinc-400' : 'text-zinc-100 font-medium'}>
                {label}
              </span>
            </div>
          )
        })}
      </div>
    </div>
  )
}
