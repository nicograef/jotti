import { createBrowserRouter } from "react-router"
import App from "./App"
import Login from "./pages/Login"

export const router = createBrowserRouter([
  {
    path: "/",
    Component: App,
    children: [{ index: true, Component: Login }],
  },
])
