// Tamagui narrows style-prop types (e.g. `bg`) to a finite literal union of
// token strings, which is great for literals written directly in JSX but
// too strict for a value looked up from a Record<Variant, string> at
// runtime — the Record's value type widens to plain `string`. This is a
// narrow, explicit escape hatch for exactly that case: the token name is
// still one of our own registered tokens, TypeScript just can't prove it
// came from the finite set without either restating the whole token union
// here or losing type safety on the Record's keys instead.
export function asToken<T>(value: string): T {
  return value as unknown as T
}
