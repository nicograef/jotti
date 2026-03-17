import {
  createBrowserRouter,
  type LoaderFunctionArgs,
  redirect,
} from 'react-router'

import { AuthSingleton } from '@/lib/Auth'

import { AdminLayout } from './admin/AdminLayout'
import { DruckerConfigPage } from './admin/DruckerConfigPage'
import { AdminProductsPage } from './admin/products/AdminProductsPage'
import { AdminDashboardPage } from './admin/reporting/AdminDashboardPage'
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
  const tableId = Number(params.tableId)
  if (!Number.isInteger(tableId) || tableId <= 0) {
    return redirect('/service/tables')
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
          { index: true, loader: () => redirect('reporting') },
          { path: 'reporting', Component: AdminDashboardPage },
          { path: 'products', Component: AdminProductsPage },
          { path: 'tables', Component: AdminTablesPage },
          { path: 'users', Component: AdminUsersPage },
          { path: 'drucker', Component: DruckerConfigPage },
        ],
      },
      {
        path: 'service',
        Component: ServiceLayout,
        loader: ServiceGuard,
        children: [
          { index: true, loader: () => redirect('tables') },
          { path: 'tables', Component: TableSelectionPage },
          {
            path: 'tables/:tableId',
            Component: TablePage,
            loader: ServiceTableGuard,
          },
          { path: '', loader: () => redirect('tables') },
        ],
      },
      { path: '', loader: () => redirect('login') },
    ],
  },
])
