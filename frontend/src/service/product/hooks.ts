import { useQuery } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import { ProduktBackend } from './ProduktBackend'

const produktBackend = new ProduktBackend(BackendSingleton)

export function useAktiveProdukte() {
  const { data = [], isPending } = useQuery({
    queryKey: ['aktive-produkte'],
    queryFn: () => produktBackend.getAktiveProdukte(),
  })
  return { produkte: data, isPending }
}
