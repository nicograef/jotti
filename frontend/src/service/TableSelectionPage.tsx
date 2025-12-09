import { useAllTables } from './table/hooks'
import { TableList, TableListSkeleton } from './TableList'

export function TableSelectionPage() {
  const { loading, tables } = useAllTables()

  return (
    <>
      <h1 className="text-2xl font-bold">Tisch auswählen</h1>
      {loading ? <TableListSkeleton /> : <TableList tables={tables} />}
    </>
  )
}
