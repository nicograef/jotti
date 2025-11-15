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
import type { Table } from '@/lib/TableBackend'

interface TableCreatedDialogProps {
  table: Table | null
  open: boolean
  close: () => void
}

export function TableCreatedDialog(props: TableCreatedDialogProps) {
  const onOpenChange = (isOpen: boolean) => {
    if (!isOpen) props.close()
  }

  return (
    <Dialog open={props.open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Tisch wurde angelegt!</DialogTitle>
          <DialogDescription>
            Der Tisch {props.table?.name} wurde angelegt.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className="mt-4">
          <DialogClose asChild>
            <Button>ooookay</Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
