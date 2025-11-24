import { createBrowserRouter, redirect } from 'react-router'

import { AuthSingleton } from '@/lib/auth'

import { AdminLayout } from './admin/AdminLayout'
import { AdminProductsPage } from './admin/products/AdminProductsPage'
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
  } else if (AuthSingleton.isAuthenticated && AuthSingleton.isService) {
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
  const isServiceOrdAdmin =
    AuthSingleton.isAuthenticated &&
    (AuthSingleton.isService || AuthSingleton.isAdmin)

  if (!isServiceOrdAdmin) {
    return redirect('/')
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
          { path: 'products', Component: AdminProductsPage },
          { path: 'tables', Component: AdminTablesPage },
          { path: 'users', Component: AdminUsersPage },
        ],
      },
      {
        path: 'service',
        Component: ServiceLayout,
        loader: ServiceGuard,
        children: [
          { index: true, loader: () => redirect('tables') },
          { path: 'tables', Component: TableSelectionPage },
          { path: 'tables/:tableId', Component: TablePage },
          { path: '', loader: () => redirect('tables') },
        ],
      },
      { path: '', loader: () => redirect('login') },
    ],
  },
])
