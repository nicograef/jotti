import { ChevronLeft } from 'lucide-react'
import { Link, Outlet, useMatch } from 'react-router'

import { UserDropdown } from '@/components/common/UserDropdown'
import { cn } from '@/lib/utils'

// Ein Segment der Modus-Umschaltung: react-router-Link in Tabs-Optik (abgeleitet
// aus tabsListVariants/TabsTrigger). Der aktive Zustand kommt aus der Route, das
// aktive Segment trägt aria-current="page". Höhe 44 px als Tap-Ziel.
function ModusSegment({
  to,
  label,
  aktiv,
}: {
  to: string
  label: string
  aktiv: boolean
}) {
  return (
    <Link
      to={to}
      aria-current={aktiv ? 'page' : undefined}
      className={cn(
        'inline-flex h-11 min-w-20 items-center justify-center rounded-md px-4 text-sm font-medium transition-all focus-visible:outline-1 focus-visible:outline-ring',
        aktiv
          ? 'bg-background text-foreground shadow-sm dark:bg-input/30 dark:text-foreground'
          : 'text-foreground/60 hover:text-foreground dark:text-muted-foreground dark:hover:text-foreground',
      )}
    >
      {label}
    </Link>
  )
}

export function ServiceLayout() {
  const onTableDetail = useMatch('/service/tische/:tischId')
  const onDirektverkauf = useMatch('/service/direktverkauf')

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
            <nav
              aria-label="Arbeitsmodus"
              className="inline-flex items-center rounded-lg bg-muted p-[3px] text-muted-foreground"
            >
              <ModusSegment
                to="/service/tische"
                label="Tische"
                aktiv={onDirektverkauf === null}
              />
              <ModusSegment
                to="/service/direktverkauf"
                label="Theke"
                aktiv={onDirektverkauf !== null}
              />
            </nav>
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
