import { useQuery, useQueryClient } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'
import { DruckerBackend, type DruckerConfig } from '@/lib/DruckerBackend'
import {
  type Betreiber,
  EinstellungenBackend,
} from '@/lib/EinstellungenBackend'

const druckerBackend = new DruckerBackend(BackendSingleton)
const einstellungenBackend = new EinstellungenBackend(BackendSingleton)

export function useDruckerConfig() {
  const queryClient = useQueryClient()
  const {
    isPending,
    data = [],
    error,
  } = useQuery({
    queryKey: ['drucker-config'],
    queryFn: () => druckerBackend.getDruckerConfig(),
  })

  const updateDruckerConfig = async (newConfig: DruckerConfig) => {
    await druckerBackend.updateDruckerConfig(newConfig)
    await queryClient.invalidateQueries({ queryKey: ['drucker-config'] })
  }

  return { drucker: data, isPending, error, updateDruckerConfig }
}

export function useSeriennummer() {
  return useQuery({
    queryKey: ['seriennummer'],
    queryFn: () => einstellungenBackend.getSeriennummer(),
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
