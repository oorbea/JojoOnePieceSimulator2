import AsyncStorage from '@react-native-async-storage/async-storage'

// Thin re-export so the persist client and any future consumer import from
// `@/shared/lib` rather than reaching into a third-party package directly —
// keeps the storage backend swappable in one place.
export { AsyncStorage }
