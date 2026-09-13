import { env } from 'cloudflare:workers'
import { createExecutionContext, waitOnExecutionContext } from 'cloudflare:test'
import { describe, expect, it } from 'vitest'

import worker from '../src/index'

const SECRET = 'test-secret' // must match vitest.config.ts's bindings.R2_PROXY_SECRET

async function call(request: Request): Promise<Response> {
	const ctx = createExecutionContext()
	const response = await worker.fetch(request, env, ctx)
	await waitOnExecutionContext(ctx)
	return response
}

function authed(url: string, init: RequestInit = {}): Request {
	return new Request(url, {
		...init,
		headers: { ...init.headers, Authorization: `Bearer ${SECRET}` },
	})
}

describe('auth', () => {
	it('rejects a request with no Authorization header', async () => {
		const res = await call(new Request('https://worker.test/objects/a'))
		expect(res.status).toBe(401)
	})

	it('rejects a request with the wrong bearer token', async () => {
		const res = await call(
			new Request('https://worker.test/objects/a', { headers: { Authorization: 'Bearer wrong' } })
		)
		expect(res.status).toBe(401)
	})

	it('rejects anything outside /objects', async () => {
		const res = await call(authed('https://worker.test/other'))
		expect(res.status).toBe(404)
	})
})

describe('PUT/GET/DELETE round trip', () => {
	it('stores, reads back with content-type, then deletes', async () => {
		const putRes = await call(
			authed('https://worker.test/objects/stands/1/main.webp', {
				method: 'PUT',
				headers: { 'Content-Type': 'image/webp' },
				body: new TextEncoder().encode('hello world'),
			})
		)
		expect(putRes.status).toBe(200)

		const getRes = await call(authed('https://worker.test/objects/stands/1/main.webp'))
		expect(getRes.status).toBe(200)
		expect(getRes.headers.get('Content-Type')).toBe('image/webp')
		expect(await getRes.text()).toBe('hello world')

		const delRes = await call(
			authed('https://worker.test/objects/stands/1/main.webp', { method: 'DELETE' })
		)
		expect(delRes.status).toBe(204)

		const afterDelete = await call(authed('https://worker.test/objects/stands/1/main.webp'))
		expect(afterDelete.status).toBe(404)
	})

	it('deleting a key that never existed is not an error', async () => {
		const res = await call(authed('https://worker.test/objects/never-existed', { method: 'DELETE' }))
		expect(res.status).toBe(204)
	})

	it('getting a missing key is a 404', async () => {
		const res = await call(authed('https://worker.test/objects/missing'))
		expect(res.status).toBe(404)
	})

	it('round-trips a key containing multiple path segments', async () => {
		await call(
			authed('https://worker.test/objects/one-piece-characters/abc-123/def-456.webp', {
				method: 'PUT',
				headers: { 'Content-Type': 'image/png' },
				body: new TextEncoder().encode('bytes'),
			})
		)
		const res = await call(authed('https://worker.test/objects/one-piece-characters/abc-123/def-456.webp'))
		expect(res.status).toBe(200)
		expect(await res.text()).toBe('bytes')
	})
})

describe('list', () => {
	it('returns every stored key with its size', async () => {
		await call(
			authed('https://worker.test/objects/list-test/a', { method: 'PUT', body: new TextEncoder().encode('12') })
		)
		await call(
			authed('https://worker.test/objects/list-test/b', {
				method: 'PUT',
				body: new TextEncoder().encode('1234'),
			})
		)

		const res = await call(authed('https://worker.test/objects'))
		expect(res.status).toBe(200)
		const body = (await res.json()) as { keys: { key: string; size: number }[]; cursor: string }
		const byKey = Object.fromEntries(body.keys.map((k) => [k.key, k.size]))
		expect(byKey['list-test/a']).toBe(2)
		expect(byKey['list-test/b']).toBe(4)
		expect(body.cursor).toBe('')
	})

	it('requires auth', async () => {
		const res = await call(new Request('https://worker.test/objects'))
		expect(res.status).toBe(401)
	})
})

describe('signed GET URL (PresignGet)', () => {
	async function sign(key: string, exp: number): Promise<string> {
		const cryptoKey = await crypto.subtle.importKey(
			'raw',
			new TextEncoder().encode(SECRET),
			{ name: 'HMAC', hash: 'SHA-256' },
			false,
			['sign']
		)
		const mac = await crypto.subtle.sign('HMAC', cryptoKey, new TextEncoder().encode(`${key}|${exp}`))
		return [...new Uint8Array(mac)].map((b) => b.toString(16).padStart(2, '0')).join('')
	}

	it('accepts a valid, unexpired signature without a bearer header', async () => {
		await call(
			authed('https://worker.test/objects/signed/a', { method: 'PUT', body: new TextEncoder().encode('x') })
		)
		const exp = Math.floor(Date.now() / 1000) + 900
		const sig = await sign('signed/a', exp)

		const res = await call(new Request(`https://worker.test/objects/signed/a?exp=${exp}&sig=${sig}`))
		expect(res.status).toBe(200)
	})

	it('rejects an expired signature', async () => {
		await call(
			authed('https://worker.test/objects/signed/b', { method: 'PUT', body: new TextEncoder().encode('x') })
		)
		const exp = Math.floor(Date.now() / 1000) - 10
		const sig = await sign('signed/b', exp)

		const res = await call(new Request(`https://worker.test/objects/signed/b?exp=${exp}&sig=${sig}`))
		expect(res.status).toBe(401)
	})

	it('rejects a signature for a different key', async () => {
		await call(
			authed('https://worker.test/objects/signed/c', { method: 'PUT', body: new TextEncoder().encode('x') })
		)
		const exp = Math.floor(Date.now() / 1000) + 900
		const sigForOtherKey = await sign('signed/not-c', exp)

		const res = await call(new Request(`https://worker.test/objects/signed/c?exp=${exp}&sig=${sigForOtherKey}`))
		expect(res.status).toBe(401)
	})

	it('rejects a signed GET but still refuses PUT/DELETE without a bearer token', async () => {
		const exp = Math.floor(Date.now() / 1000) + 900
		const sig = await sign('signed/d', exp)
		const putRes = await call(
			new Request(`https://worker.test/objects/signed/d?exp=${exp}&sig=${sig}`, {
				method: 'PUT',
				body: new TextEncoder().encode('x'),
			})
		)
		expect(putRes.status).toBe(401)
	})
})
