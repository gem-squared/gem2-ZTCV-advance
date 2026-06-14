interface Props {
  onStart: () => void
}

export function SplashScreen({ onStart }: Props) {
  return (
    <div className="min-h-screen bg-bg-base text-zinc-100 flex flex-col items-center justify-center px-6">
      <div className="text-7xl mb-8 drop-shadow-[0_0_30px_rgba(139,92,246,0.5)]">🛡️</div>
      <h1 className="text-4xl font-semibold tracking-tight mb-3">VOUCH</h1>
      <p className="text-zinc-400 text-sm mb-16">통화 연결 전, 신원을 증명합니다</p>
      <button
        onClick={onStart}
        className="px-12 py-4 bg-bg-elevated hover:bg-bg-subtle active:bg-bg-subtle transition-colors rounded-full text-zinc-100 font-medium border border-zinc-800"
      >
        시작하기
      </button>
    </div>
  )
}
