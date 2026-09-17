import { AppError } from '@/shared/api/errors'
import { inviteErrorKey } from '@/features/game/lib/invite-error'

function errorWithCode(code: string): AppError {
  return new AppError('boom', { code })
}

describe('inviteErrorKey', () => {
  it.each([
    ['INVITE_INVALID', 'game.invite.error.expired'],
    ['INVITE_REVOKED', 'game.invite.error.revoked'],
    ['GAME_ALREADY_STARTED', 'game.invite.error.started'],
    ['INVALID_STATE_TRANSITION', 'game.invite.error.started'],
    ['GAME_FULL', 'game.invite.error.full'],
    ['TEAM_FULL', 'game.invite.error.full'],
    ['LOBBY_LOCKED', 'game.invite.error.locked'],
    ['GAME_NOT_FOUND', 'game.invite.error.gone'],
    ['SOMETHING_UNMAPPED', 'game.invite.error.generic'],
  ])('%s -> %s', (code, expected) => {
    expect(inviteErrorKey(errorWithCode(code))).toBe(expected)
  })

  it('falls back to generic when there is no code at all', () => {
    expect(inviteErrorKey(new AppError('boom'))).toBe('game.invite.error.generic')
  })
})
