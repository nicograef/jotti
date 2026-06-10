import './index.css'

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RouterProvider } from 'react-router/dom'

import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { ResponseBodyError } from '@/lib/Backend'

import { router } from './routes.ts'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      refetchOnWindowFocus: false,
      retry: (failureCount, error) => {
        if (error instanceof ResponseBodyError) return false
        return failureCount < 2
      },
    },
  },
})

const documentRoot = document.getElementById('root')
if (!documentRoot) throw new Error('Failed to find the root element')

createRoot(documentRoot).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <ErrorBoundary>
        <RouterProvider router={router} />
      </ErrorBoundary>
    </QueryClientProvider>
  </StrictMode>,
)
