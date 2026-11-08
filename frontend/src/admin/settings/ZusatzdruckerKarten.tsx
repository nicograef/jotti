import { zodResolver } from '@hookform/resolvers/zod'
import { Plus, Printer, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
import { z } from 'zod'

import { NameField } from '@/components/common/FormFields'
import { LadefehlerAlert } from '@/components/common/LadefehlerAlert'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
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
import { FieldGroup } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Spinner } from '@/components/ui/spinner'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { useFormActionSubmit } from '@/hooks/use-form-action-submit'

import { BonmodusOptionen, DruckerIpFeld } from './DruckerFelder'
import type { Bonmodus } from './DruckstationBackend'
import { useZusatzdrucker } from './hooks'
import { ProduktAuswahlDialog } from './ProduktAuswahlDialog'
import {
  type ProduktAuswahl,
  type Zusatzdrucker,
  ZUSATZDRUCKER_BONMODI,
  ZusatzdruckerNameSchema,
} from './ZusatzdruckerBackend'

const AnlegenSchema = z.object({ name: ZusatzdruckerNameSchema })
type AnlegenDaten = z.infer<typeof AnlegenSchema>

function DruckerHinzufuegenDialog({
  anlegen,
}: {
  anlegen: (name: string) => Promise<void>
}) {
  const [open, setOpen] = useState(false)
  const form = useForm<AnlegenDaten>({
    defaultValues: { name: '' },
    resolver: zodResolver(AnlegenSchema),
    mode: 'onTouched',
  })
  const { loading, run } = useFormActionSubmit({
    form,
    actionLabel: 'Drucker anlegen',
  })

  const onSubmit = async (daten: AnlegenDaten) => {
    await run(async () => {
      await anlegen(daten.name)
      toast.success(`Drucker „${daten.name}“ angelegt.`)
      form.reset()
      setOpen(false)
    })
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        setOpen(isOpen)
        if (!isOpen) form.reset()
      }}
    >
      <DialogTrigger asChild>
        <Button variant="outline" size="sm" className="w-fit">
          <Plus /> Drucker hinzufügen
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader className="mb-4">
          <DialogTitle>Drucker hinzufügen</DialogTitle>
          <DialogDescription>
            Der Name steht auf dem Testbon, etwa der Ort der Ausgabe.
          </DialogDescription>
        </DialogHeader>
        <DialogBody>
          <form
            id="zusatzdrucker-form"
            onSubmit={(e) => {
              e.preventDefault()
              void form.handleSubmit(onSubmit)()
            }}
          >
            <FieldGroup>
              <NameField form={form} placeholder="z.B. Kaffee & Kuchen" />
            </FieldGroup>
          </form>
        </DialogBody>
        <DialogFooter className="mt-4">
          <DialogClose asChild>
            <Button variant="outline" disabled={loading}>
              Abbrechen
            </Button>
          </DialogClose>
          <Button
            type="submit"
            form="zusatzdrucker-form"
            disabled={loading || !form.formState.isValid}
          >
            {loading ? <Spinner /> : null} Drucker hinzufügen
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function ZugeordneteProdukte({
  zusatzdrucker,
  produktAuswahl,
  produktZuordnen,
  zuordnungEntfernen,
}: {
  zusatzdrucker: Zusatzdrucker
  produktAuswahl: ProduktAuswahl[]
  produktZuordnen: (produktId: number, zusatzdruckerId: number) => Promise<void>
  zuordnungEntfernen: (produktId: number) => Promise<void>
}) {
  const { id, name, produkte } = zusatzdrucker
  const { loading, run } = useActionSubmit({
    actionLabel: 'Produkt entfernen',
  })

  return (
    <div className="flex flex-col gap-2">
      <span className="text-sm text-muted-foreground">Produkte</span>
      {produkte.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          Noch keine Produkte. Sie drucken am Drucker ihrer Kategorie.
        </p>
      ) : (
        <ul className="flex flex-col divide-y rounded-md border">
          {produkte.map((produkt) => (
            <li
              key={produkt.id}
              className="flex items-center justify-between gap-2 py-1 pr-1 pl-3"
            >
              <span className="text-sm">{produkt.name}</span>
              <Button
                variant="ghost"
                size="sm"
                disabled={loading}
                onClick={() =>
                  void run(async () => {
                    await zuordnungEntfernen(produkt.id)
                    toast.success(
                      `„${produkt.name}“ druckt wieder am Drucker seiner Kategorie.`,
                    )
                  })
                }
              >
                Entfernen
              </Button>
            </li>
          ))}
        </ul>
      )}
      <ProduktAuswahlDialog
        zusatzdruckerName={name}
        zugeordneteIds={new Set(produkte.map((produkt) => produkt.id))}
        produktAuswahl={produktAuswahl}
        zuordnen={(produktId) => produktZuordnen(produktId, id)}
      />
    </div>
  )
}

function ZusatzdruckerCard({
  zusatzdrucker,
  produktAuswahl,
  aendern,
  loeschen,
  testbonDrucken,
  produktZuordnen,
  zuordnungEntfernen,
}: {
  zusatzdrucker: Zusatzdrucker
  produktAuswahl: ProduktAuswahl[]
  aendern: (zusatzdrucker: Zusatzdrucker) => Promise<void>
  loeschen: (id: number) => Promise<void>
  testbonDrucken: (id: number) => Promise<void>
  produktZuordnen: (produktId: number, zusatzdruckerId: number) => Promise<void>
  zuordnungEntfernen: (produktId: number) => Promise<void>
}) {
  const { id, name: gespeicherterName } = zusatzdrucker
  const [name, setName] = useState(gespeicherterName)
  const [nameError, setNameError] = useState<string | null>(null)
  const { loading: saving, run: runSave } = useActionSubmit({
    actionLabel: 'Drucker speichern',
  })
  const { loading: testing, run: runTestbon } = useActionSubmit({
    actionLabel: 'Testbon senden',
  })
  const { loading: deleting, run: runLoeschen } = useActionSubmit({
    actionLabel: 'Drucker löschen',
  })
  // Jede Speicherung sendet den ganzen Drucker; eine sperrt die anderen, bis
  // der Refetch den neuen Stand geliefert hat.

  const speichere = (aenderung: Partial<Zusatzdrucker>, meldung: string) =>
    runSave(async () => {
      await aendern({ ...zusatzdrucker, ...aenderung })
      toast.success(meldung)
    })

  const speichereName = async () => {
    const { data, error } = ZusatzdruckerNameSchema.safeParse(name)
    if (error) {
      setNameError(error.issues[0]?.message ?? null)
      return
    }
    setNameError(null)
    setName(data)
    if (data === gespeicherterName) {
      return
    }
    await speichere({ name: data }, `Name „${data}“ gespeichert.`)
  }

  const waehleBonmodus = async (bonmodus: Bonmodus) => {
    if (bonmodus === zusatzdrucker.bonmodus) {
      return
    }
    if (bonmodus !== 'pro_position' && bonmodus !== 'pro_bestellung') {
      return
    }
    await speichere(
      { bonmodus },
      `Bonmodus für „${gespeicherterName}“ geändert.`,
    )
  }

  return (
    <div className="flex flex-col gap-3 rounded-xl border p-5">
      <div>
        <span className="text-base font-semibold">{gespeicherterName}</span>{' '}
        <span className="text-xs font-normal text-muted-foreground">
          — Zusätzlicher Drucker
        </span>
      </div>

      <div className="flex flex-col gap-1">
        <Label
          htmlFor={`zusatzdrucker-name-${String(id)}`}
          className="text-muted-foreground"
        >
          Name
        </Label>
        <Input
          id={`zusatzdrucker-name-${String(id)}`}
          value={name}
          disabled={saving}
          onChange={(e) => {
            setName(e.target.value)
            if (nameError !== null) {
              setNameError(null)
            }
          }}
          onBlur={() => void speichereName()}
          onKeyDown={(e) => {
            if (e.key === 'Enter') {
              e.currentTarget.blur()
            }
          }}
          aria-invalid={nameError !== null}
        />
        {nameError !== null && (
          <p className="text-sm text-destructive">{nameError}</p>
        )}
      </div>

      <DruckerIpFeld
        id={`zusatzdrucker-ip-${String(id)}`}
        gespeicherteIp={zusatzdrucker.druckerIp}
        disabled={saving}
        onSpeichern={(druckerIp) =>
          speichere(
            { druckerIp },
            `Drucker-IP für „${gespeicherterName}“ gespeichert.`,
          )
        }
        aktion={
          <Button
            variant="outline"
            size="sm"
            className="h-9"
            disabled={testing || zusatzdrucker.druckerIp === ''}
            onClick={() =>
              void runTestbon(async () => {
                await testbonDrucken(id)
                toast.success(`Testbon an „${gespeicherterName}“ gesendet.`)
              })
            }
          >
            <Printer className="size-4" />
            Testbon
          </Button>
        }
      />

      <BonmodusOptionen
        optionen={ZUSATZDRUCKER_BONMODI}
        aktiv={zusatzdrucker.bonmodus}
        disabled={saving}
        onSelect={(bonmodus) => void waehleBonmodus(bonmodus)}
      />

      <ZugeordneteProdukte
        zusatzdrucker={zusatzdrucker}
        produktAuswahl={produktAuswahl}
        produktZuordnen={produktZuordnen}
        zuordnungEntfernen={zuordnungEntfernen}
      />

      <AlertDialog>
        <AlertDialogTrigger asChild>
          <Button
            variant="ghost"
            size="sm"
            className="w-fit text-destructive hover:text-destructive"
            disabled={deleting}
          >
            <Trash2 /> Löschen
          </Button>
        </AlertDialogTrigger>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              Drucker „{gespeicherterName}“ löschen?
            </AlertDialogTitle>
            <AlertDialogDescription>
              Die Produkte drucken danach wieder am Drucker ihrer Kategorie.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Abbrechen</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive-solid"
              disabled={deleting}
              onClick={() =>
                void runLoeschen(async () => {
                  await loeschen(id)
                  toast.success(`Drucker „${gespeicherterName}“ gelöscht.`)
                })
              }
            >
              Löschen
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

export function ZusatzdruckerAbschnitt() {
  const {
    zusatzdrucker,
    produktAuswahl,
    isPending,
    isError,
    refetch,
    anlegen,
    aendern,
    loeschen,
    testbonDrucken,
    produktZuordnen,
    zuordnungEntfernen,
  } = useZusatzdrucker()

  if (isPending) {
    return null
  }
  if (isError) {
    return (
      <div className="mt-4">
        <LadefehlerAlert
          titel="Zusätzliche Drucker konnten nicht geladen werden"
          onErneutVersuchen={() => void refetch()}
        />
      </div>
    )
  }

  return (
    <div className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2">
      {zusatzdrucker.map((zd) => (
        <ZusatzdruckerCard
          key={zd.id}
          zusatzdrucker={zd}
          produktAuswahl={produktAuswahl}
          aendern={aendern}
          loeschen={loeschen}
          testbonDrucken={testbonDrucken}
          produktZuordnen={produktZuordnen}
          zuordnungEntfernen={zuordnungEntfernen}
        />
      ))}
      <div className="flex flex-col gap-2 rounded-xl border border-dashed bg-muted/30 p-5">
        <span className="text-base font-semibold text-muted-foreground">
          Zusätzlicher Drucker
        </span>
        <p className="text-sm leading-relaxed text-muted-foreground">
          Für Produkte, die an einem anderen Ort zubereitet werden.
        </p>
        <DruckerHinzufuegenDialog anlegen={anlegen} />
      </div>
    </div>
  )
}
