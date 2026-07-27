import { onlineManager } from '@tanstack/react-query'
import { act, cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { OfflineBanner } from './OfflineBanner'

const bannerText = 'Keine Verbindung — Änderungen sind gerade nicht möglich'

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
  onlineManager.setOnline(true)
})

// navigator.onLine meldet in jsdom immer true; für den Offline-Start wird der
// Wert überschrieben, den das Banner beim Abonnieren abgleicht.
function stubNavigatorOnLine(onLine: boolean) {
  vi.spyOn(window.navigator, 'onLine', 'get').mockReturnValue(onLine)
}

describe('OfflineBanner', () => {
  it('zeigt nichts, solange eine Verbindung besteht', () => {
    render(<OfflineBanner />)

    expect(screen.queryByText(bannerText)).not.toBeInTheDocument()
  })

  it('erscheint ohne Interaktion, sobald das Gerät offline geht', () => {
    render(<OfflineBanner />)

    act(() => {
      window.dispatchEvent(new Event('offline'))
    })

    expect(screen.getByText(bannerText)).toBeInTheDocument()
  })

  it('verschwindet ohne Interaktion, sobald die Verbindung zurückkehrt', () => {
    render(<OfflineBanner />)

    act(() => {
      window.dispatchEvent(new Event('offline'))
    })
    act(() => {
      window.dispatchEvent(new Event('online'))
    })

    expect(screen.queryByText(bannerText)).not.toBeInTheDocument()
  })

  it('erscheint auch, wenn das Gerät bereits offline gestartet ist', () => {
    stubNavigatorOnLine(false)

    render(<OfflineBanner />)

    expect(screen.getByText(bannerText)).toBeInTheDocument()
  })
})
