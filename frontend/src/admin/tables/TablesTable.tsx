import { LockKeyhole } from 'lucide-react'

import { Skeleton } from '@/components/ui/skeleton'
import {
  Table as TableComponent,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { Table } from '@/lib/TableBackend'

function TableTableRow(props: {
  table: Table
  onClick: (table: Table) => void
}) {
  const { table, onClick } = props
  return (
    <TableRow
      className="cursor-pointer"
      onClick={() => {
        onClick(table)
      }}
    >
      <TableCell>{table.locked ? <LockKeyhole size="30" /> : <></>}</TableCell>
      <TableCell className="font-medium">{table.name}</TableCell>
      <TableCell className="text-right">
        {new Date(table.createdAt).toLocaleString()} Uhr
      </TableCell>
    </TableRow>
  )
}

function UserTableRowSkeleton() {
  return (
    <TableRow className="animate-pulse">
      <TableCell>
        <></>
      </TableCell>
      <TableCell>
        <Skeleton className="h-4 w-24" />
      </TableCell>
      <TableCell>
        <Skeleton className="h-4 w-20" />
      </TableCell>
    </TableRow>
  )
}

interface TablesTableProps {
  tables: Table[]
  loading: boolean
  onClick: (table: Table) => void
}

export function TablesTable(props: Readonly<TablesTableProps>) {
  return (
    <TableComponent className="text-lg">
      <TableHeader className="h-18 bg-muted">
        <TableRow>
          <TableHead className="w-[50px]">{/* Gesperrt */}</TableHead>
          <TableHead>Tischname</TableHead>
          <TableHead className="text-right">Erstellungsdatum</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {props.loading
          ? Array.from({ length: 5 }).map((_, index) => (
              // eslint-disable-next-line react-x/no-array-index-key
              <UserTableRowSkeleton key={index} />
            ))
          : props.tables.map((table) => (
              <TableTableRow
                key={table.id}
                table={table}
                onClick={props.onClick}
              />
            ))}
      </TableBody>
    </TableComponent>
  )
}
