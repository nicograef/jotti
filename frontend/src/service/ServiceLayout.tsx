import { ChevronLeft, LogOut, Moon, Sun, User } from 'lucide-react'
import { Link, Outlet, useMatch, useNavigate } from 'react-router'

import { useTheme } from '@/components/theme-provider'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { AuthSingleton } from '@/lib/Auth'

function UserDropdown() {
  const navigate = useNavigate()
  const { theme, setTheme } = useTheme()

  const logout = () => {
    AuthSingleton.logout()
    void navigate('/login')
  }

  const isDark =
    theme === 'dark' ||
    (theme === 'system' &&
      window.matchMedia('(prefers-color-scheme: dark)').matches)

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon">
          <User className="h-5 w-5" />
          <span className="sr-only">Benutzermenü</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem
          onClick={() => {
            setTheme(isDark ? 'light' : 'dark')
          }}
        >
          {isDark ? (
            <Sun className="mr-2 h-4 w-4" />
          ) : (
            <Moon className="mr-2 h-4 w-4" />
          )}
          {isDark ? 'Hell' : 'Dunkel'}
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={logout}>
          <LogOut className="mr-2 h-4 w-4" />
          Abmelden
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

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
