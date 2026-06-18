import { cleanup, render, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Bestellung } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import type { Zahlung } from '../../table/Zahlung'
import { TischHistorie } from './TischHistorie'

type HistorieEintrag = Bestellung | Zahlung

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

vi.mock('@/lib/Auth', () => ({
  AuthSingleton: { canCancel: true },
}))

// vaul's Drawer braucht Browser-APIs, die jsdom nicht bereitstellt. Hier wird
// nur die flache Historien-Liste geprüft, kein Drawer geöffnet.
vi.mock('@/components/ui/drawer', () => {
  const Passthrough = ({ children }: { children?: ReactNode }) => children
  return {
    Drawer: Passthrough,
    DrawerTrigger: Passthrough,
    DrawerContent: () => null,
    DrawerHeader: Passthrough,
    DrawerTitle: Passthrough,
    DrawerDescription: Passthrough,
    DrawerFooter: Passthrough,
    DrawerClose: Passthrough,
  }
})

afterEach(() => {
  cleanup()
})

const tisch: Tisch = { id: 1, name: 'Stammtisch', saldoCents: 0 }

function bestellung(overrides: Partial<Bestellung> = {}): Bestellung {
  return {
    art: 'bestellung',
    id: '00000000-0000-0000-0000-000000000001',
    userId: 1,
    userName: 'Tester',
    tischId: 1,
    positionen: [
      {
        positionId: '00000000-0000-0000-0000-0000000000a1',
        varianteId: 1,
        produktName: 'Bratwurst',
        varianteName: 'Normal',
        kategorie: 'essen',
        steuersatz: 'regel',
        einzelpreis: 350,
        menge: 1,
        bestellerUserId: 1,
        bestellerName: 'Tester',
      },
    ],
    gesamtPreisCents: 350,
    kommentar: '',
    aufgenommenAm: '2026-06-18T12:00:00Z',
    stornierbarePositionen: [],
    umbuchbarePositionen: [],
    ...overrides,
  }
}

function zahlung(overrides: Partial<Zahlung> = {}): Zahlung {
  return {
    art: 'zahlung',
    id: '00000000-0000-0000-0000-0000000000f1',
    userId: 2,
    userName: 'Bert',
    tischId: 1,
    positionen: [
      {
        positionId: '00000000-0000-0000-0000-0000000000a1',
        varianteId: 1,
        produktName: 'Bratwurst',
        varianteName: 'Normal',
        kategorie: 'essen',
        steuersatz: 'regel',
        einzelpreis: 350,
        menge: 1,
        bestellerUserId: 1,
        bestellerName: 'Anna',
      },
    ],
    gesamtZahlungCents: 350,
    kommentar: '',
    kassiertAm: '2026-06-18T12:05:00Z',
    ...overrides,
  }
}

function renderHistorie(historie: HistorieEintrag[]) {
  render(
    <TischHistorie
      historie={historie}
      historieLoading={false}
      tisch={tisch}
      backend={{
        stornierungErteilen: vi.fn().mockResolvedValue(undefined),
        bestellungUmbuchen: vi.fn().mockResolvedValue(undefined),
        belegDrucken: vi.fn().mockResolvedValue(undefined),
      }}
      onStornierungErteilt={vi.fn()}
      onBestellungUmgebucht={vi.fn()}
    />,
  )
}

describe('TischHistorie', () => {
  it('beschriftet jeden Eintrag mit dem Namen der handelnden Servicekraft', () => {
    renderHistorie([
      bestellung({
        id: '00000000-0000-0000-0000-000000000001',
        userName: 'Anna',
      }),
      zahlung({
        id: '00000000-0000-0000-0000-0000000000f1',
        userName: 'Bert',
      }),
    ])

    expect(screen.getByText('von Anna')).toBeInTheDocument()
    expect(screen.getByText('von Bert')).toBeInTheDocument()
  })

  it('zeigt die Historie flach — alle Einträge ohne „Alle anzeigen"-Schalter', () => {
    renderHistorie([
      bestellung({ id: '00000000-0000-0000-0000-000000000001' }),
      bestellung({ id: '00000000-0000-0000-0000-000000000002' }),
      bestellung({ id: '00000000-0000-0000-0000-000000000003' }),
    ])

    expect(screen.getAllByText(/Bestellung \+/)).toHaveLength(3)
    expect(
      screen.queryByRole('button', { name: /Alle anzeigen/ }),
    ).not.toBeInTheDocument()
  })
})
