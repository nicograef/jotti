import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Bestellung } from '../../table/Bestellung'
import type { Stornierung } from '../../table/Stornierung'
import type { Tisch } from '../../table/Tisch'
import type { Umbuchung } from '../../table/Umbuchung'
import type { Zahlung } from '../../table/Zahlung'
import { TischHistorie } from './TischHistorie'

type HistorieEintrag = Bestellung | Zahlung | Stornierung | Umbuchung

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn(), info: vi.fn() },
}))

vi.mock('@/lib/Auth', () => ({
  AuthSingleton: { canCancel: true, canRebook: true },
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

function position() {
  return {
    positionId: '00000000-0000-0000-0000-0000000000a1',
    varianteId: 1,
    produktName: 'Bratwurst',
    varianteName: 'Normal',
    kategorie: 'essen' as const,
    steuersatz: 'regel' as const,
    einzelpreisCents: 350,
    menge: 1,
    bestellerUserId: 1,
    bestellerName: 'Tester',
  }
}

function bestellung(overrides: Partial<Bestellung> = {}): Bestellung {
  return {
    art: 'bestellung',
    id: '00000000-0000-0000-0000-000000000001',
    userId: 1,
    userName: 'Tester',
    tischId: 1,
    positionen: [position()],
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
    positionen: [position()],
    gesamtZahlungCents: 350,
    kommentar: '',
    kassiertAm: '2026-06-18T12:05:00Z',
    ...overrides,
  }
}

function stornierung(overrides: Partial<Stornierung> = {}): Stornierung {
  return {
    art: 'stornierung',
    id: '00000000-0000-0000-0000-0000000000c1',
    userId: 3,
    userName: 'Clara',
    tischId: 1,
    positionen: [position()],
    gesamtStornierungCents: 350,
    kommentar: 'Falsch gebucht',
    barRueckgabe: true,
    storniertAm: '2026-06-18T12:10:00Z',
    ...overrides,
  }
}

function umbuchung(overrides: Partial<Umbuchung> = {}): Umbuchung {
  return {
    art: 'umbuchung',
    id: '00000000-0000-0000-0000-0000000000d1',
    userId: 4,
    userName: 'Dora',
    tischId: 1,
    quellTischId: 2,
    zielTischId: 1,
    positionen: [position()],
    gesamtCents: 350,
    kommentar: 'Umbuchung von Tisch 2',
    umgebuchtAm: '2026-06-18T12:15:00Z',
    stornierbarePositionen: [],
    umbuchbarePositionen: [],
    ...overrides,
  }
}

function renderHistorie(
  historie: HistorieEintrag[],
  backend: Partial<Parameters<typeof TischHistorie>[0]['backend']> = {},
) {
  render(
    <TischHistorie
      historie={historie}
      historieLoading={false}
      tisch={tisch}
      backend={{
        stornierungErteilen: vi.fn().mockResolvedValue(undefined),
        bestellungUmbuchen: vi.fn().mockResolvedValue(undefined),
        belegDrucken: vi.fn().mockResolvedValue('eingereiht'),
        stornobelegDrucken: vi.fn().mockResolvedValue('eingereiht'),
        ...backend,
      }}
      onStornierungErteilt={vi.fn()}
      onBestellungUmgebucht={vi.fn()}
    />,
  )
}

describe('TischHistorie', () => {
  it('beschriftet jede Zeile mit dem Namen der handelnden Servicekraft', () => {
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

    expect(screen.getByText(/· Anna/)).toBeInTheDocument()
    expect(screen.getByText(/· Bert/)).toBeInTheDocument()
  })

  it('zeigt Typ, farbcodierten Betrag und relative Zeit ohne Inline-Aktions-Buttons', () => {
    renderHistorie([bestellung()])

    expect(screen.getByText('Bestellung')).toBeInTheDocument()
    expect(screen.getByText('+3,50 €')).toBeInTheDocument()
    // Keine Inline-Aktionen mehr in der Zeile.
    expect(
      screen.queryByRole('button', { name: 'Stornieren' }),
    ).not.toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'Umbuchen' }),
    ).not.toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'Details anzeigen' }),
    ).not.toBeInTheDocument()
  })

  it('zeigt die Historie flach — alle Einträge ohne „Alle anzeigen"-Schalter', () => {
    renderHistorie([
      bestellung({ id: '00000000-0000-0000-0000-000000000001' }),
      bestellung({ id: '00000000-0000-0000-0000-000000000002' }),
      bestellung({ id: '00000000-0000-0000-0000-000000000003' }),
    ])

    expect(screen.getAllByText('Bestellung')).toHaveLength(3)
    expect(
      screen.queryByRole('button', { name: /Alle anzeigen/ }),
    ).not.toBeInTheDocument()
  })

  it('unterscheidet Warenrücknahme und geldneutrale Korrektur sichtbar', () => {
    renderHistorie([
      stornierung({
        id: '00000000-0000-0000-0000-0000000000c1',
        barRueckgabe: true,
      }),
      stornierung({
        id: '00000000-0000-0000-0000-0000000000c2',
        barRueckgabe: false,
        kommentar: '',
      }),
    ])

    expect(screen.getByText('Warenrücknahme')).toBeInTheDocument()
    expect(screen.getByText('Korrektur')).toBeInTheDocument()
  })

  it('bietet Stornieren und Umbuchen nur im Detail-Drawer an', () => {
    renderHistorie([
      bestellung({
        id: '00000000-0000-0000-0000-000000000001',
        stornierbarePositionen: [position()],
        umbuchbarePositionen: [position()],
      }),
    ])

    // In der Liste gibt es keine destruktiven Aktionen.
    expect(
      screen.queryByRole('button', { name: /Stornieren/ }),
    ).not.toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /Bestellung/ }))

    const dialog = screen.getByRole('dialog')
    expect(
      within(dialog).getByRole('button', { name: /Umbuchen/ }),
    ).toBeInTheDocument()
    expect(
      within(dialog).getByRole('button', { name: /Stornieren…/ }),
    ).toBeInTheDocument()
  })

  it('titelt den Detail-Drawer menschenlesbar statt mit UUID-Fragment', () => {
    renderHistorie([
      bestellung({
        id: '00000000-0000-0000-0000-000000000001',
        userName: 'Nico',
      }),
    ])

    fireEvent.click(screen.getByRole('button', { name: /Bestellung/ }))

    const dialog = screen.getByRole('dialog')
    expect(within(dialog).getByText(/^Bestellung ·/)).toHaveTextContent('Nico')
    expect(within(dialog).getByText(/Stammtisch ·/)).toBeInTheDocument()
    expect(screen.queryByText(/00000000/)).not.toBeInTheDocument()
  })

  it('zeigt den Stornobeleg-Button nur bei der Warenrücknahme im Drawer und löst ihn aus', async () => {
    const stornobelegDrucken = vi.fn().mockResolvedValue('eingereiht')
    renderHistorie(
      [
        stornierung({
          id: '00000000-0000-0000-0000-0000000000c1',
          barRueckgabe: true,
        }),
        stornierung({
          id: '00000000-0000-0000-0000-0000000000c2',
          barRueckgabe: false,
          kommentar: '',
        }),
      ],
      { stornobelegDrucken },
    )

    // Stornobeleg-Button ist vor dem Öffnen des Drawers nicht sichtbar
    expect(
      screen.queryByRole('button', { name: 'Stornobeleg drucken' }),
    ).not.toBeInTheDocument()

    // Drawer der Warenrücknahme (erster Eintrag) öffnen
    fireEvent.click(screen.getByText('Warenrücknahme'))

    // Stornobeleg-Button im Drawer sichtbar und auslösbar
    const belegButton = screen.getByRole('button', {
      name: 'Stornobeleg drucken',
    })
    fireEvent.click(belegButton)

    await waitFor(() => {
      expect(stornobelegDrucken).toHaveBeenCalledWith(
        1,
        '00000000-0000-0000-0000-0000000000c1',
      )
    })
  })

  it('nutzt bei Umbuchungen den Richtungs-Autotext als Titel — in Zeile und Detail — ohne ihn als Kommentar auszugeben', () => {
    renderHistorie([
      umbuchung({
        id: '00000000-0000-0000-0000-0000000000d1',
        kommentar: 'Umbuchung von Tisch 2',
        gesamtCents: 350,
      }),
    ])

    // Zeile: Autotext als Titel, Zugang mit +-Betrag.
    expect(screen.getByText('Umbuchung von Tisch 2')).toBeInTheDocument()
    expect(screen.getByText('+3,50 €')).toBeInTheDocument()

    fireEvent.click(
      screen.getByRole('button', { name: /Umbuchung von Tisch 2/ }),
    )

    // Detail: Titel ist der Autotext (nicht generisch „Umbuchung"); der Autotext
    // erscheint nicht als Kommentar-Textfeld.
    const dialog = screen.getByRole('dialog')
    expect(
      within(dialog).getByText(/^Umbuchung von Tisch 2 ·/),
    ).toBeInTheDocument()
    expect(
      within(dialog).queryByDisplayValue('Umbuchung von Tisch 2'),
    ).not.toBeInTheDocument()
  })

  it('bietet aus einem Umbuchungs-Zugang Stornieren und Umbuchen im Detail an', () => {
    renderHistorie([
      umbuchung({
        id: '00000000-0000-0000-0000-0000000000d1',
        stornierbarePositionen: [position()],
        umbuchbarePositionen: [position()],
      }),
    ])

    fireEvent.click(
      screen.getByRole('button', { name: /Umbuchung von Tisch 2/ }),
    )

    const dialog = screen.getByRole('dialog')
    expect(
      within(dialog).getByRole('button', { name: /Umbuchen/ }),
    ).toBeInTheDocument()
    expect(
      within(dialog).getByRole('button', { name: /Stornieren…/ }),
    ).toBeInTheDocument()
  })

  it('rendert eine geldneutrale Korrektur mit leerem Kommentar ohne Fehler', () => {
    renderHistorie([
      stornierung({
        id: '00000000-0000-0000-0000-0000000000c2',
        barRueckgabe: false,
        kommentar: '',
      }),
    ])

    expect(screen.getByText('Korrektur')).toBeInTheDocument()

    // Drawer öffnen: auch dort kein Stornobeleg-Button für eine Korrektur
    fireEvent.click(screen.getByText('Korrektur'))
    expect(
      screen.queryByRole('button', { name: 'Stornobeleg drucken' }),
    ).not.toBeInTheDocument()
  })
})
