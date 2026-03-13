import { useEffect, useRef } from 'react'

import { BackendSingleton } from '@/lib/Backend'
import { useFetch } from '@/lib/useFetch'

import { ReportingBackend } from './ReportingBackend'
import type { Dashboard } from './types'

const reportingBackend = new ReportingBackend(BackendSingleton)

const emptyDashboard: Dashboard = {
  gesamtUmsatzCents: 0,
  anzahlOffeneTische: 0,
  anzahlBestellungen: 0,
  anzahlStornierungen: 0,
  gesamtBestellungenCents: 0,
  gesamtStornierungenCents: 0,
}

/** Fetches dashboard data on mount and auto-refreshes every 60 seconds. */
export function useDashboard() {
  const result = useFetch(() => reportingBackend.getDashboard(), emptyDashboard)

  const reloadRef = useRef(result.reload)
  useEffect(() => {
    reloadRef.current = result.reload
  }, [result.reload])

  useEffect(() => {
    const interval = setInterval(() => {
      reloadRef.current()
    }, 60_000)
    return () => {
      clearInterval(interval)
    }
  }, [])

  return result
}
