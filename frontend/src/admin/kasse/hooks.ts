import { useQuery } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import { KasseBackend, type OffeneKassensitzung } from './KasseBackend'

export const kasseBackend = new KasseBackend(BackendSingleton)

export function useOffeneKassensitzung() {
  const {
    data = null as OffeneKassensitzung | null,
    isPending,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['offene-kassensitzung'],
    queryFn: () => kasseBackend.getOffeneKassensitzung(),
  })
  return { kassensitzung: data, isPending, isError, refetch }
}

export function useKassenbestand(kassensitzungNr: number | null) {
  const { data = null } = useQuery({
    queryKey: ['kassenbestand', kassensitzungNr],
    queryFn: () => kasseBackend.getKassenbestand(kassensitzungNr ?? 0),
    enabled: kassensitzungNr !== null,
  })
  return { kassenbestand: data }
}
