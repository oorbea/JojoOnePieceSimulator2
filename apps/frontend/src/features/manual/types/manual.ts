import type { ManualRules } from '@/shared/contracts/rules'

export type { ManualRules }

// Structural views over the generated rules (contracts/rules.ts is `as const`,
// so each entry is its own literal type; these are the common shapes the
// components take). Nothing here redeclares a backend enum: kinds, slots and
// levels stay plain wire strings, translated through the shared i18n keys.
export type ManualSubject = {
  readonly kind: string
  readonly slot?: string
  readonly min?: string
  readonly fruitTypes?: readonly string[]
  readonly names?: readonly string[]
  readonly group?: string
}

export type ManualConvention = {
  readonly id: string
  readonly category: string
  readonly actors: readonly ManualSubject[]
  readonly targets: readonly ManualSubject[]
  readonly immune: readonly ManualSubject[]
  readonly attenuated: readonly ManualSubject[]
  readonly gap: number
  readonly judgement: boolean
}

export type ManualPct = { readonly level: string; readonly percent: number }

export type ManualChain = {
  readonly kind: string
  readonly driver: string
  readonly stages: readonly { readonly name: string; readonly tier?: string }[]
  readonly levels: readonly string[]
  readonly rows: readonly {
    readonly drawn: string
    readonly outcomes: readonly {
      readonly level: string
      readonly power: string
      readonly levelAfter: string
      readonly change: string
    }[]
  }[]
}

export type ManualStatFloor = {
  readonly cross: boolean
  readonly causeSlot: string
  readonly when: ManualSubject
  readonly target: string
  readonly floor: string
}
