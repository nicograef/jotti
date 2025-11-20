import { Lamp, LogOut, ReceiptText, Utensils } from 'lucide-react'
import { NavLink, useNavigate } from 'react-router'
import { useLocation } from 'react-router'

import { ModeToggle } from '@/components/mode-tootle'
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
import { AuthSingleton } from '@/lib/auth'

const items = [
  {
    title: 'Bestellungen',
    url: '/service/orders',
    icon: ReceiptText,
  },
  {
    title: 'Produkte',
    url: '/service/products',
    icon: Utensils,
  },
  {
    title: 'Tische',
    url: '/service/tables',
    icon: Lamp,
  },
]

export function ServiceSidebar() {
  const location = useLocation()
  const navigate = useNavigate()

  const logout = () => {
    AuthSingleton.logout()
    void navigate('/login')
  }

  return (
    <Sidebar>
      <SidebarHeader>
        <h1 className="text-4xl text-center font-extrabold">jotti</h1>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Verwaltung</SidebarGroupLabel>
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
        <SidebarGroup>
          <SidebarGroupLabel>Einstellungen</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <ModeToggle />
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
