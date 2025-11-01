import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

interface User {
  userId: string
  username: string
  createdAt: string
  name: string
}

const users: User[] = [
  {
    userId: "1",
    username: "nicogr",
    createdAt: new Date().toLocaleString(),
    name: "Nico Gräf",
  },
  {
    userId: "2",
    username: "lucasfi",
    createdAt: new Date().toLocaleString(),
    name: "Lucas Finke",
  },
  {
    userId: "3",
    username: "silviafi",
    createdAt: new Date().toLocaleString(),
    name: "Silvia Finke",
  },
]

export function UserTable() {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Benutzername</TableHead>
          <TableHead>Erstellungsdatum</TableHead>
          <TableHead className="text-right">Aktionen</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {users.map((user) => (
          <TableRow key={user.userId}>
            <TableCell className="font-medium">{user.name}</TableCell>
            <TableCell>{user.username}</TableCell>
            <TableCell>{user.createdAt}</TableCell>
            <TableCell className="text-right">
              <button className="text-red-500 hover:text-red-700 cursor-pointer">Entfernen</button>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
