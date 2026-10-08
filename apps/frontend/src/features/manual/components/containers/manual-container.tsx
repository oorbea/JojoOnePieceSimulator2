import { useLocalSearchParams, useRouter } from 'expo-router'
import { useCallback, useRef, useState } from 'react'
import type { ScrollView } from 'react-native'

import { MANUAL_RULES } from '@/shared/contracts/rules'

import { parseSection, type ManualSectionId } from '../../lib/sections'
import { ManualScreen } from '../presentational/manual-screen'

// How long to wait after a section reports its layout before scrolling to it
// for a deep link: the row/column offsets above it report in the same layout
// pass, in no guaranteed order.
const DEEP_LINK_SETTLE_MS = 120

type Offsets = {
  row: number
  column: number
  sections: Partial<Record<ManualSectionId, number>>
}

// Wires the manual to the router: `?section=<id>` deep-links to a section
// (the lobby, vote and reveal screens link here), and the table of contents
// scrolls to one and keeps the URL in step so a section can be shared.
export function ManualContainer() {
  const params = useLocalSearchParams<{ section?: string }>()
  const router = useRouter()
  const requested = parseSection(params.section)

  const scrollRef = useRef<ScrollView>(null)
  const offsets = useRef<Offsets>({ row: 0, column: 0, sections: {} })
  const pending = useRef<ManualSectionId | null>(requested)
  const [active, setActive] = useState<ManualSectionId | null>(requested)

  const scrollToSection = useCallback((id: ManualSectionId) => {
    const y = offsets.current.sections[id]
    if (y === undefined) return
    scrollRef.current?.scrollTo({
      y: offsets.current.row + offsets.current.column + y,
      animated: true,
    })
  }, [])

  const handleMeasureSection = useCallback(
    (id: ManualSectionId, y: number) => {
      offsets.current.sections[id] = y
      if (pending.current === id) {
        pending.current = null
        setTimeout(() => scrollToSection(id), DEEP_LINK_SETTLE_MS)
      }
    },
    [scrollToSection]
  )

  const handleSelect = useCallback(
    (id: ManualSectionId) => {
      setActive(id)
      scrollToSection(id)
      router.setParams({ section: id })
    },
    [router, scrollToSection]
  )

  return (
    <ManualScreen
      rules={MANUAL_RULES}
      active={active}
      onSelectSection={handleSelect}
      onMeasureRow={(y) => {
        offsets.current.row = y
      }}
      onMeasureColumn={(y) => {
        offsets.current.column = y
      }}
      onMeasureSection={handleMeasureSection}
      scrollRef={scrollRef}
    />
  )
}
