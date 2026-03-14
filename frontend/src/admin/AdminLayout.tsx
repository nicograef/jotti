import { Outlet } from 'react-router'

import { SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar'

import { AdminSidebar } from './AdminSidebar'

export function AdminLayout() {
  return (
    <SidebarProvider defaultOpen={true}>
      <AdminSidebar />
      <main className="min-h-screen w-full">
        <SidebarTrigger className="md:hidden" />
        <div className="px-4 py-2 md:px-8 md:py-4 xl:px-12 xl:py-6">
          <Outlet />
        </div>
      </main>
    </SidebarProvider>
  )
}
