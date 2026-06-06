import {
  createBrowserRouter,
  type LoaderFunctionArgs,
  redirect,
} from 'react-router'

import { AuthSingleton } from '@/lib/Auth'

import App from './App'
import { LoginPage } from './pages/LoginPage'
import { PasswordPage } from './pages/PasswordPage'

function AuthRedirect() {
  if (AuthSingleton.isAuthenticated && AuthSingleton.isAdmin) {
    return redirect('/admin')
  } else if (
    AuthSingleton.isAuthenticated &&
    (AuthSingleton.isService || AuthSingleton.isServiceleitung)
  ) {
    return redirect('/service')
  } else if (AuthSingleton.isAuthenticated) {
    return redirect('/')
  }
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
}

export const router = createBrowserRouter([
  {
    path: '/',
    Component: App,
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
            path: 'drucker',
            lazy: async () => ({
              Component: (await import('./admin/settings/DruckerConfigPage'))
                .DruckerConfigPage,
            }),
          },
          {
            path: 'einstellungen',
            lazy: async () => ({
              Component: (await import('./admin/settings/EinstellungenPage'))
                .EinstellungenPage,
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
          { index: true, loader: () => redirect('tische') },
          {
            path: 'tische',
            lazy: async () => ({
              Component: (await import('./service/TableSelectionPage'))
                .TableSelectionPage,
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
    ],
  },
])
