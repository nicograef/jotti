import { act, renderHook } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'

import { useActionSubmit } from './use-action-submit'

vi.mock('sonner', () => ({
  toast: { error: vi.fn() },
}))

// Buchung, die erst auf Kommando endet: Nur so lässt sich der Zwischenstand
// „läuft noch" prüfen.
function steuerbareBuchung() {
  let abschliessen!: () => void
  let scheitern!: (fehler: Error) => void
  const lauf = new Promise<void>((resolve, reject) => {
    abschliessen = resolve
    scheitern = reject
  })
  return { lauf: () => lauf, abschliessen, scheitern }
}

beforeEach(() => {
  VorgangsRegisterSingleton.zuruecksetzen()
  // Die Fehlschlag-Pfade protokollieren den Fehler; das gehört nicht in die
  // Testausgabe.
  vi.spyOn(console, 'error').mockImplementation(() => undefined)
})

afterEach(() => {
  vi.restoreAllMocks()
})

describe('useActionSubmit im Vorgangs-Register', () => {
  it('meldet die laufende Buchung und gibt sie nach Erfolg frei', async () => {
    const buchung = steuerbareBuchung()
    const { result } = renderHook(() =>
      useActionSubmit({ actionLabel: 'Bestellung aufnehmen' }),
    )

    let lauf!: Promise<void>
    act(() => {
      lauf = result.current.run(buchung.lauf)
    })
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)

    await act(async () => {
      buchung.abschliessen()
      await lauf
    })
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })

  it('gibt die Buchung auch nach einem Fehlschlag frei', async () => {
    const buchung = steuerbareBuchung()
    const { result } = renderHook(() =>
      useActionSubmit({ actionLabel: 'Bestellung aufnehmen' }),
    )

    let lauf!: Promise<void>
    act(() => {
      lauf = result.current.run(buchung.lauf)
    })
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)

    await act(async () => {
      buchung.scheitern(new Error('Netzabbruch'))
      await lauf
    })
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })

  it('gibt eine noch laufende Buchung beim Aushängen frei', async () => {
    const buchung = steuerbareBuchung()
    const { result, unmount } = renderHook(() =>
      useActionSubmit({ actionLabel: 'Bestellung aufnehmen' }),
    )

    let lauf!: Promise<void>
    act(() => {
      lauf = result.current.run(buchung.lauf)
    })
    unmount()

    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)

    // Die Antwort trifft nach dem Aushängen ein und darf nichts nachtragen.
    await act(async () => {
      buchung.abschliessen()
      await lauf
    })
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })
})
