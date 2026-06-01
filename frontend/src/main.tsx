import './index.css'

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RouterProvider } from 'react-router/dom'

import { ErrorBoundary } from '@/components/common/ErrorBoundary'

import { router } from './routes.ts'

const queryClient = new QueryClient()

const documentRoot = document.getElementById('root')
if (!documentRoot) throw new Error('Failed to find the root element')

createRoot(documentRoot).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <ErrorBoundary>
        <RouterProvider router={router} />
      </ErrorBoundary>
      {import.meta.env.DEV && <ReactQueryDevtools />}
    </QueryClientProvider>
  </StrictMode>,
)
