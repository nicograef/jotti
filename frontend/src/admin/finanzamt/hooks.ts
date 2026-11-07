import {
  useQuery,
  useQueryClient,
  type UseQueryResult,
} from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import {
  type Betreiber,
  BetreiberBackend,
  type BetreiberEingabe,
  type Kassenidentitaet,
} from './BetreiberBackend'

const betreiberBackend = new BetreiberBackend(BackendSingleton)

const KASSENIDENTITAET_KEY = 'kassenidentitaet'
const BETREIBER_KEY = 'betreiber'

export function useKassenidentitaet(): {
  kassenidentitaet: Kassenidentitaet | undefined
} {
  const { data } = useQuery({
    queryKey: [KASSENIDENTITAET_KEY],
    queryFn: () => betreiberBackend.getKassenidentitaet(),
  })
  return { kassenidentitaet: data }
}

interface BetreiberResult {
  betreiber: Betreiber | undefined
  isPending: boolean
  isError: boolean
  refetchBetreiber: UseQueryResult<Betreiber>['refetch']
  saveBetreiber: (b: BetreiberEingabe) => Promise<void>
  setElsterMeldung: () => Promise<void>
  nimmElsterMeldungZurueck: () => Promise<void>
}

export function useBetreiber(): BetreiberResult {
  const queryClient = useQueryClient()
  const { isPending, isError, data, refetch } = useQuery({
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
    refetchBetreiber: refetch,
    saveBetreiber,
    setElsterMeldung,
    nimmElsterMeldungZurueck,
  }
}
