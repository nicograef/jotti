import { LayoutGrid } from 'lucide-react'

import { EmptyState } from '@/components/common/EmptyState'
import { useActionSubmit } from '@/hooks/use-action-submit'

import { HinweisKarte } from '../components/HinweisKarte'
import { type Tisch } from './Tisch'
import type { TischBackend } from './TischBackend'
import { gruppiereTische } from './tischGrouping'
import { TischItem } from './TischItem'

interface TischeProps {
  loading: boolean
  backend: Pick<TischBackend, 'activateTisch' | 'deactivateTisch'>
  tische: Tisch[]
  onEdit: (tischId: number) => void
  onStatusChange: () => void
}

export function Tische(props: TischeProps) {
  const { loading: activateLoading, run: runActivate } = useActionSubmit({
    actionLabel: 'Tisch aktivieren',
  })
  const { loading: deactivateLoading, run: runDeactivate } = useActionSubmit({
    actionLabel: 'Tisch deaktivieren',
  })

  const loading = activateLoading || deactivateLoading

  const activateTisch = async (tischId: number) => {
    await runActivate(async () => {
      await props.backend.activateTisch(tischId)
      props.onStatusChange()
    })
  }

  const deactivateTisch = async (tischId: number) => {
    await runDeactivate(async () => {
      await props.backend.deactivateTisch(tischId)
      props.onStatusChange()
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

  const gruppen = gruppiereTische(props.tische)

  return (
    <div className="my-4 flex flex-col gap-6">
      {gruppen.map((gruppe) => (
        <div key={gruppe.name}>
          <h2 className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
            {gruppe.name}
          </h2>
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6">
            {gruppe.tische.map((tisch) => (
              <TischItem
                key={tisch.id}
                loading={loading || props.loading}
                tisch={tisch}
                onActivate={activateTisch}
                onDeactivate={deactivateTisch}
                onEdit={props.onEdit}
              />
            ))}
          </div>
        </div>
      ))}
      <HinweisKarte>
        Tische mit <strong className="text-foreground">offenem Saldo</strong>{' '}
        zeigen den Betrag an und sind gegen Deaktivieren und Löschen geschützt,
        bis abgerechnet wurde. Umbenennen und Löschen: Kachel anklicken.
      </HinweisKarte>
    </div>
  )
}
