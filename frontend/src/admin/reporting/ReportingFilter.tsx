import { Loader2 } from 'lucide-react'

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

import type { Kassensitzung } from './types'
import { formatDatum } from './utils'

export function ReportingFilter({
  kassensitzungen,
  kassensitzungNr,
  loading,
  onKassensitzungNrChange,
}: {
  kassensitzungen: Kassensitzung[]
  kassensitzungNr: number | null
  loading: boolean
  onKassensitzungNrChange: (nr: number) => void
}) {
  if (loading && kassensitzungen.length === 0) {
    return <Loader2 className="size-5 animate-spin text-muted-foreground" />
  }

  if (kassensitzungen.length === 0) {
    return (
      <Select disabled>
        <SelectTrigger className="w-72">
          <SelectValue placeholder="Noch keine Kassensitzungen vorhanden" />
        </SelectTrigger>
        <SelectContent />
      </Select>
    )
  }

  return (
    <Select
      value={kassensitzungNr?.toString() ?? ''}
      onValueChange={(val) => {
        onKassensitzungNrChange(parseInt(val, 10))
      }}
      disabled={loading}
    >
      <SelectTrigger className="w-72">
        <SelectValue placeholder="Kassensitzung wählen" />
      </SelectTrigger>
      <SelectContent>
        {kassensitzungen.map((k) => (
          <SelectItem key={k.zNr} value={k.zNr.toString()}>
            {k.status === 'offen' ? '🟢 ' : '🔴 '}
            {formatDatum(k.datum)} ({k.bezeichnung})
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
