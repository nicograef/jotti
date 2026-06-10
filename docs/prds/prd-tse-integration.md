# PRD: TSE-Integration (F-02)

> Referenz-Anforderung: F-02 (`docs/anforderungen.md`)
> Fachliche Grundlagen: `docs/compliance.md` §3, `docs/handbuch.md` §3.13, `docs/steuerrecht.md` §6
> Vorarbeit: `docs/prds/prd-kassenbeleg.md` (F-03/F-07) hat den beleg-seitigen Einhängepunkt (`TSEAbschnitt`, optionaler Pointer auf die Kassenbeleg-Daten) bereits vorbereitet.

## Problem Statement

jotti ist ein elektronisches Aufzeichnungssystem im Sinne von § 1 KassenSichV und unterliegt damit der TSE-Pflicht nach § 146a AO. Heute werden alle kassenwirksamen Vorgänge (Bestellung, Zahlung, Stornierung, Auszahlung, Direktverkauf, Geldtransit, Kassendifferenz, Tagesabschluss) zwar lückenlos und unveränderlich im Kassenjournal erfasst, aber **nicht kryptografisch durch eine zertifizierte Technische Sicherheitseinrichtung abgesichert**. Ohne diese Absicherung ist ein realer Echtbetrieb eines Vereins rechtlich nicht zulässig, und der Kassenbeleg kann die nach § 6 KassenSichV vorgeschriebenen TSE-Pflichtfelder (Transaktionsnummer, Signaturzähler, TSE-Seriennummer, Zeitpunkte, Signatur) nicht ausweisen.

Gleichzeitig soll jotti **ohne** TSE voll funktionsfähig bleiben — zum Testen, Ausprobieren, Üben, für Demos und für Einsatzszenarien, die keine TSE erfordern. Die rechtliche Verantwortung für den Einsatz einer TSE liegt beim Betreiber (Verein); jotti als Hersteller muss die Anbindungs**möglichkeit** bereitstellen (§ 146a Abs. 1 Satz 5 AO i.V.m. § 379 AO).

Aus Entwicklersicht fehlt heute ein klar definiertes, anbieter-agnostisches Interface, gegen das die kassenwirksamen Abläufe entwickelt werden können, sowie eine konkrete, zertifizierte Implementierung (fiskaly Cloud-TSE).

## Solution

jotti erhält eine vollständige TSE-Integration nach dem **Festzelt-Muster** (atomare, sofort geschlossene TSE-Transaktionen, `docs/compliance.md` §3.6):

- Ein anbieter-agnostisches **`TSEClient`-Interface** (`StartTransaction` / `UpdateTransaction` / `FinishTransaction`) gemäß `docs/handbuch.md` §3.13 kapselt die TSE-Kommunikation.
- Eine **fiskaly-Implementierung** (`FiskalyTSEClient`) gegen die fiskaly **SIGN DE API v2** ist die erste konkrete Adapter-Implementierung.
- Jeder der **neun fiskalisch relevanten Vorgänge** wird beim Schreiben synchron signiert (**sign-then-persist**): Die TSE-Transaktion wird gestartet und sofort abgeschlossen, und das Ergebnis (Signatur, Transaktionsnummer, Signaturzähler, TSE-Seriennummer, Zeitpunkte) wird in das jeweilige Event eingebettet, bevor das Event im Kassenjournal landet.
- **TSE ist optional.** Ist keine TSE konfiguriert, läuft jotti unverändert weiter (keine Signatur, kein TSE-Block auf dem Beleg). Der Admin sieht einen deutlichen Hinweis und eine Warnung, solange keine TSE konfiguriert ist.
- **Die Kasse blockiert nie.** Ist eine TSE konfiguriert, fiskaly aber gerade gestört, wird der Verkauf trotzdem abgeschlossen: Das Event wird ohne Signatur persistiert, der Ausfall wird auf dem Beleg vermerkt, und ein Hintergrund-Worker **signiert den Vorgang automatisch nach** (idempotent über eine deterministische Transaktions-ID).
- Der Betreiber bringt seine eigene TSE mit (**BYOT**): Er legt TSS und Client im fiskaly-Dashboard an und trägt `api_key`, `api_secret`, `tss_id` und `client_id` über die Admin-UI in jotti ein. jotti **signiert nur** — es betreibt keinen TSS-Lebenszyklus.
- Der **Kassenbeleg** weist die TSE-Pflichtfelder als Text und zusätzlich als **DSFinV-K-QR-Code** aus.

Servicekräfte und Serviceleitung merken im Normalbetrieb nichts von der TSE — sie bestellen, kassieren, stornieren und zahlen aus wie bisher. Der Gast erhält auf Anforderung einen vollständigen, fiskalisch signierten Kassenbeleg. Der Betriebsprüfer kann jeden Vorgang lückenlos und signiert rekonstruieren.

## User Stories

### Betreiber / Admin — Konfiguration

1. Als Admin möchte ich `api_key` und `api_secret` meiner fiskaly-Organisation in jotti hinterlegen, damit jotti sich gegenüber der fiskaly-API authentifizieren kann.
2. Als Admin möchte ich die `tss_id` und `client_id` meiner im fiskaly-Dashboard angelegten TSS bzw. Client in jotti eintragen, damit jotti Transaktionen gegen die richtige TSE signiert.
3. Als Admin möchte ich jottis Kassenseriennummer (F-01) prominent angezeigt bekommen, damit ich sie im fiskaly-Dashboard als `serial_number` des Clients verwenden kann (Konsistenz für DSFinV-K).
4. Als Admin möchte ich die TSE-Konfiguration jederzeit ändern oder leeren können, damit ich Schlüssel rotieren oder von TEST auf LIVE wechseln kann.
5. Als Admin möchte ich sehen, ob jotti aktuell mit einer TEST- oder LIVE-TSE verbunden ist, damit ich nicht versehentlich einen Echtbetrieb gegen die Test-Umgebung fahre.
6. Als Admin möchte ich die Verbindung zur TSE testen können (z. B. Authentifizierung + Status der TSS abrufen), damit ich die Konfiguration vor dem Fest verifizieren kann.
7. Als Admin möchte ich einen deutlichen Hinweis im Dashboard sehen, wenn keine TSE konfiguriert ist, damit mir bewusst ist, dass jotti gerade nicht fiskalkonform aufzeichnet.
8. Als Admin möchte ich eine Warnung sehen, die erklärt, dass der Betrieb ohne TSE meine rechtliche Verantwortung ist und nur für Test/Demo/Übung gedacht ist, damit ich eine informierte Entscheidung treffe.
9. Als Admin möchte ich sehen, wie viele Vorgänge aktuell auf Nachsignierung warten (TSE-Ausfall), damit ich erkenne, ob die TSE-Anbindung gesund ist.

### Servicekraft / Serviceleitung — Betrieb

10. Als Servicekraft möchte ich Bestellungen, Zahlungen und Ausgaben unverändert aufnehmen können, egal ob eine TSE konfiguriert ist oder nicht, damit mein Arbeitsablauf gleich bleibt.
11. Als Serviceleitung möchte ich Stornierungen und Auszahlungen unverändert durchführen können, auch wenn die TSE gerade gestört ist, damit der Betrieb nie blockiert.
12. Als Servicekraft möchte ich, dass ein Verkauf auch bei einem fiskaly-Ausfall sofort abgeschlossen wird, damit keine Warteschlange am Tisch entsteht.
13. Als Servicekraft möchte ich auf Anforderung einen Kassenbeleg drucken, der die TSE-Pflichtfelder enthält, damit der Gast einen gültigen Beleg erhält.
14. Als Servicekraft möchte ich, dass ein Beleg bei TSE-Ausfall einen klaren Ausfallvermerk trägt, damit Gast und Prüfer den Sachverhalt nachvollziehen können.

### Gast — Beleg

15. Als Gast möchte ich einen Kassenbeleg mit Transaktionsnummer, Signaturzähler, TSE-Seriennummer und Zeitpunkten erhalten, damit ich einen gesetzeskonformen Beleg habe.
16. Als Gast möchte ich einen QR-Code auf dem Beleg, der alle TSE-Daten maschinenlesbar bündelt, damit der Beleg ohne Abtippen prüfbar ist.

### Betriebsprüfer — Nachvollziehbarkeit

17. Als Betriebsprüfer möchte ich, dass jeder kassenwirksame Vorgang einer eigenständigen, sofort geschlossenen TSE-Transaktion entspricht, damit das Festzelt-Muster (DSFinV-K Nr. 2.7) eingehalten ist.
18. Als Betriebsprüfer möchte ich auf jedem Zahlungsbeleg den Startzeitpunkt der ersten Bestellung des Tisches in Klarschrift sehen (Durchbedienen, `docs/compliance.md` §5.3), damit der Tischvorgang rekonstruierbar ist.
19. Als Betriebsprüfer möchte ich, dass nachsignierte Vorgänge (nach einem Ausfall) dieselbe fiskalische Aussagekraft haben wie sofort signierte, damit kein Vorgang ungesichert bleibt.

### Entwickler / System — Architektur

20. Als Entwickler möchte ich ein anbieter-agnostisches `TSEClient`-Interface, damit ich gegen eine stabile Schnittstelle entwickeln kann, ohne mich an fiskaly zu binden.
21. Als Entwickler möchte ich einen Test-Doppelgänger (Fake) des `TSEClient`, damit ich die kassenwirksamen Abläufe ohne echten fiskaly-Zugang testen kann.
22. Als System möchte ich für jeden fiskalischen Vorgang den korrekten `processType` (`Bestellung-V1`, `Kassenbeleg-V1`, `SonstigerVorgang-V1`) gemäß `docs/handbuch.md` §3.13 verwenden, damit die TSE-Transaktion korrekt typisiert ist.
23. Als System möchte ich die `processData` exakt nach `docs/compliance.md` §3.4 formatieren (`Beleg^…`-String, Punkt als Dezimaltrennzeichen, kombi 70/30 in ermäßigt/normal), damit fiskaly die Daten akzeptiert.
24. Als System möchte ich eine deterministische TSE-Transaktions-ID pro Vorgang ableiten, damit ein Nachsignieren idempotent denselben fiskaly-Vorgang trifft und keine Dubletten entstehen.
25. Als System möchte ich das fiskaly-Zugriffstoken zwischenspeichern und bei Ablauf/401 erneuern, damit nicht vor jeder Signatur neu authentifiziert wird.
26. Als System möchte ich bei wiederholbaren fiskaly-Fehlern (5xx, 499, 429) mit exponentiellem Backoff erneut versuchen, damit transiente Störungen die Signatur nicht unnötig in den Ausfallpfad zwingen.
27. Als System möchte ich die TSE-Ergebnisfelder in die bestehenden Event-Daten der neun fiskalischen Events einbetten, damit die Signatur Teil des unveränderlichen Kassenjournals wird (Happy Path).
28. Als System möchte ich bei einem TSE-Ausfall den Vorgang ohne Signatur persistieren und in eine Nachsignier-Warteschlange stellen, damit der Verkauf nie blockiert.
29. Als System möchte ich einen Hintergrund-Worker, der offene Nachsignier-Aufträge abarbeitet und die erhaltene Signatur in einer Seitentabelle (per Transaktions-ID verknüpft) ablegt, damit auch ausgefallene Vorgänge fiskalisch abgesichert werden, ohne das immutable Event zu verändern.
30. Als System möchte ich beim Beleg-Druck die Signatur entweder aus dem Event (Happy Path) oder aus der Seitentabelle (nachsigniert) beziehen, damit der Beleg unabhängig vom Signierzeitpunkt vollständig ist.
31. Als System möchte ich den Startzeitpunkt der ersten Bestellung je Tisch-Session festhalten, damit der Zahlungsbeleg die Durchbedienen-Klarschrift ausweisen kann.

## Implementation Decisions

### Schnittstelle und Adapter

- **`TSEClient`-Port** exakt nach `docs/handbuch.md` §3.13: `StartTransaction`, `UpdateTransaction`, `FinishTransaction` mit den Rückgabetypen `StartResult` (Transaktionsnummer, LogTime, TSE-Seriennummer, Signaturzähler) und `FinishResult` (Signatur, LogTime, Signaturzähler). `UpdateTransaction` wird im atomaren Festzelt-Muster **nicht** benutzt, ist aber Teil des Interfaces (Vollständigkeit, künftige Nutzung).
- **`FiskalyTSEClient`-Adapter** gegen fiskaly **SIGN DE API v2**:
  - Authentifizierung per `POST /api/v2/auth` (`api_key` + `api_secret`) → Bearer-JWT (bis 24 h gültig), Token wird gecacht und bei Ablauf/`401` erneuert.
  - Signatur per `upsertTransaction` (`PUT …/tss/{tss_id}/tx/{tx_id}?tx_revision=N`): Start mit `state=ACTIVE`, `tx_revision=1`; Abschluss mit `state=FINISHED`, `tx_revision=2`. Typ und Daten werden über das **`raw`-Schema** (`process_type` + `process_data`) übergeben.
  - Antwort-Mapping: `number` → Transaktionsnummer, `tss_serial_number` → TSE-Seriennummer, `signature.counter` → Signaturzähler (Implementierung akzeptiert Zahl **und** String — bekannter fiskaly-Hinweis), `signature.value` → Signatur, `log.timestamp` / `time_start` / `time_end` → Zeitpunkte, `qr_code_data` → QR-Inhalt für den Beleg.
  - Eine einzige Middleware-Basis-URL für alle Calls; TEST/LIVE ergibt sich aus den Credentials (im Token-Claim `env` sichtbar) und wird dem Admin angezeigt.
  - Wiederholbare Fehler (`5xx`, `499`, `429` mit `Retry-After`) werden mit exponentiellem Backoff erneut versucht; die fiskaly-API ist idempotent.

### Provisionierung (BYOT)

- Der Betreiber legt **TSS und Client im fiskaly-Dashboard** an. jotti betreibt **keinen** TSS-Lebenszyklus (kein `createTss`/`createClient`, kein Admin-PIN/PUK). Damit bleibt der Adapter minimal (nur Auth + Signieren) und es liegt **kein** mächtiges Admin-Secret in jottis Datenbank.
- jotti hält in der Konfiguration genau: `api_key`, `api_secret`, `tss_id`, `client_id`.
- Der Betreiber verwendet jottis **Kassenseriennummer** (F-01) als `serial_number` des fiskaly-Clients, damit ERS-Seriennummer und DSFinV-K-Ausweis konsistent sind.

### TSE optional + Admin-Hinweis

- **Kein** `strict`/`bypass`-Schalter. Ist keine TSE konfiguriert, läuft jotti vollständig weiter, signiert nicht und druckt keinen TSE-Block. Ist eine TSE konfiguriert, werden alle fiskalischen Vorgänge signiert.
- Das Admin-Dashboard zeigt einen **Hinweis + Warnung**, solange keine TSE konfiguriert ist (Betreiberverantwortung, nur für Test/Demo/Übung).

### Sign-then-persist + Ausfallverhalten

- Für jeden fiskalischen Vorgang ruft die Anwendungsschicht **synchron** `StartTransaction` und unmittelbar `FinishTransaction` auf, **bevor** das Event geschrieben wird (Festzelt-Muster: jede Transaktion sofort geschlossen).
- **Happy Path:** Das `FinishResult` (+ `StartResult`-Felder) wird in die Event-Daten eingebettet und mit dem Event im Kassenjournal transaktional persistiert. Die Signatur ist damit Teil des unveränderlichen Events.
- **Ausfall-Path (DON'T BLOCK THE TILL):** Schlägt die TSE-Kommunikation fehl, wird der Vorgang **ohne** Signatur persistiert, auf dem Beleg ein **Ausfallvermerk** gedruckt und der Vorgang in eine **Nachsignier-Warteschlange** gestellt.
- **Deterministische Transaktions-ID:** Pro fiskalischem Event wird eine stabile TSE-`tx_id` aus der Event-Identität abgeleitet, sodass Start/Finish und ein späteres Nachsignieren denselben fiskaly-Vorgang treffen (Idempotenz, keine Dubletten).
- **Nachsignier-Worker:** Ein Hintergrundprozess arbeitet die Warteschlange ab (idempotenter `upsertTransaction`) und legt die erhaltene Signatur in einer **Seitentabelle** ab, die per `tx_id` mit dem Event verknüpft ist. Das immutable Event bleibt unverändert.
- **Signatur-Bezug beim Beleg:** Der Beleg-Druck bezieht die TSE-Felder entweder aus dem Event (Happy Path) oder aus der Seitentabelle (nachsigniert).

### Event-Abdeckung und Typ-Mapping

- Alle **neun** fiskalisch relevanten Events erhalten TSE-Felder und werden signiert (`docs/handbuch.md` §3.13):
  - `bestellung-aufgenommen:v1` → `Bestellung-V1`
  - `zahlung-kassiert:v1` → `Kassenbeleg-V1`
  - `stornierung-erteilt:v1` → `Kassenbeleg-V1`
  - `auszahlung-geleistet:v1` → `Kassenbeleg-V1`
  - `direktverkauf-getaetigt:v1` → `Kassenbeleg-V1`
  - `direktverkauf-storniert:v1` → `Kassenbeleg-V1`
  - `geldtransit-gebucht:v1` → `SonstigerVorgang-V1`
  - `differenz-soll-ist-gebucht:v1` → `SonstigerVorgang-V1`
  - `tagesabschluss-erstellt:v1` → `SonstigerVorgang-V1`
- **Nicht** signiert: `ausgabe-bestaetigt:v1` (keine Geld-/Umsatzrelevanz). `kassensitzung-eroeffnet:v1` (Anfangsbestand) wird bewusst ausgeklammert; siehe Annahme unten.
- Die TSE-Ergebnisfelder werden als zusätzliches Struct (`TSEData`, Felder nach `docs/handbuch.md` §3.13) an die Event-Daten der neun Events angehängt. Event-Data-Structs dürfen JSON-Tags tragen (dokumentierte Ausnahme der Domain-Regel).

### processData-Formatter

- Eine **reine Funktion** erzeugt die `processData` nach `docs/compliance.md` §3.4:
  - `Kassenbeleg-V1`: `Beleg^<Normal>_<Ermäßigt>_<Null>_<BesondererSatz>_<Befreit>^<Zahlbetrag>:<Zahlungsart>`. Kombi-Positionen (70/30) fließen anteilig in „Ermäßigt" (Speisen) und „Normal" (Getränke). Die Brutto-/Netto-/Steuer-Aufteilung kommt aus der bestehenden Steuer-Aufteilungslogik der Domäne (`domain/steuer`), nicht aus einer eigenen Rechnung.
  - `Bestellung-V1`: Positionen als strukturierter Text.
  - `SonstigerVorgang-V1`: vorgangsspezifischer Inhalt (Geldtransit, Differenz, Tagesaggregat).
- Formatregeln strikt: Punkt als Dezimaltrennzeichen, keine Tausendertrenner, keine Exponentialschreibweise, mindestens eine Stelle vor dem Punkt, UTF-8 ohne BOM.

### Beleg-Erweiterung

- Der in der Kassenbeleg-Vorarbeit angelegte optionale `TSEAbschnitt` wird aus den Signaturdaten gefüllt: Transaktionsnummer, Signaturzähler, TSE-Seriennummer, Zeitpunkt Vorgangsbeginn/-ende, Signatur.
- Zusätzlich wird der **DSFinV-K-QR-Code** (`qr_code_data` aus `FinishTransaction`) gedruckt. Dafür ist eine ESC/POS-QR-Code-Fähigkeit im Bondruck-Formatter nötig.
- Der Zahlungsbeleg trägt den **Startzeitpunkt der ersten Bestellung** der Tisch-Session in Klarschrift (Durchbedienen, `docs/compliance.md` §5.3). Dazu wird der LogTime der ersten `Bestellung-V1` je Tisch-Session festgehalten.

### Schema-Änderungen

Da jotti in aktiver Entwicklung ist, werden Schemaänderungen direkt in der initialen Migration vorgenommen (keine neuen Migrationsdateien):

- Neue **Singleton-Tabelle** für die TSE-Konfiguration (`api_key`, `api_secret`, `tss_id`, `client_id`, `updated_at`), plaintext (Verschlüsselung-at-rest später), Muster wie `betreiber`/`bondruck_einstellungen`.
- Neue **Nachsignier-Warteschlange** (Outbox) für ausgefallene Signierungen, mit deterministischer `tx_id`, `process_type`, `process_data`, Status `offen`/`erledigt` — Muster wie `druckauftraege`.
- Neue **Signatur-Seitentabelle**, per `tx_id` mit dem Event verknüpft, für nachsignierte Vorgänge.
- `tisch_sessions` erhält den Startzeitpunkt der ersten Bestellung (für die Durchbedienen-Klarschrift).

### API und Frontend

- Neue **POST-only** Admin-Endpunkte: TSE-Konfiguration lesen, TSE-Konfiguration setzen/leeren, Verbindung testen, Nachsignier-Status abrufen. JSON-Keys und Response-DTOs in der HTTP-Schicht; Domain-Modelle werden nie direkt serialisiert.
- Validierung beidseitig: `zog` (Backend) + `Zod` (Frontend), deutsche Fehlermeldungen.
- Frontend: Admin-Seite für die TSE-Konfiguration (über eine `BackendClient`-basierte Backend-Klasse, kein direktes `fetch`), plus Dashboard-Hinweis/Warnung bei fehlender TSE. Alle benutzersichtbaren Strings deutsch.

### Abgrenzung des Adapters

- Der Adapter implementiert ausschließlich den `TSEClient`-Port. Die Orchestrierung (sign-then-persist, Ausfallpfad, Nachsignieren) liegt im TSE-Application-Service, nicht im Adapter. So bleibt der Adapter ein dünner Übersetzer (HTTP/JSON ↔ Domain) und der Service das tiefe, testbare Modul.

## Testing Decisions

**Was einen guten Test ausmacht:** Getestet wird **externes Verhalten**, nicht die interne Umsetzung. Tests prüfen beobachtbare Ergebnisse (welche TSE-Felder im Event/Beleg landen, welcher `processData`-String erzeugt wird, was bei einem Ausfall passiert), nicht private Hilfsfunktionen oder exakte Byte-Anordnungen. Externe HTTP-Aufrufe werden gegen einen lokalen Mock-Server geprüft, nicht gegen die echte fiskaly-API.

**Zu testende Module:**

- **TSE-Application-Service (Pflicht):** Verhalten von sign-then-persist und Ausfallpfad. Happy Path → TSE-Felder im Event; TSE-Fehler → Event ohne Signatur + Eintrag in der Nachsignier-Warteschlange + Ausfallvermerk, Verkauf wird **nicht** blockiert. Der Service wird gegen einen Fake-`TSEClient` getestet (Erfolg, Fehler, Timeout). Vorbild: bestehende Command-Level-Tests im Kasse-/Table-Kontext.
- **`FiskalyTSEClient`-Adapter (Pflicht):** Übersetzung gegen einen **Mock-HTTP-Server**: Auth-Token-Erwerb und -Caching, Start/Finish-Aufrufe mit korrektem `state`/`tx_revision`, Mapping der Antwortfelder auf `StartResult`/`FinishResult`, Retry-Verhalten bei wiederholbaren Fehlern. Vorbild: vorhandene HTTP-Adapter-/Handler-Tests im Backend.
- **processData-Formatter (Pflicht):** Reine, table-driven Unit-Tests über repräsentative Vorgänge — einzelner Satz, gemischte Sätze, kombi 70/30 in ermäßigt/normal, befreit (0 %), Stornobetrag (negativ), korrekte Formatregeln (Punkt, keine Tausendertrenner). Vorbild: die bestehenden Tests der Steuer-Aufteilungslogik (`steuer_test.go`).
- **Event-Data TSE-Feld-Round-Trip (Pflicht):** Marshal/Unmarshal der erweiterten Event-Daten — ein Event mit TSE-Feldern serialisiert und deserialisiert verlustfrei; ein Event ohne TSE-Felder (kein TSE konfiguriert) bleibt gültig. Vorbild: bestehende Event-/Domain-Tests im Kasse-Kontext.
- **Nachsignier-Worker (Pflicht):** Outbox-Drain und Idempotenz — ein offener Auftrag wird signiert und als erledigt markiert; ein erneuter Lauf gegen denselben `tx_id` erzeugt keine Dublette; die Signatur landet in der Seitentabelle. Vorbild: das Outbox-Muster der Druckauftrags-Verarbeitung.

**Regressionsschutz:** Alle bestehenden Beleg-, Command- und Handler-Tests bleiben grün. Insbesondere bleibt der Beleg ohne konfigurierte TSE **byte-identisch** zum heutigen Zustand (kein TSE-Block, kein QR), abgesichert durch die bestehenden Kassenbeleg-Tests.

## Out of Scope

- **DSFinV-K-Export (F-04):** Der maschinenlesbare ZIP-/CSV-Export bleibt eigenes Feature. **Ausnahme:** Der QR-Code auf dem Beleg ist hier enthalten.
- **ELSTER-Meldepflicht (F-05):** Manuelle Anleitung und programmatische Meldung (ERiC / fiskaly SUBMIT DE) sind nicht Teil dieser PRD.
- **eBeleg / digitaler Beleg (F-09):** PDF/HTML-Download und Beleg-Archivierung; lediglich der QR-Code wird gedruckt.
- **GoBD-Hash-Chain (F-08):** Eigenständige Maßnahme, unabhängig von der TSE-Signatur.
- **TSS-Lebenszyklus-Automatisierung:** `createTss`/`createClient`, Admin-PIN/PUK-Verwaltung, Init-/Disable-/Defekt-Behandlung. Bleibt Betreibersache im fiskaly-Dashboard.
- **Verschlüsselung-at-rest** der TSE-Schlüssel: spätere Härtung; vorerst plaintext in der Datenbank (Betreiber-geschützte DB, konsistent mit der heutigen Secrets-Haltung).
- **`UpdateTransaction`-Nutzung:** Im Interface definiert, aber im atomaren Festzelt-Muster nicht verwendet.
- **Offline-Fähigkeit (Q-05):** Lokale Zwischenspeicherung bei Internetausfall ist ein separates Nice-to-have.
- **Weitere TSE-Anbieter:** Das Interface ist anbieter-agnostisch, aber es wird nur die fiskaly-Implementierung gebaut.
- **Hardware-TSE:** Für BYOD-Smartphone-Betrieb ausgeschlossen (Cloud-TSE).

## Further Notes

- **fiskaly TEST-Umgebung:** TSS-Instanzen im TEST-Env werden regelmäßig (mind. sonntags, nach 14 Tagen Inaktivität) gelöscht; maximal fünf aktive TSS. Tests müssen damit umgehen (Teardown). Für CI wird ohnehin gegen einen Mock-HTTP-Server getestet, nicht gegen fiskaly.
- **Rate Limits / Idempotenz:** fiskaly ist idempotent; `5xx`/`499`/`429` sind wiederholbar (Backoff). Ein TSS erlaubt max. 2000 offene Transaktionen — im sofort-geschlossenen Festzelt-Muster unkritisch.
- **Bekannter fiskaly-Hinweis:** `signature.counter` kann als Zahl **oder** String zurückkommen; die Implementierung muss beides akzeptieren.
- **„DON'T BLOCK THE TILL"** ist eine ausdrückliche fiskaly-Vorgabe und deckt sich mit dem AO-Anwendungserlass zu § 146a: Bei TSE-Störung darf der Kassiervorgang nicht blockieren, sondern wird mit Ausfallvermerk fortgeführt und nachsigniert.
- **Anschlussfähigkeit:** Das Datenmodell ist auf den DSFinV-K-Export (F-04) ausgelegt — `ABRECHNUNGSKREIS` pro Tisch-Session (F-06, bereits erfüllt), TSE-Felder je Transaktion, Durchbedienen-Zeitstempel.
- **Auswirkung auf F-02-Akzeptanzkriterien:** Das ursprüngliche Kriterium „konfigurierbarer Modus `strict`/`bypass` bei fehlender TSE" wird bewusst ersetzt durch „TSE optional + Admin-Hinweis/Warnung". `docs/anforderungen.md` F-02 ist entsprechend anzugleichen.

> **Annahme (Anfangsbestand):** `kassensitzung-eroeffnet:v1` (Wechselgeld-Anfangsbestand) wird nicht signiert, da das Mapping in `docs/handbuch.md` §3.13 nur Geldtransit-Einlagen/-Entnahmen als `SonstigerVorgang-V1` führt. Sollte eine spätere DSFinV-K-Prüfung den Anfangsbestand als kassenwirksame Bewegung verlangen, wird er analog zum Geldtransit nachgezogen.
