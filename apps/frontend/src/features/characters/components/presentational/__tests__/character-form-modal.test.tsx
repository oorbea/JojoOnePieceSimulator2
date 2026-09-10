import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'

import { act, fireEvent, renderWithProviders, screen } from '@/test/render'
import { createEmptyCharacterTranslationsForm } from '@/shared/lib/character-translations'
import type { Locale } from '@/shared/contracts/enums'
import {
  jojoCharacterFormSchema,
  onePieceCharacterFormSchema,
  type CharacterKind,
  type JojoCharacterFormValues,
  type OnePieceCharacterFormValues,
} from '@/features/characters/types/characters.types'

import { CharacterFormModal } from '../character-form-modal'

function createJojoDefaultValues(): JojoCharacterFormValues {
  return {
    name: '',
    rarity: 'COMMON',
    hamon: 'NONE',
    spin: 'NONE',
    battleIq: 100,
    translations: createEmptyCharacterTranslationsForm(),
  }
}

function createOnePieceDefaultValues(): OnePieceCharacterFormValues {
  return {
    name: '',
    rarity: 'COMMON',
    physicalForm: 'PRIVATE',
    armamentHaki: 'NONE',
    observationHaki: 'NONE',
    conquerorHaki: 'NONE',
    fruitMastery: 'NONE',
    translations: createEmptyCharacterTranslationsForm(),
  }
}

type HarnessProps = {
  mode?: 'create' | 'edit'
  initialKind?: CharacterKind | null
  onSubmit: () => void
  onCancel: () => void
}

function Harness({ mode = 'create', initialKind = null, onSubmit, onCancel }: HarnessProps) {
  const jojo = useForm<JojoCharacterFormValues>({
    resolver: zodResolver(jojoCharacterFormSchema),
    defaultValues: createJojoDefaultValues(),
  })
  const onePiece = useForm<OnePieceCharacterFormValues>({
    resolver: zodResolver(onePieceCharacterFormSchema),
    defaultValues: createOnePieceDefaultValues(),
  })
  const [kind, setKind] = useState<CharacterKind | null>(initialKind)
  const [activeLocale, setActiveLocale] = useState<Locale>('en-GB')

  return (
    <CharacterFormModal
      visible
      mode={mode}
      kind={kind}
      onSelectKind={setKind}
      jojoControl={jojo.control}
      jojoErrors={jojo.formState.errors}
      onePieceControl={onePiece.control}
      onePieceErrors={onePiece.formState.errors}
      onSubmit={onSubmit}
      onCancel={onCancel}
      isSaving={false}
      pictureUri={null}
      onPickPicture={jest.fn()}
      isPictureBusy={false}
      activeLocale={activeLocale}
      onLocaleChange={setActiveLocale}
      erroredLocales={[]}
    />
  )
}

describe('CharacterFormModal', () => {
  it('asks for the manga first and hides every other field until one is picked', async () => {
    await renderWithProviders(<Harness onSubmit={jest.fn()} onCancel={jest.fn()} />)

    expect(screen.getByText('Manga')).toBeTruthy()
    expect(screen.queryByLabelText('Name')).toBeNull()
    expect(screen.queryByText('Hamon')).toBeNull()
    expect(screen.queryByText('Physical Form')).toBeNull()
  })

  it('shows only the JoJo stat fields once JOJO is picked', async () => {
    await renderWithProviders(
      <Harness initialKind="JOJO" onSubmit={jest.fn()} onCancel={jest.fn()} />
    )

    expect(screen.getByLabelText('Name')).toBeTruthy()
    expect(screen.getByText('Hamon')).toBeTruthy()
    expect(screen.getByText('Spin')).toBeTruthy()
    expect(screen.getByLabelText('Battle IQ')).toBeTruthy()
    expect(screen.queryByText('Physical Form')).toBeNull()
    expect(screen.queryByText('Armament Haki')).toBeNull()
  })

  it('shows only the One Piece stat fields once ONE_PIECE is picked', async () => {
    await renderWithProviders(
      <Harness initialKind="ONE_PIECE" onSubmit={jest.fn()} onCancel={jest.fn()} />
    )

    expect(screen.getByLabelText('Name')).toBeTruthy()
    expect(screen.getByText('Physical Form')).toBeTruthy()
    expect(screen.getByText('Armament Haki')).toBeTruthy()
    expect(screen.getByText('Observation Haki')).toBeTruthy()
    expect(screen.getByText("Conqueror's Haki")).toBeTruthy()
    expect(screen.getByText('Fruit Mastery')).toBeTruthy()
    expect(screen.queryByText('Hamon')).toBeNull()
    expect(screen.queryByLabelText('Battle IQ')).toBeNull()
  })

  it('locks the manga to a read-only panel in edit mode', async () => {
    await renderWithProviders(
      <Harness mode="edit" initialKind="JOJO" onSubmit={jest.fn()} onCancel={jest.fn()} />
    )

    expect(screen.queryByText('One Piece')).toBeNull()
    expect(screen.getByText("JoJo's Bizarre Adventure")).toBeTruthy()
    // No select control for manga in edit mode - just the resolved label.
    expect(screen.queryByLabelText('Manga')).toBeNull()
  })

  it('disables Save until a manga is picked', async () => {
    await renderWithProviders(<Harness onSubmit={jest.fn()} onCancel={jest.fn()} />)

    // Disabled Tamagui Buttons render `aria-disabled` (not RN's
    // accessibilityState) and RNTL's label queries exclude inert elements -
    // query by the still-visible text and check its disabled ancestor
    // directly, same as confirm-sheet-confirming.test.tsx.
    expect(screen.getByText('Save').parent?.props['aria-disabled']).toBe(true)
  })

  it('fires onCancel and onSubmit from their own buttons once a manga is picked', async () => {
    const onSubmit = jest.fn()
    const onCancel = jest.fn()
    await renderWithProviders(
      <Harness initialKind="JOJO" onSubmit={onSubmit} onCancel={onCancel} />
    )

    await act(async () => {
      fireEvent.press(screen.getByLabelText('Cancel'))
    })
    expect(onCancel).toHaveBeenCalledTimes(1)

    await act(async () => {
      fireEvent.press(screen.getByLabelText('Save Character'))
    })
    expect(onSubmit).toHaveBeenCalledTimes(1)
  })
})
