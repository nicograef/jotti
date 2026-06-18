import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import { Ausgabe } from './Ausgabe'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

vi.mock('@/lib/Auth', () => ({
  AuthSingleton: { canCancel: true, userId: 1 },
}))

// vaul's Drawer braucht Browser-APIs, die jsdom nicht bereitstellt. Trigger
// inline rendern, Drawer-Inhalt ausblenden — so bleibt nur die Positionsliste.
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

const position: Position = {
  positionId: '00000000-0000-0000-0000-000000000001',
  varianteId: 1,
  produktName: 'Bratwurst',
  varianteName: 'Normal',
  kategorie: 'essen',
  steuersatz: 'regel',
  einzelpreis: 350,
  menge: 2,
  bestellerUserId: 1,
  bestellerName: 'Tester',
}

// Position einer anderen Servicekraft (bestellerUserId ≠ 1).
const fremdePosition: Position = {
  positionId: '00000000-0000-0000-0000-000000000002',
  varianteId: 2,
  produktName: 'Pommes',
  varianteName: 'Normal',
  kategorie: 'essen',
  steuersatz: 'regel',
  einzelpreis: 250,
  menge: 1,
  bestellerUserId: 2,
  bestellerName: 'Kollegin',
}

function renderAusgabe(positionen: Position[] = [position]) {
  render(
    <Ausgabe
      backend={{ ausgabeBestaetigen: vi.fn().mockResolvedValue(undefined) }}
      tisch={tisch}
      positionen={positionen}
      loading={false}
      onAusgabeBestaetigt={vi.fn()}
    />,
  )
}

describe('Ausgabe Positionsgruppen', () => {
  it('zeigt eigene Positionen direkt und fremde erst nach „Alle anzeigen"', async () => {
    const user = userEvent.setup()
    renderAusgabe([position, fremdePosition])

    expect(screen.getByText('Bratwurst Normal')).toBeInTheDocument()
    expect(screen.queryByText('Pommes Normal')).not.toBeInTheDocument()

    await user.click(
      screen.getByRole('button', { name: /Alle anzeigen \(1 von/ }),
    )

    expect(screen.getByText('Pommes Normal')).toBeInTheDocument()
    expect(screen.getByText('von Kollegin')).toBeInTheDocument()
  })

  it('ohne fremde Positionen gibt es keinen „Alle anzeigen"-Schalter', () => {
    renderAusgabe([position])

    expect(screen.getByText('Bratwurst Normal')).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: /Alle anzeigen/ }),
    ).not.toBeInTheDocument()
  })
})
