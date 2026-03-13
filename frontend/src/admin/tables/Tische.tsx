import { LayoutGrid } from 'lucide-react'
import { useState } from 'react'

import { EmptyState } from '@/components/common/EmptyState'
import { ItemGroup } from '@/components/ui/item'

import { type Tisch, TischStatus } from './Tisch'
import type { TischBackend } from './TischBackend'
import { TischItem } from './TischItem'

interface TischeProps {
  loading: boolean
  backend: Pick<
    TischBackend,
    'activateTisch' | 'deactivateTisch' | 'deleteTisch'
  >
  tische: Tisch[]
  onEdit: (tischId: number) => void
  onStatusChange: (tischId: number, status: TischStatus) => void
  onDeleted: (tischId: number) => void
}

export function Tische(props: TischeProps) {
  const [loading, setLoading] = useState(props.loading)

  const activateTisch = async (tischId: number) => {
    setLoading(true)
    try {
      await props.backend.activateTisch(tischId)
      props.onStatusChange(tischId, TischStatus.ACTIVE)
    } catch (error) {
      console.error('Error activating table:', error)
    }
    setLoading(false)
  }

  const deactivateTisch = async (tischId: number) => {
    setLoading(true)
    try {
      await props.backend.deactivateTisch(tischId)
      props.onStatusChange(tischId, TischStatus.INACTIVE)
    } catch (error) {
      console.error('Error deactivating table:', error)
    }
    setLoading(false)
  }

  const deleteTisch = async (tischId: number) => {
    setLoading(true)
    try {
      await props.backend.deleteTisch(tischId)
      props.onDeleted(tischId)
    } catch (error) {
      console.error('Error deleting table:', error)
    }
    setLoading(false)
  }

  if (props.tische.length === 0 && !props.loading) {
    return (
      <EmptyState
        icon={LayoutGrid}
        title="Keine Tische vorhanden"
        description="Erstelle einen neuen Tisch, um Bestellungen aufnehmen zu können."
      />
    )
  }

  return (
    <>
      <ItemGroup className="grid gap-4 lg:grid-cols-2 2xl:grid-cols-3 my-4">
        {props.tische.map((tisch) => (
          <TischItem
            key={tisch.id}
            loading={loading || props.loading}
            tisch={tisch}
            onActivate={activateTisch}
            onDeactivate={deactivateTisch}
            onEdit={props.onEdit}
            onDelete={deleteTisch}
          />
        ))}
      </ItemGroup>
    </>
  )
}
