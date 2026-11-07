import { useQuery, useQueryClient } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'
import type { QueryResult } from '@/lib/queryClient'

import {
  DruckstationBackend,
  type DruckstationConfig,
  type FehlgeschlagenerDruckauftrag,
  type Kategorie,
} from './DruckstationBackend'

const druckstationBackend = new DruckstationBackend(BackendSingleton)

export const DRUCKSTATIONEN_KEY = 'druckstationen'
export const FEHLGESCHLAGENE_DRUCKAUFTRAEGE_KEY =
  'fehlgeschlagene-druckauftraege'

interface DruckstationenResult extends QueryResult<DruckstationConfig[]> {
  druckstationen: DruckstationConfig[]
  updateDruckstation: (newConfig: DruckstationConfig) => Promise<void>
  testbonDrucken: (kategorie: Kategorie) => Promise<void>
}

export function useDruckstationen(): DruckstationenResult {
  const queryClient = useQueryClient()
  const {
    isPending,
    isError,
    data = [],
    refetch,
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
    isError,
    refetch,
    updateDruckstation,
    testbonDrucken,
  }
}

interface FehlgeschlageneDruckauftraegeResult {
  druckauftraege: FehlgeschlagenerDruckauftrag[]
  erneutVersuchen: (id: number) => Promise<void>
  verwerfen: (id: number) => Promise<void>
  alleVerwerfen: () => Promise<number>
}

export function useFehlgeschlageneDruckauftraege(): FehlgeschlageneDruckauftraegeResult {
  const queryClient = useQueryClient()
  const { data = [] } = useQuery({
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
    erneutVersuchen,
    verwerfen,
    alleVerwerfen,
  }
}
