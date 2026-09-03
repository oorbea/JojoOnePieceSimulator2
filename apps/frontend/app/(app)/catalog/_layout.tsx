import { Slot } from 'expo-router'

// No extra guard here - app/(app)/_layout.tsx already requires a session
// for everything under (app), and the catalog is open to every logged-in
// user regardless of role (unlike /admin, which further restricts to
// ADMIN).
export default function CatalogGroupLayout() {
  return <Slot />
}
