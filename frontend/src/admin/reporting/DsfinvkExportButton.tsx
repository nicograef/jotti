import { Download, Loader2 } from 'lucide-react'

import { Button } from '@/components/ui/button'

import { useDsfinvkExport } from './hooks'

// DsfinvkExportButton lädt das DSFinV-K-Archiv der gewählten Kassensitzung
// herunter (Self-Service für nicht-technische Vereins-Admins).
export function DsfinvkExportButton({
  kassensitzungNr,
}: {
  kassensitzungNr: number | null
}) {
  const { exportieren, isPending } = useDsfinvkExport()

  return (
    <Button
      variant="outline"
      disabled={kassensitzungNr === null || isPending}
      onClick={() => {
        exportieren(kassensitzungNr)
      }}
    >
      {isPending ? (
        <Loader2 className="size-4 animate-spin" />
      ) : (
        <Download className="size-4" />
      )}
      DSFinV-K-Export
    </Button>
  )
}
