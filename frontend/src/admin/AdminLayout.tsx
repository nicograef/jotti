import { Outlet } from "react-router"
import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar"
import { AdminSidebar } from "./AdminSidebar"

export function AdminLayout() {
  return (
    <SidebarProvider>
      <AdminSidebar />
      <main className="min-h-screen w-full">
        <SidebarTrigger />
        <div className="p-4 md:p-8 lg:p-12">
          <Outlet />
        </div>
      </main>
    </SidebarProvider>
  )
}
