// R2_PROXY_SECRET is a Worker *secret* (`wrangler secret put`), not a
// `[vars]` entry in wrangler.toml, so `wrangler types` can't see it and
// leaves it out of the generated Cloudflare.Env. This merges it in by
// declaration merging so both the real runtime env and the test env
// (vitest.config.ts's miniflare.bindings) type-check.
declare namespace Cloudflare {
	interface Env {
		R2_PROXY_SECRET: string
	}
}
