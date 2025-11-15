import { LockKeyhole, ShieldUser } from 'lucide-react'

import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { type User, UserRole } from '@/lib/user'

function UserTableRow(props: { user: User; onClick: (user: User) => void }) {
  const { user, onClick } = props
  return (
    <TableRow
      className="cursor-pointer"
      onClick={() => {
        onClick(user)
      }}
    >
      <TableCell>{user.locked ? <LockKeyhole size="30" /> : <></>}</TableCell>
      <TableCell className="font-medium">{user.name}</TableCell>
      <TableCell>{user.username}</TableCell>
      <TableCell>
        {user.role === UserRole.ADMIN ? (
          <Badge className="text-sm">
            <ShieldUser />
            {user.role}
          </Badge>
        ) : (
          <Badge className="text-sm" variant="secondary">
            {user.role}
          </Badge>
        )}
      </TableCell>
      <TableCell className="text-right">
        {new Date(user.createdAt).toLocaleString()} Uhr
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
      <TableCell>
        <Skeleton className="h-4 w-16" />
      </TableCell>
      <TableCell>
        <Skeleton className="h-4 w-28" />
      </TableCell>
    </TableRow>
  )
}

interface UsersTableProps {
  users: User[]
  loading: boolean
  onClick: (user: User) => void
}

export function UserTable(props: Readonly<UsersTableProps>) {
  return (
    <Table className="text-lg">
      <TableHeader className="h-18 bg-muted">
        <TableRow>
          <TableHead className="w-[50px]">{/* Gesperrt */}</TableHead>
          <TableHead>Name</TableHead>
          <TableHead>Benutzername</TableHead>
          <TableHead>Rolle</TableHead>
          <TableHead className="text-right">Erstellungsdatum</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {props.loading
          ? Array.from({ length: 5 }).map((_, index) => (
              // eslint-disable-next-line react-x/no-array-index-key
              <UserTableRowSkeleton key={index} />
            ))
          : props.users.map((user) => (
              <UserTableRow key={user.id} user={user} onClick={props.onClick} />
            ))}
      </TableBody>
    </Table>
  )
}
