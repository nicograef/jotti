import { ClipboardList, Loader2 } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

export function ReportingFilter({
  kassensitzungNr,
  loading,
  onKassensitzungNrChange,
  onAuswerten,
}: {
  kassensitzungNr: number | null
  loading: boolean
  onKassensitzungNrChange: (nr: number | null) => void
  onAuswerten: () => void
}) {
  return (
    <div className="flex flex-wrap items-end gap-x-8 gap-y-4">
      <div>
        <Label htmlFor="kassensitzung-nr">Kassensitzungs-Nr</Label>
        <Input
          id="kassensitzung-nr"
          type="number"
          min={1}
          value={kassensitzungNr ?? ''}
          onChange={(e) => {
            const val = parseInt(e.target.value, 10)
            onKassensitzungNrChange(isNaN(val) ? null : val)
          }}
          placeholder="z.B. 1"
          className="mt-1 w-32"
        />
      </div>

      <Button onClick={onAuswerten} disabled={loading}>
        {loading ? (
          <Loader2 className="size-4 animate-spin" />
        ) : (
          <ClipboardList className="size-4" />
        )}
        Auswerten
      </Button>
    </div>
  )
}
