import { useState } from 'react'
import { toast } from 'sonner'

import { BackendSingleton } from '@/lib/Backend'

import { EditProductDialog } from './EditProductDialog'
import { useAllProdukte } from './hooks'
import { NewProductDialog } from './NewProductDialog'
import { Products } from './Products'
import type { Produkt, Variante } from './Produkt'
import { ProduktBackend } from './ProduktBackend'

const initialProduktEditState = {
  produkt: null as Produkt | null,
  open: false,
}

const produktBackend = new ProduktBackend(BackendSingleton)

export function AdminProductsPage() {
  const { loading, produkte, setProdukte } = useAllProdukte()
  const [produktEditState, setProduktEditState] = useState(
    initialProduktEditState,
  )

  const updateProdukt = (produkt: Produkt) => {
    setProdukte((prevProdukte) =>
      prevProdukte.map((p) => (p.id === produkt.id ? produkt : p)),
    )
  }

  const updateVarianteInProdukt = (
    produktId: number,
    updater: (varianten: Variante[]) => Variante[],
  ) => {
    setProdukte((prevProdukte) =>
      prevProdukte.map((p) =>
        p.id === produktId ? { ...p, varianten: updater(p.varianten) } : p,
      ),
    )
  }

  const onVarianteCreated = (produktId: number, variante: Variante) => {
    updateVarianteInProdukt(produktId, (varianten) => [...varianten, variante])
    toast.success(`Variante "${variante.name}" wurde angelegt.`)
  }

  const onVarianteUpdated = (produktId: number, variante: Variante) => {
    updateVarianteInProdukt(produktId, (varianten) =>
      varianten.map((v) => (v.id === variante.id ? variante : v)),
    )
  }

  const onVarianteStatusChange = (
    produktId: number,
    varianteId: number,
    status: 'active' | 'inactive',
  ) => {
    updateVarianteInProdukt(produktId, (varianten) =>
      varianten.map((v) => (v.id === varianteId ? { ...v, status } : v)),
    )
  }

  const onProduktDelete = async (produktId: number) => {
    await produktBackend.deleteProdukt(produktId)
    setProdukte((prev) => prev.filter((p) => p.id !== produktId))
    toast.success('Produkt wurde gelöscht.')
  }

  const onVarianteDeleted = (produktId: number, varianteId: number) => {
    updateVarianteInProdukt(produktId, (varianten) =>
      varianten.filter((v) => v.id !== varianteId),
    )
    toast.success('Variante wurde gelöscht.')
  }

  return (
    <>
      <NewProductDialog
        backend={produktBackend}
        created={(produkt) => {
          setProdukte((prevProdukte) => [...prevProdukte, produkt])
          toast.success(`Produkt "${produkt.name}" wurde angelegt.`)
        }}
      />
      {produktEditState.produkt && (
        <EditProductDialog
          backend={produktBackend}
          open={produktEditState.open}
          product={produktEditState.produkt}
          updated={(produkt) => {
            updateProdukt(produkt)
          }}
          close={() => {
            setProduktEditState(initialProduktEditState)
          }}
        />
      )}
      <h1 className="text-2xl font-bold">Produkte verwalten</h1>
      <Products
        loading={loading}
        backend={produktBackend}
        products={produkte}
        onEdit={(produktId) => {
          const produktToEdit = produkte.find((p) => p.id === produktId) ?? null
          setProduktEditState({ produkt: produktToEdit, open: true })
        }}
        onDelete={onProduktDelete}
        onVariantCreated={onVarianteCreated}
        onVariantUpdated={onVarianteUpdated}
        onVariantDeleted={onVarianteDeleted}
        onVariantStatusChange={onVarianteStatusChange}
      />
    </>
  )
}
