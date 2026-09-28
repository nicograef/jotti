import { cleanup, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ReactNode } from 'react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { Dialog, DialogContent, DialogTitle } from '@/components/ui/dialog'
import { RELOAD_VERMERK_SCHLUESSEL } from '@/hooks/use-versions-guard'
import { Seite } from '@/lib/reload'
import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'
import { FakeBackend } from '@/test/FakeBackend'
import { renderWithBackend } from '@/test/render'

import { VersionsHinweis } from './VersionsHinweis'

// The client runs as a real release; /health reports `serverVersion`.
const CLIENT = 'v1.2.3'
let serverVersion = CLIENT

// Waits until /health has answered, so the guard has decided on its state.
async function renderHinweis(
  umgebung: (hinweis: ReactNode) => ReactNode = (hinweis) => hinweis,
) {
  const backend = new FakeBackend().respond('health', () => ({
    version: serverVersion,
  }))
  const { queryClient } = renderWithBackend(
    <>{umgebung(<VersionsHinweis clientVersion={CLIENT} />)}</>,
    backend,
  )
  await waitFor(() => {
    expect(backend.bodies('health')).toHaveLength(1)
    expect(queryClient.isFetching()).toBe(0)
  })
}

// A different server release with an open Vorgang holds the reload back.
function wartet() {
  serverVersion = 'v1.2.4'
  VorgangsRegisterSingleton.anmelden()
}

// A reload that already aimed at the server release brought no new bundle.
function gebremst() {
  serverVersion = 'v1.2.4'
  sessionStorage.setItem(RELOAD_VERMERK_SCHLUESSEL, 'v1.2.4')
}

let neuLaden: ReturnType<typeof vi.spyOn>

// Vitest verarbeitet kein CSS. Ohne genau diese eine Deklaration — die, die
// Tailwind für `pointer-events-auto` erzeugt — könnte kein Test sehen, ob der
// Hinweis neben einem offenen Modal bedienbar bleibt.
const tailwindErsatz = document.createElement('style')
tailwindErsatz.textContent = '.pointer-events-auto { pointer-events: auto }'
document.head.append(tailwindErsatz)

beforeEach(() => {
  serverVersion = CLIENT
  sessionStorage.clear()
  VorgangsRegisterSingleton.zuruecksetzen()
  neuLaden = vi.spyOn(Seite, 'neuLaden').mockImplementation(() => undefined)
})

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

describe('VersionsHinweis', () => {
  it('bleibt aus, solange der Guard keinen Anlass sieht', async () => {
    await renderHinweis()

    expect(screen.queryByRole('alert')).toBeNull()
  })

  it('bleibt während des Reloads aus', async () => {
    serverVersion = 'v1.2.4'

    await renderHinweis()

    expect(neuLaden).toHaveBeenCalledTimes(1)

    expect(screen.queryByRole('alert')).toBeNull()
  })

  it('erklärt beim Warten, was den Reload noch aufhält', async () => {
    wartet()

    await renderHinweis()

    expect(screen.getByRole('alert')).toHaveTextContent(
      'Der Server läuft mit einer anderen Version als diese Seite. Bitte den laufenden Vorgang abschließen oder verwerfen — danach lädt sich die Seite von selbst neu.',
    )
    // Eine Schaltfläche wäre hier wertlos: Die Seite lädt von selbst neu.
    expect(screen.queryByRole('button')).toBeNull()
  })

  // Gebremst lädt dieser Client nicht mehr von selbst; die Zusage wäre gelogen.
  it('verspricht gebremst kein automatisches Neuladen mehr', async () => {
    gebremst()

    await renderHinweis()

    expect(screen.getByRole('alert')).toHaveTextContent(
      'Der Server läuft mit einer anderen Version als diese Seite. Das automatische Neuladen hat nicht geklappt — bitte von Hand neu laden.',
    )
  })

  it('lädt gebremst auf Knopfdruck neu', async () => {
    gebremst()

    await renderHinweis()
    await userEvent.click(
      screen.getByRole('button', { name: 'Jetzt neu laden' }),
    )

    expect(neuLaden).toHaveBeenCalledTimes(1)
  })

  // Wäre der Hinweis ein modaler Dialog, sperrte er genau die Bedienung aus,
  // auf die er wartet.
  it('lässt den laufenden Vorgang weiter bedienen und ist nicht wegklickbar', async () => {
    wartet()
    const kassieren = vi.fn()

    await renderHinweis((hinweis) => (
      <div>
        {hinweis}
        <button onClick={kassieren}>Kassieren</button>
      </div>
    ))

    await userEvent.keyboard('{Escape}')
    await userEvent.click(screen.getByRole('button', { name: 'Kassieren' }))

    expect(kassieren).toHaveBeenCalledTimes(1)
    expect(screen.getByRole('alert')).toBeInTheDocument()
  })

  // Der Hinweis erscheint zwangsläufig neben einem offenen Modal; Radix legt
  // dann die Seite außerhalb des Portals still (`body { pointer-events: none }`).
  // Die Rollen-Abfrage braucht `hidden`, weil Radix denselben Teilbaum
  // zusätzlich `aria-hidden` setzt.
  it('bleibt gebremst auch neben einem offenen Modal bedienbar', async () => {
    gebremst()

    await renderHinweis((hinweis) => (
      <>
        {hinweis}
        <Dialog open>
          <DialogContent>
            <DialogTitle>Zählhilfe</DialogTitle>
          </DialogContent>
        </Dialog>
      </>
    ))
    await userEvent.click(
      screen.getByRole('button', { name: 'Jetzt neu laden', hidden: true }),
    )

    expect(neuLaden).toHaveBeenCalledTimes(1)
  })
})
