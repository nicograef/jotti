import { useQuery, useQueryClient } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import {
  DruckstationBackend,
  type DruckstationConfig,
  type Kategorie,
} from './DruckstationBackend'

const druckstationBackend = new DruckstationBackend(BackendSingleton)

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

  const alleVerwerfen = async () => {
    const verworfen = await druckstationBackend.druckauftraegeVerwerfen()
    await queryClient.invalidateQueries({
      queryKey: ['fehlgeschlagene-druckauftraege'],
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
