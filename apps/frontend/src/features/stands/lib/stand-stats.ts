import type { StandResponse } from '@/features/stands/types/stands.types'

// Shared between StandCard's stat grid and StandDetail's - both render the
// same six stats in the same order, just at different sizes.
export const STAND_STAT_LABELS: {
  key: keyof Pick<StandResponse, 'attackPower' | 'speed' | 'attackRange' | 'endurance' | 'precision' | 'potential'>
  label: string
}[] = [
  { key: 'attackPower', label: 'PWR' },
  { key: 'speed', label: 'SPD' },
  { key: 'attackRange', label: 'RNG' },
  { key: 'endurance', label: 'END' },
  { key: 'precision', label: 'PRE' },
  { key: 'potential', label: 'DEV' },
]
