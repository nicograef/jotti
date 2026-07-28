import { useQuery } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import type { EigeneUebersicht, TischSession } from './Tisch'
import { TischBackend } from './TischBackend'

const tischBackend = new TischBackend(BackendSingleton)

// Die Hooks mit Fehlerzweig melden `isLoadingError` statt `isError`: Nur ein
// gescheitertes Erstladen (kein brauchbarer Cache-Stand) rechtfertigt einen
// Fehlerzustand statt der Daten. Scheitert ein Hintergrund-Refetch, bleiben die
// zwischengespeicherten Daten stehen; die Meldung trägt der zentrale
// Fehler-Toast aus queryClient.ts.

// Dateilokal, anders als TISCH_STATE_KEY und TISCH_HISTORIE_KEY: Diese Query
// wird nirgends invalidiert und braucht es auch nicht. Ihr einziger Konsument
// ist die Ziel-Tisch-Auswahl im HistorieUmbuchungDrawer, der nur geöffnet
// gemountet ist — ohne Aktualitätsschwelle lädt jedes Öffnen ohnehin neu.
const AKTIVE_TISCHE_KEY = 'aktive-tische'
// Kein Stammdatum: Die Nutzlast trägt `saldoCents` — den offenen Tisch-Saldo der
// laufenden Kassensitzung. Damit gilt die Regel wörtlich, statt vom aktuellen
// Verwendungszweck des Feldes abzuhängen; niemand kann später versehentlich
// einen veralteten Saldo anzeigen (Präzedenz: `useAllTische` in
// admin/tables/hooks.ts). Der Traffic-Effekt ist vernachlässigbar: Einziger
// Konsument ist die Ziel-Tisch-Auswahl beim Umbuchen.
export function useAktiveTische() {
  const {
    data: tische = [],
    isPending,
    isLoadingError,
    refetch,
  } = useQuery({
    queryKey: [AKTIVE_TISCHE_KEY],
    queryFn: () => tischBackend.getAktiveTische(),
  })
  return { tische, isPending, isLoadingError, refetch }
}

export const TISCH_HISTORIE_KEY = 'tisch-historie'
export function useTischHistorie(tischId: number) {
  const {
    data: historie = [],
    isPending,
    isLoadingError,
  } = useQuery({
    queryKey: [TISCH_HISTORIE_KEY, tischId],
    queryFn: () => tischBackend.getTischHistorie(tischId),
  })
  return { historie, isPending, isLoadingError }
}

const DEFAULT_TISCH_STATE: TischSession = {
  tischId: 0,
  tischName: '',
  saldoCents: 0,
  unbezahltePositionen: [],
  fuerMichErledigt: true,
}

export const TISCH_STATE_KEY = 'tisch-state'
export function useTischState(tischId: number) {
  const {
    data: state = DEFAULT_TISCH_STATE,
    isPending,
    isLoadingError,
  } = useQuery({
    queryKey: [TISCH_STATE_KEY, tischId],
    queryFn: () => tischBackend.getTischState(tischId),
  })
  return { state, isPending, isLoadingError }
}

export const AKTIVE_TISCHE_MIT_FAVORITEN_KEY = 'aktive-tische-mit-favoriten'
export function useAktiveTischeMitFavoriten() {
  const {
    data: tische = [],
    isLoadingError,
    refetch,
  } = useQuery({
    queryKey: [AKTIVE_TISCHE_MIT_FAVORITEN_KEY],
    queryFn: () => tischBackend.getAktiveTischeMitFavoriten(),
  })
  return { tische, isLoadingError, refetch }
}

export const MEINE_TISCHE_STATE_KEY = 'meine-tische-state'
export function useMeineTischeState() {
  const {
    data: tische = [],
    isPending,
    isLoadingError,
    refetch,
  } = useQuery({
    queryKey: [MEINE_TISCHE_STATE_KEY],
    queryFn: () => tischBackend.getMeineTischeState(),
  })
  return { tische, isPending, isLoadingError, refetch }
}

const DEFAULT_EIGENE_UEBERSICHT: EigeneUebersicht = {
  anzahlBestellungen: 0,
  bestellungenCents: 0,
  anzahlZahlungen: 0,
  zahlungenCents: 0,
  anzahlRuecknahmen: 0,
  ruecknahmenCents: 0,
  abzugebenCents: 0,
}

export const EIGENE_UEBERSICHT_KEY = 'eigene-uebersicht'
export function useEigeneUebersicht() {
  const {
    data: uebersicht = DEFAULT_EIGENE_UEBERSICHT,
    isPending,
    isLoadingError,
    refetch,
  } = useQuery({
    queryKey: [EIGENE_UEBERSICHT_KEY],
    queryFn: () => tischBackend.getEigeneUebersicht(),
  })
  return { uebersicht, isPending, isLoadingError, refetch }
}
