import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'

import { ErfolgsPop } from './ErfolgsPop'

describe('ErfolgsPop', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    VorgangsRegisterSingleton.zuruecksetzen()
  })

  afterEach(() => {
    vi.useRealTimers()
    cleanup()
  })

  it('zeigt bei geöffnetem Pop den Bestätigungstext als Status-Region', () => {
    render(<ErfolgsPop open text="Zahlung erfolgreich." onDismiss={vi.fn()} />)

    expect(screen.getByRole('status')).toHaveTextContent('Zahlung erfolgreich.')
  })

  it('hält die Status-Region auch geschlossen ohne Text im DOM', () => {
    render(
      <ErfolgsPop
        open={false}
        text="Zahlung erfolgreich."
        onDismiss={vi.fn()}
      />,
    )

    // Die Live-Region bleibt dauerhaft gemountet (nur der Inhalt wird
    // getoggelt), damit Screenreader spätere Textwechsel zuverlässig ansagen.
    expect(screen.getByRole('status')).toBeEmptyDOMElement()
  })

  it('schließt nach Ablauf der Anzeigedauer automatisch', () => {
    const onDismiss = vi.fn()
    render(<ErfolgsPop open text="Fertig" onDismiss={onDismiss} />)

    expect(onDismiss).not.toHaveBeenCalled()
    act(() => {
      vi.advanceTimersByTime(1400)
    })
    expect(onDismiss).toHaveBeenCalledTimes(1)
  })

  it('schließt beim Antippen sofort', () => {
    const onDismiss = vi.fn()
    render(<ErfolgsPop open text="Fertig" onDismiss={onDismiss} />)

    fireEvent.click(screen.getByRole('status'))

    expect(onDismiss).toHaveBeenCalledTimes(1)
  })

  it('meldet als reine Anzeige keinen offenen Vorgang', () => {
    render(<ErfolgsPop open text="Zahlung erfolgreich." onDismiss={vi.fn()} />)

    // Die Erfolgsmeldung verschwindet nach 1,4 s von selbst; sie darf einen
    // erzwungenen Reload nicht aufhalten.
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })
})
