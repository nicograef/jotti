import { useQuery, useQueryClient } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'
import { DruckerBackend, type DruckerConfig } from '@/lib/DruckerBackend'

const druckerBackend = new DruckerBackend(BackendSingleton)

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
