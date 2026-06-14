import { useState } from 'react'

export const BYO_KEY_STORAGE = 'vouch_byo_llm_key'

interface Props {
  onClose: () => void
}

export function BYOKeyModal({ onClose }: Props) {
  const [key, setKey] = useState('')
  const [show, setShow] = useState(false)
  const [saved, setSaved] = useState(false)

  const handleSave = () => {
    const trimmed = key.trim()
    if (!trimmed) return
    sessionStorage.setItem(BYO_KEY_STORAGE, trimmed)
    setSaved(true)
    setTimeout(onClose, 600)
  }

  const handleSkip = () => {
    sessionStorage.removeItem(BYO_KEY_STORAGE)
    onClose()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm px-4">
      <div className="w-full max-w-md bg-zinc-900 border border-zinc-700 rounded-2xl p-6 shadow-2xl">
        <div className="flex items-center gap-3 mb-4">
          <span className="text-2xl">🤖</span>
          <div>
            <h2 className="text-base font-semibold text-zinc-100">Enable Live AI</h2>
            <p className="text-[11px] text-zinc-500">Bring Your Own Anthropic API Key</p>
          </div>
        </div>

        <p className="text-sm text-zinc-400 mb-4 leading-relaxed">
          VOUCH's <strong className="text-zinc-200">Intent Handshake</strong> (Step 6) uses Claude AI
          to generate pre-call disclosures. Enter your key to see live responses — otherwise
          the demo uses a deterministic fallback.
        </p>

        <p className="text-[11px] text-zinc-500 mb-4">
          Your key is sent only to this session's backend via HTTPS and is never stored server-side.
          Get a key at{' '}
          <a
            href="https://console.anthropic.com/settings/keys"
            target="_blank"
            rel="noopener noreferrer"
            className="text-fuchsia-400 underline"
          >
            console.anthropic.com
          </a>
          .
        </p>

        <div className="relative mb-5">
          <input
            type={show ? 'text' : 'password'}
            value={key}
            onChange={(e) => setKey(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSave()}
            placeholder="sk-ant-api03-…"
            className="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-3 py-2.5 text-sm text-zinc-100 placeholder-zinc-600 focus:outline-none focus:border-fuchsia-600 pr-16"
            autoFocus
          />
          <button
            onClick={() => setShow((v) => !v)}
            className="absolute right-2 top-1/2 -translate-y-1/2 text-[10px] text-zinc-500 hover:text-zinc-300 px-2 py-1"
          >
            {show ? 'hide' : 'show'}
          </button>
        </div>

        <div className="flex gap-3">
          <button
            onClick={handleSave}
            disabled={!key.trim() || saved}
            className="flex-1 py-2.5 rounded-lg bg-fuchsia-700 hover:bg-fuchsia-600 disabled:opacity-40 disabled:cursor-not-allowed text-white text-sm font-medium transition-colors"
          >
            {saved ? '✓ Saved' : 'Save & Enable Live AI'}
          </button>
          <button
            onClick={handleSkip}
            className="px-4 py-2.5 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-400 text-sm transition-colors"
          >
            Use Fallback
          </button>
        </div>
      </div>
    </div>
  )
}
