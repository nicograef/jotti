import { act, renderHook } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { useVorgang } from './use-vorgang'

describe('useVorgang', () => {
  it('behält den Schlüssel über einen Retry mit gleichem Inhalt', () => {
    const { result } = renderHook(() => useVorgang<number>())
    act(() => {
      result.current.steuerung.add(1)
    })

    let erster = ''
    let zweiter = ''
    act(() => {
      erster = result.current.schluesselFuer({ positionen: [1] })
    })
    act(() => {
      zweiter = result.current.schluesselFuer({ positionen: [1] })
    })

    expect(zweiter).toBe(erster)
  })

  it('erneuert den Schlüssel, wenn sich der Inhalt nach einem Versuch ändert', () => {
    const { result } = renderHook(() => useVorgang<number>())
    act(() => {
      result.current.steuerung.add(1)
    })

    let erster = ''
    let geaendert = ''
    act(() => {
      erster = result.current.schluesselFuer({ kommentar: '' })
    })
    act(() => {
      geaendert = result.current.schluesselFuer({ kommentar: 'Ohne Zwiebeln' })
    })

    expect(geaendert).not.toBe(erster)
    expect(result.current.vorgangId).toBe(geaendert)
  })

  it('beginnt mit einer neu gefüllten Auswahl mit neuem Schlüssel und leerem Kommentar', () => {
    const { result } = renderHook(() => useVorgang<number>())
    act(() => {
      result.current.steuerung.add(1)
    })
    act(() => {
      result.current.setKommentar('Ohne Zwiebeln')
    })
    const ersterVorgang = result.current.vorgangId

    act(() => {
      result.current.steuerung.reset()
    })
    act(() => {
      result.current.steuerung.add(1)
    })

    expect(result.current.vorgangId).not.toBe(ersterVorgang)
    expect(result.current.kommentar).toBe('')
  })
})
