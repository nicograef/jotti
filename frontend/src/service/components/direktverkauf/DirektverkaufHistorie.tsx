import { Printer, X } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Item,
  ItemActions,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemTitle,
} from '@/components/ui/item'
import { Skeleton } from '@/components/ui/skeleton'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { AuthSingleton } from '@/lib/Auth'
import { formatCents } from '@/lib/utils'

import type { DirektverkaufHistorieEintrag } from '../../direktverkauf/Direktverkauf'
import type { DirektverkaufBackend } from '../../direktverkauf/DirektverkaufBackend'
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
  const [stornoVerkauf, setStornoVerkauf] =
    useState<DirektverkaufHistorieEintrag | null>(null)
  const { loading: belegDruckenLoading, run: runBelegDrucken } =
    useActionSubmit({
      actionLabel: 'Kassenbeleg drucken',
      byCode: {
        kassenbeleg_drucker_nicht_konfiguriert:
          'Kein Kassenbeleg-Drucker konfiguriert. Bitte in den Admin-Einstellungen hinterlegen.',
        verkauf_not_found: 'Der Verkauf wurde nicht gefunden.',
      },
      onSuccess: () => {
        toast.success('Kassenbeleg in die Druckwarteschlange eingereiht.')
      },
    })

  if (historieLoading) {
    return (
      <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
        {Array.from({ length: 6 }).map((_, index) => (
          // eslint-disable-next-line react-x/no-array-index-key
          <HistorieItemSkeleton key={index} />
        ))}
      </ItemGroup>
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
      <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
        {historie.map((verkauf) => {
          const stornierbar =
            AuthSingleton.canCancel && verkauf.offenePositionen.length > 0
          return (
            <Item key={verkauf.verkaufId} variant="outline">
              <ItemContent>
                <ItemTitle className="flex flex-wrap items-center gap-2">
                  {formatCents(verkauf.gesamtbetragCents)}&nbsp;€
                  {verkauf.gesamtStorniertCents > 0 && (
                    <Badge variant="destructive">
                      −{formatCents(verkauf.gesamtStorniertCents)}&nbsp;€
                      storniert
                    </Badge>
                  )}
                </ItemTitle>
                <ItemDescription>
                  {new Date(verkauf.getaetigtAm).toLocaleString('de-DE')} ·{' '}
                  {verkauf.userName}
                  {verkauf.kommentar && (
                    <>
                      <br />
                      {verkauf.kommentar}
                    </>
                  )}
                </ItemDescription>
              </ItemContent>
              <ItemActions>
                <Button
                  size="icon-sm"
                  variant="outline"
                  className="rounded-full cursor-pointer"
                  aria-label="Kassenbeleg drucken"
                  disabled={belegDruckenLoading}
                  onClick={() => {
                    void runBelegDrucken(async () => {
                      await backend.kassenbelegDrucken({
                        verkaufId: verkauf.verkaufId,
                      })
                    })
                  }}
                >
                  <Printer />
                </Button>
                {stornierbar && (
                  <Button
                    size="icon-sm"
                    variant="destructive"
                    className="rounded-full cursor-pointer"
                    aria-label="Stornieren"
                    onClick={() => {
                      setStornoVerkauf(verkauf)
                    }}
                  >
                    <X />
                  </Button>
                )}
              </ItemActions>
            </Item>
          )
        })}
      </ItemGroup>
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

function HistorieItemSkeleton() {
  return (
    <Item variant="outline">
      <ItemContent>
        <ItemTitle>
          <Skeleton className="h-6 w-32" />
        </ItemTitle>
        <Skeleton className="h-4 w-48" />
      </ItemContent>
    </Item>
  )
}
