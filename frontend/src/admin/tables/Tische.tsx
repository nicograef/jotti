import { LayoutGrid } from 'lucide-react'

import { EmptyState } from '@/components/common/EmptyState'
import { ItemGroup } from '@/components/ui/item'
import { useActionSubmit } from '@/hooks/use-action-submit'

import { adminListBottomClearance } from '../adminListLayout'
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
  const { loading: activateLoading, run: runActivate } = useActionSubmit({
    actionLabel: 'Tisch aktivieren',
  })
  const { loading: deactivateLoading, run: runDeactivate } = useActionSubmit({
    actionLabel: 'Tisch deaktivieren',
  })
  const { loading: deleteLoading, run: runDelete } = useActionSubmit({
    actionLabel: 'Tisch löschen',
  })

  const loading = activateLoading || deactivateLoading || deleteLoading

  const activateTisch = async (tischId: number) => {
    await runActivate(async () => {
      await props.backend.activateTisch(tischId)
      props.onStatusChange(tischId, TischStatus.ACTIVE)
    })
  }

  const deactivateTisch = async (tischId: number) => {
    await runDeactivate(async () => {
      await props.backend.deactivateTisch(tischId)
      props.onStatusChange(tischId, TischStatus.INACTIVE)
    })
  }

  const deleteTisch = async (tischId: number) => {
    await runDelete(async () => {
      await props.backend.deleteTisch(tischId)
      props.onDeleted(tischId)
    })
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
    <ItemGroup
      className={`grid gap-4 lg:grid-cols-2 2xl:grid-cols-3 my-4 ${adminListBottomClearance}`}
    >
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
  )
}
