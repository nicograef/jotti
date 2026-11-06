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

export interface Einmalcode {
  titel: string
  beschreibung: string
  username: string
  onetimePassword: string
}

// Zeigt den Code, mit dem der Helfer beim nächsten Anmelden sein Passwort setzt.
export function EinmalcodeDialog({
  einmalcode,
  close,
}: {
  einmalcode: Einmalcode
  close: () => void
}) {
  return (
    <Dialog
      open
      onOpenChange={(isOpen) => {
        if (!isOpen) close()
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{einmalcode.titel}</DialogTitle>
          <DialogDescription>{einmalcode.beschreibung}</DialogDescription>
        </DialogHeader>
        <DialogBody className="flex flex-col gap-6">
          <Field className="gap-1">
            <FieldLabel>Benutzername</FieldLabel>
            <p className="text-3xl">{einmalcode.username}</p>
          </Field>
          <Field className="gap-1">
            <FieldLabel>Code</FieldLabel>
            <p
              data-testid="onetime-password"
              className="text-3xl tracking-widest"
            >
              {einmalcode.onetimePassword}
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
