import { Users as UsersIcon } from 'lucide-react'

import { EmptyState } from '@/components/common/EmptyState'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { AuthSingleton } from '@/lib/Auth'

import { userBackend } from './hooks'
import { type User } from './User'
import { UserRow } from './UserRow'
import { BENUTZER_SPALTEN } from './UsersSpalten'

interface UsersProps {
  loading: boolean
  users: User[]
  onEdit: (user: User) => void
  onStatusChange: () => void
  onResetPassword: (userId: number) => Promise<void>
  onDeleted: (userId: number) => void
}

export function Users(props: UsersProps) {
  const { loading: activateLoading, run: runActivate } = useActionSubmit({
    actionLabel: 'Helfer aktivieren',
  })
  const { loading: deactivateLoading, run: runDeactivate } = useActionSubmit({
    actionLabel: 'Helfer deaktivieren',
  })
  const { loading: deleteLoading, run: runDelete } = useActionSubmit({
    actionLabel: 'Helfer löschen',
  })

  const loading = activateLoading || deactivateLoading || deleteLoading

  const activateUser = async (userId: number) => {
    await runActivate(async () => {
      await userBackend.activateUser(userId)
      props.onStatusChange()
    })
  }

  const deactivateUser = async (userId: number) => {
    await runDeactivate(async () => {
      await userBackend.deactivateUser(userId)
      props.onStatusChange()
    })
  }

  const deleteUser = async (userId: number) => {
    await runDelete(async () => {
      await userBackend.deleteUser(userId)
      props.onDeleted(userId)
    })
  }

  if (props.users.length === 0 && !props.loading) {
    return (
      <EmptyState
        icon={UsersIcon}
        title="Keine Helfer vorhanden"
        description="Erstelle einen neuen Helfer, um loszulegen."
      />
    )
  }

  return (
    <div className="overflow-hidden rounded-lg border">
      <div
        className={`${BENUTZER_SPALTEN} bg-muted px-4 py-2 text-xs font-semibold text-muted-foreground`}
      >
        <span>Name · Login</span>
        <span>Rolle</span>
        <span>Status</span>
        <span />
      </div>
      {props.users.map((user) => (
        <UserRow
          key={user.id}
          loading={loading || props.loading}
          user={user}
          isSelf={user.id === AuthSingleton.userId}
          onActivate={activateUser}
          onDeactivate={deactivateUser}
          onResetPassword={props.onResetPassword}
          onDelete={deleteUser}
          onEdit={props.onEdit}
        />
      ))}
    </div>
  )
}
