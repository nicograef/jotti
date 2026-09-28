import { useQuery, type UseQueryResult } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import type { DirektverkaufHistorieEintrag } from './Direktverkauf'
import { DirektverkaufBackend } from './DirektverkaufBackend'

export const direktverkaufBackend = new DirektverkaufBackend(BackendSingleton)

interface DirektverkaufHistorieResult {
  historie: DirektverkaufHistorieEintrag[]
  isPending: boolean
  isError: boolean
  refetch: UseQueryResult<DirektverkaufHistorieEintrag[]>['refetch']
}

export function useDirektverkaufHistorie(): DirektverkaufHistorieResult {
  const {
    data: historie = [],
    isPending,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['direktverkauf-historie'],
    queryFn: () => direktverkaufBackend.getDirektverkaufHistorie(),
  })
  return { historie, isPending, isError, refetch }
}
