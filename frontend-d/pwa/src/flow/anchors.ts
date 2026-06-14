// Anchor-point computation for SVG overlay flow lines.
// Returns viewport-relative coordinates (the SvgOverlay's viewBox
// matches the document for a one-to-one mapping).
//
// Anchors are recomputed on:
//   - First paint (after refs mount)
//   - Window resize
//   - Window scroll (in case the layout shifts vertically with header changes)
//   - Manual trigger (scenario change re-mounts panels)

import { useCallback, useEffect, useState, type RefObject } from 'react'

export type Side = 'left' | 'right' | 'top' | 'bottom' | 'center'

export interface XY {
  x: number
  y: number
}

/**
 * Compute a single anchor point on an element's bounding box.
 * Coordinates are in viewport (CSS pixel) units relative to the viewport top-left.
 */
export function computeAnchor(el: Element | null, side: Side): XY | null {
  if (!el) return null
  const r = el.getBoundingClientRect()
  switch (side) {
    case 'left':   return { x: r.left,                 y: r.top + r.height / 2 }
    case 'right':  return { x: r.right,                y: r.top + r.height / 2 }
    case 'top':    return { x: r.left + r.width / 2,   y: r.top }
    case 'bottom': return { x: r.left + r.width / 2,   y: r.bottom }
    case 'center': return { x: r.left + r.width / 2,   y: r.top + r.height / 2 }
  }
}

export interface AnchorSpec {
  name: string
  ref: RefObject<HTMLElement | null>
  side: Side
}

export type AnchorMap = Record<string, XY | null>

/**
 * React hook that maintains a map of named anchor points and
 * recomputes on resize / scroll / version bump.
 *
 * `version` is a number you bump to force recompute (e.g. when a
 * scenario change re-mounts panels and you want to re-snap to
 * post-mount positions).
 */
export function useAnchors(specs: AnchorSpec[], version = 0): AnchorMap {
  const [map, setMap] = useState<AnchorMap>({})

  const recompute = useCallback(() => {
    const next: AnchorMap = {}
    for (const s of specs) next[s.name] = computeAnchor(s.ref.current, s.side)
    setMap(next)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [specs.length, version])

  useEffect(() => {
    // First paint
    recompute()
    // Multiple recomputes shortly after to catch late layout settling
    // (font load, image decode, transitions).
    const t1 = setTimeout(recompute, 80)
    const t2 = setTimeout(recompute, 240)
    const t3 = setTimeout(recompute, 600)

    window.addEventListener('resize', recompute)
    window.addEventListener('scroll', recompute, { passive: true })

    return () => {
      clearTimeout(t1)
      clearTimeout(t2)
      clearTimeout(t3)
      window.removeEventListener('resize', recompute)
      window.removeEventListener('scroll', recompute)
    }
  }, [recompute])

  return map
}
