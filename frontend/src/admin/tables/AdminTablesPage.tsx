import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { toast } from 'sonner'

import { EntityStatus } from '@/lib/entityStatus'
import { formatAnzahl } from '@/lib/utils'

import { AdminPageHeader } from '../components/AdminPageHeader'
import { EditTischDialog } from './EditTischDialog'
import { ALLE_TISCHE_KEY, tischBackend, useAllTische } from './hooks'
import { NewTischDialog } from './NewTischDialog'
import type { Tisch } from './Tisch'
import { Tische } from './Tische'

export function AdminTablesPage() {
  const queryClient = useQueryClient()
  const { isPending, tische } = useAllTische()
  const [zuBearbeiten, setZuBearbeiten] = useState<Tisch | null>(null)

  const invalidateTische = () =>
    void queryClient.invalidateQueries({ queryKey: [ALLE_TISCHE_KEY] })

  const aktiveAnzahl = tische.filter(
    (t) => t.status === EntityStatus.ACTIVE,
  ).length
  const mitSaldoAnzahl = tische.filter((t) => t.saldoCents > 0).length
  const unterzeile = [
    formatAnzahl(tische.length, 'Tisch', 'Tische'),
    `${String(aktiveAnzahl)} aktiv`,
    mitSaldoAnzahl > 0
      ? `${String(mitSaldoAnzahl)} mit offenem Saldo`
      : 'Tische mit offenem Saldo lassen sich nicht deaktivieren',
  ].join(' · ')

  return (
    <>
      {zuBearbeiten && (
        <EditTischDialog
          backend={tischBackend}
          tisch={zuBearbeiten}
          updated={() => {
            invalidateTische()
          }}
          deleted={() => {
            invalidateTische()
            toast.success('Tisch wurde gelöscht.')
          }}
          close={() => {
            setZuBearbeiten(null)
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
        onEdit={(tisch) => {
          setZuBearbeiten(tisch)
        }}
        onStatusChange={() => {
          invalidateTische()
        }}
      />
    </>
  )
}
