import { ChevronLeft } from 'lucide-react'
import { Link, Outlet, useMatch } from 'react-router'

import { UserDropdown } from '@/components/common/UserDropdown'

export function ServiceLayout() {
  const onTableDetail = useMatch('/service/tables/:tableId')

  return (
    <div className="min-h-screen flex flex-col">
      <header className="sticky top-0 h-14 border-b bg-background z-40 flex items-center justify-between px-4">
        <div className="flex items-center gap-2">
          {onTableDetail ? (
            <Link
              to="/service/tables"
              className="flex items-center gap-1 text-sm font-medium"
            >
              <ChevronLeft className="h-4 w-4" />
              Tischauswahl
            </Link>
          ) : (
            <span className="text-sm font-bold">Tischauswahl</span>
          )}
        </div>
        <UserDropdown />
      </header>
      <main className="flex-1 px-4 py-2 md:px-8 md:py-4 xl:px-12 xl:py-6">
        <Outlet />
      </main>
    </div>
  )
}
