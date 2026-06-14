// Animated SVG path that draws on when `active` flips to true, holds,
// then fades when `active` flips back to false. The line is a smooth
// cubic bezier between two anchor points so adjacent lines don't
// overlap visually.
//
// Includes:
//   - stroke-dasharray draw-on animation
//   - optional moving dot ("data packet") for high-signal edges
//   - optional arrowhead marker at the terminal end

import { useEffect, useMemo, useState } from 'react'
import type { XY } from './anchors'
import { LINE_DRAW_MS } from './timing'

export type FlowColor = 'white' | 'blue' | 'fuchsia' | 'gold' | 'safe' | 'block'
export type FlowStyle = 'solid' | 'dashed'

interface Props {
  from: XY | null
  to:   XY | null
  color: FlowColor
  style?: FlowStyle
  active: boolean
  /** If true, also animate a moving dot along the path (data packet). */
  movingDot?: boolean
  /** If true, draw an arrowhead at `to`. */
  arrow?: boolean
  /** Curvature factor (0..1). 0 = straight, 1 = very curved. */
  curve?: number
  /** Stroke width. */
  width?: number
}

const COLOR_MAP: Record<FlowColor, string> = {
  white:   'rgba(255,255,255,0.85)',
  blue:    '#3b82f6',
  fuchsia: '#d946ef',
  gold:    '#fbbf24',
  safe:    '#22c55e',
  block:   '#ef4444',
}

const MARKER_MAP: Record<FlowColor, string> = {
  white:   'url(#arrow-white)',
  blue:    'url(#arrow-blue)',
  fuchsia: 'url(#arrow-fuchsia)',
  gold:    'url(#arrow-gold)',
  safe:    'url(#arrow-safe)',
  block:   'url(#arrow-block)',
}

/**
 * Build a smooth cubic-bezier `d` attribute between two points.
 * Curvature is computed perpendicular to the from→to vector so the
 * bend always feels natural regardless of orientation.
 */
function bezierPath(from: XY, to: XY, curve: number): string {
  const dx = to.x - from.x
  const dy = to.y - from.y
  const dist = Math.sqrt(dx * dx + dy * dy)
  // Perpendicular offset for control points.
  const off = dist * curve * 0.25
  // Normalized perpendicular (rotate (dx,dy) by 90°).
  const nx = -dy / (dist || 1)
  const ny =  dx / (dist || 1)
  // Two control points, both pulled "up" along the perpendicular.
  const c1x = from.x + dx * 0.3 + nx * off
  const c1y = from.y + dy * 0.3 + ny * off
  const c2x = from.x + dx * 0.7 + nx * off
  const c2y = from.y + dy * 0.7 + ny * off
  return `M ${from.x} ${from.y} C ${c1x} ${c1y}, ${c2x} ${c2y}, ${to.x} ${to.y}`
}

export function FlowLine({
  from, to, color, style = 'solid', active,
  movingDot = false, arrow = true, curve = 0.5, width = 2,
}: Props) {
  // Path length state (computed after first render via ref measurement).
  const [pathLen, setPathLen] = useState(0)
  const pathRef = useMemo(() => ({ current: null as SVGPathElement | null }), [])

  // Build the path string up-front so the moving-dot can reference it.
  const d = useMemo(() => {
    if (!from || !to) return ''
    return bezierPath(from, to, curve)
  }, [from, to, curve])

  // Measure path length after each path change so the dash animation
  // covers the entire stroke.
  useEffect(() => {
    if (pathRef.current) {
      try { setPathLen(pathRef.current.getTotalLength()) } catch { /* ignore */ }
    }
  }, [d, pathRef])

  if (!from || !to) return null

  const stroke = COLOR_MAP[color]
  const markerEnd = arrow ? MARKER_MAP[color] : undefined
  const dashArray = style === 'dashed' ? '6 6' : undefined

  // When inactive, render nothing (cleaner than render-with-opacity-0
  // which can flash). When active, render the path with stroke-dashoffset
  // animating from pathLen → 0 over LINE_DRAW_MS.
  if (!active) return null

  return (
    <g>
      <path
        ref={(el) => { pathRef.current = el }}
        d={d}
        stroke={stroke}
        strokeWidth={width}
        fill="none"
        strokeDasharray={dashArray ?? (pathLen ? `${pathLen}` : undefined)}
        strokeDashoffset={dashArray ? 0 : pathLen}
        markerEnd={markerEnd}
        style={
          dashArray
            ? {
                animation: `flow-dash-shift ${LINE_DRAW_MS * 2}ms linear infinite`,
                opacity: 0.9,
              }
            : {
                transition: `stroke-dashoffset ${LINE_DRAW_MS}ms ease-out`,
                strokeDashoffset: 0,
                opacity: 0.95,
              }
        }
      />
      {movingDot && (
        <circle r={4} fill={stroke}>
          <animateMotion dur={`${LINE_DRAW_MS * 2}ms`} repeatCount="indefinite" path={d} />
        </circle>
      )}
    </g>
  )
}
