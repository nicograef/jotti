import {
  cleanup,
  fireEvent,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { toast } from 'sonner'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'
import { signIn, signOut } from '@/test/auth'
import { FakeBackend } from '@/test/FakeBackend'
import { renderWithBackend } from '@/test/render'

import type { Bestellung } from './Bestellung'
import type { Stornierung } from './Stornierung'
import type { Tisch } from './Tisch'
import { TischBackend } from './TischBackend'
import { TischHistorie } from './TischHistorie'
import type { Umbuchung } from './Umbuchung'
import type { Zahlung } from './Zahlung'

type HistorieEintrag = Bestellung | Zahlung | Stornierung | Umbuchung

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn(), info: vi.fn() },
}))

beforeEach(() => {
  VorgangsRegisterSingleton.zuruecksetzen()
  // Serviceleitung may cancel and rebook.
  signIn({ role: 'serviceleitung' })
})

afterEach(() => {
  cleanup()
  signOut()
})

const tisch: Tisch = { id: 1, name: 'Stammtisch', saldoCents: 0 }

function position() {
  return {
    positionId: '00000000-0000-4000-8000-0000000000a1',
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
    id: '00000000-0000-4000-8000-000000000001',
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
    id: '00000000-0000-4000-8000-0000000000f1',
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
    id: '00000000-0000-4000-8000-0000000000c1',
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
    id: '00000000-0000-4000-8000-0000000000d1',
    userId: 4,
    userName: 'Dora',
    tischId: 1,
    quellTischId: 2,
    zielTischId: 1,
    positionen: [position()],
    gesamtCents: 350,
    kommentar: 'Umbuchung von Tisch 2',
    benutzerKommentar: '',
    umgebuchtAm: '2026-06-18T12:15:00Z',
    stornierbarePositionen: [],
    umbuchbarePositionen: [],
    ...overrides,
  }
}

// The rebooking drawer loads its target tables itself.
function backend(): FakeBackend {
  return new FakeBackend()
    .respond('service/get-aktive-tische', {
      tische: [
        { id: 1, name: 'Stammtisch', saldoCents: 0 },
        { id: 2, name: 'Nebentisch', saldoCents: 0 },
      ],
    })
    .respond('serviceleitung/stornierung-erteilen', {})
    .respond('service/bestellung-umbuchen', {})
    .respond('service/beleg-drucken', { status: 'eingereiht' })
}

function renderHistorie(
  historie: HistorieEintrag[],
  fake: FakeBackend = backend(),
  onErfolg: (nachricht: string) => void = vi.fn(),
) {
  renderWithBackend(
    <TischHistorie
      historie={historie}
      historieLoading={false}
      tisch={tisch}
      backend={new TischBackend(fake)}
      onErfolg={onErfolg}
    />,
    fake,
  )
}

describe('TischHistorie', () => {
  it('beschriftet jede Zeile mit dem Namen der handelnden Servicekraft', () => {
    renderHistorie([
      bestellung({
        id: '00000000-0000-4000-8000-000000000001',
        userName: 'Anna',
      }),
      zahlung({
        id: '00000000-0000-4000-8000-0000000000f1',
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
      bestellung({ id: '00000000-0000-4000-8000-000000000001' }),
      bestellung({ id: '00000000-0000-4000-8000-000000000002' }),
      bestellung({ id: '00000000-0000-4000-8000-000000000003' }),
    ])

    expect(screen.getAllByText('Bestellung')).toHaveLength(3)
    expect(
      screen.queryByRole('button', { name: /Alle anzeigen/ }),
    ).not.toBeInTheDocument()
  })

  it('unterscheidet Warenrücknahme und geldneutrale Korrektur sichtbar', () => {
    renderHistorie([
      stornierung({
        id: '00000000-0000-4000-8000-0000000000c1',
        barRueckgabe: true,
      }),
      stornierung({
        id: '00000000-0000-4000-8000-0000000000c2',
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
        id: '00000000-0000-4000-8000-000000000001',
        stornierbarePositionen: [position()],
        umbuchbarePositionen: [position()],
      }),
    ])

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
        id: '00000000-0000-4000-8000-000000000001',
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
    const fake = backend()
    renderHistorie(
      [
        stornierung({
          id: '00000000-0000-4000-8000-0000000000c1',
          barRueckgabe: true,
        }),
        stornierung({
          id: '00000000-0000-4000-8000-0000000000c2',
          barRueckgabe: false,
          kommentar: '',
        }),
      ],
      fake,
    )

    expect(
      screen.queryByRole('button', { name: 'Stornobeleg drucken' }),
    ).not.toBeInTheDocument()

    fireEvent.click(screen.getByText('Warenrücknahme'))

    const belegButton = screen.getByRole('button', {
      name: 'Stornobeleg drucken',
    })
    fireEvent.click(belegButton)

    await waitFor(() => {
      expect(fake.bodies('service/beleg-drucken')).toEqual([
        { tischId: 1, stornierungId: '00000000-0000-4000-8000-0000000000c1' },
      ])
    })
  })

  it('bietet im Detail einer Zahlung den Gäste-Beleg als „Kassenbeleg drucken" an', async () => {
    const fake = backend()
    renderHistorie(
      [zahlung({ id: '00000000-0000-4000-8000-0000000000f1' })],
      fake,
    )

    fireEvent.click(screen.getByRole('button', { name: /Zahlung/ }))

    // Der Gäste-Beleg heißt „Kassenbeleg", nicht generisch „Beleg".
    const belegButton = screen.getByRole('button', {
      name: 'Kassenbeleg drucken',
    })
    fireEvent.click(belegButton)

    await waitFor(() => {
      expect(fake.bodies('service/beleg-drucken')).toEqual([
        { tischId: 1, zahlungId: '00000000-0000-4000-8000-0000000000f1' },
      ])
    })
  })

  it('nutzt bei Umbuchungen den Richtungs-Autotext als Titel — in Zeile und Detail — ohne ihn als Kommentar auszugeben', () => {
    renderHistorie([
      umbuchung({
        id: '00000000-0000-4000-8000-0000000000d1',
        kommentar: 'Umbuchung von Tisch 2',
        gesamtCents: 350,
      }),
    ])

    expect(screen.getByText('Umbuchung von Tisch 2')).toBeInTheDocument()
    expect(screen.getByText('+3,50 €')).toBeInTheDocument()

    fireEvent.click(
      screen.getByRole('button', { name: /Umbuchung von Tisch 2/ }),
    )

    const dialog = screen.getByRole('dialog')
    expect(
      within(dialog).getByText(/^Umbuchung von Tisch 2 ·/),
    ).toBeInTheDocument()
    expect(
      within(dialog).queryByDisplayValue('Umbuchung von Tisch 2'),
    ).not.toBeInTheDocument()
  })

  it('zeigt das Benutzerkommentar einer Umbuchung in Anführungszeichen in Unterzeile und Detail — Titel bleibt der Autotext', () => {
    renderHistorie([
      umbuchung({
        id: '00000000-0000-4000-8000-0000000000d1',
        kommentar: 'Umbuchung von Tisch 2',
        benutzerKommentar: 'Gast gewechselt',
      }),
    ])

    expect(screen.getByText('Umbuchung von Tisch 2')).toBeInTheDocument()
    expect(screen.getByText(/„Gast gewechselt“/)).toBeInTheDocument()

    fireEvent.click(
      screen.getByRole('button', { name: /Umbuchung von Tisch 2/ }),
    )

    const dialog = screen.getByRole('dialog')
    expect(
      within(dialog).getByText(/^Umbuchung von Tisch 2 ·/),
    ).toBeInTheDocument()
    expect(
      within(dialog).getByDisplayValue('Gast gewechselt'),
    ).toBeInTheDocument()
  })

  it('bietet aus einem Umbuchungs-Zugang Stornieren und Umbuchen im Detail an', () => {
    renderHistorie([
      umbuchung({
        id: '00000000-0000-4000-8000-0000000000d1',
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

  // Storno und Umbuchung melden den Erfolg als Pop-Text an den Aufrufer; der
  // Drawer schließt, der Refetch läuft beim Pop-Schließen in TablePage.
  it('meldet den Storno-Erfolg über den Pop-Text und schließt den Drawer', async () => {
    const user = userEvent.setup()
    const onErfolg = vi.fn()
    renderHistorie(
      [
        bestellung({
          id: '00000000-0000-4000-8000-000000000001',
          stornierbarePositionen: [position()],
        }),
      ],
      backend(),
      onErfolg,
    )

    await user.click(screen.getByRole('button', { name: /Bestellung/ }))
    await user.click(screen.getByRole('button', { name: /Stornieren…/ }))
    await user.click(screen.getByRole('button', { name: /hinzufügen/ }))
    await user.type(
      screen.getByPlaceholderText('Kommentar (erforderlich)'),
      'Falsch gebucht',
    )
    await user.click(
      screen.getByRole('button', { name: 'Stornierung erteilen' }),
    )

    await waitFor(() => {
      expect(onErfolg).toHaveBeenCalledWith('Stornierung gebucht.')
    })
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('meldet die Umbuchung mit dem Ziel-Tischnamen über den Pop-Text — ohne Toast', async () => {
    const user = userEvent.setup()
    const onErfolg = vi.fn()
    renderHistorie(
      [
        bestellung({
          id: '00000000-0000-4000-8000-000000000001',
          umbuchbarePositionen: [position()],
        }),
      ],
      backend(),
      onErfolg,
    )

    await user.click(screen.getByRole('button', { name: /Bestellung/ }))
    await user.click(screen.getByRole('button', { name: /Umbuchen/ }))
    await user.click(
      screen.getByRole('button', { name: /Position(?:en)? auswählen/ }),
    )
    await user.selectOptions(screen.getByRole('combobox'), 'Nebentisch')
    await user.click(
      screen.getByRole('button', { name: 'Umbuchung ausführen' }),
    )

    await waitFor(() => {
      expect(onErfolg).toHaveBeenCalledWith('Auf Nebentisch umgebucht.')
    })
    expect(toast.success).not.toHaveBeenCalledWith('Bestellung umgebucht.')
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('rendert eine geldneutrale Korrektur mit leerem Kommentar ohne Fehler', () => {
    renderHistorie([
      stornierung({
        id: '00000000-0000-4000-8000-0000000000c2',
        barRueckgabe: false,
        kommentar: '',
      }),
    ])

    expect(screen.getByText('Korrektur')).toBeInTheDocument()

    fireEvent.click(screen.getByText('Korrektur'))
    expect(
      screen.queryByRole('button', { name: 'Stornobeleg drucken' }),
    ).not.toBeInTheDocument()
  })
})

describe('TischHistorie im Vorgangs-Register', () => {
  it('meldet das geöffnete Detail als reine Anzeige nicht', () => {
    renderHistorie([
      bestellung({
        id: '00000000-0000-4000-8000-000000000001',
        stornierbarePositionen: [position()],
        umbuchbarePositionen: [position()],
      }),
    ])

    fireEvent.click(screen.getByRole('button', { name: /Bestellung/ }))

    // Das Detail zeigt nur an; erst Stornieren oder Umbuchen wird zur Arbeit.
    expect(screen.getByRole('dialog')).toBeInTheDocument()
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })
})
