import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'

import { useDsfinvkExport } from './hooks'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

vi.mock('@/lib/download', () => ({
  triggerBrowserDownload: vi.fn(),
}))

const { exportDsfinvk } = vi.hoisted(() => ({
  exportDsfinvk: vi.fn<() => Promise<{ blob: Blob; filename: string }>>(),
}))

// hooks.ts baut sein Backend beim Import (`new ReportingBackend(...)`) — der
// Ersatz muss deshalb konstruierbar sein.
vi.mock('./ReportingBackend', () => ({
  ReportingBackend: class {
    exportDsfinvk = exportDsfinvk
  },
}))

function Wrapper({ children }: { children: ReactNode }) {
  const queryClient = new QueryClient({
    defaultOptions: { mutations: { retry: false } },
  })
  return (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

beforeEach(() => {
  VorgangsRegisterSingleton.zuruecksetzen()
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('useDsfinvkExport im Vorgangs-Register', () => {
  it('meldet den laufenden Export und gibt ihn nach dem Download frei', async () => {
    let liefern!: (archiv: { blob: Blob; filename: string }) => void
    exportDsfinvk.mockReturnValue(
      new Promise((resolve) => {
        liefern = resolve
      }),
    )
    const { result } = renderHook(() => useDsfinvkExport(), {
      wrapper: Wrapper,
    })

    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)

    act(() => {
      result.current.exportieren(1)
    })
    await waitFor(() => {
      expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)
    })

    await act(async () => {
      liefern({ blob: new Blob(), filename: 'dsfinvk.zip' })
      await Promise.resolve()
    })
    await waitFor(() => {
      expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
    })
  })

  it('gibt den Export auch nach einem Fehlschlag frei', async () => {
    let scheitern!: (fehler: Error) => void
    exportDsfinvk.mockReturnValue(
      new Promise((_, reject) => {
        scheitern = reject
      }),
    )
    const { result } = renderHook(() => useDsfinvkExport(), {
      wrapper: Wrapper,
    })

    act(() => {
      result.current.exportieren(1)
    })
    await waitFor(() => {
      expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)
    })

    await act(async () => {
      scheitern(new Error('Netzabbruch'))
      await Promise.resolve()
    })
    await waitFor(() => {
      expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
    })
  })
})
