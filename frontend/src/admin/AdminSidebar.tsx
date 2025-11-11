import { Lamp, ReceiptText, Users, Utensils } from 'lucide-react'
import { NavLink, useNavigate } from 'react-router'
import { useLocation } from 'react-router'

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
    url: '/admin/orders',
    icon: ReceiptText,
  },
  {
    title: 'Produkte',
    url: '/admin/products',
    icon: Utensils,
  },
  {
    title: 'Tische',
    url: '/admin/tables',
    icon: Lamp,
  },
  {
    title: 'Benutzer',
    url: '/admin/users',
    icon: Users,
  },
]

export function AdminSidebar() {
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
          <SidebarGroupLabel>admin</SidebarGroupLabel>
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
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton onClick={logout}>
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
