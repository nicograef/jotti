import { BackendSingleton } from '@/lib/Backend'
import { TableBackend } from '@/table/TableBackend'

import { TableList } from './TableList'

const tableBackend = new TableBackend(BackendSingleton)

export function TableSelectionPage() {
  return (
    <>
      <h1 className="text-2xl font-bold">Tisch auswählen</h1>
      <TableList tableBackend={tableBackend} />
    </>
  )
}
