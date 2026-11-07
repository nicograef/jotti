import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
import { z } from 'zod'

import { useTSEKonfiguration, useTSEStatus } from '@/admin/tse/hooks'
import { istTestUmgebung, TSE_TEST_WARNUNG } from '@/admin/tse/tseAmpel'
import { EuroField, TextField } from '@/components/common/FormFields'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { FieldGroup } from '@/components/ui/field'
import { useActionSubmit } from '@/hooks/use-action-submit'

import { kasseBackend } from './hooks'
import { KassensitzungEroeffnenSchema } from './KasseBackend'

type FormData = z.infer<typeof KassensitzungEroeffnenSchema>

export function EroeffnenSection({ onSuccess }: { onSuccess: () => void }) {
  const { tseKonfiguration } = useTSEKonfiguration()
  const { tseStatus } = useTSEStatus()
  const testUmgebung = istTestUmgebung(tseStatus)
  const [tseDialogOpen, setTseDialogOpen] = useState(false)
  const [pendingData, setPendingData] = useState<FormData | null>(null)

  const form = useForm<FormData>({
    defaultValues: {
      bezeichnung: '',
      betragCents: 0,
    },
    resolver: zodResolver(KassensitzungEroeffnenSchema),
    mode: 'onTouched',
  })

  const { loading, run } = useActionSubmit({
    actionLabel: 'Kassensitzung eröffnen',
    onCode: {
      // Retry nach verlorener Antwort: der erste Versuch hat schon eröffnet.
      kasse_bereits_geoeffnet: () => {
        toast.info('Die Kasse ist bereits geöffnet.')
        setTseDialogOpen(false)
        onSuccess()
      },
    },
  })

  const eroeffnen = async (data: FormData) => {
    await run(async () => {
      await kasseBackend.kassensitzungEroeffnen(
        data.bezeichnung,
        data.betragCents,
      )
      toast.success('Kassensitzung eröffnet.')
      setTseDialogOpen(false)
      onSuccess()
    })
  }

  const onSubmit = async (data: FormData) => {
    if (tseKonfiguration?.istKonfiguriert && !testUmgebung) {
      await eroeffnen(data)
      return
    }
    setPendingData(data)
    setTseDialogOpen(true)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Kassensitzung eröffnen</CardTitle>
        <CardDescription>Die Kasse für den Verkaufstag öffnen</CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-muted-foreground text-sm mb-4">
          Trage den Anfangsbestand ein, also das Wechselgeld, das beim Start in
          der Kasse liegt. Die Kassensitzung läuft, bis die Kasse am Ende des
          Tages abgeschlossen wird.
        </p>
        <form
          onSubmit={(e) => {
            e.preventDefault()
            void form.handleSubmit(onSubmit)()
          }}
        >
          <FieldGroup>
            <TextField
              form={form}
              name="bezeichnung"
              label="Bezeichnung"
              placeholder="z.B. Sommerfest Tag 1"
              className="w-72"
            />
            <EuroField
              form={form}
              name="betragCents"
              label="Anfangsbestand"
              placeholder="z.B. 150,00"
              className="w-44"
            />
            <div>
              <Button type="submit" disabled={loading}>
                Kassensitzung eröffnen
              </Button>
            </div>
          </FieldGroup>
        </form>

        <AlertDialog open={tseDialogOpen} onOpenChange={setTseDialogOpen}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>
                {testUmgebung
                  ? 'TSE läuft in der Test-Umgebung'
                  : 'Keine TSE konfiguriert'}
              </AlertDialogTitle>
              <AlertDialogDescription>
                {testUmgebung
                  ? `${TSE_TEST_WARNUNG} Trotzdem eröffnen?`
                  : 'Es ist keine technische Sicherheitseinrichtung (TSE) eingerichtet. Vorgänge dieser Kassensitzung werden nicht nach § 146a AO signiert. Trotzdem eröffnen?'}
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel disabled={loading}>
                Abbrechen
              </AlertDialogCancel>
              <AlertDialogAction
                disabled={loading}
                onClick={(e) => {
                  e.preventDefault()
                  if (pendingData) void eroeffnen(pendingData)
                }}
              >
                Trotzdem eröffnen
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </CardContent>
    </Card>
  )
}
