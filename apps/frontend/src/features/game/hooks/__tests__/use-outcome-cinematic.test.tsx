import { act, fireEvent, render, screen } from '@testing-library/react-native'
import { Pressable, Text } from 'react-native'

import { useOutcomeCinematic } from '@/features/game/hooks/use-outcome-cinematic'
import type {
  GameParticipant,
  GameRound,
  GameSnapshot,
  GameTeam,
  GameViewer,
} from '@/features/game/types/game.types'

function snapshot(overrides: Partial<GameSnapshot> = {}): GameSnapshot {
  return {
    id: 'g1',
    code: 'ABC123',
    state: 'FINISHED',
    mode: 'GAUNTLET',
    hostId: 'p1',
    locked: false,
    config: {
      stageMangas: ['ONE_PIECE'],
      powerMangas: ['ONE_PIECE'],
      abilitySource: 'RANDOM',
      teamSize: 4,
      allowBots: false,
      visibility: 'PRIVATE',
      votingWindowSeconds: 30,
      revealSpeed: 'NORMAL',
      summaryDurationSeconds: 60,
      poolFilter: { standRarities: [], fruitRarities: [], fruitTypes: [], banned: [] },
    },
    teams: [],
    participants: [],
    rounds: [],
    ...overrides,
  }
}

function viewer(overrides: Partial<GameViewer> = {}): GameViewer {
  return { participantId: 'p1', teamId: 't1', isHost: false, hasVoted: false, ...overrides }
}

function Harness({ snapshot: snap, you }: { snapshot: GameSnapshot; you: GameViewer }) {
  const cinematic = useOutcomeCinematic(snap, you)
  return (
    <>
      <Text testID="visible">{String(cinematic.visible)}</Text>
      <Text testID="kind">{cinematic.kind ?? 'null'}</Text>
      <Text testID="canReplay">{String(cinematic.canReplay)}</Text>
      <Pressable testID="dismiss" onPress={cinematic.dismiss}>
        <Text>dismiss</Text>
      </Pressable>
      <Pressable testID="replay" onPress={cinematic.replay}>
        <Text>replay</Text>
      </Pressable>
    </>
  )
}

describe('useOutcomeCinematic', () => {
  it('GAUNTLET: shows victory when the squad survived', async () => {
    await render(
      <Harness
        snapshot={snapshot({
          mode: 'GAUNTLET',
          result: { mode: 'GAUNTLET', winner: 'SURVIVE', roundsPlayed: 3, aborted: false },
        })}
        you={viewer()}
      />
    )
    expect(screen.getByTestId('visible').props.children).toBe('true')
    expect(screen.getByTestId('kind').props.children).toBe('victory')
  })

  it('GAUNTLET: shows defeat when the squad fell', async () => {
    await render(
      <Harness
        snapshot={snapshot({
          mode: 'GAUNTLET',
          result: { mode: 'GAUNTLET', winner: 'FALL', roundsPlayed: 2, aborted: false },
        })}
        you={viewer()}
      />
    )
    expect(screen.getByTestId('kind').props.children).toBe('defeat')
  })

  it('VERSUS: shows victory/defeat per the viewer\'s own team', async () => {
    const teamA: GameTeam = { id: 'team-a', name: 'Straw Hats', color: 0, memberIds: [] }
    const teamB: GameTeam = { id: 'team-b', name: 'Baroque Works', color: 1, memberIds: [] }
    function participantOn(teamId: string): GameParticipant {
      return {
        id: 'p1',
        displayName: 'jotaro',
        teamId,
        kind: 'HUMAN',
        connected: true,
        abandoned: false,
        avatarThumb: '',
        avatarFocalX: 0.5,
        avatarFocalY: 0.5,
      }
    }
    function versusSnap(teamId: string) {
      return snapshot({
        mode: 'VERSUS',
        teams: [teamA, teamB],
        participants: [participantOn(teamId)],
        result: { mode: 'VERSUS', winner: 'team-a', roundsPlayed: 3, aborted: false },
      })
    }

    await render(<Harness snapshot={versusSnap('team-a')} you={viewer({ teamId: 'team-a' })} />)
    expect(screen.getByTestId('kind').props.children).toBe('victory')

    await render(<Harness snapshot={versusSnap('team-b')} you={viewer({ teamId: 'team-b' })} />)
    expect(screen.getAllByTestId('kind').at(-1)?.props.children).toBe('defeat')
  })

  it('never shows for an aborted game', async () => {
    await render(
      <Harness
        snapshot={snapshot({
          state: 'ABORTED',
          result: { mode: 'GAUNTLET', winner: '', roundsPlayed: 1, aborted: true },
        })}
        you={viewer()}
      />
    )
    expect(screen.getByTestId('visible').props.children).toBe('false')
    expect(screen.getByTestId('kind').props.children).toBe('null')
  })

  it('dismiss hides it and enables replay; replay shows it again', async () => {
    await render(
      <Harness
        snapshot={snapshot({
          result: { mode: 'GAUNTLET', winner: 'SURVIVE', roundsPlayed: 1, aborted: false },
        })}
        you={viewer()}
      />
    )
    expect(screen.getByTestId('visible').props.children).toBe('true')
    expect(screen.getByTestId('canReplay').props.children).toBe('false')

    await act(async () => fireEvent.press(screen.getByTestId('dismiss')))
    expect(screen.getByTestId('visible').props.children).toBe('false')
    expect(screen.getByTestId('canReplay').props.children).toBe('true')

    await act(async () => fireEvent.press(screen.getByTestId('replay')))
    expect(screen.getByTestId('visible').props.children).toBe('true')
    expect(screen.getByTestId('canReplay').props.children).toBe('false')
  })
})
