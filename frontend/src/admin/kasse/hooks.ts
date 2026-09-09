import { useQuery } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import { type AktiveKassensitzung, KasseBackend } from './KasseBackend'

export const kasseBackend = new KasseBackend(BackendSingleton)

// Query-Keys der Kassentag-Seite. Nach einer Geldtransit-Buchung werden
// Kassenbestand und Bewegungsliste über diese Präfixe invalidiert.
export const KASSENBESTAND_KEY = 'kassenbestand'
export const GELDTRANSIT_LISTE_KEY = 'geldtransit-liste'

export function useAktiveKassensitzung() {
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

export function useKassenbestand(kassensitzungNr: number | null) {
  const { data = null, dataUpdatedAt } = useQuery({
    queryKey: [KASSENBESTAND_KEY, kassensitzungNr],
    queryFn: () => kasseBackend.getKassenbestand(kassensitzungNr ?? 0),
    enabled: kassensitzungNr !== null,
  })
  return { kassenbestand: data, dataUpdatedAt }
}

export function useGeldtransitListe(kassensitzungNr: number | null) {
  const { data = [] } = useQuery({
    queryKey: [GELDTRANSIT_LISTE_KEY, kassensitzungNr],
    queryFn: () => kasseBackend.getGeldtransitListe(kassensitzungNr ?? 0),
    enabled: kassensitzungNr !== null,
  })
  return { buchungen: data }
}
