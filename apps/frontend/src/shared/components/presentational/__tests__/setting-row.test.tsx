import { Text } from 'react-native'
import { renderWithProviders, screen } from '@/test/render'

import { SettingRow } from '../setting-row'

describe('SettingRow', () => {
  it('renders both the label and the control', async () => {
    await renderWithProviders(
      <SettingRow label="Privacy">
        <Text>Private</Text>
      </SettingRow>
    )

    expect(screen.getByText('Privacy')).toBeTruthy()
    expect(screen.getByText('Private')).toBeTruthy()
  })

  it('renders an optional help slot next to the label', async () => {
    await renderWithProviders(
      <SettingRow label="Privacy" help={<Text>?</Text>}>
        <Text>Private</Text>
      </SettingRow>
    )

    expect(screen.getByText('?')).toBeTruthy()
  })
})
