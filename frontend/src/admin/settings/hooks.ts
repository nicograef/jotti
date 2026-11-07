import { useQuery, useQueryClient } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'
import type { QueryResult } from '@/lib/queryClient'

import {
  DruckstationBackend,
  type DruckstationConfig,
  type FehlgeschlagenerDruckauftrag,
  type Kategorie,
} from './DruckstationBackend'
import {
  type Zusatzdrucker,
  ZusatzdruckerBackend,
} from './ZusatzdruckerBackend'

const druckstationBackend = new DruckstationBackend(BackendSingleton)
const zusatzdruckerBackend = new ZusatzdruckerBackend(BackendSingleton)

const DRUCKSTATIONEN_KEY = 'druckstationen'
const FEHLGESCHLAGENE_DRUCKAUFTRAEGE_KEY = 'fehlgeschlagene-druckauftraege'
const ZUSATZDRUCKER_KEY = 'zusatzdrucker'

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

interface ZusatzdruckerResult extends QueryResult<Zusatzdrucker[]> {
  zusatzdrucker: Zusatzdrucker[]
  anlegen: (name: string) => Promise<void>
  aendern: (zusatzdrucker: Zusatzdrucker) => Promise<void>
  loeschen: (id: number) => Promise<void>
  testbonDrucken: (id: number) => Promise<void>
}

export function useZusatzdrucker(): ZusatzdruckerResult {
  const queryClient = useQueryClient()
  const {
    isPending,
    isError,
    data = [],
    refetch,
  } = useQuery({
    queryKey: [ZUSATZDRUCKER_KEY],
    queryFn: () => zusatzdruckerBackend.getZusatzdrucker(),
  })

  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: [ZUSATZDRUCKER_KEY] })

  const anlegen = async (name: string) => {
    await zusatzdruckerBackend.anlegen(name)
    await invalidate()
  }

  const aendern = async (zusatzdrucker: Zusatzdrucker) => {
    await zusatzdruckerBackend.aendern(zusatzdrucker)
    await invalidate()
  }

  const loeschen = async (id: number) => {
    await zusatzdruckerBackend.loeschen(id)
    await invalidate()
  }

  const testbonDrucken = async (id: number) => {
    await zusatzdruckerBackend.testbonDrucken(id)
  }

  return {
    zusatzdrucker: data,
    isPending,
    isError,
    refetch,
    anlegen,
    aendern,
    loeschen,
    testbonDrucken,
  }
}
