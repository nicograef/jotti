import { useQuery, useQueryClient } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import { BetreiberBackend, type BetreiberEingabe } from './BetreiberBackend'

const betreiberBackend = new BetreiberBackend(BackendSingleton)

export function useKassenidentitaet() {
  return useQuery({
    queryKey: ['kassenidentitaet'],
    queryFn: () => betreiberBackend.getKassenidentitaet(),
  })
}

export function useBetreiber() {
  const queryClient = useQueryClient()
  const { isPending, isError, data, error, refetch } = useQuery({
    queryKey: ['betreiber'],
    queryFn: () => betreiberBackend.getBetreiber(),
  })

  const saveBetreiber = async (b: BetreiberEingabe) => {
    await betreiberBackend.saveBetreiber(b)
    await queryClient.invalidateQueries({ queryKey: ['betreiber'] })
  }

  const setElsterMeldung = async () => {
    await betreiberBackend.setElsterMeldung()
    await queryClient.invalidateQueries({ queryKey: ['betreiber'] })
  }

  const nimmElsterMeldungZurueck = async () => {
    await betreiberBackend.nimmElsterMeldungZurueck()
    await queryClient.invalidateQueries({ queryKey: ['betreiber'] })
  }

  return {
    betreiber: data,
    isPending,
    isError,
    error,
    refetchBetreiber: refetch,
    saveBetreiber,
    setElsterMeldung,
    nimmElsterMeldungZurueck,
  }
}
