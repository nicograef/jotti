import { useQuery, useQueryClient } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'
import { STAMMDATEN_AKTUALITAET_MS } from '@/lib/queryClient'

import {
  TSEBackend,
  type TSEEinrichten,
  type TSEEinrichtenErgebnis,
  type TSEKonfigurationSpeichern,
  type TSESetupBefund,
  type TSESetupZugangsdaten,
  type TSEUebernehmen,
  type TSEVerbindungStatus,
} from './TSEBackend'

const tseBackend = new TSEBackend(BackendSingleton)

// Query-Keys der TSE-Ansichten. Nach Speichern/Leeren/Einrichten werden
// Konfiguration und Status über diese Keys invalidiert.
export const TSE_KONFIGURATION_KEY = 'tse-konfiguration'
export const TSE_STATUS_KEY = 'tse-status'
export const TSE_SIGNATUR_QUEUE_KEY = 'tse-signatur-queue'
export const TSE_STOERUNGEN_KEY = 'tse-stoerungen'

// Ab rund einer Minute Signatur-Rückstand gilt der TSE-Signatur-Rückstand als
// kritisch (deckt sich mit der Nachsigniert-Schwelle im Backend). Dashboard und
// Sidebar teilen sich diese Schwelle.
export const RUECKSTAND_WARN_SEKUNDEN = 60

// `isLoadingError` statt `error`: Nur ein gescheitertes Erstladen (kein
// brauchbarer Cache-Stand) rechtfertigt einen Fehlerzustand statt des Formulars.
// Scheitert ein Hintergrund-Refetch — etwa der Fokus-Refetch, wenn der Admin von
// fiskaly zurückkehrt —, bleiben Formular und Eingaben stehen; die Meldung trägt
// der zentrale Fehler-Toast aus queryClient.ts.
export function useTSEKonfiguration() {
  const queryClient = useQueryClient()
  const { isPending, data, isLoadingError } = useQuery({
    queryKey: [TSE_KONFIGURATION_KEY],
    queryFn: () => tseBackend.getTSEKonfiguration(),
    staleTime: STAMMDATEN_AKTUALITAET_MS,
  })

  const saveTSEKonfiguration = async (config: TSEKonfigurationSpeichern) => {
    await tseBackend.saveTSEKonfiguration(config)
    // Speichern/Leeren ändern auch istKonfiguriert — beide Ansichten neu laden.
    await queryClient.invalidateQueries({ queryKey: [TSE_KONFIGURATION_KEY] })
    await queryClient.invalidateQueries({ queryKey: [TSE_STATUS_KEY] })
  }

  const clearTSEKonfiguration = async () => {
    await tseBackend.clearTSEKonfiguration()
    await queryClient.invalidateQueries({ queryKey: [TSE_KONFIGURATION_KEY] })
    await queryClient.invalidateQueries({ queryKey: [TSE_STATUS_KEY] })
  }

  const testTSEVerbindung = async (): Promise<TSEVerbindungStatus> => {
    return tseBackend.testTSEVerbindung()
  }

  return {
    tseKonfiguration: data,
    isPending,
    isLoadingError,
    saveTSEKonfiguration,
    clearTSEKonfiguration,
    testTSEVerbindung,
  }
}

export function checkTSESetup(
  zugangsdaten: TSESetupZugangsdaten,
): Promise<TSESetupBefund> {
  return tseBackend.checkTSESetup(zugangsdaten)
}

export function useTSEEinrichtung() {
  const queryClient = useQueryClient()

  const richteTSEEin = async (
    eingabe: TSEEinrichten,
  ): Promise<TSEEinrichtenErgebnis> => {
    const ergebnis = await tseBackend.richteTSEEin(eingabe)
    // Die Konfiguration ist jetzt gespeichert — abhängige Ansichten neu laden.
    await queryClient.invalidateQueries({ queryKey: [TSE_KONFIGURATION_KEY] })
    await queryClient.invalidateQueries({ queryKey: [TSE_STATUS_KEY] })
    return ergebnis
  }

  const uebernimmTSE = async (
    eingabe: TSEUebernehmen,
  ): Promise<TSEEinrichtenErgebnis> => {
    const ergebnis = await tseBackend.uebernimmTSE(eingabe)
    await queryClient.invalidateQueries({ queryKey: [TSE_KONFIGURATION_KEY] })
    await queryClient.invalidateQueries({ queryKey: [TSE_STATUS_KEY] })
    return ergebnis
  }

  return { richteTSEEin, uebernimmTSE }
}

export function useTSESignaturQueue() {
  const { data, isPending, error } = useQuery({
    queryKey: [TSE_SIGNATUR_QUEUE_KEY],
    queryFn: () => tseBackend.getTSESignaturQueue(),
  })

  return { queue: data, isPending, error }
}

export function useTSEStoerungen() {
  const {
    isPending,
    data = [],
    error,
  } = useQuery({
    queryKey: [TSE_STOERUNGEN_KEY],
    queryFn: () => tseBackend.getTSEStoerungen(),
  })

  return { stoerungen: data, isPending, error }
}

export function useTSEStatus() {
  const { data, isPending, error } = useQuery({
    queryKey: [TSE_STATUS_KEY],
    queryFn: () => tseBackend.getTSEStatus(),
  })

  return {
    tseStatus: data,
    isPending,
    error,
  }
}
