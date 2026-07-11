import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Bestellung } from '../../table/Bestellung'
import type { Stornierung } from '../../table/Stornierung'
import type { Tisch } from '../../table/Tisch'
import type { Zahlung } from '../../table/Zahlung'
import { TischHistorie } from './TischHistorie'

type HistorieEintrag = Bestellung | Zahlung | Stornierung

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn(), info: vi.fn() },
}))

vi.mock('@/lib/Auth', () => ({
  AuthSingleton: { canCancel: true },
}))

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
        einzelpreisCents: 350,
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
        einzelpreisCents: 350,
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

function stornierung(overrides: Partial<Stornierung> = {}): Stornierung {
  return {
    art: 'stornierung',
    id: '00000000-0000-0000-0000-0000000000c1',
    userId: 3,
    userName: 'Clara',
    tischId: 1,
    positionen: [
      {
        positionId: '00000000-0000-0000-0000-0000000000a1',
        varianteId: 1,
        produktName: 'Bratwurst',
        varianteName: 'Normal',
        kategorie: 'essen',
        steuersatz: 'regel',
        einzelpreisCents: 350,
        menge: 1,
        bestellerUserId: 1,
        bestellerName: 'Anna',
      },
    ],
    gesamtStornierungCents: 350,
    kommentar: 'Falsch gebucht',
    barRueckgabe: true,
    storniertAm: '2026-06-18T12:10:00Z',
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

    expect(screen.getByText(/Warenrücknahme -/)).toBeInTheDocument()
    expect(screen.getByText(/Korrektur -/)).toBeInTheDocument()
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
    const detailButtons = screen.getAllByRole('button', {
      name: 'Details anzeigen',
    })
    fireEvent.click(detailButtons[0])

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

  it('rendert eine geldneutrale Korrektur mit leerem Kommentar ohne Fehler', () => {
    renderHistorie([
      stornierung({
        id: '00000000-0000-0000-0000-0000000000c2',
        barRueckgabe: false,
        kommentar: '',
      }),
    ])

    expect(screen.getByText(/Korrektur -/)).toBeInTheDocument()

    // Drawer öffnen: auch dort kein Stornobeleg-Button für eine Korrektur
    fireEvent.click(screen.getByRole('button', { name: 'Details anzeigen' }))
    expect(
      screen.queryByRole('button', { name: 'Stornobeleg drucken' }),
    ).not.toBeInTheDocument()
  })
})
