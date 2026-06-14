import type { ScenariosRunResponse } from './types/passport'
import scenario1Fixture from './fixtures/scenario-1.json'
import scenario2Fixture from './fixtures/scenario-2.json'
import scenario3Fixture from './fixtures/scenario-3.json'

// BACKEND_URL: defaults to production sim-mode. Override via .env.local
// VITE_BACKEND_URL=http://localhost:8001 to point at our local Sepolia-anchoring
// session-svc for real-chain demo capture.
const BACKEND_URL =
  (import.meta.env.VITE_BACKEND_URL as string | undefined) ||
  'https://vouch.gemsquared.ai'
const FETCH_TIMEOUT_MS = 60000

const FIXTURES: Record<1 | 2 | 3, ScenariosRunResponse> = {
  1: scenario1Fixture as ScenariosRunResponse,
  2: scenario2Fixture as ScenariosRunResponse,
  3: scenario3Fixture as ScenariosRunResponse
}

export interface RunScenarioResult {
  data: ScenariosRunResponse
  source: 'live' | 'fixture'
}

// Broken demo is fatal — runScenario MUST always resolve with a usable result.
// Network/timeout/non-2xx all fall back to committed fixture silently; the
// returned `source` flag drives an "오프라인 모드" pill in the UI.
export async function runScenario(n: 1 | 2 | 3): Promise<RunScenarioResult> {
  const fixture = FIXTURES[n]
  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(), FETCH_TIMEOUT_MS)

  try {
    const res = await fetch(`${BACKEND_URL}/api/scenarios/run?n=${n}`, {
      method: 'POST',
      signal: controller.signal
    })
    clearTimeout(timer)
    if (!res.ok) {
      return { data: fixture, source: 'fixture' }
    }
    const data = (await res.json()) as ScenariosRunResponse
    if (!data || !data.passport || !Array.isArray(data.passport.stamps)) {
      return { data: fixture, source: 'fixture' }
    }
    return { data, source: 'live' }
  } catch {
    clearTimeout(timer)
    return { data: fixture, source: 'fixture' }
  }
}
