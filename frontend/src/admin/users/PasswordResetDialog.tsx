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

interface PasswordResetDialogProps {
  username: string | null
  onetimePassword: string
  open: boolean
  close: () => void
}

export function PasswordResetDialog(props: PasswordResetDialogProps) {
  const onOpenChange = (isOpen: boolean) => {
    if (!isOpen) props.close()
  }

  return (
    <Dialog open={props.open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Passwort zurückgesetzt!</DialogTitle>
          <DialogDescription>
            Das Passwort für {props.username} wurde zurückgesetzt. Beim nächsten
            Anmelden muss der Helfer mit dem untenstehenden Code sein neues
            Passwort setzen.
          </DialogDescription>
        </DialogHeader>
        <DialogBody className="flex flex-col gap-6">
          <Field className="gap-1">
            <FieldLabel>Benutzername</FieldLabel>
            <p className="text-3xl">{props.username}</p>
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
