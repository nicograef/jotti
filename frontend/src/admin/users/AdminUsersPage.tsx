import { useState } from 'react'
import { toast } from 'sonner'

import { BackendSingleton } from '@/lib/Backend'
import { useAllUsers } from '@/lib/user/hooks'
import type { User, UserStatus } from '@/lib/user/User'
import { UserBackend } from '@/lib/user/UserBackend'

import { EditUserDialog } from './EditUserDialog'
import { NewUserDialog } from './NewUserDialog'
import { UserCreatedDialog } from './UserCreatedDialog'
import { Users } from './Users'

const initialUserCreatedState = {
  user: null as User | null,
  onetimePassword: '',
  open: false,
}

const initialEditState = {
  user: null as User | null,
  open: false,
}

const userBackend = new UserBackend(BackendSingleton)

export function AdminUsersPage() {
  const { loading, users, setUsers } = useAllUsers()
  const [userCreatedState, setUserCreatedState] = useState(
    initialUserCreatedState,
  )
  const [editState, setEditState] = useState(initialEditState)

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
          toast.success(`Neuer Benutzer "${user.name}" wurde erstellt.`)
        }}
      />
      <UserCreatedDialog
        {...userCreatedState}
        close={() => {
          setUserCreatedState(initialUserCreatedState)
        }}
      />
      {editState.user && (
        <EditUserDialog
          backend={userBackend}
          open={editState.open}
          user={editState.user}
          updated={(user) => {
            updateUser(user)
          }}
          close={() => {
            setEditState(initialEditState)
          }}
        />
      )}
      <h1 className="text-2xl font-bold">Benutzer verwalten</h1>
      <Users
        loading={loading}
        backend={userBackend}
        users={users}
        onEdit={(userId) => {
          const userToEdit = users.find((u) => u.id === userId) ?? null
          setEditState({ user: userToEdit, open: true })
        }}
        onStatusChange={onStatusChange}
      />
    </>
  )
}
