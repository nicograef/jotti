import { AppBar } from "@/admin/AppBar"
import { UserTable } from "@/admin/users/UserTable"
import { Card } from "@/components/ui/card"
import { NewUserDialog } from "./NewUserDialog"

export function Users() {
  return (
    <>
      <AppBar activeTab="users" />
      <NewUserDialog />
      <div className="flex p-8 pt-24 justify-center">
        <Card className="w-full p-4 max-w-4xl">
          <UserTable />
        </Card>
      </div>
    </>
  )
}
