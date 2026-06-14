import { Component, useEffect, useState, type ErrorInfo, type ReactNode } from 'react'
import { StandaloneApp } from './pages/StandaloneApp'
import { DramaticPage } from './pages/DramaticPage'
import { FlowPage } from './pages/FlowPage'
import { SignupPage } from './pages/SignupPage'
import { CallerPanel } from './panels/CallerPanel'
import { BrokerPanel } from './panels/BrokerPanel'

interface EBState {
  hasError: boolean
}

class ErrorBoundary extends Component<{ children: ReactNode }, EBState> {
  state: EBState = { hasError: false }

  static getDerivedStateFromError(): EBState {
    return { hasError: true }
  }

  componentDidCatch(error: Error, info: ErrorInfo): void {
    console.error('[ZTCV ErrorBoundary]', error, info)
  }

  reset = (): void => this.setState({ hasError: false })

  render(): ReactNode {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen bg-bg-base text-zinc-100 flex flex-col items-center justify-center px-6">
          <div className="text-5xl mb-4">⚠️</div>
          <h2 className="text-xl mb-2">잠시 문제가 발생했습니다</h2>
          <p className="text-sm text-zinc-400 mb-8">다시 시도해 주세요</p>
          <button
            onClick={this.reset}
            className="px-8 py-3 bg-bg-elevated rounded-full text-sm"
          >
            처음으로
          </button>
        </div>
      )
    }
    return this.props.children
  }
}

// Standalone Caller view — shows the CallerPanel with a static "submitted" state
// for screenshot capture / 3-window demo use.
function StandaloneCallerView() {
  return (
    <div className="min-h-screen bg-bg-base text-zinc-100 p-6">
      <div className="max-w-md mx-auto">
        <div className="text-[10px] text-zinc-500 uppercase tracking-wider mb-2">
          Standalone View · for 3-window demo or capture
        </div>
        <CallerPanel scenarioId={1} phase="submitted" />
      </div>
    </div>
  )
}

// Standalone Broker view — shows the BrokerPanel in a complete state for capture.
function StandaloneBrokerView() {
  return (
    <div className="min-h-screen bg-bg-base text-zinc-100 p-6">
      <div className="max-w-md mx-auto">
        <div className="text-[10px] text-zinc-500 uppercase tracking-wider mb-2">
          Standalone View · for 3-window demo or capture
        </div>
        <BrokerPanel scenarioId={1} phase="verifying" result={null} />
      </div>
    </div>
  )
}

// Subpath-aware: when deployed at /dramatic/ (Vite base), strip that prefix
// before matching routes, AND treat the bare base path as the dramatic landing
// so https://host/dramatic/ opens DramaticPage directly.
const BASE_RAW = import.meta.env.BASE_URL || '/'
const BASE = BASE_RAW.replace(/\/$/, '')  // '/dramatic' or ''
const IS_SUBPATH_DEPLOY = BASE === '/dramatic'

function stripBase(p: string): string {
  if (BASE && p.startsWith(BASE)) {
    const stripped = p.slice(BASE.length)
    return stripped || '/'
  }
  return p
}

function App() {
  // We track BOTH the base-stripped path (for /dramatic/ sub-routes)
  // AND the raw window pathname (so /flow/ — served by a sibling
  // Caddy handle but sharing the same SPA bundle — can be detected
  // without conflating with the /dramatic/ base).
  const [path, setPath] = useState<string>(() =>
    typeof window !== 'undefined' ? stripBase(window.location.pathname) : '/'
  )
  const [rawPath, setRawPath] = useState<string>(() =>
    typeof window !== 'undefined' ? window.location.pathname : '/'
  )

  useEffect(() => {
    const onPop = () => {
      setPath(stripBase(window.location.pathname))
      setRawPath(window.location.pathname)
    }
    window.addEventListener('popstate', onPop)
    return () => window.removeEventListener('popstate', onPop)
  }, [])

  let body: ReactNode
  // /flow/ — sibling top-level route served by Caddy handle_path /flow/*
  // (shares the same /var/www/ztcv-dramatic bundle). Detected via raw
  // pathname since the Vite base is still /dramatic/.
  if (rawPath.startsWith('/flow') || path === '/flow') body = <FlowPage />
  else if (rawPath.startsWith('/signup') || path === '/signup') body = <SignupPage />
  else if (path === '/dramatic' || (IS_SUBPATH_DEPLOY && path === '/')) body = <DramaticPage />
  else if (path === '/caller') body = <StandaloneCallerView />
  else if (path === '/broker') body = <StandaloneBrokerView />
  else body = <StandaloneApp />

  return <ErrorBoundary>{body}</ErrorBoundary>
}

export default App
