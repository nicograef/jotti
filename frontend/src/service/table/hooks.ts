import { BackendSingleton } from '@/lib/Backend'
import { useFetch } from '@/lib/useFetch'

import type { Bestellung, Position } from './Bestellung'
import type { Lieferung } from './Lieferung'
import type { Stornierung } from './Stornierung'
import type { Tisch } from './Tisch'
import { TischBackend } from './TischBackend'
import type { Zahlung } from './Zahlung'

const tischBackend = new TischBackend(BackendSingleton)

/** Custom hook to fetch a single table from backend. */
export function useTisch(id: number) {
  const { data: tisch, ...rest } = useFetch(
    () => tischBackend.getTisch(id),
    null as Tisch | null,
    [id],
  )
  return { ...rest, tisch }
}

/** Custom hook to fetch active tables from backend. */
export function useAktiveTische() {
  const { data: tische, ...rest } = useFetch(
    () => tischBackend.getAktiveTische(),
    [] as Tisch[],
  )
  return { ...rest, tische }
}

/** Custom hook to fetch the history for a specific table from backend. */
export function useTischHistorie(tischId: number) {
  const { data: historie, ...rest } = useFetch(
    () => tischBackend.getTischHistorie(tischId),
    [] as (Bestellung | Zahlung | Stornierung | Lieferung)[],
    [tischId],
  )
  return { ...rest, historie }
}

export function useTischSaldo(tischId: number) {
  const { data: saldoCents, ...rest } = useFetch(
    () => tischBackend.getTischSaldo(tischId),
    0,
    [tischId],
  )
  return { ...rest, saldoCents }
}

export function useTischUnbezahlt(tischId: number) {
  const { data: positionen, ...rest } = useFetch(
    () => tischBackend.getTischUnbezahlt(tischId),
    [] as Position[],
    [tischId],
  )
  return { ...rest, positionen }
}

export function useTischUngeliefert(tischId: number) {
  const { data: positionen, ...rest } = useFetch(
    () => tischBackend.getTischUngeliefert(tischId),
    [] as Position[],
    [tischId],
  )
  return { ...rest, positionen }
}
