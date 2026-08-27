import type { GameSnapshot, GameViewer } from '@/features/game/types/game.types'

export type Gate = { ok: boolean; reasonKey: string; params?: Record<string, number> }

// startGate mirrors game.Game.Start's checks so the Start button can explain
// *why* it's disabled before the server ever gets asked - first-failing-wins
// so the message is deterministic. Versus needs equal, non-empty teams (not
// necessarily full to config.teamSize - that's only the per-team capacity
// ceiling); Gauntlet just needs at least one player.
export function startGate(snapshot: GameSnapshot, you: GameViewer): Gate {
  if (!you.isHost) return { ok: false, reasonKey: 'game.start.reasonNotHost' }
  if (snapshot.state !== 'LOBBY') return { ok: false, reasonKey: 'game.start.reasonAlreadyStarted' }
  if (snapshot.config.stageMangas.length === 0) return { ok: false, reasonKey: 'game.start.reasonNoStageMangas' }
  if (snapshot.config.powerMangas.length === 0) return { ok: false, reasonKey: 'game.start.reasonNoPowerMangas' }

  if (snapshot.mode === 'VERSUS') {
    if (snapshot.teams.length !== 2) return { ok: false, reasonKey: 'game.start.reasonNeedsTwoTeams' }
    const [a, b] = snapshot.teams
    if (a.memberIds.length === 0 || b.memberIds.length === 0) {
      return { ok: false, reasonKey: 'game.start.reasonEmptyTeam' }
    }
    if (a.memberIds.length !== b.memberIds.length) {
      return {
        ok: false,
        reasonKey: 'game.start.reasonUnequalTeams',
        params: { a: a.memberIds.length, b: b.memberIds.length },
      }
    }
  } else if (snapshot.participants.length === 0) {
    return { ok: false, reasonKey: 'game.start.reasonNoPlayers' }
  }

  return { ok: true, reasonKey: 'game.start.ready' }
}

// canSwitchTeam mirrors game.Game.SwitchTeam's checks.
export function canSwitchTeam(snapshot: GameSnapshot, you: GameViewer, targetParticipantId: string, teamId: string): Gate {
  if (snapshot.state !== 'LOBBY') return { ok: false, reasonKey: 'game.lobby.reasonNotInLobby' }
  if (targetParticipantId !== you.participantId && !you.isHost) {
    return { ok: false, reasonKey: 'game.lobby.reasonNotHost' }
  }
  const target = snapshot.participants.find((p) => p.id === targetParticipantId)
  const team = snapshot.teams.find((t) => t.id === teamId)
  if (!target || !team) return { ok: false, reasonKey: 'game.lobby.reasonNotFound' }
  if (target.teamId === teamId) return { ok: true, reasonKey: 'game.lobby.ready' }
  if (team.memberIds.length >= snapshot.config.teamSize) {
    return { ok: false, reasonKey: 'game.lobby.reasonTeamFull' }
  }
  return { ok: true, reasonKey: 'game.lobby.ready' }
}

// canKick mirrors game.Game.Kick's checks.
export function canKick(snapshot: GameSnapshot, you: GameViewer, targetParticipantId: string): Gate {
  if (!you.isHost) return { ok: false, reasonKey: 'game.lobby.reasonNotHost' }
  if (snapshot.state !== 'LOBBY') return { ok: false, reasonKey: 'game.lobby.reasonNotInLobby' }
  if (targetParticipantId === you.participantId) return { ok: false, reasonKey: 'game.kick.reasonSelf' }
  return { ok: true, reasonKey: 'game.lobby.ready' }
}

// canTransferHost mirrors game.Game.TransferHost's checks.
export function canTransferHost(snapshot: GameSnapshot, you: GameViewer, targetParticipantId: string): Gate {
  if (!you.isHost) return { ok: false, reasonKey: 'game.lobby.reasonNotHost' }
  const target = snapshot.participants.find((p) => p.id === targetParticipantId)
  if (!target) return { ok: false, reasonKey: 'game.lobby.reasonNotFound' }
  if (target.kind === 'BOT') return { ok: false, reasonKey: 'game.transferHost.reasonBot' }
  if (!target.connected) return { ok: false, reasonKey: 'game.lobby.reasonNotFound' }
  return { ok: true, reasonKey: 'game.lobby.ready' }
}

const TEAM_TONES = ['blue', 'red', 'green', 'grape'] as const
export type TeamTone = (typeof TEAM_TONES)[number]

// teamTone maps a team's index onto a stable palette token, never a raw
// color - the same tones ChannelTile/GlossButton already use.
export function teamTone(index: number): TeamTone {
  return TEAM_TONES[index % TEAM_TONES.length]
}

// teamToneColor resolves a tone to its actual Tamagui token - extracted from
// team-column.tsx's own TONE_BG map so both TeamColumn and the in-match
// roster (match-roster.tsx) share one definition instead of duplicating it.
const TONE_COLORS: Record<TeamTone, string> = {
  blue: '$wiiBlue',
  red: '$strawHatRed',
  green: '$meadowGreen',
  grape: '$grapeSoda',
}

export function teamToneColor(tone: TeamTone): string {
  return TONE_COLORS[tone]
}
