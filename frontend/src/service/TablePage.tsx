import { TriangleAlert } from 'lucide-react'
import { useParams } from 'react-router'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import {
  Item,
  ItemContent,
  ItemDescription,
  ItemTitle,
} from '@/components/ui/item'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useIsMobile } from '@/hooks/use-mobile'
import { BackendSingleton } from '@/lib/Backend'
import { formatCents } from '@/lib/utils'

import { Ausgabe } from './components/table/Ausgabe'
import { Bestellung } from './components/table/Bestellung'
import { TischHistorie } from './components/table/TischHistorie'
import { Zahlung } from './components/table/Zahlung'
import { useAktiveProdukte } from './product/hooks'
import { useTischHistorie, useTischState } from './table/hooks'
import { TischBackend } from './table/TischBackend'

const tischBackend = new TischBackend(BackendSingleton)

// Unterer Freiraum der Produktliste, damit die letzte Zeile über den beiden
// fixierten Leisten (StickyActionBar + Tab-Leiste) endet und antippbar bleibt.
// Die Leisten sitzen mobil bei 4rem, desktop bei 1rem über der Safe-Area und
// sind 3.5rem hoch; ihre Oberkante liegt also 7.5rem (mobil) bzw. 4.5rem
// (desktop) über der Safe-Area. Der Freiraum leitet sich davon plus 1rem
// Sicherheitsabstand ab und wächst mit env(safe-area-inset-bottom) mit. Das
// frühere statische pb-40 wuchs nicht mit und war auf Geräten mit Inset zu knapp.
const produktlistenFreiraum =
  'pb-[calc(8.5rem+env(safe-area-inset-bottom,0px))] md:pb-[calc(5.5rem+env(safe-area-inset-bottom,0px))]'

export function TablePage() {
  const { tischId } = useParams<{ tischId: string }>()
  const isMobile = useIsMobile()
  const {
    state,
    isPending: stateLoading,
    isError: stateError,
    refetch: reloadState,
  } = useTischState(Number(tischId))
  const { isPending, produkte } = useAktiveProdukte()
  const {
    isPending: historieLoading,
    isError: historieError,
    historie,
    refetch: reloadHistorie,
  } = useTischHistorie(Number(tischId))

  const reload = () => {
    void reloadState()
    void reloadHistorie()
  }

  // Expliziter Fehlerzustand statt der Leer-Defaults (Saldo 0,00 €,
  // „Alles ausgegeben!") — sonst wirkt der Tisch bei Netzabbruch abgerechnet.
  if (stateError || historieError) {
    return (
      <Alert variant="destructive">
        <TriangleAlert className="size-4" />
        <AlertTitle>Tischdaten konnten nicht geladen werden</AlertTitle>
        <AlertDescription>
          <p>Bitte die Verbindung prüfen und erneut versuchen.</p>
          <Button variant="outline" size="sm" onClick={reload}>
            Erneut versuchen
          </Button>
        </AlertDescription>
      </Alert>
    )
  }

  const tisch = {
    id: state.tischId,
    name: state.tischName,
    saldoCents: state.saldoCents,
  }

  const offenePositionen = state.ausstehendePositionen.reduce(
    (sum, position) => sum + position.menge,
    0,
  )
  const tabsLocked = stateLoading || historieLoading

  return (
    <>
      <Item>
        <ItemContent>
          <ItemTitle className="text-2xl">
            {stateLoading ? 'Tisch ??' : tisch.name}{' '}
            {!stateLoading && offenePositionen > 0 && (
              <Badge variant="destructive">{offenePositionen} offen</Badge>
            )}
            {!stateLoading && offenePositionen === 0 && (
              <Badge>Alles ausgegeben!</Badge>
            )}
          </ItemTitle>
          {!stateLoading && (
            <ItemDescription>
              {state.fuerMichErledigt ? (
                <span className="font-medium text-green-600">
                  Für dich erledigt
                </span>
              ) : (
                <span className="text-muted-foreground">
                  Für dich noch offen
                </span>
              )}
            </ItemDescription>
          )}
        </ItemContent>
        <ItemContent>
          <ItemDescription className="text-2xl">
            {stateLoading ? '?' : `${formatCents(state.saldoCents)} €`}
          </ItemDescription>
        </ItemContent>
      </Item>
      <Tabs defaultValue="order">
        <div
          className={
            isMobile
              ? 'w-full fixed bottom-[calc(1rem+env(safe-area-inset-bottom,0px))] left-0 z-50 flex flex-col items-center gap-2'
              : 'mb-4 flex flex-col items-start gap-2'
          }
        >
          {tabsLocked && (
            <p className="text-xs text-muted-foreground bg-background/90 px-3 py-1 rounded-md border">
              Lade Tischdaten. Tabs sind kurzzeitig deaktiviert.
            </p>
          )}
          <TabsList>
            <TabsTrigger value="order" className="p-4" disabled={tabsLocked}>
              Bestellen
            </TabsTrigger>
            <TabsTrigger value="payment" className="p-4" disabled={tabsLocked}>
              Kassieren
            </TabsTrigger>
            <TabsTrigger value="history" className="p-4" disabled={tabsLocked}>
              Historie
            </TabsTrigger>
          </TabsList>
        </div>
        <TabsContent value="order" className={produktlistenFreiraum}>
          {!stateLoading && (
            <>
              {offenePositionen > 0 && (
                <Card className="p-2 gap-0 mb-4">
                  <Ausgabe
                    backend={tischBackend}
                    tisch={tisch}
                    positionen={state.ausstehendePositionen}
                    loading={stateLoading}
                    onAusgabeBestaetigt={reload}
                  />
                </Card>
              )}
              <Bestellung
                backend={tischBackend}
                tisch={tisch}
                products={produkte}
                productsLoading={isPending}
                onBestellungAufgenommen={reload}
              />
            </>
          )}
        </TabsContent>
        <TabsContent value="payment" className={produktlistenFreiraum}>
          {!stateLoading && (
            <Zahlung
              backend={tischBackend}
              tisch={tisch}
              positionen={state.unbezahltePositionen}
              loading={stateLoading}
              onZahlungKassiert={reload}
            />
          )}
        </TabsContent>
        <TabsContent value="history" className={isMobile ? 'pb-24' : ''}>
          {!stateLoading && (
            <TischHistorie
              historie={historie}
              historieLoading={historieLoading}
              tisch={tisch}
              backend={tischBackend}
              onStornierungErteilt={reload}
              onBestellungUmgebucht={reload}
            />
          )}
        </TabsContent>
      </Tabs>
    </>
  )
}
