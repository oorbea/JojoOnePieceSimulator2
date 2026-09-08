import { __resetImageQueue } from '@/shared/lib/image-queue'

// image-queue.ts is module-level state shared across every test file in a
// worker - without this, a test that enqueues an image and never settles it
// leaks a slot into the next test. Registered via `setupFilesAfterEach`, not
// `jest.setup.ts` (`setupFiles`) - `afterEach` isn't defined yet at the
// `setupFiles` stage, which runs before Jest's test framework globals are
// installed.
afterEach(() => {
  __resetImageQueue()
})
