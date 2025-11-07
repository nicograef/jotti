import { UserTable } from "@/admin/users/UserTable"
import { Card } from "@/components/ui/card"
import { NewUserDialog } from "./NewUserDialog"

export function Users() {
  return (
    <>
      <NewUserDialog />
      <div className="flex p-8 justify-center">
        <Card className="w-full p-4">
          <UserTable />
        </Card>
      </div>
    </>
  )
}
