import { useState } from 'react'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Spinner } from '@/components/ui/spinner'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { formatCents, parseCents } from '@/lib/utils'

import type { DirektverkaufBackend } from '../../direktverkauf/DirektverkaufBackend'
import type { Produkt } from '../../product/Produkt'
import { KommentarField } from '../table/CommentField'
import {
  calculateTotalPrice,
  calculateZahlungsbetraege,
} from '../table/drawerUtils'
import { ProductList, ProductListSkeleton } from '../table/ProductList'

interface DirektverkaufProps {
  backend: Pick<DirektverkaufBackend, 'direktverkaufTaetigen'>
  products: Produkt[]
  productsLoading: boolean
  onVerkauft?: () => void
}

interface SelectedItem {
  produktId: number
  varianteId: number
  einzelpreis: number
  menge: number
}

function selectItems(
  products: Produkt[],
  mengen: Record<number, number>,
): SelectedItem[] {
  return products.flatMap((product) =>
    product.varianten
      .filter((variant) => (mengen[variant.id] || 0) > 0)
      .map((variant) => ({
        produktId: product.id,
        varianteId: variant.id,
        einzelpreis: variant.preisCents,
        menge: mengen[variant.id],
      })),
  )
}

export function Direktverkauf({
  backend,
  products,
  productsLoading,
  onVerkauft,
}: DirektverkaufProps) {
  const [mengen, setMengen] = useState<Record<number, number>>({})
  const [erhaltenEuro, setErhaltenEuro] = useState('')
  const [kommentar, setKommentar] = useState('')

  const items = selectItems(products, mengen)
  const total = calculateTotalPrice(items)
  const { rueckgeldCents } = calculateZahlungsbetraege(
    total,
    parseCents(erhaltenEuro),
    0,
  )
  const noPositionenSelected = items.length === 0

  const { loading, run } = useActionSubmit({
    actionLabel: 'Verkauf abschließen',
    byCode: {
      kasse_nicht_geoeffnet:
        'Es ist keine Kassensitzung geöffnet. Bitte zuerst die Kasse öffnen.',
      produkt_not_found:
        'Ein ausgewähltes Produkt ist nicht mehr verfügbar. Bitte Auswahl aktualisieren.',
    },
    onSuccess: () => {
      setMengen({})
      setErhaltenEuro('')
      setKommentar('')
      toast.success('Verkauf abgeschlossen.')
      onVerkauft?.()
    },
  })

  const onSubmit = async () => {
    await run(async () => {
      await backend.direktverkaufTaetigen({
        positionen: items.map((item) => ({
          produktId: item.produktId,
          varianteId: item.varianteId,
          menge: item.menge,
        })),
        kommentar,
      })
    })
  }

  if (productsLoading) {
    return <ProductListSkeleton />
  }

  return (
    <div className="space-y-4">
      <Card className="p-4 space-y-3 sticky top-14 z-30">
        <div className="flex justify-between text-lg font-semibold">
          <span>Gesamt</span>
          <span>{formatCents(total)}&nbsp;€</span>
        </div>
        <div className="flex items-center justify-between gap-3">
          <Label htmlFor="erhalten">Erhalten</Label>
          <div className="flex items-center gap-1.5">
            <Input
              id="erhalten"
              inputMode="decimal"
              placeholder="0,00"
              value={erhaltenEuro}
              onChange={(e) => {
                setErhaltenEuro(e.target.value)
              }}
              className="w-24 text-right"
              spellCheck={false}
            />
            <span>€</span>
          </div>
        </div>
        {rueckgeldCents !== null && (
          <div className="flex justify-between font-medium">
            <span>Rückgeld</span>
            <span>{formatCents(rueckgeldCents)}&nbsp;€</span>
          </div>
        )}
        <KommentarField
          onChange={(value) => {
            setKommentar(value)
          }}
        />
        <Button
          disabled={noPositionenSelected || loading}
          onClick={() => {
            void onSubmit()
          }}
          className="w-full"
        >
          {loading ? <Spinner /> : null} Verkauf abschließen
        </Button>
      </Card>
      <ProductList
        products={products}
        variantMengen={mengen}
        onAdd={(variantId) => {
          setMengen((prev) => ({
            ...prev,
            [variantId]: (prev[variantId] || 0) + 1,
          }))
        }}
        onRemove={(variantId) => {
          setMengen((prev) => {
            const aktuelleMenge = prev[variantId] || 0
            if (aktuelleMenge <= 0) return prev
            return {
              ...prev,
              [variantId]: aktuelleMenge - 1,
            }
          })
        }}
      />
    </div>
  )
}
