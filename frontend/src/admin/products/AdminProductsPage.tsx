import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { toast } from 'sonner'

import type { Produkt } from '@/lib/produktSchemas'

import { AdminPageHeader } from '../components/AdminPageHeader'
import { useDruckstationen } from '../settings/hooks'
import { EditProductDialog } from './EditProductDialog'
import { ALLE_PRODUKTE_KEY, produktBackend, useAllProdukte } from './hooks'
import { NewProductDialog } from './NewProductDialog'
import { produktUnterzeile } from './productGrouping'
import { Products } from './Products'

export function AdminProductsPage() {
  const queryClient = useQueryClient()
  const { isPending, produkte } = useAllProdukte()
  const { druckstationen } = useDruckstationen()
  const [zuBearbeiten, setZuBearbeiten] = useState<Produkt | null>(null)

  const invalidateProdukte = () =>
    void queryClient.invalidateQueries({ queryKey: [ALLE_PRODUKTE_KEY] })

  const onProduktDelete = async (produktId: number) => {
    await produktBackend.deleteProdukt(produktId)
    invalidateProdukte()
    toast.success('Produkt wurde gelöscht.')
  }

  return (
    <>
      {zuBearbeiten && (
        <EditProductDialog
          product={zuBearbeiten}
          updated={() => {
            invalidateProdukte()
          }}
          close={() => {
            setZuBearbeiten(null)
          }}
        />
      )}
      <AdminPageHeader
        titel="Produkte & Preise"
        unterzeile={produktUnterzeile(produkte)}
        glowFarben={['green', 'blue']}
        aktionen={
          <NewProductDialog
            created={(produkt) => {
              invalidateProdukte()
              toast.success(`Produkt "${produkt.name}" wurde angelegt.`)
            }}
          />
        }
      />
      <Products
        loading={isPending}
        products={produkte}
        druckstationen={druckstationen}
        onEdit={(produkt) => {
          setZuBearbeiten(produkt)
        }}
        onDelete={onProduktDelete}
        onChanged={invalidateProdukte}
        onVariantCreated={(variante) => {
          invalidateProdukte()
          toast.success(`Variante "${variante.name}" wurde angelegt.`)
        }}
        onVariantDeleted={() => {
          invalidateProdukte()
          toast.success('Variante wurde gelöscht.')
        }}
      />
    </>
  )
}
