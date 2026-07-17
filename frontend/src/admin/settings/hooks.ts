import { useQuery, useQueryClient } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import {
  DruckstationBackend,
  type DruckstationConfig,
  type Kategorie,
} from './DruckstationBackend'

const druckstationBackend = new DruckstationBackend(BackendSingleton)

// Query-Keys der Druckstation-Ansichten. Nach Konfigurations- und
// Druckauftrag-Aktionen werden die Listen über diese Keys invalidiert.
export const DRUCKSTATIONEN_KEY = 'druckstationen'
export const FEHLGESCHLAGENE_DRUCKAUFTRAEGE_KEY =
  'fehlgeschlagene-druckauftraege'

export function useDruckstationen() {
  const queryClient = useQueryClient()
  const {
    isPending,
    data = [],
    error,
  } = useQuery({
    queryKey: [DRUCKSTATIONEN_KEY],
    queryFn: () => druckstationBackend.getDruckstationen(),
  })

  const updateDruckstation = async (newConfig: DruckstationConfig) => {
    await druckstationBackend.updateDruckstation(newConfig)
    await queryClient.invalidateQueries({ queryKey: [DRUCKSTATIONEN_KEY] })
  }

  const testbonDrucken = async (kategorie: Kategorie) => {
    await druckstationBackend.testbonDrucken(kategorie)
  }

  return {
    druckstationen: data,
    isPending,
    error,
    updateDruckstation,
    testbonDrucken,
  }
}

export function useFehlgeschlageneDruckauftraege() {
  const queryClient = useQueryClient()
  const {
    isPending,
    data = [],
    error,
  } = useQuery({
    queryKey: [FEHLGESCHLAGENE_DRUCKAUFTRAEGE_KEY],
    queryFn: () => druckstationBackend.getFehlgeschlageneDruckauftraege(),
  })

  const erneutVersuchen = async (id: number) => {
    await druckstationBackend.druckauftragErneutVersuchen(id)
    await queryClient.invalidateQueries({
      queryKey: [FEHLGESCHLAGENE_DRUCKAUFTRAEGE_KEY],
    })
  }

  const verwerfen = async (id: number) => {
    await druckstationBackend.druckauftragVerwerfen(id)
    await queryClient.invalidateQueries({
      queryKey: [FEHLGESCHLAGENE_DRUCKAUFTRAEGE_KEY],
    })
  }

  const alleVerwerfen = async () => {
    const verworfen = await druckstationBackend.druckauftraegeVerwerfen()
    await queryClient.invalidateQueries({
      queryKey: [FEHLGESCHLAGENE_DRUCKAUFTRAEGE_KEY],
    })
    return verworfen
  }

  return {
    druckauftraege: data,
    isPending,
    error,
    erneutVersuchen,
    verwerfen,
    alleVerwerfen,
  }
}
