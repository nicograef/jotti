import { useQuery } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import type { Tisch } from './Tisch'
import { TischBackend } from './TischBackend'

const tischBackend = new TischBackend(BackendSingleton)

export const ALLE_TISCHE_KEY = 'alle-tische'

export function useAllTische() {
  const { data: tische = [] as Tisch[], isPending } = useQuery({
    queryKey: [ALLE_TISCHE_KEY],
    queryFn: () => tischBackend.getAllTische(),
  })
  return { tische, isPending }
}
