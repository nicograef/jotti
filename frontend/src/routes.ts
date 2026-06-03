import {
  createBrowserRouter,
  type LoaderFunctionArgs,
  redirect,
} from 'react-router'

import { AuthSingleton } from '@/lib/Auth'

import { AdminLayout } from './admin/AdminLayout'
import { KassensitzungPage } from './admin/kasse/KassensitzungPage'
import { AdminProductsPage } from './admin/products/AdminProductsPage'
import { AdminDashboardPage } from './admin/reporting/AdminDashboardPage'
import { DruckerConfigPage } from './admin/settings/DruckerConfigPage'
import { EinstellungenPage } from './admin/settings/EinstellungenPage'
import { AdminTablesPage } from './admin/tables/AdminTablesPage'
import { AdminUsersPage } from './admin/users/AdminUsersPage'
import App from './App'
import { LoginPage } from './pages/LoginPage'
import { PasswordPage } from './pages/PasswordPage'
import { ServiceLayout } from './service/ServiceLayout'
import { TablePage } from './service/TablePage'
import { TableSelectionPage } from './service/TableSelectionPage'

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
        Component: AdminLayout,
        loader: AdminGuard,
        children: [
          { index: true, loader: () => redirect('auswertung') },
          { path: 'auswertung', Component: AdminDashboardPage },
          { path: 'produkte', Component: AdminProductsPage },
          { path: 'tische', Component: AdminTablesPage },
          { path: 'benutzer', Component: AdminUsersPage },
          { path: 'kasse', Component: KassensitzungPage },
          { path: 'drucker', Component: DruckerConfigPage },
          { path: 'einstellungen', Component: EinstellungenPage },
        ],
      },
      {
        path: 'service',
        Component: ServiceLayout,
        loader: ServiceGuard,
        children: [
          { index: true, loader: () => redirect('tische') },
          { path: 'tische', Component: TableSelectionPage },
          {
            path: 'tische/:tischId',
            Component: TablePage,
            loader: ServiceTableGuard,
          },
          { path: '', loader: () => redirect('tische') },
        ],
      },
      { path: '', loader: () => redirect('login') },
    ],
  },
])
