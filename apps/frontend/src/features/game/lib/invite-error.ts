import type { AppError } from '@/shared/api/errors'

// Maps a JoinByInvite failure onto one of the join-invite screen's
// dedicated error messages (game.invite.error.*) - see the backend's
// JoinByInvite doc (game_invite.go) for exactly which code each cause
// returns. Falls back to a generic message for anything unmapped rather
// than surfacing a raw error code.
export function inviteErrorKey(error: AppError): string {
  switch (error.code) {
    case 'INVITE_INVALID':
      return 'game.invite.error.expired'
    case 'INVITE_REVOKED':
      return 'game.invite.error.revoked'
    case 'GAME_ALREADY_STARTED':
    case 'INVALID_STATE_TRANSITION':
      return 'game.invite.error.started'
    case 'GAME_FULL':
    case 'TEAM_FULL':
      return 'game.invite.error.full'
    case 'LOBBY_LOCKED':
      return 'game.invite.error.locked'
    case 'GAME_NOT_FOUND':
      return 'game.invite.error.gone'
    default:
      return 'game.invite.error.generic'
  }
}
