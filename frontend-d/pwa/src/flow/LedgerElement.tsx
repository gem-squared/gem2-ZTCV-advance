// Ledger visual — three isometric cubes connected by two chain links,
// representing the on-chain receipt blockchain (per David's reference
// image 2026-05-30). When the chain-anchor phase fires, the entire
// row lights up gold and pulses; otherwise it sits in a muted/dim
// state to telegraph "not yet anchored".
//
// Renders a clickable Etherscan link below if a real receiptTxHash is
// available.

import { forwardRef } from 'react'

interface Props {
  receiptTxHash?: string
  explorerUrl?: string
  /** True after phase 6 fires — full row glows gold + pulse animation. */
  anchored: boolean
}

function truncateHash(h: string): string {
  if (h.length <= 16) return h
  return `${h.slice(0, 10)}…${h.slice(-6)}`
}

interface ChainCubeProps {
  active: boolean
  size?: number
}

/** Single isometric cube — 3 visible faces, outlined.
 *  Active = warm orange/gold; inactive = muted dark amber. */
function ChainCube({ active, size = 72 }: ChainCubeProps) {
  const topFill   = active ? '#fbbf24' : '#3f2e0a'
  const leftFill  = active ? '#d97706' : '#2a1f06'
  const rightFill = active ? '#f59e0b' : '#3a2a08'
  const stroke    = active ? '#1f2937' : '#6b7280'
  return (
    <svg
      viewBox="0 0 100 100"
      width={size}
      height={size}
      style={{ transition: 'all 320ms ease-out', filter: active ? 'drop-shadow(0 0 12px rgba(251,191,36,0.5))' : undefined }}
      aria-hidden="true"
    >
      {/* Top face (rhombus) */}
      <polygon
        points="50,12 88,32 50,52 12,32"
        fill={topFill}
        stroke={stroke}
        strokeWidth="3"
        strokeLinejoin="round"
      />
      {/* Left face (parallelogram) */}
      <polygon
        points="12,32 50,52 50,90 12,70"
        fill={leftFill}
        stroke={stroke}
        strokeWidth="3"
        strokeLinejoin="round"
      />
      {/* Right face (parallelogram) */}
      <polygon
        points="88,32 50,52 50,90 88,70"
        fill={rightFill}
        stroke={stroke}
        strokeWidth="3"
        strokeLinejoin="round"
      />
    </svg>
  )
}

/** Chain link between two cubes — two interlocked oval rings. */
function ChainLink({ active }: { active: boolean }) {
  const stroke = active ? '#fbbf24' : '#525252'
  return (
    <svg
      viewBox="0 0 60 28"
      width={44}
      height={20}
      style={{ transition: 'stroke 320ms ease-out' }}
      aria-hidden="true"
    >
      <ellipse
        cx="20"
        cy="14"
        rx="13"
        ry="8"
        fill="none"
        stroke={stroke}
        strokeWidth="2.6"
      />
      <ellipse
        cx="40"
        cy="14"
        rx="13"
        ry="8"
        fill="none"
        stroke={stroke}
        strokeWidth="2.6"
      />
    </svg>
  )
}

export const LedgerElement = forwardRef<HTMLDivElement, Props>(
  function LedgerElement({ receiptTxHash, explorerUrl, anchored }, ref) {
    const isMock = !!receiptTxHash && receiptTxHash.toUpperCase().startsWith('0XMOCK')

    return (
      <div className="flex flex-col items-center gap-2.5 relative">
        {/* Glow halo when anchored — sits behind the cubes for a soft pulse.
            Note: this is intentionally positioned RELATIVE to the cube row
            (not the outer flex column) so the halo wraps the cubes, not the
            label/link below. */}
        {anchored && (
          <div
            className="absolute pointer-events-none rounded-full"
            style={{
              top: '-20px',
              left: '50%',
              transform: 'translateX(-50%)',
              width: '320px',
              height: '120px',
              background:
                'radial-gradient(ellipse at center, rgba(251,191,36,0.22) 0%, rgba(251,191,36,0) 70%)',
              animation: 'ledger-pulse 1.4s ease-out 1',
            }}
          />
        )}

        {/* Three cubes joined by two chain links.
            ⊢ ref attaches HERE (not on the outer wrapper) so the
              `ledger.top` anchor lands exactly on the top center of the
              cube row — where the gold Chain-Receipt 앵커링 line should
              terminate. The label + Etherscan link below are layout-only
              and excluded from the anchor box. */}
        <div ref={ref} className="flex items-center gap-1.5 relative z-10">
          <ChainCube active={anchored} />
          <ChainLink active={anchored} />
          <ChainCube active={anchored} />
          <ChainLink active={anchored} />
          <ChainCube active={anchored} />
        </div>

        <div className="text-[10px] uppercase tracking-wider text-amber-300/80 font-semibold mt-1">
          Blockchain Receipt Ledger · Sepolia
        </div>
        {receiptTxHash && (
          isMock ? (
            <div className="font-mono text-[10px] text-zinc-400">
              {truncateHash(receiptTxHash)} <span className="text-zinc-600">· sim</span>
            </div>
          ) : (
            <a
              href={explorerUrl ?? '#'}
              target="_blank"
              rel="noopener noreferrer"
              className="font-mono text-[10px] text-amber-300 hover:text-amber-200 transition-colors underline decoration-amber-300/40 hover:decoration-amber-200"
            >
              {truncateHash(receiptTxHash)} · Etherscan ↗
            </a>
          )
        )}
      </div>
    )
  }
)
