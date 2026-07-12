import { CheckCircle2, Printer, ShieldCheck, TriangleAlert } from 'lucide-react'
import type { ReactNode } from 'react'
import { NavLink } from 'react-router'

import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

// Statuszelle der Übersicht: im Normalzustand neutral gerahmt, im Fehlerfall rot
// (Rahmen destructive/40, Fläche destructive/4) mit „Beheben"-Button zur
// zuständigen Admin-Seite. Ersetzt die früheren Alert-Banner (Design-Handoff 1a).
function StatusZelle({
  icon,
  titel,
  text,
  fehler,
  behebenHref,
}: {
  icon: ReactNode
  titel: string
  text: string
  fehler?: boolean
  behebenHref?: string
}) {
  return (
    <div
      className={cn(
        'flex items-center gap-2.5 rounded-lg border p-3',
        fehler && 'border-destructive/40 bg-destructive/4',
      )}
    >
      <span className={cn('shrink-0', fehler && 'text-destructive')}>
        {icon}
      </span>
      <div className="flex min-w-0 flex-1 flex-col">
        <span
          className={cn(
            'text-[13px] font-semibold',
            fehler && 'text-destructive',
          )}
        >
          {titel}
        </span>
        <span className="truncate text-xs text-muted-foreground">{text}</span>
      </div>
      {fehler && behebenHref && (
        <Button
          asChild
          size="sm"
          variant="destructive"
          className="h-7 shrink-0 px-2.5 text-xs"
        >
          <NavLink to={behebenHref}>Beheben</NavLink>
        </Button>
      )}
    </div>
  )
}

// UebersichtStatusZeile bündelt die drei Statuszellen (Kasse, TSE, Drucker) der
// Übersicht. Die Fehlerlogik und die Schwellen liegen im Aufrufer
// (AdminDashboardPage), diese Komponente ist reine Darstellung.
export function UebersichtStatusZeile({
  kasseText,
  tseFehler,
  tseText,
  druckFehler,
  druckTitel,
  druckText,
}: {
  kasseText: string
  tseFehler: boolean
  tseText: string
  druckFehler: boolean
  druckTitel: string
  druckText: string
}) {
  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
      <StatusZelle
        icon={<CheckCircle2 className="size-4 text-primary" />}
        titel="Kasse offen"
        text={kasseText}
      />
      <StatusZelle
        icon={
          tseFehler ? (
            <TriangleAlert className="size-4" />
          ) : (
            <ShieldCheck className="size-4 text-primary" />
          )
        }
        titel={tseFehler ? 'TSE benötigt Aufmerksamkeit' : 'TSE signiert'}
        text={tseText}
        fehler={tseFehler}
        behebenHref="/admin/finanzamt"
      />
      <StatusZelle
        icon={
          druckFehler ? (
            <TriangleAlert className="size-4" />
          ) : (
            <Printer className="size-4 text-primary" />
          )
        }
        titel={druckTitel}
        text={druckText}
        fehler={druckFehler}
        behebenHref="/admin/druckstationen"
      />
    </div>
  )
}
