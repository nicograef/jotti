import { useQuery, type UseQueryResult } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import { type AktiveKassensitzung, KasseBackend } from './KasseBackend'
import type { GeldtransitBuchung, Kassenbestand } from './Kassensitzung'

export const kasseBackend = new KasseBackend(BackendSingleton)

// Präfixe für die Invalidierung nach einer Geldtransit-Buchung.
export const KASSENBESTAND_KEY = 'kassenbestand'
export const GELDTRANSIT_LISTE_KEY = 'geldtransit-liste'

interface AktiveKassensitzungResult {
  kassensitzung: AktiveKassensitzung | null
  isPending: boolean
  isError: boolean
  refetch: UseQueryResult<AktiveKassensitzung | null>['refetch']
}

export function useAktiveKassensitzung(): AktiveKassensitzungResult {
  const {
    data = null as AktiveKassensitzung | null,
    isPending,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['aktive-kassensitzung'],
    queryFn: () => kasseBackend.getAktiveKassensitzung(),
  })
  return { kassensitzung: data, isPending, isError, refetch }
}

export function useKassenbestand(kassensitzungNr: number | null): {
  kassenbestand: Kassenbestand | null
  dataUpdatedAt: number
} {
  const { data = null, dataUpdatedAt } = useQuery({
    queryKey: [KASSENBESTAND_KEY, kassensitzungNr],
    queryFn: () => kasseBackend.getKassenbestand(kassensitzungNr ?? 0),
    enabled: kassensitzungNr !== null,
  })
  return { kassenbestand: data, dataUpdatedAt }
}

export function useGeldtransitListe(kassensitzungNr: number | null): {
  buchungen: GeldtransitBuchung[]
} {
  const { data = [] } = useQuery({
    queryKey: [GELDTRANSIT_LISTE_KEY, kassensitzungNr],
    queryFn: () => kasseBackend.getGeldtransitListe(kassensitzungNr ?? 0),
    enabled: kassensitzungNr !== null,
  })
  return { buchungen: data }
}
