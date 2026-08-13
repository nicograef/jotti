import { ChevronLeft, Package } from 'lucide-react'
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

// gewaehlteMenge summiert die Auswahl über alle Varianten eines Produkts. Auf
// der Produktebene ist das die einzige Rückmeldung darüber, was schon im Korb
// liegt — ohne sie verschwindet die Auswahl beim Zurückgehen aus dem Blick.
function gewaehlteMenge(
  product: Produkt,
  mengen: Record<number, number>,
): number {
  return product.varianten.reduce(
    (summe, variante) => summe + (mengen[variante.id] || 0),
    0,
  )
}

export function ProductList(props: ProductListComponentProps) {
  const kategorien = belegteKategorien(props.products)
  const [aktiveKategorie, setAktiveKategorie] = useState<Kategorie | undefined>(
    kategorien[0],
  )
  const [aktivesProduktId, setAktivesProduktId] = useState<number | undefined>()

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
  // Kein expliziter Reset beim Kategoriewechsel nötig: gehört das gewählte
  // Produkt nicht zur sichtbaren Kategorie, greift die Suche ins Leere und die
  // Ansicht fällt von selbst auf die Produktebene zurück.
  const aktivesProdukt = sichtbareProdukte.find(
    (p) => p.id === aktivesProduktId,
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

      {aktivesProdukt ? (
        <div className="mt-4">
          <div className="mb-1 flex items-center gap-1">
            <button
              type="button"
              className="-ml-1 flex items-center gap-0.5 rounded-md py-1 pl-1 pr-2 text-[13px] font-semibold uppercase tracking-wide text-muted-foreground"
              onClick={() => {
                setAktivesProduktId(undefined)
              }}
            >
              <ChevronLeft className="size-4" />
              Produkte
            </button>
            <span className="text-[13px] font-semibold uppercase tracking-wide">
              {aktivesProdukt.name}
            </span>
          </div>
          <div>
            {aktivesProdukt.varianten.map((variant) => (
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
      ) : (
        <div className="mt-4 grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-4">
          {sichtbareProdukte.map((product) => (
            <ProductTile
              key={product.id}
              product={product}
              gewaehlt={gewaehlteMenge(product, props.variantMengen)}
              onOpen={() => {
                setAktivesProduktId(product.id)
              }}
            />
          ))}
        </div>
      )}
    </div>
  )
}

// ProductTile ist die Produktebene: ein Produkt pro Kachel, darunter die Anzahl
// seiner Varianten. Liegt schon etwas im Korb, zeigt die Kachel die Summe, damit
// die Auswahl auf dieser Ebene nicht unsichtbar wird.
function ProductTile({
  product,
  gewaehlt,
  onOpen,
}: {
  product: Produkt
  gewaehlt: number
  onOpen: () => void
}) {
  const anzahl = product.varianten.length

  return (
    <button
      type="button"
      onClick={onOpen}
      className={cn(
        'flex h-full min-h-20 flex-col rounded-lg border bg-card p-3 text-left transition-transform duration-100 ease-linear active:scale-[.98]',
        gewaehlt > 0 && 'border-primary/50 bg-primary/[0.04]',
      )}
    >
      <span className="break-words text-[15px] font-medium leading-snug">
        {product.name}
      </span>
      <span className="mt-auto pt-2 text-sm text-muted-foreground">
        {/* Komma vor der Zahl: sonst zieht der Accessibility-Name von Name und
            Anzahl zu "Bratwurst1 Variante" zusammen. */}
        {', '}
        {anzahl} {anzahl === 1 ? 'Variante' : 'Varianten'}
        {gewaehlt > 0 && (
          <span className="ml-2 font-semibold text-foreground tabular-nums">
            {gewaehlt} gewählt
          </span>
        )}
      </span>
    </button>
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
    <div className="mt-4 grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-4">
      {Array.from({ length: 4 }).map((_, kachel) => (
        <div
          key={`skeleton-kachel-${kachel.toString()}`}
          className="min-h-20 rounded-lg border bg-card p-3"
        >
          <Skeleton className="h-5 w-24" />
          <Skeleton className="mt-4 h-4 w-16" />
        </div>
      ))}
    </div>
  )
}
