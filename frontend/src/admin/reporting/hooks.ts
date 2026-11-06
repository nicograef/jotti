import {
  useMutation,
  useQuery,
  type UseQueryResult,
} from '@tanstack/react-query'
import { toast } from 'sonner'

import { useOffenerVorgang } from '@/hooks/use-offener-vorgang'
import { BackendSingleton } from '@/lib/Backend'
import { triggerBrowserDownload } from '@/lib/download'
import { getActionErrorMessage } from '@/lib/errorMessages'

import { ReportingBackend } from './ReportingBackend'
import type {
  AbgeschlosseneSitzung,
  LiveReportingData,
  ReportingData,
} from './types'

const reportingBackend = new ReportingBackend(BackendSingleton)

export const ABGESCHLOSSENE_KASSENSITZUNGEN_KEY =
  'abgeschlossene-kassensitzungen'
export const REPORT_KEY = 'report'
export const LIVE_REPORTING_KEY = 'live-reporting'

export function useAbgeschlosseneKassensitzungen(): {
  kassensitzungen: AbgeschlosseneSitzung[]
  isPending: boolean
  isError: boolean
  refetch: UseQueryResult<AbgeschlosseneSitzung[]>['refetch']
} {
  const {
    data: kassensitzungen = [] as AbgeschlosseneSitzung[],
    isPending,
    isError,
    refetch,
  } = useQuery({
    queryKey: [ABGESCHLOSSENE_KASSENSITZUNGEN_KEY],
    queryFn: () => reportingBackend.getAbgeschlosseneKassensitzungen(),
  })
  return { kassensitzungen, isPending, isError, refetch }
}

export function useReport(kassensitzungNr: number | null): {
  result: ReportingData | null
  isPending: boolean
  isError: boolean
  refetch: UseQueryResult<ReportingData>['refetch']
} {
  const {
    data: result = null as ReportingData | null,
    isPending,
    isError,
    refetch,
  } = useQuery({
    queryKey: [REPORT_KEY, kassensitzungNr],
    queryFn: () => reportingBackend.getReporting(kassensitzungNr ?? 0),
    enabled: kassensitzungNr !== null,
  })
  return { result, isPending, isError, refetch }
}

interface LiveReportingResult {
  liveData: LiveReportingData | null
  isPending: boolean
  dataUpdatedAt: number
  refetch: UseQueryResult<LiveReportingData | null>['refetch']
}

export function useLiveReporting(): LiveReportingResult {
  const {
    data: liveData = null as LiveReportingData | null,
    isPending,
    dataUpdatedAt,
    refetch,
  } = useQuery({
    queryKey: [LIVE_REPORTING_KEY],
    queryFn: () => reportingBackend.getLiveReporting(),
    // 30 s Auto-Refresh ist vertretbar: lokaler Server, eine Admin-Session.
    refetchInterval: 30_000,
  })
  return { liveData, isPending, dataUpdatedAt, refetch }
}

export function useDsfinvkExport(): {
  exportieren: (kassensitzungNr: number) => void
  isPending: boolean
} {
  const mutation = useMutation({
    mutationFn: (kassensitzungNr: number) =>
      reportingBackend.exportDsfinvk(kassensitzungNr),
    onSuccess: ({ blob, filename }) => {
      triggerBrowserDownload(blob, filename)
      toast.success('DSFinV-K-Archiv heruntergeladen.')
    },
    onError: (error) => {
      toast.error(
        getActionErrorMessage({
          actionLabel: 'DSFinV-K-Export',
          error,
          byCode: {
            leere_kassensitzung:
              'Diese Kassensitzung enthält keine Vorgänge zum Exportieren.',
            kassensitzung_nicht_gefunden:
              'Die gewählte Kassensitzung wurde nicht gefunden. Bitte neu auswählen und erneut versuchen.',
          },
        }),
      )
    },
  })

  // Ein Reload mitten im Lauf bräche das Archiv ab, bevor der Browser es als
  // Download übernimmt.
  useOffenerVorgang(mutation.isPending)

  return { exportieren: mutation.mutate, isPending: mutation.isPending }
}
