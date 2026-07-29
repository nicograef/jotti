import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { Dialog, DialogContent, DialogTitle } from '@/components/ui/dialog'
import type { VersionsZustand } from '@/hooks/use-versions-guard'
import { seiteNeuLaden } from '@/lib/reload'

import { VersionsHinweis } from './VersionsHinweis'

const guardState = vi.hoisted<{ versionsZustand: VersionsZustand }>(() => ({
  versionsZustand: 'aus',
}))

vi.mock('@/hooks/use-versions-guard', () => ({
  useVersionsGuard: () => guardState.versionsZustand,
}))

vi.mock('@/lib/reload', () => ({ seiteNeuLaden: vi.fn() }))

// Vitest verarbeitet kein CSS, Tailwind-Klassen bleiben im Test also
// wirkungslos. Ohne diese eine Deklaration — genau die, die Tailwind für
// `pointer-events-auto` erzeugt — könnte kein Test sehen, ob der Hinweis neben
// einem offenen Modal noch bedienbar ist. Die Stapelreihenfolge selbst ist in
// jsdom nicht prüfbar (kein Layout, kein Malen).
const tailwindErsatz = document.createElement('style')
tailwindErsatz.textContent = '.pointer-events-auto { pointer-events: auto }'
document.head.append(tailwindErsatz)

beforeEach(() => {
  guardState.versionsZustand = 'aus'
  vi.mocked(seiteNeuLaden).mockClear()
})

afterEach(() => {
  cleanup()
})

describe('VersionsHinweis', () => {
  it('bleibt aus, solange der Guard keinen Anlass sieht', () => {
    render(<VersionsHinweis />)

    expect(screen.queryByRole('alert')).toBeNull()
  })

  // Der Regelfall des Handshakes ist unsichtbar: Bei leerem Vorgangs-Register
  // lädt die Seite sofort neu, und bis zum Entladen gibt es nichts zu erklären.
  it('bleibt während des Reloads aus', () => {
    guardState.versionsZustand = 'laedt'

    render(<VersionsHinweis />)

    expect(screen.queryByRole('alert')).toBeNull()
  })

  it('erklärt beim Warten, was den Reload noch aufhält', () => {
    guardState.versionsZustand = 'wartet'

    render(<VersionsHinweis />)

    expect(screen.getByRole('alert')).toHaveTextContent(
      'Der Server läuft mit einer anderen Version als diese Seite. Bitte den laufenden Vorgang abschließen oder verwerfen — danach lädt sich die Seite von selbst neu.',
    )
    // Im Regelfall lädt die Seite von selbst; eine Schaltfläche wäre hier nur
    // ignorierbar und damit wertlos.
    expect(screen.queryByRole('button')).toBeNull()
  })

  // Gebremst lädt dieser Client nicht mehr von selbst. Die Zusage „es lädt
  // gleich von selbst neu" wäre dann schlicht gelogen.
  it('verspricht gebremst kein automatisches Neuladen mehr', () => {
    guardState.versionsZustand = 'gebremst'

    render(<VersionsHinweis />)

    expect(screen.getByRole('alert')).toHaveTextContent(
      'Der Server läuft mit einer anderen Version als diese Seite. Das automatische Neuladen hat nicht geklappt — bitte von Hand neu laden.',
    )
  })

  it('lädt gebremst auf Knopfdruck neu', async () => {
    guardState.versionsZustand = 'gebremst'

    render(<VersionsHinweis />)
    await userEvent.click(
      screen.getByRole('button', { name: 'Jetzt neu laden' }),
    )

    expect(seiteNeuLaden).toHaveBeenCalledTimes(1)
  })

  // Der Hinweis wartet auf das Abschließen oder Verwerfen des laufenden
  // Vorgangs. Wäre er ein modaler Dialog, sperrte er genau die Bedienung aus,
  // auf die er wartet.
  it('lässt den laufenden Vorgang weiter bedienen und ist nicht wegklickbar', async () => {
    guardState.versionsZustand = 'wartet'
    const kassieren = vi.fn()

    render(
      <div>
        <VersionsHinweis />
        <button onClick={kassieren}>Kassieren</button>
      </div>,
    )

    await userEvent.keyboard('{Escape}')
    await userEvent.click(screen.getByRole('button', { name: 'Kassieren' }))

    expect(kassieren).toHaveBeenCalledTimes(1)
    expect(screen.getByRole('alert')).toBeInTheDocument()
  })

  // Mehrere der meldenden Vorgänge leben nur, solange ein Dialog oder Drawer
  // offen ist — der Hinweis erscheint dort also zwangsläufig neben einem
  // offenen Modal. Radix legt dann die ganze Seite außerhalb des Portals stumm
  // (`body { pointer-events: none }`); ohne eigenen Stapelplatz wäre der
  // einzige Ausweg aus dem gebremsten Zustand nicht zu treffen. Die Rollen-
  // Abfrage braucht `hidden`, weil Radix denselben Teilbaum zusätzlich
  // `aria-hidden` setzt.
  it('bleibt gebremst auch neben einem offenen Modal bedienbar', async () => {
    guardState.versionsZustand = 'gebremst'

    render(
      <>
        <VersionsHinweis />
        <Dialog open>
          <DialogContent>
            <DialogTitle>Zählhilfe</DialogTitle>
          </DialogContent>
        </Dialog>
      </>,
    )
    await userEvent.click(
      screen.getByRole('button', { name: 'Jetzt neu laden', hidden: true }),
    )

    expect(seiteNeuLaden).toHaveBeenCalledTimes(1)
  })
})
