import { useRouter } from 'expo-router'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import type { BannableItem } from '@/features/game/components/presentational/fields/banlist-field'
import { CreateLobbyScreen } from '@/features/game/components/presentational/create-lobby-screen'
import { useCreateGame } from '@/features/game/hooks/use-create-game'
import type { PoolFilter } from '@/features/game/types/game.types'
import { toAppError } from '@/shared/api/errors'
import type { FruitType, GameMode, LobbyVisibility, Manga, Rarity } from '@/shared/lib/zod'
import { useDevilFruits } from '@/features/devil-fruits'
import { useStands } from '@/features/stands'

const GAUNTLET_MIN = 1
const GAUNTLET_MAX = 10
const VERSUS_MIN = 1
const VERSUS_MAX = 5

const EMPTY_POOL_FILTER: PoolFilter = {
  standRarities: [],
  fruitRarities: [],
  fruitTypes: [],
  banned: [],
}

export function CreateLobbyContainer() {
  const router = useRouter()
  const { t } = useTranslation()
  const createGame = useCreateGame()
  const standsQuery = useStands()
  const devilFruitsQuery = useDevilFruits()

  const [mode, setMode] = useState<GameMode>('GAUNTLET')
  const [mangas, setMangas] = useState<Manga[]>(['JOJO'])
  const [teamSize, setTeamSize] = useState(5)
  const [allowBots, setAllowBots] = useState(false)
  const [visibility, setVisibility] = useState<LobbyVisibility>('PRIVATE')
  const [votingWindowSeconds, setVotingWindowSeconds] = useState(30)
  const [poolFilter, setPoolFilter] = useState<PoolFilter>(EMPTY_POOL_FILTER)

  const banlistItems: BannableItem[] = useMemo(
    () => [
      ...(standsQuery.data ?? []).map((s) => ({ id: s.id, name: s.name, kind: 'STAND' as const })),
      ...(devilFruitsQuery.data ?? []).map((f) => ({ id: f.id, name: f.name, kind: 'DEVIL_FRUIT' as const })),
    ],
    [standsQuery.data, devilFruitsQuery.data]
  )

  const handleChangeMode = (next: GameMode) => {
    setMode(next)
    if (next === 'GAUNTLET') {
      setAllowBots(false)
      setTeamSize((size) => Math.min(Math.max(size, GAUNTLET_MIN), GAUNTLET_MAX))
    } else {
      setTeamSize((size) => Math.min(Math.max(size, VERSUS_MIN), VERSUS_MAX))
    }
  }

  const handleToggleManga = (manga: Manga) => {
    setMangas((current) =>
      current.includes(manga) ? current.filter((m) => m !== manga) : [...current, manga]
    )
  }

  const handleToggleRarity = (rarity: Rarity) => {
    setPoolFilter((current) => {
      const has = current.standRarities.includes(rarity)
      return {
        ...current,
        standRarities: has
          ? current.standRarities.filter((r) => r !== rarity)
          : [...current.standRarities, rarity],
        fruitRarities: has
          ? current.fruitRarities.filter((r) => r !== rarity)
          : [...current.fruitRarities, rarity],
      }
    })
  }

  const handleToggleFruitType = (fruitType: FruitType) => {
    setPoolFilter((current) => ({
      ...current,
      fruitTypes: current.fruitTypes.includes(fruitType)
        ? current.fruitTypes.filter((f) => f !== fruitType)
        : [...current.fruitTypes, fruitType],
    }))
  }

  const handleAddBan = (id: string) => {
    setPoolFilter((current) =>
      current.banned.includes(id) ? current : { ...current, banned: [...current.banned, id] }
    )
  }

  const handleRemoveBan = (id: string) => {
    setPoolFilter((current) => ({ ...current, banned: current.banned.filter((b) => b !== id) }))
  }

  const poolActiveCount =
    poolFilter.standRarities.length +
    poolFilter.fruitRarities.length +
    poolFilter.fruitTypes.length +
    poolFilter.banned.length

  const handleSubmit = () => {
    createGame.mutate(
      {
        mode,
        mangas,
        abilitySource: 'RANDOM',
        teamSize,
        allowBots,
        visibility,
        votingWindowSeconds,
        poolFilter,
      },
      {
        onSuccess: (data) => router.replace(`/play/${data.game.id}` as never),
      }
    )
  }

  return (
    <CreateLobbyScreen
      onBack={() => router.back()}
      mode={mode}
      onChangeMode={handleChangeMode}
      mangas={mangas}
      onToggleManga={handleToggleManga}
      teamSize={teamSize}
      teamSizeMin={mode === 'GAUNTLET' ? GAUNTLET_MIN : VERSUS_MIN}
      teamSizeMax={mode === 'GAUNTLET' ? GAUNTLET_MAX : VERSUS_MAX}
      onChangeTeamSize={setTeamSize}
      allowBots={allowBots}
      onToggleAllowBots={() => setAllowBots((v) => !v)}
      visibility={visibility}
      onToggleVisibility={() => setVisibility((v) => (v === 'PUBLIC' ? 'PRIVATE' : 'PUBLIC'))}
      votingWindowSeconds={votingWindowSeconds}
      onChangeVotingWindow={setVotingWindowSeconds}
      poolFilter={poolFilter}
      poolActiveCount={poolActiveCount}
      banlistItems={banlistItems}
      onToggleRarity={handleToggleRarity}
      onToggleFruitType={handleToggleFruitType}
      onAddBan={handleAddBan}
      onRemoveBan={handleRemoveBan}
      onClearPoolFilter={() => setPoolFilter(EMPTY_POOL_FILTER)}
      submitting={createGame.isPending}
      error={createGame.error ? t(`errors.${toAppError(createGame.error).code}`, { defaultValue: toAppError(createGame.error).message }) : undefined}
      onSubmit={handleSubmit}
    />
  )
}
