import { useQuery, useQueryClient } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import { BetreiberBackend, type BetreiberEingabe } from './BetreiberBackend'

const betreiberBackend = new BetreiberBackend(BackendSingleton)

export const KASSENIDENTITAET_KEY = 'kassenidentitaet'
export const BETREIBER_KEY = 'betreiber'

export function useKassenidentitaet() {
  const { data, isPending, error } = useQuery({
    queryKey: [KASSENIDENTITAET_KEY],
    queryFn: () => betreiberBackend.getKassenidentitaet(),
  })
  return { kassenidentitaet: data, isPending, error }
}

export function useBetreiber() {
  const queryClient = useQueryClient()
  const { isPending, isError, data, error, refetch } = useQuery({
    queryKey: [BETREIBER_KEY],
    queryFn: () => betreiberBackend.getBetreiber(),
  })

  const saveBetreiber = async (b: BetreiberEingabe) => {
    await betreiberBackend.saveBetreiber(b)
    await queryClient.invalidateQueries({ queryKey: [BETREIBER_KEY] })
  }

  const setElsterMeldung = async () => {
    await betreiberBackend.setElsterMeldung()
    await queryClient.invalidateQueries({ queryKey: [BETREIBER_KEY] })
  }

  const nimmElsterMeldungZurueck = async () => {
    await betreiberBackend.nimmElsterMeldungZurueck()
    await queryClient.invalidateQueries({ queryKey: [BETREIBER_KEY] })
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
