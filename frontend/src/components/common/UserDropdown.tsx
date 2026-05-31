import { LayoutDashboard, LogOut, Moon, Sun, User } from 'lucide-react'
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

export function UserDropdown() {
  const navigate = useNavigate()
  const location = useLocation()
  const { isDark, setTheme } = useTheme()

  const logout = () => {
    AuthSingleton.logout()
    void navigate('/login')
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon">
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
