import { Package } from 'lucide-react'
import { useState } from 'react'

import { EmptyState } from '@/components/common/EmptyState'
import { Skeleton } from '@/components/ui/skeleton'
import type { Kategorie, Produkt, Variante } from '@/lib/produktSchemas'
import { cn, formatEuro } from '@/lib/utils'

import { Stepper } from '../Stepper'

// Deutsche Labels und feste Anzeigereihenfolge der Kategorie-Abschnitte dieser
// Liste.
const KategorieLabels: Record<Kategorie, string> = {
  essen: 'Essen',
  getraenk: 'Getränke',
  sonstiges: 'Sonstiges',
}
const KategorieOrder: Kategorie[] = ['essen', 'getraenk', 'sonstiges']

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
            <h2 className="mb-1 text-[13px] font-semibold text-muted-foreground">
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
// Preis darunter, Mengensteuerung rechts. Der Name kürzt nie, sondern bricht
// um — „Schorle weiß, sauer" und „Schorle weiß, süß" kürzen sich auf denselben
// Text, und die Servicekraft bucht dann die falsche Variante. Lange Namen
// belegen deshalb mehrere Zeilen.
// Die Mengensteuerung sitzt in einem Slot fester Breite: 8,25 rem (132 px bei
// 16-px-Wurzelschrift: Minus 2,75 + Lücke 0,5 + Menge 1,75 + Lücke 0,5 + Plus
// 2,75) — die volle Stepper-Breite in derselben Einheit wie der Stepper selbst.
// Dadurch bleibt die Namensspalte unabhängig von der Menge gleich breit:
// Solange nichts ausgewählt ist, zeigt die Zeile nur das rechtsbündige Plus
// (minusNurAbEins), und der erste Tap bricht weder den Namen neu um noch
// schiebt er die Zeilen darunter nach unten.
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
      <div className="flex w-[8.25rem] shrink-0 justify-end">
        <Stepper
          menge={menge}
          onAdd={onAdd}
          onRemove={onRemove}
          addLabel="Variante hinzufügen"
          removeLabel="Variante entfernen"
          minusNurAbEins
        />
      </div>
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
            {Array.from({ length: 2 }).map((_, zeile) => (
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
