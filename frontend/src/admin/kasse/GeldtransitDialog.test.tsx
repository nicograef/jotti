import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { GeldtransitDialog } from './GeldtransitDialog'
import { type GeldtransitRichtung } from './Kassensitzung'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

const { geldtransitBuchen } = vi.hoisted(() => ({
  geldtransitBuchen:
    vi.fn<
      (
        geldtransitId: string,
        richtung: GeldtransitRichtung,
        betragCents: number,
        kommentar: string,
      ) => Promise<void>
    >(),
}))

vi.mock('./hooks', () => ({
  kasseBackend: { geldtransitBuchen },
  GELDTRANSIT_LISTE_KEY: 'geldtransit-liste',
  KASSENBESTAND_KEY: 'kassenbestand',
}))

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
})

// Spiegelt LaufenderBetriebSection: Der Dialog bleibt über Öffnen und Schließen
// hinweg montiert (nur `open` wechselt), und der Erfolgspfad schließt ihn selbst.
// Genau deshalb überlebt der Idempotenz-Schlüssel eine geschlossene Runde.
function Harness() {
  const [richtung, setRichtung] = useState<GeldtransitRichtung | null>(
    'entnahme',
  )
  return (
    <>
      <button
        type="button"
        onClick={() => {
          setRichtung('entnahme')
        }}
      >
        Dialog öffnen
      </button>
      <GeldtransitDialog
        open={richtung !== null}
        onOpenChange={(open) => {
          if (!open) setRichtung(null)
        }}
        richtung={richtung}
      />
    </>
  )
}

function renderHarness() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  render(
    <QueryClientProvider client={queryClient}>
      <Harness />
    </QueryClientProvider>,
  )
  return { queryClient }
}

const buchen = () => screen.getByRole('button', { name: 'Geld entnehmen' })

const schluessel = (aufruf: number) => geldtransitBuchen.mock.calls[aufruf][0]
const betrag = (aufruf: number) => geldtransitBuchen.mock.calls[aufruf][2]

describe('GeldtransitDialog', () => {
  // Das Befund-Szenario: Die Buchung ist beim Server angekommen, nur ihre
  // Antwort ging verloren. Der Admin hält sie für gescheitert, korrigiert den
  // Betrag und bucht erneut. Der zweite Versuch muss denselben Schlüssel tragen
  // — nur dann erkennt der Server die Abweichung (409
  // `vorgang_daten_abweichend`), statt ein zweites Mal zu buchen.
  it('behält den Schlüssel, wenn der Betrag nach einem Fehlversuch korrigiert wird', async () => {
    const user = userEvent.setup()
    geldtransitBuchen
      .mockRejectedValueOnce(new Error('Netz weg'))
      .mockResolvedValue(undefined)
    renderHarness()

    await user.type(screen.getByLabelText('Betrag'), '25,00')
    await user.type(screen.getByLabelText('Kommentar'), 'Abschöpfung')
    await user.click(buchen())
    await waitFor(() => {
      expect(geldtransitBuchen).toHaveBeenCalledTimes(1)
    })

    // Korrektur am stehenden Betrag (25,00 → 25,50): Der Dialog bleibt nach dem
    // Fehlversuch offen, das Feld behält seinen Wert.
    await user.type(screen.getByLabelText('Betrag'), '{backspace}{backspace}50')
    await user.click(buchen())
    await waitFor(() => {
      expect(geldtransitBuchen).toHaveBeenCalledTimes(2)
    })

    expect(betrag(0)).toBe(2500)
    expect(betrag(1)).toBe(2550)
    expect(schluessel(1)).toBe(schluessel(0))
  })

  // Derselbe Fehlablauf, nur nimmt der Admin den Umweg über die Bewegungsliste:
  // Er schließt den Dialog, um nachzusehen, ob die Buchung angekommen ist, und
  // öffnet ihn wieder. Leerte der Dialog sein Formular beim Öffnen, liefe er
  // dabei durch den Leerzustand und der nächste Tastendruck vergäbe einen neuen
  // Schlüssel — die Buchung liefe ein zweites Mal durch.
  it('behält den Schlüssel über Schließen und Wiederöffnen nach einem Fehlversuch', async () => {
    const user = userEvent.setup()
    geldtransitBuchen
      .mockRejectedValueOnce(new Error('Netz weg'))
      .mockResolvedValue(undefined)
    const { queryClient } = renderHarness()
    const invalidate = vi.spyOn(queryClient, 'invalidateQueries')

    await user.type(screen.getByLabelText('Betrag'), '25,00')
    await user.type(screen.getByLabelText('Kommentar'), 'Abschöpfung')
    await user.click(buchen())
    await waitFor(() => {
      expect(geldtransitBuchen).toHaveBeenCalledTimes(1)
    })

    await user.click(screen.getByRole('button', { name: 'Abbrechen' }))
    await waitFor(() => {
      expect(screen.queryByLabelText('Betrag')).not.toBeInTheDocument()
    })
    // Der Blick in die Bewegungsliste ist nur dann eine Auskunft, wenn sie
    // nachlädt — sonst zeigt sie den Stand von vor der Buchung.
    expect(invalidate).toHaveBeenCalledWith({
      queryKey: ['geldtransit-liste'],
    })
    expect(invalidate).toHaveBeenCalledWith({ queryKey: ['kassenbestand'] })

    // Wiederöffnen: Betrag und Kommentar stehen noch, der Vorgang läuft weiter.
    await user.click(screen.getByRole('button', { name: 'Dialog öffnen' }))
    expect(screen.getByLabelText('Betrag')).toHaveValue('25,00')
    await user.click(buchen())
    await waitFor(() => {
      expect(geldtransitBuchen).toHaveBeenCalledTimes(2)
    })

    expect(betrag(1)).toBe(betrag(0))
    expect(schluessel(1)).toBe(schluessel(0))
  })

  it('wechselt den Schlüssel nach einer erfolgreichen Buchung, auch bei gleichem Betrag', async () => {
    const user = userEvent.setup()
    geldtransitBuchen.mockResolvedValue(undefined)
    renderHarness()

    await user.type(screen.getByLabelText('Betrag'), '50,00')
    await user.type(screen.getByLabelText('Kommentar'), 'Abschöpfung')
    await user.click(buchen())
    await waitFor(() => {
      expect(geldtransitBuchen).toHaveBeenCalledTimes(1)
    })

    // Der Erfolg schließt den Dialog und leert das Formular.
    await waitFor(() => {
      expect(screen.queryByLabelText('Betrag')).not.toBeInTheDocument()
    })
    await user.click(screen.getByRole('button', { name: 'Dialog öffnen' }))
    expect(screen.getByLabelText('Betrag')).toHaveValue('')

    // Zweite Entnahme über denselben Betrag: eine eigene Geldbewegung, kein
    // Wiederholversuch. Mit dem alten Schlüssel verschluckte der Server sie als
    // Duplikat.
    await user.type(screen.getByLabelText('Betrag'), '50,00')
    await user.type(screen.getByLabelText('Kommentar'), 'Abschöpfung')
    await user.click(buchen())
    await waitFor(() => {
      expect(geldtransitBuchen).toHaveBeenCalledTimes(2)
    })

    expect(betrag(1)).toBe(betrag(0))
    expect(schluessel(1)).not.toBe(schluessel(0))
  })

  // Die Semantik von useVorgangId: Nicht der Erfolg vergibt den nächsten
  // Schlüssel, sondern der Übergang von leer zu nicht leer. Ein vollständig
  // geleertes Betragsfeld beendet den Vorgang wie eine geleerte Auswahl im
  // Service-Pfad.
  it('beginnt mit dem geleerten Betragsfeld einen neuen Vorgang', async () => {
    const user = userEvent.setup()
    geldtransitBuchen
      .mockRejectedValueOnce(new Error('Netz weg'))
      .mockResolvedValue(undefined)
    renderHarness()

    await user.type(screen.getByLabelText('Betrag'), '25,00')
    await user.type(screen.getByLabelText('Kommentar'), 'Abschöpfung')
    await user.click(buchen())
    await waitFor(() => {
      expect(geldtransitBuchen).toHaveBeenCalledTimes(1)
    })

    await user.clear(screen.getByLabelText('Betrag'))
    await user.type(screen.getByLabelText('Betrag'), '25,00')
    await user.click(buchen())
    await waitFor(() => {
      expect(geldtransitBuchen).toHaveBeenCalledTimes(2)
    })

    expect(betrag(1)).toBe(betrag(0))
    expect(schluessel(1)).not.toBe(schluessel(0))
  })
})
