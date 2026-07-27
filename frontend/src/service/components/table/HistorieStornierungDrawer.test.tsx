import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Bestellung, Position } from '../../table/Bestellung'
import type { StornierungErteilen } from '../../table/Stornierung'
import type { Tisch } from '../../table/Tisch'
import { HistorieStornierungDrawer } from './HistorieStornierungDrawer'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

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
  einzelpreisCents: 350,
  menge: 2,
  bestellerUserId: 1,
  bestellerName: 'Tester',
}

const quelle: Bestellung = {
  art: 'bestellung',
  id: '00000000-0000-0000-0000-000000000042',
  userId: 1,
  userName: 'Nico',
  tischId: 1,
  positionen: [position],
  gesamtPreisCents: 700,
  kommentar: '',
  aufgenommenAm: '2026-06-18T12:00:00Z',
  stornierbarePositionen: [position],
  umbuchbarePositionen: [position],
}

describe('HistorieStornierungDrawer', () => {
  it('rendert Positionsliste im DrawerBody; Pflicht-Kommentar und Buttons im sichtbaren Footer', () => {
    render(
      <HistorieStornierungDrawer
        backend={{ stornierungErteilen: vi.fn().mockResolvedValue(undefined) }}
        tisch={tisch}
        quelle={quelle}
        onClose={vi.fn()}
        onStornierungErteilt={vi.fn()}
      />,
    )

    const dialog = screen.getByRole('dialog')
    const body = dialog.querySelector('[data-slot="drawer-body"]')
    const footer = dialog.querySelector('[data-slot="drawer-footer"]')
    expect(body).not.toBeNull()
    expect(footer).not.toBeNull()
    expect(body).toContainElement(screen.getByText(/Bratwurst/))
    // Das Pflichtfeld steht im nicht-scrollenden Footer, nicht im Body.
    const kommentar = screen.getByPlaceholderText('Kommentar (erforderlich)')
    expect(footer).toContainElement(kommentar)
    expect(body).not.toContainElement(kommentar)
    expect(body).not.toContainElement(
      screen.getByRole('button', { name: 'Stornierung erteilen' }),
    )
    expect(body).not.toContainElement(
      screen.getByRole('button', { name: 'Abbrechen' }),
    )
  })

  it('titelt menschenlesbar mit Vorgangstyp und Name statt UUID-Fragment', () => {
    render(
      <HistorieStornierungDrawer
        backend={{ stornierungErteilen: vi.fn().mockResolvedValue(undefined) }}
        tisch={tisch}
        quelle={quelle}
        onClose={vi.fn()}
        onStornierungErteilt={vi.fn()}
      />,
    )

    const title = screen.getByText(/^Bestellung ·/)
    expect(title).toBeInTheDocument()
    expect(title).toHaveTextContent('Nico')
    expect(screen.queryByText(/00000000/)).not.toBeInTheDocument()
  })

  it('nennt den Sperrgrund neben der Aktion und gibt frei, sobald er erfüllt ist', async () => {
    const user = userEvent.setup()
    render(
      <HistorieStornierungDrawer
        backend={{ stornierungErteilen: vi.fn().mockResolvedValue(undefined) }}
        tisch={tisch}
        quelle={quelle}
        onClose={vi.fn()}
        onStornierungErteilt={vi.fn()}
      />,
    )

    const button = screen.getByRole('button', { name: 'Stornierung erteilen' })

    // Ohne Auswahl: der Grund nennt die fehlende Positionswahl, die Aktion sperrt.
    expect(button).toBeDisabled()
    expect(screen.getByText('Positionen auswählen')).toBeVisible()

    // Position gewählt, aber Kommentar fehlt noch: Der Positions-Grund
    // verschwindet, die Kommentar-Pflicht bleibt am Feld genannt.
    await user.click(screen.getByRole('button', { name: /hinzufügen/ }))
    expect(screen.queryByText('Positionen auswählen')).not.toBeInTheDocument()
    expect(screen.getByText(/Kommentar ist erforderlich/)).toBeVisible()
    expect(button).toBeDisabled()

    // Gültiger Kommentar: die Aktion wird frei.
    await user.type(
      screen.getByPlaceholderText('Kommentar (erforderlich)'),
      'Falsch bestellt',
    )
    expect(button).toBeEnabled()
  })

  it('behält die vorgangId über einen Retry und wechselt sie beim neuen Vorgang', async () => {
    const user = userEvent.setup()
    const stornierungErteilen = vi
      .fn<(s: StornierungErteilen) => Promise<void>>()
      .mockRejectedValueOnce(new Error('kaputt'))
      .mockResolvedValue(undefined)
    render(
      <HistorieStornierungDrawer
        backend={{ stornierungErteilen }}
        tisch={tisch}
        quelle={quelle}
        onClose={vi.fn()}
        onStornierungErteilt={vi.fn()}
      />,
    )

    const erteilen = () =>
      screen.getByRole('button', { name: 'Stornierung erteilen' })

    await user.click(screen.getByRole('button', { name: /hinzufügen/ }))
    await user.type(
      screen.getByPlaceholderText('Kommentar (erforderlich)'),
      'Falsch bestellt',
    )

    await user.click(erteilen())
    await waitFor(() => {
      expect(stornierungErteilen).toHaveBeenCalledTimes(1)
    })
    const ersterKey = stornierungErteilen.mock.calls[0][0].vorgangId
    expect(ersterKey).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/,
    )

    // Wiederholversuch desselben Vorgangs: derselbe Schlüssel.
    await user.click(erteilen())
    await waitFor(() => {
      expect(stornierungErteilen).toHaveBeenCalledTimes(2)
    })
    expect(stornierungErteilen.mock.calls[1][0].vorgangId).toBe(ersterKey)

    // Neuer logischer Vorgang: Auswahl leeren und neu füllen.
    await user.click(screen.getByRole('button', { name: /verringern/ }))
    await user.click(screen.getByRole('button', { name: /hinzufügen/ }))
    await user.click(erteilen())
    await waitFor(() => {
      expect(stornierungErteilen).toHaveBeenCalledTimes(3)
    })
    expect(stornierungErteilen.mock.calls[2][0].vorgangId).not.toBe(ersterKey)
  })
})
