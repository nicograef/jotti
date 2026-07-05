import { useQuery, useQueryClient } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import { type Betreiber, BetreiberBackend } from './BetreiberBackend'

const betreiberBackend = new BetreiberBackend(BackendSingleton)

export function useKassenidentitaet() {
  return useQuery({
    queryKey: ['kassenidentitaet'],
    queryFn: () => betreiberBackend.getKassenidentitaet(),
  })
}

export function useBetreiber() {
  const queryClient = useQueryClient()
  const { isPending, data, error } = useQuery({
    queryKey: ['betreiber'],
    queryFn: () => betreiberBackend.getBetreiber(),
  })

  const saveBetreiber = async (b: Betreiber) => {
    await betreiberBackend.saveBetreiber(b)
    await queryClient.invalidateQueries({ queryKey: ['betreiber'] })
  }

  return { betreiber: data, isPending, error, saveBetreiber }
}
