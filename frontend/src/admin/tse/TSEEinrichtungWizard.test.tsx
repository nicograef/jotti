import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { BackendError } from '@/lib/Backend'
import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'

import type {
  TSEEinrichtenErgebnis,
  TSESetupBefund,
  TSEVerbindungStatus,
} from './TSEBackend'
import { TSEEinrichtungWizard } from './TSEEinrichtungWizard'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

const { checkTSESetup, richteTSEEin, uebernimmTSE, testTSEVerbindung } =
  vi.hoisted(() => ({
    checkTSESetup: vi.fn<() => Promise<TSESetupBefund>>(),
    richteTSEEin: vi.fn<() => Promise<TSEEinrichtenErgebnis>>(),
    uebernimmTSE: vi.fn<() => Promise<TSEEinrichtenErgebnis>>(),
    testTSEVerbindung: vi.fn<() => Promise<TSEVerbindungStatus>>(),
  }))

vi.mock('./hooks', () => ({
  checkTSESetup,
  useTSEEinrichtung: () => ({ richteTSEEin, uebernimmTSE }),
  useTSEKonfiguration: () => ({ testTSEVerbindung }),
}))

// Eine bereits eingerichtete TSE, die jotti übernehmen kann — der Weg, auf dem
// Admin-PIN und Admin-PUK abgefragt werden.
const uebernehmbarerBefund: TSESetupBefund = {
  umgebung: 'TEST',
  vorhandeneTss: [
    {
      id: 'tss-1',
      state: 'UNINITIALIZED',
      passenderClient: null,
    },
  ],
}

// Eine TSE, in der diese Kasse schon angemeldet ist: INITIALIZED plus
// REGISTERED Client — einsatzbereit ohne privilegierte fiskaly-Operation und
// damit ohne Admin-PIN.
const einsatzbereiterBefund: TSESetupBefund = {
  umgebung: 'TEST',
  vorhandeneTss: [
    {
      id: 'tss-1',
      state: 'INITIALIZED',
      passenderClient: {
        id: 'client-1',
        serialNumber: 'jotti-1',
        state: 'REGISTERED',
      },
    },
  ],
}

const neueTseErgebnis: TSEEinrichtenErgebnis = {
  tssId: 'tss-1',
  clientId: 'client-1',
  puk: 'PUK-123456',
  adminPin: '12345',
  umgebung: 'TEST',
}

beforeEach(() => {
  VorgangsRegisterSingleton.zuruecksetzen()
})

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
})

// Führt den Wizard bis zum Befund-Schritt: Zugangsdaten eintippen und prüfen.
async function bisZumBefund(
  user: ReturnType<typeof userEvent.setup>,
  befund: TSESetupBefund,
) {
  checkTSESetup.mockResolvedValue(befund)
  render(<TSEEinrichtungWizard />)

  await user.type(screen.getByLabelText('API-Key'), 'key')
  await user.type(screen.getByLabelText('API-Secret'), 'secret')
  await user.click(screen.getByRole('button', { name: 'fiskaly-Konto prüfen' }))
  // Der Befund-Schritt ist erreicht, sobald sein Rückweg dasteht.
  await screen.findByRole('button', { name: 'Andere Zugangsdaten' })
}

describe('TSEEinrichtungWizard im Vorgangs-Register', () => {
  it('meldet getippte Zugangsdaten und gibt sie beim Leeren frei', async () => {
    const user = userEvent.setup()
    render(<TSEEinrichtungWizard />)

    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)

    await user.type(screen.getByLabelText('API-Key'), 'key')
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)

    // Das zweite Feld ist derselbe Vorgang, kein zweiter.
    await user.type(screen.getByLabelText('API-Secret'), 'secret')
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)

    await user.clear(screen.getByLabelText('API-Key'))
    await user.clear(screen.getByLabelText('API-Secret'))
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })

  it('meldet die abgetippte Admin-PIN zusätzlich', async () => {
    const user = userEvent.setup()
    await bisZumBefund(user, uebernehmbarerBefund)

    const vorher = VorgangsRegisterSingleton.anzahlOffen()

    await user.type(screen.getByLabelText('Admin-PIN'), '12345')
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(vorher + 1)

    await user.clear(screen.getByLabelText('Admin-PIN'))
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(vorher)
  })

  it('meldet den abgetippten Admin-PUK, das bloße Aufklappen nicht', async () => {
    const user = userEvent.setup()
    await bisZumBefund(user, uebernehmbarerBefund)

    const vorher = VorgangsRegisterSingleton.anzahlOffen()

    await user.click(
      screen.getByRole('button', { name: 'Ich habe den Admin-PUK' }),
    )
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(vorher)

    await user.type(screen.getByLabelText('Admin-PUK'), 'PUK-123456')
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(vorher + 1)
  })

  it('meldet die begonnene LIVE-Tippbestätigung', async () => {
    const user = userEvent.setup()
    await bisZumBefund(user, { umgebung: 'LIVE', vorhandeneTss: [] })

    const vorher = VorgangsRegisterSingleton.anzahlOffen()

    await user.type(screen.getByLabelText(/Zur Bestätigung/), 'LIVE')
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(vorher + 1)
  })

  it('hält den Ergebnis-Schritt gemeldet, bis er verlassen wird — der Verwahr-Haken gibt nicht frei', async () => {
    const user = userEvent.setup()
    richteTSEEin.mockResolvedValue(neueTseErgebnis)
    testTSEVerbindung.mockResolvedValue({
      umgebung: 'TEST',
      tssState: 'INITIALIZED',
      clientState: 'REGISTERED',
      clientSerialNumber: 'jotti-1',
      seriennummerKorrekt: true,
    })
    await bisZumBefund(user, { umgebung: 'TEST', vorhandeneTss: [] })

    const vorZurAnlage = VorgangsRegisterSingleton.anzahlOffen()

    await user.click(screen.getByRole('button', { name: 'TSE einrichten' }))
    // PUK und Admin-PIN stehen jetzt genau einmal auf dem Schirm.
    await screen.findByText('PUK-123456')
    const mitGeheimnissen = VorgangsRegisterSingleton.anzahlOffen()
    expect(mitGeheimnissen).toBe(vorZurAnlage + 1)

    await user.click(screen.getByRole('checkbox'))
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(mitGeheimnissen)

    await user.click(
      screen.getByRole('button', { name: 'Verbindung testen & abschließen' }),
    )
    await screen.findByText('Verbindung bestätigt')
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(mitGeheimnissen)

    // Erst „Fertig" verlässt den Schritt — und gibt mit den Geheimnissen auch
    // die Zugangsdaten frei, die nach der Einrichtung niemand mehr braucht.
    await user.click(screen.getByRole('button', { name: 'Fertig' }))
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })
})

describe('TSEEinrichtungWizard — Sperren der Einrichtung', () => {
  it('sperrt die LIVE-Anlage, bis „LIVE" abgetippt ist', async () => {
    const user = userEvent.setup()
    await bisZumBefund(user, { umgebung: 'LIVE', vorhandeneTss: [] })

    const einrichten = screen.getByRole('button', { name: 'TSE einrichten' })
    expect(einrichten).toBeDisabled()

    // Ein anderes Wort gibt nicht frei.
    await user.type(screen.getByLabelText(/Zur Bestätigung/), 'TEST')
    expect(einrichten).toBeDisabled()

    await user.clear(screen.getByLabelText(/Zur Bestätigung/))
    await user.type(screen.getByLabelText(/Zur Bestätigung/), 'LIVE')
    expect(einrichten).toBeEnabled()
  })

  it('meldet eine abgelehnte Admin-PIN als bleibenden Hinweis', async () => {
    const user = userEvent.setup()
    uebernimmTSE.mockRejectedValue(
      new BackendError(409, 'tse_setup_pin_unbekannt'),
    )
    await bisZumBefund(user, uebernehmbarerBefund)

    await user.type(screen.getByLabelText('Admin-PIN'), '99999')
    await user.click(screen.getByRole('button', { name: 'TSE übernehmen' }))

    expect(
      await screen.findByText('Admin-PIN nicht akzeptiert'),
    ).toBeInTheDocument()
  })

  it('verlangt für eine einsatzbereite TSE keine Admin-PIN', async () => {
    const user = userEvent.setup()
    await bisZumBefund(user, einsatzbereiterBefund)

    expect(screen.queryByLabelText('Admin-PIN')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'TSE übernehmen' })).toBeEnabled()
  })

  it('bietet bei ausschließlich deaktivierten TSE die Neuanlage an', async () => {
    const user = userEvent.setup()
    await bisZumBefund(user, {
      umgebung: 'TEST',
      vorhandeneTss: [
        { id: 'tss-alt', state: 'DISABLED', passenderClient: null },
      ],
    })

    expect(
      screen.getByRole('button', { name: 'TSE einrichten' }),
    ).toBeInTheDocument()
    expect(
      screen.queryByText('TSE in nicht übernehmbarem Zustand'),
    ).not.toBeInTheDocument()
  })
})
