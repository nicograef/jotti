import { ChevronLeft } from 'lucide-react'
import { Link, Outlet, useMatch } from 'react-router'

import { UserDropdown } from '@/components/common/UserDropdown'

// Höhe einer Seite mit fester Spalte ab lg: Viewport minus Header (h-14) und
// vertikalem Main-Padding (md:py-4, xl:py-6); ändert sich eins, muss sie mit.
export const seitenHoehe = 'h-[calc(100dvh-5.5rem)] xl:h-[calc(100dvh-6.5rem)]'

export function ServiceLayout() {
  const onTableDetail = useMatch('/service/tische/:tischId')
  const onDirektverkauf = useMatch('/service/direktverkauf')
  const titel = onDirektverkauf ? 'Direktverkauf' : 'Meine Tische'

  return (
    <div className="min-h-screen flex flex-col">
      <header className="sticky top-0 h-14 border-b bg-background z-40 flex items-center justify-between px-4">
        <div className="flex items-center gap-2">
          {onTableDetail ? (
            <Link
              to="/service/tische"
              className="flex items-center gap-1 text-sm font-medium"
            >
              <ChevronLeft className="h-4 w-4" />
              Meine Tische
            </Link>
          ) : (
            <span className="text-sm font-bold">{titel}</span>
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
