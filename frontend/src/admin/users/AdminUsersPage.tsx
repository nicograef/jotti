import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { toast } from 'sonner'

import { useActionSubmit } from '@/hooks/use-action-submit'
import { BackendSingleton } from '@/lib/Backend'

import { AdminPageHeader } from '../components/AdminPageHeader'
import { EditUserDialog } from './EditUserDialog'
import { HelferPanels } from './HelferPanels'
import { ALLE_USERS_KEY, useAllUsers } from './hooks'
import { NewUserDialog } from './NewUserDialog'
import { PasswordResetDialog } from './PasswordResetDialog'
import { type User, UserStatus } from './User'
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

  const { run: runResetPassword } = useActionSubmit({
    actionLabel: 'Passwort zurücksetzen',
  })

  const invalidateUsers = () =>
    void queryClient.invalidateQueries({ queryKey: [ALLE_USERS_KEY] })

  // Passwort-Reset direkt aus dem Zeilen-Menü (Design-Handoff 1e): setzt das
  // Passwort zurück und zeigt das neue Einmalpasswort im bestehenden Dialog.
  const resetPassword = async (userId: number) => {
    const user = users.find((u) => u.id === userId)
    if (!user) return
    await runResetPassword(async () => {
      const onetimePassword = await userBackend.resetPassword(userId)
      setPasswordResetState({
        username: user.username,
        onetimePassword,
        open: true,
      })
    })
  }

  const aktiveAnzahl = users.filter(
    (u) => u.status === UserStatus.ACTIVE,
  ).length
  const unterzeile = `${String(users.length)} ${users.length === 1 ? 'Zugang' : 'Zugänge'} · ${String(aktiveAnzahl)} aktiv`

  return (
    <>
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
          close={() => {
            setEditState(initialEditState)
          }}
        />
      )}
      <AdminPageHeader
        titel="Helfer & Zugänge"
        unterzeile={unterzeile}
        glowFarben={['blue', 'red']}
        aktionen={
          <NewUserDialog
            backend={userBackend}
            created={(user, onetimePassword) => {
              invalidateUsers()
              setUserCreatedState({ user, onetimePassword, open: true })
              toast.success(`Neuer Benutzer "${user.name}" wurde erstellt.`)
            }}
          />
        }
      />
      <div className="mt-4 grid grid-cols-1 items-start gap-5 lg:grid-cols-[1fr_320px]">
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
          onResetPassword={resetPassword}
          onDeleted={() => {
            invalidateUsers()
            toast.success('Benutzer wurde gelöscht.')
          }}
        />
        <HelferPanels />
      </div>
    </>
  )
}
