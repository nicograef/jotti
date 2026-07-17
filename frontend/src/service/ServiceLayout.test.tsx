import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { ServiceLayout } from './ServiceLayout'

// UserDropdown zieht Auth-Singleton und Theme-Provider herein; für den
// Layout-Test der Modus-Umschaltung reicht ein Stub.
vi.mock('@/components/common/UserDropdown', () => ({
  UserDropdown: () => null,
}))

function renderLayout(initialPath: string) {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <Routes>
        <Route path="/service" element={<ServiceLayout />}>
          <Route path="tische" element={<div>Tischauswahl-Seite</div>} />
          <Route path="direktverkauf" element={<div>Theke-Seite</div>} />
          <Route
            path="tische/:tischId"
            element={<div>Tischdetail-Seite</div>}
          />
        </Route>
      </Routes>
    </MemoryRouter>,
  )
}

afterEach(() => {
  cleanup()
})

describe('ServiceLayout — Modus-Umschaltung', () => {
  it('zeigt auf der Tischauswahl die Segmented Control mit aktivem „Tische"', () => {
    renderLayout('/service/tische')

    const tische = screen.getByRole('link', { name: 'Tische' })
    const theke = screen.getByRole('link', { name: 'Theke' })
    expect(tische).toHaveAttribute('aria-current', 'page')
    expect(theke).not.toHaveAttribute('aria-current')
  })

  it('zeigt im Direktverkauf die Segmented Control mit aktivem „Theke"', () => {
    renderLayout('/service/direktverkauf')

    const tische = screen.getByRole('link', { name: 'Tische' })
    const theke = screen.getByRole('link', { name: 'Theke' })
    expect(theke).toHaveAttribute('aria-current', 'page')
    expect(tische).not.toHaveAttribute('aria-current')
  })

  it('navigiert über die Control von Tischen zur Theke und zurück', async () => {
    const user = userEvent.setup()
    renderLayout('/service/tische')

    expect(screen.getByText('Tischauswahl-Seite')).toBeInTheDocument()

    await user.click(screen.getByRole('link', { name: 'Theke' }))
    expect(screen.getByText('Theke-Seite')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Theke' })).toHaveAttribute(
      'aria-current',
      'page',
    )

    await user.click(screen.getByRole('link', { name: 'Tische' }))
    expect(screen.getByText('Tischauswahl-Seite')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Tische' })).toHaveAttribute(
      'aria-current',
      'page',
    )
  })

  it('zeigt auf der Tisch-Detailseite den Backlink statt der Control', () => {
    renderLayout('/service/tische/7')

    expect(screen.getByText('Tischdetail-Seite')).toBeInTheDocument()
    // Backlink „Meine Tische" ist vorhanden, aber kein Theke-Segment.
    expect(
      screen.getByRole('link', { name: /Meine Tische/ }),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('link', { name: 'Theke' }),
    ).not.toBeInTheDocument()
  })

  it('macht die Tap-Ziele der Segmente mindestens 44 px hoch', () => {
    renderLayout('/service/tische')

    expect(screen.getByRole('link', { name: 'Tische' })).toHaveClass('h-11')
    expect(screen.getByRole('link', { name: 'Theke' })).toHaveClass('h-11')
  })
})
