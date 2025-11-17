import { useEffect, useState } from 'react'

import { BackendSingleton } from '@/lib/Backend'
import { type User, UserBackend, UserStatus } from '@/lib/UserBackend'

import { EditUserDialog } from '../users/EditUserDialog'
import { NewUserDialog } from '../users/NewUserDialog'
import { UserCreatedDialog } from '../users/UserCreatedDialog'
import { Users } from './Users'

const initialUserCreatedState = {
  user: null as User | null,
  onetimePassword: '',
  open: false,
}

const initialEditUserState = {
  user: null as User | null,
  open: false,
}

const userBackend = new UserBackend(BackendSingleton)

export function AdminUsersPage() {
  const [loading, setLoading] = useState(false)
  const [users, setUsers] = useState<User[]>([])
  const [userCreatedState, setUserCreatedState] = useState(
    initialUserCreatedState,
  )
  const [editUserState, setEditUserState] = useState(initialEditUserState)

  useEffect(() => {
    async function fetchUsers() {
      const response = await userBackend.getUsers()
      setUsers(response)
      setLoading(false)
    }
    void fetchUsers()
  }, [])

  const updateUser = (user: User) => {
    setUsers((prevUsers) => prevUsers.map((u) => (u.id === user.id ? user : u)))
  }

  const onStatusChange = (userId: number, status: UserStatus) => {
    setUsers((prevUsers) =>
      prevUsers.map((u) => (u.id === userId ? { ...u, status } : u)),
    )
  }

  return (
    <>
      <NewUserDialog
        backend={userBackend}
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
          backend={userBackend}
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
      <h1 className="text-2xl font-bold">Benutzerverwaltung</h1>
      <Users
        loading={loading}
        backend={userBackend}
        users={users}
        onEdit={(userId) => {
          const userToEdit = users.find((u) => u.id === userId) ?? null
          setEditUserState({ user: userToEdit, open: true })
        }}
        onStatusChange={onStatusChange}
      />
    </>
  )
}
