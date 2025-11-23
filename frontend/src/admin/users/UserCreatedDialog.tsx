import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Field, FieldLabel } from '@/components/ui/field'
import type { User } from '@/user/User'

interface UserCreatedDialogProps {
  user: User | null
  onetimePassword: string
  open: boolean
  close: () => void
}

export function UserCreatedDialog(props: UserCreatedDialogProps) {
  const onOpenChange = (isOpen: boolean) => {
    if (!isOpen) props.close()
  }

  return (
    <Dialog open={props.open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Benutzer wurde angelegt!</DialogTitle>
          <DialogDescription>
            Für {props.user?.name} wurde ein {props.user?.role}-Benutzer
            angelegt. Beim erstmaligen Anmelden muss der Benutzer mit dem
            untenstehenden Code sein Passwort setzen.
          </DialogDescription>
        </DialogHeader>
        <Field className="gap-1">
          <FieldLabel htmlFor="username">Benutzername</FieldLabel>
          <p className="text-3xl">{props.user?.username}</p>
        </Field>
        <Field className="gap-1">
          <FieldLabel htmlFor="onetimePassword">Code</FieldLabel>
          <p className="text-3xl tracking-widest">{props.onetimePassword}</p>
        </Field>
        <DialogFooter className="mt-4">
          <DialogClose asChild>
            <Button>Okay</Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
