import { useQuery } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import type { EigeneUebersicht, TischSession } from './Tisch'
import { TischBackend } from './TischBackend'

const tischBackend = new TischBackend(BackendSingleton)

export function useAktiveTische() {
  const {
    data: tische = [],
    isPending,
    refetch,
  } = useQuery({
    queryKey: ['aktive-tische'],
    queryFn: () => tischBackend.getAktiveTische(),
  })
  return { tische, isPending, refetch }
}

export function useTischHistorie(tischId: number) {
  const {
    data: historie = [],
    isPending,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['tisch-historie', tischId],
    queryFn: () => tischBackend.getTischHistorie(tischId),
  })
  return { historie, isPending, isError, refetch }
}

const DEFAULT_TISCH_STATE: TischSession = {
  tischId: 0,
  tischName: '',
  saldoCents: 0,
  unbezahltePositionen: [],
  gesamtZahlungenCents: 0,
  fuerMichErledigt: true,
}

export function useTischState(tischId: number) {
  const {
    data: state = DEFAULT_TISCH_STATE,
    isPending,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['tisch-state', tischId],
    queryFn: () => tischBackend.getTischState(tischId),
  })
  return { state, isPending, isError, refetch }
}

export const AKTIVE_TISCHE_MIT_FAVORITEN_KEY = 'aktive-tische-mit-favoriten'
export function useAktiveTischeMitFavoriten() {
  const { data: tische = [] } = useQuery({
    queryKey: [AKTIVE_TISCHE_MIT_FAVORITEN_KEY],
    queryFn: () => tischBackend.getAktiveTischeMitFavoriten(),
  })
  return { tische }
}

export const MEINE_TISCHE_STATE_KEY = 'meine-tische-state'
export function useMeineTischeState() {
  const { data: tische = [], isPending } = useQuery({
    queryKey: [MEINE_TISCHE_STATE_KEY],
    queryFn: () => tischBackend.getMeineTischeState(),
  })
  return { tische, isPending }
}

const DEFAULT_EIGENE_UEBERSICHT: EigeneUebersicht = {
  anzahlBestellungen: 0,
  bestellungenCents: 0,
  anzahlZahlungen: 0,
  zahlungenCents: 0,
}

export function useEigeneUebersicht() {
  const { data: uebersicht = DEFAULT_EIGENE_UEBERSICHT, isPending } = useQuery({
    queryKey: ['eigene-uebersicht'],
    queryFn: () => tischBackend.getEigeneUebersicht(),
  })
  return { uebersicht, isPending }
}
