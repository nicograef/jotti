import {
  FileText,
  Lamp,
  Landmark,
  LayoutDashboard,
  LogOut,
  type LucideIcon,
  Moon,
  Printer,
  Sun,
  Users,
  Utensils,
  Wallet,
} from 'lucide-react'
import { NavLink, useLocation, useNavigate } from 'react-router'

import { useTheme } from '@/components/theme-provider'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/components/ui/sidebar'
import { AuthSingleton } from '@/lib/Auth'

import { useVersion } from './hooks'

const reportingItems = [
  {
    title: 'Live-Dashboard',
    url: '/admin/auswertung',
    icon: LayoutDashboard,
  },
  {
    title: 'Kassenberichte',
    url: '/admin/kassenberichte',
    icon: FileText,
  },
]

const adminItems = [
  {
    title: 'Produkte',
    url: '/admin/produkte',
    icon: Utensils,
  },
  {
    title: 'Tische',
    url: '/admin/tische',
    icon: Lamp,
  },
  {
    title: 'Benutzer',
    url: '/admin/benutzer',
    icon: Users,
  },
  {
    title: 'Kasse',
    url: '/admin/kasse',
    icon: Wallet,
  },
  {
    title: 'Druckstationen',
    url: '/admin/druckstationen',
    icon: Printer,
  },
  {
    title: 'Finanzamt',
    url: '/admin/finanzamt',
    icon: Landmark,
  },
]

const serviceItems = [
  {
    title: 'Tischauswahl',
    url: '/service/tische',
    icon: Lamp,
  },
]

function NavGroup({
  label,
  items,
}: {
  label: string
  items: { title: string; url: string; icon: LucideIcon }[]
}) {
  const location = useLocation()

  return (
    <SidebarGroup>
      <SidebarGroupLabel>{label}</SidebarGroupLabel>
      <SidebarGroupContent>
        <SidebarMenu>
          {items.map((item) => (
            <SidebarMenuItem key={item.title}>
              <SidebarMenuButton
                asChild
                isActive={location.pathname === item.url}
              >
                <NavLink to={item.url}>
                  <item.icon />
                  <span>{item.title}</span>
                </NavLink>
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  )
}

export function AdminSidebar() {
  const navigate = useNavigate()
  const version = useVersion()
  const { isDark, setTheme } = useTheme()
  const toggleTheme = () => {
    setTheme(isDark ? 'light' : 'dark')
  }

  const logout = () => {
    AuthSingleton.logout()
    void navigate('/login')
  }

  return (
    <Sidebar collapsible="offcanvas">
      <SidebarHeader>
        <h1 className="text-4xl text-center font-extrabold">jotti</h1>
      </SidebarHeader>
      <SidebarContent>
        <NavGroup label="Auswertungen" items={reportingItems} />
        <NavGroup label="Verwaltung" items={adminItems} />
        <NavGroup label="Service" items={serviceItems} />
        <SidebarGroup>
          <SidebarGroupLabel>Einstellungen</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton onClick={toggleTheme}>
                  {isDark ? <Sun /> : <Moon />}
                  <span>{isDark ? 'Hell' : 'Dunkel'}</span>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton onClick={logout}>
                  <LogOut />
                  <span>Abmelden</span>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter>
        {version !== undefined && (
          <p className="text-center text-sm text-muted-foreground">
            jotti {version}
          </p>
        )}
        <p className="text-center text-sm text-muted-foreground">
          Entwickelt von{' '}
          <a
            href="https://nicograef.de"
            target="_blank"
            rel="noopener noreferrer"
            className="hover:underline"
          >
            Nico Gräf
          </a>
        </p>
      </SidebarFooter>
    </Sidebar>
  )
}
