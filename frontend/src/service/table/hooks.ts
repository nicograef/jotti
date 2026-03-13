import { BackendSingleton } from '@/lib/Backend'
import { useFetch } from '@/lib/useFetch'

import type { Bestellung } from './Bestellung'
import type { Lieferung } from './Lieferung'
import type { Stornierung } from './Stornierung'
import type { Tisch } from './Tisch'
import type { TischState } from './Tisch'
import { TischBackend } from './TischBackend'
import type { Zahlung } from './Zahlung'

const tischBackend = new TischBackend(BackendSingleton)

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

export function useTischState(tischId: number) {
  const { data: state, ...rest } = useFetch(
    () => tischBackend.getTischState(tischId),
    {
      tischId: 0,
      tischName: '',
      saldoCents: 0,
      unbezahltePositionen: [],
      ungeliefertePositionen: [],
      gesamtZahlungenCents: 0,
    } as TischState,
    [tischId],
  )
  return { ...rest, state }
}
