import { useQuery, useQueryClient } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'
import { STAMMDATEN_AKTUALITAET_MS } from '@/lib/queryClient'

import { BetreiberBackend, type BetreiberEingabe } from './BetreiberBackend'

const betreiberBackend = new BetreiberBackend(BackendSingleton)

export const KASSENIDENTITAET_KEY = 'kassenidentitaet'
export const BETREIBER_KEY = 'betreiber'

export function useKassenidentitaet() {
  const { data, isPending, error } = useQuery({
    queryKey: [KASSENIDENTITAET_KEY],
    queryFn: () => betreiberBackend.getKassenidentitaet(),
    staleTime: STAMMDATEN_AKTUALITAET_MS,
  })
  return { kassenidentitaet: data, isPending, error }
}

// `isLoadingError` statt `isError`: Nur ein gescheitertes Erstladen (kein
// brauchbarer Cache-Stand) rechtfertigt einen Fehlerzustand statt der Daten.
// Scheitert ein Hintergrund-Refetch, bleibt die Einrichtungs-Checkliste stehen;
// die Meldung trägt der zentrale Fehler-Toast aus queryClient.ts.
export function useBetreiber() {
  const queryClient = useQueryClient()
  const { isPending, isLoadingError, data, error, refetch } = useQuery({
    queryKey: [BETREIBER_KEY],
    queryFn: () => betreiberBackend.getBetreiber(),
    staleTime: STAMMDATEN_AKTUALITAET_MS,
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
    isLoadingError,
    error,
    refetchBetreiber: refetch,
    saveBetreiber,
    setElsterMeldung,
    nimmElsterMeldungZurueck,
  }
}
