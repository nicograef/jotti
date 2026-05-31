import { useCallback } from 'react'

import { BackendSingleton } from '@/lib/Backend'
import { useFetch } from '@/lib/useFetch'

import type { Ausgabe } from './Ausgabe'
import type { Bestellung } from './Bestellung'
import type { Stornierung } from './Stornierung'
import type { AktiverTischMitFavorit, Tisch, TischSession } from './Tisch'
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
  const fetcher = useCallback(
    () => tischBackend.getTischHistorie(tischId),
    [tischId],
  )
  const { data: historie, ...rest } = useFetch(
    fetcher,
    [] as (Bestellung | Zahlung | Stornierung | Ausgabe)[],
  )
  return { ...rest, historie }
}

export function useTischState(tischId: number) {
  const fetcher = useCallback(
    () => tischBackend.getTischState(tischId),
    [tischId],
  )
  const { data: state, ...rest } = useFetch(fetcher, {
    tischId: 0,
    tischName: '',
    saldoCents: 0,
    unbezahltePositionen: [],
    ausstehendePositionen: [],
    gesamtZahlungenCents: 0,
  })
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
    [] as TischSession[],
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
    },
  )
  return { ...rest, uebersicht }
}
