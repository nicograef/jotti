import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { toast } from 'sonner'

import { BackendSingleton } from '@/lib/Backend'

import { EditUserDialog } from './EditUserDialog'
import { ALLE_USERS_KEY, useAllUsers } from './hooks'
import { NewUserDialog } from './NewUserDialog'
import { PasswordResetDialog } from './PasswordResetDialog'
import type { User } from './User'
import { UserBackend } from './UserBackend'
import { UserCreatedDialog } from './UserCreatedDialog'
import { Users } from './Users'

const initialUserCreatedState = {
  user: null as User | null,
  onetimePassword: '',
  open: false,
}

const initialPasswordResetState = {
  username: null as string | null,
  onetimePassword: '',
  open: false,
}

const initialEditState = {
  user: null as User | null,
  open: false,
}

const userBackend = new UserBackend(BackendSingleton)

export function AdminUsersPage() {
  const queryClient = useQueryClient()
  const { isPending, users } = useAllUsers()
  const [userCreatedState, setUserCreatedState] = useState(
    initialUserCreatedState,
  )
  const [passwordResetState, setPasswordResetState] = useState(
    initialPasswordResetState,
  )
  const [editState, setEditState] = useState(initialEditState)

  const invalidateUsers = () =>
    void queryClient.invalidateQueries({ queryKey: [ALLE_USERS_KEY] })

  return (
    <>
      <NewUserDialog
        backend={userBackend}
        created={(user, onetimePassword) => {
          invalidateUsers()
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
      <PasswordResetDialog
        {...passwordResetState}
        close={() => {
          setPasswordResetState(initialPasswordResetState)
        }}
      />
      {editState.user && (
        <EditUserDialog
          backend={userBackend}
          open={editState.open}
          user={editState.user}
          updated={() => {
            invalidateUsers()
          }}
          onPasswordReset={(username, onetimePassword) => {
            setPasswordResetState({ username, onetimePassword, open: true })
          }}
          close={() => {
            setEditState(initialEditState)
          }}
        />
      )}
      <h1 className="text-2xl font-bold">Benutzer verwalten</h1>
      <Users
        loading={isPending}
        backend={userBackend}
        users={users}
        onEdit={(userId) => {
          const userToEdit = users.find((u) => u.id === userId) ?? null
          setEditState({ user: userToEdit, open: true })
        }}
        onStatusChange={() => {
          invalidateUsers()
        }}
        onDeleted={() => {
          invalidateUsers()
          toast.success('Benutzer wurde gelöscht.')
        }}
      />
    </>
  )
}
