import { useAudioSettingsStore } from '../audio-settings.store'
import { AsyncStorage } from '@/shared/lib/async-storage'

describe('useAudioSettingsStore', () => {
  beforeEach(async () => {
    await AsyncStorage.clear()
    useAudioSettingsStore.setState({ muted: false, volume: 1, isHydrated: false })
  })

  it('defaults to unmuted, full volume', () => {
    const state = useAudioSettingsStore.getState()
    expect(state.muted).toBe(false)
    expect(state.volume).toBe(1)
  })

  it('toggleMute flips the flag and persists it', async () => {
    await useAudioSettingsStore.getState().toggleMute()
    expect(useAudioSettingsStore.getState().muted).toBe(true)
    expect(await AsyncStorage.getItem('jops.audio.muted')).toBe('true')

    await useAudioSettingsStore.getState().toggleMute()
    expect(useAudioSettingsStore.getState().muted).toBe(false)
    expect(await AsyncStorage.getItem('jops.audio.muted')).toBe('false')
  })

  it('setVolume clamps to [0, 1] and persists it', async () => {
    await useAudioSettingsStore.getState().setVolume(0.4)
    expect(useAudioSettingsStore.getState().volume).toBe(0.4)
    expect(await AsyncStorage.getItem('jops.audio.volume')).toBe('0.4')

    await useAudioSettingsStore.getState().setVolume(5)
    expect(useAudioSettingsStore.getState().volume).toBe(1)

    await useAudioSettingsStore.getState().setVolume(-2)
    expect(useAudioSettingsStore.getState().volume).toBe(0)
  })

  it('hydrate loads persisted values', async () => {
    await AsyncStorage.setItem('jops.audio.muted', 'true')
    await AsyncStorage.setItem('jops.audio.volume', '0.25')

    await useAudioSettingsStore.getState().hydrate()

    const state = useAudioSettingsStore.getState()
    expect(state.muted).toBe(true)
    expect(state.volume).toBe(0.25)
    expect(state.isHydrated).toBe(true)
  })

  it('hydrate falls back to defaults on missing or invalid persisted values', async () => {
    await AsyncStorage.setItem('jops.audio.volume', 'not-a-number')

    await useAudioSettingsStore.getState().hydrate()

    const state = useAudioSettingsStore.getState()
    expect(state.muted).toBe(false)
    expect(state.volume).toBe(1)
    expect(state.isHydrated).toBe(true)
  })
})
