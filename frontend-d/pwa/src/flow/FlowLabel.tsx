// Flow line label — SVG text with a dark pill background, anchored
// near the midpoint between two anchor points. Includes an optional
// step-number badge (① ② … ⑦) so the protocol sequence is readable.

import type { XY } from './anchors'
import type { FlowColor } from './FlowLine'

interface Props {
  from: XY | null
  to:   XY | null
  text: string
  step?: number  // 1..7 — renders as a circled numeral
  color: FlowColor
  active: boolean
  /** Vertical offset from midpoint (positive = down). */
  offsetY?: number
  /** Horizontal offset from midpoint (positive = right). */
  offsetX?: number
}

const COLOR_FILL: Record<FlowColor, string> = {
  white:   'rgba(255,255,255,0.9)',
  blue:    '#93c5fd',
  fuchsia: '#f5d0fe',
  gold:    '#fde68a',
  safe:    '#86efac',
  block:   '#fca5a5',
}

const COLOR_BORDER: Record<FlowColor, string> = {
  white:   'rgba(255,255,255,0.4)',
  blue:    '#3b82f6',
  fuchsia: '#d946ef',
  gold:    '#fbbf24',
  safe:    '#22c55e',
  block:   '#ef4444',
}

const CIRCLED: Record<number, string> = {
  1: '①', 2: '②', 3: '③', 4: '④', 5: '⑤', 6: '⑥', 7: '⑦',
}

export function FlowLabel({ from, to, text, step, color, active, offsetY = -14, offsetX = 0 }: Props) {
  if (!active || !from || !to) return null

  const mx = (from.x + to.x) / 2 + offsetX
  const my = (from.y + to.y) / 2 + offsetY

  // Approximate text width — used to size the background pill.
  const fontSize = 12
  const padX = 10
  const padY = 5
  const approxCharW = fontSize * 0.62
  const prefix = step ? `${CIRCLED[step]} ` : ''
  const display = prefix + text
  const w = Math.max(60, display.length * approxCharW + padX * 2)
  const h = fontSize + padY * 2

  return (
    <g style={{ animation: 'flow-label-in 320ms ease-out forwards', opacity: 0 }}>
      <rect
        x={mx - w / 2}
        y={my - h / 2}
        width={w}
        height={h}
        rx={h / 2}
        fill="rgba(10, 10, 15, 0.92)"
        stroke={COLOR_BORDER[color]}
        strokeWidth={1}
      />
      <text
        x={mx}
        y={my + fontSize * 0.35}
        textAnchor="middle"
        fontSize={fontSize}
        fontWeight={600}
        fill={COLOR_FILL[color]}
        style={{
          fontFamily:
            "'Pretendard', 'Apple SD Gothic Neo', 'Noto Sans KR', system-ui, sans-serif",
        }}
      >
        {display}
      </text>
    </g>
  )
}
