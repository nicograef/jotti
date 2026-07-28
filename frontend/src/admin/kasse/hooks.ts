import { useQuery } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import { KasseBackend, type OffeneKassensitzung } from './KasseBackend'

export const kasseBackend = new KasseBackend(BackendSingleton)

// Query-Keys der Kassentag-Seite. Der GeldtransitDialog invalidiert beide
// Präfixe, wenn er sich schließt — nach einer Buchung ebenso wie nach einem
// Abbruch, denn auch ein gescheiterter Versuch kann serverseitig gebucht haben.
export const KASSENBESTAND_KEY = 'kassenbestand'
export const GELDTRANSIT_LISTE_KEY = 'geldtransit-liste'

// `isLoadingError` statt `isError`: Nur ein gescheitertes Erstladen (kein
// brauchbarer Cache-Stand) rechtfertigt einen Fehlerzustand statt der Daten.
// Scheitert ein Hintergrund-Refetch, bleibt die Kassentag-Ansicht samt
// eingetipptem Ist-Bestand stehen; die Meldung trägt der zentrale Fehler-Toast
// aus queryClient.ts.
export function useOffeneKassensitzung() {
  const {
    data = null as OffeneKassensitzung | null,
    isPending,
    isLoadingError,
    refetch,
  } = useQuery({
    queryKey: ['offene-kassensitzung'],
    queryFn: () => kasseBackend.getOffeneKassensitzung(),
  })
  return { kassensitzung: data, isPending, isLoadingError, refetch }
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
