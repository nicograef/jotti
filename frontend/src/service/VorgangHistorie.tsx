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
import { AuthSingleton } from '@/lib/Auth'
import {
  formatEuro,
  formatPositionName,
  formatRelativeTime,
  formatZeitpunkt,
} from '@/lib/utils'

import { type BelegStatus, useBelegDrucken } from './beleg'
import { Kommentar } from './CommentField'
import { toReceiptItems } from './drawerUtils'
import { HistorieSkeletonGrid } from './HistorieSkeletonGrid'
import { Receipt } from './Receipt'
import type {
  GebuchtePosition,
  PositionRef,
  VorgangHistorieBasis,
} from './schemas'
import { VorgangStornoDrawer } from './VorgangStornoDrawer'

// id ist die Vorgangs-ID (verkaufId oder rueckgabeId).
export type VorgangEintrag = VorgangHistorieBasis & { id: string }

function positionenZusammenfassung(positionen: GebuchtePosition[]): string {
  return positionen
    .map(
      (position) =>
        `${position.menge.toString()}× ${formatPositionName(position.produktName, position.varianteName)}`,
    )
    .join(', ')
}

interface VorgangHistorieProps {
  eintraege: VorgangEintrag[]
  loading: boolean
  // Zeilen- und Drawer-Titel, etwa „Verkauf“ oder „Pfandrückgabe“.
  bezeichnung: string
  leerHinweis: string
  stornieren: (
    id: string,
    schluessel: string,
    positionen: PositionRef[],
    kommentar: string,
  ) => Promise<void>
  belegDrucken: (id: string, stornierungId?: string) => Promise<BelegStatus>
  onErfolg: (nachricht: string) => void
}

export function VorgangHistorie({
  eintraege,
  loading,
  bezeichnung,
  leerHinweis,
  stornieren,
  belegDrucken,
  onErfolg,
}: VorgangHistorieProps) {
  const [detail, setDetail] = useState<VorgangEintrag | null>(null)
  const [stornoVorgang, setStornoVorgang] = useState<VorgangEintrag | null>(
    null,
  )
  const { loading: belegDruckenLoading, drucken } = useBelegDrucken()

  const belegAnfordern = (id: string, stornierungId?: string) => {
    drucken(
      () => belegDrucken(id, stornierungId),
      'Kassenbeleg in die Druckwarteschlange eingereiht.',
    )
  }

  if (loading) {
    return <HistorieSkeletonGrid />
  }

  if (eintraege.length === 0) {
    return (
      <p className="text-muted-foreground text-center py-8">{leerHinweis}</p>
    )
  }

  return (
    <>
      <div className="grid grid-cols-1 gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
        {eintraege.map((verkauf) => (
          <button
            key={verkauf.id}
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
                {bezeichnung}
                {verkauf.gesamtStorniertCents > 0 && (
                  <Badge variant="destructive">
                    −{formatEuro(verkauf.gesamtStorniertCents)} storniert
                  </Badge>
                )}
              </span>
              <span className="text-sm text-muted-foreground">
                {formatRelativeTime(verkauf.getaetigtAm)} · {verkauf.userName}
                {verkauf.kommentar && ` · „${verkauf.kommentar}“`}
              </span>
              {/* line-clamp-1 statt truncate: truncate (nowrap) lässt die
                  Zusammenfassung in die Max-Content-Breite der Seite eingehen —
                  mobile Browser weiten dann den Layout-Viewport (Zoom-out) und
                  Tap-Koordinaten in Drawern gehen daneben. */}
              <span className="line-clamp-1 text-sm text-muted-foreground">
                {positionenZusammenfassung(verkauf.positionen)}
              </span>
            </span>
            <span className="shrink-0 text-[15px] font-bold tabular-nums">
              {formatEuro(verkauf.gesamtbetragCents)}
            </span>
            <ChevronRight className="size-4 shrink-0 text-muted-foreground" />
          </button>
        ))}
      </div>
      {detail && (
        <Details
          verkauf={detail}
          bezeichnung={bezeichnung}
          belegDruckenLoading={belegDruckenLoading}
          onKassenbeleg={() => {
            belegAnfordern(detail.id)
          }}
          onStornobeleg={(stornierungId) => {
            belegAnfordern(detail.id, stornierungId)
          }}
          onStornieren={
            AuthSingleton.canCancel && detail.offenePositionen.length > 0
              ? () => {
                  setStornoVorgang(detail)
                  setDetail(null)
                }
              : undefined
          }
          onClose={() => {
            setDetail(null)
          }}
        />
      )}
      {stornoVorgang && (
        <VorgangStornoDrawer
          bezeichnung={bezeichnung}
          offenePositionen={stornoVorgang.offenePositionen}
          stornieren={(schluessel, positionen, kommentar) =>
            stornieren(stornoVorgang.id, schluessel, positionen, kommentar)
          }
          onClose={() => {
            setStornoVorgang(null)
          }}
          onStorniert={() => {
            setStornoVorgang(null)
            onErfolg('Stornierung gebucht.')
          }}
        />
      )}
    </>
  )
}

function Details({
  verkauf,
  bezeichnung,
  belegDruckenLoading,
  onKassenbeleg,
  onStornobeleg,
  onStornieren,
  onClose,
}: {
  verkauf: VorgangEintrag
  bezeichnung: string
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
            {bezeichnung} · {formatRelativeTime(verkauf.getaetigtAm)} ·{' '}
            {verkauf.userName}
          </DrawerTitle>
          <DrawerDescription>
            {formatZeitpunkt(verkauf.getaetigtAm)}
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
                    Storno −{formatEuro(storno.gesamtStornierungCents)} ·{' '}
                    {formatRelativeTime(storno.storniertAm)}
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
