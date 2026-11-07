import { Building2, Download, FileText } from 'lucide-react'
import { useState } from 'react'
import { NavLink } from 'react-router'

import { useAktiveKassensitzung } from '@/admin/kasse/hooks'
import { LadefehlerAlert } from '@/components/common/LadefehlerAlert'
import { Button } from '@/components/ui/button'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty'
import { Spinner } from '@/components/ui/spinner'

import { AdminPageHeader } from '../components/AdminPageHeader'
import {
  useAbgeschlosseneKassensitzungen,
  useDsfinvkExport,
  useReport,
} from './hooks'
import { ReportingResults } from './ReportingResults'
import { SitzungsListe } from './SitzungsListe'

function ExportBlock({ kassensitzungNr }: { kassensitzungNr: number }) {
  const { exportieren, isPending } = useDsfinvkExport()

  return (
    <div className="flex flex-col gap-4 rounded-xl border bg-sidebar p-4 sm:flex-row sm:items-center print:hidden">
      <Building2 className="size-5 shrink-0 text-primary" aria-hidden />
      <div className="flex-1">
        <p className="text-sm font-semibold">Für Steuerberater & Finanzamt</p>
        <p className="mt-0.5 text-sm text-muted-foreground">
          Das DSFinV-K-Archiv ist das maschinenlesbare Kassenprotokoll dieser
          Sitzung. Bei einer Prüfung wird genau diese Datei verlangt — einfach
          herunterladen und weitergeben.
        </p>
      </div>
      <Button
        className="shrink-0"
        disabled={isPending}
        onClick={() => {
          exportieren(kassensitzungNr)
        }}
      >
        {isPending ? <Spinner /> : <Download className="size-4" />}
        Archiv herunterladen (ZIP)
      </Button>
    </div>
  )
}

export function KassenberichtePage() {
  const {
    kassensitzungen,
    isPending: listLoading,
    isError: listFehler,
    refetch: listNeuLaden,
  } = useAbgeschlosseneKassensitzungen()
  const { kassensitzung: aktiveSitzung, abschlussUnterbrochen } =
    useAktiveKassensitzung()
  const [selectedNr, setSelectedNr] = useState<number | null>(null)

  const effectiveNr = selectedNr ?? kassensitzungen.at(0)?.zNr ?? null
  const selectedSitzung =
    kassensitzungen.find((k) => k.zNr === effectiveNr) ?? null
  const {
    result,
    isPending: reportLoading,
    isError: reportFehler,
    refetch: reportNeuLaden,
  } = useReport(effectiveNr)

  return (
    <>
      {/* Der generische Seitenkopf gehört nicht auf den gedruckten Z-Bon. */}
      <div className="print:hidden">
        <AdminPageHeader
          titel="Berichte & Export"
          unterzeile="Jede abgeschlossene Kassensitzung ergibt einen Tagesbericht (Z-Bon)."
        />
      </div>

      {/* Ohne Liste wäre der Leer-Zustand eine falsche Aussage über die Z-Bons. */}
      {listFehler ? (
        <LadefehlerAlert
          titel="Kassenberichte konnten nicht geladen werden"
          onErneutVersuchen={() => void listNeuLaden()}
          className="mt-6"
        />
      ) : !listLoading && kassensitzungen.length === 0 ? (
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
        <div className="mt-4 grid grid-cols-1 gap-5 lg:grid-cols-[280px_minmax(0,1fr)]">
          <div className="print:hidden">
            <SitzungsListe
              sitzungen={kassensitzungen}
              aktiveSitzung={aktiveSitzung}
              abschlussUnterbrochen={abschlussUnterbrochen}
              selectedNr={effectiveNr}
              onSelect={setSelectedNr}
            />
          </div>

          <div className="flex flex-col gap-4">
            {reportFehler ? (
              <LadefehlerAlert
                titel="Tagesbericht konnte nicht geladen werden"
                onErneutVersuchen={() => void reportNeuLaden()}
              />
            ) : reportLoading || !result || !selectedSitzung ? (
              <div className="flex items-center justify-center py-16">
                <Spinner className="size-6 text-muted-foreground" />
              </div>
            ) : (
              <>
                <ReportingResults result={result} sitzung={selectedSitzung} />
                <ExportBlock kassensitzungNr={selectedSitzung.zNr} />
              </>
            )}
          </div>
        </div>
      )}
    </>
  )
}
