export type ScenarioId = 1 | 2 | 3

export type BadgeTone = 'safe' | 'block' | 'warn'

export interface ScenarioMeta {
  id: ScenarioId
  title: string
  subtitle: string
  badge: string
  badgeTone: BadgeTone
  callerOrg: string
  callerLabel: string
  callerPhone: string
  callerBadge: 'AI' | '상담원' | '수사관'
}

export const SCENARIOS: Record<ScenarioId, ScenarioMeta> = {
  1: {
    id: 1,
    title: '카카오뱅크 AI 상담',
    subtitle: '대출 상담 전화',
    badge: '안전',
    badgeTone: 'safe',
    callerOrg: '카카오뱅크',
    callerLabel: '카카오뱅크 AI 상담원',
    callerPhone: '02-3456-7890',
    callerBadge: 'AI'
  },
  2: {
    id: 2,
    title: '검찰청 사칭 전화',
    subtitle: '송금 요구',
    badge: '차단',
    badgeTone: 'block',
    callerOrg: '대한민국 검찰청',
    callerLabel: '검찰청 수사관',
    callerPhone: '02-0000-0000',
    callerBadge: '수사관'
  },
  3: {
    id: 3,
    title: '카카오뱅크 보안 알림 AI',
    subtitle: '대출 가입 유도 (권한 외 목적)',
    badge: '차단',
    badgeTone: 'warn',
    callerOrg: '카카오뱅크',
    callerLabel: '카카오뱅크 보안 알림 AI',
    callerPhone: '02-3456-7891',
    callerBadge: 'AI'
  }
}

export const SCENARIO_LIST: ScenarioMeta[] = [
  SCENARIOS[1],
  SCENARIOS[2],
  SCENARIOS[3]
]
