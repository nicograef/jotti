import { Package } from 'lucide-react'
import { useState } from 'react'

import { EmptyState } from '@/components/common/EmptyState'
import { Skeleton } from '@/components/ui/skeleton'
import { cn, formatEuro } from '@/lib/utils'

import {
  type Kategorie,
  KategorieLabels,
  KategorieOrder,
  type Produkt,
  type Variante,
} from '../../product/Produkt'
import { Stepper } from '../Stepper'

interface ProductListComponentProps {
  products: Produkt[]
  variantMengen: Record<number, number>
  onAdd: (variantId: number) => void
  onRemove: (variantId: number) => void
}

function belegteKategorien(products: Produkt[]): Kategorie[] {
  return KategorieOrder.filter((kategorie) =>
    products.some((p) => p.kategorie === kategorie),
  )
}

export function ProductList(props: ProductListComponentProps) {
  const kategorien = belegteKategorien(props.products)
  const [aktiveKategorie, setAktiveKategorie] = useState<Kategorie | undefined>(
    kategorien[0],
  )

  if (props.products.length === 0) {
    return (
      <EmptyState
        icon={Package}
        title="Keine Produkte verfügbar"
        description="Bitte im Admin-Bereich mindestens eine aktive Variante anlegen."
      />
    )
  }

  const angezeigteKategorie =
    aktiveKategorie && kategorien.includes(aktiveKategorie)
      ? aktiveKategorie
      : kategorien[0]
  const sichtbareProdukte = props.products.filter(
    (p) => p.kategorie === angezeigteKategorie,
  )

  return (
    <div>
      {kategorien.length > 1 && (
        <div className="sticky top-14 z-20 -mx-4 border-b bg-background px-4 pb-3 pt-1.5 md:-mx-8 md:px-8 lg:top-0 lg:mx-0 lg:px-0">
          <div className="flex flex-wrap gap-2">
            {kategorien.map((kategorie) => {
              const aktiv = kategorie === angezeigteKategorie
              return (
                <button
                  key={kategorie}
                  type="button"
                  onClick={() => {
                    setAktiveKategorie(kategorie)
                  }}
                  className={cn(
                    'h-9 rounded-full px-4 text-sm font-medium transition-colors',
                    aktiv
                      ? 'bg-foreground text-background'
                      : 'border text-foreground',
                  )}
                >
                  {KategorieLabels[kategorie]}
                </button>
              )
            })}
          </div>
        </div>
      )}
      <div className="mt-4 space-y-5">
        {sichtbareProdukte.map((product) => (
          <div key={product.id}>
            <h2 className="mb-1 text-[13px] font-semibold uppercase tracking-wide text-muted-foreground">
              {product.name}
            </h2>
            <div>
              {product.varianten.map((variant) => (
                <VariantRow
                  key={variant.id}
                  variant={variant}
                  menge={props.variantMengen[variant.id] || 0}
                  onAdd={() => {
                    props.onAdd(variant.id)
                  }}
                  onRemove={() => {
                    props.onRemove(variant.id)
                  }}
                />
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

// VariantRow ist die Bestellzeile einer Variante: Name in der ersten Zeile,
// Preis darunter, Mengensteuerung rechts. Der Name teilt sich die Breite nur
// mit der Steuerung und braucht deshalb in der Praxis keinen Umbruch mehr; er
// kürzt auch nicht, denn „Schorle weiß, sauer" und „Schorle weiß, süß" kürzen
// sich auf denselben Text und die Servicekraft bucht die falsche Variante.
// Solange nichts ausgewählt ist, zeigt die Zeile nur das Plus — das hält die
// Liste ruhig und gibt dem Namen die volle Breite.
function VariantRow({
  variant,
  menge,
  onAdd,
  onRemove,
}: {
  variant: Variante
  menge: number
  onAdd: () => void
  onRemove: () => void
}) {
  return (
    <div
      className={cn(
        'flex items-center gap-3 border-b py-2 last:border-b-0',
        menge > 0 && 'bg-primary/[0.04]',
      )}
    >
      <div className="min-w-0 flex-1">
        <div className="break-words text-[15px] font-medium leading-snug">
          {variant.name}
        </div>
        <div className="text-sm tabular-nums text-muted-foreground">
          {formatEuro(variant.preisCents)}
        </div>
      </div>
      <Stepper
        menge={menge}
        onAdd={onAdd}
        onRemove={onRemove}
        addLabel="Variante hinzufügen"
        removeLabel="Variante entfernen"
        minusNurAbEins
      />
    </div>
  )
}

export function ProductListSkeleton() {
  return (
    <div className="mt-4 space-y-5">
      {Array.from({ length: 2 }).map((_, gruppe) => (
        <div key={`skeleton-gruppe-${gruppe.toString()}`}>
          <Skeleton className="mb-1 h-4 w-24" />
          <div>
            {Array.from({ length: 3 }).map((_, zeile) => (
              <div
                key={`skeleton-zeile-${gruppe.toString()}-${zeile.toString()}`}
                className="flex items-center justify-between gap-3 border-b py-2 last:border-b-0"
              >
                <div className="flex-1">
                  <Skeleton className="h-5 w-40" />
                  <Skeleton className="mt-1 h-4 w-12" />
                </div>
                <Skeleton className="size-11 rounded-full" />
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}
