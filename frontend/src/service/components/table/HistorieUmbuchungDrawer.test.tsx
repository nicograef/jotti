import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Bestellung, Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import { HistorieUmbuchungDrawer } from './HistorieUmbuchungDrawer'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

vi.mock('../../table/hooks', () => ({
  useAktiveTische: () => ({
    tische: [
      { id: 1, name: 'Stammtisch', saldoCents: 0 },
      { id: 2, name: 'Nebentisch', saldoCents: 0 },
    ],
    isPending: false,
  }),
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

function renderDrawer(
  bestellungUmbuchen = vi.fn().mockResolvedValue(undefined),
) {
  render(
    <HistorieUmbuchungDrawer
      backend={{ bestellungUmbuchen }}
      tisch={tisch}
      quelle={quelle}
      onClose={vi.fn()}
      onBestellungUmgebucht={vi.fn()}
    />,
  )
  return bestellungUmbuchen
}

describe('HistorieUmbuchungDrawer', () => {
  it('rendert Positionsliste im DrawerBody; Ziel-Tisch-Auswahl und Buttons im sichtbaren Footer', () => {
    renderDrawer()

    const dialog = screen.getByRole('dialog')
    const body = dialog.querySelector('[data-slot="drawer-body"]')
    const footer = dialog.querySelector('[data-slot="drawer-footer"]')
    expect(body).not.toBeNull()
    expect(footer).not.toBeNull()
    expect(body).toContainElement(screen.getByText(/Bratwurst/))
    // Die Ziel-Tisch-Auswahl (Pflichtfeld) steht im nicht-scrollenden Footer.
    const select = screen.getByRole('combobox')
    expect(footer).toContainElement(select)
    expect(body).not.toContainElement(select)
    expect(body).not.toContainElement(
      screen.getByRole('button', { name: 'Umbuchung ausführen' }),
    )
    expect(body).not.toContainElement(
      screen.getByRole('button', { name: 'Abbrechen' }),
    )
  })

  it('startet mit leerer Auswahl und sperrt, bis Positionen und Ziel-Tisch gewählt sind', async () => {
    const user = userEvent.setup()
    const bestellungUmbuchen = renderDrawer()

    const button = screen.getByRole('button', { name: 'Umbuchung ausführen' })

    // Ohne Wahl steht der Placeholder, nicht ein vorbelegter Tisch.
    const select = screen.getByRole('combobox')
    expect(select).toHaveValue('')
    expect(screen.getByText('Ziel-Tisch wählen…')).toBeInTheDocument()

    // Ziel-Tisch allein reicht nicht: ohne ausgewählte Positionen bleibt gesperrt.
    await user.selectOptions(select, 'Nebentisch')
    expect(button).toBeDisabled()

    // „Alle auswählen" wählt die volle umbuchbare Menge und gibt den Button frei.
    await user.click(
      screen.getByRole('button', { name: /^1 Position auswählen/ }),
    )
    expect(button).toBeEnabled()

    await user.click(button)
    expect(bestellungUmbuchen).toHaveBeenCalledWith(
      expect.objectContaining({
        quellTischId: 1,
        zielTischId: 2,
        positionen: [{ positionId: position.positionId, menge: 2 }],
      }),
    )
  })

  it('leert die Auswahl beim zweiten Tap auf „Alle auswählen"', async () => {
    const user = userEvent.setup()
    renderDrawer()

    const button = screen.getByRole('button', { name: 'Umbuchung ausführen' })
    await user.selectOptions(screen.getByRole('combobox'), 'Nebentisch')

    await user.click(
      screen.getByRole('button', { name: /^1 Position auswählen/ }),
    )
    expect(button).toBeEnabled()

    await user.click(screen.getByRole('button', { name: 'Auswahl aufheben' }))
    expect(button).toBeDisabled()
  })

  it('titelt menschenlesbar mit Vorgangstyp und Name statt UUID-Fragment', () => {
    renderDrawer()

    const title = screen.getByText(/^Bestellung ·/)
    expect(title).toHaveTextContent('Nico')
    expect(screen.queryByText(/00000000/)).not.toBeInTheDocument()
  })

  it('nennt den Sperrgrund neben der Aktion und gibt frei, sobald er erfüllt ist', async () => {
    const user = userEvent.setup()
    renderDrawer()

    const button = screen.getByRole('button', { name: 'Umbuchung ausführen' })

    // Ohne Auswahl: der Grund nennt die fehlende Positionswahl.
    expect(button).toBeDisabled()
    expect(screen.getByText('Positionen auswählen')).toBeVisible()

    // Positionen gewählt, aber Ziel-Tisch fehlt: der Grund wechselt.
    await user.click(
      screen.getByRole('button', { name: /^1 Position auswählen/ }),
    )
    expect(screen.queryByText('Positionen auswählen')).not.toBeInTheDocument()
    expect(screen.getByText('Ziel-Tisch wählen')).toBeVisible()
    expect(button).toBeDisabled()

    // Ziel-Tisch gewählt: der Grund verschwindet, die Aktion wird frei.
    await user.selectOptions(screen.getByRole('combobox'), 'Nebentisch')
    expect(screen.queryByText('Ziel-Tisch wählen')).not.toBeInTheDocument()
    expect(button).toBeEnabled()
  })

  it('reicht ein optionales Kommentar an die Umbuchung durch', async () => {
    const user = userEvent.setup()
    const bestellungUmbuchen = renderDrawer()

    await user.type(
      screen.getByPlaceholderText('Kommentar (optional)'),
      'Gast gewechselt',
    )
    await user.click(
      screen.getByRole('button', { name: /^1 Position auswählen/ }),
    )
    await user.selectOptions(screen.getByRole('combobox'), 'Nebentisch')
    await user.click(
      screen.getByRole('button', { name: 'Umbuchung ausführen' }),
    )

    expect(bestellungUmbuchen).toHaveBeenCalledWith(
      expect.objectContaining({
        quellTischId: 1,
        zielTischId: 2,
        benutzerKommentar: 'Gast gewechselt',
      }),
    )
  })

  it('beschriftet den Sammel-Button bei mehreren Positionen im Plural', () => {
    const zweite: Position = {
      ...position,
      positionId: '00000000-0000-0000-0000-000000000002',
      produktName: 'Pommes',
      einzelpreisCents: 250,
      menge: 1,
    }
    render(
      <HistorieUmbuchungDrawer
        backend={{ bestellungUmbuchen: vi.fn().mockResolvedValue(undefined) }}
        tisch={tisch}
        quelle={{ ...quelle, umbuchbarePositionen: [position, zweite] }}
        onClose={vi.fn()}
        onBestellungUmgebucht={vi.fn()}
      />,
    )

    expect(
      screen.getByRole('button', { name: /^Alle 2 Positionen auswählen/ }),
    ).toBeInTheDocument()
  })
})
