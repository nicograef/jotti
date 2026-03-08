import { BackendSingleton } from '@/lib/Backend'
import { useFetch } from '@/lib/useFetch'

import type { Tisch } from './Tisch'
import { TischBackend } from './TischBackend'

const tischBackend = new TischBackend(BackendSingleton)

/** Custom hook to fetch all tables from backend. */
export function useAllTische() {
  const {
    data: tische,
    setData: setTische,
    ...rest
  } = useFetch(() => tischBackend.getAllTische(), [] as Tisch[])
  return { ...rest, tische, setTische }
}
