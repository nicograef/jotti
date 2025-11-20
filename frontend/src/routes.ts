import { createBrowserRouter, redirect } from 'react-router'

import { AuthSingleton } from '@/lib/auth'

import { AdminLayout } from './admin/AdminLayout'
import { AdminProductsPage } from './admin/products/AdminProductsPage'
import { AdminTablesPage } from './admin/tables/AdminTablesPage'
import { AdminUsersPage } from './admin/users/AdminUsersPage'
import App from './App'
import { LoginPage } from './pages/LoginPage'
import { PasswordPage } from './pages/PasswordPage'

function AuthRedirect() {
  if (AuthSingleton.isAuthenticated && AuthSingleton.isAdmin) {
    return redirect('/admin')
  } else if (AuthSingleton.isAuthenticated) {
    return redirect('/')
  }
}

export function AdminGuard() {
  if (!AuthSingleton.isAuthenticated || !AuthSingleton.isAdmin) {
    return redirect('/')
  }
}

export const router = createBrowserRouter([
  {
    path: '/',
    Component: App,
    children: [
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
      { path: '', loader: () => redirect('login') },
    ],
  },
])
