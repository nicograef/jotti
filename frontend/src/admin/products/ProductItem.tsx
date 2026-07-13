import { MoreHorizontal, Pen, Plus, PowerOff, Trash2 } from 'lucide-react'
import { useState } from 'react'

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { useActionSubmit } from '@/hooks/use-action-submit'

import { NewVariantDialog } from './NewVariantDialog'
import { type Produkt, type Variante, VarianteStatus } from './Produkt'
import type { ProduktBackend } from './ProduktBackend'
import { VariantChip } from './VariantChip'

interface ProductItemProps {
  loading: boolean
  product: Produkt
  backend: Pick<
    ProduktBackend,
    | 'aktiviereVariante'
    | 'deaktiviereVariante'
    | 'createVariante'
    | 'updateVariante'
    | 'deleteVariante'
  >
  onEdit: (produktId: number) => void
  onDelete: (produktId: number) => Promise<void>
  onVariantCreated: (variante: Variante) => void
  onVariantUpdated: (variante: Variante) => void
  onVariantStatusChange: (varianteId: number, status: VarianteStatus) => void
  onVariantDeleted: () => void
}

export function ProductItem(props: ProductItemProps) {
  const [deleteOpen, setDeleteOpen] = useState(false)

  const { loading: deleteLoading, run: runDeleteProduct } = useActionSubmit({
    actionLabel: 'Produkt löschen',
  })
  const { loading: activateVariantLoading, run: runActivateVariant } =
    useActionSubmit({ actionLabel: 'Variante aktivieren' })
  const { loading: deactivateVariantLoading, run: runDeactivateVariant } =
    useActionSubmit({ actionLabel: 'Variante deaktivieren' })
  const {
    loading: deactivateAllVariantsLoading,
    run: runDeactivateAllVariants,
  } = useActionSubmit({ actionLabel: 'Varianten deaktivieren' })

  const variantLoading =
    activateVariantLoading ||
    deactivateVariantLoading ||
    deactivateAllVariantsLoading

  const activeVarianten = props.product.varianten.filter(
    (v) => v.status === VarianteStatus.ACTIVE,
  )

  const handleActivateVariant = async (variantId: number) => {
    await runActivateVariant(async () => {
      await props.backend.aktiviereVariante(variantId)
      props.onVariantStatusChange(variantId, VarianteStatus.ACTIVE)
    })
  }

  const handleDeactivateVariant = async (variantId: number) => {
    await runDeactivateVariant(async () => {
      await props.backend.deaktiviereVariante(variantId)
      props.onVariantStatusChange(variantId, VarianteStatus.INACTIVE)
    })
  }

  const handleDeactivateAllVariants = async () => {
    await runDeactivateAllVariants(async () => {
      for (const variante of activeVarianten) {
        await props.backend.deaktiviereVariante(variante.id)
        props.onVariantStatusChange(variante.id, VarianteStatus.INACTIVE)
      }
    })
  }

  const handleDelete = async () => {
    await runDeleteProduct(async () => {
      await props.onDelete(props.product.id)
      setDeleteOpen(false)
    })
  }

  return (
    <div className="flex flex-wrap items-start gap-x-4 gap-y-2 border-b py-3 last:border-b-0">
      <span className="min-w-32 shrink-0 pt-1 font-medium">
        {props.product.name}
      </span>

      <div className="flex flex-1 flex-wrap items-center gap-2">
        {props.product.varianten.map((variant) => (
          <VariantChip
            key={variant.id}
            produktId={props.product.id}
            variant={variant}
            loading={props.loading || variantLoading}
            backend={props.backend}
            onActivate={handleActivateVariant}
            onDeactivate={handleDeactivateVariant}
            onUpdated={props.onVariantUpdated}
            onDeleted={props.onVariantDeleted}
          />
        ))}
        <NewVariantDialog
          productId={props.product.id}
          backend={props.backend}
          created={props.onVariantCreated}
        >
          <button
            type="button"
            className="inline-flex h-8 cursor-pointer items-center gap-1 rounded-full border border-dashed px-3 text-sm text-muted-foreground"
          >
            <Plus className="size-3.5" /> Variante
          </button>
        </NewVariantDialog>
      </div>

      <div className="flex shrink-0 items-center gap-1">
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              size="icon-sm"
              variant="ghost"
              className="cursor-pointer rounded-full"
              aria-label="Produkt bearbeiten"
              onClick={() => {
                props.onEdit(props.product.id)
              }}
            >
              <Pen />
            </Button>
          </TooltipTrigger>
          <TooltipContent>Bearbeiten</TooltipContent>
        </Tooltip>

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              size="icon-sm"
              variant="ghost"
              className="cursor-pointer rounded-full"
              aria-label="Weitere Aktionen"
            >
              <MoreHorizontal />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem
              onSelect={() => {
                props.onEdit(props.product.id)
              }}
            >
              <Pen /> Umbenennen
            </DropdownMenuItem>
            <DropdownMenuItem
              disabled={activeVarianten.length === 0 || variantLoading}
              onSelect={() => void handleDeactivateAllVariants()}
            >
              <PowerOff /> Alle Varianten deaktivieren
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              variant="destructive"
              onSelect={() => {
                setDeleteOpen(true)
              }}
            >
              <Trash2 /> Löschen…
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      <AlertDialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Produkt löschen?</AlertDialogTitle>
            <AlertDialogDescription>
              Das Produkt &quot;{props.product.name}&quot; und alle zugehörigen
              Varianten werden unwiderruflich gelöscht.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Abbrechen</AlertDialogCancel>
            <AlertDialogAction
              className="bg-destructive text-white hover:bg-destructive/90"
              onClick={(e) => {
                e.preventDefault()
                void handleDelete()
              }}
              disabled={deleteLoading}
            >
              Löschen
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
