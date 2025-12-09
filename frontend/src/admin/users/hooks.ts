import { useEffect, useState } from 'react'

import { BackendSingleton } from '@/lib/Backend'

import type { User } from './User'
import { UserBackend } from './UserBackend'

const userBackend = new UserBackend(BackendSingleton)

/** Custom hook to fetch all users from backend. */
export function useAllUsers() {
  const [loading, setLoading] = useState(false)
  const [users, setUsers] = useState<User[]>([])

  useEffect(() => {
    async function fetchUsers() {
      setLoading(true)

      try {
        const response = await userBackend.getAllUsers()
        setUsers(response)
      } catch (error) {
        console.error('Failed to fetch users:', error)
      }

      setLoading(false)
    }

    void fetchUsers()
  }, [])

  return { loading, users, setUsers }
}
