import { Plus } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

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
  DialogTrigger,
} from '@/components/ui/dialog'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { KATEGORIE_LABEL } from '@/lib/produktSchemas'

import type { ProduktAuswahl } from './ZusatzdruckerBackend'

// Antippen ordnet zu; ein Produkt dieses Druckers ist gesperrt.
export function ProduktAuswahlDialog({
  zusatzdruckerName,
  zugeordneteIds,
  produktAuswahl,
  zuordnen,
}: {
  zusatzdruckerName: string
  zugeordneteIds: ReadonlySet<number>
  produktAuswahl: ProduktAuswahl[]
  zuordnen: (produktId: number) => Promise<void>
}) {
  const [open, setOpen] = useState(false)
  const { loading, run } = useActionSubmit({
    actionLabel: 'Produkt zuordnen',
  })

  const waehle = (produkt: ProduktAuswahl) =>
    run(async () => {
      await zuordnen(produkt.id)
      toast.success(`„${produkt.name}“ druckt jetzt an „${zusatzdruckerName}“.`)
      setOpen(false)
    })

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm" className="w-fit">
          <Plus /> Produkt hinzufügen
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader className="mb-4">
          <DialogTitle>Produkt hinzufügen</DialogTitle>
          <DialogDescription>
            Das Produkt druckt danach an „{zusatzdruckerName}“.
          </DialogDescription>
        </DialogHeader>
        <DialogBody>
          {produktAuswahl.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              Noch keine Produkte angelegt.
            </p>
          ) : (
            <ul className="flex flex-col gap-1">
              {produktAuswahl.map((produkt) => {
                const hier = zugeordneteIds.has(produkt.id)
                return (
                  <li key={produkt.id}>
                    <Button
                      variant="ghost"
                      className="h-auto w-full justify-between gap-3 py-2 text-left"
                      disabled={loading || hier}
                      onClick={() => void waehle(produkt)}
                    >
                      <span className="flex flex-col">
                        <span className="font-medium">{produkt.name}</span>
                        <span className="text-xs text-muted-foreground">
                          {KATEGORIE_LABEL[produkt.kategorie]}
                        </span>
                      </span>
                      {produkt.zusatzdruckerName !== '' && (
                        <span className="text-xs text-muted-foreground">
                          {produkt.zusatzdruckerName}
                        </span>
                      )}
                    </Button>
                  </li>
                )
              })}
            </ul>
          )}
        </DialogBody>
        <DialogFooter className="mt-4">
          <DialogClose asChild>
            <Button variant="outline" disabled={loading}>
              Abbrechen
            </Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
