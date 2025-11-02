import { Button } from "@/components/ui/button"
import { AuthSingleton } from "@/lib/auth"
import { NavLink, useNavigate } from "react-router"

type Pages = "dashboard" | "products" | "tables" | "users"

interface AppBarProps {
  activeTab?: Pages
}

export function AppBar({ activeTab }: AppBarProps) {
  const navigate = useNavigate()

  const handleLogout = async () => {
    AuthSingleton.logout()
    await navigate("/login")
  }

  return (
    <header className="fixed inset-x-0 top-0 z-50 shadow-md">
      <div className="mx-auto flex h-16 items-center gap-6 px-8">
        {/* Left: brand/logo */}
        <div className="shrink-0 font-semibold">jotti</div>

        {/* Center: navigation (full-width, centered) */}
        <div className="flex-1 flex justify-center gap-8">
          <Button
            asChild
            variant={activeTab === "dashboard" ? "default" : "ghost"}
          >
            <NavLink to="/admin/dashboard">Dashboard</NavLink>
          </Button>
          <Button
            asChild
            variant={activeTab === "products" ? "default" : "ghost"}
          >
            <NavLink to="/admin/products">Produkte</NavLink>
          </Button>
          <Button
            asChild
            variant={activeTab === "tables" ? "default" : "ghost"}
          >
            <NavLink to="/admin/tables">Tische</NavLink>
          </Button>
          <Button asChild variant={activeTab === "users" ? "default" : "ghost"}>
            <NavLink to="/admin/users">Benutzer</NavLink>
          </Button>
        </div>

        <div className="ml-auto flex items-center gap-2">
          <div className="shrink-0 font-semibold">nicog</div>
          <div className="shrink-0">/</div>
          <div
            className="shrink-0 cursor-pointer font-semibold hover:underline"
            onClick={(e) => {
              e.preventDefault()
              void handleLogout()
            }}
          >
            Logout
          </div>
        </div>
      </div>
    </header>
  )
}
