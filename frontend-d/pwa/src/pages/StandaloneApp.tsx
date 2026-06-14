import { useState, useCallback } from 'react'
import { SplashScreen } from '../screens/SplashScreen'
import { PickerScreen } from '../screens/PickerScreen'
import { RingScreen } from '../screens/RingScreen'
import { BrokerVerificationScreen } from '../screens/BrokerVerificationScreen'
import { PassportScreen } from '../screens/PassportScreen'
import type { ScenarioId } from '../scenarios'
import type { RunScenarioResult } from '../api'

type View =
  | { kind: 'splash' }
  | { kind: 'picker' }
  | { kind: 'ring'; scenarioId: ScenarioId }
  | { kind: 'broker'; scenarioId: ScenarioId; result: RunScenarioResult }
  | { kind: 'passport'; scenarioId: ScenarioId; result: RunScenarioResult }

/**
 * StandaloneApp — the original Primary Receiver PWA flow.
 * Splash → Picker → Ring → BrokerVerification → Passport → Reset.
 * Unchanged from prior implementation. Used at `/` and `/receiver`.
 */
export function StandaloneApp() {
  const [view, setView] = useState<View>({ kind: 'splash' })

  const handleStart = useCallback(() => setView({ kind: 'picker' }), [])
  const handlePick = useCallback(
    (n: ScenarioId) => setView({ kind: 'ring', scenarioId: n }),
    []
  )
  const handleRingComplete = useCallback(
    (scenarioId: ScenarioId, result: RunScenarioResult) =>
      setView({ kind: 'broker', scenarioId, result }),
    []
  )
  const handleBrokerComplete = useCallback(() => {
    setView(prev => {
      if (prev.kind === 'broker') {
        return { kind: 'passport', scenarioId: prev.scenarioId, result: prev.result }
      }
      return prev
    })
  }, [])
  const handleReset = useCallback(() => setView({ kind: 'picker' }), [])

  return (
    <>
      {view.kind === 'splash' && <SplashScreen onStart={handleStart} />}
      {view.kind === 'picker' && <PickerScreen onPick={handlePick} />}
      {view.kind === 'ring' && (
        <RingScreen scenarioId={view.scenarioId} onComplete={handleRingComplete} />
      )}
      {view.kind === 'broker' && (
        <BrokerVerificationScreen
          scenarioId={view.scenarioId}
          result={view.result}
          onComplete={handleBrokerComplete}
        />
      )}
      {view.kind === 'passport' && (
        <PassportScreen
          scenarioId={view.scenarioId}
          result={view.result}
          onReset={handleReset}
        />
      )}
    </>
  )
}
