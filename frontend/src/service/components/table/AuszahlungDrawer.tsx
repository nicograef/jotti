import { useState } from 'react'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from '@/components/ui/drawer'
import { Field } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Spinner } from '@/components/ui/spinner'
import { Textarea } from '@/components/ui/textarea'
import { getActionErrorMessage } from '@/lib/errorMessages'
import { formatCents, parseCents } from '@/lib/utils'

import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'

interface AuszahlungDrawerProps {
  backend: Pick<TischBackend, 'auszahlungLeisten'>
  tisch: Tisch
  saldoCents: number
  auszahlungGeleistet: () => void
}

export function AuszahlungDrawer(props: AuszahlungDrawerProps) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)

  const initialBetragEuro =
    props.saldoCents < 0 ? formatCents(Math.abs(props.saldoCents)) : ''
  const [betragEuro, setBetragEuro] = useState(initialBetragEuro)
  const [kommentar, setKommentar] = useState('')
  const [kommentarTouched, setKommentarTouched] = useState(false)

  const betragCents = parseCents(betragEuro)
  const betragInvalid = betragCents < 1
  const kommentarInvalid = kommentar.trim().length < 3

  const onSubmit = async () => {
    setLoading(true)

    try {
      await props.backend.auszahlungLeisten({
        tischId: props.tisch.id,
        betragCents,
        kommentar,
      })
      props.auszahlungGeleistet()
      setOpen(false)
    } catch (error: unknown) {
      console.error(error)
      toast.error(
        getActionErrorMessage({
          actionLabel: 'Auszahlung leisten',
          error,
        }),
      )
    }

    setLoading(false)
  }

  const onOpenChange = (isOpen: boolean) => {
    setOpen(isOpen)
    if (!isOpen) {
      const reset =
        props.saldoCents < 0 ? formatCents(Math.abs(props.saldoCents)) : ''
      setBetragEuro(reset)
      setKommentar('')
      setKommentarTouched(false)
    }
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerTrigger asChild>
        <Button
          variant="outline"
          className="cursor-pointer hover:shadow-sm w-full"
        >
          Auszahlung
        </Button>
      </DrawerTrigger>
      <DrawerContent>
        <div className="mx-auto w-full max-w-sm">
          <DrawerHeader>
            <DrawerTitle>Auszahlung für {props.tisch.name}</DrawerTitle>
            <DrawerDescription>
              Betrag und Kommentar angeben, dann bestätigen.
            </DrawerDescription>
          </DrawerHeader>
          <div className="px-4 flex flex-col gap-3">
            <Field>
              <Input
                type="number"
                min="0.01"
                step="0.01"
                placeholder="Betrag in €"
                value={betragEuro}
                onChange={(e) => {
                  setBetragEuro(e.target.value)
                }}
                spellCheck={false}
              />
            </Field>
            <Field>
              <Textarea
                className="resize-none"
                placeholder="Kommentar (erforderlich)"
                rows={3}
                maxLength={100}
                value={kommentar}
                onChange={(e) => {
                  setKommentar(e.target.value)
                  setKommentarTouched(true)
                }}
                spellCheck={false}
              />
              {kommentarTouched && kommentarInvalid && (
                <p className="text-sm text-destructive mt-1">
                  Kommentar ist erforderlich (mind. 3 Zeichen).
                </p>
              )}
            </Field>
          </div>
          <DrawerFooter>
            <Button
              variant="secondary"
              disabled={loading || betragInvalid || kommentarInvalid}
              onClick={() => {
                void onSubmit()
              }}
            >
              {loading ? <Spinner /> : <></>} Auszahlung leisten
            </Button>
            <DrawerClose asChild>
              <Button variant="outline" disabled={loading}>
                Abbrechen
              </Button>
            </DrawerClose>
          </DrawerFooter>
        </div>
      </DrawerContent>
    </Drawer>
  )
}
