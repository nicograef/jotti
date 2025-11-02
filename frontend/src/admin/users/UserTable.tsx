import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import type { User } from "./User"
import { LockKeyhole } from "lucide-react"

const users: User[] = [
  {
    id: 1,
    username: "nicograef",
    createdAt: new Date().toLocaleString(),
    name: "Nico Gräf",
    role: "admin",
    gesperrt: false,
  },
  {
    id: 2,
    username: "lucasfi",
    createdAt: new Date().toLocaleString(),
    name: "Lucas Finke",
    role: "service",
    gesperrt: true,
  },
  {
    id: 3,
    username: "silviafi",
    createdAt: new Date().toLocaleString(),
    name: "Silvia Finke",
    role: "service",
    gesperrt: false,
  },
]

export function UserTable() {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>{/* Gesperrt */}</TableHead>
          <TableHead>Name</TableHead>
          <TableHead>Benutzername</TableHead>
          <TableHead>Rolle</TableHead>
          <TableHead>Erstellungsdatum</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {users.map((user) => (
          <TableRow key={user.id} className="cursor-pointer">
            <TableCell>
              {user.gesperrt ? <LockKeyhole size="16" /> : <></>}
            </TableCell>
            <TableCell className="font-medium">{user.name}</TableCell>
            <TableCell>{user.username}</TableCell>
            <TableCell>{user.role}</TableCell>
            <TableCell>{user.createdAt}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
