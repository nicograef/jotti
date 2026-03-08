import { useState } from 'react'
import { toast } from 'sonner'

import { BackendSingleton } from '@/lib/Backend'

import { EditTischDialog } from './EditTischDialog'
import { useAllTische } from './hooks'
import { NewTischDialog } from './NewTischDialog'
import type { Tisch, TischStatus } from './Tisch'
import { TischBackend } from './TischBackend'
import { Tische } from './Tische'

const initialEditState = {
  tisch: null as Tisch | null,
  open: false,
}

const tischBackend = new TischBackend(BackendSingleton)

export function AdminTablesPage() {
  const { loading, tische, setTische } = useAllTische()
  const [editState, setEditState] = useState(initialEditState)

  const updateTisch = (tisch: Tisch) => {
    setTische((prevTische) =>
      prevTische.map((t) => (t.id === tisch.id ? tisch : t)),
    )
  }

  const onStatusChange = (tischId: number, status: TischStatus) => {
    setTische((prevTische) =>
      prevTische.map((t) => (t.id === tischId ? { ...t, status } : t)),
    )
  }

  return (
    <>
      <NewTischDialog
        backend={tischBackend}
        created={(tisch) => {
          setTische((prevTische) => [...prevTische, tisch])
          toast.success(`Tisch "${tisch.name}" wurde angelegt.`)
        }}
      />
      {editState.tisch && (
        <EditTischDialog
          backend={tischBackend}
          open={editState.open}
          tisch={editState.tisch}
          updated={(tisch) => {
            updateTisch(tisch)
          }}
          close={() => {
            setEditState(initialEditState)
          }}
        />
      )}
      <h1 className="text-2xl font-bold">Tische verwalten</h1>
      <Tische
        loading={loading}
        backend={tischBackend}
        tische={tische}
        onEdit={(tischId) => {
          const tischToEdit = tische.find((t) => t.id === tischId) ?? null
          setEditState({ tisch: tischToEdit, open: true })
        }}
        onStatusChange={onStatusChange}
      />
    </>
  )
}
