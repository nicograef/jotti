import { useQueryClient } from '@tanstack/react-query'
import { useCallback, useState } from 'react'
import { useParams } from 'react-router'

import { LadefehlerAlert } from '@/components/common/LadefehlerAlert'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useCountUp } from '@/hooks/use-count-up'
import { useErstAufbau } from '@/hooks/use-erst-aufbau'
import { useMengen } from '@/hooks/use-mengen'
import { useIsMobile } from '@/hooks/use-mobile'
import { useVorgangId } from '@/hooks/use-vorgang-id'
import { BackendSingleton } from '@/lib/Backend'
import { formatEuro } from '@/lib/utils'

import { ErfolgsPop } from './components/ErfolgsPop'
import { dockFreiraum, ServiceDock } from './components/ServiceDock'
import { Bestellung } from './components/table/Bestellung'
import { TischHistorie } from './components/table/TischHistorie'
import { Zahlung } from './components/table/Zahlung'
import { useAktiveProdukte } from './product/hooks'
import type { Produkt } from './product/Produkt'
import {
  AKTIVE_TISCHE_MIT_FAVORITEN_KEY,
  EIGENE_UEBERSICHT_KEY,
  MEINE_TISCHE_STATE_KEY,
  TISCH_HISTORIE_KEY,
  TISCH_STATE_KEY,
  useTischHistorie,
  useTischState,
} from './table/hooks'
import { TischBackend } from './table/TischBackend'

const tischBackend = new TischBackend(BackendSingleton)

// Deckelt die gehobene Kassieren-Auswahl auf die noch unbezahlte Menge je
// Position: Einträge über ihrer Obergrenze sinken auf die Obergrenze, Einträge
// für verschwundene Positionen (Obergrenze 0) fallen heraus. Gibt `null`
// zurück, wenn nichts zu deckeln ist — damit der State-Abgleich im Render nur
// bei echter Änderung ein setAll auslöst und keine Render-Schleife dreht.
function deckeleAuswahl(
  auswahl: Record<string, number>,
  obergrenzen: Record<string, number>,
): Record<string, number> | null {
  let geaendert = false
  const gedeckelt: Record<string, number> = {}
  for (const [positionId, menge] of Object.entries(auswahl)) {
    const obergrenze = obergrenzen[positionId] || 0
    if (obergrenze <= 0) {
      geaendert = true
      continue
    }
    if (menge > obergrenze) {
      gedeckelt[positionId] = obergrenze
      geaendert = true
    } else {
      gedeckelt[positionId] = menge
    }
  }
  return geaendert ? gedeckelt : null
}

// Wirft aus dem Bestell-Korb, was nicht mehr in der aktiven Produktliste steht:
// Eine deaktivierte Variante hat keine Zeile mehr, ihr Korb-Eintrag wäre also
// weder sichtbar noch herunterzählbar. Gibt `null` zurück, wenn nichts zu
// entfernen ist — damit der State-Abgleich im Render nur bei echter Änderung ein
// setAll auslöst und keine Render-Schleife dreht (wie bei deckeleAuswahl).
function entferneUnbekannteVarianten(
  korb: Record<number, number>,
  produkte: Produkt[],
): Record<number, number> | null {
  const aktiveVarianten = new Set(
    produkte.flatMap((produkt) => produkt.varianten.map(({ id }) => id)),
  )
  let geaendert = false
  const bereinigt: Record<number, number> = {}
  for (const [varianteId, menge] of Object.entries(korb)) {
    if (!aktiveVarianten.has(Number(varianteId))) {
      geaendert = true
      continue
    }
    bereinigt[Number(varianteId)] = menge
  }
  return geaendert ? bereinigt : null
}

// Fachlicher Leerzustand einer Auswahl: Ein Eintrag, der auf 0 heruntergezählt
// wurde, bleibt in der Mengen-Karte stehen, zählt aber nicht als Auswahl.
function istAuswahlLeer(
  mengen: Record<string, number> | Record<number, number>,
): boolean {
  return Object.values(mengen).every((menge) => menge <= 0)
}

// Das Status-Badge poppt bei jedem Wertwechsel (Motion-Inventar „Statuswechsel",
// 350 ms), aber nicht beim ersten Aufbau. Der key-Wechsel auf den Zählwert
// remountet StatusBadgeInhalt, das seine Pop-Entscheidung beim Mount erfasst:
// Beim ersten Aufbau ist `animieren` noch false; nach dem ersten Aufbau mountet
// jeder neue Wert mit true und poppt.
function TischStatusBadge({ anzahlUnbezahlt }: { anzahlUnbezahlt: number }) {
  const erstAufbau = useErstAufbau(true)
  return (
    <StatusBadgeInhalt
      key={anzahlUnbezahlt}
      anzahlUnbezahlt={anzahlUnbezahlt}
      animieren={!erstAufbau}
    />
  )
}

function StatusBadgeInhalt({
  anzahlUnbezahlt,
  animieren,
}: {
  anzahlUnbezahlt: number
  animieren: boolean
}) {
  const [poppen] = useState(animieren)
  const popKlasse = poppen ? 'animate-pop' : undefined

  return anzahlUnbezahlt > 0 ? (
    <Badge variant="warn" className={popKlasse}>
      {anzahlUnbezahlt} unbezahlt
    </Badge>
  ) : (
    <Badge className={popKlasse}>Alles bezahlt</Badge>
  )
}

export function TablePage() {
  const isMobile = useIsMobile()
  const { tischId } = useParams<{ tischId: string }>()
  const {
    state,
    isPending: stateLoading,
    isLoadingError: stateError,
  } = useTischState(Number(tischId))
  const {
    isPending,
    isLoadingError: produkteError,
    produkte,
    refetch: reloadProdukte,
  } = useAktiveProdukte()
  const {
    isPending: historieLoading,
    isLoadingError: historieError,
    historie,
  } = useTischHistorie(Number(tischId))

  // Der Saldo zählt bei jeder Änderung animiert zum neuen Wert (u. a. nach dem
  // Schließen des Erfolgs-Pops, wenn der Refetch den Tischzustand aktualisiert).
  const animierterSaldo = useCountUp(state.saldoCents)

  // Eine Buchung ändert nicht nur diesen Tisch, sondern auch die Zahlen der
  // Tischübersicht (Meine Tische, Alle Tische, eigene Summen). Deren Queries
  // hängen an keiner Komponente dieser Seite und zeigten nach der Rückkehr sonst
  // den Cache-Stand von vor der Buchung — ein soeben kassierter Tisch stünde
  // weiter unter „Noch offen".
  // Zustand und Historie werden über ihr Präfix invalidiert statt über den
  // refetch dieses Tischs: Eine Umbuchung ändert auch den Ziel-Tisch, dessen
  // Queries hier nicht gemountet sind. Inaktive Queries werden dabei nur als
  // veraltet markiert und lösen keinen zusätzlichen Request aus; der gemountete
  // Tisch lädt wie bisher sofort neu.
  const queryClient = useQueryClient()
  const reload = useCallback(() => {
    void queryClient.invalidateQueries({ queryKey: [TISCH_STATE_KEY] })
    void queryClient.invalidateQueries({ queryKey: [TISCH_HISTORIE_KEY] })
    void queryClient.invalidateQueries({ queryKey: [MEINE_TISCHE_STATE_KEY] })
    void queryClient.invalidateQueries({
      queryKey: [AKTIVE_TISCHE_MIT_FAVORITEN_KEY],
    })
    void queryClient.invalidateQueries({ queryKey: [EIGENE_UEBERSICHT_KEY] })
  }, [queryClient])

  // Erfolgs-Pop: Bestellen, Kassieren, Stornieren und Umbuchen öffnen ihn mit
  // ihrer Meldung (statt eines Erfolgs-Toasts). Der nachgelagerte Refetch
  // (reload) läuft erst beim Schließen, damit sichtbare Statuswechsel (Saldo,
  // Badge, Listen) dem Pop folgen.
  const [erfolg, setErfolg] = useState({ open: false, text: '' })
  const zeigeErfolg = useCallback((nachricht: string) => {
    setErfolg({ open: true, text: nachricht })
  }, [])
  const erfolgSchliessen = useCallback(() => {
    setErfolg((prev) => ({ ...prev, open: false }))
    reload()
  }, [reload])

  // Bestell-Korb (Variante-ID → Menge) und Kassieren-Auswahl (Position-ID →
  // Menge) liegen hier, damit sie das Aus- und Wiedereinhängen der Radix-Tab-
  // Inhalte überstehen; ein Tab-Wechsel würde die Auswahl sonst verlieren. Die
  // Kassieren-Auswahl ist auf die noch unbezahlte Menge je Position gedeckelt.
  const bestellKorb = useMengen<number>()
  const unbezahlteMengen: Record<string, number> = {}
  state.unbezahltePositionen.forEach((position) => {
    unbezahlteMengen[position.positionId] = position.menge
  })
  const kassierenAuswahl = useMengen<string>(
    (positionId) => unbezahlteMengen[positionId] || 0,
  )

  // Beim Tischwechsel bleibt TablePage gemountet (nur der :tischId-Param
  // ändert sich), daher wird die gehobene Auswahl pro Tisch zurückgesetzt.
  // React-idiomatisches Zurücksetzen von State bei Prop-Wechsel im Render.
  const [aktiverTisch, setAktiverTisch] = useState(tischId)
  if (tischId !== aktiverTisch) {
    setAktiverTisch(tischId)
    bestellKorb.reset()
    kassierenAuswahl.reset()
  }

  // Der useMengen-`max` deckelt nur beim `add`, nicht die schon gespeicherte
  // Auswahl. Schrumpft die unbezahlte Menge einer bereits ausgewählten Position,
  // während die Auswahl bestehen bleibt (z. B. eine Stornierung auf der
  // Historie, deren Refetch erst beim Schließen des Erfolgs-Pops eintrifft),
  // wird die gespeicherte Auswahl beim Eintreffen der kleineren Obergrenzen im
  // Render abgeglichen — React-idiomatischer State-Sync wie beim Tischwechsel.
  const gedeckelteAuswahl = deckeleAuswahl(
    kassierenAuswahl.mengen,
    unbezahlteMengen,
  )
  if (gedeckelteAuswahl) {
    kassierenAuswahl.setAll(gedeckelteAuswahl)
  }

  // Derselbe Abgleich für den Bestell-Korb: Wird eine Variante deaktiviert,
  // während sie im Korb liegt, verschwindet ihre Zeile aus der Produktliste —
  // der Korb-Eintrag bliebe unsichtbar und nicht mehr herunterzählbar stehen.
  // Der Korb gälte damit nie wieder als leer, und die bestellungId rotierte für
  // die restliche Lebensdauer der Seite nicht mehr: Die nächste Bestellung liefe
  // unter dem Schlüssel der vorherigen. Erst abgleichen, wenn die Produktliste
  // geladen ist — vorher steht ihr Leerzustand für „noch nichts geladen", nicht
  // für „nicht mehr aktiv".
  const bereinigterKorb = isPending
    ? null
    : entferneUnbekannteVarianten(bestellKorb.mengen, produkte)
  if (bereinigterKorb) {
    bestellKorb.setAll(bereinigterKorb)
  }

  // Die Idempotenz-Schlüssel liegen bei ihrer Zusammenstellung — also hier, nicht
  // in den Abschluss-Komponenten: Radix hängt den Inhalt des inaktiven Tabs aus,
  // und ein dort gehaltener Schlüssel bekäme beim Tab-Wechsel einen neuen Wert.
  // Eine unveränderte Auswahl würde dann als zweiter Vorgang gebucht, obwohl die
  // erste Einreichung nur ihre Antwort verloren hat. Schlüssel und Auswahl
  // müssen dieselbe Lebensdauer haben.
  const bestellungId = useVorgangId(istAuswahlLeer(bestellKorb.mengen))
  const zahlungVorgangId = useVorgangId(istAuswahlLeer(kassierenAuswahl.mengen))

  // 409 `vorgang_daten_abweichend`: Der Vorgang unter diesem Schlüssel ist
  // gebucht, nur seine Antwort ging verloren — er ist damit abgeschlossen, auch
  // wenn die zuletzt gesendete Auswahl eine andere war. Die Auswahl wird geleert
  // (erst der Leerzustand rotiert den Schlüssel) und der Tischzustand neu
  // geladen. Ohne beides folgte auf jeden weiteren Versuch derselbe 409: Die
  // Meldung rät zur Differenz, aber mit unverändertem Schlüssel und
  // unverändertem Server-Hash bleibt sie unbuchbar.
  const bestellungBereitsGebucht = () => {
    bestellKorb.reset()
    reload()
  }
  const zahlungBereitsGebucht = () => {
    kassierenAuswahl.reset()
    reload()
  }

  // Expliziter Fehlerzustand statt der Leer-Defaults (Saldo 0,00 €) — sonst
  // wirkt der Tisch bei Netzabbruch abgerechnet. Nur beim gescheiterten
  // Erstladen: Scheitert ein Hintergrund-Refetch, bleibt der zuletzt geladene
  // Tischzustand stehen, statt eine geöffnete Ansicht wegzureißen.
  // Nur der Tischzustand trägt die ganze Seite; ein Ladefehler der Historie
  // ersetzt allein den Historie-Tab (siehe historieInhalt), damit Bestellen und
  // Kassieren bedienbar bleiben.
  if (stateError) {
    return (
      <LadefehlerAlert
        titel="Tischdaten konnten nicht geladen werden"
        onErneutVersuchen={reload}
      />
    )
  }

  const tisch = {
    id: state.tischId,
    name: state.tischName,
    saldoCents: state.saldoCents,
  }

  const tabsLocked = stateLoading || historieLoading
  const anzahlUnbezahlt = state.unbezahltePositionen.length

  const kopf = (
    <div className="flex items-start justify-between gap-4">
      <div>
        {stateLoading ? (
          <Skeleton className="h-7 w-40" />
        ) : (
          <h1 className="font-heading text-[22px] font-semibold leading-tight">
            {tisch.name}
          </h1>
        )}
        {!stateLoading && (
          <div className="mt-1.5 flex flex-wrap items-center gap-2">
            <TischStatusBadge anzahlUnbezahlt={anzahlUnbezahlt} />
            <span className="text-sm text-muted-foreground">
              {state.fuerMichErledigt
                ? 'Für dich erledigt'
                : 'Für dich noch offen'}
            </span>
          </div>
        )}
      </div>
      <div className="text-right">
        <div className="text-[11px] font-medium uppercase tracking-[0.04em] text-muted-foreground">
          Offen
        </div>
        {stateLoading ? (
          <Skeleton className="ml-auto h-7 w-20" />
        ) : (
          <div
            data-slot="tisch-saldo"
            className="text-xl font-bold tabular-nums"
          >
            {formatEuro(animierterSaldo)}
          </div>
        )}
      </div>
    </div>
  )

  const tabsLockedHinweis = tabsLocked && (
    <p className="rounded-md border bg-background/90 px-3 py-1 text-center text-xs text-muted-foreground">
      Lade Tischdaten. Tabs sind kurzzeitig deaktiviert.
    </p>
  )

  const tabsList = (
    <TabsList className="h-10 w-full">
      <TabsTrigger value="order" className="flex-1" disabled={tabsLocked}>
        Bestellen
      </TabsTrigger>
      <TabsTrigger value="payment" className="flex-1" disabled={tabsLocked}>
        Kassieren
      </TabsTrigger>
      <TabsTrigger value="history" className="flex-1" disabled={tabsLocked}>
        Historie
      </TabsTrigger>
    </TabsList>
  )

  const bestellenInhalt = !stateLoading && (
    <Bestellung
      backend={tischBackend}
      tisch={tisch}
      products={produkte}
      productsLoading={isPending}
      productsError={produkteError}
      onErneutVersuchen={() => {
        void reloadProdukte()
      }}
      mengenSteuerung={bestellKorb}
      bestellungId={bestellungId}
      onErfolg={zeigeErfolg}
      onVorgangBereitsGebucht={bestellungBereitsGebucht}
    />
  )
  const kassierenInhalt = !stateLoading && (
    <Zahlung
      backend={tischBackend}
      tisch={tisch}
      positionen={state.unbezahltePositionen}
      mengenSteuerung={kassierenAuswahl}
      vorgangId={zahlungVorgangId}
      onErfolg={zeigeErfolg}
      onVorgangBereitsGebucht={zahlungBereitsGebucht}
    />
  )
  const historieInhalt = !stateLoading && (
    <TischHistorie
      historie={historie}
      historieLoading={historieLoading}
      historieError={historieError}
      onErneutVersuchen={reload}
      tisch={tisch}
      backend={tischBackend}
      onErfolg={zeigeErfolg}
    />
  )

  return (
    <>
      {isMobile ? (
        <>
          {kopf}
          <Tabs defaultValue="order" className="mt-4">
            <ServiceDock
              leiste={
                <>
                  {tabsLockedHinweis}
                  {tabsList}
                </>
              }
            >
              <TabsContent value="order" className={dockFreiraum}>
                {bestellenInhalt}
              </TabsContent>
              <TabsContent value="payment" className={dockFreiraum}>
                {kassierenInhalt}
              </TabsContent>
              <TabsContent value="history" className={dockFreiraum}>
                {historieInhalt}
              </TabsContent>
            </ServiceDock>
          </Tabs>
        </>
      ) : (
        // Höhenbegrenzte Flex-Spalte (Viewport minus Header und Content-Padding
        // aus ServiceLayout); Kopf und Reiter-Zeile ergeben sich per Flex, der
        // aktive Tab füllt via flex-1 den Rest und scrollt in sich bzw. seinen
        // Spalten.
        <Tabs
          defaultValue="order"
          className="flex h-[calc(100dvh-5.5rem)] flex-col xl:h-[calc(100dvh-6.5rem)]"
        >
          {kopf}
          <div className="mt-4 mb-4 max-w-md space-y-1">
            {tabsLockedHinweis}
            {tabsList}
          </div>
          <TabsContent value="order" className="min-h-0 flex-1">
            {bestellenInhalt}
          </TabsContent>
          <TabsContent value="payment" className="min-h-0 flex-1">
            {kassierenInhalt}
          </TabsContent>
          <TabsContent
            value="history"
            className="min-h-0 flex-1 overflow-y-auto"
          >
            {historieInhalt}
          </TabsContent>
        </Tabs>
      )}
      <ErfolgsPop
        open={erfolg.open}
        text={erfolg.text}
        onDismiss={erfolgSchliessen}
      />
    </>
  )
}
