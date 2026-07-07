import './index.css'

import { QueryClientProvider } from '@tanstack/react-query'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RouterProvider } from 'react-router/dom'

import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { createQueryClient } from '@/lib/queryClient'

import { router } from './routes.ts'

const queryClient = createQueryClient()

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
