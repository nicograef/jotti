import { createBrowserRouter } from "react-router"
import App from "./App"
import { LoginPage } from "./pages/LoginPage"
import { AdminPage } from "./pages/AdminPage"

export const router = createBrowserRouter([
  {
    path: "/",
    Component: App,
    children: [
      { index: true, Component: LoginPage },
      { path: "login", Component: LoginPage },
      { path: "admin", Component: AdminPage },
    ],
  },
])
