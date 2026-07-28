import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Produkt } from '../../product/Produkt'
import type { Tisch } from '../../table/Tisch'
import { ServiceDock } from '../ServiceDock'
import { BestellungDrawer } from './BestellungDrawer'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

afterEach(() => {
  cleanup()
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

// Der Aktionsbutton rendert per Portal in den ServiceDock; deshalb wird der
// Drawer hier in ein Dock eingebettet (leiste bleibt leer).
function renderDrawer(mengen: Record<number, number>) {
  render(
    <ServiceDock leiste={null}>
      <BestellungDrawer
        backend={{ bestellungAufnehmen: vi.fn().mockResolvedValue(undefined) }}
        tisch={tisch}
        products={[testProdukt]}
        mengen={mengen}
        bestellungId="11111111-1111-4111-8111-111111111111"
        bestellungAufgenommen={vi.fn()}
        vorgangBereitsGebucht={vi.fn()}
      />
    </ServiceDock>,
  )
}

describe('BestellungDrawer', () => {
  it('öffnet bei leerer Auswahl nicht (Guard über deaktivierten Trigger)', async () => {
    const user = userEvent.setup()
    renderDrawer({})

    // Ohne Auswahl ist der Trigger-Button deaktiviert; ein Klick darf den
    // Drawer nicht öffnen (der onOpenChange-Guard sichert zusätzlich ab).
    const trigger = screen.getByRole('button', {
      name: /Bestellung überprüfen/,
    })
    expect(trigger).toBeDisabled()
    await user.click(trigger)

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('öffnet mit Auswahl über den Trigger; Gesamtsumme und Footer-Buttons liegen außerhalb des Scrollbereichs', async () => {
    const user = userEvent.setup()
    renderDrawer({ 1: 2 })

    await user.click(
      screen.getByRole('button', { name: /Bestellung überprüfen/ }),
    )

    const dialog = await screen.findByRole('dialog')
    const body = dialog.querySelector('[data-slot="drawer-body"]')
    const footer = dialog.querySelector('[data-slot="drawer-footer"]')
    expect(body).not.toBeNull()
    expect(footer).not.toBeNull()
    expect(body).toContainElement(screen.getByText(/Bratwurst/))
    // Die Gesamtsumme steht im nicht-scrollenden Footer, nicht im Body.
    const gesamt = screen.getByText('Gesamt')
    expect(footer).toContainElement(gesamt)
    expect(body).not.toContainElement(gesamt)
    expect(body).not.toContainElement(
      screen.getByRole('button', { name: 'Bestellung aufnehmen' }),
    )
    expect(body).not.toContainElement(
      screen.getByRole('button', { name: 'Abbrechen' }),
    )

    await user.click(screen.getByRole('button', { name: 'Abbrechen' }))
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })
  })

  it('öffnet ohne Auto-Fokus auf ein Eingabefeld (keine ungefragte Tastatur)', async () => {
    const user = userEvent.setup()
    renderDrawer({ 1: 2 })

    await user.click(
      screen.getByRole('button', { name: /Bestellung überprüfen/ }),
    )
    const dialog = await screen.findByRole('dialog')

    // Das optionale Kommentarfeld ist das erste Feld im Drawer. Beim Öffnen darf
    // KEIN Eingabefeld den Fokus erhalten — sonst öffnet sich auf dem Handy
    // ungefragt die Tastatur (zentrale onOpenAutoFocus-Unterdrückung im Drawer).
    const kommentar = screen.getByPlaceholderText(/Kommentar/)
    expect(kommentar).not.toHaveFocus()
    const active = document.activeElement
    expect(active?.tagName).not.toBe('INPUT')
    expect(active?.tagName).not.toBe('TEXTAREA')
    // Der Fokus liegt aber IM Dialog (auf dem Container), damit Fokus-Falle,
    // Escape und Fokusrückgabe weiter funktionieren.
    expect(active === dialog || dialog.contains(active)).toBe(true)
  })

  it('zeigt den Tischnamen als dominante Überschrift ohne Prosa-Description', async () => {
    const user = userEvent.setup()
    renderDrawer({ 1: 2 })

    await user.click(
      screen.getByRole('button', { name: /Bestellung überprüfen/ }),
    )

    await screen.findByRole('dialog')
    expect(
      screen.getByRole('heading', { name: tisch.name }),
    ).toBeInTheDocument()
    expect(screen.queryByText(/vor dem Absenden/)).not.toBeInTheDocument()
  })
})
