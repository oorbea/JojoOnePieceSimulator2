// Root key factory. Each feature extends this under its own namespace, e.g.
//   export const standKeys = {
//     all: [...queryKeys.root, 'stands'] as const,
//     detail: (id: string) => [...standKeys.all, id] as const,
//   }
// so invalidation ("all stands") and precise keys ("this one stand") share a
// prefix without every feature re-deciding the convention.
export const queryKeys = {
  root: ['jops'] as const,
}
