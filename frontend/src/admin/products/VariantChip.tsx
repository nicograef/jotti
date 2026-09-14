import { ChevronLeft, ChevronRight } from 'lucide-react'
import { useState } from 'react'

import { VariantNamePreis } from '@/components/common/VariantNamePreis'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import type { Variante } from '@/lib/produktSchemas'
import { cn } from '@/lib/utils'

import { EditVariantDialog } from './EditVariantDialog'
import { Richtung, VarianteStatus } from './Produkt'
import type { ProduktBackend } from './ProduktBackend'

interface VariantChipProps {
  produktId: number
  variant: Variante
  loading: boolean
  isFirst: boolean
  isLast: boolean
  backend: Pick<ProduktBackend, 'updateVariante' | 'deleteVariante'>
  onActivate: (varianteId: number) => Promise<void>
  onDeactivate: (varianteId: number) => Promise<void>
  onMove: (varianteId: number, richtung: Richtung) => Promise<void>
  onUpdated: (variante: Variante) => void
  onDeleted: () => void
}

// Die Chevrons zeigen nach links und rechts, weil die Chips horizontal
// umbrechen — die Pfeilrichtung folgt der sichtbaren Anordnung, nicht der
// Richtungs-Benennung der API.
export function VariantChip(props: VariantChipProps) {
  const [editOpen, setEditOpen] = useState(false)
  const isActive = props.variant.status === VarianteStatus.ACTIVE
  const verschiebbar = !(props.isFirst && props.isLast)

  // `relative z-10` ist Pflicht: Der Switch bringt eine unsichtbare
  // Trefferflächen-Erweiterung mit (after:-inset-x-3, 12 px je Seite) und liegt
  // sonst über den unpositionierten Geschwistern — ein Tipp auf die zugewandte
  // Kante von Pfeil oder Name-Button schaltete dann die Variante.
  const chevronClass =
    'relative z-10 -my-1 shrink-0 cursor-pointer rounded-full'

  return (
    <>
      <span
        className={cn(
          'inline-flex items-center gap-1.5 rounded-full border py-1 pl-1 pr-1.5 text-sm',
          isActive ? 'bg-background' : 'bg-muted/50 text-muted-foreground',
        )}
      >
        {verschiebbar && (
          <Button
            size="icon-sm"
            variant="ghost"
            className={chevronClass}
            disabled={props.loading || props.isFirst}
            aria-label={`Variante „${props.variant.name}" nach vorne`}
            onClick={() => {
              void props.onMove(props.variant.id, Richtung.HOCH)
            }}
          >
            <ChevronLeft />
          </Button>
        )}

        <button
          type="button"
          className="relative z-10 flex min-w-0 cursor-pointer items-center gap-1.5"
          aria-label={`Variante „${props.variant.name}" bearbeiten`}
          onClick={() => {
            setEditOpen(true)
          }}
        >
          <VariantNamePreis
            name={props.variant.name}
            preisCents={props.variant.preisCents}
          />
          {!isActive && (
            <span className="shrink-0 text-xs uppercase tracking-wide">
              aus
            </span>
          )}
        </button>

        <Switch
          className="shrink-0 cursor-pointer"
          disabled={props.loading}
          checked={isActive}
          aria-label={
            isActive
              ? `Variante „${props.variant.name}" deaktivieren`
              : `Variante „${props.variant.name}" aktivieren`
          }
          onCheckedChange={(checked) => {
            if (checked) {
              void props.onActivate(props.variant.id)
            } else {
              void props.onDeactivate(props.variant.id)
            }
          }}
        />

        {verschiebbar && (
          <Button
            size="icon-sm"
            variant="ghost"
            className={chevronClass}
            disabled={props.loading || props.isLast}
            aria-label={`Variante „${props.variant.name}" nach hinten`}
            onClick={() => {
              void props.onMove(props.variant.id, Richtung.RUNTER)
            }}
          >
            <ChevronRight />
          </Button>
        )}
      </span>

      <EditVariantDialog
        open={editOpen}
        produktId={props.produktId}
        variant={props.variant}
        backend={props.backend}
        updated={props.onUpdated}
        deleted={props.onDeleted}
        close={() => {
          setEditOpen(false)
        }}
      />
    </>
  )
}
