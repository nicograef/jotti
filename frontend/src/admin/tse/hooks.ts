import { useQuery, useQueryClient } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

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

export function useTSEKonfiguration() {
  const queryClient = useQueryClient()
  const { isPending, data, error } = useQuery({
    queryKey: ['tse-konfiguration'],
    queryFn: () => tseBackend.getTSEKonfiguration(),
  })

  const saveTSEKonfiguration = async (config: TSEKonfigurationSpeichern) => {
    await tseBackend.saveTSEKonfiguration(config)
    // Speichern/Leeren ändern auch istKonfiguriert — beide Ansichten neu laden.
    await queryClient.invalidateQueries({ queryKey: ['tse-konfiguration'] })
    await queryClient.invalidateQueries({ queryKey: ['tse-status'] })
  }

  const clearTSEKonfiguration = async () => {
    await tseBackend.clearTSEKonfiguration()
    await queryClient.invalidateQueries({ queryKey: ['tse-konfiguration'] })
    await queryClient.invalidateQueries({ queryKey: ['tse-status'] })
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

export function useTSEEinrichtung() {
  const queryClient = useQueryClient()

  const richteTSEEin = async (
    eingabe: TSEEinrichten,
  ): Promise<TSEEinrichtenErgebnis> => {
    const ergebnis = await tseBackend.richteTSEEin(eingabe)
    // Die Konfiguration ist jetzt gespeichert — abhängige Ansichten neu laden.
    await queryClient.invalidateQueries({ queryKey: ['tse-konfiguration'] })
    await queryClient.invalidateQueries({ queryKey: ['tse-status'] })
    return ergebnis
  }

  const uebernimmTSE = async (
    eingabe: TSEUebernehmen,
  ): Promise<TSEEinrichtenErgebnis> => {
    const ergebnis = await tseBackend.uebernimmTSE(eingabe)
    await queryClient.invalidateQueries({ queryKey: ['tse-konfiguration'] })
    await queryClient.invalidateQueries({ queryKey: ['tse-status'] })
    return ergebnis
  }

  return { richteTSEEin, uebernimmTSE }
}

export function useTSESignaturQueue() {
  const { data, isPending, error } = useQuery({
    queryKey: ['tse-signatur-queue'],
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
    queryKey: ['tse-stoerungen'],
    queryFn: () => tseBackend.getTSEStoerungen(),
  })

  return { stoerungen: data, isPending, error }
}

export function useTSEStatus() {
  const { data, isPending, error } = useQuery({
    queryKey: ['tse-status'],
    queryFn: () => tseBackend.getTSEStatus(),
  })

  return {
    tseStatus: data,
    isPending,
    error,
  }
}
