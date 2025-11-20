import { Outlet } from 'react-router'

import { SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar'

import { ServiceSidebar } from './ServiceSidebar'

export function ServiceLayout() {
  return (
    <SidebarProvider>
      <ServiceSidebar />
      <main className="min-h-screen max-h-screen w-full">
        <SidebarTrigger />
        <div className="px-4 py-2 md:px-8 md:py-4 xl:px-12 xl:py-6">
          <Outlet />
        </div>
      </main>
    </SidebarProvider>
  )
}
