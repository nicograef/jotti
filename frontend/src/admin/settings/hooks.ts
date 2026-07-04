import { useQuery, useQueryClient } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'
import {
  DruckstationBackend,
  type DruckstationConfig,
} from '@/lib/DruckstationBackend'
import {
  type Betreiber,
  EinstellungenBackend,
  type TSEEinrichten,
  type TSEEinrichtenErgebnis,
  type TSEKonfigurationSpeichern,
  type TSESetupBefund,
  type TSESetupZugangsdaten,
  type TSEUebernehmen,
  type TSEVerbindungStatus,
} from '@/lib/EinstellungenBackend'

const druckstationBackend = new DruckstationBackend(BackendSingleton)
const einstellungenBackend = new EinstellungenBackend(BackendSingleton)

export function useDruckstationen() {
  const queryClient = useQueryClient()
  const {
    isPending,
    data = [],
    error,
  } = useQuery({
    queryKey: ['druckstationen'],
    queryFn: () => druckstationBackend.getDruckstationen(),
  })

  const updateDruckstation = async (newConfig: DruckstationConfig) => {
    await druckstationBackend.updateDruckstation(newConfig)
    await queryClient.invalidateQueries({ queryKey: ['druckstationen'] })
  }

  return { druckstationen: data, isPending, error, updateDruckstation }
}

export function useFehlgeschlageneDruckauftraege() {
  const queryClient = useQueryClient()
  const {
    isPending,
    data = [],
    error,
  } = useQuery({
    queryKey: ['fehlgeschlagene-druckauftraege'],
    queryFn: () => druckstationBackend.getFehlgeschlageneDruckauftraege(),
  })

  const erneutVersuchen = async (id: number) => {
    await druckstationBackend.druckauftragErneutVersuchen(id)
    await queryClient.invalidateQueries({
      queryKey: ['fehlgeschlagene-druckauftraege'],
    })
  }

  const verwerfen = async (id: number) => {
    await druckstationBackend.druckauftragVerwerfen(id)
    await queryClient.invalidateQueries({
      queryKey: ['fehlgeschlagene-druckauftraege'],
    })
  }

  return { druckauftraege: data, isPending, error, erneutVersuchen, verwerfen }
}

export function useKassenidentitaet() {
  return useQuery({
    queryKey: ['kassenidentitaet'],
    queryFn: () => einstellungenBackend.getKassenidentitaet(),
  })
}

export function useBetreiber() {
  const queryClient = useQueryClient()
  const { isPending, data, error } = useQuery({
    queryKey: ['betreiber'],
    queryFn: () => einstellungenBackend.getBetreiber(),
  })

  const saveBetreiber = async (b: Betreiber) => {
    await einstellungenBackend.saveBetreiber(b)
    await queryClient.invalidateQueries({ queryKey: ['betreiber'] })
  }

  return { betreiber: data, isPending, error, saveBetreiber }
}

export function useTSEKonfiguration() {
  const queryClient = useQueryClient()
  const { isPending, data, error } = useQuery({
    queryKey: ['tse-konfiguration'],
    queryFn: () => einstellungenBackend.getTSEKonfiguration(),
  })

  const saveTSEKonfiguration = async (config: TSEKonfigurationSpeichern) => {
    await einstellungenBackend.saveTSEKonfiguration(config)
    // Speichern/Leeren ändern auch istKonfiguriert — beide Ansichten neu laden.
    await queryClient.invalidateQueries({ queryKey: ['tse-konfiguration'] })
    await queryClient.invalidateQueries({ queryKey: ['tse-status'] })
  }

  const clearTSEKonfiguration = async () => {
    await einstellungenBackend.clearTSEKonfiguration()
    await queryClient.invalidateQueries({ queryKey: ['tse-konfiguration'] })
    await queryClient.invalidateQueries({ queryKey: ['tse-status'] })
  }

  const testTSEVerbindung = async (): Promise<TSEVerbindungStatus> => {
    return einstellungenBackend.testTSEVerbindung()
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

export function pruefeTSESetup(
  zugangsdaten: TSESetupZugangsdaten,
): Promise<TSESetupBefund> {
  return einstellungenBackend.pruefeTSESetup(zugangsdaten)
}

export function useTSEEinrichtung() {
  const queryClient = useQueryClient()

  const richteTSEEin = async (
    eingabe: TSEEinrichten,
  ): Promise<TSEEinrichtenErgebnis> => {
    const ergebnis = await einstellungenBackend.richteTSEEin(eingabe)
    // Die Konfiguration ist jetzt gespeichert — abhängige Ansichten neu laden.
    await queryClient.invalidateQueries({ queryKey: ['tse-konfiguration'] })
    await queryClient.invalidateQueries({ queryKey: ['tse-status'] })
    return ergebnis
  }

  const uebernimmTSE = async (
    eingabe: TSEUebernehmen,
  ): Promise<TSEEinrichtenErgebnis> => {
    const ergebnis = await einstellungenBackend.uebernimmTSE(eingabe)
    await queryClient.invalidateQueries({ queryKey: ['tse-konfiguration'] })
    await queryClient.invalidateQueries({ queryKey: ['tse-status'] })
    return ergebnis
  }

  return { richteTSEEin, uebernimmTSE }
}

export function useTSESignaturQueue() {
  const { data, isPending, error } = useQuery({
    queryKey: ['tse-signatur-queue'],
    queryFn: () => einstellungenBackend.getTSESignaturQueue(),
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
    queryFn: () => einstellungenBackend.getTSEStoerungen(),
  })

  return { stoerungen: data, isPending, error }
}

export function useTSEStatus() {
  const { data, isPending, error } = useQuery({
    queryKey: ['tse-status'],
    queryFn: () => einstellungenBackend.getTSEStatus(),
  })

  return {
    tseStatus: data,
    isPending,
    error,
  }
}
