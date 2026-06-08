import { useQuery } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import { DirektverkaufBackend } from './DirektverkaufBackend'

const direktverkaufBackend = new DirektverkaufBackend(BackendSingleton)

export function useDirektverkaufHistorie() {
  const {
    data: historie = [],
    isPending,
    refetch,
  } = useQuery({
    queryKey: ['direktverkauf-historie'],
    queryFn: () => direktverkaufBackend.getDirektverkaufHistorie(),
  })
  return { historie, isPending, refetch }
}
