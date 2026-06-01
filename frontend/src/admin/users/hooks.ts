import { useQuery } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import type { User } from './User'
import { UserBackend } from './UserBackend'

const userBackend = new UserBackend(BackendSingleton)

export const ALLE_USERS_KEY = 'alle-users'

export function useAllUsers() {
  const { data: users = [] as User[], isPending } = useQuery({
    queryKey: [ALLE_USERS_KEY],
    queryFn: () => userBackend.getAllUsers(),
  })
  return { users, loading: isPending }
}
