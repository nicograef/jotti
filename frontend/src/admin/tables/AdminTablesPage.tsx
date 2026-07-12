import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { toast } from 'sonner'

import { BackendSingleton } from '@/lib/Backend'

import { AdminPageHeader } from '../components/AdminPageHeader'
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

  const aktiveAnzahl = tische.filter((t) => t.status === 'active').length
  const mitSaldoAnzahl = tische.filter((t) => t.saldoCents > 0).length
  const unterzeile = [
    `${String(tische.length)} ${tische.length === 1 ? 'Tisch' : 'Tische'}`,
    `${String(aktiveAnzahl)} aktiv`,
    mitSaldoAnzahl > 0
      ? `${String(mitSaldoAnzahl)} mit offenem Saldo`
      : 'Tische mit offenem Saldo lassen sich nicht deaktivieren',
  ].join(' · ')

  return (
    <>
      {editState.tisch && (
        <EditTischDialog
          backend={tischBackend}
          open={editState.open}
          tisch={editState.tisch}
          updated={() => {
            invalidateTische()
          }}
          deleted={() => {
            invalidateTische()
            toast.success('Tisch wurde gelöscht.')
          }}
          close={() => {
            setEditState(initialEditState)
          }}
        />
      )}
      <AdminPageHeader
        titel="Tische"
        unterzeile={unterzeile}
        aktionen={
          <NewTischDialog
            backend={tischBackend}
            created={(tisch) => {
              invalidateTische()
              toast.success(`Tisch "${tisch.name}" wurde angelegt.`)
            }}
          />
        }
      />
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
      />
    </>
  )
}
