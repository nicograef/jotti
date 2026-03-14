import { describe, expect, it, vi } from 'vitest'
import { z } from 'zod'

import { Backend, BackendError } from './Backend'

const dummyTokenGetter = {
  getToken(): string | null {
    return null
  },
}

function createClient() {
  return new Backend('/api', dummyTokenGetter)
}

describe('Backend.post', () => {
  it('throws BackendError with backend code/details from JSON error payload', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({ code: 'validation_failed', details: 'x' }),
          {
            status: 400,
            headers: { 'Content-Type': 'application/json' },
          },
        ),
      ),
    )

    const backend = createClient()

    await expect(
      backend.post('service/foo', {}, z.object({ ok: z.boolean() })),
    ).rejects.toMatchObject({
      status: 400,
      code: 'validation_failed',
      message: 'BackendError: validation_failed - x',
    })
  })

  it('throws unknown BackendError for non-JSON error payloads', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockImplementation(
          () => new Response('<html>upstream error</html>', { status: 502 }),
        ),
    )

    const backend = createClient()
    const request = backend.post(
      'service/foo',
      {},
      z.object({ ok: z.boolean() }),
    )

    await expect(request).rejects.toBeInstanceOf(BackendError)
    await expect(request).rejects.toMatchObject({
      status: 502,
      code: 'unknown',
    })
    await expect(request).rejects.toThrow('upstream error')
  })
})
