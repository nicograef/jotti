import { useEffect, useState } from 'react'

import { UserTable } from '@/admin/users/UserTable'
import { Card } from '@/components/ui/card'
import { BackendSingleton } from '@/lib/backend'
import type { User } from '@/lib/user'

import { EditUserDialog } from '../users/EditUserDialog'
import { NewUserDialog } from '../users/NewUserDialog'
import { UserCreatedDialog } from '../users/UserCreatedDialog'

const initialUserCreatedState = {
  user: null as User | null,
  onetimePassword: '',
  open: false,
}

const initialEditUserState = {
  user: null as User | null,
  open: false,
}

export function Users() {
  const [loading, setLoading] = useState(true)
  const [users, setUsers] = useState<User[]>([])
  const [userCreatedState, setUserCreatedState] = useState(
    initialUserCreatedState,
  )
  const [editUserState, setEditUserState] = useState(initialEditUserState)

  useEffect(() => {
    async function fetchUsers() {
      const response = await BackendSingleton.getUsers()
      setUsers(response)
      setLoading(false)
    }
    void fetchUsers()
  }, [])

  const updateUser = (user: User) => {
    setUsers((prevUsers) => prevUsers.map((u) => (u.id === user.id ? user : u)))
  }

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
        close={() => {
          setUserCreatedState(initialUserCreatedState)
        }}
      />
      {editUserState.user && (
        <EditUserDialog
          open={editUserState.open}
          user={editUserState.user}
          updated={(user) => {
            updateUser(user)
          }}
          close={() => {
            setEditUserState(initialEditUserState)
          }}
        />
      )}
      <Card className="p-0">
        <UserTable
          users={users}
          loading={loading}
          onClick={(user) => {
            setEditUserState({ user, open: true })
          }}
        />
      </Card>
    </>
  )
}
