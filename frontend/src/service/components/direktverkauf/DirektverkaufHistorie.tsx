import { Banknote, ChevronRight, Printer, RotateCcw } from 'lucide-react'
import { useState } from 'react'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Drawer,
  DrawerBody,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
} from '@/components/ui/drawer'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { AuthSingleton } from '@/lib/Auth'
import { formatCents, formatRelativeTime } from '@/lib/utils'

import { belegDruckenMitNachfassen, meldeBelegStatus } from '../../beleg'
import type {
  DirektverkaufHistorieEintrag,
  DirektverkaufKassenbelegDrucken,
} from '../../direktverkauf/Direktverkauf'
import type { DirektverkaufBackend } from '../../direktverkauf/DirektverkaufBackend'
import { Kommentar } from '../table/CommentField'
import { toReceiptItems } from '../table/drawerUtils'
import { HistorieRowSkeleton } from '../table/HistorieRowSkeleton'
import { Receipt } from '../table/Receipt'
import { DirektverkaufStornoDrawer } from './DirektverkaufStornoDrawer'

interface DirektverkaufHistorieProps {
  historie: DirektverkaufHistorieEintrag[]
  historieLoading: boolean
  backend: Pick<
    DirektverkaufBackend,
    'direktverkaufStornieren' | 'kassenbelegDrucken'
  >
  onStorniert: () => void
}

export function DirektverkaufHistorie({
  historie,
  historieLoading,
  backend,
  onStorniert,
}: DirektverkaufHistorieProps) {
  const [detail, setDetail] = useState<DirektverkaufHistorieEintrag | null>(
    null,
  )
  const [stornoVerkauf, setStornoVerkauf] =
    useState<DirektverkaufHistorieEintrag | null>(null)
  const { loading: belegDruckenLoading, run: runBelegDrucken } =
    useActionSubmit({
      actionLabel: 'Kassenbeleg drucken',
      byCode: {
        kassenbeleg_drucker_nicht_konfiguriert:
          'Kein Kassenbeleg-Drucker konfiguriert. Bitte in den Admin-Einstellungen hinterlegen.',
        verkauf_not_found: 'Der Verkauf wurde nicht gefunden.',
        stornierung_not_found: 'Die Stornierung wurde nicht gefunden.',
      },
    })

  const belegAnfordern = (cmd: DirektverkaufKassenbelegDrucken) => {
    void runBelegDrucken(async () => {
      const status = await belegDruckenMitNachfassen(() =>
        backend.kassenbelegDrucken(cmd),
      )
      meldeBelegStatus(
        status,
        'Kassenbeleg in die Druckwarteschlange eingereiht.',
      )
    })
  }

  if (historieLoading) {
    return (
      <div className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
        {Array.from({ length: 6 }).map((_, index) => (
          // eslint-disable-next-line react-x/no-array-index-key
          <HistorieRowSkeleton key={index} />
        ))}
      </div>
    )
  }

  if (historie.length === 0) {
    return (
      <p className="text-muted-foreground text-center py-8">
        Noch keine Direktverkäufe in dieser Kassensitzung.
      </p>
    )
  }

  return (
    <>
      <div className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
        {historie.map((verkauf) => (
          <button
            key={verkauf.verkaufId}
            type="button"
            onClick={() => {
              setDetail(verkauf)
            }}
            className="flex w-full items-center gap-3 rounded-md border bg-card px-3 py-3 text-left transition-colors hover:bg-muted/50"
          >
            <span className="flex size-10 shrink-0 items-center justify-center rounded-full bg-muted text-muted-foreground [&_svg]:size-5">
              <Banknote />
            </span>
            <span className="flex min-w-0 flex-1 flex-col gap-0.5">
              <span className="flex flex-wrap items-center gap-2 text-[15px] font-medium">
                Verkauf
                {verkauf.gesamtStorniertCents > 0 && (
                  <Badge variant="destructive">
                    −{formatCents(verkauf.gesamtStorniertCents)}&nbsp;€
                    storniert
                  </Badge>
                )}
              </span>
              <span className="text-sm text-muted-foreground">
                {formatRelativeTime(verkauf.getaetigtAm)} · {verkauf.userName}
                {verkauf.kommentar && ` · „${verkauf.kommentar}“`}
              </span>
            </span>
            <span className="shrink-0 text-[15px] font-bold tabular-nums">
              {formatCents(verkauf.gesamtbetragCents)}&nbsp;€
            </span>
            <ChevronRight className="size-4 shrink-0 text-muted-foreground" />
          </button>
        ))}
      </div>
      {detail && (
        <Details
          verkauf={detail}
          belegDruckenLoading={belegDruckenLoading}
          onKassenbeleg={() => {
            belegAnfordern({ verkaufId: detail.verkaufId })
          }}
          onStornobeleg={(stornierungId) => {
            belegAnfordern({ verkaufId: detail.verkaufId, stornierungId })
          }}
          onStornieren={
            AuthSingleton.canCancel && detail.offenePositionen.length > 0
              ? () => {
                  setStornoVerkauf(detail)
                  setDetail(null)
                }
              : undefined
          }
          onClose={() => {
            setDetail(null)
          }}
        />
      )}
      {stornoVerkauf && (
        <DirektverkaufStornoDrawer
          backend={backend}
          verkauf={stornoVerkauf}
          onClose={() => {
            setStornoVerkauf(null)
          }}
          onStorniert={() => {
            setStornoVerkauf(null)
            onStorniert()
          }}
        />
      )}
    </>
  )
}

function Details({
  verkauf,
  belegDruckenLoading,
  onKassenbeleg,
  onStornobeleg,
  onStornieren,
  onClose,
}: {
  verkauf: DirektverkaufHistorieEintrag
  belegDruckenLoading: boolean
  onKassenbeleg: () => void
  onStornobeleg: (stornierungId: string) => void
  onStornieren?: () => void
  onClose: () => void
}) {
  return (
    <Drawer
      open
      onOpenChange={(open) => {
        if (!open) onClose()
      }}
    >
      <DrawerContent>
        <DrawerHeader className="mx-auto w-full max-w-sm">
          <DrawerTitle>
            Verkauf · {formatRelativeTime(verkauf.getaetigtAm)} ·{' '}
            {verkauf.userName}
          </DrawerTitle>
          <DrawerDescription>
            {new Date(verkauf.getaetigtAm).toLocaleString('de-DE', {
              day: 'numeric',
              month: 'numeric',
              year: 'numeric',
              hour: '2-digit',
              minute: '2-digit',
            })}
          </DrawerDescription>
        </DrawerHeader>
        <DrawerBody className="mx-auto w-full max-w-sm">
          <Receipt
            positionen={toReceiptItems(verkauf.positionen)}
            totalPrice={verkauf.gesamtbetragCents}
          />
          {verkauf.kommentar && (
            <div className="px-4">
              <Kommentar value={verkauf.kommentar} />
            </div>
          )}
          {verkauf.stornierungen.length > 0 && (
            <div className="mt-2 flex flex-col gap-1 px-4">
              {verkauf.stornierungen.map((storno) => (
                <div
                  key={storno.stornierungId}
                  className="flex items-center justify-between gap-2 text-sm text-muted-foreground"
                >
                  <span>
                    Storno −{formatCents(storno.gesamtStornierungCents)}&nbsp;€
                    · {formatRelativeTime(storno.storniertAm)}
                  </span>
                  <Button
                    size="icon-sm"
                    variant="outline"
                    className="rounded-full cursor-pointer"
                    aria-label="Stornobeleg drucken"
                    disabled={belegDruckenLoading}
                    onClick={() => {
                      onStornobeleg(storno.stornierungId)
                    }}
                  >
                    <Printer />
                  </Button>
                </div>
              ))}
            </div>
          )}
        </DrawerBody>
        <DrawerFooter className="mx-auto w-full max-w-sm">
          <div className="flex gap-2">
            <Button
              variant="outline"
              className="flex-1"
              disabled={belegDruckenLoading}
              onClick={onKassenbeleg}
            >
              <Printer /> Kassenbeleg drucken
            </Button>
            {onStornieren && (
              <Button
                variant="outline"
                className="flex-1 border-destructive/40 text-destructive"
                disabled={belegDruckenLoading}
                onClick={onStornieren}
              >
                <RotateCcw /> Stornieren…
              </Button>
            )}
          </div>
          <DrawerClose asChild>
            <Button variant="outline" disabled={belegDruckenLoading}>
              Schließen
            </Button>
          </DrawerClose>
        </DrawerFooter>
      </DrawerContent>
    </Drawer>
  )
}
