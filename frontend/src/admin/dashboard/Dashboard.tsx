import { AppBar } from "@/admin/AppBar"

export function Dashboard() {
  return (
    <>
      <AppBar activeTab="dashboard" />
      <div className="flex p-8 pt-24">Willkommen</div>
    </>
  )
}
