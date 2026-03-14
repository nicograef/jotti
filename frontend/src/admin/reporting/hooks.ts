import { useCallback, useEffect, useRef, useState } from 'react'
import { toast } from 'sonner'

import { BackendSingleton } from '@/lib/Backend'
import { useFetch } from '@/lib/useFetch'

import { ReportingBackend } from './ReportingBackend'
import type { Dashboard, Tagesabrechnung } from './types'

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

function initTodayMidnight(): Date {
  const d = new Date()
  d.setHours(0, 0, 0, 0)
  return d
}

function initNowTime(): string {
  const pad = (n: number) => String(n).padStart(2, '0')
  const now = new Date()
  return `${pad(now.getHours())}:${pad(now.getMinutes())}`
}

function combineDateTime(date: Date | undefined, time: string): Date | null {
  if (!date) return null
  const [hours, minutes] = time.split(':').map(Number)
  const result = new Date(date)
  result.setHours(hours, minutes, 0, 0)
  return result
}

/** Manages Tagesabrechnung filter state and fetches data on mount. */
export function useTagesabrechnung() {
  const [vonDate, setVonDate] = useState<Date | undefined>(initTodayMidnight)
  const [vonTime, setVonTime] = useState('00:00')
  const [vonOpen, setVonOpen] = useState(false)
  const [bisDate, setBisDate] = useState<Date | undefined>(() => new Date())
  const [bisTime, setBisTime] = useState(initNowTime)
  const [bisOpen, setBisOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<Tagesabrechnung | null>(null)

  const auswerten = useCallback(async () => {
    const vonDT = combineDateTime(vonDate, vonTime)
    const bisDT = combineDateTime(bisDate, bisTime)

    if (!vonDT || !bisDT) {
      toast.error('Bitte Datum und Uhrzeit auswählen.')
      return
    }

    const vonUTC = vonDT.toISOString()
    const bisUTC = bisDT.toISOString()

    if (vonUTC >= bisUTC) {
      toast.error('"Von" muss vor "Bis" liegen.')
      return
    }

    setLoading(true)
    try {
      const data = await reportingBackend.getTagesabrechnung(vonUTC, bisUTC)
      setResult(data)
    } catch {
      toast.error('Fehler beim Laden der Tagesabrechnung.')
    } finally {
      setLoading(false)
    }
  }, [vonDate, vonTime, bisDate, bisTime])

  useEffect(() => {
    void auswerten()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return {
    vonDate,
    vonTime,
    vonOpen,
    bisDate,
    bisTime,
    bisOpen,
    loading,
    result,
    setVonDate,
    setVonTime,
    setVonOpen,
    setBisDate,
    setBisTime,
    setBisOpen,
    auswerten: () => void auswerten(),
  }
}
