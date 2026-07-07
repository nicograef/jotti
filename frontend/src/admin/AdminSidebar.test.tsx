import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { SidebarProvider } from '@/components/ui/sidebar'

import { AdminSidebar } from './AdminSidebar'

vi.mock('@/components/theme-provider', () => ({
  useTheme: () => ({ isDark: false, setTheme: vi.fn() }),
}))

// jsdom kennt window.matchMedia nicht (von SidebarProvider via useIsMobile benötigt).
vi.mock('@/hooks/use-mobile', () => ({
  useIsMobile: () => false,
}))

const versionState = vi.hoisted<{ version: string | undefined }>(() => ({
  version: undefined,
}))

vi.mock('./hooks', () => ({
  useVersion: () => versionState.version,
}))

function renderSidebar() {
  render(
    <MemoryRouter>
      <SidebarProvider>
        <AdminSidebar />
      </SidebarProvider>
    </MemoryRouter>,
  )
}

afterEach(() => {
  cleanup()
})

describe('AdminSidebar', () => {
  it('zeigt die Version im Footer, sobald sie geladen ist', () => {
    versionState.version = 'v1.0.0'
    renderSidebar()

    expect(screen.getByText('jotti v1.0.0')).toBeInTheDocument()
  })

  it('zeigt keine Versionszeile, solange die Version nicht geladen ist', () => {
    versionState.version = undefined
    renderSidebar()

    expect(screen.queryByText(/jotti v/)).not.toBeInTheDocument()
  })
})
