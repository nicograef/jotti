import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { toast } from 'sonner'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Bestellung, Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import type { BestellungUmbuchen } from '../../table/Umbuchung'
import { HistorieUmbuchungDrawer } from './HistorieUmbuchungDrawer'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

const tischeState = vi.hoisted(() => ({
  isLoadingError: false,
  refetch: vi.fn(),
}))

vi.mock('../../table/hooks', () => ({
  useAktiveTische: () => ({
    tische: tischeState.isLoadingError
      ? []
      : [
          { id: 1, name: 'Stammtisch', saldoCents: 0 },
          { id: 2, name: 'Nebentisch', saldoCents: 0 },
          { id: 3, name: 'Ecktisch', saldoCents: 0 },
        ],
    isPending: false,
    isLoadingError: tischeState.isLoadingError,
    refetch: tischeState.refetch,
  }),
}))

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
  tischeState.isLoadingError = false
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

  // A2: Der Erfolg meldet den Namen des Ziel-Tischs für den Erfolgs-Pop; der
  // frühere „Bestellung umgebucht."-Toast entfällt.
  it('meldet den Ziel-Tischnamen an den Aufrufer und zeigt keinen Toast', async () => {
    const user = userEvent.setup()
    const onBestellungUmgebucht = vi.fn()
    render(
      <HistorieUmbuchungDrawer
        backend={{ bestellungUmbuchen: vi.fn().mockResolvedValue(undefined) }}
        tisch={tisch}
        quelle={quelle}
        onClose={vi.fn()}
        onBestellungUmgebucht={onBestellungUmgebucht}
      />,
    )

    await user.click(
      screen.getByRole('button', { name: /^1 Position auswählen/ }),
    )
    await user.selectOptions(screen.getByRole('combobox'), 'Nebentisch')
    await user.click(
      screen.getByRole('button', { name: 'Umbuchung ausführen' }),
    )

    await waitFor(() => {
      expect(onBestellungUmgebucht).toHaveBeenCalledWith('Nebentisch')
    })
    expect(toast.success).not.toHaveBeenCalledWith('Bestellung umgebucht.')
  })

  it('behält die vorgangId über einen Retry und wechselt sie beim neuen Vorgang', async () => {
    const user = userEvent.setup()
    const bestellungUmbuchen = vi
      .fn<(u: BestellungUmbuchen) => Promise<void>>()
      .mockRejectedValueOnce(new Error('kaputt'))
      .mockResolvedValue(undefined)
    render(
      <HistorieUmbuchungDrawer
        backend={{ bestellungUmbuchen }}
        tisch={tisch}
        quelle={quelle}
        onClose={vi.fn()}
        onBestellungUmgebucht={vi.fn()}
      />,
    )

    const ausfuehren = () =>
      screen.getByRole('button', { name: 'Umbuchung ausführen' })

    await user.click(
      screen.getByRole('button', { name: /^1 Position auswählen/ }),
    )
    await user.selectOptions(screen.getByRole('combobox'), 'Nebentisch')

    await user.click(ausfuehren())
    await waitFor(() => {
      expect(bestellungUmbuchen).toHaveBeenCalledTimes(1)
    })
    const ersterKey = bestellungUmbuchen.mock.calls[0][0].vorgangId
    expect(ersterKey).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/,
    )

    // Wiederholversuch desselben Vorgangs: derselbe Schlüssel.
    await user.click(ausfuehren())
    await waitFor(() => {
      expect(bestellungUmbuchen).toHaveBeenCalledTimes(2)
    })
    expect(bestellungUmbuchen.mock.calls[1][0].vorgangId).toBe(ersterKey)

    // Neuer logischer Vorgang: Auswahl leeren und neu füllen.
    await user.click(screen.getByRole('button', { name: 'Auswahl aufheben' }))
    await user.click(
      screen.getByRole('button', { name: /^1 Position auswählen/ }),
    )
    await user.click(ausfuehren())
    await waitFor(() => {
      expect(bestellungUmbuchen).toHaveBeenCalledTimes(3)
    })
    expect(bestellungUmbuchen.mock.calls[2][0].vorgangId).not.toBe(ersterKey)
  })

  // Scheitert der erste Versuch scheinbar (Antwort im WLAN verloren) und ändert
  // sich die Auswahl danach, muss der zweite Versuch denselben Schlüssel tragen:
  // Nur dann erkennt der Server den Konflikt mit der bereits gebuchten Umbuchung
  // und meldet ihn, statt ein zweites Mal umzubuchen.
  it('behält die vorgangId, wenn sich die Auswahl nach einem Fehlversuch ändert', async () => {
    const user = userEvent.setup()
    const bestellungUmbuchen = vi
      .fn<(u: BestellungUmbuchen) => Promise<void>>()
      .mockRejectedValueOnce(new Error('kaputt'))
      .mockResolvedValue(undefined)
    render(
      <HistorieUmbuchungDrawer
        backend={{ bestellungUmbuchen }}
        tisch={tisch}
        quelle={quelle}
        onClose={vi.fn()}
        onBestellungUmgebucht={vi.fn()}
      />,
    )

    const ausfuehren = () =>
      screen.getByRole('button', { name: 'Umbuchung ausführen' })

    await user.click(
      screen.getByRole('button', { name: /^1 Position auswählen/ }),
    )
    await user.selectOptions(screen.getByRole('combobox'), 'Nebentisch')
    await user.click(ausfuehren())
    await waitFor(() => {
      expect(bestellungUmbuchen).toHaveBeenCalledTimes(1)
    })
    const ersterKey = bestellungUmbuchen.mock.calls[0][0].vorgangId

    // Geänderte Nutzdaten nach dem Fehlversuch (Menge 2 → 1): derselbe Vorgang,
    // derselbe Schlüssel — die Abweichung beanstandet der Server.
    await user.click(screen.getByRole('button', { name: /verringern/ }))
    await user.click(ausfuehren())
    await waitFor(() => {
      expect(bestellungUmbuchen).toHaveBeenCalledTimes(2)
    })
    const zweiterAufruf = bestellungUmbuchen.mock.calls[1][0]
    expect(zweiterAufruf.vorgangId).toBe(ersterKey)
    expect(zweiterAufruf.positionen).toEqual([
      { positionId: position.positionId, menge: 1 },
    ])
  })

  it('behält die vorgangId, wenn nach einem Fehlversuch der Ziel-Tisch wechselt', async () => {
    const user = userEvent.setup()
    const bestellungUmbuchen = vi
      .fn<(u: BestellungUmbuchen) => Promise<void>>()
      .mockRejectedValueOnce(new Error('kaputt'))
      .mockResolvedValue(undefined)
    render(
      <HistorieUmbuchungDrawer
        backend={{ bestellungUmbuchen }}
        tisch={tisch}
        quelle={quelle}
        onClose={vi.fn()}
        onBestellungUmgebucht={vi.fn()}
      />,
    )

    const ausfuehren = () =>
      screen.getByRole('button', { name: 'Umbuchung ausführen' })

    await user.click(
      screen.getByRole('button', { name: /^1 Position auswählen/ }),
    )
    await user.selectOptions(screen.getByRole('combobox'), 'Nebentisch')
    await user.click(ausfuehren())
    await waitFor(() => {
      expect(bestellungUmbuchen).toHaveBeenCalledTimes(1)
    })
    const ersterKey = bestellungUmbuchen.mock.calls[0][0].vorgangId

    // Der Ziel-Tisch gehört zu den Nutzdaten und geht damit in den
    // serverseitigen Hash ein. Genau deshalb bleibt der Schlüssel gleich: Lag
    // die Umbuchung auf den Nebentisch bereits vor, beanstandet der Server den
    // Wechsel auf den Ecktisch, statt ihn ein zweites Mal zu buchen.
    await user.selectOptions(screen.getByRole('combobox'), 'Ecktisch')
    await user.click(ausfuehren())
    await waitFor(() => {
      expect(bestellungUmbuchen).toHaveBeenCalledTimes(2)
    })
    const zweiterAufruf = bestellungUmbuchen.mock.calls[1][0]
    expect(zweiterAufruf.vorgangId).toBe(ersterKey)
    expect(zweiterAufruf.zielTischId).toBe(3)
  })

  // Der leere Select behauptet „Kein aktiver Ziel-Tisch verfügbar" — bei einem
  // Ladefehler ist das falsch: Die Tische gibt es, nur die Liste kam nicht an.
  it('zeigt bei gescheitertem Erstladen der Tische einen Ladefehler statt des Select-Platzhalters', async () => {
    tischeState.isLoadingError = true
    const user = userEvent.setup()
    renderDrawer()

    expect(
      screen.getByText('Tische konnten nicht geladen werden'),
    ).toBeInTheDocument()
    expect(
      screen.queryByText('Kein aktiver Ziel-Tisch verfügbar'),
    ).not.toBeInTheDocument()
    expect(screen.queryByRole('combobox')).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Erneut versuchen' }))
    expect(tischeState.refetch).toHaveBeenCalled()
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
