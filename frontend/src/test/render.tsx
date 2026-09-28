import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  render,
  renderHook,
  type RenderHookResult,
  type RenderResult,
} from '@testing-library/react'
import type { ReactElement, ReactNode } from 'react'
import { vi } from 'vitest'

import { BackendSingleton } from '@/lib/Backend'
import { createQueryClient } from '@/lib/queryClient'

import type { FakeBackend } from './FakeBackend'

// Domain backends are module singletons built on BackendSingleton, so the fake
// takes over its two methods instead of being passed in.
function installBackend(backend: FakeBackend): void {
  vi.spyOn(BackendSingleton, 'post').mockImplementation(
    (endpoint, body, responseSchema) =>
      backend.post(endpoint, body, responseSchema),
  )
  vi.spyOn(BackendSingleton, 'download').mockImplementation((endpoint, body) =>
    backend.download(endpoint, body),
  )
}

function queryWrapper(queryClient: QueryClient) {
  return function QueryWrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    )
  }
}

export function renderWithBackend(
  ui: ReactElement,
  backend: FakeBackend,
): RenderResult & { queryClient: QueryClient } {
  installBackend(backend)
  const queryClient = createQueryClient()
  const result = render(ui, { wrapper: queryWrapper(queryClient) })
  return { ...result, queryClient }
}

export function renderHookWithBackend<TResult>(
  hook: () => TResult,
  backend: FakeBackend,
): RenderHookResult<TResult, unknown> & { queryClient: QueryClient } {
  installBackend(backend)
  const queryClient = createQueryClient()
  const result = renderHook(hook, { wrapper: queryWrapper(queryClient) })
  return { ...result, queryClient }
}

/** useIsMobile reads window.innerWidth; jsdom starts at 1024 px (desktop). */
export function setViewportWidth(px: number): void {
  window.innerWidth = px
}
