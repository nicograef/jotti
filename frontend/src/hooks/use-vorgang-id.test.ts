import { renderHook } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { useVorgangId } from './use-vorgang-id'

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/

describe('useVorgangId', () => {
  it('behält den Schlüssel über die ganze Zusammenstellung', () => {
    const { result, rerender } = renderHook(
      (istLeer: boolean) => useVorgangId(istLeer),
      { initialProps: false },
    )

    const erster = result.current
    expect(erster).toMatch(UUID)

    // Auswahl, Mengen und Kommentar ändern sich hier laufend — der Schlüssel
    // nicht: Eine abweichende Zweiteinreichung soll der Server erkennen können.
    rerender(false)
    rerender(false)
    expect(result.current).toBe(erster)
  })

  it('wechselt den Schlüssel erst beim Übergang von leer zu nicht leer', () => {
    const { result, rerender } = renderHook(
      (istLeer: boolean) => useVorgangId(istLeer),
      { initialProps: false },
    )
    const erster = result.current

    // Das Leeren allein (etwa nach einer erfolgreichen Buchung) rotiert noch
    // nicht — ein Submit ist im Leerzustand ohnehin gesperrt.
    rerender(true)
    expect(result.current).toBe(erster)

    // Erst die neue Zusammenstellung beginnt einen neuen Vorgang.
    rerender(false)
    const zweiter = result.current
    expect(zweiter).toMatch(UUID)
    expect(zweiter).not.toBe(erster)

    rerender(false)
    expect(result.current).toBe(zweiter)
  })

  it('beginnt eine leer gemountete Zusammenstellung mit einem neuen Schlüssel', () => {
    // Die Storno- und Umbuchungs-Drawer mounten ohne Auswahl. Der beim Mount
    // erzeugte Schlüssel wird nie gesendet; die erste Auswahl vergibt einen
    // eigenen.
    const { result, rerender } = renderHook(
      (istLeer: boolean) => useVorgangId(istLeer),
      { initialProps: true },
    )
    const beimMount = result.current

    rerender(false)
    expect(result.current).toMatch(UUID)
    expect(result.current).not.toBe(beimMount)
  })
})
