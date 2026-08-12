import { useRouter } from 'expo-router'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { CreateLobbyScreen } from '@/features/game/components/presentational/create-lobby-screen'
import { useCreateGame } from '@/features/game/hooks/use-create-game'
import { toAppError } from '@/shared/api/errors'
import type { GameMode, LobbyVisibility, Manga } from '@/shared/lib/zod'

const GAUNTLET_MIN = 1
const GAUNTLET_MAX = 10
const VERSUS_MIN = 1
const VERSUS_MAX = 5

export function CreateLobbyContainer() {
  const router = useRouter()
  const { t } = useTranslation()
  const createGame = useCreateGame()

  const [mode, setMode] = useState<GameMode>('GAUNTLET')
  const [mangas, setMangas] = useState<Manga[]>(['JOJO'])
  const [teamSize, setTeamSize] = useState(5)
  const [allowBots, setAllowBots] = useState(false)
  const [visibility, setVisibility] = useState<LobbyVisibility>('PRIVATE')
  const [votingWindowSeconds, setVotingWindowSeconds] = useState(30)

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
      submitting={createGame.isPending}
      error={createGame.error ? t(`errors.${toAppError(createGame.error).code}`, { defaultValue: toAppError(createGame.error).message }) : undefined}
      onSubmit={handleSubmit}
    />
  )
}
