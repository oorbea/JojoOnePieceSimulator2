import { useRouter } from 'expo-router'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import type { BannableItem } from '@/features/game/components/presentational/fields/banlist-field'
import { CreateLobbyScreen } from '@/features/game/components/presentational/create-lobby-screen'
import { useCreateGame } from '@/features/game/hooks/use-create-game'
import { TEAM_SIZE_LIMITS, clampTeamSize } from '@/features/game/lib/config-form'
import type { PoolFilter } from '@/features/game/types/game.types'
import { toAppError } from '@/shared/api/errors'
import type { GameMode, LobbyVisibility, Manga } from '@/shared/contracts/enums'
import { useDevilFruits } from '@/features/devil-fruits'
import { useStands } from '@/features/stands'

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
  // Both mangas on by default, on both axes - a new host most likely wants
  // the full pool, not a JoJo-only lobby by accident just because it's
  // first in the toggle row.
  const [stageMangas, setStageMangas] = useState<Manga[]>(['JOJO', 'ONE_PIECE'])
  const [powerMangas, setPowerMangas] = useState<Manga[]>(['JOJO', 'ONE_PIECE'])
  const [teamSize, setTeamSize] = useState(5)
  const [allowBots, setAllowBots] = useState(false)
  const [visibility, setVisibility] = useState<LobbyVisibility>('PRIVATE')
  const [votingWindowSeconds, setVotingWindowSeconds] = useState(30)
  const [summaryDurationSeconds, setSummaryDurationSeconds] = useState(60)
  const [poolFilter, setPoolFilter] = useState<PoolFilter>(EMPTY_POOL_FILTER)

  const banlistItems: BannableItem[] = useMemo(
    () => [
      ...(standsQuery.data ?? []).map((s) => ({
        id: s.id,
        name: s.name,
        kind: 'STAND' as const,
        rarity: s.rarity,
        stats: {
          attackPower: s.attackPower,
          speed: s.speed,
          attackRange: s.attackRange,
          endurance: s.endurance,
          precision: s.precision,
          potential: s.potential,
        },
      })),
      ...(devilFruitsQuery.data ?? []).map((f) => ({
        id: f.id,
        name: f.name,
        kind: 'DEVIL_FRUIT' as const,
        rarity: f.rarity,
        fruitType: f.fruitType,
      })),
    ],
    [standsQuery.data, devilFruitsQuery.data]
  )

  const handleChangeMode = (next: GameMode) => {
    setMode(next)
    if (next === 'GAUNTLET') setAllowBots(false)
    setTeamSize((size) => clampTeamSize(next, size))
  }

  const handleToggleStageManga = (manga: Manga) => {
    setStageMangas((current) =>
      current.includes(manga) ? current.filter((m) => m !== manga) : [...current, manga]
    )
  }

  const handleTogglePowerManga = (manga: Manga) => {
    setPowerMangas((current) =>
      current.includes(manga) ? current.filter((m) => m !== manga) : [...current, manga]
    )
  }

  const handleAddBan = (id: string) => {
    setPoolFilter((current) =>
      current.banned.includes(id) ? current : { ...current, banned: [...current.banned, id] }
    )
  }

  const handleRemoveBan = (id: string) => {
    setPoolFilter((current) => ({ ...current, banned: current.banned.filter((b) => b !== id) }))
  }

  const handleBanMatching = (ids: string[]) => {
    setPoolFilter((current) => ({
      ...current,
      banned: [...current.banned, ...ids.filter((id) => !current.banned.includes(id))],
    }))
  }

  // Whitelisting by rarity/fruit-type isn't exposed in the UI (owner call -
  // only banning specific powers is needed) even though `PoolFilter` still
  // carries those fields for the backend, so the active-count badge only
  // ever reflects the banlist.
  const poolActiveCount = poolFilter.banned.length

  const handleSubmit = () => {
    createGame.mutate(
      {
        mode,
        stageMangas,
        powerMangas,
        abilitySource: 'RANDOM',
        teamSize,
        allowBots,
        visibility,
        votingWindowSeconds,
        summaryDurationSeconds,
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
      stageMangas={stageMangas}
      powerMangas={powerMangas}
      onToggleStageManga={handleToggleStageManga}
      onTogglePowerManga={handleTogglePowerManga}
      teamSize={teamSize}
      teamSizeMin={TEAM_SIZE_LIMITS[mode].min}
      teamSizeMax={TEAM_SIZE_LIMITS[mode].max}
      onChangeTeamSize={setTeamSize}
      allowBots={allowBots}
      onToggleAllowBots={() => setAllowBots((v) => !v)}
      visibility={visibility}
      onToggleVisibility={() => setVisibility((v) => (v === 'PUBLIC' ? 'PRIVATE' : 'PUBLIC'))}
      votingWindowSeconds={votingWindowSeconds}
      onChangeVotingWindow={setVotingWindowSeconds}
      summaryDurationSeconds={summaryDurationSeconds}
      onChangeSummaryDuration={setSummaryDurationSeconds}
      poolFilter={poolFilter}
      poolActiveCount={poolActiveCount}
      banlistItems={banlistItems}
      onAddBan={handleAddBan}
      onRemoveBan={handleRemoveBan}
      onBanMatching={handleBanMatching}
      onClearPoolFilter={() => setPoolFilter(EMPTY_POOL_FILTER)}
      submitting={createGame.isPending}
      error={createGame.error ? t(`errors.${toAppError(createGame.error).code}`, { defaultValue: toAppError(createGame.error).message }) : undefined}
      onSubmit={handleSubmit}
    />
  )
}
