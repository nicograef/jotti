import { Outlet } from 'react-router'

import { OfflineBanner } from '@/components/common/OfflineBanner'
import { ThemeProvider } from '@/components/theme-provider'
import { Toaster } from '@/components/ui/sonner'
import { TooltipProvider } from '@/components/ui/tooltip'

export default function App() {
  return (
    <ThemeProvider storageKey="vite-ui-theme">
      <TooltipProvider>
        <Toaster position="top-right" />
        <OfflineBanner />
        <Outlet />
      </TooltipProvider>
    </ThemeProvider>
  )
}
