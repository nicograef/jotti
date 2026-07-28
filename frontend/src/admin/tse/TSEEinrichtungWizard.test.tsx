import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { toast } from 'sonner'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { NetzwerkFehler } from '@/lib/Backend'

import { TSEEinrichtungWizard } from './TSEEinrichtungWizard'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
}))

const mocks = vi.hoisted(() => ({
  checkTSESetup: vi.fn(),
  richteTSEEin: vi.fn(),
  uebernimmTSE: vi.fn(),
}))

vi.mock('./hooks', () => ({
  checkTSESetup: mocks.checkTSESetup,
  useTSEEinrichtung: () => ({
    richteTSEEin: mocks.richteTSEEin,
    uebernimmTSE: mocks.uebernimmTSE,
  }),
  useTSEKonfiguration: () => ({ testTSEVerbindung: vi.fn() }),
}))

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
})

// Bis zum Bestätigungsschritt eines leeren fiskaly-Kontos durchklicken — von
// dort startet die Neuanlage.
async function bisZumEinrichtenKlicken() {
  const user = userEvent.setup()
  mocks.checkTSESetup.mockResolvedValue({ umgebung: 'TEST', vorhandeneTss: [] })

  render(<TSEEinrichtungWizard />)

  await user.type(screen.getByLabelText('API-Key'), 'api-key')
  await user.type(screen.getByLabelText('API-Secret'), 'api-secret')
  await user.click(screen.getByRole('button', { name: 'fiskaly-Konto prüfen' }))
  await user.click(
    await screen.findByRole('button', { name: 'TSE einrichten' }),
  )
}

describe('TSEEinrichtungWizard', () => {
  // Nach dem Client-Zeitlimit läuft der fiskaly-Lebenszyklus serverseitig
  // weiter. Der allgemeine Rat „erneut versuchen" wäre hier der teuerste
  // mögliche: Ein zweiter Start legt eine zweite, kostenpflichtige TSE an.
  it('rät nach einer Zeitüberschreitung zum Prüfen statt zum erneuten Starten', async () => {
    mocks.richteTSEEin.mockRejectedValue(
      new NetzwerkFehler('zeitueberschreitung'),
    )

    await bisZumEinrichtenKlicken()

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalled()
    })
    const meldung = vi.mocked(toast.error).mock.calls[0][0]
    expect(meldung).toBe(
      'Die Einrichtung läuft im Hintergrund weiter. Bitte nicht erneut starten – das legt eine zweite, kostenpflichtige TSE an. Bitte kurz warten und dann unten die TSE-Konfiguration prüfen.',
    )
  })

  // Der Verbindungsabbruch bleibt bei der allgemeinen Meldung: Ob die Anfrage
  // den Server überhaupt erreicht hat, ist dabei offen.
  it('behält für einen Verbindungsabbruch die allgemeine Netzmeldung', async () => {
    mocks.richteTSEEin.mockRejectedValue(
      new NetzwerkFehler('verbindungsabbruch'),
    )

    await bisZumEinrichtenKlicken()

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalled()
    })
    const meldung = vi.mocked(toast.error).mock.calls[0][0]
    expect(meldung).toBe(
      'Keine Verbindung zum Server. Bitte WLAN prüfen und erneut versuchen.',
    )
  })
})
