/**
 * Fronts one R2 bucket via a native Workers R2 binding, so the backend can
 * reach R2 through Cloudflare's generic *.workers.dev pool instead of R2's
 * S3 API endpoint (<account>.r2.cloudflarestorage.com) - see
 * apps/backend/internal/infrastructure/storage/workerproxy's doc comment
 * for why that distinction matters (an ISP can have a broken route to
 * that one Cloudflare anycast prefix while everything else Cloudflare
 * fronts, this Worker included, works fine).
 *
 * Wire protocol (must stay in lockstep with workerproxy/backend.go):
 *   PUT    /objects/<key>            body=raw bytes, header Content-Type
 *   GET    /objects/<key>            -> body=raw bytes, header Content-Type
 *   DELETE /objects/<key>            -> 204 whether or not <key> existed
 *   GET    /objects?cursor=<cursor>  -> {"keys":[{"key","size"}...],"cursor"}
 *
 * Auth: every request needs either
 *   - "Authorization: Bearer <R2_PROXY_SECRET>", or
 *   - (GET only) "?exp=<unix-seconds>&sig=<hex-hmac-sha256>" where sig is
 *     HMAC-SHA256(secret, `${key}|${exp}`) - see Backend.PresignGet.
 */

export interface Env {
	BUCKET: R2Bucket
	R2_PROXY_SECRET: string
}

const OBJECTS_PREFIX = '/objects'

export default {
	async fetch(request: Request, env: Env, _ctx: ExecutionContext): Promise<Response> {
		const url = new URL(request.url)
		if (!url.pathname.startsWith(OBJECTS_PREFIX)) {
			return new Response('not found', { status: 404 })
		}
		const rest = url.pathname.slice(OBJECTS_PREFIX.length)

		if (rest === '' || rest === '/') {
			if (request.method !== 'GET') return new Response('method not allowed', { status: 405 })
			if (!hasValidBearer(request, env.R2_PROXY_SECRET)) return unauthorized()
			return handleList(url, env)
		}

		const key = decodeKey(rest)
		if (key === null) return new Response('bad key', { status: 400 })

		switch (request.method) {
			case 'PUT':
				if (!hasValidBearer(request, env.R2_PROXY_SECRET)) return unauthorized()
				return handlePut(request, env, key)
			case 'GET':
				if (!hasValidBearer(request, env.R2_PROXY_SECRET) && !(await hasValidSignature(url, key, env.R2_PROXY_SECRET))) {
					return unauthorized()
				}
				return handleGet(env, key)
			case 'DELETE':
				if (!hasValidBearer(request, env.R2_PROXY_SECRET)) return unauthorized()
				return handleDelete(env, key)
			default:
				return new Response('method not allowed', { status: 405 })
		}
	},
} satisfies ExportedHandler<Env>

// decodeKey turns "/stands/1/main.webp" back into "stands/1/main.webp",
// decoding each path segment individually - the mirror image of
// Backend.objectURL's per-segment url.PathEscape on the Go side.
function decodeKey(pathAfterPrefix: string): string | null {
	const trimmed = pathAfterPrefix.replace(/^\//, '')
	if (trimmed === '') return null
	try {
		return trimmed
			.split('/')
			.map((segment) => decodeURIComponent(segment))
			.join('/')
	} catch {
		return null
	}
}

function unauthorized(): Response {
	return new Response('unauthorized', { status: 401 })
}

function hasValidBearer(request: Request, secret: string): boolean {
	const header = request.headers.get('Authorization') ?? ''
	const want = `Bearer ${secret}`
	return timingSafeEqual(header, want)
}

async function hasValidSignature(url: URL, key: string, secret: string): Promise<boolean> {
	const expRaw = url.searchParams.get('exp')
	const sig = url.searchParams.get('sig')
	if (!expRaw || !sig) return false
	const exp = Number(expRaw)
	if (!Number.isFinite(exp) || exp < Date.now() / 1000) return false
	return timingSafeEqual(sig, await expectedSignature(key, expRaw, secret))
}

// expectedSignature mirrors Backend.PresignGet exactly: hex-encoded
// HMAC-SHA256 of "<key>|<exp>" under the shared secret.
async function expectedSignature(key: string, expRaw: string, secret: string): Promise<string> {
	const cryptoKey = await crypto.subtle.importKey(
		'raw',
		new TextEncoder().encode(secret),
		{ name: 'HMAC', hash: 'SHA-256' },
		false,
		['sign']
	)
	const mac = await crypto.subtle.sign('HMAC', cryptoKey, new TextEncoder().encode(`${key}|${expRaw}`))
	return [...new Uint8Array(mac)].map((b) => b.toString(16).padStart(2, '0')).join('')
}

// timingSafeEqual compares two strings without leaking their length
// difference/prefix match through timing - both defenses matter for a
// bearer token and for an HMAC signature.
function timingSafeEqual(a: string, b: string): boolean {
	// Walk the longer string's full length regardless of the outcome, so a
	// length mismatch doesn't return measurably faster than a same-length
	// mismatch.
	const len = Math.max(a.length, b.length)
	let diff = a.length ^ b.length
	for (let i = 0; i < len; i++) {
		diff |= (a.charCodeAt(i) || 0) ^ (b.charCodeAt(i) || 0)
	}
	return diff === 0
}

async function handlePut(request: Request, env: Env, key: string): Promise<Response> {
	const contentType = request.headers.get('Content-Type') ?? 'application/octet-stream'
	await env.BUCKET.put(key, request.body, { httpMetadata: { contentType } })
	return new Response(null, { status: 200 })
}

async function handleGet(env: Env, key: string): Promise<Response> {
	const obj = await env.BUCKET.get(key)
	if (obj === null) return new Response('not found', { status: 404 })
	const headers = new Headers()
	headers.set('Content-Type', obj.httpMetadata?.contentType ?? 'application/octet-stream')
	headers.set('Content-Length', String(obj.size))
	return new Response(obj.body, { status: 200, headers })
}

async function handleDelete(env: Env, key: string): Promise<Response> {
	await env.BUCKET.delete(key)
	return new Response(null, { status: 204 })
}

async function handleList(url: URL, env: Env): Promise<Response> {
	const cursor = url.searchParams.get('cursor') ?? undefined
	const listed = await env.BUCKET.list({ cursor, limit: 1000 })
	const body = {
		keys: listed.objects.map((o) => ({ key: o.key, size: o.size })),
		cursor: listed.truncated ? listed.cursor : '',
	}
	return new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } })
}
