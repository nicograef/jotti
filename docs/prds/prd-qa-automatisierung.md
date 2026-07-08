# PRD: QA-Automatisierung für v1.0.0 und den laufenden Betrieb

> Quelle: QA-Analyse vom 08.07.2026 zum Release-Guide
> (`docs/plans/plan-v1.0.0-release.md`) und zum manuellen QA-Guide
> (`docs/plans/guide-manuelle-qa-v1.0.0.md`). Diese PRD baut die
> automatisierbaren und agentengetriebenen Teile der v1.0.0-QA als
> dauerhafte Artefakte, führt die QA einmal vollständig durch und
> reduziert den manuellen Guide auf die Punkte, die wirklich
> menschliche Handarbeit brauchen.

## Problem Statement

Vor dem v1.0.0-Tag muss die gesamte Anwendung geprüft werden: alle
Screens und Flows, Admin-Funktionen, Schnittstellen, Exporte,
Fehlerpfade, Druck und TSE, dazu Installations-, Update-, Backup- und
Restore-Pfad auf einem produktionsähnlichen System. Die heutige
Absicherung endet an der Unit- und Integrationstest-Grenze: Es gibt
keine End-to-End-Tests durch den Browser, keine automatisierte
DSFinV-K-Prüfung, keine systematische Berechtigungsprüfung über alle
Routen, keine Schwachstellen-Scans und keine wiederholbare Prüfung
der Ops-Pfade. Der manuelle QA-Guide bündelt deshalb rund vierzig
Prüfpunkte als Handarbeit, obwohl der größte Teil davon
automatisierbar oder von einem Agenten durchführbar ist.

Das hat zwei Folgen. Erstens ist die v1.0.0-QA als einmaliger
manueller Kraftakt geplant, der Tage kostet und nicht wiederholbar
ist. Zweitens fehlt nach dem Release die Schicht, die künftige
Updates absichert: Ein Fehler, der erst im Zusammenspiel von
Frontend, Backend und Datenbank sichtbar wird, fällt heute erst beim
Klicken auf, im schlimmsten Fall beim Verein im Betrieb.

## Solution

Die QA-Arbeit wird in drei Kategorien zerlegt und entsprechend
behandelt:

Kategorie A, voll automatisiert: Eine Playwright-E2E-Suite prüft bei
jedem Pull Request die Kernflows durch den echten Stack. Ein
DSFinV-K-Validator prüft jeden Export automatisch gegen die
Spezifikation. Eine Berechtigungs-Matrix testet jede Route gegen jede
Rolle. Schwachstellen-Scans für Go- und npm-Abhängigkeiten laufen in
CI. Kleine Fuzz-Targets sichern die Replay- und Encoder-Kanten. Ein
Parallelzugriffstest prüft die Datenkonsistenz bei gleichzeitiger
Bedienung.

Kategorie B, wiederholbar mit einem Befehl, lokal oder
agentengetrieben: Eine TSE-Live-Suite spielt alle Geschäftsvorfälle,
den Ausfall- und Nachsignierungspfad und eine Latenzmessung gegen die
fiskaly-TEST-Umgebung durch. Ein Ops-Smoke-Skript fährt auf einem
frischen Server den kompletten Roundtrip aus Installation, Backup,
Backup-Verifikation und Update und dient zugleich als Release-Smoke
für gepinnte Images.

Kategorie C, menschliche Handarbeit: Der bestehende manuelle QA-Guide
wird zum Rest-Guide umgebaut. Er enthält nur noch, was Automatisierung
nicht leisten kann: physische Hardware (Bondrucker, QR-Scan mit dem
Handy), der echte Windows-Rechner, fiskaly-Konto und PUK/PIN-
Verwahrung, destruktives Restore, TLS-Abnahme, Usability mit echten
Vereinshelfern und die Abnahme-Entscheidungen selbst.

Nach Fertigstellung der Artefakte wird die QA einmal vollständig
durchgeführt: Die Suiten laufen, ein Agent spielt zusätzlich einen
Screen-Sweep über alle Seiten und eine heuristische UX-Review durch,
und alle Befunde landen in einem Befund-Report, aus dem Fixes werden.
Was danach übrig bleibt, ist der Rest-Guide als konzentrierter
Handarbeits-Nachmittag statt eines mehrtägigen QA-Marathons.

## User Stories

1. Als Entwickler will ich eine E2E-Suite, die die Kernflows
   (Anmelden, Bestellen, Ausgeben, Kassieren, Direktverkauf,
   Stornieren, Umbuchen, Kassenabschluss) durch Browser, Backend und
   Datenbank prüft, damit Regressionen vor dem Merge auffallen und
   nicht beim Verein.
2. Als Entwickler will ich, dass die E2E-Suite bei jedem Pull Request
   in CI läuft, damit kein ungeprüfter Stand auf main kommt.
3. Als Entwickler will ich die E2E-Suite lokal mit einem Befehl gegen
   den Dev-Stack starten können, damit ich Fehler ohne CI-Umweg
   reproduzieren kann.
4. Als Entwickler will ich, dass jeder E2E-Lauf von einem definierten
   Seed-Zustand ausgeht, damit Tests deterministisch sind und nicht
   voneinander abhängen.
5. Als Entwickler will ich E2E-Tests für die Fehlerpfade (Server
   antwortet mit Fehler, Netz bricht ab), damit die Kasse sichtbar
   einen Fehler zeigt statt stiller Leer-Defaults wie Saldo 0,00.
6. Als Entwickler will ich E2E-Tests für die Admin-Flows (Produkte,
   Tische, Benutzer, Druckstationen, Kassenführung, Reporting),
   damit auch die selten bedienten Verwaltungsseiten abgesichert sind.
7. Als Entwickler will ich einen E2E-Test für den Export-Download aus
   der Admin-Oberfläche, damit der Weg vom Klick bis zur ZIP-Datei
   geprüft ist.
8. Als Entwickler will ich, dass die E2E-Suite die Servicekraft-Flows
   auch im Handy-Viewport prüft, weil die Zielnutzer im Service mit
   dem Telefon arbeiten.
9. Als Entwickler will ich einen DSFinV-K-Validator, der jedes
   Export-ZIP strukturell gegen die Spezifikation prüft (CSV-Regeln,
   Dateinamen, Spaltenreihenfolge, index.xml gegen die DTD), damit
   Formatfehler nicht erst ein Prüfer des Finanzamts findet.
10. Als Entwickler will ich, dass der Validator auch die Inhalte
    prüft (Storno-Referenzen, Kombi-Steueraufteilung, Bediener-
    Felder, Tagesabschluss-Zeile, TSE-Stammdaten), damit die
    fiskalische Außenkante bei jeder Änderung mitgetestet wird.
11. Als Entwickler will ich, dass ein Integrationstest aus Seed-Daten
    einen echten Export erzeugt und durch den Validator schickt,
    damit die Prüfung bei jedem Verify-Lauf automatisch stattfindet.
12. Als Betreiber-Vertreter will ich eine TSE-Live-Suite, die jeden
    Geschäftsvorfall (Bestellung, Teil- und Vollzahlung,
    Direktverkauf und dessen Storno, Warenrücknahme, geldneutrale
    Korrektur, Umbuchung, Geldtransit, Kassendifferenz,
    Tagesabschluss) real gegen die fiskaly-TEST-TSE signiert und die
    Signatur im Journal prüft, damit der Compliance-Anspruch auf
    einem durchgespielten Durchlauf ruht statt auf Doku.
13. Als Entwickler will ich den TSE-Ausfall- und Nachsignierungspfad
    automatisiert prüfen (Vorgänge bleiben buchbar, Störungsprotokoll
    erfasst den Zeitraum, Nachsignierung läuft nach Wiederherstellung,
    Abschluss-Gate antwortet korrekt), damit der kritischste
    Betriebsfall nicht nur einmal von Hand getestet wird.
14. Als Betreiber-Vertreter will ich eine reproduzierbare
    p95-Latenzmessung der Signaturen unter Burst, damit die Zusage
    der Verfahrensdokumentation (p95 unter 5 Sekunden) belegt oder
    korrigiert werden kann.
15. Als Entwickler will ich, dass der TSS-anlegende Setup-Durchlauf
    nur explizit gestartet wird, weil jeder Lauf eine nicht löschbare
    TSS im TEST-Konto hinterlässt.
16. Als Entwickler will ich eine Berechtigungs-Matrix, die jede
    HTTP-Route gegen jede Rolle und gegen fremde Objekt-IDs prüft,
    damit Rechte-Lücken (fehlende Autorisierung, IDOR) systematisch
    statt stichprobenhaft auffallen.
17. Als Entwickler will ich, dass neue Routen automatisch in die
    Berechtigungs-Matrix fallen und ein Test fehlschlägt, wenn eine
    Route nicht klassifiziert ist, damit die Prüfung nicht veraltet.
18. Als Entwickler will ich automatische Assertions auf Security-
    Header und Login-Rate-Limiting, damit diese Härtungen nicht still
    verloren gehen.
19. Als Entwickler will ich Schwachstellen-Scans der Go- und
    npm-Abhängigkeiten in CI, damit bekannte Lücken in Dependencies
    zeitnah auffallen.
20. Als Entwickler will ich kleine Fuzz-Targets für die
    Event-Replay-Kante und die Encoder (DSFinV-K, ESC/POS), deren
    Korpus bei jedem Testlauf mitläuft, weil Events zehn Jahre
    replaybar bleiben müssen.
21. Als Entwickler will ich einen Integrationstest für parallelen
    Zugriff (zwei Clients, gleicher Tisch), damit die
    Datenkonsistenz-Zusage geprüft ist, bevor zwei echte Handys sie
    im Vereinsheim testen.
22. Als Release-Verantwortlicher will ich ein Ops-Smoke-Skript, das
    auf einem frischen Server den Roundtrip aus prod-init, erstem
    Login, Backup, Backup-Verifikation und prod-update fährt und
    protokolliert, damit der Installations- und Update-Pfad
    wiederholbar geprüft ist statt einmalig.
23. Als Release-Verantwortlicher will ich dasselbe Skript mit
    gepinnten Release-Images als Release-Smoke nutzen (Installation,
    Login, ein Verkauf, ein Beleg, ein Export), damit Gate 6 des
    Release-Guides skriptgestützt abläuft.
24. Als Betreiber will ich, dass der manuelle QA-Guide nur noch
    echte Handarbeit enthält und für alles andere auf die
    automatisierte Abdeckung verweist, damit klar ist, was ich vor
    dem Release wirklich selbst tun muss.
25. Als Release-Verantwortlicher will ich nach Fertigstellung der
    Artefakte eine einmalige vollständige QA-Durchführung mit
    Befund-Report (Suiten-Läufe, Screen-Sweep über alle Seiten,
    heuristische UX-Review, TSE-Blöcke, Ops-Smoke), damit die
    v1.0.0-QA am Ende erledigt ist und nicht nur ermöglicht.
26. Als Entwickler will ich, dass alle neuen Suiten dauerhaft im
    Verify- und CI-Umfang bleiben, damit jedes Update nach v1.0.0
    denselben Sicherheitsstandard durchläuft wie das Release selbst.

## Implementation Decisions

- Die E2E-Suite lebt in einem eigenen Top-Level-Verzeichnis (sie
  testet den ganzen Stack, nicht das Frontend allein), nutzt
  Playwright mit TypeScript und pnpm und läuft als eigener CI-Job bei
  jedem Pull Request gegen den Docker-Compose-Stack mit Seed-Daten.
- Die E2E-Suite läuft ohne konfigurierte TSE. Das ist ein legaler,
  im Produkt vorgesehener Betriebsmodus (Vorgänge tragen den
  Hinweis „keine TSE konfiguriert"). Alles TSE-Spezifische deckt die
  TSE-Live-Suite ab; so bleibt CI frei von externen API-Abhängigkeiten
  und deren Flakiness.
- Browser-Matrix bewusst klein: Chromium in zwei Viewports (Desktop
  für Admin, Handy-Format für Service). Keine Multi-Browser-Matrix.
- Selektoren folgen der bestehenden Testing-Library-Philosophie:
  zugängliche Rollen und sichtbare Beschriftungen zuerst, Test-IDs
  nur wo nötig. Spec-Namen folgen der Fachsprache aus language.md.
- Jede Spec-Datei startet von einem definierten Seed-Zustand über den
  bestehenden Reset-und-Seed-Pfad; Specs sind untereinander
  unabhängig, innerhalb einer Spec-Datei dürfen Schritte aufeinander
  aufbauen (ein Kassenabschluss-Test braucht vorherige Umsätze).
- Der DSFinV-K-Validator ist ein eigenes Go-Paket mit einer schmalen
  Schnittstelle (Export-ZIP rein, Befundliste raus) und eigenen
  Unit-Tests. Die Prüfregeln zitieren die jeweilige Stelle der
  DSFinV-K-2.4-Spezifikation aus den lokalen Rechtsquellen; die
  beiliegende DTD wird für die index.xml-Prüfung verwendet. Ein
  Integrationstest erzeugt aus Seed-Daten (inklusive Fake-TSE für
  gefüllte TSE-Felder) einen echten Export und muss befundfrei sein.
- Die TSE-Live-Suite baut auf dem bestehenden Live-Test-Muster auf
  (Umgebungsvariablen, Skip wenn nicht gesetzt, Guard gegen
  Nicht-TEST-Umgebungen) und wird über ein eigenes make-Target
  gestartet, das die Credentials aus der lokalen, gitignorierten
  Credentials-Datei liest. Entschieden: kein CI-Workflow für
  TSE-Tests; Ausführung nur lokal beziehungsweise agentengetrieben.
  Ein Secrets-basierter GitHub-Workflow kann später ergänzt werden,
  ohne an der Suite etwas zu ändern.
- Der TSS-anlegende Setup-Durchlauf bleibt ein separates, explizit
  aufzurufendes Target und ist nicht Teil des Standard-Laufs.
- Die Latenzmessung ist Teil der TSE-Live-Suite (kein eigenes
  Lasttest-Werkzeug): eine definierte Zahl Signaturaufträge im Burst,
  Ausgabe von p50/p95. Das Ergebnis fließt in die
  Verfahrensdokumentation ein (Zusage bestätigen oder anpassen).
- Die Berechtigungs-Matrix ist ein table-driven Integrationstest, der
  aus der Routen-Registrierung gespeist wird: Für jede Route ist die
  erlaubte Rollenmenge deklariert; eine nicht deklarierte Route lässt
  den Test fehlschlagen. Geprüft werden erlaubte Rollen (2xx/fachlich),
  verbotene Rollen (403), fehlende und ungültige Tokens (401) und der
  Zugriff auf fremde Objekt-IDs.
- Security-Header und Login-Rate-Limiting werden auf zwei Ebenen
  geprüft: als Integrationstest auf Middleware-Ebene (schnell, pro
  Verify-Lauf) und durch das Ops-Smoke-Skript am real deployten
  Stack (mit Reverse-Proxy davor).
- Schwachstellen-Scans als eigene CI-Jobs: govulncheck für alle
  Go-Module, pnpm audit für das Frontend. Fehlschlag bei erreichbaren
  beziehungsweise hoch eingestuften Befunden; die bestehende
  Supply-Chain-Policy (Mindestalter neuer Pakete) bleibt unberührt.
- Fuzzing bewusst klein (Ermessens-Entscheidung): Go-native
  Fuzz-Targets nur für die Event-JSON-Deserialisierung (Replay-Kante)
  und die beiden Encoder (DSFinV-K-CSV, ESC/POS). Der Seed-Korpus
  läuft automatisch als Unit-Test mit; längere Fuzz-Läufe nur lokal
  über ein make-Target, kein Dauerlauf in CI. Kein k6, keine
  Lastszenarien über die Latenzmessung hinaus.
- Das Ops-Smoke-Skript ist ein Shell-Skript im bestehenden
  Skript-Verzeichnis (shellcheck-pflichtig wie alle anderen), das auf
  einem frischen Ubuntu-Host läuft. Modi: Erstinstallation
  (prod-init bis erster Login-Roundtrip per API), Betrieb (prod-backup,
  prod-backup-verify, prod-update-Roundtrip) und Release-Smoke
  (gepinnte Image-Version als Parameter, dann Installation, Login,
  ein Verkauf, ein Beleg, ein Export per API). Es protokolliert jeden
  Schritt maschinenlesbar. Die Host-Provisionierung bleibt beim
  Menschen; destruktives prod-restore und die TLS-Abnahme bleiben im
  Rest-Guide.
- Der bestehende manuelle QA-Guide wird in place zum Rest-Guide
  umgebaut: Automatisierte Punkte werden entfernt und durch einen
  Verweis auf die jeweilige Suite ersetzt; übrig bleiben Hardware
  (Bondrucker-Druckbild, QR-Scan mit dem Handy), der echte
  Windows-Rechner, fiskaly-Konto samt TEST-zu-LIVE-Umschaltung und
  PUK/PIN-Verwahrung, Zwei-Geräte-Test in echt, destruktives Restore,
  TLS, Usability mit Vereinshelfern und die Abnahmen. Eine einzige
  Checkliste, keine Doppelpflege.
- Die einmalige QA-Durchführung ist Teil des Vorhabens: Nach
  Fertigstellung laufen alle Suiten; zusätzlich führt ein Agent einen
  Screen-Sweep über sämtliche Seiten und Zustände (auch die, die
  keine eigene Spec haben) und eine heuristische UX-Review durch.
  Alle Befunde landen in einem Befund-Report als Arbeitsdokument;
  Fixes sind Folgearbeit außerhalb dieser PRD, der Report ordnet sie
  nach Schwere.
- Belegdruck-Prüfung bleibt zweigeteilt: Der ESC/POS-Bytestrom mit
  den Pflichtangaben wird durch die bestehenden Formatter-Tests
  abgedeckt; die Reihenfolge Signatur vor Druckbefehl prüft die
  TSE-Live-Suite am Ablauf. Das physische Druckbild und die
  QR-Scanbarkeit bleiben im Rest-Guide.

## Testing Decisions

- Ein guter Test prüft von außen sichtbares Verhalten, nie
  Implementierungsdetails: Die E2E-Suite assertet, was eine
  Servicekraft oder ein Admin sieht (Texte, Beträge, Zustände), nicht
  interne Requests oder Store-Zustände. Der Validator assertet Fakten
  der Spezifikation, nicht die Struktur des Export-Codes; jede Regel
  trägt die Fundstelle als Kommentar. Die Berechtigungs-Matrix
  assertet HTTP-Statuscodes, nicht Middleware-Interna.
- E2E-Stabilitätsregeln: keine festen Wartezeiten, nur Playwrights
  wartende Assertions; Determinismus über Seed-Zustand statt über
  Testreihenfolge; ein fehlgeschlagener Test hinterlässt Trace und
  Screenshot als CI-Artefakt.
- Vorbilder im Bestand: das Integrationstest-Muster (Build-Tag,
  serialisierte Pakete gegen eine echte Datenbank), das
  Live-Test-Muster gegen fiskaly-TEST (Skip ohne Credentials, Guard
  gegen LIVE), der Event-JSON-Contract-Test als Guard-Muster, die
  Seed-Engine mit Fake-TSE, die ESC/POS-Formatter-Tests und die
  Testing-Library-Tests im Frontend.
- Getestete Module: Der DSFinV-K-Validator bekommt als einziges
  Nicht-Test-Modul eigene Unit-Tests (gute und absichtlich kaputte
  Export-Fixtures). Die Skripte werden durch shellcheck (bestehender
  CI-Job) und durch reale Ausführung auf einem Wegwerf-Host
  verifiziert. Alle übrigen Module sind selbst Tests; ihre Qualität
  wird über die einmalige Durchführung und Review gesichert.
- Neue Suiten werden in die bestehende Verify-Kette und CI
  eingehängt, damit „grün" weiterhin genau eine Bedeutung hat.

## Out of Scope

- Externe Penetrationstests und DAST-Scanner (etwa ZAP): Die
  Berechtigungs-Matrix, Header-Checks und Dependency-Scans decken die
  realistische Fehlerklasse ab; ein Scanner brächte bei einer
  JSON-API mit schmaler Oberfläche wenig zusätzlich und kostet
  Stabilität.
- Lasttests über die Signatur-Latenzmessung hinaus (kein k6, keine
  Lastszenarien): Eine Vereinskasse hat eine Handvoll gleichzeitiger
  Nutzer; der Parallelzugriffstest deckt die Konsistenzfrage ab.
- Usability-Tests mit echten Nutzern, physische Hardware-Prüfungen,
  Windows-Smoke auf echtem Rechner: bleiben bewusst im Rest-Guide.
- Automatisierte Provisionierung von Cloud-VMs (hcloud) für den
  Ops-Smoke: Der Host wird manuell gestellt.
- TSE-Tests in GitHub Actions: entschieden gegen einen CI-Workflow;
  die Suite ist so gebaut, dass er später ergänzt werden kann.
- Visuelle Regressionstests (Screenshot-Diffs) und
  Multi-Browser-Matrix.
- Gegenlesen des Exports mit IDEA oder fiskaly-Prüftooling: bleibt
  als Abnahme im Rest-Guide.
- Nacharbeit Block 6 (CHANGELOG, Versionsstellen, TODO-Grep) und der
  Release-Schnitt selbst: eigener Plan, eigene Session.
- Behebung der Befunde aus der QA-Durchführung: Der Report ordnet
  und priorisiert sie; die Fixes sind Folgearbeit.

## Further Notes

- Die fiskaly-TEST-Credentials liegen vollständig in der lokalen,
  gitignorierten Datei nach dem Muster der committeten Vorlage
  (.env.fiskaly-test.example); die bestehenden make-Targets
  test-tse-live und test-tse-live-setup lesen sie. Die TSE-Live-Suite
  baut darauf auf.
- Beim Neuaufsetzen der TEST-TSS am 08.07.2026 wurde ein
  Feld-Mapping-Fehler gefunden und behoben: Die TSS-Seriennummer wurde
  nie gelesen, weil die TSS-Ressource das Feld serial_number nennt
  (tss_serial_number gibt es nur auf Transaktions-Responses). Der Fall
  ist das Musterbeispiel für die TSE-Live-Suite: Sie muss die
  Vollständigkeit der persistierten Stammdaten explizit prüfen, nicht
  nur den Signaturerfolg.
- Das fiskaly-TEST-Konto füllt sich mit jeder per Setup-Durchlauf
  angelegten TSS (nicht löschbar). Deshalb nutzt der Standard-Lauf
  die bestehende TSS und der Setup-Durchlauf bleibt opt-in.
- Bezug zum Release-Guide: Diese PRD automatisiert Gate 1 (Rest),
  große Teile von Gate 2 (Blöcke 3, 4 und 6 des QA-Guides), Gate 3
  (Ops-Roundtrips, soweit nicht destruktiv) und den Smoke-Teil von
  Gate 6. Der umgebaute Rest-Guide trägt die verbleibenden
  Abnahmepunkte; die Checkboxen der Gates bleiben im Release-Guide
  führend.
- Der bestehende Upgrade-Pfad-CI-Job (Migration auf befüllter
  Vorversions-Datenbank, Boot, Projektions-Rebuild) bleibt das
  Rückgrat der Update-Sicherheit und wird von dieser PRD nicht
  angefasst, nur ergänzt.
- Empfohlene Umsetzungsreihenfolge nach Hebelwirkung: E2E-Suite,
  DSFinV-K-Validator, Berechtigungs-Matrix und Scans, TSE-Live-Suite,
  Ops-Smoke, Guide-Umbau, dann die Durchführung mit Befund-Report.
