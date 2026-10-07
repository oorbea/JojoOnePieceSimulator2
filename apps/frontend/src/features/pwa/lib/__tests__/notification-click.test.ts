import { parseNotificationClick } from '../notification-click'

describe('parseNotificationClick', () => {
  it('returns the path of a notification tap', () => {
    expect(parseNotificationClick({ type: 'NOTIFICATION_CLICK', url: '/play/abc' })).toBe('/play/abc')
  })

  it('ignores other messages', () => {
    expect(parseNotificationClick({ type: 'SKIP_WAITING', url: '/play/abc' })).toBeNull()
    expect(parseNotificationClick(null)).toBeNull()
    expect(parseNotificationClick('NOTIFICATION_CLICK')).toBeNull()
    expect(parseNotificationClick(undefined)).toBeNull()
  })

  it.each(['https://evil.example/', '//evil.example/', 'javascript:alert(1)', '', 7, null])(
    'never navigates to a non-app url (%p)',
    (url) => {
      expect(parseNotificationClick({ type: 'NOTIFICATION_CLICK', url })).toBeNull()
    }
  )
})
