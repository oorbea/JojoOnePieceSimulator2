import { clockOffsetMs, recordServerTime, resetServerClock, serverNow } from '../server-clock'

describe('server-clock', () => {
  beforeEach(() => {
    resetServerClock()
  })

  it('reads as plain Date.now() before any sample arrives', () => {
    expect(clockOffsetMs()).toBe(0)
  })

  it('estimates the offset from a single sample', () => {
    // Server clock is 5s ahead of this device's.
    recordServerTime(15_000, 10_000)
    expect(clockOffsetMs()).toBe(5_000)
  })

  it('picks the LARGEST observed offset - the least-latency sample', () => {
    // True offset is 5000ms. Three samples of the same server instant
    // arriving at different local receive times, each implying a smaller
    // offset the more network latency it carried.
    recordServerTime(20_000, 15_200) // implies +4800 (200ms latency)
    recordServerTime(20_000, 15_050) // implies +4950 (50ms latency)
    recordServerTime(20_000, 15_500) // implies +4500 (500ms latency)
    expect(clockOffsetMs()).toBe(4_950)
  })

  it('ignores a non-finite sample instead of poisoning the window', () => {
    recordServerTime(15_000, 10_000)
    recordServerTime(Number.NaN, 10_000)
    expect(clockOffsetMs()).toBe(5_000)
  })

  it('drops samples once the window is full, oldest first', () => {
    // Window is 20. Seed it with a strong (large) offset, then push 20 weak
    // ones - the strong one should eventually age out.
    recordServerTime(100_000, 0) // offset +100000
    for (let i = 0; i < 20; i++) {
      recordServerTime(1_000, 0) // offset +1000
    }
    expect(clockOffsetMs()).toBe(1_000)
  })

  it('serverNow() applies the estimated offset to Date.now()', () => {
    jest.spyOn(Date, 'now').mockReturnValue(10_000)
    recordServerTime(15_000, 10_000)
    expect(serverNow()).toBe(15_000)
  })
})
