import { useState } from 'react'

import { Switch } from '@/components/ui/switch'
import { cn, formatCents } from '@/lib/utils'

import { EditVariantDialog } from './EditVariantDialog'
import { type Variante, VarianteStatus } from './Produkt'
import type { ProduktBackend } from './ProduktBackend'

interface VariantChipProps {
  produktId: number
  variant: Variante
  loading: boolean
  backend: Pick<ProduktBackend, 'updateVariante' | 'deleteVariante'>
  onActivate: (varianteId: number) => Promise<void>
  onDeactivate: (varianteId: number) => Promise<void>
  onUpdated: (variante: Variante) => void
  onDeleted: () => void
}

// Varianten-Chip der Preisliste (Design-Handoff 1c): Name und Preis öffnen per
// Klick den Bearbeiten-Dialog, der Mini-Switch schaltet die Variante direkt
// (aktiv/inaktiv) ohne Dialog. Inaktive Chips sind gedämpft und mit „aus"
// markiert.
export function VariantChip(props: VariantChipProps) {
  const [editOpen, setEditOpen] = useState(false)
  const isActive = props.variant.status === VarianteStatus.ACTIVE

  return (
    <>
      <span
        className={cn(
          'inline-flex items-center gap-2 rounded-full border py-1 pl-3 pr-2 text-sm',
          isActive ? 'bg-background' : 'bg-muted/50 text-muted-foreground',
        )}
      >
        <button
          type="button"
          className="cursor-pointer inline-flex items-center gap-1.5"
          aria-label={`Variante „${props.variant.name}" bearbeiten`}
          onClick={() => {
            setEditOpen(true)
          }}
        >
          <span>{props.variant.name}</span>
          <strong className="font-semibold">
            {formatCents(props.variant.preisCents)} €
          </strong>
          {!isActive && (
            <span className="text-xs uppercase tracking-wide">aus</span>
          )}
        </button>
        <Switch
          className="cursor-pointer"
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
