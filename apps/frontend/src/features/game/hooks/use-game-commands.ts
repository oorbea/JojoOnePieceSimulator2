import { useGameSocketStore } from '@/features/game/stores/game-socket.store'
import { CLIENT_COMMAND } from '@/features/game/types/game-ws.types'
import type { UpdateGameConfigInput } from '@/features/game/types/game.types'

// Typed command senders over the socket store's raw send() - one function
// per WS command, so containers never hand-build a payload shape.
export function useGameCommands() {
  const send = useGameSocketStore((s) => s.send)

  return {
    leave: () => send(CLIENT_COMMAND.LEAVE),
    addBot: (teamId: string) => send(CLIENT_COMMAND.ADD_BOT, { teamId }),
    removeBot: (botId: string) => send(CLIENT_COMMAND.REMOVE_BOT, { botId }),
    start: () => send(CLIENT_COMMAND.START),
    abort: () => send(CLIENT_COMMAND.ABORT),
    vote: (option: string) => send(CLIENT_COMMAND.VOTE, { option }),
    revealReady: () => send(CLIENT_COMMAND.REVEAL_READY),
    resync: () => send(CLIENT_COMMAND.RESYNC),
    switchTeam: (teamId: string) => send(CLIENT_COMMAND.SWITCH_TEAM, { teamId }),
    movePlayer: (participantId: string, teamId: string) =>
      send(CLIENT_COMMAND.MOVE_PLAYER, { participantId, teamId }),
    kick: (participantId: string) => send(CLIENT_COMMAND.KICK, { participantId }),
    transferHost: (participantId: string) => send(CLIENT_COMMAND.TRANSFER_HOST, { participantId }),
    setLocked: (locked: boolean) => send(CLIENT_COMMAND.SET_LOCK, { locked }),
    // UPDATE_CONFIG is a full replacement (mirrors CreateGameRequest, plus
    // the fields only editable once a lobby exists) - callers must build
    // the whole payload from current + edited fields, never a patch.
    updateConfig: (input: UpdateGameConfigInput) =>
      send(CLIENT_COMMAND.UPDATE_CONFIG, input as unknown as Record<string, unknown>),
  }
}
