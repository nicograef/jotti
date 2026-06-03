import {
  Lamp,
  LayoutDashboard,
  LogOut,
  Moon,
  Printer,
  Settings2,
  Sun,
  Users,
  Utensils,
  Wallet,
} from 'lucide-react'
import { NavLink, useNavigate } from 'react-router'
import { useLocation } from 'react-router'

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

const reportingItems = [
  {
    title: 'Dashboard',
    url: '/admin/auswertung',
    icon: LayoutDashboard,
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
    title: 'Drucker',
    url: '/admin/drucker',
    icon: Printer,
  },
  {
    title: 'Einstellungen',
    url: '/admin/einstellungen',
    icon: Settings2,
  },
]

const serviceItems = [
  {
    title: 'Tischauswahl',
    url: '/service/tische',
    icon: Lamp,
  },
]

export function AdminSidebar() {
  const location = useLocation()
  const navigate = useNavigate()
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
        <SidebarGroup>
          <SidebarGroupLabel>Auswertungen</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {reportingItems.map((item) => (
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
        <SidebarGroup>
          <SidebarGroupLabel>Verwaltung</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {adminItems.map((item) => (
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
        <SidebarGroup>
          <SidebarGroupLabel>Service</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {serviceItems.map((item) => (
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
