// Public barrel — this is the ONLY way other features/routes should import
// from this feature. Everything else (lib/, internal hooks) stays
// unexported and reachable only from inside this folder.
export { useServiceWorkerUpdate } from './hooks/use-service-worker-update'
export { useInstallPromptCapture } from './hooks/use-install-prompt-capture'
export { useWakeLock } from './hooks/use-wake-lock'
