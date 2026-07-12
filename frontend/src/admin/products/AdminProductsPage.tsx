import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { toast } from 'sonner'

import { BackendSingleton } from '@/lib/Backend'

import { AdminPageHeader } from '../components/AdminPageHeader'
import { useDruckstationen } from '../settings/hooks'
import { EditProductDialog } from './EditProductDialog'
import { ALLE_PRODUKTE_KEY, useAllProdukte } from './hooks'
import { NewProductDialog } from './NewProductDialog'
import { produktUnterzeile } from './productGrouping'
import { Products } from './Products'
import type { Produkt } from './Produkt'
import { ProduktBackend } from './ProduktBackend'

const initialProduktEditState = {
  produkt: null as Produkt | null,
  open: false,
}

const produktBackend = new ProduktBackend(BackendSingleton)

export function AdminProductsPage() {
  const queryClient = useQueryClient()
  const { isPending, produkte } = useAllProdukte()
  const { druckstationen } = useDruckstationen()
  const [produktEditState, setProduktEditState] = useState(
    initialProduktEditState,
  )

  const invalidateProdukte = () =>
    void queryClient.invalidateQueries({ queryKey: [ALLE_PRODUKTE_KEY] })

  const onProduktDelete = async (produktId: number) => {
    await produktBackend.deleteProdukt(produktId)
    invalidateProdukte()
    toast.success('Produkt wurde gelöscht.')
  }

  return (
    <>
      {produktEditState.produkt && (
        <EditProductDialog
          backend={produktBackend}
          open={produktEditState.open}
          product={produktEditState.produkt}
          updated={() => {
            invalidateProdukte()
          }}
          close={() => {
            setProduktEditState(initialProduktEditState)
          }}
        />
      )}
      <AdminPageHeader
        titel="Produkte & Preise"
        unterzeile={produktUnterzeile(produkte)}
        aktionen={
          <NewProductDialog
            backend={produktBackend}
            created={(produkt) => {
              invalidateProdukte()
              toast.success(`Produkt "${produkt.name}" wurde angelegt.`)
            }}
          />
        }
      />
      <Products
        loading={isPending}
        backend={produktBackend}
        products={produkte}
        druckstationen={druckstationen}
        onEdit={(produktId) => {
          const produktToEdit = produkte.find((p) => p.id === produktId) ?? null
          setProduktEditState({ produkt: produktToEdit, open: true })
        }}
        onDelete={onProduktDelete}
        onVariantCreated={(_produktId, variante) => {
          invalidateProdukte()
          toast.success(`Variante "${variante.name}" wurde angelegt.`)
        }}
        onVariantUpdated={() => {
          invalidateProdukte()
        }}
        onVariantStatusChange={() => {
          invalidateProdukte()
        }}
        onVariantDeleted={() => {
          invalidateProdukte()
          toast.success('Variante wurde gelöscht.')
        }}
      />
    </>
  )
}
