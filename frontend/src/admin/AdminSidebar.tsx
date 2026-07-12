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

import { useFehlgeschlageneDruckauftraege } from '@/admin/settings/hooks'
import { useTSESignaturQueue, useTSEStatus } from '@/admin/tse/hooks'
import { tseAmpel } from '@/admin/tse/tseAmpel'
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

import { StatusDot, type StatusDotZustand } from './components/StatusDot'
import { useVersion } from './hooks'
import { useOffeneKassensitzung } from './kasse/hooks'

interface NavItem {
  title: string
  url: string
  icon: LucideIcon
  status?: { zustand: StatusDotZustand; label: string }
}

function NavGroup({ label, items }: { label: string; items: NavItem[] }) {
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
                  <span className="flex-1">{item.title}</span>
                  {item.status && (
                    <StatusDot
                      zustand={item.status.zustand}
                      label={item.status.label}
                    />
                  )}
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

  const { kassensitzung } = useOffeneKassensitzung()
  const { druckauftraege } = useFehlgeschlageneDruckauftraege()
  const { tseStatus, isPending: tseLoading } = useTSEStatus()
  const { queue } = useTSESignaturQueue()

  const kasseOffen = kassensitzung !== null
  const bondruckerFehler = druckauftraege.length > 0
  // Finanzamt & TSE ist kritisch nach derselben Regel wie die „Läuft alles?"-
  // Karte: tseAmpel ist die Single Source of Truth für den TSE-Fehlerzustand.
  const finanzamtFehler = tseAmpel(tseStatus, tseLoading, queue).fehler

  const heuteItems: NavItem[] = [
    {
      title: 'Übersicht',
      url: '/admin/auswertung',
      icon: LayoutDashboard,
    },
    {
      title: 'Kassentag',
      url: '/admin/kasse',
      icon: Wallet,
      status: kasseOffen ? { zustand: 'ok', label: 'Kasse offen' } : undefined,
    },
  ]

  const vorbereitungItems: NavItem[] = [
    {
      title: 'Produkte & Preise',
      url: '/admin/produkte',
      icon: Utensils,
    },
    {
      title: 'Tische',
      url: '/admin/tische',
      icon: Lamp,
    },
    {
      title: 'Helfer & Zugänge',
      url: '/admin/benutzer',
      icon: Users,
    },
    {
      title: 'Bondrucker',
      url: '/admin/druckstationen',
      icon: Printer,
      status: bondruckerFehler
        ? { zustand: 'fehler', label: 'Druckauftrag fehlgeschlagen' }
        : undefined,
    },
  ]

  const nachDemFestItems: NavItem[] = [
    {
      title: 'Berichte & Export',
      url: '/admin/kassenberichte',
      icon: FileText,
    },
    {
      title: 'Finanzamt & TSE',
      url: '/admin/finanzamt',
      icon: Landmark,
      status: finanzamtFehler
        ? { zustand: 'fehler', label: 'TSE benötigt Aufmerksamkeit' }
        : undefined,
    },
  ]

  const serviceItems: NavItem[] = [
    {
      title: 'Zum Service-Bereich',
      url: '/service/tische',
      icon: LogOut,
    },
  ]

  const kasseStatusText = kasseOffen
    ? `Kasse offen · seit ${new Date(
        kassensitzung.eroeffnetAm,
      ).toLocaleTimeString('de-DE', { hour: '2-digit', minute: '2-digit' })}`
    : 'Kasse geschlossen'

  const toggleTheme = () => {
    setTheme(isDark ? 'light' : 'dark')
  }

  const logout = () => {
    AuthSingleton.logout()
    void navigate('/login')
  }

  return (
    <Sidebar collapsible="offcanvas" className="print:hidden">
      <SidebarHeader className="gap-3">
        <h1 className="px-1 text-3xl font-extrabold">jotti</h1>
        <NavLink
          to="/admin/kasse"
          className="flex flex-col gap-1 rounded-lg border bg-background px-3 py-2.5 shadow-sm"
        >
          <span className="text-sm font-semibold">
            {kassensitzung?.bezeichnung ?? 'Kein Kassentag'}
          </span>
          <span className="flex items-center gap-1.5 text-xs text-muted-foreground">
            <StatusDot
              zustand={kasseOffen ? 'ok' : 'neutral'}
              label={kasseOffen ? 'Kasse offen' : 'Kasse geschlossen'}
            />
            {kasseStatusText}
          </span>
        </NavLink>
      </SidebarHeader>
      <SidebarContent>
        <NavGroup label="Heute" items={heuteItems} />
        <NavGroup label="Vorbereitung" items={vorbereitungItems} />
        <NavGroup label="Nach dem Fest" items={nachDemFestItems} />
        <NavGroup label="Service" items={serviceItems} />
      </SidebarContent>
      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton onClick={toggleTheme}>
              {isDark ? <Sun /> : <Moon />}
              <span>{isDark ? 'Helles Design' : 'Dunkles Design'}</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem>
            <SidebarMenuButton onClick={logout}>
              <LogOut />
              <span>Abmelden</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
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
