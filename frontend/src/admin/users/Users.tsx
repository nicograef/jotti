import { useState } from 'react'

import { ItemGroup } from '@/components/ui/item'
import { type User, type UserBackend, UserStatus } from '@/lib/UserBackend'

import { UserItem } from './UserItem'

interface UsersProps {
  loading: boolean
  backend: Pick<UserBackend, 'activateUser' | 'deactivateUser'>
  users: User[]
  onEdit: (userId: number) => void
  onStatusChange: (userId: number, status: UserStatus) => void
}

export function Users(props: UsersProps) {
  const [loading, setLoading] = useState(props.loading)

  const activateUser = async (userId: number) => {
    setLoading(true)
    try {
      await props.backend.activateUser(userId)
      props.onStatusChange(userId, UserStatus.ACTIVE)
    } catch (error) {
      console.error('Error activating user:', error)
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
    }
    setLoading(false)
  }

  return (
    <>
      <ItemGroup className="grid gap-4 lg:grid-cols-2 2xl:grid-cols-3 my-4">
        {props.users.map((user) => (
          <UserItem
            key={user.id}
            loading={loading || props.loading}
            user={user}
            onActivate={activateUser}
            onDeactivate={deactivateUser}
            onEdit={props.onEdit}
          />
        ))}
      </ItemGroup>
    </>
  )
}
