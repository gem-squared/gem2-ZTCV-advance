// Full-viewport SVG overlay layer. Sits ABOVE the 3 panels visually
// but pointer-events: none so clicks pass through to the underlying
// panels and scenario picker.
//
// The viewBox is sized to the current window so coordinates returned
// by `useAnchors()` (which are in viewport CSS pixels) map 1:1.

import { useEffect, useState, type ReactNode } from 'react'

interface Props {
  children: ReactNode
}

export function SvgOverlay({ children }: Props) {
  const [dims, setDims] = useState(() => ({
    w: typeof window !== 'undefined' ? window.innerWidth  : 1920,
    h: typeof window !== 'undefined' ? window.innerHeight : 1080,
  }))

  useEffect(() => {
    const onResize = () =>
      setDims({ w: window.innerWidth, h: window.innerHeight })
    window.addEventListener('resize', onResize)
    return () => window.removeEventListener('resize', onResize)
  }, [])

  return (
    <svg
      width={dims.w}
      height={dims.h}
      viewBox={`0 0 ${dims.w} ${dims.h}`}
      style={{
        position: 'fixed',
        inset: 0,
        width: '100vw',
        height: '100vh',
        pointerEvents: 'none',
        zIndex: 30,
      }}
      aria-hidden="true"
    >
      <defs>
        {/* Reusable arrowhead markers for different colored lines. */}
        <marker
          id="arrow-white" viewBox="0 0 10 10" refX="9" refY="5"
          markerWidth="6" markerHeight="6" orient="auto-start-reverse"
        >
          <path d="M0 0 L10 5 L0 10 z" fill="rgba(255,255,255,0.85)" />
        </marker>
        <marker
          id="arrow-blue" viewBox="0 0 10 10" refX="9" refY="5"
          markerWidth="6" markerHeight="6" orient="auto-start-reverse"
        >
          <path d="M0 0 L10 5 L0 10 z" fill="#3b82f6" />
        </marker>
        <marker
          id="arrow-fuchsia" viewBox="0 0 10 10" refX="9" refY="5"
          markerWidth="6" markerHeight="6" orient="auto-start-reverse"
        >
          <path d="M0 0 L10 5 L0 10 z" fill="#d946ef" />
        </marker>
        <marker
          id="arrow-gold" viewBox="0 0 10 10" refX="9" refY="5"
          markerWidth="6" markerHeight="6" orient="auto-start-reverse"
        >
          <path d="M0 0 L10 5 L0 10 z" fill="#fbbf24" />
        </marker>
        <marker
          id="arrow-safe" viewBox="0 0 10 10" refX="9" refY="5"
          markerWidth="6" markerHeight="6" orient="auto-start-reverse"
        >
          <path d="M0 0 L10 5 L0 10 z" fill="#22c55e" />
        </marker>
        <marker
          id="arrow-block" viewBox="0 0 10 10" refX="9" refY="5"
          markerWidth="6" markerHeight="6" orient="auto-start-reverse"
        >
          <path d="M0 0 L10 5 L0 10 z" fill="#ef4444" />
        </marker>
      </defs>
      {children}
    </svg>
  )
}
