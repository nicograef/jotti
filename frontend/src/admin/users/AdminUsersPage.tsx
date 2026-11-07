import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { toast } from 'sonner'

import { useActionSubmit } from '@/hooks/use-action-submit'
import { AuthSingleton } from '@/lib/Auth'
import { EntityStatus } from '@/lib/entityStatus'
import { formatAnzahl } from '@/lib/utils'

import { AdminPageHeader } from '../components/AdminPageHeader'
import { EditUserDialog } from './EditUserDialog'
import { type Einmalcode, EinmalcodeDialog } from './EinmalcodeDialog'
import { HelferPanels } from './HelferPanels'
import { ALLE_USERS_KEY, useAllUsers, userBackend } from './hooks'
import { NewUserDialog } from './NewUserDialog'
import type { User } from './User'
import { rolleLabel } from './UserRolle'
import { Users } from './Users'

export function AdminUsersPage() {
  const queryClient = useQueryClient()
  const { isPending, users } = useAllUsers()
  const [einmalcode, setEinmalcode] = useState<Einmalcode | null>(null)
  const [zuBearbeiten, setZuBearbeiten] = useState<User | null>(null)

  const { run: runResetPassword } = useActionSubmit({
    actionLabel: 'Passwort zurücksetzen',
  })

  const invalidateUsers = () =>
    void queryClient.invalidateQueries({ queryKey: [ALLE_USERS_KEY] })

  const resetPassword = async (userId: number) => {
    const user = users.find((u) => u.id === userId)
    if (!user) return
    await runResetPassword(async () => {
      const onetimePassword = await userBackend.resetPassword(userId)
      setEinmalcode({
        titel: 'Passwort zurückgesetzt!',
        beschreibung: `Das Passwort für ${user.username} wurde zurückgesetzt. Beim nächsten Anmelden muss der Helfer mit dem untenstehenden Code sein neues Passwort setzen.`,
        username: user.username,
        onetimePassword,
      })
    })
  }

  const aktiveAnzahl = users.filter(
    (u) => u.status === EntityStatus.ACTIVE,
  ).length
  const unterzeile = `${formatAnzahl(users.length, 'Zugang', 'Zugänge')} · ${String(aktiveAnzahl)} aktiv`

  return (
    <>
      {einmalcode && (
        <EinmalcodeDialog
          einmalcode={einmalcode}
          close={() => {
            setEinmalcode(null)
          }}
        />
      )}
      {zuBearbeiten && (
        <EditUserDialog
          user={zuBearbeiten}
          isSelf={zuBearbeiten.id === AuthSingleton.userId}
          updated={() => {
            invalidateUsers()
          }}
          close={() => {
            setZuBearbeiten(null)
          }}
        />
      )}
      <AdminPageHeader
        titel="Helfer & Zugänge"
        unterzeile={unterzeile}
        glowFarben={['blue', 'red']}
        aktionen={
          <NewUserDialog
            created={(user, onetimePassword) => {
              invalidateUsers()
              setEinmalcode({
                titel: 'Helfer wurde angelegt!',
                beschreibung: `Für ${user.name} wurde ein ${rolleLabel[user.role]}-Helfer angelegt. Beim erstmaligen Anmelden muss der Helfer mit dem untenstehenden Code sein Passwort setzen.`,
                username: user.username,
                onetimePassword,
              })
              toast.success(`Neuer Helfer "${user.name}" wurde erstellt.`)
            }}
          />
        }
      />
      <div className="mt-4 grid grid-cols-1 items-start gap-5 lg:grid-cols-[1fr_320px]">
        <Users
          loading={isPending}
          users={users}
          onEdit={(user) => {
            setZuBearbeiten(user)
          }}
          onStatusChange={() => {
            invalidateUsers()
          }}
          onResetPassword={resetPassword}
          onDeleted={() => {
            invalidateUsers()
            toast.success('Helfer wurde gelöscht.')
          }}
        />
        <HelferPanels />
      </div>
    </>
  )
}
