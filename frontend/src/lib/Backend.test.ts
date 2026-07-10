import { describe, expect, it, vi } from 'vitest'
import { z } from 'zod'

import { AuthSingleton } from './Auth'
import { Backend, BackendError, ResponseBodyError } from './Backend'

const dummyTokenGetter = {
  getToken(): string | null {
    return null
  },
}

function createClient() {
  return new Backend('/api', dummyTokenGetter)
}

describe('Backend.post', () => {
  // Ein abgelaufenes Token beantwortet das Backend mit 401 (invalid_jwt):
  // Der Client meldet den Benutzer ab und leitet zur Login-Seite um.
  it('logs the user out on 401, e.g. for an expired token', async () => {
    // jsdom kann die Redirect-Navigation nicht ausführen und loggt einen Fehler.
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    const logoutSpy = vi.spyOn(AuthSingleton, 'logout')
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ code: 'invalid_jwt' }), {
          status: 401,
          headers: { 'Content-Type': 'application/json' },
        }),
      ),
    )

    const backend = createClient()

    await expect(backend.post('service/foo', {})).rejects.toMatchObject({
      status: 401,
      code: 'unauthorized',
    })
    expect(logoutSpy).toHaveBeenCalled()
  })

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

  it('exposes the X-Correlation-ID response header as referenz on BackendError', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ code: 'internal_server_error' }), {
          status: 500,
          headers: {
            'Content-Type': 'application/json',
            'X-Correlation-ID': 'abc12345',
          },
        }),
      ),
    )

    const backend = createClient()

    await expect(backend.post('service/foo', {})).rejects.toMatchObject({
      status: 500,
      code: 'internal_server_error',
      referenz: 'abc12345',
    })
  })

  it('accepts non-string error details and exposes them on BackendError', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            code: 'validation_error',
            details: { name: ['Name ist erforderlich'] },
          }),
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
      code: 'validation_error',
      details: { name: ['Name ist erforderlich'] },
    })
  })

  it('includes endpoint and issue path for response schema mismatch without leaking values', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            unbezahltePositionen: [
              { steuersatz: 'SECRET', secretToken: 'SECRET' },
            ],
          }),
          {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          },
        ),
      ),
    )

    const backend = createClient()
    const request = backend.post(
      'service/get-tisch-state',
      {},
      z.object({
        unbezahltePositionen: z.array(
          z.object({
            steuersatz: z.number().int(),
          }),
        ),
      }),
    )

    await expect(request).rejects.toBeInstanceOf(ResponseBodyError)
    await expect(request).rejects.toThrow(
      'Response of service/get-tisch-state is invalid: unbezahltePositionen[0].steuersatz (invalid_type)',
    )

    let thrown: Error | null = null
    try {
      await request
    } catch (e) {
      thrown = e as Error
    }

    expect(thrown).not.toBeNull()
    expect(thrown?.message).not.toContain('SECRET')
  })
})
