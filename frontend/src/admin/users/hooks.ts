import { BackendSingleton } from '@/lib/Backend'
import { useFetch } from '@/lib/useFetch'

import type { User } from './User'
import { UserBackend } from './UserBackend'

const userBackend = new UserBackend(BackendSingleton)

/** Custom hook to fetch all users from backend. */
export function useAllUsers() {
  const { data: users, setData: setUsers, ...rest } = useFetch(
    () => userBackend.getAllUsers(),
    [] as User[],
  )
  return { ...rest, users, setUsers }
}
