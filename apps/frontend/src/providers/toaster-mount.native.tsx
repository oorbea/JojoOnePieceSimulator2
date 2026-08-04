// Native platforms render toasts through their own native overlay (burnt
// wraps Alert/Toast APIs there) - no mount point is needed, unlike web's
// sonner-backed <Toaster/>. See toaster-mount.web.tsx.
export function ToasterMount() {
  return null
}
