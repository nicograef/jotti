import { createBrowserRouter, redirect } from "react-router"
import App from "./App"
import { LoginPage } from "./pages/LoginPage"
import { AdminUsersPage } from "./pages/AdminUsersPage"
import { AdminProductsPage } from "./pages/AdminProductsPage"
import { AdminTablesPage } from "./pages/AdminTablesPage"
import { PasswordPage } from "./pages/PasswordPage"

import { AuthSingleton } from "@/lib/auth"
import { AdminLayout } from "./admin/AdminLayout"

function AuthRedirect() {
  if (AuthSingleton.isAuthenticated && AuthSingleton.isAdmin) {
    return redirect("/admin")
  } else if (AuthSingleton.isAuthenticated) {
    return redirect("/")
  }
}

export function AdminGuard() {
  if (!AuthSingleton.isAuthenticated || !AuthSingleton.isAdmin) {
    return redirect("/")
  }
}

export const router = createBrowserRouter([
  {
    path: "/",
    Component: App,
    children: [
      { path: "login", Component: LoginPage, loader: AuthRedirect },
      { path: "set-password", Component: PasswordPage, loader: AuthRedirect },
      {
        path: "admin",
        Component: AdminLayout,
        loader: AdminGuard,
        children: [
          { path: "orders", Component: AdminUsersPage },
          { path: "products", Component: AdminProductsPage },
          { path: "tables", Component: AdminTablesPage },
          {
            path: "users",
            Component: AdminUsersPage,
          },
          { path: "", loader: () => redirect("orders") },
        ],
      },
      { path: "", loader: () => redirect("login") },
    ],
  },
])
