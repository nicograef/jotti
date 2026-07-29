import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'

import type { Produkt } from '../product/Produkt'
import { Direktverkauf } from './direktverkauf/Direktverkauf'
import { ErfolgsPop } from './ErfolgsPop'
import { ServiceDock } from './ServiceDock'

// toast wird über useActionSubmit (Fehlerpfad) importiert; die Erfolgs-Flows
// dürfen ihn nicht mehr aufrufen, was der Flow-Test unten prüft.
vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

// Der Buchungsflow-Test unten prüft den Handy-Pfad (Dock-Button plus
// Bottom-Sheet); ab lg trägt die feste Spalte den Abschluss.
vi.mock('@/hooks/use-mobile', () => ({
  useIsMobile: () => true,
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

    // Ein Tap schließt den Pop früher als der Auto-Dismiss-Timer.
    expect(onDismiss).toHaveBeenCalledTimes(1)
  })

  it('meldet als reine Anzeige keinen offenen Vorgang', () => {
    render(<ErfolgsPop open text="Zahlung erfolgreich." onDismiss={vi.fn()} />)

    // Die Erfolgsmeldung verschwindet nach 1,4 s von selbst; sie darf einen
    // erzwungenen Reload nicht aufhalten.
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })
})

const testProdukt: Produkt = {
  id: 1,
  name: 'Bratwurst',
  kategorie: 'essen',
  status: 'active',
  varianten: [
    {
      id: 1,
      name: 'Normal',
      preisCents: 350,
      status: 'active',
      createdAt: '2025-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z',
    },
  ],
  createdAt: '2025-01-01T00:00:00Z',
  updatedAt: '2025-01-01T00:00:00Z',
}

// Spiegelt die Verdrahtung der Buchungsseiten wider: Der Erfolg öffnet den Pop
// (statt eines Toasts), und der nachgelagerte Refetch (hier `reload`) läuft erst
// beim Schließen.
function DirektverkaufMitPop({
  direktverkaufTaetigen,
  reload,
}: {
  direktverkaufTaetigen: () => Promise<void>
  reload: () => void
}) {
  const [erfolg, setErfolg] = useState({ open: false, text: '' })
  return (
    <ServiceDock leiste={null}>
      <Direktverkauf
        backend={{ direktverkaufTaetigen }}
        products={[testProdukt]}
        productsLoading={false}
        onErfolg={(nachricht) => {
          setErfolg({ open: true, text: nachricht })
        }}
      />
      <ErfolgsPop
        open={erfolg.open}
        text={erfolg.text}
        onDismiss={() => {
          setErfolg((prev) => ({ ...prev, open: false }))
          reload()
        }}
      />
    </ServiceDock>
  )
}

describe('Erfolgs-Pop im Buchungsflow', () => {
  afterEach(() => {
    cleanup()
  })

  it('öffnet den Pop ohne Erfolgs-Toast und lädt erst nach dem Schließen nach', async () => {
    const { toast } = await import('sonner')
    const user = userEvent.setup()
    const reload = vi.fn()
    const direktverkaufTaetigen = vi.fn().mockResolvedValue(undefined)
    render(
      <DirektverkaufMitPop
        direktverkaufTaetigen={direktverkaufTaetigen}
        reload={reload}
      />,
    )

    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    await user.click(screen.getByRole('button', { name: /Kassieren/ }))
    await screen.findByRole('dialog')
    await user.click(
      screen.getByRole('button', { name: 'Verkauf abschließen' }),
    )

    // Der Pop erscheint mit der Bestätigung; der Refetch läuft noch nicht und es
    // gibt keinen Erfolgs-Toast mehr.
    await screen.findByText('Verkauf abgeschlossen.')
    expect(reload).not.toHaveBeenCalled()
    expect(toast.success).not.toHaveBeenCalled()

    // Der Auto-Dismiss-Timer schließt den Pop und löst erst dann den
    // nachgelagerten Refetch aus.
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
