import { Users as UsersIcon } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

import { EmptyState } from '@/components/common/EmptyState'
import { ItemGroup } from '@/components/ui/item'
import { AuthSingleton } from '@/lib/Auth'
import { getActionErrorMessage } from '@/lib/errorMessages'

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
  const [loading, setLoading] = useState(false)

  const activateUser = async (userId: number) => {
    setLoading(true)
    try {
      await props.backend.activateUser(userId)
      props.onStatusChange(userId, UserStatus.ACTIVE)
    } catch (error) {
      console.error('Error activating user:', error)
      toast.error(
        getActionErrorMessage({ actionLabel: 'Benutzer aktivieren', error }),
      )
    }
    setLoading(false)
  }

  const deactivateUser = async (userId: number) => {
    setLoading(true)
    try {
      await props.backend.deactivateUser(userId)
      props.onStatusChange(userId, UserStatus.INACTIVE)
    } catch (error) {
      console.error('Error deactivating user:', error)
      toast.error(
        getActionErrorMessage({ actionLabel: 'Benutzer deaktivieren', error }),
      )
    }
    setLoading(false)
  }

  const deleteUser = async (userId: number) => {
    setLoading(true)
    try {
      await props.backend.deleteUser(userId)
      props.onDeleted(userId)
    } catch (error) {
      console.error('Error deleting user:', error)
      toast.error(
        getActionErrorMessage({ actionLabel: 'Benutzer löschen', error }),
      )
    }
    setLoading(false)
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
    <>
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
    </>
  )
}
