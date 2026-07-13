import { Users as UsersIcon } from 'lucide-react'

import { EmptyState } from '@/components/common/EmptyState'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { AuthSingleton } from '@/lib/Auth'

import { type User } from './User'
import type { UserBackend } from './UserBackend'
import { UserRow } from './UserRow'
import { BENUTZER_SPALTEN } from './UsersSpalten'

interface UsersProps {
  loading: boolean
  backend: Pick<UserBackend, 'activateUser' | 'deactivateUser' | 'deleteUser'>
  users: User[]
  onEdit: (userId: number) => void
  onStatusChange: () => void
  onResetPassword: (userId: number) => Promise<void>
  onDeleted: (userId: number) => void
}

// Benutzer als Tabelle (Design-Handoff 1e): Spalten Name·Login, Rolle, Status
// und Aktionen. Ersetzt das frühere Kachel-Grid.
export function Users(props: UsersProps) {
  const { loading: activateLoading, run: runActivate } = useActionSubmit({
    actionLabel: 'Benutzer aktivieren',
  })
  const { loading: deactivateLoading, run: runDeactivate } = useActionSubmit({
    actionLabel: 'Benutzer deaktivieren',
  })
  const { loading: deleteLoading, run: runDelete } = useActionSubmit({
    actionLabel: 'Benutzer löschen',
  })

  const loading = activateLoading || deactivateLoading || deleteLoading

  const activateUser = async (userId: number) => {
    await runActivate(async () => {
      await props.backend.activateUser(userId)
      props.onStatusChange()
    })
  }

  const deactivateUser = async (userId: number) => {
    await runDeactivate(async () => {
      await props.backend.deactivateUser(userId)
      props.onStatusChange()
    })
  }

  const deleteUser = async (userId: number) => {
    await runDelete(async () => {
      await props.backend.deleteUser(userId)
      props.onDeleted(userId)
    })
  }

  if (props.users.length === 0 && !props.loading) {
    return (
      <EmptyState
        icon={UsersIcon}
        title="Keine Benutzer vorhanden"
        description="Erstelle einen neuen Benutzer, um loszulegen."
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
