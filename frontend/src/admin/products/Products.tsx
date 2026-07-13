import { Package } from 'lucide-react'

import { EmptyState } from '@/components/common/EmptyState'

import { HinweisKarte } from '../components/HinweisKarte'
import type { DruckstationConfig } from '../settings/DruckstationBackend'
import { groupProdukteByKategorie, kategorieZusatz } from './productGrouping'
import { ProductItem } from './ProductItem'
import { type Produkt, type Variante, VarianteStatus } from './Produkt'
import { type ProduktBackend } from './ProduktBackend'

interface ProductsProps {
  loading: boolean
  backend: Pick<
    ProduktBackend,
    | 'aktiviereVariante'
    | 'deaktiviereVariante'
    | 'createVariante'
    | 'updateVariante'
    | 'deleteVariante'
  >
  products: Produkt[]
  druckstationen: DruckstationConfig[]
  onEdit: (produktId: number) => void
  onDelete: (produktId: number) => Promise<void>
  onVariantCreated: (produktId: number, variante: Variante) => void
  onVariantUpdated: (produktId: number, variante: Variante) => void
  onVariantStatusChange: (
    produktId: number,
    varianteId: number,
    status: VarianteStatus,
  ) => void
  onVariantDeleted: () => void
}

export function Products(props: ProductsProps) {
  if (props.products.length === 0 && !props.loading) {
    return (
      <EmptyState
        icon={Package}
        title="Keine Produkte vorhanden"
        description="Erstelle ein neues Produkt und füge Varianten mit Preisen hinzu."
      />
    )
  }

  const gruppen = groupProdukteByKategorie(props.products)

  return (
    <div className="my-4 flex flex-col gap-8">
      {gruppen.map((gruppe) => {
        const zusatz = kategorieZusatz(
          gruppe.kategorie,
          gruppe.produkte,
          props.druckstationen,
        )
        return (
          <section key={gruppe.kategorie}>
            <h2 className="mb-2 text-sm font-semibold uppercase tracking-wide text-muted-foreground">
              {gruppe.label}
              {zusatz !== '' && (
                <span className="font-normal normal-case tracking-normal">
                  {' · '}
                  {zusatz}
                </span>
              )}
            </h2>
            <div className="rounded-lg border px-4">
              {gruppe.produkte.map((product) => (
                <ProductItem
                  key={product.id}
                  loading={props.loading}
                  product={product}
                  backend={props.backend}
                  onEdit={props.onEdit}
                  onDelete={props.onDelete}
                  onVariantCreated={(variant) => {
                    props.onVariantCreated(product.id, variant)
                  }}
                  onVariantUpdated={(variant) => {
                    props.onVariantUpdated(product.id, variant)
                  }}
                  onVariantStatusChange={(variantId, status) => {
                    props.onVariantStatusChange(product.id, variantId, status)
                  }}
                  onVariantDeleted={props.onVariantDeleted}
                />
              ))}
            </div>
          </section>
        )
      })}

      <HinweisKarte>
        <strong className="text-foreground">Ausverkauft?</strong> Schalter aus
        statt löschen — die Variante verschwindet sofort von den Service-Handys,
        bleibt aber in allen Berichten. Löschen ist im „···"-Menü möglich.
      </HinweisKarte>
    </div>
  )
}
