import { createBrowserRouter, redirect } from "react-router"
import App from "./App"
import { LoginPage } from "./pages/LoginPage"
import { AdminDashboard } from "./pages/AdminDashboardPage"
import { AdminUsersPage } from "./pages/AdminUsersPage"
import { AdminProductsPage } from "./pages/AdminProductsPage"
import { AdminTablesPage } from "./pages/AdminTablesPage"

export const router = createBrowserRouter([
  {
    path: "/",
    Component: App,
    children: [
      { index: true, Component: LoginPage },
      { path: "login", Component: LoginPage },
      {
        path: "admin",
        children: [
          { path: "dashboard", Component: AdminDashboard },
          { path: "products", Component: AdminProductsPage },
          { path: "tables", Component: AdminTablesPage },
          { path: "users", Component: AdminUsersPage },
          { path: "", loader: () => redirect("dashboard") },
        ],
      },
    ],
  },
])
