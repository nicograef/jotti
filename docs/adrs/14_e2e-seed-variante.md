# ADR 14: Ein Demo-Drehbuch für die E2E-Suite, keine zweite Seed-Variante

- **Status:** akzeptiert (2026-09-09)
- **Kontext-Dokumente:** `docs/plans/plan-jotti-audit-fixes.md` Phase 14 (nach
  Merge gelöscht, siehe Git-Historie);
  `e2e/tests/kassenabschluss.mobile.spec.ts`; `backend/seed/`

## Kontext

`kassenabschluss.mobile.spec.ts` (80 Zeilen) prüft eine Zusicherung: nach dem
Kassenabschluss erscheint die Meldung „Kasse abgeschlossen." (Zeile 78). Um
dorthin zu kommen, gleicht die Spec zuerst jeden Tisch mit offenem Saldo aus —
der Abschluss ist sonst gesperrt. Dafür hebt sie ihr Zeitbudget auf 120 s
(Zeile 37). Sie ist die einzige Spec, die das tut; global gelten 60 s
(`e2e/playwright.config.ts:18`). Die Suite läuft seriell (`fullyParallel: false`,
`workers: 1`, Zeilen 24–25), das Budget zählt also voll zur Laufzeit.

Vorgeschlagen ist eine zweite Seed-Variante mit ausgeglichenen Tischen, damit
die Klickstrecke entfällt.

### Was die Spec braucht

- eine laufende Kassensitzung (die Spec erwartet „Kassentag Nr. 3", Zeile 60),
- frischen Umsatz aus einer vollständigen Runde am Tisch (Zeilen 42–45),
- jeden Tisch ausgeglichen, damit das Gate `tische_saldo_offen` nicht greift,
- einen Ist-Bestand zum Eintippen (342,00 €, Zeile 64).

### Wie heute geseedet wird

`backend/seed` (11 Dateien, 4065 Zeilen) kennt genau ein Drehbuch:
`demoSzenario()` in `szenario.go` (846 Zeilen), das
„3-Tage-Sommerfest TSV Musterstadt e.V." — Freitag und Samstag abgeschlossen,
Sonntag offen. Zwei Einstiege schreiben es: `Run` (`writer.go:22`) für
`make seed` mit Kassenjournal-Guard und `ResetAndSeed` (`writer.go:31`), das
zuerst leert. Beide rufen dieselbe `seedInTransaction` und damit dasselbe
Drehbuch.

Die E2E-Umgebung nimmt den zweiten Weg: `resetAndSeed` in `e2e/support/seed.ts`
schickt `POST /api/test/reset-and-seed` (Zeile 26) und liest die Zugangsdaten
aus der Antwort. Die Route existiert nur bei `JOTTI_ENABLE_TEST_API=1`
(`backend/app/app.go:60`, `backend/config/config.go:64`), gesetzt in
`docker-compose.e2e.yml:69`. Alle 25 Spec-Dateien der Suite rufen
`resetAndSeed`.

### Die offenen Tische sind Absicht

Das Drehbuch führt 22 Tische (`szenario.go:320-341`): 20 aktiv, einer inaktiv,
einer gelöscht. Der offene Sonntag lässt neun der aktiven Tische mit einem
Saldo ungleich null zurück — teilbezahlt (Tisch 2), frisch bestellt (6, 14),
mehrere offene Bestellungen (8, 11), Nachbestellung nach abgeschlossenen Runden
(16, 17, 20) und die auf Tisch 4 umgebuchte Bestellung (`szenario.go:751-841`).
`offeneTischNamen` (`e2e/support/servicekraft.ts:211`) liest alle aktiven Tische
aus dem „Alle Tische"-Drawer und gibt die mit offenem Saldo zurück;
`settleAlleOffenenTische` (Zeile 314) kassiert sie der Reihe nach.

Diese Zustandsvielfalt ist festgeschrieben: `backend/seed/engine_test.go:147`
(`TestBuildSeedDaten_SonntagsTischZustaende`) prüft leere, frisch bestellte,
mehrfach bestellte, teilbezahlte, warenrückgenommene und abgeschlossene Tische
einzeln.

Zwei weitere Specs hängen daran:

- `admin-kassenfuehrung.spec.ts:51-54` erwartet, dass der Abschluss mit „Es gibt
  noch offene Tische mit ausstehenden Beträgen." blockiert wird.
- `admin-live-reporting.spec.ts:41-43` erwartet die Überschrift „Offene Tische"
  im Live-Reporting.

Ein ausgeglichenes Drehbuch macht beide Zusicherungen unprüfbar.

### Erwogene Alternativen

1. **Zweite Seed-Variante mit ausgeglichenen Tischen** (der Vorschlag). Braucht
   ein zweites Drehbuch neben `demoSzenario()` und einen Schalter am
   Reset-Endpoint. Der Schalter steht in Produktionscode, auch wenn die Route
   nur bei `JOTTI_ENABLE_TEST_API=1` existiert.
2. **Das bestehende Drehbuch ausgleichen.** Bricht
   `admin-kassenfuehrung.spec.ts`, `admin-live-reporting.spec.ts` und
   `TestBuildSeedDaten_SonntagsTischZustaende`.
3. **Die Vorbedingung über die API herstellen statt über die UI.** Die Spec hat
   einen `APIRequestContext`, aber das Token entsteht bei der UI-Anmeldung. Die
   Suite bräuchte einen zweiten Anmeldeweg und müsste die Nutzlast von
   `/service/zahlung-kassieren` kennen — Domänenwissen, das heute nur im
   Frontend steht.
4. **Alles lassen.** Eine Spec trägt ein doppeltes Zeitbudget und eine
   Klickstrecke, die dem realen Abendabschluss entspricht.

## Entscheidung

**Die E2E-Suite behält ein einziges Demo-Drehbuch; die Spec gleicht die Tische
weiter über die Oberfläche aus** (Alternative 4).

- **Korrektheit:** Kein Vorschlag behebt einen Fehler. Alternative 2 nimmt zwei
  Zusicherungen weg. Alternative 1 bringt zwei Drehbücher, die auseinanderlaufen
  können: der Sonntag der einen Variante ist dann ein anderer als der der
  anderen, und welche eine Spec meint, steht nirgends im Testcode.
- **Einfachheit:** Ein Drehbuch, ein Reset-Endpoint ohne Parameter, ein
  Ausgangszustand für alle 25 Spec-Dateien. Das ist die einfachste Form, die
  die Suite haben kann.
- **Konsistenz:** Jede Spec startet heute mit demselben Aufruf und demselben
  Zustand. Ein Parameter macht aus dieser einen Zeile eine Entscheidung, die
  jede künftige Spec treffen muss.
- **Produkt-Konservatismus:** Der Vorschlag baut Produktionscode um, um eine
  Testlaufzeit zu senken. Die Klickstrecke ist zudem der Abendabschluss, wie
  ein Verein ihn wirklich durchführt — alle Tische kassieren, dann die Kasse
  schließen.

**Empfehlung: nie.**

## Konsequenzen

- `POST /api/test/reset-and-seed` bleibt parameterlos. Wer einen anderen
  Ausgangszustand braucht, stellt ihn in der Spec her, wie
  `kassenabschluss.mobile.spec.ts` es tut.
- `settleAlleOffenenTische` bleibt der einzige Helfer mit dieser Aufgabe und
  hat weiterhin genau einen Aufrufer.
- Das Zeitbudget von 120 s bleibt die einzige Ausnahme von den globalen 60 s.
  Es ist eine Obergrenze, keine gemessene Laufzeit.
- **Wieder aufgreifen, wenn** eine zweite Spec einen ausgeglichenen
  Ausgangszustand braucht, oder wenn diese Spec ihr Budget in der CI
  tatsächlich ausschöpft. Dann ist der erste Schritt, `settleAlleOffenenTische`
  schneller zu machen, und erst der zweite ein zweites Drehbuch — und dieses
  bräuchte ein eigenes ADR mit der Antwort darauf, wie beide Drehbücher
  gemeinsam gepflegt werden.
