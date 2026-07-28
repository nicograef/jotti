import { afterEach, describe, expect, it, vi } from 'vitest'
import { z } from 'zod'

import { AuthSingleton } from './Auth'
import {
  Backend,
  BackendError,
  NetzwerkFehler,
  ResponseBodyError,
} from './Backend'

const dummyTokenGetter = {
  getToken(): string | null {
    return null
  },
}

function createClient() {
  return new Backend('/api', dummyTokenGetter)
}

// Antwort-Body, dessen Stream nie etwas liefert und erst durch das
// Abbruch-Signal endet — so verhält sich fetch, wenn das Zeitlimit zuschlägt,
// während der Body noch überträgt.
function stockenderBodyStream(
  signal: AbortSignal | null | undefined,
): ReadableStream<Uint8Array> {
  return new ReadableStream<Uint8Array>({
    start(controller) {
      signal?.addEventListener('abort', () => {
        controller.error(new DOMException('Aborted', 'AbortError'))
      })
    },
  })
}

// Antwort-Body, dessen Übertragung abreißt, ohne dass abgebrochen wurde.
function abgerissenerBodyStream(): ReadableStream<Uint8Array> {
  return new ReadableStream<Uint8Array>({
    start(controller) {
      controller.error(new TypeError('network error'))
    },
  })
}

afterEach(() => {
  vi.useRealTimers()
  vi.unstubAllGlobals()
  vi.restoreAllMocks()
})

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

describe('Backend Verbindungsfehler', () => {
  // Ein Request, der nie antwortet (klassisches WLAN-Loch), muss clientseitig
  // enden — sonst bleibt die Query dauerhaft im Ladezustand.
  it('bricht einen hängenden Request nach 8 Sekunden als Zeitüberschreitung ab', async () => {
    vi.useFakeTimers()
    vi.stubGlobal(
      'fetch',
      vi.fn(
        (_url: string, init: RequestInit) =>
          new Promise<Response>((_resolve, reject) => {
            init.signal?.addEventListener('abort', () => {
              reject(new DOMException('Aborted', 'AbortError'))
            })
          }),
      ),
    )

    const backend = createClient()
    const request = backend
      .post('service/foo', {})
      .catch((error: unknown) => error)

    await vi.advanceTimersByTimeAsync(8000)

    const fehler = await request
    expect(fehler).toBeInstanceOf(NetzwerkFehler)
    expect(fehler).toMatchObject({ art: 'zeitueberschreitung' })
  })

  // Das Zeitlimit endet nicht mit den Headern: Antwortet der Server mit 200,
  // liefert dann aber keinen Body mehr, muss der Aufruf ebenfalls abbrechen.
  it('bricht einen in der Body-Phase stockenden Request als Zeitüberschreitung ab', async () => {
    vi.useFakeTimers()
    vi.stubGlobal(
      'fetch',
      vi.fn(
        (_url: string, init: RequestInit) =>
          new Response(stockenderBodyStream(init.signal), {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          }),
      ),
    )

    const backend = createClient()
    const request = backend
      .post('service/foo', {}, z.object({ ok: z.boolean() }))
      .catch((error: unknown) => error)

    await vi.advanceTimersByTimeAsync(8000)

    const fehler = await request
    expect(fehler).toBeInstanceOf(NetzwerkFehler)
    expect(fehler).toMatchObject({ art: 'zeitueberschreitung' })
  })

  // Reißt die Verbindung in der Body-Phase ab, ohne dass das Zeitlimit
  // abgelaufen ist, bleibt es ein Verbindungsabbruch.
  it('meldet einen Abriss in der Body-Phase als Verbindungsabbruch', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(
        () =>
          new Response(abgerissenerBodyStream(), {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          }),
      ),
    )

    const backend = createClient()
    const request = backend.post(
      'service/foo',
      {},
      z.object({ ok: z.boolean() }),
    )

    await expect(request).rejects.toBeInstanceOf(NetzwerkFehler)
    await expect(request).rejects.toMatchObject({ art: 'verbindungsabbruch' })
  })

  // Langlaufende Endpunkte (TSE-Einrichtung, DSFinV-K-Export) setzen ihr eigenes
  // Zeitlimit; die Voreinstellung von 8 Sekunden darf sie nicht abbrechen.
  it('verwendet ein explizit übergebenes Zeitlimit statt der Voreinstellung', async () => {
    vi.useFakeTimers()
    vi.stubGlobal(
      'fetch',
      vi.fn(
        (_url: string, init: RequestInit) =>
          new Promise<Response>((_resolve, reject) => {
            init.signal?.addEventListener('abort', () => {
              reject(new DOMException('Aborted', 'AbortError'))
            })
          }),
      ),
    )

    const backend = createClient()
    let erledigt = false
    const request = backend
      .post('admin/tse-einrichten', {}, undefined, { zeitlimitMs: 30_000 })
      .catch((error: unknown) => error)
      .finally(() => {
        erledigt = true
      })

    await vi.advanceTimersByTimeAsync(8000)
    expect(erledigt).toBe(false)

    await vi.advanceTimersByTimeAsync(22_000)

    const fehler = await request
    expect(fehler).toBeInstanceOf(NetzwerkFehler)
    expect(fehler).toMatchObject({ art: 'zeitueberschreitung' })
  })

  it('meldet einen Verbindungsabbruch als NetzwerkFehler', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockRejectedValue(new TypeError('Failed to fetch')),
    )

    const backend = createClient()
    const request = backend.post('service/foo', {})

    await expect(request).rejects.toBeInstanceOf(NetzwerkFehler)
    await expect(request).rejects.toMatchObject({ art: 'verbindungsabbruch' })
  })

  // Ein abgeschnittener Body wirft in der Web-API einen rohen SyntaxError; für
  // den Aufrufer ist das ein Verbindungsproblem, kein Schema-Verstoß.
  it('meldet einen abgebrochenen Antwort-Body als NetzwerkFehler', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response('{"ok":tr', {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        }),
      ),
    )

    const backend = createClient()
    const request = backend.post(
      'service/foo',
      {},
      z.object({ ok: z.boolean() }),
    )

    await expect(request).rejects.toBeInstanceOf(NetzwerkFehler)
    await expect(request).rejects.not.toBeInstanceOf(ResponseBodyError)
  })

  it('leitet bei mehreren gleichzeitigen 401-Antworten nur einmal zum Login weiter', async () => {
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
    const ergebnisse = await Promise.allSettled([
      backend.post('service/a', {}),
      backend.post('service/b', {}),
      backend.post('service/c', {}),
    ])

    expect(ergebnisse.map((ergebnis) => ergebnis.status)).toEqual([
      'rejected',
      'rejected',
      'rejected',
    ])
    expect(logoutSpy).toHaveBeenCalledTimes(1)
  })
})
