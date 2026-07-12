import { Users as UsersIcon } from 'lucide-react'

import { EmptyState } from '@/components/common/EmptyState'
import { ItemGroup } from '@/components/ui/item'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { AuthSingleton } from '@/lib/Auth'

import { type User, UserStatus } from './User'
import type { UserBackend } from './UserBackend'
import { UserItem } from './UserItem'

interface UsersProps {
  loading: boolean
  backend: Pick<UserBackend, 'activateUser' | 'deactivateUser' | 'deleteUser'>
  users: User[]
  onEdit: (userId: number) => void
  onStatusChange: (userId: number, status: UserStatus) => void
  onDeleted: (userId: number) => void
}

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
      props.onStatusChange(userId, UserStatus.ACTIVE)
    })
  }

  const deactivateUser = async (userId: number) => {
    await runDeactivate(async () => {
      await props.backend.deactivateUser(userId)
      props.onStatusChange(userId, UserStatus.INACTIVE)
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
    <ItemGroup className="grid gap-4 lg:grid-cols-2 2xl:grid-cols-3 my-4">
      {props.users.map((user) => (
        <UserItem
          key={user.id}
          loading={loading || props.loading}
          user={user}
          isSelf={user.id === AuthSingleton.userId}
          onActivate={activateUser}
          onDeactivate={deactivateUser}
          onDelete={deleteUser}
          onEdit={props.onEdit}
        />
      ))}
    </ItemGroup>
  )
}
