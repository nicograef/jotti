import { UserTable } from "@/admin/users/UserTable"
import { Card } from "@/components/ui/card"
import { NewUserDialog } from "./NewUserDialog"
import { UserCreatedDialog } from "./UserCreatedDialog"
import { useEffect, useState } from "react"
import type { User } from "@/lib/user"
import { BackendSingleton } from "@/lib/backend"

const initialUserCreatedState = {
  user: null as User | null,
  onetimePassword: "",
  open: false,
}

export function Users() {
  const [loading, setLoading] = useState(true)
  const [users, setUsers] = useState<User[]>([])
  const [userCreatedState, setUserCreatedState] = useState(
    initialUserCreatedState,
  )

  useEffect(() => {
    async function fetchUsers() {
      const response = await BackendSingleton.getUsers()
      setUsers(response)
      setLoading(false)
    }
    fetchUsers()
  }, [])

  return (
    <>
      <NewUserDialog
        created={(user, onetimePassword) => {
          setUsers((prevUsers) => [...prevUsers, user])
          setUserCreatedState({ user, onetimePassword, open: true })
        }}
      />
      <UserCreatedDialog
        {...userCreatedState}
        close={() => setUserCreatedState(initialUserCreatedState)}
      />
      <Card className="p-0">
        <UserTable users={users} loading={loading} />
      </Card>
    </>
  )
}
