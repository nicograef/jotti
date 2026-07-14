import {
  createBrowserRouter,
  type LoaderFunctionArgs,
  redirect,
} from 'react-router'

import { getArbeitsmodus, setArbeitsmodus } from '@/lib/arbeitsmodus'
import { AuthSingleton } from '@/lib/Auth'

import App from './App'
import { ErrorPage, NotFoundPage } from './pages/ErrorPage'
import { HydrateFallbackPage } from './pages/HydrateFallbackPage'
import { LoginPage } from './pages/LoginPage'
import { PasswordPage } from './pages/PasswordPage'

function AuthRedirect() {
  if (!AuthSingleton.isAuthenticated) return

  if (AuthSingleton.isAdmin) {
    return redirect('/admin')
  }
  if (AuthSingleton.isService || AuthSingleton.isServiceleitung) {
    return redirect('/service')
  }
  return redirect('/')
}

export function AdminGuard() {
  if (!AuthSingleton.isAuthenticated || !AuthSingleton.isAdmin) {
    return redirect('/')
  }
}

export function ServiceGuard() {
  const hasServiceAccess =
    AuthSingleton.isAuthenticated &&
    (AuthSingleton.isService ||
      AuthSingleton.isServiceleitung ||
      AuthSingleton.isAdmin)

  if (!hasServiceAccess) {
    return redirect('/')
  }
}

export function ServiceTableGuard({ params }: LoaderFunctionArgs) {
  const tischId = Number(params.tischId)
  if (!Number.isInteger(tischId) || tischId <= 0) {
    return redirect('/service/tische')
  }
  setArbeitsmodus('tischservice')
}

// Service-Einstieg: in den zuletzt genutzten Modus weiterleiten.
export function ServiceIndexRedirect() {
  return redirect(
    getArbeitsmodus() === 'direktverkauf' ? 'direktverkauf' : 'tische',
  )
}

// Beim Besuch einer Modus-Route den zugehörigen Modus persistieren
// (auch Deep-Links und Lesezeichen zählen so als „zuletzt genutzt").
export function ServiceTischauswahlLoader() {
  setArbeitsmodus('tischservice')
  return null
}

export function ServiceDirektverkaufLoader() {
  setArbeitsmodus('direktverkauf')
  return null
}

export const router = createBrowserRouter([
  {
    path: '/',
    Component: App,
    ErrorBoundary: ErrorPage,
    HydrateFallback: HydrateFallbackPage,
    children: [
      { index: true, loader: () => redirect('login') },
      { path: 'login', Component: LoginPage, loader: AuthRedirect },
      { path: 'set-password', Component: PasswordPage, loader: AuthRedirect },
      {
        path: 'admin',
        lazy: async () => ({
          Component: (await import('./admin/AdminLayout')).AdminLayout,
        }),
        loader: AdminGuard,
        children: [
          { index: true, loader: () => redirect('auswertung') },
          {
            path: 'auswertung',
            lazy: async () => ({
              Component: (await import('./admin/reporting/AdminDashboardPage'))
                .AdminDashboardPage,
            }),
          },
          {
            path: 'kassenberichte',
            lazy: async () => ({
              Component: (await import('./admin/reporting/KassenberichtePage'))
                .KassenberichtePage,
            }),
          },
          {
            path: 'produkte',
            lazy: async () => ({
              Component: (await import('./admin/products/AdminProductsPage'))
                .AdminProductsPage,
            }),
          },
          {
            path: 'tische',
            lazy: async () => ({
              Component: (await import('./admin/tables/AdminTablesPage'))
                .AdminTablesPage,
            }),
          },
          {
            path: 'benutzer',
            lazy: async () => ({
              Component: (await import('./admin/users/AdminUsersPage'))
                .AdminUsersPage,
            }),
          },
          {
            path: 'kasse',
            lazy: async () => ({
              Component: (await import('./admin/kasse/KassensitzungPage'))
                .KassensitzungPage,
            }),
          },
          {
            path: 'druckstationen',
            lazy: async () => ({
              Component: (
                await import('./admin/settings/DruckstationConfigPage')
              ).DruckstationConfigPage,
            }),
          },
          {
            path: 'finanzamt',
            lazy: async () => ({
              Component: (await import('./admin/finanzamt/FinanzamtPage'))
                .FinanzamtPage,
            }),
          },
          {
            path: 'tse-einrichtung',
            lazy: async () => ({
              Component: (await import('./admin/tse/TSEEinrichtungPage'))
                .TSEEinrichtungPage,
            }),
          },
        ],
      },
      {
        path: 'service',
        lazy: async () => ({
          Component: (await import('./service/ServiceLayout')).ServiceLayout,
        }),
        loader: ServiceGuard,
        children: [
          { index: true, loader: ServiceIndexRedirect },
          {
            path: 'tische',
            loader: ServiceTischauswahlLoader,
            lazy: async () => ({
              Component: (await import('./service/TableSelectionPage'))
                .TableSelectionPage,
            }),
          },
          {
            path: 'direktverkauf',
            loader: ServiceDirektverkaufLoader,
            lazy: async () => ({
              Component: (await import('./service/DirektverkaufPage'))
                .DirektverkaufPage,
            }),
          },
          {
            path: 'tische/:tischId',
            lazy: async () => ({
              Component: (await import('./service/TablePage')).TablePage,
            }),
            loader: ServiceTableGuard,
          },
        ],
      },
      { path: '*', Component: NotFoundPage },
    ],
  },
])
