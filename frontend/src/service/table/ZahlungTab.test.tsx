import { cleanup, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ComponentProps } from 'react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { useMengen } from '@/hooks/use-mengen'
import { signIn, signOut } from '@/test/auth'
import { FakeBackend } from '@/test/FakeBackend'
import { setViewportWidth } from '@/test/render'

import { ServiceDock } from '../ServiceDock'
import type { Position } from './Bestellung'
import type { Tisch } from './Tisch'
import { TischBackend } from './TischBackend'
import { ZahlungTab } from './ZahlungTab'

// Die Kassieren-Auswahl liegt in TablePage; dieser Harness stellt die gehobene
// Steuerung inklusive der Deckelung auf die unbezahlte Menge bereit.
function ZahlungHarness(
  props: Omit<ComponentProps<typeof ZahlungTab>, 'mengenSteuerung'>,
) {
  const unbezahlteMengen: Record<string, number> = {}
  props.positionen.forEach((position) => {
    unbezahlteMengen[position.positionId] = position.menge
  })
  const mengenSteuerung = useMengen<string>(
    (positionId) => unbezahlteMengen[positionId] || 0,
  )
  return <ZahlungTab {...props} mengenSteuerung={mengenSteuerung} />
}

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

// Standardmäßig Handy-Layout (Dock-Button, Restbetrag im Dock-Slot); ein Test
// unten schaltet auf Desktop, um die Verdrahtung der festen Spalte zu prüfen.
// Deren container-neutrales Verhalten deckt ZahlungAbschluss.test.tsx ab.
beforeEach(() => {
  setViewportWidth(375)
  signIn({ userId: 1 })
})

afterEach(() => {
  cleanup()
  setViewportWidth(1024)
  signOut()
})

function tischBackend(): TischBackend {
  return new TischBackend(
    new FakeBackend().respond('service/zahlung-kassieren', {}),
  )
}

const tisch: Tisch = { id: 1, name: 'Stammtisch', saldoCents: 900 }

const position: Position = {
  positionId: '00000000-0000-4000-8000-000000000001',
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
  positionId: '00000000-0000-4000-8000-000000000002',
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
        backend={tischBackend()}
        tisch={tisch}
        positionen={positionen}
        onErfolg={vi.fn()}
      />
    </ServiceDock>,
  )
}

describe('ZahlungTab feste Spalte (ab lg)', () => {
  it('rendert die Abschluss-Spalte mit Restbetrag statt Dock und Drawer', async () => {
    setViewportWidth(1024)
    const user = userEvent.setup()
    // Kein ServiceDock: die feste Spalte trägt Aktionsbutton und Restbetrag.
    render(
      <ZahlungHarness
        backend={tischBackend()}
        tisch={tisch}
        positionen={[position]}
        onErfolg={vi.fn()}
      />,
    )

    expect(screen.getByText('Bratwurst Normal')).toBeInTheDocument()
    expect(screen.getByText('Nach dieser Zahlung noch offen')).toBeVisible()
    const button = screen.getByRole('button', { name: 'Kassieren' })
    expect(button).toBeDisabled()

    await user.click(screen.getByRole('button', { name: 'Produkt hinzufügen' }))
    expect(screen.getByRole('button', { name: 'Kassieren' })).toBeEnabled()
  })
})

describe('ZahlungTab Aktionsleiste', () => {
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

describe('ZahlungTab Positionsgruppen', () => {
  it('zeigt eigene Positionen direkt und fremde erst nach dem Aufklappen von „Von anderen"', async () => {
    const user = userEvent.setup()
    renderZahlung([position, fremdePosition])

    expect(screen.getByText('Bratwurst Normal')).toBeInTheDocument()
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

describe('ZahlungTab „Meine auswählen"', () => {
  it('wählt nur eigene Positionen voll aus und leert die Auswahl beim zweiten Tap', async () => {
    const user = userEvent.setup()
    renderZahlung([position, fremdePosition])

    expect(screen.getByText('2 unbezahlt · 7,00 €')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Kassieren/ })).toBeDisabled()

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

    await user.click(screen.getByRole('button', { name: 'Auswahl aufheben' }))
    expect(screen.getByText('2 unbezahlt · 7,00 €')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Kassieren/ })).toBeDisabled()
  })
})

describe('ZahlungTab Restbetrag', () => {
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
      positionId: '00000000-0000-4000-8000-000000000003',
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
