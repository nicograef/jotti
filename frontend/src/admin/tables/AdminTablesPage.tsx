import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { toast } from 'sonner'

import { BackendSingleton } from '@/lib/Backend'

import { EditTischDialog } from './EditTischDialog'
import { ALLE_TISCHE_KEY, useAllTische } from './hooks'
import { NewTischDialog } from './NewTischDialog'
import type { Tisch } from './Tisch'
import { TischBackend } from './TischBackend'
import { Tische } from './Tische'

const initialEditState = {
  tisch: null as Tisch | null,
  open: false,
}

const tischBackend = new TischBackend(BackendSingleton)

export function AdminTablesPage() {
  const queryClient = useQueryClient()
  const { isPending, tische } = useAllTische()
  const [editState, setEditState] = useState(initialEditState)

  const invalidateTische = () =>
    void queryClient.invalidateQueries({ queryKey: [ALLE_TISCHE_KEY] })

  return (
    <>
      <NewTischDialog
        backend={tischBackend}
        created={(tisch) => {
          invalidateTische()
          toast.success(`Tisch "${tisch.name}" wurde angelegt.`)
        }}
      />
      {editState.tisch && (
        <EditTischDialog
          backend={tischBackend}
          open={editState.open}
          tisch={editState.tisch}
          updated={() => {
            invalidateTische()
          }}
          close={() => {
            setEditState(initialEditState)
          }}
        />
      )}
      <h1 className="text-2xl font-bold">Tische verwalten</h1>
      <Tische
        loading={isPending}
        backend={tischBackend}
        tische={tische}
        onEdit={(tischId) => {
          const tischToEdit = tische.find((t) => t.id === tischId) ?? null
          setEditState({ tisch: tischToEdit, open: true })
        }}
        onStatusChange={() => {
          invalidateTische()
        }}
        onDeleted={() => {
          invalidateTische()
          toast.success('Tisch wurde gelöscht.')
        }}
      />
    </>
  )
}
