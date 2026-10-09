// The section list lives in shared/ (any screen can open the manual overlay at
// a section); the manual feature re-exports it for its own imports.
export { MANUAL_SECTIONS, parseSection, type ManualSectionId } from '@/shared/lib/manual-sections'
