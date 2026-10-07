import { foregroundAction, onlineAction } from '../reconnect-policy'

describe('foregroundAction', () => {
  it('asks for a fresh snapshot when the socket still looks open', () => {
    expect(foregroundAction('open')).toBe('resync')
  })

  it.each(['reconnecting', 'connecting', 'closed', 'unavailable'] as const)(
    'retries immediately when the socket is %s',
    (status) => {
      expect(foregroundAction(status)).toBe('retry')
    }
  )
})

describe('onlineAction', () => {
  it('leaves an open socket alone', () => {
    expect(onlineAction('open')).toBe('none')
  })

  it.each(['reconnecting', 'connecting', 'closed'] as const)(
    'retries when the socket is %s',
    (status) => {
      expect(onlineAction(status)).toBe('retry')
    }
  )
})
