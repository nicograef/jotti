import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogBody,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Field, FieldLabel } from '@/components/ui/field'

import type { User } from './User'

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
          <DialogTitle>Helfer wurde angelegt!</DialogTitle>
          <DialogDescription>
            Für {props.user?.name} wurde ein {props.user?.role}-Helfer angelegt.
            Beim erstmaligen Anmelden muss der Helfer mit dem untenstehenden
            Code sein Passwort setzen.
          </DialogDescription>
        </DialogHeader>
        <DialogBody className="flex flex-col gap-6">
          <Field className="gap-1">
            <FieldLabel>Benutzername</FieldLabel>
            <p className="text-3xl">{props.user?.username}</p>
          </Field>
          <Field className="gap-1">
            <FieldLabel>Code</FieldLabel>
            <p
              data-testid="onetime-password"
              className="text-3xl tracking-widest"
            >
              {props.onetimePassword}
            </p>
          </Field>
        </DialogBody>
        <DialogFooter className="mt-4">
          <DialogClose asChild>
            <Button>Okay</Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
