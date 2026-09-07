import { ChevronLeft, ChevronRight } from 'lucide-react'
import { useState } from 'react'

import { VariantNamePreis } from '@/components/common/VariantNamePreis'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { cn } from '@/lib/utils'

import { EditVariantDialog } from './EditVariantDialog'
import { type Richtung, type Variante, VarianteStatus } from './Produkt'
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

// Varianten-Chip der Preisliste (Design-Handoff 1c): Name und Preis öffnen per
// Klick den Bearbeiten-Dialog, der Mini-Switch schaltet die Variante direkt
// (aktiv/inaktiv) ohne Dialog. Inaktive Chips sind gedämpft und mit „aus"
// markiert.
//
// Die Chevrons an den Chip-Rändern verschieben die Variante innerhalb ihres
// Produkts. Sie zeigen nach links und rechts, weil die Chips horizontal
// umbrechen — die Pfeilrichtung folgt der sichtbaren Anordnung, nicht der
// Richtungs-Benennung der API. Bei nur einer Variante gibt es nichts zu
// tauschen; dann trägt der Chip die Pfeile gar nicht erst.
export function VariantChip(props: VariantChipProps) {
  const [editOpen, setEditOpen] = useState(false)
  const isActive = props.variant.status === VarianteStatus.ACTIVE
  const verschiebbar = !(props.isFirst && props.isLast)

  // 32-px-Ziele wie die Produkt-Pfeile in ProductItem. Das negative my hält den
  // Chip auf seiner Höhe, obwohl der Button höher ist als sein Inhalt.
  const chevronClass = '-my-1 shrink-0 cursor-pointer rounded-full'

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
              void props.onMove(props.variant.id, 'hoch')
            }}
          >
            <ChevronLeft />
          </Button>
        )}

        <button
          type="button"
          className="flex min-w-0 cursor-pointer items-center gap-1.5"
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
              void props.onMove(props.variant.id, 'runter')
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
