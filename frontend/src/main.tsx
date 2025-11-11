import "./index.css"

import { StrictMode } from "react"
import { createRoot } from "react-dom/client"
import { RouterProvider } from "react-router/dom"

import { router } from "./routes.ts"

const documentRoot = document.getElementById("root")
if (!documentRoot) throw new Error("Failed to find the root element")

createRoot(documentRoot).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>,
)
