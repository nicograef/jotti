import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import { Zahlung } from './Zahlung'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

vi.mock('@/lib/Auth', () => ({
  AuthSingleton: { canCancel: true, userId: 1 },
}))

// vaul's Drawer braucht Browser-APIs, die jsdom nicht bereitstellt. Trigger
// inline rendern, Drawer-Inhalt ausblenden — so bleiben nur die Aktionsleiste
// und der Auszahlung-Trigger übrig.
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

function renderZahlung() {
  render(
    <Zahlung
      backend={{
        zahlungKassieren: vi.fn().mockResolvedValue(undefined),
        auszahlungLeisten: vi.fn().mockResolvedValue(undefined),
      }}
      tisch={tisch}
      positionen={[position]}
      saldoCents={0}
      loading={false}
      onZahlungKassiert={vi.fn()}
      onAuszahlungGeleistet={vi.fn()}
    />,
  )
}

describe('Zahlung Aktionsleiste', () => {
  it('ist ohne Auswahl deaktiviert und zeigt nach Auswahl Anzahl und Summe', async () => {
    const user = userEvent.setup()
    renderZahlung()

    expect(screen.getByRole('button', { name: /Kassieren/ })).toBeDisabled()

    await user.click(screen.getByRole('button', { name: 'Produkt hinzufügen' }))

    const bar = screen.getByRole('button', { name: /Kassieren/ })
    expect(bar).toBeEnabled()
    expect(bar).toHaveTextContent('3,50')
  })

  it('hält die Auszahlung weiterhin erreichbar', () => {
    renderZahlung()

    expect(
      screen.getByRole('button', { name: 'Auszahlung' }),
    ).toBeInTheDocument()
  })
})
