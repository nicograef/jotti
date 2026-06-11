import { useQuery, useQueryClient } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'
import {
  DruckstationBackend,
  type DruckstationConfig,
} from '@/lib/DruckstationBackend'
import {
  type Betreiber,
  EinstellungenBackend,
  type TSEKonfigurationSpeichern,
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
    await queryClient.invalidateQueries({ queryKey: ['tse-konfiguration'] })
  }

  const clearTSEKonfiguration = async () => {
    await einstellungenBackend.clearTSEKonfiguration()
    await queryClient.invalidateQueries({ queryKey: ['tse-konfiguration'] })
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
