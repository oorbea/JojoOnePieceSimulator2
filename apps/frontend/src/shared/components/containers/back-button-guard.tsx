import { useBackButtonGuard } from '@/shared/hooks/use-back-button-guard'

// Renders nothing; exists so a screen can attach the back-button guard from
// JSX positioned after its own early returns (hooks can't be called past
// them, but the handler it needs is only defined down there).
export function BackButtonGuard({
  enabled,
  onBackAttempt,
}: {
  enabled: boolean
  onBackAttempt: () => void
}) {
  useBackButtonGuard(enabled, onBackAttempt)
  return null
}
