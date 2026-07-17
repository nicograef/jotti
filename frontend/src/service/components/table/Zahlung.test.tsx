import { cleanup, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ComponentProps } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { useMengen } from '@/hooks/use-mengen'
import { useIsMobile } from '@/hooks/use-mobile'

import type { Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import { ServiceDock } from '../ServiceDock'
import { Zahlung } from './Zahlung'

// Die Kassieren-Auswahl liegt seit A1 in TablePage; für die isolierten
// Komponenten-Tests stellt dieser Harness die gehobene Steuerung inklusive der
// Deckelung auf die unbezahlte Menge bereit.
function ZahlungHarness(
  props: Omit<ComponentProps<typeof Zahlung>, 'mengenSteuerung'>,
) {
  const unbezahlteMengen: Record<string, number> = {}
  props.positionen.forEach((position) => {
    unbezahlteMengen[position.positionId] = position.menge
  })
  const mengenSteuerung = useMengen<string>(
    (positionId) => unbezahlteMengen[positionId] || 0,
  )
  return <Zahlung {...props} mengenSteuerung={mengenSteuerung} />
}

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

// Standardmäßig Handy-Layout (Dock-Button, Restbetrag im Dock-Slot); ein Test
// unten schaltet auf Desktop, um die Verdrahtung der festen Spalte zu prüfen.
// Deren container-neutrales Verhalten deckt ZahlungAbschluss.test.tsx ab.
vi.mock('@/hooks/use-mobile', () => ({
  useIsMobile: vi.fn(() => true),
}))

vi.mock('@/lib/Auth', () => ({
  AuthSingleton: { userId: 1 },
}))

afterEach(() => {
  cleanup()
  vi.mocked(useIsMobile).mockReturnValue(true)
})

const tisch: Tisch = { id: 1, name: 'Stammtisch', saldoCents: 900 }

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

// Position einer anderen Servicekraft (bestellerUserId ≠ 1).
const fremdePosition: Position = {
  positionId: '00000000-0000-0000-0000-000000000002',
  varianteId: 2,
  produktName: 'Pommes',
  varianteName: 'Normal',
  kategorie: 'essen',
  steuersatz: 'regel',
  einzelpreisCents: 250,
  menge: 1,
  bestellerUserId: 2,
  bestellerName: 'Kollegin',
}

function renderZahlung(positionen: Position[] = [position]) {
  render(
    <ServiceDock leiste={null}>
      <ZahlungHarness
        backend={{
          zahlungKassieren: vi.fn().mockResolvedValue(undefined),
        }}
        tisch={tisch}
        positionen={positionen}
        onErfolg={vi.fn()}
      />
    </ServiceDock>,
  )
}

describe('Zahlung feste Spalte (ab lg)', () => {
  it('rendert die Abschluss-Spalte mit Restbetrag statt Dock und Drawer', async () => {
    vi.mocked(useIsMobile).mockReturnValue(false)
    const user = userEvent.setup()
    // Kein ServiceDock: die feste Spalte trägt Aktionsbutton und Restbetrag.
    render(
      <ZahlungHarness
        backend={{ zahlungKassieren: vi.fn().mockResolvedValue(undefined) }}
        tisch={tisch}
        positionen={[position]}
        onErfolg={vi.fn()}
      />,
    )

    expect(screen.getByText('Bratwurst Normal')).toBeInTheDocument()
    // Restbetrag steht in der Spalte (nicht im Dock-Slot).
    expect(screen.getByText('Nach dieser Zahlung noch offen')).toBeVisible()
    const button = screen.getByRole('button', { name: 'Kassieren' })
    expect(button).toBeDisabled()

    // Auswahl links aktiviert den Kassieren-Button rechts.
    await user.click(screen.getByRole('button', { name: 'Produkt hinzufügen' }))
    expect(screen.getByRole('button', { name: 'Kassieren' })).toBeEnabled()
  })
})

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
})

describe('Zahlung Positionsgruppen', () => {
  it('zeigt eigene Positionen direkt und fremde erst nach dem Aufklappen von „Von anderen"', async () => {
    const user = userEvent.setup()
    renderZahlung([position, fremdePosition])

    expect(screen.getByText('Bratwurst Normal')).toBeInTheDocument()
    // Eingeklappt: Summen-/Namenszeile statt einzelner Positionsstepper; die
    // Besteller-Angabe „von Kollegin" erscheint erst nach dem Aufklappen.
    expect(screen.getByText(/Pommes Normal · 2,50/)).toBeInTheDocument()
    expect(screen.queryByText('von Kollegin')).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /Von anderen · 1/ }))

    expect(screen.getByText('von Kollegin')).toBeInTheDocument()
  })

  it('zeigt eine getroffene Fremd-Auswahl auch in der eingeklappten Summenzeile', async () => {
    const user = userEvent.setup()
    renderZahlung([position, fremdePosition])

    // Aufklappen, fremde Position wählen, wieder einklappen. Der „+"-Button der
    // fremden Zeile wird über ihre Item-Zeile (enthält „von Kollegin")
    // eingegrenzt, um ihn vom „+" der eigenen Zeile zu trennen.
    await user.click(screen.getByRole('button', { name: /Von anderen · 1/ }))
    const fremdeZeile = screen
      .getByText('von Kollegin')
      .closest<HTMLElement>('[data-slot="item"]')
    if (fremdeZeile === null) throw new Error('Fremd-Zeile nicht gefunden')
    await user.click(
      within(fremdeZeile).getByRole('button', { name: 'Produkt hinzufügen' }),
    )
    await user.click(screen.getByRole('button', { name: /Von anderen · 1/ }))

    // Eingeklappt spiegelt die Auswahl statt der vollen Fremd-Summe wider,
    // deckungsgleich mit dem, was Kassieren/Restbetrag verrechnen.
    expect(screen.getByText('1 ausgewählt · 2,50 €')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Kassieren/ })).toHaveTextContent(
      '2,50',
    )
  })

  it('ohne fremde Positionen gibt es keine „Von anderen"-Gruppe', () => {
    renderZahlung([position])

    expect(screen.getByText('Bratwurst Normal')).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: /Von anderen/ }),
    ).not.toBeInTheDocument()
  })
})

describe('Zahlung „Alle auswählen"', () => {
  it('wählt nur eigene Positionen voll aus und leert die Auswahl beim zweiten Tap', async () => {
    const user = userEvent.setup()
    renderZahlung([position, fremdePosition])

    // Ausgangszustand: nichts ausgewählt, Restbetrag = voller Saldo.
    expect(screen.getByText('2 unbezahlt · 7,00 €')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Kassieren/ })).toBeDisabled()

    // Erster Tap: eigene Position (2× 3,50 €) voll ausgewählt, fremde nicht.
    // Genau eine eigene Position → Singular „Meine Position auswählen".
    await user.click(
      screen.getByRole('button', {
        name: /^Meine Position auswählen · 7,00/,
      }),
    )
    expect(screen.getByText('2 von 2 ausgewählt · 7,00 €')).toBeInTheDocument()
    const bar = screen.getByRole('button', { name: /Kassieren/ })
    expect(bar).toBeEnabled()
    expect(bar).toHaveTextContent('7,00')

    // Zweiter Tap: Button heißt jetzt „Auswahl aufheben" und leert alles.
    await user.click(screen.getByRole('button', { name: 'Auswahl aufheben' }))
    expect(screen.getByText('2 unbezahlt · 7,00 €')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Kassieren/ })).toBeDisabled()
  })
})

describe('Zahlung Restbetrag', () => {
  it('zeigt den Restbetrag nach der Zahlung und rechnet mit der Auswahl live', async () => {
    const user = userEvent.setup()
    // Saldo 9,00 €; volle eigene Auswahl 7,00 € → Rest 2,00 €.
    renderZahlung([position, fremdePosition])

    expect(screen.getByText('9,00 €')).toBeInTheDocument()

    await user.click(
      screen.getByRole('button', { name: /^Meine Position auswählen/ }),
    )

    expect(screen.getByText('2,00 €')).toBeInTheDocument()
  })

  it('beschriftet den Sammel-Button bei mehreren eigenen Positionen im Plural', () => {
    // Zweite eigene Position (bestellerUserId 1) → Plural „Meine 2 Positionen".
    const zweite: Position = {
      ...position,
      positionId: '00000000-0000-0000-0000-000000000003',
      produktName: 'Brezel',
      einzelpreisCents: 200,
      menge: 1,
    }
    renderZahlung([position, zweite])

    expect(
      screen.getByRole('button', { name: /^Meine 2 Positionen auswählen/ }),
    ).toBeInTheDocument()
  })
})
