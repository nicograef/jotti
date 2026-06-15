import { useQuery } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import { KasseBackend } from './KasseBackend'
import type { Kassensitzung } from './Kassensitzung'

export const kasseBackend = new KasseBackend(BackendSingleton)

export function useOffeneKassensitzung() {
  const {
    data = null as Kassensitzung | null,
    isPending,
    refetch,
  } = useQuery({
    queryKey: ['offene-kassensitzung'],
    queryFn: () => kasseBackend.getOffeneKassensitzung(),
  })
  return { kassensitzung: data, isPending, refetch }
}

export function useKassenbestand(kassensitzungNr: number | null) {
  const { data = null } = useQuery({
    queryKey: ['kassenbestand', kassensitzungNr],
    queryFn: () => kasseBackend.getKassenbestand(kassensitzungNr ?? 0),
    enabled: kassensitzungNr !== null,
  })
  return { kassenbestand: data }
}
