import { useQuery } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'
import type { QueryResult } from '@/lib/queryClient'

import type { DirektverkaufHistorieEintrag } from './Direktverkauf'
import { DirektverkaufBackend } from './DirektverkaufBackend'

export const direktverkaufBackend = new DirektverkaufBackend(BackendSingleton)

export function useDirektverkaufHistorie(): QueryResult<
  DirektverkaufHistorieEintrag[]
> & { historie: DirektverkaufHistorieEintrag[] } {
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
