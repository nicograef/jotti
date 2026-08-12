import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ComponentProps } from 'react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { useMengen } from '@/hooks/use-mengen'
import { useIsMobile } from '@/hooks/use-mobile'
import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'

import type { Produkt } from '../../product/Produkt'
import type { Tisch } from '../../table/Tisch'
import { ServiceDock } from '../ServiceDock'
import { Bestellung } from './Bestellung'

// Mit der Produktebene liegt die Variantenliste eine Navigationsebene tiefer:
// erst das Produkt oeffnen, dann ist der Hinzufuegen-Knopf da.
async function produktOeffnen(
  user: ReturnType<typeof userEvent.setup>,
  name = 'Bratwurst',
) {
  await user.click(screen.getByRole('button', { name: new RegExp(name) }))
}

// Der Bestell-Korb liegt seit A1 in TablePage; für die isolierten Komponenten-
// Tests stellt dieser Harness die gehobene Steuerung bereit.
function BestellungHarness(
  props: Omit<ComponentProps<typeof Bestellung>, 'mengenSteuerung'>,
) {
  const mengenSteuerung = useMengen<number>()
  return <Bestellung {...props} mengenSteuerung={mengenSteuerung} />
}

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

// Standardmäßig Handy-Layout (Dock-Aktionsbutton); ein Test unten schaltet auf
// Desktop, um die Verdrahtung der festen Spalte zu prüfen. Deren
// container-neutrales Verhalten deckt BestellungAbschluss.test.tsx ab.
vi.mock('@/hooks/use-mobile', () => ({
  useIsMobile: vi.fn(() => true),
}))

beforeEach(() => {
  VorgangsRegisterSingleton.zuruecksetzen()
})

afterEach(() => {
  cleanup()
  vi.mocked(useIsMobile).mockReturnValue(true)
})

const tisch: Tisch = { id: 1, name: 'Stammtisch', saldoCents: 0 }

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

describe('Bestellung Aktionsleiste', () => {
  it('ist ohne Auswahl deaktiviert und zeigt nach Auswahl Anzahl und Summe', async () => {
    const user = userEvent.setup()
    render(
      <ServiceDock leiste={null}>
        <BestellungHarness
          backend={{
            bestellungAufnehmen: vi.fn().mockResolvedValue(undefined),
          }}
          tisch={tisch}
          products={[testProdukt]}
          productsLoading={false}
          onErfolg={vi.fn()}
        />
      </ServiceDock>,
    )

    expect(
      screen.getByRole('button', { name: /Bestellung überprüfen/ }),
    ).toBeDisabled()

    await produktOeffnen(user)
    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )

    const bar = screen.getByRole('button', { name: /Bestellung überprüfen/ })
    expect(bar).toBeEnabled()
    expect(bar).toHaveTextContent('3,50')
  })

  it('rendert ab lg die feste Abschluss-Spalte statt Dock und Drawer', async () => {
    vi.mocked(useIsMobile).mockReturnValue(false)
    const user = userEvent.setup()
    // Kein ServiceDock: die feste Spalte trägt den Aktionsbutton selbst.
    render(
      <BestellungHarness
        backend={{ bestellungAufnehmen: vi.fn().mockResolvedValue(undefined) }}
        tisch={tisch}
        products={[testProdukt]}
        productsLoading={false}
        onErfolg={vi.fn()}
      />,
    )

    expect(screen.getByText('Bratwurst')).toBeInTheDocument()
    const button = screen.getByRole('button', { name: 'Bestellung aufnehmen' })
    expect(button).toBeDisabled()
    expect(
      screen.queryByRole('button', { name: /Bestellung überprüfen/ }),
    ).not.toBeInTheDocument()

    await produktOeffnen(user)
    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    expect(
      screen.getByRole('button', { name: 'Bestellung aufnehmen' }),
    ).toBeEnabled()
  })
})

// Der Wechsel zwischen Drawer- und Spaltenlayout ist einer der beiden realen
// Auslöser für ein Zähler-Leck im Vorgangs-Register: Der harte
// if(!isMobile)-Zweig tauscht ganze Teilbäume aus. Der Korb selbst liegt
// darüber (in TablePage bzw. hier im Harness) und bleibt derselbe Vorgang — er
// darf sich beim Wechsel weder ein zweites Mal melden noch stehen bleiben.
describe('Bestellung im Vorgangs-Register', () => {
  it('meldet den Korb über einen Layout-Wechsel hinweg genau einmal', async () => {
    const user = userEvent.setup()
    const renderUi = () => (
      <ServiceDock leiste={null}>
        <BestellungHarness
          backend={{
            bestellungAufnehmen: vi.fn().mockResolvedValue(undefined),
          }}
          tisch={tisch}
          products={[testProdukt]}
          productsLoading={false}
          onErfolg={vi.fn()}
        />
      </ServiceDock>
    )
    const { rerender, unmount } = render(renderUi())

    await produktOeffnen(user)
    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)

    vi.mocked(useIsMobile).mockReturnValue(false)
    rerender(renderUi())

    // Der Teilbaum ist getauscht (feste Abschluss-Spalte statt Drawer), der
    // Korb steht weiterhin.
    expect(
      screen.getByRole('button', { name: 'Bestellung aufnehmen' }),
    ).toBeEnabled()
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)

    unmount()
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })
})
