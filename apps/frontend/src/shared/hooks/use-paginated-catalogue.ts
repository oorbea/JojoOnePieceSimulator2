import { useInfiniteQuery, type Query } from '@tanstack/react-query'
import { useMemo, useRef } from 'react'
import { Platform } from 'react-native'

// Shared cursor-paginated list contract every `?limit=`/`?cursor=`
// catalogue endpoint returns (dto.PageInfo embedded + `items`) - see
// ObsidianVault/catalogue-pagination.md for the backend side.
export type CataloguePage<T> = {
  items: T[]
  nextCursor?: string
  total?: number
}

// See use-stands.ts's useStands for why polling only happens on native (web
// uses PictureEventsBridge's SSE push instead), and why the attempt count
// lives in a ref instead of query.state.dataUpdateCount - identical
// reasoning applies to the infinite-query form.
const MAX_POLL_ATTEMPTS = 8
const MAX_POLL_INTERVAL_MS = 30_000
const BASE_POLL_INTERVAL_MS = 2_000

type Options<T> = {
  // Called once per query key change (locale/filters), not per page - a
  // pending picture on *any* loaded page keeps the whole infinite query
  // polling until it resolves.
  hasPendingPicture: (item: T) => boolean
}

// Generic `useInfiniteQuery` wrapper over a cursor-paginated catalogue
// endpoint - collapses what used to be three near-identical hooks
// (use-stands.ts/use-devil-fruits.ts/use-stages.ts's *unpaginated* form) is
// left alone; this is the paginated counterpart for screens that opt into
// `?limit=`/`?cursor=` instead of the legacy bare-array fetch. The cursor
// lives entirely in TanStack Query's own pageParam machinery, never in the
// query key - `queryKey` should be the same shape as e.g. standKeys.list(),
// no cursor spliced in.
export function usePaginatedCatalogue<T>(
  queryKey: readonly unknown[],
  fetchPage: (cursor: string | undefined, limit: number) => Promise<CataloguePage<T>>,
  { hasPendingPicture }: Options<T>,
  limit = 24
) {
  const pollAttempts = useRef(0)

  const query = useInfiniteQuery({
    queryKey,
    queryFn: ({ pageParam }) => fetchPage(pageParam, limit),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => lastPage.nextCursor,
    refetchInterval:
      Platform.OS === 'web'
        ? undefined
        : (q: Query<{ pages: CataloguePage<T>[] }>) => {
            const hasPending = q.state.data?.pages.some((page) => page.items.some(hasPendingPicture))
            if (!hasPending) {
              pollAttempts.current = 0
              return false
            }
            if (pollAttempts.current >= MAX_POLL_ATTEMPTS) return false
            const interval = Math.min(BASE_POLL_INTERVAL_MS * 2 ** pollAttempts.current, MAX_POLL_INTERVAL_MS)
            pollAttempts.current += 1
            return interval
          },
    refetchIntervalInBackground: true,
  })

  const items = useMemo(() => query.data?.pages.flatMap((page) => page.items) ?? [], [query.data])
  const total = query.data?.pages[0]?.total

  return { ...query, items, total }
}
