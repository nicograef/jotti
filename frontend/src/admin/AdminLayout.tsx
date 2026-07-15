import { Menu } from 'lucide-react'
import { Outlet } from 'react-router'

import { UserDropdown } from '@/components/common/UserDropdown'
import { Wortmarke } from '@/components/common/Wortmarke'
import { Button } from '@/components/ui/button'
import { SidebarProvider, useSidebar } from '@/components/ui/sidebar'

import { AdminSidebar } from './AdminSidebar'

function AdminMobileHeader() {
  const { toggleSidebar } = useSidebar()

  return (
    <header className="sticky top-0 h-14 border-b bg-background z-40 flex items-center justify-between px-4 md:hidden print:hidden">
      <div className="flex items-center gap-2">
        <Button
          variant="ghost"
          size="icon"
          className="size-11"
          onClick={toggleSidebar}
          aria-label="Menü öffnen"
        >
          <Menu className="h-5 w-5" />
        </Button>
        <Wortmarke className="text-sm" />
      </div>
      <UserDropdown />
    </header>
  )
}

export function AdminLayout() {
  return (
    <SidebarProvider defaultOpen={true}>
      <AdminSidebar />
      <main className="min-h-screen w-full min-w-0">
        <AdminMobileHeader />
        <div className="px-4 py-2 md:px-8 md:py-4 xl:px-12 xl:py-6">
          <Outlet />
        </div>
      </main>
    </SidebarProvider>
  )
}
