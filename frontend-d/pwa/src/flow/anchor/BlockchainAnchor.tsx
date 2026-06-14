// BlockchainAnchor — visual "verification settled on chain" status card.
//
// Ported from feature/dramatic-desktop-upgrade Next.js branch and adapted
// for our Vite tree. Placed below LedgerElement in /flow/ column 2 — adds
// network badge + animated confirmations counter + metrics table.
//
// DELIBERATELY OMITS tx hash display + Etherscan link to avoid duplicating
// the affordance already provided by LedgerElement.
//
// HONESTY: the confirmations counter is a local visual animation (0→12 over
// ~1.8s) — NOT polled from Sepolia. The caption at the bottom makes this
// explicit so we never claim "live chain polling" when we don't do it.
import { useEffect, useState } from 'react'

interface BlockchainAnchorProps {
  active: boolean
  txHash?: string
}

export function BlockchainAnchor({ active, txHash }: BlockchainAnchorProps) {
  const [confirmations, setConfirmations] = useState(0)
  const isMock = !!txHash && txHash.toUpperCase().startsWith('0XMOCK')

  useEffect(() => {
    if (!active) { setConfirmations(0); return }
    if (isMock) { setConfirmations(12); return }
    setConfirmations(0)
    const interval = setInterval(() => {
      setConfirmations((c) => {
        if (c >= 12) { clearInterval(interval); return 12 }
        return c + 1
      })
    }, 150)
    return () => clearInterval(interval)
  }, [active, isMock, txHash])

  if (!active) return null

  return (
    <div
      className="w-full max-w-[260px] flex flex-col items-center px-3 py-3 bg-zinc-950/60 border border-zinc-800/60 rounded-lg"
      style={{ animation: 'anchor-card-in 360ms ease-out 1' }}
    >
      <div className="inline-flex items-center gap-1.5 bg-zinc-900/80 border border-zinc-700/50 rounded-full px-2.5 py-1 mb-3">
        <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
        <span className="text-[9px] text-zinc-300 font-mono">Sepolia · Chain 11155111</span>
      </div>

      <div className="bg-emerald-950/30 border border-emerald-500/30 rounded-lg px-3 py-2 mb-3 w-full text-center">
        <div className="text-[9px] text-emerald-400/70 uppercase tracking-wider mb-0.5">Confirmations</div>
        <div className="text-lg font-mono font-bold text-emerald-300">{confirmations}</div>
      </div>

      <div className="w-full bg-zinc-900/40 border border-zinc-800/40 rounded-lg p-2.5 mb-2 space-y-1.5 text-[10px]">
        <div className="flex justify-between"><span className="text-zinc-500">상태</span><span className="text-emerald-300">Success</span></div>
        <div className="flex justify-between"><span className="text-zinc-500">네트워크</span><span className="text-zinc-300">Sepolia Testnet</span></div>
        <div className="flex justify-between"><span className="text-zinc-500">컨트랙트</span><span className="text-zinc-300 font-mono">ZTCVReceiptAnchor</span></div>
      </div>

      <p className="text-[9px] text-zinc-600 text-center leading-snug">
        ※ confirmations는 로컬 시각화 (실시간 체인 폴링 아님)
      </p>
    </div>
  )
}
