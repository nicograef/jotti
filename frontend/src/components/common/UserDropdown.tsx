import {
  ArrowRightLeft,
  LayoutDashboard,
  LogOut,
  Moon,
  Sun,
  User,
} from 'lucide-react'
import { useLocation, useNavigate } from 'react-router'

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

// Wechsel-Eintrag für den Service-Bereich, abgeleitet aus der aktuellen Route.
// Außerhalb von /service erscheint kein Eintrag. Die Geräte-Präferenz wird vom
// Loader der Zielroute geschrieben, nicht hier.
// eslint-disable-next-line react-refresh/only-export-components
export function moduswechselEintrag(
  pathname: string,
): { label: string; ziel: string } | null {
  if (!pathname.startsWith('/service')) return null
  return pathname.startsWith('/service/direktverkauf')
    ? { label: 'Zu Tischservice wechseln', ziel: '/service/tische' }
    : { label: 'Zu Direktverkauf wechseln', ziel: '/service/direktverkauf' }
}

export function UserDropdown() {
  const navigate = useNavigate()
  const location = useLocation()
  const { isDark, setTheme } = useTheme()

  const moduswechsel = moduswechselEintrag(location.pathname)

  const logout = () => {
    AuthSingleton.logout()
    void navigate('/login')
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" className="size-11">
          <User className="h-5 w-5" />
          <span className="sr-only">Benutzermenü</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {AuthSingleton.isAdmin && !location.pathname.startsWith('/admin') && (
          <>
            <DropdownMenuItem onClick={() => void navigate('/admin')}>
              <LayoutDashboard className="mr-2 h-4 w-4" />
              Verwaltung
            </DropdownMenuItem>
            <DropdownMenuSeparator />
          </>
        )}
        {moduswechsel && (
          <>
            <DropdownMenuItem onClick={() => void navigate(moduswechsel.ziel)}>
              <ArrowRightLeft className="mr-2 h-4 w-4" />
              {moduswechsel.label}
            </DropdownMenuItem>
            <DropdownMenuSeparator />
          </>
        )}
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
