import { useQuery, useQueryClient } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import {
  TSEBackend,
  type TSEEinrichten,
  type TSEEinrichtenErgebnis,
  type TSEKonfiguration,
  type TSEKonfigurationSpeichern,
  type TSESetupBefund,
  type TSESetupZugangsdaten,
  type TSESignaturQueue,
  type TSEStatus,
  type TSEStoerung,
  type TSEUebernehmen,
  type TSEVerbindungStatus,
} from './TSEBackend'

const tseBackend = new TSEBackend(BackendSingleton)

export const TSE_KONFIGURATION_KEY = 'tse-konfiguration'
export const TSE_STATUS_KEY = 'tse-status'
export const TSE_SIGNATUR_QUEUE_KEY = 'tse-signatur-queue'
export const TSE_STOERUNGEN_KEY = 'tse-stoerungen'

// Deckt sich mit der Nachsigniert-Schwelle im Backend; Dashboard und Sidebar
// teilen diese Schwelle.
export const RUECKSTAND_WARN_SEKUNDEN = 60

interface TSEKonfigurationResult {
  tseKonfiguration: TSEKonfiguration | undefined
  isPending: boolean
  error: Error | null
  saveTSEKonfiguration: (config: TSEKonfigurationSpeichern) => Promise<void>
  clearTSEKonfiguration: () => Promise<void>
  testTSEVerbindung: () => Promise<TSEVerbindungStatus>
}

export function useTSEKonfiguration(): TSEKonfigurationResult {
  const queryClient = useQueryClient()
  const { isPending, data, error } = useQuery({
    queryKey: [TSE_KONFIGURATION_KEY],
    queryFn: () => tseBackend.getTSEKonfiguration(),
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
    error,
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

interface TSEEinrichtungResult {
  richteTSEEin: (eingabe: TSEEinrichten) => Promise<TSEEinrichtenErgebnis>
  uebernimmTSE: (eingabe: TSEUebernehmen) => Promise<TSEEinrichtenErgebnis>
}

export function useTSEEinrichtung(): TSEEinrichtungResult {
  const queryClient = useQueryClient()

  const richteTSEEin = async (
    eingabe: TSEEinrichten,
  ): Promise<TSEEinrichtenErgebnis> => {
    const ergebnis = await tseBackend.richteTSEEin(eingabe)
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

export function useTSESignaturQueue(): {
  queue: TSESignaturQueue | undefined
  isPending: boolean
  error: Error | null
} {
  const { data, isPending, error } = useQuery({
    queryKey: [TSE_SIGNATUR_QUEUE_KEY],
    queryFn: () => tseBackend.getTSESignaturQueue(),
  })

  return { queue: data, isPending, error }
}

export function useTSEStoerungen(): {
  stoerungen: TSEStoerung[]
  isPending: boolean
  error: Error | null
} {
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

export function useTSEStatus(): {
  tseStatus: TSEStatus | undefined
  isPending: boolean
  error: Error | null
} {
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
