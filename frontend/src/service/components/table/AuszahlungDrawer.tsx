import { HandCoins } from 'lucide-react'
import { useState } from 'react'

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
import { useActionSubmit } from '@/hooks/use-action-submit'
import { formatCents, parseCents } from '@/lib/utils'

import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { KommentarField } from './CommentField'

interface AuszahlungDrawerProps {
  backend: Pick<TischBackend, 'auszahlungLeisten'>
  tisch: Tisch
  saldoCents: number
  auszahlungGeleistet: () => void
}

export function AuszahlungDrawer(props: AuszahlungDrawerProps) {
  const [open, setOpen] = useState(false)

  const initialBetragEuro =
    props.saldoCents < 0 ? formatCents(Math.abs(props.saldoCents)) : ''
  const [betragEuro, setBetragEuro] = useState(initialBetragEuro)
  const [kommentar, setKommentar] = useState('')

  const betragCents = parseCents(betragEuro)
  const betragInvalid = betragCents < 1
  const kommentarInvalid = kommentar.trim().length < 3

  const { loading, run } = useActionSubmit({
    actionLabel: 'Auszahlung leisten',
    onSuccess: () => {
      props.auszahlungGeleistet()
      setOpen(false)
    },
  })

  const onSubmit = async () => {
    await run(async () => {
      await props.backend.auszahlungLeisten({
        tischId: props.tisch.id,
        betragCents,
        kommentar,
      })
    })
  }

  const onOpenChange = (isOpen: boolean) => {
    setOpen(isOpen)
    if (!isOpen) {
      const reset =
        props.saldoCents < 0 ? formatCents(Math.abs(props.saldoCents)) : ''
      setBetragEuro(reset)
      setKommentar('')
    }
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerTrigger asChild>
        <Button
          variant="outline"
          aria-label="Auszahlung"
          className="cursor-pointer hover:shadow-sm w-full"
        >
          {props.saldoCents < 0 ? 'Auszahlung' : <HandCoins />}
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
            <KommentarField
              required
              invalid={kommentarInvalid}
              onChange={setKommentar}
            />
          </div>
          <DrawerFooter>
            <Button
              variant="secondary"
              disabled={loading || betragInvalid || kommentarInvalid}
              onClick={() => {
                void onSubmit()
              }}
            >
              {loading ? <Spinner /> : null} Auszahlung leisten
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
