import type { LoadoutSlot } from '@/shared/contracts/enums'
import { isLowestLevel, slotLevelKey, slotTraitKey } from '@/shared/lib/loadout-slots'

import type { ManualSubject } from '../types/manual'

type T = (key: string, options?: Record<string, unknown>) => string

// A convention's "who" as a short phrase: "Stand users", "Spin (any level)",
// "Spin Golden or higher", "Logia users". Power names are the catalogue's own
// display names, shown as stored.
export function subjectLabel(t: T, s: ManualSubject): string {
  switch (s.kind) {
    case 'ANY_STAND':
      return t('manual.subject.anyStand')
    case 'ANY_FRUIT':
      return t('manual.subject.anyFruit')
    case 'FRUIT_TYPE': {
      const types = (s.fruitTypes ?? []).map((ft) => t(`enums.fruitType.${ft}`)).join(' / ')
      return t('manual.subject.fruitType', { types })
    }
    case 'SLOT_MIN': {
      const slot = s.slot as LoadoutSlot
      const stat = t(slotTraitKey(slot))
      if (isLowestLevel(slot, s.min ?? '')) return t('manual.subject.slotAny', { stat })
      return t('manual.subject.slotMin', { stat, level: t(slotLevelKey(slot, s.min ?? '')) })
    }
    case 'POWERS': {
      const names = (s.names ?? []).join(', ')
      return s.group ? t(`manual.groups.${s.group}`, { names }) : names
    }
    default:
      return s.kind
  }
}
