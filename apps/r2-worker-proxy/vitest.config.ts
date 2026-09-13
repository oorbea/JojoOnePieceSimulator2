import { cloudflareTest } from '@cloudflare/vitest-plugin'
import { defineConfig } from 'vitest/config'

export default defineConfig({
	plugins: [
		cloudflareTest({
			wrangler: { configPath: './wrangler.toml' },
			miniflare: {
				// The test env needs its own R2 bucket binding value - Miniflare
				// creates it in-memory, it never touches the real prod bucket.
				r2Buckets: ['BUCKET'],
				bindings: { R2_PROXY_SECRET: 'test-secret' },
			},
		}),
	],
})
