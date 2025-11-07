import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { UserRole, type User } from "../../lib/user"
import { LockKeyhole, ShieldUser } from "lucide-react"
import { Skeleton } from "@/components/ui/skeleton"
import { Badge } from "@/components/ui/badge"

function UserTableRow(props: { user: User }) {
  const { user } = props
  return (
    <TableRow className="cursor-pointer">
      <TableCell>{user.locked ? <LockKeyhole size="16" /> : <></>}</TableCell>
      <TableCell className="font-medium">{user.name}</TableCell>
      <TableCell>{user.username}</TableCell>
      <TableCell>
        {user.role === UserRole.ADMIN ? (
          <Badge className="text-sm">
            <ShieldUser />
            {user.role}
          </Badge>
        ) : (
          <Badge className="text-sm" variant="secondary">{user.role}</Badge>
        )}
      </TableCell>
      <TableCell>{new Date(user.createdAt).toLocaleString()} Uhr</TableCell>
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

type UsersTableProps = {
  users: User[]
  loading: boolean
}

export function UserTable(props: Readonly<UsersTableProps>) {
  return (
    <Table className="text-lg">
      <TableHeader className="h-18 bg-muted">
        <TableRow>
          <TableHead>{/* Gesperrt */}</TableHead>
          <TableHead>Name</TableHead>
          <TableHead>Benutzername</TableHead>
          <TableHead>Rolle</TableHead>
          <TableHead>Erstellungsdatum</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {props.loading
          ? Array.from({ length: 5 }).map((_, index) => (
              <UserTableRowSkeleton key={index} />
            ))
          : props.users.map((user) => (
              <UserTableRow key={user.id} user={user} />
            ))}
      </TableBody>
    </Table>
  )
}
