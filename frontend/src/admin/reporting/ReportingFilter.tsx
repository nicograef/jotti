import { format } from 'date-fns'
import { de } from 'date-fns/locale'
import { ChevronDownIcon, ClipboardList, Loader2 } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Calendar } from '@/components/ui/calendar'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'

export function ReportingFilter({
  vonDate,
  vonTime,
  vonOpen,
  bisDate,
  bisTime,
  bisOpen,
  loading,
  onVonDateChange,
  onVonTimeChange,
  onVonOpenChange,
  onBisDateChange,
  onBisTimeChange,
  onBisOpenChange,
  onAuswerten,
}: {
  vonDate: Date | undefined
  vonTime: string
  vonOpen: boolean
  bisDate: Date | undefined
  bisTime: string
  bisOpen: boolean
  loading: boolean
  onVonDateChange: (d: Date | undefined) => void
  onVonTimeChange: (t: string) => void
  onVonOpenChange: (open: boolean) => void
  onBisDateChange: (d: Date | undefined) => void
  onBisTimeChange: (t: string) => void
  onBisOpenChange: (open: boolean) => void
  onAuswerten: () => void
}) {
  return (
    <div className="flex flex-wrap items-end gap-x-8 gap-y-4">
      {/* Von */}
      <div className="flex items-end gap-2">
        <div>
          <Label>Von</Label>
          <Popover open={vonOpen} onOpenChange={onVonOpenChange}>
            <PopoverTrigger asChild>
              <Button
                variant="outline"
                className="mt-1 w-36 justify-between font-normal"
              >
                {vonDate
                  ? format(vonDate, 'dd. MMM yyyy', { locale: de })
                  : 'Datum wählen'}
                <ChevronDownIcon className="size-4" />
              </Button>
            </PopoverTrigger>
            <PopoverContent
              className="w-auto overflow-hidden p-0"
              align="start"
            >
              <Calendar
                mode="single"
                selected={vonDate}
                captionLayout="dropdown"
                defaultMonth={vonDate}
                onSelect={(d) => {
                  onVonDateChange(d)
                  onVonOpenChange(false)
                }}
              />
            </PopoverContent>
          </Popover>
        </div>
        <div className="w-28">
          <Label htmlFor="von-time">Uhrzeit</Label>
          <Input
            type="time"
            id="von-time"
            value={vonTime}
            onChange={(e) => {
              onVonTimeChange(e.target.value)
            }}
            className="mt-1 appearance-none bg-background [&::-webkit-calendar-picker-indicator]:hidden [&::-webkit-calendar-picker-indicator]:appearance-none"
          />
        </div>
      </div>

      {/* Bis */}
      <div className="flex items-end gap-2">
        <div>
          <Label>Bis</Label>
          <Popover open={bisOpen} onOpenChange={onBisOpenChange}>
            <PopoverTrigger asChild>
              <Button
                variant="outline"
                className="mt-1 w-36 justify-between font-normal"
              >
                {bisDate
                  ? format(bisDate, 'dd. MMM yyyy', { locale: de })
                  : 'Datum wählen'}
                <ChevronDownIcon className="size-4" />
              </Button>
            </PopoverTrigger>
            <PopoverContent
              className="w-auto overflow-hidden p-0"
              align="start"
            >
              <Calendar
                mode="single"
                selected={bisDate}
                captionLayout="dropdown"
                defaultMonth={bisDate}
                onSelect={(d) => {
                  onBisDateChange(d)
                  onBisOpenChange(false)
                }}
              />
            </PopoverContent>
          </Popover>
        </div>
        <div className="w-28">
          <Label htmlFor="bis-time">Uhrzeit</Label>
          <Input
            type="time"
            id="bis-time"
            value={bisTime}
            onChange={(e) => {
              onBisTimeChange(e.target.value)
            }}
            className="mt-1 appearance-none bg-background [&::-webkit-calendar-picker-indicator]:hidden [&::-webkit-calendar-picker-indicator]:appearance-none"
          />
        </div>
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
