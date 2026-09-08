import { act, screen, waitFor } from '@testing-library/react-native'
import { Text } from 'react-native'

import { renderWithProviders } from '@/test/render'
import { usePaginatedCatalogue, type CataloguePage } from '@/shared/hooks/use-paginated-catalogue'

type Item = { id: string; pending: boolean }

function page(items: Item[], nextCursor?: string, total?: number): CataloguePage<Item> {
  return { items, nextCursor, total }
}

// Renders through a tiny probe component - see use-debounced-value.test.tsx
// for why this codebase has no renderHook precedent.
function Probe({ fetchPage }: { fetchPage: (cursor: string | undefined, limit: number) => Promise<CataloguePage<Item>> }) {
  const { items, hasNextPage, isFetchingNextPage, isFetchNextPageError, total, fetchNextPage, isLoading } =
    usePaginatedCatalogue(['probe'], fetchPage, { hasPendingPicture: (i) => i.pending })

  return (
    <>
      <Text testID="loading">{String(isLoading)}</Text>
      <Text testID="items">{items.map((i) => i.id).join(',')}</Text>
      <Text testID="hasNext">{String(hasNextPage)}</Text>
      <Text testID="fetchingNext">{String(isFetchingNextPage)}</Text>
      <Text testID="nextError">{String(isFetchNextPageError)}</Text>
      <Text testID="total">{String(total)}</Text>
      <Text testID="loadMore" onPress={() => void fetchNextPage()}>
        load more
      </Text>
    </>
  )
}

describe('usePaginatedCatalogue', () => {
  it('loads the first page', async () => {
    const fetchPage = jest.fn().mockResolvedValue(page([{ id: 'a', pending: false }], 'cursor-1', 2))
    await renderWithProviders(<Probe fetchPage={fetchPage} />)

    await waitFor(() => expect(screen.getByTestId('items').props.children).toBe('a'))
    expect(screen.getByTestId('hasNext').props.children).toBe('true')
    expect(screen.getByTestId('total').props.children).toBe('2')
    expect(fetchPage).toHaveBeenCalledWith(undefined, 24)
  })

  it('fetchNextPage appends, preserving order, and passes the cursor through', async () => {
    const fetchPage = jest
      .fn()
      .mockResolvedValueOnce(page([{ id: 'a', pending: false }], 'cursor-1', 2))
      .mockResolvedValueOnce(page([{ id: 'b', pending: false }], undefined, 2))
    await renderWithProviders(<Probe fetchPage={fetchPage} />)
    await waitFor(() => expect(screen.getByTestId('items').props.children).toBe('a'))

    await act(async () => {
      screen.getByTestId('loadMore').props.onPress()
    })

    await waitFor(() => expect(screen.getByTestId('items').props.children).toBe('a,b'))
    expect(fetchPage).toHaveBeenLastCalledWith('cursor-1', 24)
    expect(screen.getByTestId('hasNext').props.children).toBe('false')
  })

  it('hasNextPage is false without a nextCursor', async () => {
    const fetchPage = jest.fn().mockResolvedValue(page([{ id: 'a', pending: false }], undefined, 1))
    await renderWithProviders(<Probe fetchPage={fetchPage} />)

    await waitFor(() => expect(screen.getByTestId('items').props.children).toBe('a'))
    expect(screen.getByTestId('hasNext').props.children).toBe('false')
  })

  it('a rejected second page leaves the first page intact', async () => {
    const fetchPage = jest
      .fn()
      .mockResolvedValueOnce(page([{ id: 'a', pending: false }], 'cursor-1', 2))
      .mockRejectedValueOnce(new Error('boom'))
    await renderWithProviders(<Probe fetchPage={fetchPage} />)
    await waitFor(() => expect(screen.getByTestId('items').props.children).toBe('a'))

    await act(async () => {
      screen.getByTestId('loadMore').props.onPress()
    })

    await waitFor(() => expect(screen.getByTestId('nextError').props.children).toBe('true'))
    expect(screen.getByTestId('items').props.children).toBe('a')
  })
})
