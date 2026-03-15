import { BackendSingleton } from '@/lib/Backend'
import { useFetch } from '@/lib/useFetch'

import type { Ausgabe } from './Ausgabe'
import type { Bestellung } from './Bestellung'
import type { Stornierung } from './Stornierung'
import type {
  AktiverTischMitFavorit,
  EigeneUebersicht,
  Tisch,
  TischState,
} from './Tisch'
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
    [] as (Bestellung | Zahlung | Stornierung | Ausgabe)[],
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
      ausstehendePositionen: [],
      gesamtZahlungenCents: 0,
    } as TischState,
    [tischId],
  )
  return { ...rest, state }
}

export function useAktiveTischeMitFavoriten() {
  const { data: tische, ...rest } = useFetch(
    () => tischBackend.getAktiveTischeMitFavoriten(),
    [] as AktiverTischMitFavorit[],
  )
  return { ...rest, tische }
}

export function useMeineTischeState() {
  const { data: tische, ...rest } = useFetch(
    () => tischBackend.getMeineTischeState(),
    [] as TischState[],
  )
  return { ...rest, tische }
}

export function useEigeneUebersicht() {
  const { data: uebersicht, ...rest } = useFetch(
    () => tischBackend.getEigeneUebersicht(),
    {
      anzahlBestellungen: 0,
      bestellungenCents: 0,
      anzahlZahlungen: 0,
      zahlungenCents: 0,
    } as EigeneUebersicht,
  )
  return { ...rest, uebersicht }
}
