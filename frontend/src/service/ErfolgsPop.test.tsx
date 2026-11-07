import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ComponentProps } from 'react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { useErfolgsPop } from '@/hooks/use-erfolgs-pop'
import { useVorgang } from '@/hooks/use-vorgang'
import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'
import { FakeBackend } from '@/test/FakeBackend'
import { testProdukt } from '@/test/fixtures'
import { renderWithBackend, setViewportWidth } from '@/test/render'

import { DirektverkaufTab } from './direktverkauf/DirektverkaufTab'
import { ErfolgsPop } from './ErfolgsPop'
import { ServiceDock } from './ServiceDock'

// Verkauf und Pfand-Vorgang liegen auf der Seite; der Harness übernimmt ihre Rolle.
function DirektverkaufTabHarness(
  props: Omit<
    ComponentProps<typeof DirektverkaufTab>,
    'vorgang' | 'pfandVorgang'
  >,
) {
  return (
    <DirektverkaufTab
      {...props}
      vorgang={useVorgang<number>()}
      pfandVorgang={useVorgang<number>()}
    />
  )
}

// toast kommt über den Fehlerpfad von useActionSubmit; die Erfolgs-Flows dürfen
// ihn nicht aufrufen (prüft der Flow-Test unten).
vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

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

// Verdrahtet wie die Buchungsseiten: Der Erfolg öffnet den Pop, der Refetch
// (hier `reload`) läuft erst beim Schließen.
function DirektverkaufMitPop({ reload }: { reload: () => void }) {
  const { zeigeErfolg, pop } = useErfolgsPop(reload)
  return (
    <ServiceDock leiste={null}>
      <DirektverkaufTabHarness
        pfandProdukte={[]}
        products={[testProdukt]}
        productsLoading={false}
        onErfolg={zeigeErfolg}
      />
      <ErfolgsPop {...pop} />
    </ServiceDock>
  )
}

describe('Erfolgs-Pop im Buchungsflow', () => {
  // Der Buchungsflow-Test unten prüft den Handy-Pfad (Dock-Button plus
  // Bottom-Sheet); ab lg trägt die feste Spalte den Abschluss.
  beforeEach(() => {
    setViewportWidth(375)
  })

  afterEach(() => {
    cleanup()
    setViewportWidth(1024)
  })

  it('öffnet den Pop ohne Erfolgs-Toast und lädt erst nach dem Schließen nach', async () => {
    const { toast } = await import('sonner')
    const user = userEvent.setup()
    const reload = vi.fn()
    const fake = new FakeBackend().respond('service/direktverkauf-taetigen', {})
    renderWithBackend(<DirektverkaufMitPop reload={reload} />, fake)

    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    await user.click(screen.getByRole('button', { name: /Kassieren/ }))
    await screen.findByRole('dialog')
    await user.click(
      screen.getByRole('button', { name: 'Verkauf abschließen' }),
    )

    await screen.findByText('Verkauf abgeschlossen.')
    expect(reload).not.toHaveBeenCalled()
    expect(toast.success).not.toHaveBeenCalled()

    await waitFor(
      () => {
        expect(
          screen.queryByText('Verkauf abgeschlossen.'),
        ).not.toBeInTheDocument()
      },
      { timeout: 2000 },
    )
    expect(reload).toHaveBeenCalledTimes(1)
  })
})
