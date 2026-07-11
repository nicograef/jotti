import { FileText } from 'lucide-react'
import { useState } from 'react'
import { NavLink } from 'react-router'

import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty'

import { DsfinvkExportButton } from './DsfinvkExportButton'
import { useAbgeschlosseneKassensitzungen, useReport } from './hooks'
import { ReportingFilter } from './ReportingFilter'
import { ReportingResults } from './ReportingResults'

// Kassenberichte zeigen die historische Auswertung abgeschlossener
// Kassensitzungen samt Steuersatz-Tabelle und DSFinV-K-Export. Laufende
// (offene) Sitzungen erscheinen ausschließlich auf dem Live-Dashboard.
export function KassenberichtePage() {
  const { kassensitzungen, isPending: listLoading } =
    useAbgeschlosseneKassensitzungen()
  const [selectedNr, setSelectedNr] = useState<number | null>(null)

  const effectiveNr = selectedNr ?? kassensitzungen.at(0)?.zNr ?? null
  const { result, isPending: reportLoading } = useReport(effectiveNr)

  return (
    <>
      <h1 className="text-2xl font-bold">Kassenberichte</h1>

      {!listLoading && kassensitzungen.length === 0 ? (
        <Empty className="mt-6">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <FileText />
            </EmptyMedia>
            <EmptyTitle>Noch keine abgeschlossene Kassensitzung</EmptyTitle>
            <EmptyDescription>
              Kassenberichte erscheinen hier, sobald eine Kassensitzung
              abgeschlossen wurde.{' '}
              <NavLink
                to="/admin/kasse"
                className="underline underline-offset-4"
              >
                Zur Kassensitzungs-Seite
              </NavLink>
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      ) : (
        <>
          <div className="mt-4 flex flex-wrap items-center gap-3">
            <ReportingFilter
              kassensitzungen={kassensitzungen}
              kassensitzungNr={effectiveNr}
              loading={listLoading}
              onKassensitzungNrChange={setSelectedNr}
            />
            <DsfinvkExportButton kassensitzungNr={effectiveNr} />
          </div>

          {result && (
            <div className="my-6">
              <ReportingResults result={result} loading={reportLoading} />
            </div>
          )}
        </>
      )}
    </>
  )
}
