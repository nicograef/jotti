# Findings: Vollreview jotti (2026-09-08)

> Quelle: Multi-Experten-Review aller Dateien des Repos (Stand `claude/plan-orchestrierung-phase-a-2qcafv` @ 5edad60),
> 22 Einheiten × 3 Linsen (Cleanup-Skill, Korrektheit/Security, Konventionen/Doku) plus
> 8 Cross-Layer-Flüsse; je Bereich ein Fable-Sweep als Übergabe, Reviewer und Prüfer: Opus 5.
> Jeder Blocker/Major-Befund wurde von 2–3 unabhängigen Skeptikern gegengeprüft; Minor-Befunde sind ungeprüft.
> Ausgeschlossen: `backend/sqlc/dbgen/`, Lockfiles, Binärdateien, `docs/rechtsquellen/`.
> Dieses Dokument ist die Eingabe für `plan-jotti-audit-fixes.md` (create-plan). Es enthält keine Personendaten.

## Zahlen

| Kennzahl                          | Wert                                  |
| --------------------------------- | ------------------------------------- |
| Gelesene Dateien (laut Reviewern) | 2609                                  |
| Rohbefunde                        | 748                                   |
| Nach Dedupe                       | 747                                   |
| Verifiziert                       | 177                                   |
| Verworfen                         | 49                                    |
| Verbleibend                       | 698 (Blocker 1, Major 127, Minor 570) |
| Reviewer ohne Ergebnis            | 0                                     |

## Top 10 repo-weit

- **Blocker — TSE-Assistent ist nach Anleitung nicht erreichbar**: der gesetzlich zwingende Einrichtungsschritt hat keinen auffindbaren Einstieg (`docs/leitfaden/tse-einrichten.md:26-32`, UI-Lücke in `frontend/src/admin/finanzamt/EinrichtungSection.tsx:181-198`).
- **Alle Beleg- und Bon-Zeitstempel in UTC statt Europe/Berlin**: im Sommer zwei Stunden falsch, nahe Mitternacht steht der falsche Belegtag auf dem Bon (`backend/api/druck/bondruck/application/escpos/formatter.go:113-305`).
- **Variante wird nie gegen das Produkt geprüft**: beliebiger Steuersatz zu beliebigem Preis buchbar, der falsche Satz wird eingefroren und TSE-signiert (`backend/api/kasse/enrichment/enrichment.go:72-99`, `backend/repository/produkt_repo/batch.go:16-63`).
- **Menge ohne Obergrenze**: `EinzelpreisCents * Menge` läuft über und schreibt einen falschen Gesamtpreis unveränderlich ins Kassenjournal (`backend/domain/kasse/bestellung.go:97-106`).
- **Letzter Admin kann sich selbst deaktivieren oder herabstufen** und sperrt die Instanz ohne DB-Zugriff dauerhaft aus (`backend/api/stammdaten/user/http/command_handler.go:166-219`, `frontend/src/admin/users/UserRow.tsx:69-84`).
- **Idempotenz-Schlüssel überleben geänderte Payloads**: Bestellungen und Geldtransits werden idempotent verworfen und trotzdem als Erfolg gemeldet (`frontend/src/service/components/table/BestellungAbschluss.tsx:50-81`, `frontend/src/admin/kasse/GeldtransitDialog.tsx:50-70`).
- **Kassenabschluss-Retry-Guard sieht nur den Kassensitzungs-Stream**: Tisch-Zahlungen erscheinen als Soll-Ist-Differenz auf dem signierten Z-Bon (`backend/api/kasse/kassenfuehrung/application/command.go:364-376`).
- **Passwort wird beim Setzen getrimmt, beim Login nicht**: ein Konto mit Rand-Leerzeichen bleibt dauerhaft unbenutzbar (`backend/domain/user/user.go:202-228`).
- **Backup-Dumps mit TSE-API-Secret und allen Passwort-Hashes liegen weltlesbar** — zehn Jahre lang (`scripts/prod-backup.sh:66-94`).
- **Der Restore startet die gescheiterte Migration erneut**, der Stack bleibt unten und der dokumentierte Recovery-Pfad führt ins Leere (`packaging/windows/jotti-restore.cmd:31-37`, `packaging/windows/KURZANLEITUNG.md:118-125`).

## Defektklassen und vorgeschlagene Gates

- **Regel 18 — historisierende Prosa** in Kommentaren, Docs und Konsolen-Strings („früher/bisher/nicht mehr", datierte Änderungseinträge, „Phase N", Handoff- und Review-Verweise, Zeiger nach `docs/plans/`). Gate: CI-Grep über `**/*.{go,ts,tsx,astro,css,md,cmd,sh}` auf diese Wortliste, Ausnahmen nur `CHANGELOG.md` und `docs/adrs/`.
- **Tote Datei- und Dokumentverweise** (gelöschte PRDs und Pläne, `docs/leitfaden.md`, „Design-Handoff", NEU02/NEU07, F2/F8). Gate: `make links` auf zitierte Repo-Pfade in Quellkommentaren ausweiten und als CI-Job führen.
- **Zeitstempel ohne definierte Zone** an Präsentations-, Dateinamens- und `CURRENT_DATE`-Grenzen. Gate: Grep — `.Format(` im Druck-, Beleg- und Exportpfad nur nach `.In(berlin)`; Test mit fixem UTC-Mitternachtszeitpunkt.
- **Schemas mit unterer, aber ohne obere Schranke** (zog: Menge, Betreiber-Stammdaten) und Zod ohne `.trim()/.min()` gegenüber zog. Gate: Test über jedes persistierte Feld auf GTE+LTE plus Vergleichstest Zod↔zog je DTO.
- **Anwendungs-Sentinels ohne HTTP-Mapping** werden zu 500 statt 400/409, während das Frontend einen nie gesendeten Code mappt. Gate: Test, der jedes exportierte `Err*` gegen die `helper.MapError`-Karten und die Frontend-`byCode`-Tabellen prüft.
- **Repository-Fehler nicht normalisiert oder nicht durchgereicht** (`reporting_repo` ohne `db.Error()`, fehlende `errors.Is`-Zweige an einzelnen Call-Sites). Gate: Lint-Regel je `repository`-Paket plus aktiviertes `errorlint`.
- **ID-Paare werden getrennt geladen, die Zugehörigkeit nie geprüft** (Produkt/Variante beim Buchen und beim Löschen). Gate: Lookup-Signaturen mit Elternschlüssel; Integrationstest „fremde Variante wird abgelehnt".
- **Zwischenstatus `wird_abgeschlossen` ignoriert** — `GetOffeneKassensitzung` statt `GetAktiveKassensitzung` in Lesepfaden und Guards. Gate: Grep-Gate auf `GetOffeneKassensitzung` außerhalb des Abschlusspfads plus Absturztest.
- **Event-JSON wird gegen ungetaggte Domain-Structs dekodiert** statt über die `*EventData`-Vertragstypen (Regel 10). Gate: Lint — `json.Unmarshal`/`json.Marshal` nur in Event-Data-Typen; Contract-Test auf jeden Lesepfad ausweiten.
- **Idempotenz-Schlüssel rotieren an UI-Zuständen statt am Payload**. Gate: Schlüssel aus einem Payload-Hash ableiten; Vitest, der nach geändertem Warenkorb einen neuen Schlüssel erzwingt.
- **React-Query-Hooks verwerfen `isError` und liefern Defaults**, sodass Fehlladungen wie echte Nulldaten aussehen. Gate: ESLint-Regel „jeder `useQuery`-Konsument reicht `isError` durch" plus Test „Fehlladung zeigt keine Nulldaten".
- **Ein Backend-Vertrag in mehreren Frontend-Fassungen** (Produkt-, Steuersatz-, Status-, Kategorie-Schemas doppelt bis dreifach, Response-Schemas mit Formular-Validierung). Gate: ein Schema-Modul je Domänenbegriff, `no-restricted-imports` gegen Zweitkopien.
- **Geteilte Schichten importieren aufwärts aus Feature-Bereichen** (`components/common` → `admin`, `hooks` → `tseAmpel`). Gate: `import/no-restricted-paths` — `components/`, `lib/`, `hooks/` dürfen nicht aus `admin/` oder `service/` importieren.
- **Gates, die nicht rot werden können**: `eslint --fix` im lint-Skript, `goimports -l` ohne Exit-Guard, tautologische CI-`if`, `make lint` schwächer als CI, `src/components/ui` von Lint und Prettier ausgenommen. Gate: Exit-Code-Guard-Pflicht, `make lint` identisch zum Merge-Gate, Negativtest mit bewusst kaputter Datei.
- **Unit-Tests werden nie gelintet** — fehlende oder inkonsistente `//go:build`-Tags, golangci-lint nur mit `--build-tags=integration`. Gate: zweiter golangci-lint-Lauf mit `unit`-Tag; Grep, dass jede `_test.go` genau ein Build-Tag trägt.
- **Tests prüfen Implementierungsdetails oder setzen weiche Fallbacks** (Tailwind-Klassennamen, wortgleiche Produktionsdaten, `?? 0`, `networkidle`, 50-Klick-Bound ohne Nachbedingung). Gate: ESLint-Verbot für `networkidle` und `?? 0` in `e2e/`, Review-Regel gegen Klassennamen-Assertions.
- **DRY-Verstöße in Test-, E2E- und Ops-Pipelines** (Fixtures und DB-Cleanup in sieben Paketen, `data-slot`-Selektoren, axe-Block, Makefile-`check-*`, kopierte CI-Jobs und Compose-Stacks). Gate: gemeinsamer Test-Helfer neben `OpenTestDatabase`, Selektoren nur aus `support/`, CI-Matrix statt Kopien.
- **Duplizierte Wahrheit ohne Kopplung** (TERMS-Datum, „Beta 1.0", Sidebar vs. `publishedDocs`, `MaxDruckversuche`, Retry-Schwelle, `01_initial.up.sql` als „vollständiges Schema" neben den Migrationen 02–08). Gate: je ein Test, der die Kopien vergleicht.
- **Werkzeug- und Versionsdrift**: `typescript`, `packageManager`-Hash und `minimumReleaseAgeExclude` zwischen `frontend`/`website`/`e2e`, ungepinnte Images neben exakten Pins (`postgres:17` vs. `17.8`, pnpm 11 vs. 11.6.0). Gate: Pin-Vergleichsskript in CI plus Grep auf `image:`-Zeilen ohne Patch-Version.
- **Doku-Drift gegen Code**: zitierte Menüpunkte und Bedienelemente, Konfigurierbarkeits-Zusagen über Konstanten, Planungs-Tempus („Zielbild", „Entscheidung bei Implementierung"), Beispiel-Configs mit konkreten Release-Versionen. Gate: CI-Check, der in `docs/leitfaden/**` zitierte Bedienelemente gegen `frontend/src` grept.
- **Kommentare behaupten Verhalten, das der Code nicht hat** (POST-Filter, `data-theme`, „ausschließlich zugängliche Selektoren", „content-hashed", 0600 unter Windows). Gate: Review-Checkpunkt bei jeder Kommentaränderung plus ein Test je behauptetem Verhalten.
- **Sprach- und Zeichenkonventionen**: Umlaut und `ae/oe/ue` in derselben Datei, englische Bezeichner und a11y-Strings in deutscher Domäne, Nicht-ASCII in Windows-Konsolen-Strings, LF-Zeilenenden in `.cmd`. Gate: Grep-Gate pro Paket, ASCII-Check über `windows/**`, `.gitattributes` mit `*.cmd text eol=crlf`.
- **Secret-tragende Dateien mit Standardrechten** (Caddyfile 0644, Dumps 0644, `.env` unter `%PROGRAMDATA%`). Gate: Konvention 0600 bzw. `umask 077`; CI-Test der erzeugten Dateimodi in den Ops-Skripten.
- **Domänenwissen in Repository und HTTP dupliziert statt importiert** (Betreiber-Schema, Druckstation-Kategorien, Row-Mapper, Enum-Literale). Gate: Grep auf Enum-Literale außerhalb `domain/`; Schemas nur aus `domain/` beziehen.
- **`log.Error()` ohne `.Err(err)`**. Gate: zerolog-Lint, der Error-Level ohne Fehlerobjekt ablehnt.
- **Self-referential Test-Fixtures**, die den geprüften Produktionstyp selbst marshallen. Gate: Regel — Fixtures über Domänen-Konstruktoren oder goldene JSON-Dateien bauen.
- **Opake Audit-IDs als Testbegründung** (NEU13, NEU14, „Muster 05", „Befund #1"). Gate: Grep-Verbot dieser Muster in `e2e/tests`.
- **`byCode`-Overrides formulieren zentrale Fehlermeldungen um** statt Kontext zu ergänzen. Gate: Test, der jede `byCode`-Meldung gegen die zentrale Tabelle prüft.

## Backend

Go-Backend (api/, domain/, repository/, app/, config/, seed/, sqlc-Queries, Dockerfile) inklusive Fiskal-, Druck- und Kasse-Kontexte; generierter Code unter `sqlc/dbgen` ausgenommen.

0 Blocker · 28 Major · 233 Minor — 261 Befunde nach Dedupe, 942 Dateien geprüft, 42 verifiziert, 14 verworfen.

### backend/domain/kasse/bestellung.go

**Menge ohne Obergrenze: int-Overflow schreibt Müll ins Kassenjournal** (major · security, AGENTS.md Regel 5)
Datei: backend/domain/kasse/bestellung.go:97-106
Warum: `EinzelpreisCents * Menge` läuft über und wickelt auf einen plausiblen Kleinbetrag, der jede weitere Prüfung passiert.
Vorschlag: `"Menge": z.Int().GTE(1, …).LTE(999, z.Message("Menge zu hoch")).Required()` ergänzen und in beiden HTTP-Schemas spiegeln.
Aufwand: S · Status: bestätigt

### backend/api/kasse/enrichment/enrichment.go

**Variante wird nicht gegen das Produkt geprüft** (major · security, architecture.md → Anti-Corruption Layer)
Datei: backend/api/kasse/enrichment/enrichment.go:76-99
Warum: Steuersatz und Kategorie stammen aus dem Produkt, der Preis aus einer beliebigen fremden Variante.
Vorschlag: `produkt_id` in `GetVariantenByIDs` mitselektieren und bei Abweichung mit eigenem Sentinel neben `ErrVarianteNichtAktiv` ablehnen.
Aufwand: M · Status: bestätigt

**Datierte Review-Attribution im Paket-Doc und in drei errors.go** (minor · convention, AGENTS.md Regel 18)
Datei: backend/api/kasse/enrichment/enrichment.go:1-6
Warum: Kommentare dokumentieren ein Refactoring-Ereignis statt des aktuellen Stands.
Vorschlag: „(2026-07-17 …)" hier sowie in tischgeschaeft/, direktverkauf/ und kassenfuehrung/application/errors.go streichen, Sachaussage behalten.
Aufwand: S · Status: unverifiziert (minor)

**Geteiltes Enrichment-Paket ohne jeden Test** (minor · test-quality, AGENTS.md Self-Review 6)
Datei: backend/api/kasse/enrichment/enrichment.go:43-70
Warum: Die ID-Deduplikation vor den Batch-Lookups ist ein stiller Vertrag ohne Regressionsschutz.
Vorschlag: `enrichment_test.go` (`//go:build unit`) mit Fake-Repo: Duplikate, Reihenfolge, ErrProduktNotFound, ErrVarianteNichtAktiv.
Aufwand: S · Status: unverifiziert (minor)

### backend/repository/produkt_repo/batch.go

**GetVariantenByIDs liefert kein produkt_id und filtert nicht danach** (major · security, architecture.md → Repository Pattern)
Datei: backend/repository/produkt_repo/batch.go:16-63
Warum: Keine Schicht kann das Paar (produktId, varianteId) prüfen, weil die Zugehörigkeit gar nicht geladen wird.
Vorschlag: `produkt_id` in die Query und auf den Rückgabetyp aufnehmen, damit `EnrichPositionen` die Paarung ablehnen kann.
Aufwand: M · Status: bestätigt

**Zwei handgeschriebene SQL-Literale außerhalb von sqlc** (minor · convention, AGENTS.md Regel 14/15)
Datei: backend/repository/produkt_repo/batch.go:23-27
Warum: Diese Statements entgehen `make sqlc` und brechen erst zur Laufzeit bei Spaltenumbenennungen.
Vorschlag: Beide Queries nach `sqlc/queries/produkte.sql` verschieben, `make sqlc` laufen lassen, `toInt32Slice` entfällt.
Aufwand: M · Status: unverifiziert (minor)

**nolint-Begründung behauptet nicht-nutzergesteuerte IDs** (minor · docs-accuracy, readability.md → Prose Slop)
Datei: backend/repository/produkt_repo/batch.go:117-125
Warum: Beide ID-Listen kommen aus dem Request-Body und haben nur eine untere Schranke.
Vorschlag: Begründung durch die echte Invariante ersetzen: der Ergebnis-Map-Key ist die DB-ID, eine gekappte ID verfehlt den Lookup.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/stammdaten/user/http/command_handler.go

**Selbst-Aussperr-Schutz deckt nur /delete-user ab** (major · security, docs/anforderungen.md A-01)
Datei: backend/api/stammdaten/user/http/command_handler.go:166-219
Warum: Der letzte Admin kann sich selbst deaktivieren oder herabstufen und sperrt die Instanz dauerhaft aus.
Vorschlag: Die `body.ID == currentUserID`-Prüfung aus DeleteUserHandler auf DeactivateUserHandler und den Rollenwechsel ausdehnen.
Aufwand: M · Status: bestätigt

**ErrInvalidUserData ist in keiner Fehler-Map** (minor · boundary-consistency, principles.md → Least Surprise)
Datei: backend/api/stammdaten/user/http/command_handler.go:44-95
Warum: Ein Eingabefehler fällt durch `MapError` auf 500, während tisch und produkt 400 liefern.
Vorschlag: `application.ErrInvalidUserData: "invalid_user_data"` in beide Maps aufnehmen oder den Fehlerwert löschen.
Aufwand: S · Status: unverifiziert (minor)

**Value-Receiver statt Pointer-Receiver** (minor · convention, readability.md → Naming)
Datei: backend/api/stammdaten/user/http/command_handler.go:44-193
Warum: Alle vier Schwesterpakete verwenden `*CommandHandler`, nur user weicht ohne Grund ab.
Vorschlag: Die sechs Methoden auf `func (h *CommandHandler)` umstellen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/kasse/kassenfuehrung/application/command.go

**Retry-Guard des Kassenabschlusses sieht Tisch-Buchungen nicht** (major · correctness, docs/handbuch.md → Tagesabschluss)
Datei: backend/api/kasse/kassenfuehrung/application/command.go:364-376
Warum: Nach einem Teilabbruch gebuchte Zahlungen erscheinen als Soll-Ist-Differenz auf dem signierten Z-Bon.
Vorschlag: Den wiederverwendeten Sturz gegen den neu berechneten Soll-Bestand prüfen und bei Abweichung mit ErrBuchungenNachKassensturz abbrechen.
Aufwand: M · Status: bestätigt

**Signatur-Gate läuft vor dem Setzen der Abschluss-Barriere** (minor · correctness, docs/compliance.md §3.8)
Datei: backend/api/kasse/kassenfuehrung/application/command.go:296-316
Warum: Zwischen Gate und Barriere kann eine Buchung committen und einen ungeprüften Signaturauftrag erzeugen.
Vorschlag: Barriere zuerst setzen, Gate danach auf dem eingefrorenen Datenbestand ausführen; der vorhandene defer setzt bereits zurück.
Aufwand: S · Status: unverifiziert (minor)

**Kommentar verweist auf „Resolved decisions" in docs/plans/** (minor · docs-accuracy, AGENTS.md Git-Workflow)
Datei: backend/api/kasse/kassenfuehrung/application/command.go:418-420
Warum: Die tragende Begründung zeigt auf ein Dokument, das nach dem Merge gelöscht wird.
Vorschlag: Die Tatsache selbst nennen und auf `domain/kasse/tagesabschluss_summen.go` verweisen.
Aufwand: S · Status: unverifiziert (minor)

**Umlaute und ae/oe/ue-Transliteration in derselben Datei** (minor · readability, AGENTS.md Bewertungsmetriken → Konsistenz)
Datei: backend/api/kasse/kassenfuehrung/application/command.go:42-58
Warum: Dieselben Wörter stehen wenige Zeilen auseinander in zwei Schreibweisen.
Vorschlag: Auf echte Umlaute vereinheitlichen (Mehrheitsschreibweise der Unit), Kommentare und zerolog-Messages gleichermaßen.
Aufwand: S · Status: unverifiziert (minor)

### backend/domain/user/user.go

**SetPassword trimmt, VerifyPassword nicht** (major · correctness, docs/handbuch.md §5.2)
Datei: backend/domain/user/user.go:202-228
Warum: Ein Passwort mit Rand-Leerzeichen wird getrimmt gespeichert und meldet den Benutzer nie wieder an.
Vorschlag: In VerifyPassword dieselbe Normalisierung anwenden oder `.Trim()` aus PasswordSchema entfernen; Frontend-Grenze mitziehen.
Aufwand: S · Status: bestätigt

**time.Now() zweimal im Konstruktor** (minor · code-smell, code-smells.md → Unnecessary Complexity)
Datei: backend/domain/user/user.go:121-122
Warum: CreatedAt und UpdatedAt eines frischen Datensatzes unterscheiden sich, die „nie geändert"-Prüfung trägt nicht mehr.
Vorschlag: `now := time.Now().UTC()` einmal binden; ebenso in tisch.go:78-79, product.go:108-109, variant.go:79-80.
Aufwand: S · Status: unverifiziert (minor)

### backend/domain/betreiber/betreiber.go

**Betreiber-Stammdaten ohne Längengrenzen der DSFinV-K** (major · correctness, DSFinV-K 2.4 index.xml)
Datei: backend/domain/betreiber/betreiber.go:24-32
Warum: Ein Vereinsname über 60 Zeichen erzeugt einen Export, der seine eigene index.xml verletzt.
Vorschlag: Max-Längen (Name/Straße 60, PLZ 10, Ort 62, StNr 20, UstID 15) plus `.Trim()` ergänzen und im Admin-Formular spiegeln.
Aufwand: S · Status: bestätigt

### backend/domain/kasse/tisch_session_events.go

**Kommentar-Schema prüft Bytes, der Erzeuger kürzt auf Runen** (major · boundary-consistency, docs/language.md)
Datei: backend/domain/kasse/tisch_session_events.go:118-128
Warum: Tische mit Umlauten im Namen machen Umbuchungen dauerhaft unmöglich, der Fehler endet als 500.
Vorschlag: `buildUmbuchungKommentar` byteweise kürzen; das eingefrorene Schema bleibt unverändert.
Aufwand: S · Status: bestätigt

**Validate-Flatten-Wrap 17-mal kopiert** (minor · principle, principles.md → DRY)
Datei: backend/domain/kasse/tisch_session_events.go:151-154
Warum: Jeder neue Event-Typ tippt vier Zeilen Fehler-Plumbing nach, Abweichungen fallen im Review nicht auf.
Vorschlag: Ein paket-privater Helfer `validateEventData(schema, data, name) error`, aufgerufen aus jedem Konstruktor.
Aufwand: M · Status: unverifiziert (minor)

**Identischer „Muss positiv"-Kommentar an 11 Stellen** (minor · readability, readability.md → Unnecessary Comments)
Datei: backend/domain/kasse/tisch_session_events.go:34-35
Warum: Elf Kopien einer Regel bedeuten elf Änderungsorte ohne Informationsgewinn.
Vorschlag: Regel einmal über dem geteilten positionSchema in bestellung.go festhalten, die übrigen zehn Kopien löschen.
Aufwand: S · Status: unverifiziert (minor)

**Begründung für die fehlende Nachvalidierung trägt nicht** (minor · docs-accuracy, readability.md → Unnecessary Comments)
Datei: backend/domain/kasse/tisch_session_events.go:363-368
Warum: Der genannte Grund gilt genauso für die vier Geschwister-Builder, die sehr wohl erneut validieren.
Vorschlag: Auf den unterscheidenden Grund kürzen: der Korrektur-Kommentar ist optional, stornierungSchema fordert Min(3).
Aufwand: S · Status: unverifiziert (minor)

### backend/api/druck/bondruck/application/escpos/formatter.go

**Alle gedruckten Zeitstempel ohne Europe/Berlin** (major · correctness, docs/compliance.md §5.3)
Datei: backend/api/druck/bondruck/application/escpos/formatter.go:113-259
Warum: Im Sommer zeigen Kassenbeleg und Arbeitsbon zwei Stunden vor der Wanduhr, ohne Zeitzonenangabe.
Vorschlag: Paket-Variable `berlin` laden und jeden Zeitstempel als `zeitpunkt.In(berlin)` formatieren; Test mit UTC-Eingabe ergänzen.
Aufwand: S · Status: bestätigt

**QR-Kapazitätstabelle enthält Datencodewörter statt Byte-Kapazitäten** (minor · correctness, ISO/IEC 18004 Tab. 7)
Datei: backend/api/druck/bondruck/application/escpos/formatter.go:461-497
Warum: In schmalen Längenbändern wird eine zu kleine Version gewählt, der TSE-QR-Code wird abgeschnitten.
Vorschlag: Array durch die ECL-M-Byte-Kapazitäten (14, 26, 42, …) ersetzen und die drei hartkodierten Testwerte anpassen.
Aufwand: S · Status: unverifiziert (minor)

**Stornobeleg-Flag setzt nur den Titel, der Aufrufer muss negieren** (minor · code-smell, code-smells.md → Leaky Abstraction)
Datei: backend/api/druck/bondruck/application/escpos/formatter.go:33-35
Warum: Ein zweiter Aufrufer erzeugt ohne Vorwissen einen Stornobeleg mit positiven Beträgen.
Vorschlag: Entweder im Formatter negieren und die zwei Helfer streichen oder das Feld `TitelStorno` nennen — nicht beides.
Aufwand: M · Status: unverifiziert (minor)

**toWPC1252 auf den Formatstring statt auf das Ergebnis** (minor · code-smell, readability.md → Least Surprise)
Datei: backend/api/druck/bondruck/application/escpos/formatter.go:304
Warum: Die Codepage-Konvertierung greift auf der falschen Seite und ist nur zufällig ein No-Op.
Vorschlag: `buf.WriteString(toWPC1252(fmt.Sprintf("  Nachsigniert am %s\n", …)))` wie in Zeile 275.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/druck/bondruck/application/arbeitsbon_policy.go

**Event-JSON wird in den ungetaggten Domain-Typ kasse.Position dekodiert** (major · boundary-consistency, AGENTS.md Regel 10)
Datei: backend/api/druck/bondruck/application/arbeitsbon_policy.go:15-21
Warum: Der Decode funktioniert nur über das case-insensitive Fallback und ist an keinen eingefrorenen Vertrag gebunden.
Vorschlag: `Positionen []kasse.PositionEventData` deklarieren und mit `kasse.PositionFromEventData` mappen, wie kassenbeleg_command.go:22-28.
Aufwand: S · Status: bestätigt

**Kategorien als rohe String-Literale statt Domänen-Konstanten** (minor · convention, readability.md → Naming)
Datei: backend/api/druck/bondruck/application/arbeitsbon_policy.go:60-144
Warum: Eine Umbenennung im Domänenpaket kompiliert sauber und deaktiviert stillschweigend Abholbon-Regel und Küchen-Beep.
Vorschlag: `druckstationen[string(druckstation.KategorieAbholbon)]` und `kategorie == string(druckstation.KategorieEssen)` verwenden.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/druck/bondruck/application/arbeitsbon_policy_test.go

**Fixtures marshallen die Struktur, die sie prüfen sollen** (major · test-quality, test-quality → self-referential fixtures)
Datei: backend/api/druck/bondruck/application/arbeitsbon_policy_test.go:18-47
Warum: Die Tests bleiben grün, selbst wenn Decode-Struct und echtes Event-JSON vollständig auseinanderlaufen.
Vorschlag: Events mit `kasse.NewBestellungAufgenommenEvent` bzw. `kasse.NewDirektverkaufGetaetigtEvent` bauen, wie kassenbeleg_command_test.go.
Aufwand: M · Status: bestätigt

### backend/api/fiskal/dsfinvk/mapper.go

**ABRECHNUNGSKREIS kann die amtliche MaxLength 50 überschreiten** (major · correctness, DSFinV-K 2.4 index.xml)
Datei: backend/api/fiskal/dsfinvk/mapper.go:535-544
Warum: Tischnamen dürfen 100 Zeichen haben, das Archiv widerspricht dann seiner eigenen index.xml.
Vorschlag: In `abrechnungskreis()` runensicher auf 50 kürzen oder TischNameSchema auf 50 begrenzen; Mapper-Test ergänzen.
Aufwand: S · Status: bestätigt

**buildVat-Doc beschreibt das Gegenteil des Codes** (minor · docs-accuracy, readability.md → Prose Slop)
Datei: backend/api/fiskal/dsfinvk/mapper.go:632-665
Warum: Der Kommentar verspricht sitzungsabhängige Steuersätze, der Rumpf schreibt immer alle sieben amtlichen Schlüssel.
Vorschlag: Doc auf das tatsächliche Verhalten umschreiben und den ungenutzten Parameter `_ []beleg` samt Aufrufstelle entfernen.
Aufwand: S · Status: unverifiziert (minor)

**„byte-identisch zum bisherigen Export"** (minor · convention, AGENTS.md Regel 18)
Datei: backend/api/fiskal/dsfinvk/mapper.go:494-498
Warum: Der Vergleich zu einem nicht mehr existierenden Stand trägt keine Information.
Vorschlag: Nur die Regel nennen (Autotext allein bzw. mit „; " verkettet); dieselbe Kürzung in mapper_test.go:749-750.
Aufwand: S · Status: unverifiziert (minor)

**Archive-Doc behauptet fehlende slaves.csv und pa.csv** (minor · docs-accuracy, AGENTS.md Regel 18)
Datei: backend/api/fiskal/dsfinvk/mapper.go:47-50
Warum: Beide Tabellen werden header-only ausgegeben und die amtliche index.xml deklariert alle 20.
Vorschlag: Doc auf „eine Table je CSV in amtlicher Reihenfolge, nicht befüllte Tabellen header-only" ändern.
Aufwand: S · Status: unverifiziert (minor)

**zahlartReihenfolge als Ein-Eintrag-Prioritätsmap** (minor · principle, principles.md → YAGNI)
Datei: backend/api/fiskal/dsfinvk/mapper.go:1139-1167
Warum: Neun Zeilen Sortier-Maschinerie für ein Produkt, das Kartenzahlung bewusst ausschließt.
Vorschlag: Map und Sortierung streichen; bei einer zweiten Zahlart wieder einführen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/fiskal/dsfinvk/table.go

**Tote Spaltenmetadaten `typ` und `accuracy`** (major · code-smell, code-smells.md → Dead Code)
Datei: backend/api/fiskal/dsfinvk/table.go:13-34
Warum: Die Felder werden an ~20 Tabellen gesetzt, nirgends gelesen, und der Kommentar behauptet eine index.xml-Erzeugung, die es nicht gibt.
Vorschlag: `column` auf `name` reduzieren — oder den Index-Test gegen `<AlphaNumeric/>`/`<Accuracy>` erweitern; Doc in jedem Fall korrigieren.
Aufwand: S · Status: bestätigt

### backend/api/middleware/middleware.go

**BEDIENER_NAME stammt aus dem 12-Stunden-JWT statt aus dem frischen Datensatz** (major · correctness, docs/handbuch.md:81)
Datei: backend/api/middleware/middleware.go:269-304
Warum: Ein umbenannter Bediener schreibt bis zu 12 Stunden einen veralteten Namen ins unveränderliche Kassenjournal.
Vorschlag: `ctx = context.WithValue(ctx, UserNameKey, u.Username)` aus dem bereits geladenen Datensatz; Test in middleware_test.go:512-541 anpassen.
Aufwand: S · Status: bestätigt

**Client-Fehler auf Error-Level, nicht rate-limitiert** (minor · ops, code-smells.md → Log Level)
Datei: backend/api/middleware/middleware.go:205-274
Warum: Abgelaufene Tokens und falsche Methoden füllen das gedeckelte Docker-Log und verdrängen echte Serverfehler.
Vorschlag: Zeilen 206, 257, 264 und 271 auf `Warn` senken, wie die Zweige 279/288/296 im selben Handler.
Aufwand: S · Status: unverifiziert (minor)

**POST-only-Ausnahme /health nur im Handbuch dokumentiert** (minor · docs-accuracy, AGENTS.md Regel 1)
Datei: backend/api/middleware/middleware.go:199-203
Warum: Die beiden zuerst gelesenen Regeldateien nennen ein Absolut, das der Code bewusst bricht.
Vorschlag: In AGENTS.md Regel 1 und copilot-instructions Guardrail 1 die Ausnahme „GET /health für Healthchecks" ergänzen.
Aufwand: S · Status: unverifiziert (minor)

**RateLimitMiddleware startet eine nicht stoppbare Goroutine** (minor · ops, code-smells.md → Hidden Lifecycle)
Datei: backend/api/middleware/middleware.go:135-140
Warum: Der Konstruktor hat einen undokumentierten Seiteneffekt, jeder Test hinterlässt eine laufende Goroutine.
Vorschlag: Lebensdauer im Doc-Kommentar benennen, wie throttle.go:137-139; ein Stop-Kanal wäre eine Verhaltensänderung.
Aufwand: M · Status: unverifiziert (minor)

**Drei Rückgabeformen für dieselbe Middleware-Art** (minor · convention, principles.md → Least Surprise)
Datei: backend/api/middleware/middleware.go:194-195
Warum: Die Verengung auf `http.HandlerFunc` lässt die Kette in app.go wie verschiedene Konzepte aussehen.
Vorschlag: `http.Handler` bzw. `func(http.Handler) http.Handler` zurückgeben und die Wortdopplung im Doc-Kommentar entfernen.
Aufwand: S · Status: unverifiziert (minor)

**Drei Kommentarstile in einer Datei** (minor · readability, AGENTS.md Bewertungsmetriken → Konsistenz)
Datei: backend/api/middleware/middleware.go:80-84
Warum: Englisch, transliteriertes Deutsch und Deutsch mit Umlauten stehen unmotiviert nebeneinander.
Vorschlag: Die transliterierten Blöcke (80-84, 118-124, 226-233) auf echte Umlaute umschreiben.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/stammdaten/tisch/http/command_handler.go

**ErrTischAlreadyExists beim Umbenennen nicht gemappt** (major · correctness, AGENTS.md Regel 9)
Datei: backend/api/stammdaten/tisch/http/command_handler.go:77-84
Warum: Ein Namenskonflikt liefert 500 statt 400 `tisch_already_exists`, die Frontend-Meldung greift nie.
Vorschlag: `application.ErrTischAlreadyExists: "tisch_already_exists"` in die Fehler-Map von TischAktualisierenHandler aufnehmen.
Aufwand: S · Status: bestätigt

### backend/api/stammdaten/produkt/application/command.go

**UpdateProdukt klassifiziert db.ErrAlreadyExists nicht** (major · correctness, AGENTS.md Regel 9)
Datei: backend/api/stammdaten/produkt/application/command.go:76-80
Warum: Derselbe Unique-Konflikt liefert beim Anlegen 400 und beim Umbenennen 500.
Vorschlag: `errors.Is(err, db.ErrAlreadyExists)` auf ErrProduktAlreadyExists abbilden und den Code im Handler ergänzen.
Aufwand: S · Status: bestätigt

**DeleteVariante prüft die Zugehörigkeit zum Produkt nicht** (minor · correctness, code-smells.md → Defensive Overkill)
Datei: backend/api/stammdaten/produkt/application/command.go:245-278
Warum: Das geladene Produkt wird verworfen, jede fremde varianteId wird trotzdem gelöscht.
Vorschlag: Entweder die Zugehörigkeit prüfen und ErrVarianteNotFound liefern oder produktID aus Command, DTO und Schema entfernen.
Aufwand: S · Status: unverifiziert (minor)

**16 Error-Logs ohne .Err(err)** (minor · ops, zerolog-Konvention des Repos)
Datei: backend/api/stammdaten/produkt/application/command.go:49
Warum: Ein Produktionsfehler hinterlässt weder in der Antwort noch im Log eine Ursache.
Vorschlag: `.Err(err)` an allen betroffenen Stellen in produkt/command.go, produkt/query.go und user/command.go ergänzen.
Aufwand: S · Status: unverifiziert (minor)

**Command-Port deklariert nie aufgerufene Query-Methoden** (minor · architecture, principles.md → Interface Segregation)
Datei: backend/api/stammdaten/produkt/application/command.go:13-26
Warum: Jeder Mock muss Methoden erfüllen, die die Command-Schicht nicht benutzt.
Vorschlag: GetAllProdukte/GetActiveProdukte hier, GetAllTables in tisch und GetAllUsers in user aus den Command-Ports streichen.
Aufwand: S · Status: unverifiziert (minor)

**Lokale Variablen verdecken das gleichnamige Domänenpaket** (minor · readability, readability.md → Naming)
Datei: backend/api/stammdaten/produkt/application/command.go:37-46
Warum: `produkt` bedeutet in einer Zeile das Paket und in der nächsten den Wert.
Vorschlag: Werte `p`, `t`, `u` benennen, wie die Helfer im selben Paket es bereits tun.
Aufwand: S · Status: unverifiziert (minor)

**Transliterierter Doc-Kommentar** (minor · readability, AGENTS.md Bewertungsmetriken → Konsistenz)
Datei: backend/api/stammdaten/produkt/application/command.go:280-282
Warum: „gewoehnliche"/„laesst" stehen neben Umlaut-Kommentaren derselben Datei.
Vorschlag: „gewöhnliche" und „lässt" schreiben.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/stammdaten/betreiber/http/command_handler.go

**HTTP-Schema dupliziert die Domänen-Validierung samt Meldungen** (major · principle, principles.md → DRY)
Datei: backend/api/stammdaten/betreiber/http/command_handler.go:31-51
Warum: Zwei Wahrheitsquellen für Stammdaten, die in Beleg und DSFinV-K landen; der NewBetreiber-Zweig antwortet dabei mit 500.
Vorschlag: Feld-Schemas aus domain/betreiber exportieren und `updateBetreiberSchema` daraus bauen, wie user/tisch/produkt es tun.
Aufwand: S · Status: bestätigt

### backend/repository/tisch_repo/repo.go

**Öffentliche API übersetzt das Domänen-Nomen Tisch nach Table** (major · convention, docs/language.md Regel 1/5)
Datei: backend/repository/tisch_repo/repo.go:11-146
Warum: Persistenz spricht eine andere Sprache als Domäne, DB und sqlc; `DeleteTableMitFavoriten` mischt beide in einem Bezeichner.
Vorschlag: Auf GetTisch/GetAlleTische/CreateTisch/… umbenennen und die vier Consumer-Interfaces mitziehen — oder die Abweichung in docs/language.md festhalten.
Aufwand: M · Status: bestätigt

**Offener Saldo wird außerhalb der Lösch-Transaktion geprüft** (minor · correctness, docs/handbuch.md → Invarianten)
Datei: backend/repository/tisch_repo/repo.go:156-173
Warum: Eine dazwischen committende Bestellung erzeugt einen gelöschten Tisch mit saldo_cents > 0.
Vorschlag: Den Guard `TischHatOffenenSaldo` in dieselbe Transaktion ziehen wie den Soft-Delete.
Aufwand: S · Status: unverifiziert (minor)

### backend/repository/user_repo/types.go

**Drei byte-identische Row-Mapper** (major · principle, principles.md → DRY)
Datei: backend/repository/user_repo/types.go:19-62
Warum: Ein neues Feld muss dreimal ergänzt werden; ein vergessener Pfad verliert es still, etwa bei SetPasswordTx.
Vorschlag: Einen Mapper behalten und am Aufrufort konvertieren (`userRowToDomain(dbgen.GetUserRow(row))`) oder die SELECT-Spalten angleichen.
Aufwand: S · Status: bestätigt

### backend/repository/produkt_repo/repo.go

**GetAllProdukte und GetActiveProdukte sind zeilengleich** (major · principle, principles.md → DRY)
Datei: backend/repository/produkt_repo/repo.go:38-92
Warum: Dieselbe Varianten-JSON- und Feldabbildung existiert dreimal in einer Datei.
Vorschlag: `produktRowToDomain` nach types.go ziehen und aus allen drei Lesepfaden aufrufen.
Aufwand: S · Status: bestätigt

**Repository erfindet eigenen updated_at-Zeitstempel** (minor · architecture, architecture.md → Inversion of Control)
Datei: backend/repository/produkt_repo/repo.go:163-168
Warum: Für Reihenfolge und Sortieren entscheidet die Infrastruktur einen Domänenwert, sonst die Domäne.
Vorschlag: Zeitstempel vom Aufrufer übernehmen (wie UpdateProdukt) oder den Bump in reinen Ordnungsoperationen weglassen.
Aufwand: S · Status: unverifiziert (minor)

**Reihenfolge-Algorithmus für Produkt und Variante doppelt** (minor · code-smell, code-smells.md → Structural Smells)
Datei: backend/repository/produkt_repo/repo.go:142-288
Warum: Drei Funktionspaare implementieren dieselbe normalize-then-swap-Regel auf ~95 Zeilen.
Vorschlag: Entweder den Tausch einmal parametrisieren oder die bewusste Dopplung in einem Kommentar festhalten.
Aufwand: M · Status: unverifiziert (minor)

**Kommentarsprache und Umlaut-Schreibweise gemischt** (minor · convention, AGENTS.md Bewertungsmetriken → Konsistenz)
Datei: backend/repository/produkt_repo/repo.go:290-295
Warum: Deutsch und Englisch sowie „genuegt" neben „Hälfte" stehen in einer Datei.
Vorschlag: Pro Paket eine Kommentarsprache und durchgängig echte Umlaute; Umfang bewusst über die Unit hinaus planen.
Aufwand: M · Status: unverifiziert (minor)

### backend/repository/kassenjournal_repo/repo.go

**Doc-Kommentar beschreibt einen Ablauf, den der Code nicht hat** (major · docs-accuracy, AGENTS.md Regel 18)
Datei: backend/repository/kassenjournal_repo/repo.go:264-267
Warum: Der kassensitzungen-Insert passiert in `EroeffneKassensitzung` in derselben Transaktion, nicht vorher in der Anwendungsschicht.
Vorschlag: Kommentar auf den echten Ablauf umschreiben, die Routing-Zeile 38 auf `UPDATE` korrigieren und repo_test.go:1519-1521 mitziehen.
Aufwand: S · Status: bestätigt

**Zwei kaum unterscheidbare Lesemethoden mit verschiedenem Ergebnis** (minor · readability, readability.md → Naming)
Datei: backend/repository/kassenjournal_repo/repo.go:581-636
Warum: ReadEventsByKassensitzung liefert zusätzlich Signaturen, ReadKassensitzungEvents nicht — der Name sagt es nicht.
Vorschlag: Die Export-Variante `ReadKassensitzungEventsMitSignaturen` nennen und die zwei Interfaces anpassen.
Aufwand: S · Status: unverifiziert (minor)

**Vier identische Row-zu-Event-Mapper** (minor · principle, principles.md → DRY)
Datei: backend/repository/kassenjournal_repo/repo.go:486-528
Warum: Jede neue Event-Spalte muss an vier Stellen ergänzt werden, nur weil sqlc je Query einen Row-Typ erzeugt.
Vorschlag: Die vier Queries auf dieselbe Spaltenform bringen (`sqlc.embed`) und einen Mapper behalten.
Aufwand: M · Status: unverifiziert (minor)

**Roher Event-JSON-Key als Repository-Parameter** (minor · code-smell, code-smells.md → Leaky Abstraction)
Datei: backend/repository/kassenjournal_repo/repo.go:649-666
Warum: Der eingefrorene JSON-Vertrag liegt als String-Literal in drei Anwendungs-Paketen, ein Tippfehler meldet still „kein Event".
Vorschlag: Den Key im Repository aus dem Event-Typ ableiten oder eine typisierte Konstante aus domain/kasse entgegennehmen.
Aufwand: M · Status: unverifiziert (minor)

### backend/repository/kassenjournal_repo/repo_test.go

**Test ohne jede Zusicherung** (major · test-quality, code-smells.md → Dead Code)
Datei: backend/repository/kassenjournal_repo/repo_test.go:1583-1592
Warum: `TestGetOffeneKassensitzung_NoneOpen` ruft die namensgebende Methode nie auf, sie existiert im Paket gar nicht.
Vorschlag: Test löschen und die Zusicherung in kassensitzungen_repo ergänzen: nach Abschluss liefert GetOffeneKassensitzung `(nil, nil)`.
Aufwand: S · Status: bestätigt

**Reporting-Query wird über das private `q`-Feld getestet** (minor · boundary-consistency, architecture.md → Bounded Contexts)
Datei: backend/repository/kassenjournal_repo/repo_test.go:1308-1350
Warum: Eine Änderung an der Reporting-Aggregation bricht einen Test im fremden Paket.
Vorschlag: Test nach backend/repository/reporting_repo verschieben und über die Repository-Methode zusichern.
Aufwand: S · Status: unverifiziert (minor)

**Fixture-Fehler werden verworfen** (minor · test-quality, principles.md → Fail Fast)
Datei: backend/repository/kassenjournal_repo/repo_test.go:60-64
Warum: Ein ungültiges Fixture erzeugt ein Null-Event, der Test scheitert später mit irreführender Meldung.
Vorschlag: `newTestEvent` `*testing.T` entgegennehmen und `t.Fatalf` rufen; die `_, _ = insertEventRaw`-Aufrufe prüfen.
Aufwand: S · Status: unverifiziert (minor)

**Toter `_ = userID`-Ausdruck** (minor · code-smell, code-smells.md → Dead Code)
Datei: backend/repository/kassenjournal_repo/repo_test.go:959
Warum: Die Zuweisung existiert nur, um eine ungenutzte Variable stillzustellen.
Vorschlag: Zeile 937 auf `_, _, repo, teardown := setup(t)` ändern und Zeile 959 löschen.
Aufwand: S · Status: unverifiziert (minor)

### backend/repository/druckauftrag_repo/repo.go

**MaxDruckversuche ist 6, die Dokumentation sagt 3** (major · docs-accuracy, AGENTS.md Regel 18)
Datei: backend/repository/druckauftrag_repo/repo.go:14-16
Warum: Betreiber erwarten den Abbruch nach drei Fehlversuchen und fehldiagnostizieren laufende Retries.
Vorschlag: docs/handbuch.md:308 und docs/language.md:425 auf sechs korrigieren und das Backoff-Schema dokumentieren.
Aufwand: S · Status: bestätigt

**Kommentar vergleicht mit einem früheren Status-Guard** (minor · convention, AGENTS.md Regel 18)
Datei: backend/repository/druckauftrag_repo/repo.go:137-140
Warum: Rule 18 verbietet „bisher"-Prosa; der Vergleichsstand ist nicht mehr nachlesbar.
Vorschlag: Auf „idempotenter No-Op (Status-Guard 'offen')" kürzen.
Aufwand: S · Status: unverifiziert (minor)

**Umlaute und Transliteration in einer Datei** (minor · convention, AGENTS.md Bewertungsmetriken → Konsistenz)
Datei: backend/repository/druckauftrag_repo/repo.go:163-203
Warum: „Aufträge/älteste" und „Auftraege/unberuehrt" beschreiben dasselbe Konzept 40 Zeilen auseinander.
Vorschlag: In dieser Unit eine Schreibweise wählen; die Repository-Pakete sind überwiegend transliteriert.
Aufwand: S · Status: unverifiziert (minor)

### docs/handbuch.md

**Kassenbeleg-Quellen unvollständig dokumentiert** (major · docs-accuracy, AGENTS.md Regel 18)
Datei: docs/handbuch.md:306
Warum: Der Endpunkt erzeugt vier Belegformen, dokumentiert sind nur Tischzahlung und Direktverkauf.
Vorschlag: Alle vier Quellen nennen und in docs/compliance.md:280 die Form `tischId` + `stornierungId` ergänzen.
Aufwand: S · Status: bestätigt

**Einmalpasswort-Sperre nach fünf Fehlversuchen nicht dokumentiert** (minor · docs-accuracy, AGENTS.md Regel 18)
Datei: docs/handbuch.md:363-367
Warum: Ein ehrenamtlicher Admin kann nicht wissen, warum ein korrektes Einmalpasswort plötzlich abgelehnt wird.
Vorschlag: Einen Satz in §5.2 und den Benutzer-Invarianten ergänzen und in docs/language.md §Einmalpasswort spiegeln.
Aufwand: S · Status: unverifiziert (minor)

### .github/instructions/backend.instructions.md

**Auth-Abschnitt beschreibt eine Middleware, die es nicht gibt** (major · docs-accuracy, AGENTS.md Regel 11/18)
Datei: .github/instructions/backend.instructions.md:49-50
Warum: Die Rolle liegt nie im Request-Context, und der `username`-Claim fehlt in der Aufzählung.
Vorschlag: Claims (iss, iat, exp, sub, username, role) und Context-Keys (UserIDKey, UserNameKey) korrekt benennen, Rolle aus dem DB-Datensatz erwähnen.
Aufwand: S · Status: bestätigt

**Verzeichnisbaum ohne bootstrap/, seed/ und dsfinvkpruefung/** (minor · docs-accuracy, AGENTS.md Regel 18)
Datei: .github/instructions/backend.instructions.md:22-41
Warum: Drei Top-Level-Pakete fehlen in der Karte, kleinere Geschwister sind gelistet.
Vorschlag: Die drei Zeilen mit einer Kurzbeschreibung ergänzen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/kasse/tischgeschaeft/application/command.go

**db.ErrConflict in BestellungUmbuchen nicht gemappt** (minor · correctness, AGENTS.md Regel 9)
Datei: backend/api/kasse/tischgeschaeft/application/command.go:347-364
Warum: Umbuchung ist der einzige Pfad mit zwei Zeilensperren; der Deadlock-Verlierer bekommt 500 statt eines wiederholbaren 409.
Vorschlag: `errors.Is(err, db.ErrConflict)` wie in writeEventOCC ergänzen.
Aufwand: S · Status: unverifiziert (minor)

**getOffeneKassensitzungOderFehler viermal kopiert** (minor · principle, principles.md → DRY)
Datei: backend/api/kasse/tischgeschaeft/application/command.go:65-80
Warum: Die Buchungsbarriere ist eine Fiskal-Invariante; vier Kopien lassen eine Regeländerung eine davon verfehlen.
Vorschlag: Klassifikation der Kassensitzung als Prädikat in domain/kasse; jedes Paket mappt nur noch auf eigene Sentinels.
Aufwand: M · Status: unverifiziert (minor)

**Zwei unerreichbare Zweige im Umbuchungs-Kommentarbau** (minor · code-smell, code-smells.md → Defensive Overkill)
Datei: backend/api/kasse/tischgeschaeft/application/command.go:255-276
Warum: Beide Aufrufer übergeben ein 14-Runen-Präfix, die Guards können nie greifen.
Vorschlag: Die zwei toten Zweige entfernen, die Runen-Kürzung des Tischnamens behalten.
Aufwand: S · Status: unverifiziert (minor)

**Fünfstelliges Rückgabetupel mit drei unbenannten Skalaren** (minor · readability, readability.md → Naming)
Datei: backend/api/kasse/tischgeschaeft/application/command.go:153-180
Warum: Vertauschte Strings kompilieren und äußern sich erst als falscher Umbuchungskommentar.
Vorschlag: Kleines benanntes Struct `tischKontext{Subject, KassensitzungNr, TischName, Session}` zurückgeben.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/kasse/tischgeschaeft/http/query_handler.go

**Query-Handler validieren nichts und mappen ErrTischNotFound auf 500** (minor · boundary-consistency, AGENTS.md Regel 5/9)
Datei: backend/api/kasse/tischgeschaeft/http/query_handler.go:253-313
Warum: Dieselbe Ursache liefert im Command-Handler 400 `tisch_not_found`, hier 500.
Vorschlag: Body mit `tisch.TischIDSchema` über ReadAndValidateBody prüfen und `helper.MapError` mit `tisch_not_found` verwenden.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/kasse/kassenfuehrung/application/errors.go

**ErrKasseAlreadyAbgeschlossen wird nirgends erzeugt** (minor · code-smell, code-smells.md → Dead Code)
Datei: backend/api/kasse/kassenfuehrung/application/errors.go:21-22
Warum: Ein dokumentiertes Sentinel ohne Erzeuger verleitet dazu, den Fall für abgedeckt zu halten.
Vorschlag: Variable und Kommentar löschen; der Fall wird von ErrKasseNichtGeoeffnet getragen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/kasse/kassenfuehrung/http/query_handler.go

**Zwei identische Request-Structs mit handgebautem Guard** (minor · code-smell, AGENTS.md Regel 5)
Datei: backend/api/kasse/kassenfuehrung/http/query_handler.go:32-48
Warum: Die manuelle Prüfung erzeugt eine andere Fehlerform als jeder Command-Handler der Unit.
Vorschlag: Einen `kassensitzungNrRequest` mit zog-Schema (`z.Int().GTE(1)`) über ReadAndValidateBody verwenden.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/kasse/kassenfuehrung/http/command_handler.go

**409-Details führen nur `anzahl`, die Doku verspricht auch das Alter** (minor · docs-accuracy, AGENTS.md Regel 18)
Datei: backend/api/kasse/kassenfuehrung/http/command_handler.go:73-77
Warum: DTO, Gate-Struct und Frontend kennen kein Alters-Feld, Handbuch und helper-Kommentar behaupten es.
Vorschlag: docs/handbuch.md:242 und die Kommentare in api/helper/http.go auf die reine Anzahl korrigieren.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/kasse/direktverkauf/application/command.go

**Nil-Guard für DruckstationRepo mit falscher Begründung** (minor · docs-accuracy, code-smells.md → Defensive Overkill)
Datei: backend/api/kasse/direktverkauf/application/command.go:127-135
Warum: Der Kommentar nennt Tests, tatsächlich verdrahtet api/serviceleitung.go die Command ohne dieses Repo.
Vorschlag: Den echten Grund nennen (Storno erzeugt keine Druckaufträge) oder den Guard streichen und einen Stub verdrahten.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/kasse/direktverkauf/http/command_handler.go

**PositionRef- und Position-Input-DTOs wortgleich dupliziert** (minor · code-smell, principles.md → DRY)
Datei: backend/api/kasse/direktverkauf/http/command_handler.go:97-113
Warum: Beide Kopien mappen auf dieselben geteilten Typen; eine Regeländerung muss zweimal erfolgen.
Vorschlag: DTOs, zog-Schemas und Mapper neben `enrichment.PositionInput` ziehen, das bereits für beide Pakete geteilt wird.
Aufwand: M · Status: unverifiziert (minor)

### backend/api/kasse/tischgeschaeft/application/command_test.go

**Eingefrorener Event-Vertrag im Test handkopiert** (minor · test-quality, principles.md → DRY)
Datei: backend/api/kasse/tischgeschaeft/application/command_test.go:97-115
Warum: Die lokale Kopie hat bereits `steuersatz` verloren, ohne dass ein Test scheitert.
Vorschlag: In `kasse.BestellungUmgebuchtV1Data` unmarshallen, die 17 inline-Command-Literale durch die vorhandenen Konstruktoren ersetzen.
Aufwand: M · Status: unverifiziert (minor)

**Drei ungenutzte Methoden am Repo-Mock** (minor · code-smell, code-smells.md → Dead Code)
Datei: backend/api/kasse/tischgeschaeft/application/command_test.go:129-139
Warum: Der Double wirkt breiter als das Interface, das er erfüllt, und wird nicht gelintet.
Vorschlag: CreateTable, UpdateTable und GetAllTables löschen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/kasse/tischgeschaeft/http/command_handler_test.go

**Sechszeiliges Request-Setup in allen sieben Tests wiederholt** (minor · test-quality, principles.md → DRY)
Datei: backend/api/kasse/tischgeschaeft/http/command_handler_test.go:38-155
Warum: Die Schwesterpakete haben dafür bereits `requestWithUser`, hier entstehen zwei Konventionen.
Vorschlag: Denselben Helfer übernehmen oder auf einen tabellengetriebenen Test über {Fehler, Body, Status} umstellen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/kasse/kassenfuehrung/application/command_test.go

**Test-Doc definiert Verhalten über eine entfernte Implementierung** (minor · convention, AGENTS.md Regel 18)
Datei: backend/api/kasse/kassenfuehrung/application/command_test.go:301-303
Warum: „nicht mehr aus einem separaten Reporting-Repository" beschreibt einen Stand, den der Code nicht hat.
Vorschlag: Positiv formulieren: Summen werden aus den Journal-Events der Kassensitzung berechnet.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/fiskal/export/application/export.go

**Archiv-Dateiname ohne UTC-Normalisierung** (minor · correctness, docs/compliance.md → Archivierung)
Datei: backend/api/fiskal/export/application/export.go:117-119
Warum: Dateiname und Z_ERSTELLUNG desselben Exports können verschiedene Uhrzeiten nennen.
Vorschlag: In `dateiname` `zeitpunkt.UTC()` formatieren, wie `Map` es tut.
Aufwand: S · Status: unverifiziert (minor)

**resolveKassensitzung ohne jeden Test** (minor · test-quality, AGENTS.md Self-Review 6)
Datei: backend/api/fiskal/export/application/export.go:124-155
Warum: Der „jüngste abgeschlossene"-Fallback hängt an einer nur kommentierten Sortiergarantie.
Vorschlag: `export_test.go` (`//go:build unit`) mit Fake-Repos für die fünf Zweige, inklusive Auswahl der jüngsten Sitzung.
Aufwand: M · Status: unverifiziert (minor)

### backend/api/fiskal/dsfinvk/archive.go

**Doc nennt einen index.xml-Generator, den es nicht gibt** (minor · docs-accuracy, AGENTS.md Regel 18)
Datei: backend/api/fiskal/dsfinvk/archive.go:34-38
Warum: Die amtliche index.xml wird unverändert eingebettet, nicht erzeugt — genau das prüft der Test.
Vorschlag: Doc auf „legt die amtliche index.xml und die DTD unverändert bei" ändern.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/fiskal/dsfinvk/dsfinvk.go

**„konfigurierbar gehalten" ohne jede Konfiguration** (minor · docs-accuracy, AGENTS.md Regel 18)
Datei: backend/api/fiskal/dsfinvk/dsfinvk.go:18-21
Warum: Der Versionsstring ist eine Konstante; docs/compliance.md:296 wiederholt die falsche Zusage gegenüber dem Betreiber.
Vorschlag: „An einer Stelle als Konstante gehalten, ein Versionswechsel bleibt ein Ein-Zeilen-Change" — und compliance.md angleichen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/fiskal/setup/application/setup.go

**Verweis auf eine „spaetere Phase", die bereits implementiert ist** (minor · docs-accuracy, AGENTS.md Regel 18)
Datei: backend/api/fiskal/setup/application/setup.go:150-156
Warum: `UebernimmTSE` steht 60 Zeilen tiefer in derselben Datei.
Vorschlag: Den Klammerzusatz streichen und nur auf `UebernimmTSE` verweisen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/fiskal/setup/application/errors.go

**Englische snake_case-Sentinels neben deutschen Wire-Codes** (minor · convention, docs/language.md)
Datei: backend/api/fiskal/setup/application/errors.go:9-19
Warum: `tse_setup_credentials_invalid` sieht wie ein Wire-Code aus, gesendet wird `tse_setup_zugangsdaten_ungueltig`.
Vorschlag: Sentinel-Texte auf die tatsächlich gesendeten Codes setzen — oder alle auf reine Prosa.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/fiskal/setup/http/command_handler.go

**Retry-Budget eines anderen Pakets zweimal in Prosa nachgerechnet** (minor · docs-accuracy, handbook → Single Source of Truth)
Datei: backend/api/fiskal/setup/http/command_handler.go:15-62
Warum: Eine Konstantenänderung in fiskaly_client.go verfälscht stillschweigend beide Kommentarblöcke.
Vorschlag: Die Rechnung einmal über `tseSetupWriteTimeout` festhalten, die zweite Konstante darauf verweisen lassen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/fiskal/signatur/tse_signatur_worker.go

**Deutsche Prosa je Unterpaket unterschiedlich geschrieben** (minor · convention, AGENTS.md Bewertungsmetriken → Konsistenz)
Datei: backend/api/fiskal/signatur/tse_signatur_worker.go:199-213
Warum: Log-Zeilen desselben Subsystems lesen sich wie zwei Projekte; jede Suche nach einem Begriff braucht zwei Läufe.
Vorschlag: In api/fiskal/setup und api/fiskal/signatur auf echte Umlaute umstellen (Mehrheit im Backend).
Aufwand: M · Status: unverifiziert (minor)

### backend/api/fiskal/dsfinvk/mapper_test.go

**Lokale Variable `tse` verdeckt das gleichnamige Paket** (minor · readability, readability.md → Naming)
Datei: backend/api/fiskal/dsfinvk/mapper_test.go:1147
Warum: Im selben Test bedeutet `tse` zeilenabhängig Paket oder Tabelle; ein späterer Paketzugriff kompiliert nicht.
Vorschlag: Lokale in `tseTab` umbenennen (Zeile 1268 macht es bereits so); auch die `min`-Shadows in den Handler-Tests.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/fiskal/dsfinvk/archive_test.go

**equalStrings implementiert slices.Equal nach** (minor · code-smell, code-smells.md → Redundant Abstractions)
Datei: backend/api/fiskal/dsfinvk/archive_test.go:70-80
Warum: Zwei Vergleichsstile koexistieren in einem Paket, obwohl die Standardbibliothek reicht.
Vorschlag: `slices.Equal(got, want)` verwenden und den Helfer löschen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/fiskal/export/http/handler_test.go

**deadlineCapturingWriter dreimal im Repo** (minor · test-quality, principles.md → DRY)
Datei: backend/api/fiskal/export/http/handler_test.go:44-79
Warum: Wie Write-Deadlines beobachtet werden, ist ein Wissen an drei Orten.
Vorschlag: Writer in ein internes Test-Helper-Paket ziehen und aus allen drei Tests importieren.
Aufwand: M · Status: unverifiziert (minor)

### backend/api/fiskal/setup/application/query_test.go

**Zwei Namen für dasselbe Fixture in einem Paket** (minor · test-quality, principles.md → DRY)
Datei: backend/api/fiskal/setup/application/query_test.go:266-268
Warum: `gueltigeZugangsdaten()` liefert exakt das, was `zugangsdaten()` liefert.
Vorschlag: `gueltigeZugangsdaten` löschen und die drei Aufrufstellen umbiegen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/fiskal/tse_live/tse_live_ausfall_test.go

**Verweis auf einen nicht existierenden QA-Block** (minor · docs-accuracy, AGENTS.md Git-Workflow)
Datei: backend/api/fiskal/tse_live/tse_live_ausfall_test.go:3-6
Warum: „Block 4" gibt es nicht, und das Ziel liegt in docs/plans/, das nach dem Merge gelöscht wird.
Vorschlag: Den existierenden Block nennen oder den externen Verweis streichen — die Aufzählung darunter trägt allein.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/druck/beleg/application/kassenbeleg_command.go

**Fehlender Signaturauftrag druckt still einen Beleg ohne TSE-Abschnitt** (minor · correctness, § 6 KassenSichV)
Datei: backend/api/druck/beleg/application/kassenbeleg_command.go:144-152
Warum: Alle vier Belegformen sind signaturpflichtig; der Zweig ist ein Fail-Open für ein inkonsistentes System.
Vorschlag: ErrNotFound hier als Fehler behandeln oder mindestens den Ausfallvermerk plus Warn-Log setzen.
Aufwand: S · Status: unverifiziert (minor)

**Vier nahezu identische Event-Finder** (minor · code-smell, code-smells.md → Structural Smells)
Datei: backend/api/druck/beleg/application/kassenbeleg_command.go:42-124
Warum: Eine Änderung der Lookup-Regel muss viermal erfolgen.
Vorschlag: Einen generischen `findEvent[T]`-Helfer einführen und die vier Wrapper nur behalten, wenn die Aufrufstellen davon lesbarer werden.
Aufwand: M · Status: unverifiziert (minor)

### backend/api/druck/beleg/http/command_handler.go

**Beide 409-Pfade des Beleg-Endpunkts ungetestet** (minor · test-quality, AGENTS.md Self-Review 6)
Datei: backend/api/druck/beleg/http/command_handler.go:85-89
Warum: Es sind die Barriere-Pfade des Kassenabschlusses und die einzigen ungetesteten Antworten des Handlers.
Vorschlag: Zwei Fälle mit mockCommand ergänzen, die 409 und den Code zusichern.
Aufwand: S · Status: unverifiziert (minor)

**Einziger Handler der Unit ohne `// POST /<route>`-Kommentar** (minor · convention, AGENTS.md Regel 1)
Datei: backend/api/druck/beleg/http/command_handler.go:76-77
Warum: Der Routen-Kommentar ist die einzige Stelle im Go-Code, die Handler und Pfad verbindet.
Vorschlag: `// POST /service/beleg-drucken` ergänzen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/druck/beleg/application/command.go

**Begründung verweist auf ein nicht auffindbares Review** (minor · docs-accuracy, AGENTS.md Regel 18)
Datei: backend/api/druck/beleg/application/command.go:55-59
Warum: „2026-07-17 multi-expert review" existiert nirgends im Repo, die Begründung ist nicht prüfbar.
Vorschlag: Datum und Review streichen, den sachlichen Grund behalten oder als ADR festhalten und verlinken.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/druck/station/http/handler.go

**Kategorie-Aufzählung dreimal, davon zweimal als Literal** (minor · principle, principles.md → DRY)
Datei: backend/api/druck/station/http/handler.go:72-81
Warum: Eine sechste Kategorie erfordert drei Änderungen in zwei Paketen, nur eine davon ist getestet.
Vorschlag: `druckstation.AlleKategorien()` aus der Domäne exportieren und beide `OneOf`-Aufrufe daraus speisen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/druck/station/http/handler_test.go

**Vier Zeilen Request-Aufbau in allen 15 Tests** (minor · test-quality, principles.md → DRY)
Datei: backend/api/druck/station/http/handler_test.go:40-264
Warum: Die eigentliche Zusicherung verschwindet unter identischem Setup.
Vorschlag: Den `postJSON`-Helfer des Schwesterpakets übernehmen und die Validierungsfälle zu einem Tabellentest zusammenziehen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/druck/relay/http/handler.go

**Falscher Relay-Token liefert 400 statt 401** (minor · convention, AGENTS.md Regel 9)
Datei: backend/api/druck/relay/http/handler.go:89-93
Warum: Authentifizierungsfehler sind in Logs nicht von fehlerhaften Payloads zu trennen.
Vorschlag: `helper.SendUnauthorized(w, "unauthorized")` an beiden Stellen; Tests anpassen, das Windows-Relay prüft nur auf 200.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/druck/bondruck/application/escpos/formatter_test.go

**Einzige Test-Datei des Backends ohne Build-Tag** (minor · convention, Makefile/golangci-Konfiguration)
Datei: backend/api/druck/bondruck/application/escpos/formatter_test.go:1-2
Warum: Sie kompiliert in jedem Tag-Kontext, während die getaggte Test-Menge nie gelintet wird.
Vorschlag: `//go:build unit` als erste Zeile ergänzen, wie in den beiden Nachbardateien.
Aufwand: S · Status: unverifiziert (minor)

**13-mal wiederholtes KassenbelegData-Literal** (minor · test-quality, principles.md → DRY)
Datei: backend/api/druck/bondruck/application/escpos/formatter_test.go:177-640
Warum: Ein neues Pflichtfeld bedeutet 14 Editierstellen; die Hälfte der Datei ist Fixture-Rauschen.
Vorschlag: `kassenbelegFixture(overrides…)` einführen und die acht Ein-Zusicherungs-Tests zu einem Tabellentest verschmelzen.
Aufwand: M · Status: unverifiziert (minor)

### backend/api/druck/bondruck/application/escpos/formatter_fuzz_test.go

**Verweis auf „Befund N7" ohne Entsprechung im Repo** (minor · docs-accuracy, AGENTS.md Regel 18)
Datei: backend/api/druck/bondruck/application/escpos/formatter_fuzz_test.go:69-72
Warum: Die Kennung ist nicht auflösbar, der Kommentar erklärt damit nichts.
Vorschlag: Die geprüfte Eigenschaft selbst nennen: eine QR-Nutzlast mit GS-(-k-Präfix darf die Längenzusicherung nicht täuschen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/druck/beleg/application/kassenbeleg_command_test.go

**Helfer-Rümpfe in den Tests darüber wortgleich wiederholt** (minor · test-quality, principles.md → DRY)
Datei: backend/api/druck/beleg/application/kassenbeleg_command_test.go:451-491
Warum: Die Helfer stehen in der Dateimitte und bleiben für die 15 Tests davor unsichtbar.
Vorschlag: `belegZahlungFixture`/`belegTestCommand` verwenden und die drei Helfer an den Dateianfang zu den Mocks ziehen.
Aufwand: M · Status: unverifiziert (minor)

### backend/api/stammdaten/user/application/command.go

**Stammdaten-Update schreibt Passwortspalten ohne Sperre mit** (minor · correctness, docs/handbuch.md → Benutzer)
Datei: backend/api/stammdaten/user/application/command.go:46-77
Warum: Ein gleichzeitig committender Passwort-Vorgang wird still überschrieben, der Benutzer kommt nicht mehr hinein.
Vorschlag: Die drei Kommandos wie `SetPasswordTx` mit `FOR UPDATE` in einer Transaktion ausführen oder das UPDATE auf geänderte Spalten begrenzen.
Aufwand: M · Status: unverifiziert (minor)

### backend/api/stammdaten/user/application/errors.go

**Zwei tote exportierte Sentinels** (minor · code-smell, code-smells.md → Dead Code)
Datei: backend/api/stammdaten/user/application/errors.go:15-21
Warum: `ErrNoPassword` und `ErrNoOnetimePassword` werden hier nie erzeugt oder verglichen.
Vorschlag: Beide löschen; die produktiven Kopien liegen in api/auth/application/errors.go.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/stammdaten/user/http/query_handler.go

**Deutscher DTO-Name im dokumentierten englischen Ausnahmepaket** (minor · convention, docs/language.md)
Datei: backend/api/stammdaten/user/http/query_handler.go:20-53
Warum: `benutzer` steht neben `toUser`, `getUsersResponse` und dem JSON-Key `users`.
Vorschlag: DTO in `user` bzw. `userResponse` umbenennen, analog zu produkt/tisch/variante.
Aufwand: S · Status: unverifiziert (minor)

**Kein Test für den User-Query-Handler** (minor · test-coverage, AGENTS.md Self-Review 6)
Datei: backend/api/stammdaten/user/http/query_handler.go:55-65
Warum: Betreiber, produkt und tisch haben je einen; das DTO-Mapping bleibt hier ungeschützt.
Vorschlag: `query_handler_test.go` analog zu tisch anlegen: Erfolgsfall mit `users`-Hülle und ErrDatabase auf 500.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/stammdaten/user/http/command_handler_test.go

**Mock-Feld `passwordHash` liefert das Einmalpasswort** (minor · test-quality, readability.md → Naming)
Datei: backend/api/stammdaten/user/http/command_handler_test.go:16-45
Warum: Im Credential-Code legt der Name nahe, ein Hash gehe an den Client zurück.
Vorschlag: Feld in `onetimePassword` umbenennen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/stammdaten/betreiber/application/command.go

**ELSTER-Meldedatum aus CURRENT_DATE der DB-Sitzung** (minor · correctness, docs/handbuch.md → Tagesabschluss)
Datei: backend/api/stammdaten/betreiber/application/command.go:31-42
Warum: Die Container laufen in UTC; zwischen 00:00 und 02:00 Berliner Zeit erscheint das Vortagesdatum.
Vorschlag: Datum in Europe/Berlin bestimmen — in der Query oder als Parameter aus der Anwendungsschicht.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/stammdaten/produkt/http/query_handler.go

**Zwei strukturgleiche Response-Typen** (minor · code-smell, code-smells.md → Redundant Abstractions)
Datei: backend/api/stammdaten/produkt/http/query_handler.go:41-97
Warum: Zwei Namen für einen Vertrag laden zu divergenten Änderungen ein.
Vorschlag: Einen `produkteResponse` behalten und in beiden Handlern verwenden.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/stammdaten/produkt/http/query_handler_test.go

**GET-Requests gegen eine POST-only-API** (minor · convention, AGENTS.md Regel 1)
Datei: backend/api/stammdaten/produkt/http/query_handler_test.go:53-99
Warum: Der Test kodiert eine Anfrageform, die die API nicht akzeptiert.
Vorschlag: `http.MethodPost` verwenden und den fehlenden GetActiveProdukteHandler-Fall ergänzen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/stammdaten/tisch/http/query_handler_test.go

*_Paket-Alias `t` kollidiert mit *testing.T*_ (minor · readability, readability.md → Naming)
Datei: backend/api/stammdaten/tisch/http/query_handler_test.go:13-22
Warum: `t.Tisch` und `t.Errorf` bedeuten wenige Zeilen auseinander Verschiedenes.
Vorschlag: Import beschreibend aliasen und das DTO in query_handler.go umbenennen, damit die Kollision entfällt.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/stammdaten/tisch/application/command_test.go

**Helfer nimmt einen verworfenen Parameter entgegen** (minor · test-quality, code-smells.md → Dead Code)
Datei: backend/api/stammdaten/tisch/application/command_test.go:17-26
Warum: Der Parameter ist der einzige Grund für den domain/produkt-Import.
Vorschlag: Zweiten Parameter und Import entfernen; die einzige Aufrufstelle übergibt nil.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/helper/http.go

**MapError wählt den Code über Map-Iteration und ist ungetestet** (minor · code-smell, code-smells.md → Unnecessary Complexity)
Datei: backend/api/helper/http.go:173-181
Warum: Bei zwei passenden Sentinels entscheidet Gos zufällige Map-Reihenfolge über die Antwort.
Vorschlag: Geordnete `{error, code}`-Liste entgegennehmen oder Disjunktheit dokumentieren; Unit-Tests für Treffer, Nicht-Treffer und Wrapping ergänzen.
Aufwand: S · Status: unverifiziert (minor)

**SendConflictError ist ein reiner Weiterleitungs-Wrapper** (minor · code-smell, code-smells.md → Redundant Abstractions)
Datei: backend/api/helper/http.go:54-56
Warum: Beide Schreibweisen stehen in einem switch, der Leser muss den Helfer öffnen.
Vorschlag: An den 8 Aufrufstellen `helper.SendConflict(w, "conflict")` inlinen und den Wrapper löschen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/admin.go

**Endpunkt-Pfade und Handler-Namen mischen deutsche und englische Verben** (minor · convention, docs/language.md)
Datei: backend/api/admin.go:34-71
Warum: `/update-tisch` zeigt auf `TischAktualisierenHandler`; die verbindliche Sprachregel kennt keine Ausnahme dafür.
Vorschlag: Einmal entscheiden — Pfade und Handler auf deutsche Verben umstellen oder die Ausnahme in docs/language.md aufnehmen.
Aufwand: L · Status: unverifiziert (minor)

### backend/api/auth/http/command_handler.go

**Login unterscheidet Kontozustände entgegen der Doku** (minor · docs-accuracy, docs/handbuch.md:455)
Datei: backend/api/auth/http/command_handler.go:51-56
Warum: `user_inactive` und `no_password_set` belegen die Existenz des Kontos, das Handbuch verspricht generische Fehler.
Vorschlag: Entweder beide Zweige auf `invalid_credentials` zusammenziehen oder die Handbuch-Zeile auf den bewussten Ist-Stand korrigieren.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/auth/http/command_handler_test.go

**Login-Tests prüfen nur den Statuscode** (minor · test-quality, AGENTS.md Self-Review 6)
Datei: backend/api/auth/http/command_handler_test.go:43-92
Warum: Der Fehler-Code ist der Vertrag mit dem Frontend, eine Umbenennung bliebe grün.
Vorschlag: Body dekodieren und `code` zusichern, wie es der Throttling-Test bereits tut.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/auth/application/command_test.go

**Argon2id-Hash-Literal dreimal vor seiner Konstante** (minor · test-quality, principles.md → DRY)
Datei: backend/api/auth/application/command_test.go:27-75
Warum: Der Leser trifft den magischen String dreimal, bevor die erklärende Konstante folgt.
Vorschlag: `activeUserHash` nach oben ziehen und in den drei Tests verwenden.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/middleware/middleware_test.go

**Tests benennen und rufen eine nicht existierende Route** (minor · test-quality, readability.md → Naming)
Datei: backend/api/middleware/middleware_test.go:490-511
Warum: `/service/cancel` existiert nirgends; der echte Storno-Pfad ist /serviceleitung/stornierung-erteilen.
Vorschlag: Tests nach dem geprüften Rollen-Set benennen und den realen Pfad verwenden.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/reporting/http/query_handler.go

**anzahlZahlungen wird durch fünf Schichten getragen und nie angezeigt** (minor · code-smell, code-smells.md → Dead Code)
Datei: backend/api/reporting/http/query_handler.go:86-95
Warum: Der Live-Zweig führt das Feld nicht, wodurch auch der „dieselbe Abrechnung"-Kommentar falsch ist.
Vorschlag: Feld aus DTO, Mapper, Domain-Struct, SQL und Frontend-Typ entfernen — oder anzeigen; Kommentar in beiden Fällen korrigieren.
Aufwand: M · Status: unverifiziert (minor)

### backend/api/reporting/application/query.go

**Deutscher Verbname für eine reine Read-Side-Ableitung** (minor · convention, docs/language.md Regel 5)
Datei: backend/api/reporting/application/query.go:214
Warum: Drei direkte Nachbarn heißen compute…, aggregate… und merge…; gruppieren steht nicht auf der Ausnahmeliste.
Vorschlag: In `groupProduktStatistik` umbenennen, zwei Aufrufstellen und zwei Tests mitziehen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/reporting/application/query_export_konsistenz_test.go

**Nackte CSV-Spaltenindizes in der Prüfschleife** (minor · readability, readability.md → Clever Code)
Datei: backend/api/reporting/application/query_export_konsistenz_test.go:181-187
Warum: `record[6..9]` ist ohne den 1292-Zeilen-Mapper nicht lesbar und bricht still bei Layoutänderungen.
Vorschlag: Vier benannte Konstanten neben der Schleife einführen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/health/health_integration_test.go

**Reiner Unit-Test hinter dem Integration-Tag** (minor · test-quality, Makefile-Konvention)
Datei: backend/api/health/health_integration_test.go:1-22
Warum: Der Test braucht keine Datenbank, läuft aber nur im langsamen Docker-Harness.
Vorschlag: Auf `//go:build unit` umstellen, umbenennen und den Body plus den 200/"ok"-Fall zusichern.
Aufwand: S · Status: unverifiziert (minor)

### backend/domain/kasse/storno_aufteilung.go

**Unpassende ZahlungID lässt bereits erstattete Mengen stornierbar** (minor · correctness, docs/handbuch.md → Stornierung)
Datei: backend/domain/kasse/storno_aufteilung.go:111-123
Warum: Das Event wird still verworfen, dieselbe bezahlte Position kann ein zweites Mal erstattet werden.
Vorschlag: Treffer verfolgen und bei fehlender Zuordnung `false` liefern, wie es der default-Zweig bereits begründet.
Aufwand: S · Status: unverifiziert (minor)

**Ein bool verdeckt sechs verschiedene Ursachen** (minor · code-smell, code-smells.md → Leaky Abstraction)
Datei: backend/domain/kasse/storno_aufteilung.go:39-47
Warum: Ein korruptes Kassenjournal erscheint dem Service als gewöhnliche „nicht stornierbar"-Meldung.
Vorschlag: Mindestens alle false-Ursachen dokumentieren; sauberer ist ein error, der Journal-Korruption von Ablehnung trennt.
Aufwand: M · Status: unverifiziert (minor)

**Drei Replay-Zweige ohne Test** (minor · test-quality, AGENTS.md Self-Review 6)
Datei: backend/domain/kasse/storno_aufteilung.go:70-129
Warum: Die Funktion entscheidet über geldneutrale Korrektur gegen kassenwirksame Warenrücknahme.
Vorschlag: Drei Fälle ergänzen: Umbuchungs-Abgang, vorangegangene Warenrücknahme, unbekannter Event-Typ.
Aufwand: S · Status: unverifiziert (minor)

### backend/domain/kasse/positionen.go

**ResolvePositionen überspringt unbekannte Refs still** (minor · code-smell, code-smells.md → Leaky Abstraction)
Datei: backend/domain/kasse/positionen.go:30-55
Warum: Die Funktion ist nur korrekt, wenn zuvor ValidatePositionRefs lief — das steht nirgends.
Vorschlag: Vorbedingung im Doc-Kommentar festhalten oder die Prüfung einfalten und die drei Aufrufstellen anpassen.
Aufwand: M · Status: unverifiziert (minor)

### backend/domain/kasse/tagesabschluss_summen.go

**Summen-Switch ohne default und ohne Vollständigkeits-Gate** (minor · test-coverage, docs/compliance.md → Z-Bon)
Datei: backend/domain/kasse/tagesabschluss_summen.go:33-81
Warum: Ein künftiger summen-wirksamer Event-Typ fällt still aus den signierten Z-Bon-Summen.
Vorschlag: Test über `allEventTypes`, der jeden Typ entweder als summen-wirksam oder ausdrücklich neutral einfordert.
Aufwand: S · Status: unverifiziert (minor)

**Zeilennummern-Verweis auf reporting.sql veraltet** (minor · docs-accuracy, AGENTS.md Regel 18)
Datei: backend/domain/kasse/tagesabschluss_summen.go:18-19
Warum: „reporting.sql:10-43" beginnt auf einer Leerzeile und schneidet den kassensitzung_nr-Filter ab.
Vorschlag: Per Query-Namen referenzieren; die zwei Kopien im Test mitziehen.
Aufwand: S · Status: unverifiziert (minor)

### backend/domain/kasse/event_json_contract_test.go

**contractedTypes ist eine zweite Handliste derselben Typen** (minor · test-quality, code-smells.md → Dead Code)
Datei: backend/domain/kasse/event_json_contract_test.go:40-64
Warum: Der Guard prüft zwei handgepflegte Listen gegeneinander, nicht gegen die Konstanten oder die Tests.
Vorschlag: `contractedTypes` streichen und das Pin-Set aus etwas ableiten, das Compiler oder Testlauf erzwingt.
Aufwand: S · Status: unverifiziert (minor)

**„Format vor Phase 6" als Datierung** (minor · docs-accuracy, AGENTS.md Regel 18)
Datei: backend/domain/kasse/event_json_contract_test.go:275
Warum: Die Phase ist im Repo nicht auflösbar, der Verweis zerfällt mit den Plan-Dateien.
Vorschlag: Den Fall inhaltlich beschreiben: Event ohne benutzerKommentar parst mit Leerstring.
Aufwand: S · Status: unverifiziert (minor)

### backend/domain/kasse/offene_arbeit.go

**Exportierte Funktion mit ausschließlich paketinternem Aufrufer** (minor · code-smell, code-smells.md → Redundant Abstractions)
Datei: backend/domain/kasse/offene_arbeit.go:63-67
Warum: Exportierte Symbole gelten als unterstützte API und binden künftige Änderungen.
Vorschlag: `ComputeOffeneArbeitRollup` (und ggf. den Typ) unexportieren; die Tests liegen im selben Paket.
Aufwand: S · Status: unverifiziert (minor)

### backend/domain/kasse/replay_fuzz_test.go

**Zweite Kopie des eingefrorenen Position-Literals** (minor · test-quality, principles.md → DRY)
Datei: backend/domain/kasse/replay_fuzz_test.go:228-229
Warum: Ändert sich das gepinnte Literal, testet das Seed-Korpus still die alte Form weiter.
Vorschlag: `fuzzPositionLiteral` löschen und `positionLiteral` aus demselben Paket verwenden.
Aufwand: S · Status: unverifiziert (minor)

### backend/domain/kasse/fiskalische_projektion.go

**Zwei Switch-Idiome für dieselbe Event-Typ-Verzweigung** (minor · convention, readability.md → Naming)
Datei: backend/domain/kasse/fiskalische_projektion.go:25-26
Warum: `switch EventType(evt.Type)` und `switch evt.Type` mit `case string(...)` stehen im selben Paket nebeneinander.
Vorschlag: Auf die typisierte Form vereinheitlichen und die vier String-Switches umstellen.
Aufwand: S · Status: unverifiziert (minor)

### backend/domain/kasse/tisch_session_test.go

**Kommentar beschreibt einen überholten Zustand als aktuellen** (minor · docs-accuracy, AGENTS.md Regel 18)
Datei: backend/domain/kasse/tisch_session_test.go:404-408
Warum: Die Helfer arbeiten heute auf einer Kopie, der Kommentar behauptet Mutation der Eingabe.
Vorschlag: Aktuell formulieren: die Helfer kopieren, der Test sichert die Unveränderlichkeit des übergebenen State.
Aufwand: S · Status: unverifiziert (minor)

**Englische und deutsche Namen für dieselben Domänen-Events** (minor · convention, docs/language.md Regel 1)
Datei: backend/domain/kasse/tisch_session_test.go:13-47
Warum: `mustCreateOrderEvent`/`mustCreateCancelationEvent` stehen neben `mustCreateKorrekturEvent`, inklusive Tippfehler.
Vorschlag: Auf mustCreateBestellungEvent/…ZahlungEvent/…StornierungEvent umbenennen und die drei Testdateien anpassen.
Aufwand: S · Status: unverifiziert (minor)

**Verweise auf „Phase 4"/„Phase 6"** (minor · convention, AGENTS.md Regel 18)
Datei: backend/domain/kasse/tisch_session_test.go:520-522
Warum: Meilenstein-Labels existieren nur in gelöschten Plan-Dateien und datieren den Kommentar ohne Aussage.
Vorschlag: Sachaussage behalten („SaldoCents ist abgeleitet"), Phasenbezug streichen.
Aufwand: S · Status: unverifiziert (minor)

### backend/domain/kasse/tagesabschluss_summen_test.go

**map[string]interface{} statt any** (minor · convention, code-smells.md → Type Escape Hatches)
Datei: backend/domain/kasse/tagesabschluss_summen_test.go:341-344
Warum: Vier von sechs verbliebenen `interface{}`-Vorkommen des Backends stehen in dieser Datei.
Vorschlag: Die vier Stellen auf `map[string]any` umstellen.
Aufwand: S · Status: unverifiziert (minor)

### backend/domain/steuer/steuer.go

**Prozent() liefert ein unerreichbares -1-Sentinel** (minor · code-smell, code-smells.md → Dead Code)
Datei: backend/domain/steuer/steuer.go:31-42
Warum: Der einzige Aufrufer erreicht den default-Zweig nie, und steuerAnteil behandelt -1 wie 0.
Vorschlag: Prozent unexportieren und den toten default streichen oder im Doc festhalten, dass Kombi keinen Einzelsatz hat.
Aufwand: S · Status: unverifiziert (minor)

**Negativ-Zweig von roundHalfUpDivide unerreichbar** (minor · code-smell, code-smells.md → Dead Code)
Datei: backend/domain/steuer/steuer.go:108-118
Warum: Beide Aufrufer schließen negative Zähler vorher aus, kein Test deckt den Zweig ab.
Vorschlag: Zweig entfernen oder begründen, warum er stehen bleibt.
Aufwand: S · Status: unverifiziert (minor)

### backend/domain/produkt/product.go

**SteuersatzSchema-Alias verdeckt die Mutation fremder Paketstate** (minor · code-smell, code-smells.md → Redundant Abstractions)
Datei: backend/domain/produkt/product.go:66-72
Warum: `.Required()` mutiert das exportierte Schema von domain/steuer in place, der Alias verschleiert das.
Vorschlag: Alias löschen und an den drei Stellen `steuer.SteuersatzSchema` verwenden, wie der HTTP-Handler es tut.
Aufwand: S · Status: unverifiziert (minor)

**Validierung in NewProdukt und UpdateDetails doppelt** (minor · principle, principles.md → DRY)
Datei: backend/domain/produkt/product.go:89-134
Warum: Dieselbe Regel steht zweimal in einer Datei, dasselbe Muster in drei weiteren Aggregaten.
Vorschlag: `validateProduktFelder(...)` extrahieren und aus beiden aufrufen; analog in variant.go, user.go, tisch.go.
Aufwand: S · Status: unverifiziert (minor)

### backend/domain/produkt/variant.go

**Kommentar warnt vor einem Muster, das neun Zeilen tiefer angewandt wird** (minor · docs-accuracy, readability.md → Prose Slop)
Datei: backend/domain/produkt/variant.go:31-53
Warum: VarianteSchema ruft `.Required()` auf geteilten Schemas genau so auf, wie der Kommentar es als Footgun bezeichnet.
Vorschlag: Kommentar auf die tragende Aussage kürzen: PreisCentsSchema ist bereits Required und darf nicht erneut mutiert werden.
Aufwand: S · Status: unverifiziert (minor)

### backend/domain/tse/signaturstatus.go

**Signatur wird im erledigt-Zweig ohne nil-Prüfung dereferenziert** (minor · correctness, code-smells.md → Leaky Abstraction)
Datei: backend/domain/tse/signaturstatus.go:58-64
Warum: Die Invariante hält nur das Repository, weder Typ noch DB-Constraint erzwingen sie; ein Panic träfe Beleg und Abschluss-Gate.
Vorschlag: Zweig gegen `auftrag.Signatur != nil` absichern oder die Invariante im Doc von SignaturauftragStand festschreiben.
Aufwand: S · Status: unverifiziert (minor)

**domain/tse schreibt als einziges Domänenpaket ohne Umlaute** (minor · convention, AGENTS.md Bewertungsmetriken → Konsistenz)
Datei: backend/domain/tse/signaturstatus.go:5-18
Warum: Dieselbe Schicht schreibt dieselben Wörter zweifach, sogar innerhalb des Pakets uneinheitlich.
Vorschlag: Die deutsche Prosa der neun tse-Dateien auf echte Umlaute umstellen; Bezeichner bleiben transliteriert.
Aufwand: S · Status: unverifiziert (minor)

### backend/domain/tse/stammdaten.go

**Fünf gleichtypige String-Parameter und ein doppelter Feldsatz** (minor · readability, readability.md → Naming)
Datei: backend/domain/tse/stammdaten.go:13-37
Warum: `Stammdaten` und `TSSStammdaten` deklarieren dieselben fünf Felder, jede Vertauschung im Aufruf kompiliert.
Vorschlag: `NewStammdaten(gelesen TSSStammdaten) Stammdaten` und Felder namentlich kopieren; der doppelte Doc-Kommentar entfällt.
Aufwand: S · Status: unverifiziert (minor)

### backend/domain/reporting/reporting.go

**ServicekraftLive wiederholt sieben Felder der Abrechnung** (minor · principle, principles.md → DRY)
Datei: backend/domain/reporting/reporting.go:194-210
Warum: Der Doc-Kommentar erklärt beide als dasselbe Konzept, gepflegt werden zwei Feldlisten von Hand.
Vorschlag: `AbrechnungServicekraft` bzw. mindestens `ServicekraftRef` einbetten; DTO-Mapper als eigene Änderung planen.
Aufwand: M · Status: unverifiziert (minor)

### backend/domain/jwt/jwt.go

**sub = 0 passiert die Token-Validierung** (minor · boundary-consistency, domain/user IDSchema)
Datei: backend/domain/jwt/jwt.go:42-46
Warum: Jede andere ID-Prüfung im Repo fordert >= 1; hier stoppt erst der DB-Lookup.
Vorschlag: Guard auf `userIDFloat < 1` ändern und den Fall in die Malformed-Claims-Tabelle aufnehmen.
Aufwand: S · Status: unverifiziert (minor)

**Rückgabe (int, string, string, error) ohne benannte Ergebnisse** (minor · readability, readability.md → Naming)
Datei: backend/domain/jwt/jwt.go:31-53
Warum: Username und Rolle sind nur über die Position unterscheidbar, eine Vertauschung kompiliert.
Vorschlag: Ein `Claims`-Struct zurückgeben oder mindestens die Ergebnisse benennen.
Aufwand: S · Status: unverifiziert (minor)

### backend/domain/druckstation/druckstation.go

**Drei Switches über dieselben fünf Kategorien, zwei mit stillem Fallback** (minor · code-smell, principles.md → Open/Closed)
Datei: backend/domain/druckstation/druckstation.go:25-81
Warum: Eine sechste Station erfordert drei Änderungen; der Compiler meldet keine davon, und `isValid` ist zudem unexportiert.
Vorschlag: Eine Tabelle `kategorien` einführen, aus der isValid, Anzeigename und HatBonmodus lesen, und sie der HTTP-Schicht exportieren.
Aufwand: S · Status: unverifiziert (minor)

### backend/domain/tisch/tisch.go

**Kein einziger Test im Paket** (minor · test-quality, AGENTS.md Self-Review 6)
Datei: backend/domain/tisch/tisch.go:70-107
Warum: Das Schwesterpaket domain/user testet die baugleichen Mutatoren vollständig; domain/betreiber fehlt ebenfalls.
Vorschlag: `tisch_test.go` (`//go:build unit`) analog zu user_test.go: Namensvalidierung plus Status und UpdatedAt.
Aufwand: S · Status: unverifiziert (minor)

### backend/repository/kassensitzungen_repo/repo.go

**GetOffeneKassensitzung mappt die Zeile inline** (minor · principle, principles.md → DRY)
Datei: backend/repository/kassensitzungen_repo/repo.go:66-73
Warum: `kassensitzungRowToDomain` existiert für genau diesen Row-Typ und wird von den Nachbarn benutzt.
Vorschlag: `ks := kassensitzungRowToDomain(row); return &ks, nil`, wie GetAktiveKassensitzung.
Aufwand: S · Status: unverifiziert (minor)

### backend/repository/kassensitzungen_repo/repo_test.go

**DB-Cleanup-Sequenz in sieben Testpaketen kopiert** (minor · code-smell, principles.md → DRY)
Datei: backend/repository/kassensitzungen_repo/repo_test.go:16-38
Warum: Die Löschreihenfolge kodiert FK- und Trigger-Wissen; eine neue Tabelle bricht die vergessenen Kopien.
Vorschlag: Einen Helfer neben `OpenTestDatabase` bereitstellen und aus allen Setups aufrufen.
Aufwand: M · Status: unverifiziert (minor)

**Fixture kodiert Subject-Nummern, die die DB nie vergibt** (minor · test-quality, docs/language.md → Subject-Format)
Datei: backend/repository/kassensitzungen_repo/repo_test.go:96-109
Warum: Die Query joint über kassensitzung_nr, die hartkodierten Subjects sind dekorativ und irreführend.
Vorschlag: Subjects aus der zurückgegebenen z_nr bauen (`kasse.KassensitzungSubject(ks10)`) und die Variablen sprechend benennen.
Aufwand: S · Status: unverifiziert (minor)

### backend/repository/reporting_repo/repo.go

**Einziges Repository ohne db.Error-Normalisierung** (minor · convention, architecture.md → Repository Pattern)
Datei: backend/repository/reporting_repo/repo.go:84-96
Warum: `errors.Is(err, db.ErrNotFound)` kann für dieses Repo nie greifen, das Paket importiert backend/db nicht einmal.
Vorschlag: Alle Rückgaben mit `db.Error(err)` umhüllen, wie in den sechs Schwester-Repositories.
Aufwand: S · Status: unverifiziert (minor)

### backend/repository/reporting_repo/repo_test.go

**Hartkodierte Event-Subjects mit Sitzungsnummer 1** (minor · convention, docs/language.md → Subject-Format)
Datei: backend/repository/reporting_repo/repo_test.go:255-272
Warum: Die tatsächliche Nummer stammt aus der Sequenz; die ~35 Literale beschreiben unmögliche Zeilen.
Vorschlag: `kasse.TischSessionSubject`/`KassensitzungSubject`/`DirektverkaufSubject` verwenden, wie summen_abschluss_test.go.
Aufwand: S · Status: unverifiziert (minor)

### backend/repository/tse_repo/repo.go

**Zweite Störungsursache wird nie protokolliert** (minor · ops, AEAO zu § 146a, 1.14.1)
Datei: backend/repository/tse_repo/repo.go:207-222
Warum: Bei offenem 'rueckstand'-Zeitraum ist das Öffnen ein No-Op, echte TSE-Fehler und Fehlertext gehen verloren.
Vorschlag: Bei No-Op die konkrete Ursache auf dem aktiven Zeitraum nachtragen oder den Unique-Index additiv je grund_art führen.
Aufwand: M · Status: unverifiziert (minor)

### backend/repository/tse_repo/repo_test.go

**Kommentare beschreiben entfernte Artefakte** (minor · convention, AGENTS.md Regel 18)
Datei: backend/repository/tse_repo/repo_test.go:221-377
Warum: „die Admin-Lese-Query gibt es nicht mehr" und „kein Head-of-Line-Blocking mehr" zwingen zum Nachdenken über toten Code.
Vorschlag: Nur den Ist-Stand nennen: liest direkt aus der Tabelle; ein neuerer Auftrag bleibt abholbar.
Aufwand: S · Status: unverifiziert (minor)

### backend/repository/tse_repo/einrichtung_test.go

**Zweites Fixture-Set für dasselbe Schema im selben Paket** (minor · test-quality, principles.md → DRY)
Datei: backend/repository/tse_repo/einrichtung_test.go:16-97
Warum: Struct, Reset-Liste und Insert-Helfer duplizieren repo_test.go bis auf eine Zeile.
Vorschlag: Das vorhandene Fixture um das tse_konfiguration-Reset erweitern und die Kopien löschen.
Aufwand: M · Status: unverifiziert (minor)

### backend/repository/tse_repo/fiskaly_client_test.go

**Auth-Stub und Client-Konstruktion neunmal wortgleich** (minor · test-quality, principles.md → DRY)
Datei: backend/repository/tse_repo/fiskaly_client_test.go:40-124
Warum: Rund 150 Zeilen tragen keine fallspezifische Information, sechs weitere Kopien liegen im Setup-Test.
Vorschlag: `writeAuthResponse(w, env)` und `newTestClient(t, handler)` einführen und nur die abweichenden Zweige belassen.
Aufwand: M · Status: unverifiziert (minor)

**Retry-After 0 erzeugt echte 600 ms Schlaf** (minor · test-quality, code-smells.md → Defensive Defaults)
Datei: backend/repository/tse_repo/fiskaly_client_test.go:239-248
Warum: parseRetryAfter akzeptiert nur Werte > 0, der Test fällt auf den exponentiellen Backoff zurück.
Vorschlag: Header weglassen und die vorhandene sleepFn injizieren; den Retry-After-Pfad separat mit positivem Wert prüfen.
Aufwand: S · Status: unverifiziert (minor)

### backend/repository/tse_repo/fiskaly_client.go

**doJSONRequest mit sieben positionellen Parametern** (minor · readability, code-smells.md → Long Parameter List)
Datei: backend/repository/tse_repo/fiskaly_client.go:439-447
Warum: Aufrufe lesen sich als `nil, nil, true, &resp`; ein vertauschter Body oder fehlendes Auth kompiliert.
Vorschlag: In doJSONRequest/doAuthRequest aufteilen oder ein kleines Request-Struct übergeben.
Aufwand: M · Status: unverifiziert (minor)

### backend/repository/tse_repo/fiskaly_setup.go

**Derselbe tssID-Guard in acht Methoden** (minor · principle, principles.md → DRY)
Datei: backend/repository/tse_repo/fiskaly_setup.go:118-288
Warum: Eine Änderung an Prüfung oder Meldung muss achtmal erfolgen.
Vorschlag: Einen Helfer `tssPath(tssID, suffix...) (string, error)` einführen, der trimmt, prüft und den Pfad baut.
Aufwand: S · Status: unverifiziert (minor)

### backend/repository/tse_repo/fiskaly_setup_live_test.go

**Warnung verweist auf eine nicht existierende Plan-Passage** (minor · docs-accuracy, AGENTS.md Git-Workflow)
Datei: backend/repository/tse_repo/fiskaly_setup_live_test.go:21-24
Warum: Ausgerechnet die Warnung vor einer irreversiblen TSS-Anlage zeigt ins Leere.
Vorschlag: Klammerverweis streichen und die Einschränkung inline nennen.
Aufwand: S · Status: unverifiziert (minor)

### backend/repository/kassenjournal_repo/mock.go

**Neue Event-IDs aus len(events)+1 überschreiben seeded Events** (minor · test-quality, test-quality → fixture integrity)
Datei: backend/repository/kassenjournal_repo/mock.go:97-99
Warum: Bei nicht lückenlosen IDs geht ein Event ohne Fehlersignal verloren.
Vorschlag: Nächste ID als max(vorhandene)+1 in einem Helfer berechnen und aus allen drei Schreibpfaden nutzen.
Aufwand: S · Status: unverifiziert (minor)

### backend/repository/produkt_repo/mock.go

**Mock liefert bei „nicht gefunden" nil statt db.ErrNotFound** (minor · test-quality, test-quality → contract fidelity)
Datei: backend/repository/produkt_repo/mock.go:41-47
Warum: Unit-Tests laufen auf dem Erfolgspfad mit Null-Entitäten, die die Produktion nie erzeugt.
Vorschlag: `db.ErrNotFound` zurückgeben; dasselbe in produkt (Variante), tisch und user.
Aufwand: S · Status: unverifiziert (minor)

**GetActiveProdukte-Mock filtert anders als die Query** (minor · test-quality, test-quality → contract fidelity)
Datei: backend/repository/produkt_repo/mock.go:113-121
Warum: Der Mock liefert Produkte ohne aktive Variante, die der INNER JOIN bewusst ausblendet.
Vorschlag: Produkte ohne aktive Variante überspringen und nur aktive Varianten zurückgeben.
Aufwand: S · Status: unverifiziert (minor)

### backend/repository/produkt_repo/repo_test.go

**Setup-Helfer ohne t.Helper()** (minor · test-quality, Go-Testkonvention)
Datei: backend/repository/produkt_repo/repo_test.go:17-42
Warum: Fehler werden an der Helferzeile statt am Test gemeldet; druckstation_repo macht es richtig.
Vorschlag: `t.Helper()` als erste Anweisung ergänzen, ebenso in tisch_repo und user_repo.
Aufwand: S · Status: unverifiziert (minor)

**Arrange-Fehler werden verworfen** (minor · test-quality, principles.md → Fail Fast)
Datei: backend/repository/produkt_repo/repo_test.go:81-82
Warum: Ein fehlgeschlagenes Fixture äußert sich als irreführende Zusicherung statt als klare Ursache.
Vorschlag: Fixture-Aufrufe in Helfer kapseln, die `t.Fatalf` rufen, wie erstelleTisch/createUser es tun.
Aufwand: M · Status: unverifiziert (minor)

### backend/repository/tisch_repo/repo_test.go

**itoaLast baut eine Ziffernkonvertierung von Hand** (minor · readability, readability.md → Clever Code)
Datei: backend/repository/tisch_repo/repo_test.go:391-395
Warum: Die zugesagte Eindeutigkeit gilt nur, solange die Tisch-IDs sich in der letzten Ziffer unterscheiden.
Vorschlag: `strconv.Itoa(tischID)` verwenden und den Helfer löschen.
Aufwand: S · Status: unverifiziert (minor)

### backend/repository/user_repo/repo_test.go

**Toter Fehlerrückgabewert und Paket-Shadowing** (minor · test-quality, code-smells.md → Dead Code)
Datei: backend/repository/user_repo/repo_test.go:16-79
Warum: `createTestUser` kann nur nil liefern, und die Variable `user` verdeckt das gleichnamige Domänenpaket.
Vorschlag: Nur `user.User` zurückgeben, die tote Prüfung entfernen und die Variable `u` bzw. `seeded` nennen.
Aufwand: S · Status: unverifiziert (minor)

### backend/repository/favorit_repo/mock.go

**Mock-Methode ohne produktives Gegenstück** (minor · code-smell, test-quality → contract fidelity)
Datei: backend/repository/favorit_repo/mock.go:45-52
Warum: `RemoveByTisch` existiert im echten Repository nicht; die Produktion räumt in der tisch_repo-Transaktion auf.
Vorschlag: Methode streichen oder klar als Test-Seam benennen und die Cleanup-Closure an der Aufrufstelle bauen.
Aufwand: S · Status: unverifiziert (minor)

### backend/app/app.go

**Shutdown verschluckt den Fehler und wartet nicht auf die Worker** (minor · ops, code-smells.md → Dead Code)
Datei: backend/app/app.go:104-121
Warum: Ein am 30-s-Limit gescheiterter Shutdown meldet Erfolg, und `defer db.Close()` kann laufende TSE-Ticks abschneiden.
Vorschlag: Den Server-Shutdown-Fehler zurückgeben und Worker sowie Watchdog über eine WaitGroup joinen, bevor Run zurückkehrt.
Aufwand: M · Status: unverifiziert (minor)

### backend/app/routes.go

**Drei Kommentare beschreiben den vorherigen Codestand** (minor · convention, AGENTS.md Regel 18)
Datei: backend/app/routes.go:32-49
Warum: „bisherige", „bislang" und „früheren imperativen Registrierung" sagen nichts über das Jetzt.
Vorschlag: Als Ist-Aussagen formulieren bzw. den letzten Satz streichen.
Aufwand: S · Status: unverifiziert (minor)

**Area.Name wird gesetzt und nie gelesen** (minor · code-smell, code-smells.md → Dead Code)
Datei: backend/app/routes.go:19-20
Warum: Der Kommentar behauptet Nutzung „für Logs und Tests"; beides trifft nicht zu.
Vorschlag: Feld und die sechs Zuweisungen löschen oder es in `mountArea` tatsächlich loggen.
Aufwand: S · Status: unverifiziert (minor)

### backend/config/config.go

**Unerreichbarer log.Fatalf-Zweig in parseEnvString** (minor · code-smell, code-smells.md → Dead Code)
Datei: backend/config/config.go:106-117
Warum: Alle fünf Aufrufstellen übergeben einen nicht leeren Default.
Vorschlag: Auf `if v := os.Getenv(name); v != "" { return v }; return defaultValue` reduzieren.
Aufwand: S · Status: unverifiziert (minor)

**Einziges Paket mit stdlib-log und os.Stderr statt zerolog** (minor · convention, principles.md → Least Surprise)
Datei: backend/config/config.go:3-134
Warum: Startdiagnosen landen in anderem Format und Stream, und zwei Fehlerpolitiken widersprechen sich in einer Datei.
Vorschlag: Auf zerolog umstellen (main.go konfiguriert es vor config.Load) und eine Politik für ungültige Werte wählen.
Aufwand: S · Status: unverifiziert (minor)

**placeholderSecrets-Kommentar begründet über einen entfernten Default** (minor · convention, AGENTS.md Regel 18)
Datei: backend/config/config.go:33-36
Warum: Der Kommentar dokumentiert einen früheren Codestand statt der heutigen Ablehnungsregel.
Vorschlag: „öffentlich bekannte Werte aus .env.example sowie der triviale Wert `admin`".
Aufwand: S · Status: unverifiziert (minor)

### backend/config/config_test.go

**os.Clearenv/os.Setenv ohne Wiederherstellung** (minor · test-quality, Go-Testkonvention)
Datei: backend/config/config_test.go:18-148
Warum: Die Umgebung des Testbinaries wird für alle folgenden Tests gelöscht; app_test.go nutzt bereits t.Setenv.
Vorschlag: Auf `t.Setenv` umstellen und statt Clearenv gezielt die benötigten Variablen leeren.
Aufwand: S · Status: unverifiziert (minor)

**„has no default anymore" und „old admin default"** (minor · convention, AGENTS.md Regel 18)
Datei: backend/config/config_test.go:50-178
Warum: Kommentar und Testfallname verweisen auf ein entferntes Feature.
Vorschlag: Auf Ist-Aussagen umschreiben, Testfall in „trivial value POSTGRES_PASSWORD" umbenennen.
Aufwand: S · Status: unverifiziert (minor)

### backend/db/testing.go

**Test-Helfer ohne Build-Tag im Produktionspaket** (minor · architecture, architecture.md → Separation from Frameworks)
Datei: backend/db/testing.go:1-38
Warum: `OpenTestDatabase` mit `os.Exit(1)` und dem Fallback-Passwort „admin" wird ins ausgelieferte Binary kompiliert.
Vorschlag: `//go:build integration` ergänzen — alle 22 Aufrufer tragen den Tag bereits — oder in ein dbtest-Paket verschieben.
Aufwand: S · Status: unverifiziert (minor)

### backend/main.go

**DSN wird unescaped zusammengesetzt** (minor · correctness, libpq Keyword/Value-Format)
Datei: backend/main.go:36-38
Warum: Ein Passwort mit Leerzeichen oder Anführungszeichen erzeugt einen fehlerhaften Connection-String.
Vorschlag: Werte nach libpq-Regeln quoten oder den DSN als URL mit `url.QueryEscape` bauen.
Aufwand: S · Status: unverifiziert (minor)

**DSN-Format an zwei Orten** (minor · principle, principles.md → DRY)
Datei: backend/main.go:36
Warum: Produktion und Test-Helfer beschreiben dieselbe Verbindungspolitik, unterscheiden sich aber bereits im Port-Typ.
Vorschlag: Einen exportierten DSN-Builder in package db bereitstellen und aus beiden aufrufen.
Aufwand: S · Status: unverifiziert (minor)

### backend/app/app_test.go

**Datenbank-Handle als `&sql.DB{}` erfunden** (minor · test-quality, code-smells.md → Type Escape Hatches)
Datei: backend/app/app_test.go:25-48
Warum: Ein Null-Handle paniert bei jeder Nutzung, und eine Datei nutzt drei Erzeugungsstile.
Vorschlag: Einen Helfer mit `sql.Open("pgx", "invalid-connection-string")` einführen und die fünf Literale ersetzen.
Aufwand: S · Status: unverifiziert (minor)

**Test bindet einen realen TCP-Port und schläft** (minor · test-quality, code-smells.md → Environment-dependent test)
Datei: backend/app/app_test.go:170-193
Warum: Der Test scheitert, sobald Port 3000 belegt ist, und die 100 ms sind eine Wette gegen den TSE-Worker-Tick.
Vorschlag: Port 0 konfigurieren oder einen Listener injizieren und statt des Sleeps auf ein Bereitschaftssignal warten.
Aufwand: S · Status: unverifiziert (minor)

**Fünf Kommentare wiederholen die Codezeile darunter** (minor · readability, readability.md → Unnecessary Comments)
Datei: backend/app/app_test.go:175-191
Warum: Sie erhöhen das Lesevolumen ohne Information und setzen einen Stil, dem das Paket sonst nicht folgt.
Vorschlag: Alle fünf löschen; höchstens den Grund für das Warten erklären.
Aufwand: S · Status: unverifiziert (minor)

### backend/seed/seed_integration_test.go

**cleanSeedDB byte-identisch in zwei Paketen** (minor · principle, principles.md → DRY)
Datei: backend/seed/seed_integration_test.go:49-79
Warum: Tabellenliste, FK-Reihenfolge und Trigger-Dance sind ein Wissen an zwei Orten.
Vorschlag: Cleanup einmal als exportierten Test-Helfer im seed-Paket bereitstellen und aus beiden Tests aufrufen.
Aufwand: S · Status: unverifiziert (minor)

**Ein 230-Zeilen-Test mit rund zwanzig Zusicherungen** (minor · test-quality, principles.md → Single Responsibility)
Datei: backend/seed/seed_integration_test.go:81-315
Warum: Das erste t.Fatalf verdeckt alle weiteren Prüfungen, der Name beschreibt die Hälfte davon nicht.
Vorschlag: Auf t.Run-Subtests über ein gemeinsames Fixture aufteilen, ohne Zusicherungen zu ändern.
Aufwand: M · Status: unverifiziert (minor)

### backend/seed/writer.go

**Projektionen werden erst nach dem Commit neu gebaut** (minor · ops, docs/handbuch.md → Projektion)
Datei: backend/seed/writer.go:61-68
Warum: Ein Fehler hinterlässt committete Events ohne Projektion, und der Guard verhindert danach den Retry.
Vorschlag: Rebuild in dieselbe Transaktion ziehen oder mindestens `jotti rebuild-projections` in der Fehlermeldung nennen.
Aufwand: M · Status: unverifiziert (minor)

**writeSeed mit neun Parametern** (minor · code-smell, code-smells.md → Long Parameter List)
Datei: backend/seed/writer.go:73
Warum: Sechs davon sind Teile eines bereits berechneten Ergebnisses.
Vorschlag: Die Nutzlast in ein `seedSchreibsatz`-Struct bündeln und `(ctx, database, satz, reset)` übergeben.
Aufwand: S · Status: unverifiziert (minor)

### backend/seed/faketse_test.go

**Deutsche Testnamen für englisch benannte Funktionen** (minor · convention, docs/language.md Regel 5)
Datei: backend/seed/faketse_test.go:75-173
Warum: Wer nach `buildSignaturauftraege` greppt, findet die Tests nicht.
Vorschlag: In `TestBuildSignaturauftraege_*` und `TestBuildDruckauftraege_*` umbenennen.
Aufwand: S · Status: unverifiziert (minor)

### backend/seed/bondruck.go

**„im neuen Modell" im Doc-Kommentar** (minor · convention, AGENTS.md Regel 18)
Datei: backend/seed/bondruck.go:80
Warum: „neu" hat nur Bedeutung gegenüber einem alten Modell, das der Leser nicht kennt.
Vorschlag: Qualifikator streichen.
Aufwand: S · Status: unverifiziert (minor)

### backend/sqlc/queries/produkte.sql

**Varianten-JSON-Projektion dreimal ausgeschrieben** (minor · principle, principles.md → DRY)
Datei: backend/sqlc/queries/produkte.sql:10-96
Warum: Die camelCase-Keys sind der Vertrag mit produkt_repo; ein fehlendes Feld zeigt sich erst zur Laufzeit.
Vorschlag: Aggregation in eine View oder eine geteilte CTE ziehen; erfordert eine additive Migration.
Aufwand: M · Status: unverifiziert (minor)

### backend/sqlc/queries/reporting.sql

**Query-Doku unterhalb der `-- name:`-Zeile** (minor · convention, sqlc-Konvention)
Datei: backend/sqlc/queries/reporting.sql:1-12
Warum: Genau die ausführlich dokumentierten Queries verlieren ihre Doku an der Go-Oberfläche.
Vorschlag: Kommentarblöcke über die `-- name:`-Zeile ziehen und `make sqlc` laufen lassen.
Aufwand: M · Status: unverifiziert (minor)

### backend/dsfinvkpruefung/csv.go

**Zeilentrennung ignoriert das Text-Einschlusszeichen** (minor · correctness, DSFinV-K 2.4 → CSV-Format)
Datei: backend/dsfinvkpruefung/csv.go:314-351
Warum: Ein Zeilenumbruch in BON_NOTIZ zerreißt den Datensatz, der Prüfer meldet Falschbefunde und überspringt die Zeilen.
Vorschlag: Datensätze quote-bewusst trennen (inQuotes-Zustand) oder Zeilenumbrüche im Exporter aus Freitextfeldern entfernen.
Aufwand: M · Status: unverifiziert (minor)

### backend/Dockerfile

**Image kompiliert ./main.go statt des Pakets** (minor · ops, Makefile-Konvention)
Datei: backend/Dockerfile:14
Warum: Eine zweite Datei in package main hält `make build` grün und bricht erst den Image-Build.
Vorschlag: `go build -ldflags "…" -o /go/bin/jotti .` verwenden, ebenso `go run .` in docker-compose.yml.
Aufwand: S · Status: unverifiziert (minor)

## Frontend

Umfang: React/TypeScript-App unter `frontend/` — Service-Bereich, Admin-Bereich, geteilte Bausteine, Test- und Build-Konfiguration.

0 Blocker · 25 Major · 109 Minor (134 Befunde, nach Zusammenführung 94 Einträge).

### frontend/package.json

**[Lint-Gate kann bei autofixbaren Regeln nicht rot werden]** (ops, AGENTS.md Regel 15)
Datei: frontend/package.json:14
Warum: `lint` ruft `eslint --fix`; behobene Verstöße werden nicht gemeldet, CI verwirft den Fix.
Vorschlag: `"lint": "eslint --max-warnings=0 ."` als Gate, `"lint:fix"` separat für lokal (Makefile `fmt-frontend`).
Aufwand: S · Status: bestätigt

**[Sieben Runtime-Dependencies ohne jeden Import]** (ops, AGENTS.md Regel 16)
Datei: frontend/package.json:21-43
Warum: `@base-ui/react`, `cmdk`, `date-fns`, `embla-carousel-react`, `react-day-picker`, `react-resizable-panels`, `recharts` werden nirgends importiert.
Vorschlag: Die sieben Einträge entfernen und `pnpm-lock.yaml` neu erzeugen.
Aufwand: S · Status: bestätigt

### frontend/src/components/ui/sonner.tsx

**[Toaster liest das Theme aus einem nicht gemounteten Provider]** (correctness, cleanup/code-smells.md → Redundant Abstractions)
Datei: frontend/src/components/ui/sonner.tsx:1-8
Warum: `useTheme` aus `next-themes` liefert ohne Provider kein Theme; der Default `system` gewinnt immer.
Vorschlag: `useTheme` aus `@/components/theme-provider` importieren und `theme={isDark ? 'dark' : 'light'}` setzen; `next-themes` entfällt.
Aufwand: S · Status: bestätigt

### frontend/src/components/ui/dialog.tsx

**[Englische Screenreader-Texte im deutschen UI]** (convention, AGENTS.md Regel 6 / docs/language.md §3)
Datei: frontend/src/components/ui/dialog.tsx:93 (ebenso sheet.tsx:90, sidebar.tsx:195-197, spinner.tsx:6)
Warum: Jeder Dialog und der Login-Spinner geben „Close“ bzw. „Loading“ an Screenreader aus.
Vorschlag: „Schließen“ (dialog.tsx:93, sheet.tsx:90), „Seitenleiste“/„Zeigt die mobile Seitenleiste.“ (sidebar.tsx:196-197), „Wird geladen“ (spinner.tsx:6).
Aufwand: S · Status: bestätigt

### frontend/src/test/input-otp.ts

**[Kommentar behauptet falsche Paketversion, Workaround ist tot]** (docs-accuracy, AGENTS.md Regel 18)
Datei: frontend/src/test/input-otp.ts:1-16
Warum: Installiert ist 1.5.0 mit Timer-Cleanup; „1.4.2 ist die letzte Version (Stand 2026-07)“ ist falsch.
Vorschlag: `src/test/input-otp.ts` löschen und die `afterEach`-Aufrufe in OTPField.test.tsx:28 und PasswordForm.test.tsx:32 entfernen.
Aufwand: S · Status: bestätigt

### frontend/src/service/table/hooks.ts

**[Query-Fehler werden als echte Leer-Daten angezeigt]** (correctness)
Datei: frontend/src/service/table/hooks.ts:56-90
Warum: Die Hooks verwerfen `isError` und liefern Defaults; die Servicekraft liest „Kassiert 0 · 0,00 €“ als Tatsache.
Vorschlag: `isError` mitgeben und in TableSelectionPage `LadefehlerAlert` rendern, wie TablePage.tsx:164-173; gleich für direktverkauf/hooks.ts:9-19 und product/hooks.ts:9-15.
Aufwand: S · Status: bestätigt

**[`refetch` wird von keinem Aufrufer genutzt]** (code-smell, cleanup/code-smells.md → Dead Code)
Datei: frontend/src/service/table/hooks.ts:10-20
Warum: `useAktiveTische` gibt ein `refetch` zurück, das kein Konsument liest.
Vorschlag: `refetch` aus Destructuring und Rückgabe streichen.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/service/components/table/BestellungAbschluss.tsx

**[Idempotenz-Schlüssel überlebt geänderten Warenkorb]** (correctness)
Datei: frontend/src/service/components/table/BestellungAbschluss.tsx:50-81
Warum: `bestellungId` rotiert nur bei leer→nicht-leer; ein Retry mit geänderten Positionen wird idempotent verworfen.
Vorschlag: In `onSubmit` Payload (`positionen` + `kommentar`) gegen den letzten Versuch vergleichen und bei Abweichung neue UUID erzeugen.
Aufwand: S · Status: bestätigt

**[Container-Wrapper dreifach kopiert]** (principle, cleanup/principles.md → DRY)
Datei: frontend/src/service/components/table/BestellungAbschluss.tsx:129-143
Warum: Der `sheet`/`aside`-Wrapper ist byte-identisch in ZahlungAbschluss.tsx:197-211 und DirektverkaufAbschluss.tsx:194-208.
Vorschlag: `AbschlussContainer({ variant, pending, children })` neben AbschlussHeader extrahieren und dreimal verwenden.
Aufwand: S · Status: bestätigt

### frontend/src/service/components/table/ZahlungAbschluss.tsx

**[Barzahlungs-Eingabeblock 48 Zeilen dupliziert]** (principle, cleanup/principles.md → DRY)
Datei: frontend/src/service/components/table/ZahlungAbschluss.tsx:123-170
Warum: Der Block ist bis auf eine Klassen-Reihenfolge identisch mit DirektverkaufAbschluss.tsx:123-170.
Vorschlag: `BarzahlungFelder`-Komponente neben AufrundenChips extrahieren und aus beiden Abschluss-Komponenten rendern.
Aufwand: M · Status: bestätigt

### frontend/src/service/components/ServiceDock.tsx

**[Kommentare erzählen Entwicklungshistorie statt Ist-Zustand]** (convention, AGENTS.md Regel 18)
Datei: frontend/src/service/components/ServiceDock.tsx:4-7 (ebenso TischAuswahlDrawer.tsx:27,53-54; Direktverkauf.tsx:17,72; Bestellung.tsx:22,77; Zahlung.tsx:40,240)
Warum: „früher“, „jetzt“, „nicht mehr“, „in Phase 3“, „seit A1“ beschreiben einen Stand, den es nicht mehr gibt.
Vorschlag: Historien-Klauseln streichen, nur den geltenden Satz behalten; gleiches in den genannten Testkommentaren.
Aufwand: M · Status: bestätigt

### frontend/src/service/product/Produkt.ts

**[Zweite, bereits abgedriftete Produkt-Schema-Definition]** (principle, cleanup/principles.md → DRY)
Datei: frontend/src/service/product/Produkt.ts:24-55
Warum: Dieselbe Backend-Entität hat zwei zod-Definitionen; der Service-Kopie fehlt `steuersatz`, die Preisgrenzen weichen ab.
Vorschlag: Gemeinsames Response-Schema (Produkt/Variante/Kategorie) in ein Modul ziehen, das beide Bereiche importieren; Formular-Regeln bleiben in admin.
Aufwand: M · Status: bestätigt

**[Eingabe-Regeln mit UI-Meldungen in Response-Schemas]** (boundary-consistency)
Datei: frontend/src/service/product/Produkt.ts:26-33
Warum: Unerreichbare Formularmeldungen; die Grenzen widersprechen dem Backend-Vertrag (`min(1)`/`max(99999)`).
Vorschlag: Response-Schemas auf den Transport-Vertrag reduzieren (`z.string().min(1)`, `z.number().int()`), Meldungen nur im Admin-Formular.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/admin/reporting/ReportingResults.tsx

**[Kassensturz-Differenz im Tagesbericht mit umgekehrtem Vorzeichen]** (correctness)
Datei: frontend/src/admin/reporting/ReportingResults.tsx:35-39
Warum: Der Bericht druckt den Event-Rohwert (Soll − Ist), der Abschluss-Bildschirm zeigt Ist − Soll — ein Fehlbetrag liest sich als Überschuss.
Vorschlag: `formatEuroMitVorzeichen(-metadaten.kassensturzDifferenzCents)` rendern oder das Label eindeutig auf „(Soll − Ist)“ setzen; Test:129 mitziehen.
Aufwand: S · Status: bestätigt

**[Tote `loading`-Prop mit unerreichbarem Spinner-Zweig]** (code-smell, cleanup/code-smells.md → Dead Code)
Datei: frontend/src/admin/reporting/ReportingResults.tsx:70-88
Warum: Der einzige Aufrufer übergibt immer `loading={false}` und blendet den Bericht vorher selbst aus.
Vorschlag: Prop und `if (loading)`-Zweig entfernen, die zehn Render-Aufrufe im Test anpassen.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/admin/kasse/GeldtransitDialog.tsx

**[Idempotenz-Schlüssel überlebt das Schließen des Dialogs]** (correctness)
Datei: frontend/src/admin/kasse/GeldtransitDialog.tsx:50-70
Warum: `geldtransitId` rotiert nur nach Erfolg; der Dialog bleibt gemountet, jede Folgebuchung nutzt denselben Schlüssel.
Vorschlag: Die ID im Öffnen-Effekt (Zeile 68-70) zusammen mit `form.reset` erneuern; Retries innerhalb einer Öffnung behalten sie.
Aufwand: S · Status: bestätigt

### frontend/src/admin/finanzamt/LaeuftAllesSection.tsx

**[Signatur-Panel verschweigt fehlgeschlagene TSE-Signaturen]** (correctness, § 146a AO)
Datei: frontend/src/admin/finanzamt/LaeuftAllesSection.tsx:65-110
Warum: Der Klartext hängt allein an `offene === 0`, meldet sonst „normal bei vollem Betrieb“ und zeigt `letzterFehler` nie.
Vorschlag: Fehlgeschlagene zuerst melden, unabhängig von `offene`; `letzterFehler` und einen Fehler-Zähler in die Kennzahlen aufnehmen; Beruhigung nur unterhalb `RUECKSTAND_WARN_SEKUNDEN`.
Aufwand: S · Status: bestätigt

### frontend/src/admin/reporting/UebersichtStatusZeile.tsx

**[Kommentare beschreiben abgelöste Komponenten]** (convention, AGENTS.md Regel 18)
Datei: frontend/src/admin/reporting/UebersichtStatusZeile.tsx:8-10 (ebenso GeldtransitDialog.tsx:37-38, LaeuftAllesSection.tsx:44, SitzungsListe.tsx:11-12)
Warum: „Ersetzt die früheren …“ und „nicht mehr … Unterschied zur früheren …“ verweisen auf nicht mehr existierende Artefakte.
Vorschlag: Historien-Klauseln streichen, nur den Ist-Zustand beschreiben.
Aufwand: S · Status: bestätigt

### frontend/src/admin/tse/TSEEinrichtungWizard.test.tsx

**[LIVE-TSE-Anlage ohne Test der Sperren]** (test-quality)
Datei: frontend/src/admin/tse/TSEEinrichtungWizard.test.tsx:77-167
Warum: Einziger describe-Block prüft das Vorgangs-Register; die unumkehrbare LIVE-Anlage hat keinen Test der Tippbestätigung.
Vorschlag: Tests ergänzen für Button-Sperre bis `tippBestaetigung === 'LIVE'`, `tse_setup_pin_unbekannt`, `istEinsatzbereit` ohne PIN, `nurDisabledOderLeer`.
Aufwand: M · Status: bestätigt

### frontend/src/admin/users/UserRow.tsx

**[Admin kann sich selbst deaktivieren und aussperren]** (security)
Datei: frontend/src/admin/users/UserRow.tsx:69-84
Warum: Der Status-Switch hat keinen `isSelf`-Guard; der Refetch danach liefert 401 und erzwingt Logout.
Vorschlag: Switch in der eigenen Zeile sperren, analog zu Zeile 117, und im Backend `cannot_deactivate_self` in `DeactivateUserHandler` ergänzen.
Aufwand: S · Status: bestätigt

### frontend/src/admin/components/AdminPageHeader.tsx

**[Rule-18-Verstöße im Admin-Stammdaten-Bereich]** (docs-accuracy, AGENTS.md Regel 18)
Datei: frontend/src/admin/components/AdminPageHeader.tsx:5-10 (ebenso users/UserRolle.tsx:5-10, users/Users.tsx:23, components/HinweisKarte.tsx:7, components/WarnKarte.tsx:7-8)
Warum: Kommentare erzählen, was ersetzt wurde („Ersetzt die früheren losen H1“, „kehrt die frühere Drei-Varianten-Absicht um“, „Ab Phase 3“).
Vorschlag: Historien- und Phasen-Klauseln streichen, die geltende Regel als Aussage stehen lassen.
Aufwand: S · Status: bestätigt

### frontend/src/service/components/table/TischHistorie.tsx

**[Unerreichbarer Fallback in `Details`]** (code-smell, cleanup/code-smells.md → Dead Code)
Datei: frontend/src/service/components/table/TischHistorie.tsx:436-444
Warum: `positionen` und `totalPrice` sind im Zeilenmodell nicht optional und werden vom einzigen Aufrufer immer gesetzt.
Vorschlag: Beide Props verpflichtend machen, `<Receipt …/>` unbedingt rendern, Zeilen 438-444 löschen.
Aufwand: S · Status: unverifiziert (minor)

**[`Zeilenmodell` mischt deutsche und englische Feldnamen]** (convention, docs/language.md §TypeScript-Typen)
Datei: frontend/src/service/components/table/TischHistorie.tsx:52-65
Warum: `title`, `date`, `totalPrice` stehen neben `betrag`, `kommentar`; `HistoryRow` neben `HistorieDetail`.
Vorschlag: `title`→`titel`, `date`→`zeitpunkt`, `totalPrice`→`gesamtCents`, `HistoryRow`→`HistorieZeile`; `ReceiptPosition.totalPrice` mitziehen.
Aufwand: M · Status: unverifiziert (minor)

**[Datums-/Zeitformat inline dupliziert]** (code-smell, cleanup/principles.md → DRY)
Datei: frontend/src/service/components/table/TischHistorie.tsx:424-433
Warum: Das `toLocaleString('de-DE', …)`-Optionsobjekt steht identisch in DirektverkaufHistorie.tsx:219-227.
Vorschlag: `formatDatumZeit(iso)` in `frontend/src/lib/utils.ts` ergänzen und aus beiden Drawern aufrufen.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/service/components/table/Zahlung.tsx

**[`unbezahlteMengen` ist eine Lookup-Tabelle auf sich selbst]** (code-smell, cleanup/code-smells.md → Unnecessary Complexity)
Datei: frontend/src/service/components/table/Zahlung.tsx:57-60
Warum: Die Map wird aus `positionen` gebaut und nur mit denselben IDs gelesen; `|| 0` kann nie greifen.
Vorschlag: Zeilen 57-60 löschen und in `renderPosition` (Zeile 139) `unbezahlteMenge={position.menge}` übergeben.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/service/components/table/Receipt.tsx

**[React-Key aus Name und Preis ist nicht eindeutig]** (correctness)
Datei: frontend/src/service/components/table/Receipt.tsx:19-23
Warum: Dieselbe Variante aus zwei Bestellungen bleibt zwei Positionen; der Key kollidiert im geldführenden Beleg.
Vorschlag: Stabile ID durch `ReceiptPosition` reichen (`positionId` bzw. `varianteId` aus drawerUtils.ts:135-161) und darauf keyen.
Aufwand: S · Status: unverifiziert (minor)

**[Englische Namen für deutsche Domänenbegriffe]** (convention, docs/language.md)
Datei: frontend/src/service/components/table/Receipt.tsx:3-15
Warum: `Receipt`/`ReceiptPosition` für den Beleg, `CommentField.tsx` mit Export `KommentarField`, `ProductList`-Props neben `Produkt`/`Variante`.
Vorschlag: Auf die dokumentierten Begriffe umbenennen (Beleg/BelegPosition, KommentarField.tsx, `produkte`/`variantenMengen`) oder die Abweichung in docs/language.md festhalten.
Aufwand: M · Status: unverifiziert (minor)

### frontend/src/service/components/direktverkauf/DirektverkaufStornoDrawer.tsx

**[Gesamt-Zeile handgebaut statt `GesamtZeile`]** (principle, cleanup/principles.md → DRY)
Datei: frontend/src/service/components/direktverkauf/DirektverkaufStornoDrawer.tsx:112-117
Warum: Der Schwester-Drawer nutzt `GesamtZeile` mit identischem Label; die Paddings weichen bereits ab.
Vorschlag: Durch `<GesamtZeile label="Stornierung gesamt" betrag={totalPrice} />` ersetzen.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/service/components/direktverkauf/DirektverkaufAbschluss.tsx

**[Reset-Feldliste je Komponente doppelt geschrieben]** (principle, cleanup/principles.md → DRY)
Datei: frontend/src/service/components/direktverkauf/DirektverkaufAbschluss.tsx:62-94
Warum: Dieselben vier Setter stehen im warLeer-Effekt und in `onSuccess`; dasselbe Muster in ZahlungAbschluss und BestellungAbschluss.
Vorschlag: Pro Komponente ein `resetEingaben()` definieren und aus beiden Stellen aufrufen.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/service/components/direktverkauf/DirektverkaufHistorie.test.tsx

**[Englische Testtitel in einer sonst deutschen Suite]** (convention, cleanup/readability.md → Naming)
Datei: frontend/src/service/components/direktverkauf/DirektverkaufHistorie.test.tsx:56-167
Warum: Vier englische `it()`-Titel neben deutschen Inline-Kommentaren; drawerUtils.test.ts mischt beide Sprachen in einer Datei.
Vorschlag: Titel auf Deutsch übersetzen und drawerUtils.test.ts:14-231 vereinheitlichen.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/service/components/Stepper.test.tsx

**[Assertion auf Tailwind-Klassen statt auf Verhalten]** (test-quality)
Datei: frontend/src/service/components/Stepper.test.tsx:26-31
Warum: Geprüft wird die Abwesenheit zweier Klassennamen eines nicht mehr existierenden Designs.
Vorschlag: Zeilen 28-31 löschen; `toBeDisabled()` und die Klick-Assertion decken die Anforderung ab.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/service/components/table/HistorieUmbuchungDrawer.tsx

**[Verschachtelte Ternär-Kette mit `null`-Zweig, uneinheitlich benannt]** (readability, cleanup/readability.md → Clever Code)
Datei: frontend/src/service/components/table/HistorieUmbuchungDrawer.tsx:94-100
Warum: Vier Ausgänge in einer Kette; derselbe Wert heißt hier `disabledReason`, im Schwester-Drawer `hinweisGrund`.
Vorschlag: Als lokale Funktion mit Guard Clauses schreiben und in beiden Drawern `hinweisGrund` nennen.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/service/components/ErfolgsPop.tsx

**[Kommentare zitieren ein gelöschtes Design-Dokument]** (docs-accuracy)
Datei: frontend/src/service/components/ErfolgsPop.tsx:4-5 (ebenso MeinTischCard.tsx:43-45, Zahlung.tsx:288-289, DockActionButton.tsx:39-41)
Warum: „Motion-Inventar“/„Handoff“ verweisen auf `docs/prds/design_handoff_spektral_redesign/`, das nicht existiert.
Vorschlag: Auf `docs/adrs/06_spektral-branding-app.md` verweisen oder den Wert samt Grund ausschreiben.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/service/components/table/Zahlung.test.tsx

**[Testkommentare tragen tote Plan-Kennungen]** (docs-accuracy, AGENTS.md Regel 18)
Datei: frontend/src/service/components/table/Zahlung.test.tsx:14-16 (ebenso Bestellung.test.tsx:15, HistorieUmbuchungDrawer.test.tsx:204-205, TischHistorie.test.tsx:309,400-402, DockActionButton.test.tsx:61)
Warum: „seit A1“, „A2:“, „(#7)“ lösen sich nirgends auf; „der frühere …-Toast entfällt“ verstößt gegen Regel 18.
Vorschlag: Aktuellen Sachverhalt ohne Phasen-Label formulieren.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/service/components/MeinTischCard.tsx

**[Backend-Regel „eigene offene Arbeit“ im Browser nachgebaut]** (principle, AGENTS.md Regel 9)
Datei: frontend/src/service/components/MeinTischCard.tsx:89-108
Warum: `countOffenePositionen` dupliziert `ComputeEigeneArbeitAnTisch`; die Zählungen weichen bereits ab (Set vs. Einträge).
Vorschlag: Zählwerte im Tisch-Session-DTO liefern, `fuerMichErledigt` lesen und `countOffenePositionen` löschen.
Aufwand: M · Status: unverifiziert (minor)

### frontend/src/service/components/table/ProductList.tsx

**[Kategorie-Tabs ohne Test]** (test-quality)
Datei: frontend/src/service/components/table/ProductList.tsx:24-52
Warum: Der Fallback für eine verschwundene aktive Kategorie ist ungetestet; alle Fixtures nutzen eine Kategorie.
Vorschlag: `ProductList.test.tsx` mit zwei Kategorien: Tabs sichtbar, Filterung, Einzelkategorie ohne Tabs, Fallback auf `kategorien[0]`.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/service/table/TischBackend.ts

**[Vier exportierte Request-Typen ungenutzt]** (code-smell, cleanup/code-smells.md → Dead Code)
Datei: frontend/src/service/table/TischBackend.ts:50-106
Warum: Die Methoden leiten dieselben Typen inline per `z.infer` ab; jeder Command hat zwei Schreibweisen.
Vorschlag: Benannte Typen importieren wie in DirektverkaufBackend.ts:24-26 oder die Aliase löschen.
Aufwand: S · Status: unverifiziert (minor)

**[Beleg-Druck bricht die Command-Schema-Konvention und dupliziert sich]** (convention, cleanup/code-smells.md → Unnecessary Complexity)
Datei: frontend/src/service/table/TischBackend.ts:64-92
Warum: Positionale Primitive plus anonyme Inline-Bodies; ein Endpunkt trägt drei Methodennamen.
Vorschlag: `KassenbelegDruckenSchema`/`StornobelegDruckenSchema` exportieren, einen privaten Helper für den gemeinsamen Aufruf, `belegDrucken`→`kassenbelegDrucken`.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/service/schemas.ts

**[Pass-Through-Re-Export und dritte Steuersatz-Definition]** (convention, cleanup/principles.md → DRY)
Datei: frontend/src/service/schemas.ts:3-10
Warum: `DateStringSchema` ist über zwei Pfade erreichbar; das Steuersatz-Enum steht dreimal im Code.
Vorschlag: Re-Export löschen und `@/lib/utils` direkt importieren; Steuersatz einmal definieren und in beiden Bereichen importieren.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/service/DirektverkaufPage.tsx

**[Erfolgs-Pop und Höhen-Shell aus TablePage kopiert]** (principle, cleanup/principles.md → DRY)
Datei: frontend/src/service/DirektverkaufPage.tsx:26-95
Warum: Die `calc(100dvh-…)`-Klasse und das erfolg/zeigeErfolg/erfolgSchliessen-Trio sind byte-identisch mit TablePage.tsx:118-125,299-305.
Vorschlag: Gemeinsame Shell samt benannter Höhen-Konstante und den Erfolgs-State an eine Stelle ziehen.
Aufwand: M · Status: unverifiziert (minor)

### frontend/src/service/TableSelectionPage.tsx

**[Fünfstufige Ternär-Kette über 44 JSX-Zeilen]** (readability, cleanup/readability.md → Deep Nesting)
Datei: frontend/src/service/TableSelectionPage.tsx:82-125
Warum: Der Normalfall steht im innersten Zweig hinter vier Bedingungen.
Vorschlag: Body in eine Funktion mit Guard Clauses für Suche, Laden und Leerzustand auslagern.
Aufwand: S · Status: unverifiziert (minor)

**[Fußleiste dupliziert ServiceDock samt zweiter Freiraum-Konstante]** (principle, cleanup/principles.md → DRY)
Datei: frontend/src/service/TableSelectionPage.tsx:22,127-140
Warum: Die Klassenkette ist byte-identisch mit ServiceDock.tsx:46; `fussleisteFreiraum` und `dockFreiraum` sind zwei Quellen derselben Geometrie.
Vorschlag: ServiceDock wiederverwenden oder Leiste plus Freiraum-Konstante in ein Modul ziehen.
Aufwand: M · Status: unverifiziert (minor)

### frontend/src/service/table/Tisch.ts

**[Formular-Regel mit UI-Meldung in einem Response-Schema]** (code-smell, cleanup/code-smells.md → Leaky Abstraction)
Datei: frontend/src/service/table/Tisch.ts:6-13
Warum: `min(3)` kann Server-Daten ablehnen, die Meldung ist unerreichbar; dasselbe Feld ist bei Zeile 20 und 31 unbeschränkt.
Vorschlag: Im Response-Schema `z.string()` verwenden wie in den Schwester-Schemas; die Eingaberegel bleibt im Admin-Formular.
Aufwand: S · Status: unverifiziert (minor)

**[Glossar-Eintrag `EigeneUebersicht` unvollständig]** (docs-accuracy)
Datei: frontend/src/service/table/Tisch.ts:42-50
Warum: docs/language.md:242 nennt vier von sieben JSON-Schlüsseln, darunter fehlt `abzugebenCents`.
Vorschlag: Alle sieben Schlüssel in docs/language.md:242 aufnehmen und die Beschreibung um Rücknahmen/Abzugeben erweitern.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/service/table/Bestellung.ts

**[Englische Kommentare in einem sonst deutschen Kontext]** (convention, AGENTS.md Regel 6)
Datei: frontend/src/service/table/Bestellung.ts:49-50 (ebenso Umbuchung.ts:33)
Warum: Zwei Kommentare sind englisch, einer mischt beide Sprachen in einem Satz.
Vorschlag: Beide auf Deutsch umschreiben, analog zu Bestellung.ts:15-16 und Stornierung.ts:15-19.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/service/TablePage.tsx

**[Backend-Client dreifach instanziiert]** (code-smell, cleanup/code-smells.md → Redundant Abstractions)
Datei: frontend/src/service/TablePage.tsx:24
Warum: Dieselbe zustandslose Client-Instanz entsteht in TablePage, hooks.ts:8 und TischAuswahlDrawer.tsx:23.
Vorschlag: Instanz aus dem hooks-Modul exportieren und importieren statt neu zu konstruieren.
Aufwand: S · Status: unverifiziert (minor)

**[Tab-Bezeichner englisch, Schwesterseite deutsch]** (convention, docs/language.md §1)
Datei: frontend/src/service/TablePage.tsx:231-239
Warum: `order`/`payment`/`history` stehen neben `verkaufen`/`historie` in DirektverkaufPage.tsx:40-44.
Vorschlag: Werte auf `bestellen`/`kassieren`/`historie` umbenennen; nur interne Bezeichner, Tests bleiben unberührt.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/service/direktverkauf/Direktverkauf.ts

**[Zwei TS-Namen ohne Entsprechung in Go und Glossar]** (convention, docs/language.md)
Datei: frontend/src/service/direktverkauf/Direktverkauf.ts:10-15
Warum: `VerkaufPositionInput` ist feldgleich mit `BestellPositionInput`, `VerkaufPosition` ist die dokumentierte `Position`.
Vorschlag: Bestehende Schemas wiederverwenden oder beide Namen in die Glossar-Tabelle von docs/language.md aufnehmen.
Aufwand: M · Status: unverifiziert (minor)

**[Kommentarpflicht bei Direktverkauf-Storno ungetestet]** (test-quality)
Datei: frontend/src/service/direktverkauf/Direktverkauf.ts:24-28
Warum: Die identische Regel am Tisch hat einen Test; `meldeBelegStatus` ist ebenfalls ungetestet.
Vorschlag: `Direktverkauf.test.ts` analog zu Stornierung.test.ts anlegen und beleg.test.ts um die zwei Status-Zweige erweitern.
Aufwand: M · Status: unverifiziert (minor)

### frontend/src/admin/tse/tseAmpel.ts

**[Reine Logik importiert aus dem Hook-Modul]** (architecture, architecture.md → Dependency Direction)
Datei: frontend/src/admin/tse/tseAmpel.ts:1-24
Warum: `RUECKSTAND_WARN_SEKUNDEN` liegt in hooks.ts, das react-query und einen Backend-Client zieht; drei Tests stubben den Wert nach.
Vorschlag: Konstante nach tseAmpel.ts verschieben und in hooks.ts von dort re-exportieren.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/admin/reporting/utils.ts

**[Allgemeine Datums-Formatierer im Reporting-Modul, Exklusivitäts-Aussage falsch]** (docs-accuracy, architecture.md → Bounded Context Violations)
Datei: frontend/src/admin/reporting/utils.ts:7-58
Warum: admin/kasse und admin/finanzamt importieren quer; lib/utils.ts:105-108 setzt dieselben Uhrzeit-Optionen erneut.
Vorschlag: Formatierer nach `src/lib/utils.ts` ziehen, nur `formatServicekraft` bleibt; die „einzige Stelle“-Aussage erst danach schreiben.
Aufwand: M · Status: unverifiziert (minor)

### frontend/src/admin/reporting/AdminDashboardPage.tsx

**[`formatStand` zweckentfremdet für einen ISO-Zeitpunkt]** (readability, cleanup/code-smells.md → Leaky Abstraction)
Datei: frontend/src/admin/reporting/AdminDashboardPage.tsx:59-61
Warum: Date→getTime()→Date-Umweg, obwohl `formatLocalTime(utcString)` genau diese Eingabe nimmt.
Vorschlag: `formatLocalTime(kassensitzung.eroeffnetAm)` verwenden; identische Ausgabe.
Aufwand: S · Status: unverifiziert (minor)

**[Befund-Kennung „NEU02“ ohne Ziel]** (docs-accuracy)
Datei: frontend/src/admin/reporting/AdminDashboardPage.tsx:44-45
Warum: Der zitierte Plan wurde nach dem Merge gelöscht; die Kennung löst nirgends auf.
Vorschlag: Die Regel selbst ausschreiben oder auf den erhaltenen ADR verweisen; gleiches bei DruckstationConfigPage.tsx:458.
Aufwand: S · Status: unverifiziert (minor)

**[Fixture ohne `bonArt` prüft nur den Fallback]** (test-quality)
Datei: frontend/src/admin/reporting/AdminDashboardPage.test.tsx:60-66
Warum: Die Druckauftrag-Fixtures tragen nur `{ id }`; die Bon-Art-Logik könnte brechen, ohne den Test rot zu machen.
Vorschlag: Reale `bonArt` setzen und die daraus folgende Formulierung prüfen.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/admin/kasse/Kassensitzung.ts

**[Kommentar behauptet einen Re-Export, den es nicht gibt]** (docs-accuracy, AGENTS.md Regel 18)
Datei: frontend/src/admin/kasse/Kassensitzung.ts:41-42
Warum: Kein Modul außerhalb admin/kasse importiert `KassensitzungSchema`; reporting hat ein eigenes Schema. Zudem einziger englischer Kommentar der Datei.
Vorschlag: Auf Deutsch umschreiben und nur den tatsächlichen Nutzer (`KasseBackend.ts`) nennen.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/admin/kasse/KasseAbschliessenSection.tsx

**[Submit-Vertrag von Hand nachgebaut statt `useActionSubmit`]** (principle, cleanup/principles.md → DRY)
Datei: frontend/src/admin/kasse/KasseAbschliessenSection.tsx:92-179
Warum: Loading, try/catch, console.error und Toast sind dupliziert; alle anderen Aktionen des Bereichs nutzen den Hook.
Vorschlag: `useActionSubmit` verwenden und den `signaturen_ausstehend`-Sonderfall im übergebenen `fn` behalten; Verhalten unverändert.
Aufwand: M · Status: unverifiziert (minor)

### frontend/src/admin/tse/TSEKonfigurationSection.tsx

**[„Client ist registriert“ zweimal, einmal case-sensitiv]** (principle, cleanup/principles.md → DRY)
Datei: frontend/src/admin/tse/TSEKonfigurationSection.tsx:209-210
Warum: `=== 'REGISTERED'` widerspricht der geteilten, case-insensitiven Regel in TSEBackend.ts:27-34.
Vorschlag: Die geteilte Regel aufrufen oder `clientIstRegistriert(status)` aus TSEBackend exportieren.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/admin/kasse/EroeffnenSection.tsx

**[TSE-Warnung leitet aus unbekanntem Ladezustand eine Tatsache ab]** (correctness)
Datei: frontend/src/admin/kasse/EroeffnenSection.tsx:80-87
Warum: Bei Fehler oder laufender Query ist `tseKonfiguration` undefined; der Dialog behauptet „keine TSE eingerichtet“.
Vorschlag: `isPending`/`error` mitnehmen: Submit zurückstellen bzw. „TSE-Status konnte nicht geprüft werden“ anzeigen; gleiches Muster in tseAmpel.ts:21.
Aufwand: S · Status: unverifiziert (minor)

**[Formularschema im Komponentenrumpf statt auf Modulebene]** (convention, cleanup/readability.md → Naming)
Datei: frontend/src/admin/kasse/EroeffnenSection.tsx:44-48
Warum: Drei Kasse-Formulare bauen ihr Schema pro Render neu; acht andere Admin-Formulare definieren es einmal.
Vorschlag: Die drei Schemas samt Typ auf Modulebene heben, wie NewTischDialog.tsx:27.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/admin/kasse/KassensitzungPage.tsx

**[Re-Exports existieren nur für den Test-Import]** (code-smell, cleanup/code-smells.md → Redundant Abstractions)
Datei: frontend/src/admin/kasse/KassensitzungPage.tsx:22-23
Warum: Kein Produktionscode nutzt die beiden Re-Exports; die Seite scheint Komponenten zu besitzen, die sie nur durchreicht.
Vorschlag: Beide Zeilen löschen und im Test direkt aus den Modulen importieren.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/admin/tse/hooks.ts

**[Invalidierungs-Paar viermal ausgeschrieben]** (principle, cleanup/principles.md → DRY)
Datei: frontend/src/admin/tse/hooks.ts:37-90
Warum: `TSE_KONFIGURATION_KEY` plus `TSE_STATUS_KEY` steht in vier Mutationen identisch.
Vorschlag: Lokalen Helper `invalidateTSE(queryClient)` extrahieren und viermal aufrufen.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/admin/finanzamt/BetreiberBackend.ts

**[Pflichtfelder werden clientseitig nicht validiert]** (correctness, AGENTS.md Regel 5)
Datei: frontend/src/admin/finanzamt/BetreiberBackend.ts:8-15
Warum: Vereinsname, Straße, PLZ, Ort sind blanke `z.string()`; leere Belegpflichtangaben erzeugen nur einen generischen Toast.
Vorschlag: `.trim().min(1, …)` je Feld spiegeln und die Fehler in BetreiberForm.tsx pro Feld anzeigen.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/admin/reporting/VerkaufStatistik.tsx

**[`KATEGORIE_LABEL` ein zweites Mal definiert]** (convention, cleanup/principles.md → DRY)
Datei: frontend/src/admin/reporting/VerkaufStatistik.tsx:7-15
Warum: Dieselbe Zuordnung wird bereits aus `admin/products/productGrouping.ts` exportiert.
Vorschlag: Von dort importieren und nur die String-Fallback-Logik lokal behalten.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/admin/reporting/LiveReportingSection.tsx

**[Kommentare zitieren nicht auflösbare Kennungen]** (docs-accuracy)
Datei: frontend/src/admin/reporting/LiveReportingSection.tsx:36-38 (ebenso SitzungsListe.tsx:12, TSEBackend.ts:80, TSEEinrichtungWizard.tsx:46,566)
Warum: „Design-Handoff 1a/1b“, „NEU02/NEU07“, „F2“/„F8“ existieren nicht; F2/F8 kollidieren mit dem Schema F-01…F-14.
Vorschlag: Tags entfernen und die Begründung ausschreiben oder auf ADR bzw. eine echte Anforderungs-ID verweisen.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/admin/tse/TSEEinrichtungWizard.tsx

**[934 Zeilen mit 14 Komponenten und Prop-Drilling der Zugangsdaten]** (large-refactor, architecture.md → Deep vs. Shallow Modules)
Datei: frontend/src/admin/tse/TSEEinrichtungWizard.tsx:1-934
Warum: Der sicherheitsrelevanteste Admin-Fluss liegt vollständig in einer Datei; apiKey/apiSecret laufen durch fünf Prop-Ebenen.
Vorschlag: Als eigenen Refactor planen: je Schritt eine Datei unter `admin/tse/einrichtung/`, Zugangsdaten über einen lokalen Context.
Aufwand: L · Status: unverifiziert (minor)

### frontend/src/admin/products/Produkt.ts

**[Ein Backend-Enum in vier Frontend-Fassungen]** (code-smell, cleanup/principles.md → DRY)
Datei: frontend/src/admin/products/Produkt.ts:28-33
Warum: `VarianteStatus`, `TischStatus`, `UserStatus` wiederholen `EntityStatus`; Produkt.ts nutzt zwei Fassungen gleichzeitig.
Vorschlag: Überall `EntityStatusSchema` aus lib/entityStatus.ts verwenden und die drei lokalen Kopien löschen.
Aufwand: M · Status: unverifiziert (minor)

### frontend/src/admin/tables/Tisch.ts

**[Namensschemas driften: nur Produkte trimmen]** (convention, cleanup/principles.md → DRY)
Datei: frontend/src/admin/tables/Tisch.ts:6-9
Warum: Drei Namen aus Leerzeichen passieren Tisch und Benutzer clientseitig, das Backend trimmt und lehnt generisch ab.
Vorschlag: `.trim()` ergänzen und die dreifach kopierte Meldung in ein gemeinsames Namensschema ziehen.
Aufwand: S · Status: unverifiziert (minor)

**[Kommentar begründet die Platzierung mit einer falschen Testbarkeit]** (docs-accuracy)
Datei: frontend/src/admin/tables/Tisch.ts:30-34
Warum: `tischGrouping.test.ts` importiert `WEITERE_GRUPPE` nie; die Präfix-Regel steht doppelt.
Vorschlag: Konstante zum einzigen Konsumenten in tischGrouping.ts verschieben und die doppelte Erklärung streichen.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/admin/products/productGrouping.ts

**[Stationsnamen doppelt, obwohl die Quelle schon importiert wird]** (code-smell, cleanup/principles.md → DRY)
Datei: frontend/src/admin/products/productGrouping.ts:23-30
Warum: `KATEGORIE_STATION_LABEL` wiederholt das geteilte `KATEGORIE_LABEL` aus settings/DruckstationBackend.ts:105-110.
Vorschlag: `KATEGORIE_LABEL` von dort importieren und die lokale Kopie löschen.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/admin/products/Products.tsx

**[Callback-Parameter, die niemand liest]** (code-smell, cleanup/principles.md → YAGNI)
Datei: frontend/src/admin/products/Products.tsx:30-37
Warum: `produktId`, `varianteId` und `status` werden durch zwei Komponenten gereicht und am Aufrufort verworfen.
Vorschlag: Callbacks auf das Genutzte reduzieren und die Wrapper-Closures in AdminProductsPage.tsx:84-92 entfernen.
Aufwand: S · Status: unverifiziert (minor)

**[Englische Bezeichner für deutsche Domänenbegriffe]** (large-refactor, docs/language.md §1)
Datei: frontend/src/admin/products/Products.tsx:12-38
Warum: `admin/products` benennt englisch, das Schwestermodul `admin/tables` deutsch — die Spaltung zieht sich durchs Repo.
Vorschlag: Einmal entscheiden: Abweichung in docs/language.md dokumentieren oder je Modul in einem eigenen Change umbenennen.
Aufwand: L · Status: unverifiziert (minor)

### frontend/src/admin/products/ProduktBackend.ts

**[Drei Request-Schemas exportiert, aber nur lokal genutzt]** (code-smell, cleanup/code-smells.md → Redundant Abstractions)
Datei: frontend/src/admin/products/ProduktBackend.ts:28-52
Warum: Der Export suggeriert einen geteilten Vertrag; Nachbarmethoden bauen den Body ohne benanntes Schema.
Vorschlag: `export` entfernen (auch bei `KATEGORIE_ORDER`) und einen Stil für alle Move-/Status-Methoden wählen.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/admin/users/Users.test.tsx

**[ResizeObserver-Stub je Testdatei kopiert]** (test-quality, cleanup/principles.md → DRY)
Datei: frontend/src/admin/users/Users.test.tsx:17-33 (ebenso products/Products.test.tsx:7-23)
Warum: Das Projekt hat mit `src/test/setup.ts` bereits eine Stelle für jsdom-Lücken (matchMedia-Polyfill).
Vorschlag: Stub einmal in `src/test/setup.ts` registrieren und die Klassen pro Datei löschen.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/admin/settings/DruckstationConfigPage.test.tsx

**[Assertion auf Tailwind-Klassennamen]** (test-quality)
Datei: frontend/src/admin/settings/DruckstationConfigPage.test.tsx:151-173
Warum: `container.querySelectorAll('[class*="overflow-y-auto"]')` bricht bei rein visuellem Refactor; das Verhalten ist darunter schon geprüft.
Vorschlag: Die beiden Klassen-Assertions löschen, die Verhaltens-Erwartungen behalten.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/admin/settings/DruckstationBackend.test.ts

**[Testtitel wechseln die Sprache innerhalb einer Datei]** (convention, cleanup/readability.md → Naming)
Datei: frontend/src/admin/settings/DruckstationBackend.test.ts:13-56
Warum: Englische und deutsche `it()`-Titel stehen nebeneinander; im Unit gilt dieselbe Uneinheitlichkeit dateiübergreifend.
Vorschlag: Eine Sprache festlegen (Deutsch passt zu den deutschen Assertions) und in `.github/instructions/frontend.instructions.md` notieren.
Aufwand: M · Status: unverifiziert (minor)

### frontend/src/admin/tables/NewTischDialog.tsx

**[Anlege-Dialoge erfinden Backend-Felder]** (principle, AGENTS.md Regel 9)
Datei: frontend/src/admin/tables/NewTischDialog.tsx:56-63 (ebenso NewUserDialog.tsx:66-75, NewProductDialog.tsx:76-83, NewVariantDialog.tsx:60-66)
Warum: `status`, `createdAt`, `updatedAt`, `saldoCents` werden clientseitig gesetzt, obwohl kein Konsument sie liest.
Vorschlag: `created`-Callbacks auf id plus Formulardaten verengen; die Listen werden ohnehin invalidiert.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/admin/users/UsersSpalten.ts

**[Zehn Kommentare zitieren einen fehlenden Design-Handoff]** (docs-accuracy)
Datei: frontend/src/admin/users/UsersSpalten.ts:1-5 (u. a. HeaderGlow.tsx:1, StatusDot.tsx:14, UserRow.tsx:40, HelferPanels.tsx:3-5, VariantChip.tsx:27, TischItem.tsx:14, DruckstationConfigPage.tsx:39,67)
Warum: `docs/prds/design_handoff_spektral_redesign/` existiert nicht; die zitierte Autorität ist nicht auflösbar.
Vorschlag: Je Stelle die Regel selbst ausschreiben oder auf `docs/adrs/06_spektral-branding-app.md` verweisen.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/admin/products/ProductItem.tsx

**[Drei Anführungszeichen-Konventionen für dieselbe deutsche Meldung]** (convention, AGENTS.md Regel 6)
Datei: frontend/src/admin/products/ProductItem.tsx:282-285 (ebenso EditVariantDialog.tsx:167, EditTischDialog.tsx:155, UserRow.tsx:139, Toasts in den drei Admin-Seiten)
Warum: `&quot;`, `&ldquo;/&rdquo;` und `"` stehen neben korrektem `„…“` in denselben Dateien.
Vorschlag: Durchgängig `„{name}“` verwenden, wie in den aria-Labels bereits geschehen.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/admin/users/User.ts

**[`toUsername` ohne Test]** (test-quality)
Datei: frontend/src/admin/users/User.ts:13-22
Warum: Die Funktion formt den Login-Bezeichner (Umlaute, ß, Sonderzeichen); jede andere reine Funktion des Bereichs ist getestet.
Vorschlag: Fälle „Jürgen Müller“→`juergenmueller`, „Weiß“→`weiss` und reine Satzzeichen in users/User.test.ts ergänzen.
Aufwand: S · Status: unverifiziert (minor)

### frontend/eslint.config.js

**[Handgeschriebener UI-Code außerhalb aller Qualitäts-Gates]** (ops, cleanup/code-smells.md → Template Pollution)
Datei: frontend/eslint.config.js:17
Warum: `src/components/ui` ist von ESLint und Prettier ausgenommen, enthält aber eigene Effekte (tabs.tsx:68-98, drawer.tsx:74-77).
Vorschlag: Eigene Dateien nach `components/common` ziehen oder das Ignore auf unveränderte shadcn-Dateien einengen.
Aufwand: M · Status: unverifiziert (minor)

### frontend/src/lib/errorMessages.test.ts

**[Fehlermeldungs-Katalog im Test wortgleich gespiegelt]** (test-quality, cleanup/principles.md → DRY)
Datei: frontend/src/lib/errorMessages.test.ts:9-165
Warum: Der Test erkennt nur unabsichtliche Änderungen; vier von 45 Codes fehlen bereits in der Kopie.
Vorschlag: Über `Object.entries(commonErrorMessages)` iterieren und nur das Zweigverhalten handgeschrieben prüfen.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/lib/errorMessages.ts

**[Kein Text für `login_throttled` und `invalid_kassensitzung`]** (correctness)
Datei: frontend/src/lib/errorMessages.ts:69-70
Warum: Die gedrosselte Helferin sieht „Anmeldung fehlgeschlagen“ statt eines Wartehinweises; der Cooldown dauert bis 15 Minuten.
Vorschlag: Beide Codes in `commonErrorMessages` ergänzen, alphabetisch neben `rate_limited` bzw. `invalid_kassensitzung_nr`.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/lib/Backend.ts

**[`post` lügt über seinen Rückgabetyp ohne Schema]** (code-smell, cleanup/code-smells.md → Type Escape Hatches)
Datei: frontend/src/lib/Backend.ts:172-174
Warum: `return {} as TResponse` liefert ein leeres Objekt unter beliebigem Typ; heute nur zufällig sicher.
Vorschlag: Overloads trennen: ohne Schema `Promise<void>`, mit Schema `Promise<T>`.
Aufwand: M · Status: unverifiziert (minor)

**[HTTP-Client navigiert selbst]** (architecture, architecture.md → Separation from Frameworks)
Datei: frontend/src/lib/Backend.ts:134-138
Warum: `window.location.href = '/login'` umgeht den Router; der Test muss dafür einen jsdom-Fehler stummschalten.
Vorschlag: Nur ausloggen und werfen; den Redirect den Route-Guards überlassen — bewusst planen, da verhaltensändernd.
Aufwand: M · Status: unverifiziert (minor)

### frontend/src/routes.ts

**[`AdminGuard` und `ServiceGuard` ungetestet]** (test-quality)
Datei: frontend/src/routes.ts:28-44
Warum: Die einzigen clientseitigen Autorisierungs-Gates haben weder Unit- noch e2e-Abdeckung, `ServiceTableGuard` dagegen schon.
Vorschlag: In routes.test.ts ergänzen: ohne Token Redirect, Service-Token scheitert am AdminGuard, Admin-Token passiert beide.
Aufwand: S · Status: unverifiziert (minor)

**[Sprachmix in Kommentaren und Testnamen]** (convention, cleanup/readability.md → Naming)
Datei: frontend/src/routes.test.ts:25-48
Warum: Englische und deutsche `it()`-Titel wechseln in einer Datei; dasselbe in lib/Backend.ts:46, hooks/use-mengen.ts:6-19.
Vorschlag: Eine Konvention festlegen (Deutsch passt zur Ubiquitous Language) und die Ausreißer angleichen.
Aufwand: M · Status: unverifiziert (minor)

### frontend/src/components/common/FormFields.tsx

**[Geteilte Schicht importiert aus dem Admin-Bereich]** (architecture, architecture.md → Dependency Direction)
Datei: frontend/src/components/common/FormFields.tsx:14-19
Warum: Der Login-Pfad hängt über `toUsername` am Admin-Modulbaum; die Abhängigkeitsrichtung ist invertiert.
Vorschlag: `Kategorie`, `Steuersatz`, `UserRole`, `toUsername` nach `lib/` ziehen und aus beiden Bereichen importieren.
Aufwand: M · Status: unverifiziert (minor)

**[`any`-Generics und neun `as Path<…>`-Casts]** (code-smell, cleanup/code-smells.md → Type Escape Hatches)
Datei: frontend/src/components/common/FormFields.tsx:47-49
Warum: Jede Feldkomponente nennt ihren Feldnamen zweimal; ein Tippfehler im String fällt dem Compiler nicht auf.
Vorschlag: Einen typisierten Helper einführen, der Name und Props-Typ einmal erzeugt, und den Resolver-Typ statt `any` setzen.
Aufwand: M · Status: unverifiziert (minor)

### frontend/src/components/ui/button.tsx

**[Kommentar kündigt eine Phase an, die es nicht gibt]** (convention, AGENTS.md Regel 18)
Datei: frontend/src/components/ui/button.tsx:12-15
Warum: „Phase 9 verdrahtet dieses Treatment …“ verweist auf keinen existierenden Plan.
Vorschlag: Den zweiten Satz löschen, die Begründung für das Disabled-Token behalten.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/components/common/AuthLayout.tsx

**[Verweis auf „Handoff Delta B“ ohne Ziel]** (docs-accuracy)
Datei: frontend/src/components/common/AuthLayout.tsx:3
Warum: Weder `docs/prds/design_handoff_spektral_redesign/` noch „Delta B“ existieren im Repo.
Vorschlag: Klammer streichen, den sachlichen Teil (dekorativ, aria-hidden, print-hidden) behalten.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/components/common/VersionsHinweis.tsx

**[Vergleich mit einem nicht mehr existierenden Vorgänger]** (convention, AGENTS.md Regel 18)
Datei: frontend/src/components/common/VersionsHinweis.tsx:23
Warum: „wie sein Vorgänger“ trägt für einen neuen Leser keine Information und datiert den Kommentar.
Vorschlag: Die zwei Wörter streichen; der Rest des Satzes begründet die Platzierung eigenständig.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/hooks/use-count-up.test.ts

**[Kommentar nennt den falschen Mechanismus]** (docs-accuracy)
Datei: frontend/src/hooks/use-count-up.test.ts:21-22
Warum: jsdom hat hier sehr wohl `matchMedia`; `src/test/setup.ts` meldet bewusst `prefers-reduced-motion`.
Vorschlag: Den echten Grund nennen und die Aussage in use-count-up.ts:12,55 angleichen.
Aufwand: S · Status: unverifiziert (minor)

### frontend/tsconfig.app.json

**[`baseUrl` samt Deprecation-Unterdrückung und doppeltem Alias]** (code-smell, cleanup/code-smells.md → Redundant Explicit Defaults)
Datei: frontend/tsconfig.app.json:12-14
Warum: `ignoreDeprecations` wird nur wegen `baseUrl` getragen; tsconfig.json deklariert denselben Alias inert erneut.
Vorschlag: Mit `tsc -b` prüfen, dass `paths` allein auflöst, dann `baseUrl`, `ignoreDeprecations` und den toten Block entfernen.
Aufwand: S · Status: unverifiziert (minor)

### docs/handbuch.md

**[Seitenstruktur-Tabelle widerspricht dem Router]** (docs-accuracy)
Datei: docs/handbuch.md:420
Warum: Die Admin-Zeile nennt vier Seiten und `DruckerConfigPage`; routes.ts:89-155 registriert neun Routen mit `DruckstationConfigPage`.
Vorschlag: Zeile auf die neun registrierten Routen aktualisieren und den Komponentennamen korrigieren.
Aufwand: S · Status: unverifiziert (minor)

## Website und E2E

Umfang: Astro-Website (`website/`) und Playwright-Suite mit Website-Tooling (`e2e/`).

60 Einzelbefunde, nach Zusammenführung 44 Einträge: 0 Blocker, 6 Major, 38 Minor.

### website/src/styles/brand.css

**[Elf Kommentarstellen zitieren die nicht existierende `docs/prds/prd-website-redesign.md`]** (docs-accuracy, AGENTS.md Regel 18 / No dead links)
Datei: website/src/styles/brand.css:2-3
Warum: Jeder Leser folgt einem Pfad, der ins Leere zeigt, und die Begründung wird unprüfbar.
Vorschlag: PRD-/Plan-Zitate ersatzlos streichen, nur die im Kommentar bereits stehende Ist-Begründung behalten. Stellen: brand.css:3, links.ts:9-11, live-demo.ts:9, anfrage-mailto.ts:12, link-rewriter.ts:3, externalize-inline-scripts.ts:6, astro.config.mjs:52, Hero.astro:11, Preis.astro:8, Ablauf.astro:4+8, Download.astro:5, Faq.astro:4, FaqAccordion.tsx:4, FeatureExplorer.tsx:13, LiveDemo.tsx:21, AnfrageFormular.tsx:16.
Aufwand: M · Status: bestätigt

**[„Übergangsregel" erklärt `--brand-*` und `.btn` zu auslaufenden Aliassen]** (docs-accuracy, code-smells.md → Unnecessary Comments)
Datei: website/src/styles/brand.css:11-17
Warum: Der Kommentar lädt zum Löschen von Tokens ein, die landing.css und starlight.css tragen.
Vorschlag: Absatz 11-17 löschen; stattdessen benennen, was gilt: `--brand-green*` sind die konstanten Fülltöne für weiße Schrift in beiden Themes, `.btn` ist in landing.css:352 definiert.
Aufwand: S · Status: bestätigt

**[`--spectral`/`--spectral-v` hardcodieren die Hellwerte statt der `--sp-*`-Tokens]** (code-smell, principles.md → DRY)
Datei: website/src/styles/brand.css:59-84
Warum: Die Dark-Overrides in Zeile 124-129 erreichen den Markenverlauf nie, der Kommentar behauptet das Gegenteil.
Vorschlag: Verlauf aus den Tokens bilden (`linear-gradient(100deg, var(--sp-red), …)`) oder im Kommentar sagen, dass der Verlauf in beiden Themes identisch bleibt.
Aufwand: S · Status: unverifiziert (minor)

### website/src/layouts/Landing.astro

**[Rund 80 Kommentarstellen erzählen Projektgeschichte statt Ist-Zustand]** (convention, AGENTS.md Regel 18)
Datei: website/src/layouts/Landing.astro:23-24
Warum: „Phase 8", „Handoff-Prototyp", „das frühere public/mobile-nav.js", „Referenz 09-light/04-dark" verweisen auf nichts im Repo.
Vorschlag: Jeden Kommentar auf die heute gültige Aussage kürzen, Phasennummern und Vorher/Nachher-Klauseln streichen. Hauptstellen: Landing.astro:6,23-24,36-39; Hero.astro:16-19; Preis.astro:10-14; MobileNav.tsx:6; ThemeImage.astro:4; FaqAccordion.tsx:13-20,68,97; FeatureExplorer.tsx:20; links.ts:9-10; landing.css:11-12,68,282,326,382; LiveDemo.tsx:32.
Aufwand: L · Status: bestätigt

### website/src/components/Hero.astro

**[Drei `ThemeImage`-Paare laden sechs Bilder mit `loading="eager"`]** (code-smell, principles.md → KISS)
Datei: website/src/components/Hero.astro:94-124
Warum: 25–61 KB der falschen Theme-Variante werden auf dem LCP-Pfad geladen und nie gezeigt.
Vorschlag: `ThemeImage` so erweitern, dass nur die ohne `data-theme` sichtbare Variante `eager` lädt; Frames 2 und 3 lazy laden.
Aufwand: S · Status: unverifiziert (minor)

**[„Beta 1.0" steht dreimal hartkodiert und widerspricht CHANGELOG.md]** (boundary-consistency, principles.md → DRY)
Datei: website/src/components/Hero.astro:42
Warum: Der Release-Status hat drei unabgestimmte Kopien; CHANGELOG.md führt bereits `## [1.0.0]` als stabile Version.
Vorschlag: Label einmal in `src/lib/links.ts` exportieren und in Hero.astro:42, Download.astro:50, fuer-vereine.astro:166 verwenden.
Aufwand: S · Status: unverifiziert (minor)

### website/src/lib/externalize-inline-scripts.ts

**[`\b` im Type-Regex liest `data-type` als Script-Type]** (correctness, code-smells.md → fehlerhafte Parsing-Heuristik)
Datei: website/src/lib/externalize-inline-scripts.ts:43-48
Warum: Ein Skript mit `data-type` bliebe inline und würde von der Produktiv-CSP `script-src 'self'` stumm blockiert.
Vorschlag: Attribut verankern: `/(?:^|\s)type\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s>]+))/i`.
Aufwand: S · Status: unverifiziert (minor)

### website/src/pages/fuer-vereine.astro

**[Schritt-Badges unterschreiten die WCAG-AA-Textschwelle 4,5:1]** (correctness, brand.css:78-84)
Datei: website/src/pages/fuer-vereine.astro:117-123
Warum: Weiß auf `--sp-teal` ergibt 3,44:1 und auf `--sp-red` 4,42:1 bei 13px fettem Text.
Vorschlag: Badges auf `--brand-green` umstellen oder auf das Muster aus Ablauf.astro (getönte Fläche plus `--sp-*-text`); dasselbe für LiveDemo.tsx:132-141.
Aufwand: S · Status: unverifiziert (minor)

### website/src/styles/starlight.css

**[Helle Palette liegt auf blankem `:root`, Starlight ist dark-first]** (correctness, Theme-Kontrakt)
Datei: website/src/styles/starlight.css:16-30
Warum: Ohne `data-theme` (JS aus) mischen sich heller Hintergrund und Starlights dunkle Grauwerte, Inline-Code erreicht ~1,3:1.
Vorschlag: Polarität spiegeln — dunkle Werte auf `:root`, helle auf einen Light-Selektor — oder `<html data-theme="light">` serverseitig setzen.
Aufwand: M · Status: unverifiziert (minor)

### website/src/lib/anfrage-mailto.ts

**[Fassungsdatum „7. September 2026" existiert vierfach ohne Kopplung]** (boundary-consistency, principles.md → DRY)
Datei: website/src/lib/anfrage-mailto.ts:77-91
Warum: Nach einer TERMS-Revision verschickt die Website weiter die Annahme einer überholten Fassung.
Vorschlag: `termsFassung` einmal in `links.ts` exportieren, Quelle `TERMS.md:3` im Kommentar nennen, Test darauf beziehen.
Aufwand: S · Status: unverifiziert (minor)

### website/astro.config.mjs

**[Sidebar-Slugs duplizieren `publishedDocs` ohne Prüfung]** (test-coverage, principles.md → Single Source of Truth)
Datei: website/astro.config.mjs:54-142
Warum: Ein neues Dokument ohne Sidebar-Eintrag baut fehlerfrei und ist aus der Navigation unerreichbar.
Vorschlag: Vitest-Fall ergänzen, der die Sidebar-Slugs flach zieht und Mengengleichheit mit `publishedDocs` prüft.
Aufwand: S · Status: unverifiziert (minor)

### website/src/components/Download.astro

**[Kommentar begründet die Copy mit einem fehlenden Windows-Paket]** (docs-accuracy, AGENTS.md Regel 18)
Datei: website/src/components/Download.astro:9-13
Warum: Der Doppelklick-Starter existiert (`windows/starter/`, `packaging/windows/KURZANLEITUNG.md:23`), die Begründung ist heute falsch.
Vorschlag: Kommentar auf den Ist-Zustand umschreiben: ZIP enthält `jotti-start.exe`, Docker Desktop bleibt Voraussetzung.
Aufwand: S · Status: unverifiziert (minor)

### website/src/lib/live-demo.ts

**[Domänentypen englisch (`DemoProduct`, `Cart`, `cartTotalCents`)]** (convention, docs/language.md, AGENTS.md Regel 6)
Datei: website/src/lib/live-demo.ts:13-59
Warum: Das Schwestermodul `anfrage-mailto.ts` im selben Ordner ist durchgängig deutsch benannt.
Vorschlag: Auf Glossarbegriffe umbenennen (`DemoProdukt`, `DemoVariante`, `Warenkorb`, `warenkorbSummeCents`) und LiveDemo.tsx plus Test nachziehen.
Aufwand: M · Status: unverifiziert (minor)

### website/public/theme-init.js

**[Kommentarblock englisch, die zwei Schwesterdateien deutsch]** (convention, AGENTS.md Regel 6)
Datei: website/public/theme-init.js:1-14
Warum: Ein Mechanismus ist über drei Dateien in zwei Sprachen dokumentiert, ohne Regel für den Unterschied.
Vorschlag: Header ins Deutsche übersetzen, passend zu ThemeProvider.astro:6-8 und ThemeToggle.tsx:4-9.
Aufwand: S · Status: unverifiziert (minor)

### website/pnpm-workspace.yaml

**[`minimumReleaseAgeExclude: astro@6.4.8` betrifft eine nicht genutzte Version]** (convention, Version consistency)
Datei: website/pnpm-workspace.yaml:4-5
Warum: package.json fordert `^7.2.9`, aufgelöst wird 7.3.1 — die Ausnahme greift nie mehr.
Vorschlag: Eintrag löschen oder auf die tatsächlich betroffene Astro-Version zeigen lassen.
Aufwand: S · Status: unverifiziert (minor)

### e2e/website/csp-check.mjs

**[Header nennt Zweck und Plan, die es nicht mehr gibt]** (docs-accuracy, AGENTS.md Regel 18)
Datei: e2e/website/csp-check.mjs:1-12
Warum: `docs/plans/plan-website-redesign.md` existiert nicht; der genannte Zweck („jede Phase des Redesigns") ist erledigt.
Vorschlag: Zweck im Ist-Zustand formulieren (prüft den Build gegen die Produktiv-CSP aus `reverse-proxy/nginx.rocks.conf`) und Planlink plus Phasenwortwahl streichen.
Aufwand: S · Status: bestätigt

### e2e/website/csp-server.mjs

**[Konsumentenliste verweist auf den gelöschten Redesign-Plan]** (docs-accuracy, AGENTS.md Regel 18)
Datei: e2e/website/csp-server.mjs:7-9
Warum: Der Pfad lässt sich nicht auflösen und „Phase-9" benennt einen Ablauf, den das Repo nicht kennt.
Vorschlag: Planlink und „Phase-9" streichen, die zwei echten Konsumenten stehen lassen: `csp-check.mjs` und der OG-Bild-Modus von `screenshots.mjs`.
Aufwand: S · Status: bestätigt

**[Traversal-Schutz vergleicht ohne Pfadtrenner]** (security, Pfad-Normalisierung)
Datei: e2e/website/csp-server.mjs:58-76
Warum: `decodeURIComponent` löst `%2f` nach der URL-Normalisierung auf, `startsWith(root)` lässt Geschwisterverzeichnisse durch.
Vorschlag: Gegen `root + sep` prüfen (`filePath !== root && !filePath.startsWith(rootPrefix)`) und denselben Test auf den Fallback-Pfad in Zeile 73 anwenden.
Aufwand: S · Status: unverifiziert (minor)

**[Kommentarsprache englisch, `screenshots.mjs` im selben Ordner deutsch]** (convention, AGENTS.md Regel 6)
Datei: e2e/website/csp-server.mjs:1-11
Warum: Keine Regel trennt diese Dateien, die nächste Tooling-Datei hat kein Vorbild.
Vorschlag: `e2e/website/*.mjs` auf eine Sprache festlegen — Deutsch passt zur übrigen E2E-Suite — und die drei englischen Header umstellen.
Aufwand: S · Status: unverifiziert (minor)

### e2e/website/screenshots.mjs

**[Header behauptet ein `data-theme`-Attribut, das das Frontend nicht setzt]** (docs-accuracy, AGENTS.md Regel 11/18)
Datei: e2e/website/screenshots.mjs:12-15
Warum: `theme-provider.tsx:52-60` schaltet die Klassen `light`/`dark`; `data-theme` kommt in `frontend/src` nicht vor.
Vorschlag: Mechanismus richtig beschreiben: `emulateMedia({ colorScheme })` setzt `prefers-color-scheme`, der ThemeProvider mappt das auf die `dark`/`light`-Klasse; Phasen-Zusatz streichen.
Aufwand: S · Status: bestätigt

**[Kommentar behauptet gleiche BASE-URL-Defaults wie die Suite]** (docs-accuracy, code-smells.md → Unnecessary Comments)
Datei: e2e/website/screenshots.mjs:17-20
Warum: Das Skript defaultet auf `http://localhost:8080`, die Suite auf `http://localhost`.
Vorschlag: Beide Defaults als das benennen, was sie sind, oder auf einen Default vereinheitlichen.
Aufwand: S · Status: unverifiziert (minor)

**[`zeigeAlleAn` ist inline nachgebaut statt importiert]** (principle, principles.md → DRY)
Datei: e2e/website/screenshots.mjs:122-123
Warum: Das Wissen über die „Von anderen"-Gruppe liegt doppelt, obwohl sechs Helfer aus demselben Modul importiert werden.
Vorschlag: `zeigeAlleAn` aus `support/servicekraft.ts` exportieren und hier aufrufen.
Aufwand: S · Status: unverifiziert (minor)

**[„der neue Hero" hat keinen Vorgänger mehr im Baum]** (convention, AGENTS.md Regel 18)
Datei: e2e/website/screenshots.mjs:9
Warum: „neu" ist relativ zu einem Zustand, den es nicht mehr gibt, und veraltet mit der nächsten Änderung.
Vorschlag: „den Hero" schreiben; auch Zeile 222 entsprechend absolut formulieren.
Aufwand: S · Status: unverifiziert (minor)

### e2e/website/browser.mjs

**[Catch-Block verwirft den Startfehler und probiert einen CI-Pfad]** (ops, code-smells.md → Defensive Defaults)
Datei: e2e/website/browser.mjs:10-19
Warum: Jeder Startfehler erscheint als Fehler von `/opt/pw-browsers/chromium-1194/...`, einem Pfad, den kein Entwicklerrechner hat.
Vorschlag: Fallback streichen (`CHROMIUM_EXECUTABLE` plus `playwright install chromium` decken den Fall ab) oder mindestens den Originalfehler als `cause` weiterreichen.
Aufwand: S · Status: unverifiziert (minor)

### e2e/tests/admin-kontrast-axe.spec.ts

**[Begründung besteht aus IDs eines verschwundenen Audits]** (docs-accuracy, AGENTS.md Regel 18)
Datei: e2e/tests/admin-kontrast-axe.spec.ts:8-20
Warum: „Phase 8" und „Muster 05" lassen sich repo-weit nicht auflösen.
Vorschlag: IDs durch die geprüfte Invariante ersetzen: welche Screens gegatet sind und welche Elemente (Grün-Aktionen, Outline-Buttons, WarnKarten) in beiden Themes messbar sein müssen.
Aufwand: S · Status: unverifiziert (minor)

**[axe-Lauf und Auswertung sind zweimal wortgleich vorhanden]** (principle, principles.md → DRY)
Datei: e2e/tests/admin-kontrast-axe.spec.ts:128-138
Warum: Regelauswahl und Fehlerformat müssen an zwei Stellen gepflegt werden und können auseinanderlaufen.
Vorschlag: Eine Funktion `erwarteKeineKontrastVerstoesse(page, kontext, include?)` extrahieren und aus beiden Prüfern aufrufen.
Aufwand: S · Status: unverifiziert (minor)

**[`waitForLoadState('networkidle')` als Bereitschaftssignal, fünfmal]** (test-quality, Playwright-Doku rät davon ab)
Datei: e2e/tests/admin-kontrast-axe.spec.ts:118-126
Warum: Der Zustand ist ein Zeitproxy; axe kann vor dem Mount der Warn-Inhalte messen und meldet dann eine Scheinregression.
Vorschlag: Je Screen auf das gemessene Element warten, z. B. `await expect(page.getByRole('button', { name: 'Als erledigt markieren' })).toBeVisible()`.
Aufwand: M · Status: unverifiziert (minor)

**[`isVisible().catch(() => false)` prüft ein frisch geöffnetes Radix-Menü ohne Warten]** (test-quality, Auto-Waiting)
Datei: e2e/tests/admin-kontrast-axe.spec.ts:92-103
Warum: Ein Menü mit Löschen-Eintrag wird übersprungen, wenn das Portal noch nicht committet ist.
Vorschlag: Vor der Prüfung `await page.getByRole('menu').waitFor()` oder `expect(loeschen).toBeVisible({ timeout: 1000 })` in try/catch.
Aufwand: S · Status: unverifiziert (minor)

### e2e/tests/produktliste-sticky-split.spec.ts

**[19 Zeilen Kommentar beschreiben den Zustand vor dem Fix]** (convention, AGENTS.md Regel 18)
Datei: e2e/tests/produktliste-sticky-split.spec.ts:8-26
Warum: Die alten Pixelwerte verrotten und verdecken, was die Spec heute zusichert.
Vorschlag: Nur die Invariante behalten (Sticky-Leiste bündig, kein horizontaler Überlauf, geprüft bei 1024px und 1280px); gleiche Kürzung in admin-finanzamt-einrichtung.spec.ts:8-13 und den „nicht mehr"-Sätzen der Fehlerpfad-Specs.
Aufwand: S · Status: unverifiziert (minor)

### e2e/tests/drawer-tableiste-inert.mobile.spec.ts

**[Specs begründen sich mit `NEU13`, `NEU14`, `Befund #1`]** (convention, AGENTS.md Regel 18)
Datei: e2e/tests/drawer-tableiste-inert.mobile.spec.ts:7
Warum: Die IDs matchen nur in diesen Specs, kein Dokument definiert sie.
Vorschlag: Jede ID durch die bewachte Invariante ersetzen, z. B. „Bei offenem Service-Drawer ist die Tab-Leiste weder per Zeiger noch per Tastatur erreichbar." Auch admin-theme-umschalter.spec.ts:6, admin-finanzamt-einrichtung.spec.ts:8, tischservice-viewport-ueberlauf.mobile.spec.ts:12, tischservice-lange-bestellung.mobile.spec.ts:19.
Aufwand: S · Status: unverifiziert (minor)

### e2e/tests/admin-theme-umschalter.spec.ts

**[Assertion sucht den Button erneut per Text-Regex statt über den Locator]** (readability, readability.md → Clever Code)
Datei: e2e/tests/admin-theme-umschalter.spec.ts:36-49
Warum: Bei geänderter Beschriftung liefert die Query still `null`, während der Locator laut fehlschlagen würde.
Vorschlag: `toggle.evaluate(el => ({ dataActive: el.getAttribute('data-active'), bg: getComputedStyle(el).backgroundColor, focused: document.activeElement === el }))` verwenden.
Aufwand: S · Status: unverifiziert (minor)

### e2e/tests/kassenabschluss.mobile.spec.ts

**[Spec klickt den vom Seed erzeugten Zustand per UI zurück]** (large-refactor, architecture.md)
Datei: e2e/tests/kassenabschluss.mobile.spec.ts:32-53
Warum: 120s Timeout und rund 20 Tische Klickarbeit stehen einer einzigen Assertion gegenüber.
Vorschlag: Zum Einplanen markieren: `POST /api/test/reset-and-seed` um eine Variante mit ausgeglichenen Tischen erweitern; nicht im Cleanup-Durchgang anfassen.
Aufwand: L · Status: unverifiziert (minor)

### e2e/tests/bestellen-kassieren.spec.ts

**[`[data-slot="tisch-saldo"]` steht in vier Specs plus Helfer]** (principle, principles.md → DRY)
Datei: e2e/tests/bestellen-kassieren.spec.ts:60-64
Warum: Die Kopplung an ein Frontend-Attribut liegt fünffach, ein Rename in TablePage.tsx trifft alle fünf.
Vorschlag: `tischSaldo(page): Locator` einmal in `support/servicekraft.ts` exportieren und in den Specs verwenden.
Aufwand: S · Status: unverifiziert (minor)

### e2e/tests/umbuchung.mobile.spec.ts

**[Kommentar schreibt dem `.mobile`-Suffix eine Routing-Wirkung zu]** (convention, playwright.config.ts:44,52)
Datei: e2e/tests/umbuchung.mobile.spec.ts:15
Warum: Die Projekte selektieren allein über das `admin-`-Präfix; das Suffix tragen nur etwa die Hälfte der Handy-Specs.
Vorschlag: Die echte Regel notieren und das Suffix entweder konsequent vergeben oder ganz entfernen.
Aufwand: M · Status: unverifiziert (minor)

### e2e/support/servicekraft.ts

**[`waehleAlleVollAus` bricht nach 50 Klicks ohne Nachbedingung ab]** (test-quality, stille Fehlschlag-Unterdrückung)
Datei: e2e/support/servicekraft.ts:253-265
Warum: Die Funktion kehrt scheinbar erfolgreich zurück, der Fehler erscheint später als „Es gibt noch offene Tische".
Vorschlag: Nach der Schleife prüfen, dass die Zeile voll ausgewählt ist, z. B. `await expect(zeile).toContainText(/(\d+) von \1 ausgewählt/)`, sonst mit Zeilentext scheitern.
Aufwand: S · Status: unverifiziert (minor)

**[Dateikommentar behauptet „ausschließlich zugängliche Selektoren"]** (docs-accuracy, code-smells.md → Unnecessary Comments)
Datei: e2e/support/servicekraft.ts:4-6
Warum: Zeile 242 und 274 nutzen `[data-slot="item"]` bzw. `[data-slot="tisch-saldo"]`.
Vorschlag: Regel wie gelebt formulieren: zugängliche Selektoren zuerst, `data-slot` nur ohne zugänglichen Anker; „seit dem Redesign" in Zeile 31 streichen.
Aufwand: S · Status: unverifiziert (minor)

**[`settleAlleOffenenTische` mischt englisches Verb mit deutscher Domäne]** (convention, AGENTS.md Regel 6, docs/language.md)
Datei: e2e/support/servicekraft.ts:285
Warum: Alle Schwester-Exporte der Datei sind vollständig deutsch benannt.
Vorschlag: In `gleicheAlleOffenenTischeAus` umbenennen und den Aufruf in kassenabschluss.mobile.spec.ts:10,53 nachziehen.
Aufwand: S · Status: unverifiziert (minor)

### e2e/support/viewport.ts

**[`?? 0` macht eine fehlgeschlagene Messung zur bestandenen Zusicherung]** (correctness, Assertion-Härte)
Datei: e2e/support/viewport.ts:11-18
Warum: `0 <= innerWidth` ist immer wahr, die Überlaufprüfung kann ohne Messung grün werden.
Vorschlag: Rohwert zurückgeben und vor dem Vergleich `expect(scrollWidth).toBeGreaterThan(0)` prüfen.
Aufwand: S · Status: unverifiziert (minor)

### e2e/helpers/fehlerpfade.ts

**[Kommentar verspricht einen POST-Filter, den der Code nicht hat]** (docs-accuracy, code-smells.md → Unnecessary Comments)
Datei: e2e/helpers/fehlerpfade.ts:3-6
Warum: Beide Handler filtern nur über `url.pathname` und lesen `route.request().method()` nie.
Vorschlag: „ausschließlich POST" streichen und auf `/api/<endpoint>` abstellen (die API ist laut Regel 1 POST-only) oder die Methodenprüfung ergänzen.
Aufwand: S · Status: unverifiziert (minor)

**[`simuliereServerfehler` und `simuliereNetzabbruch` sind bis auf eine Zeile identisch]** (principle, principles.md → DRY)
Datei: e2e/helpers/fehlerpfade.ts:33-62
Warum: Route-Glob und Matching-Prädikat liegen doppelt; eine Korrektur kann die zweite Kopie verfehlen.
Vorschlag: `fangeEndpunkteAb(page, endpoints, handler)` extrahieren und beide Exporte nur den Handler übergeben lassen.
Aufwand: S · Status: unverifiziert (minor)

### e2e/package.json

**[TypeScript 7.0.2 gegen `~6.0.3` in Frontend und Website]** (convention, Version consistency, AGENTS.md Tech-Stack)
Datei: e2e/package.json:17-22
Warum: Das Typecheck-Gate läuft mit einem anderen Compiler-Major als der getippte Code, und AGENTS.md nennt 6.0.
Vorschlag: Auf `~6.0.3` angleichen oder TS 7 in der Tech-Stack-Tabelle dokumentieren; Pinning-Stil im Block vereinheitlichen.
Aufwand: S · Status: unverifiziert (minor)

**[`packageManager` ohne Integritäts-Hash]** (convention, Supply-Chain)
Datei: e2e/package.json:10
Warum: Corepack prüft das pnpm-Tarball nur mit Hash; Frontend und Website haben ihn, e2e nicht.
Vorschlag: Den `pnpm@11.6.0+sha512…`-String aus frontend/package.json:76 wörtlich übernehmen.
Aufwand: S · Status: unverifiziert (minor)

### e2e/tsconfig.json

**[`helpers/` und `support/` erfüllen dieselbe Rolle ohne Regel]** (architecture, architecture.md → Modul-Layout)
Datei: e2e/tsconfig.json:17
Warum: `helpers/` enthält eine einzige Datei, Specs importieren geteilte Helfer aus zwei Pfaden.
Vorschlag: `helpers/fehlerpfade.ts` nach `support/` verschieben, den `helpers/**`-Glob entfernen und die vier importierenden Specs anpassen.
Aufwand: S · Status: unverifiziert (minor)

**[`e2e/website/*.mjs` liegt außerhalb des Typecheck-Gates]** (test-quality, CI-Abdeckung)
Datei: e2e/tsconfig.json:17
Warum: `screenshots.mjs` importiert die TS-Helfer direkt, eine Signaturänderung bricht es unbemerkt.
Vorschlag: `website/**/*.mjs` mit `allowJs`/`checkJs` in die e2e-tsconfig aufnehmen oder im Skriptheader festhalten, dass nur `make website-screenshots` es prüft.
Aufwand: S · Status: unverifiziert (minor)

## Windows, Edge und Ops

Umfang: Windows-Starter und -Relay, Release-Packaging (.cmd, Kurzanleitung), Reverse-Proxy und Resolver, Compose-Stacks, Makefile, CI und Ops-Skripte.

Befunde: 0 blocker · 14 major · 56 minor (326 Dateien, 101 Rohbefunde zusammengeführt).

### packaging/windows/jotti-restore.cmd

**[Restore startet die gescheiterte Migration erneut]** (correctness, AGENTS.md Bewertungsmetrik Korrektheit)
Datei: packaging/windows/jotti-restore.cmd:31-37
Warum: Der Dump setzt `schema_migrations` zurück; `up -d` fährt dieselbe defekte Migration erneut und der Stack bleibt unten.
Vorschlag: Restore-Erfolg getrennt melden; bei fehlschlagendem `up -d` auf das vorherige Release-ZIP verweisen und KURZANLEITUNG.md:118-121 angleichen.
Aufwand: M · Status: bestätigt

**[Recovery-Skripte starten den Stack ohne LAN_IP]** (correctness, docker-compose.release.yml:139-142)
Datei: packaging/windows/jotti-restore.cmd:36-37, packaging/windows/jotti-repair.cmd:15,46
Warum: `LAN_IP` interpoliert leer, der Reverse-Proxy wird neu erzeugt und fällt auf die Docker-Bridge-IP (172.x) zurück.
Vorschlag: Beide Skripte nach der Datenbankarbeit beenden und auf `jotti-start.exe` verweisen; alternativ LAN_IP mitgeben oder `--no-recreate` nutzen.
Aufwand: M · Status: bestätigt

**[Em-Dash und LF-Zeilenenden in den Batch-Dateien]** (convention, Windows-Konsolenkonvention)
Datei: packaging/windows/jotti-repair.cmd:51 (alle drei .cmd)
Warum: cmd.exe liest OEM-Codepage; zusätzlich sind alle .cmd-Dateien LF-only, ohne normalisierende .gitattributes.
Vorschlag: Em-Dash in der `echo`-Zeile durch `-` ersetzen und `*.cmd text eol=crlf` in eine Root-.gitattributes aufnehmen.
Aufwand: S · Status: unverifiziert (minor)

### packaging/windows/KURZANLEITUNG.md

**[Manuelles Backup schreibt in einen nicht existierenden Ordner]** (docs-accuracy, windows/starter/backup.go:169)
Datei: packaging/windows/KURZANLEITUNG.md:83-91
Warum: `%PROGRAMDATA%\jotti\backups` entsteht erst beim ersten Pre-Update-Backup; ohne Update bricht die Umleitung ab.
Vorschlag: `md "%PROGRAMDATA%\jotti\backups" 2>nul` voranstellen und auf eine nicht leere Datei prüfen.
Aufwand: S · Status: bestätigt

**[Historische Formulierungen in der ausgelieferten Kurzanleitung]** (docs-accuracy, AGENTS.md Regel 18)
Datei: packaging/windows/KURZANLEITUNG.md:62-65, 136-141
Warum: „vor diesem Update" und „lag früher im Programmordner" sind für den Leser der aktuellen Fassung nicht prüfbar.
Vorschlag: Bedingung im Ist-Zustand beschreiben: Warnung bei Volumes ohne Compose-Label; .env „bei älteren Installationen im Programmordner".
Aufwand: S · Status: unverifiziert (minor)

**[Subjekt und Objekt vertauscht]** (docs-accuracy, reverse-proxy/statuspage.go:24)
Datei: packaging/windows/KURZANLEITUNG.md:44-46
Warum: Nicht die Router-Anleitung verlinkt die Status-Seite, sondern die Status-Seite die Router-Anleitung.
Vorschlag: „Die Status-Seite verlinkt die Router-Anleitung; sie steht auch online unter …".
Aufwand: S · Status: unverifiziert (minor)

### windows/starter/backup.go

**[Nicht-ASCII-Zeichen in Windows-Konsolenausgaben]** (convention, core/diagnose.go:16-18)
Datei: windows/starter/backup.go:67,69; windows/starter/main.go:238; windows/starter/system.go:58,114,251; windows/relay/env.go:84; windows/relay/main.go:148
Warum: Die dokumentierte ASCII-Transliteration ist nur halb umgesetzt; Em-Dash, Pfeil und „ü" erscheinen als Mojibake.
Vorschlag: In gedruckten Strings „—" durch „-", „→" durch „->" und „ü" durch „ue" ersetzen; Quellkommentare unverändert lassen.
Aufwand: S · Status: bestätigt

### windows/starter/main.go

**[Erfolgsmeldung behauptet ungeprüfte Firewall-Freigabe]** (correctness, system.go:225-233)
Datei: windows/starter/main.go:241-243
Warum: `ensureFirewall` schluckt Fehler mit einer Warnung, `printSuccess` meldet die Freigabe trotzdem als eingerichtet.
Vorschlag: `ensureFirewall` einen Bool zurückgeben lassen und die Zeile nur bei Erfolg drucken.
Aufwand: S · Status: unverifiziert (minor)

### windows/starter/system.go

**[Preflight prüft UDP 443 nicht]** (correctness, docker-compose.release.yml:146-149)
Datei: windows/starter/system.go:112-126
Warum: Der Stack veröffentlicht `443:443/udp`; ein UDP-Konflikt passiert den Preflight und endet im rohen Docker-Bind-Fehler.
Vorschlag: `net.ListenPacket("udp", ":443")` ergänzen oder im Kommentar begründen, warum UDP bewusst ausgenommen bleibt.
Aufwand: S · Status: unverifiziert (minor)

**[Zwei Compose-Aufrufe ohne --env-file]** (convention, system.go:450-455)
Datei: windows/starter/system.go:128-146
Warum: Ohne `--env-file` warnt Compose je Variable, obwohl der Start gesund ist — genau das Rauschen, das der Starter vermeidet.
Vorschlag: Beide Helfer über eine lesende Variante von `runCompose` mit `--env-file` führen oder stderr abfangen.
Aufwand: S · Status: unverifiziert (minor)

**[Exportierter Name in package main]** (convention, readability.md Naming)
Datei: windows/starter/system.go:18-20
Warum: `DockerCliPath` signalisiert eine Paketgrenze, die es nicht gibt, und trennt die Konstante von ihren Geschwistern in `core`.
Vorschlag: In `dockerCliPath` umbenennen oder neben `DockerBinPath`/`DockerDesktopPath` nach core/diagnose.go verschieben.
Aufwand: S · Status: unverifiziert (minor)

### windows/starter/core/env.go

**[Verbotene Formulierung „wie bisher"]** (docs-accuracy, AGENTS.md Regel 18)
Datei: windows/starter/core/env.go:14,88; windows/starter/main.go:89
Warum: Der Zusatz beschreibt einen früheren Stand und trägt zum aktuellen Verhalten nichts bei.
Vorschlag: Die drei Zusätze streichen; die Sätze sind ohne sie vollständig.
Aufwand: S · Status: unverifiziert (minor)

### windows/starter/core/backup.go

**[Backup-Regel doppelt dokumentiert, mit historischer Begründung]** (docs-accuracy, principles.md DRY)
Datei: windows/starter/core/backup.go:13-21; windows/starter/backup.go:47-55
Warum: Zwei vollständige Kopien einer Bedingung driften auseinander, und „vor Einfuehrung des automatischen Pre-Update-Backups" datiert ein Release.
Vorschlag: Regel einmal in `ShouldBackup` ohne historische Klausel formulieren; der Aufrufer verweist per `siehe core.ShouldBackup`.
Aufwand: S · Status: unverifiziert (minor)

### windows/starter/core/diagnose.go

**[„frueher" in einer nutzersichtbaren Meldung]** (docs-accuracy, AGENTS.md Regel 18)
Datei: windows/starter/core/diagnose.go:36-40
Warum: Der Ort gilt heute noch für Altinstallationen; nur die historische Rahmung ist verboten.
Vorschlag: „(bei aelteren Installationen im Programmordner neben jotti-start.exe)" schreiben.
Aufwand: S · Status: unverifiziert (minor)

### windows/starter/rsrc_windows_amd64.syso

**[Generierte Ressource passt nicht zum Manifest]** (docs-accuracy, AGENTS.md Regel 18)
Datei: windows/starter/rsrc_windows_amd64.syso:1-25
Warum: Das eingebettete Manifest nennt einen gelöschten Plan-Pfad; das Artefakt reproduziert nicht mehr aus seiner Quelle.
Vorschlag: `make starter-syso` ausführen und die neu erzeugte .syso committen.
Aufwand: S · Status: unverifiziert (minor)

### windows/relay/main.go

**[Statusabfrage ohne Write-Deadline]** (code-smell, main.go:402-405)
Datei: windows/relay/main.go:361-372
Warum: Jeder andere I/O-Schritt des Relays ist zeitbegrenzt; diese eine Schreiboperation bricht die Invariante auf dem Papier.
Vorschlag: Vor dem DLE-EOT-Write `conn.SetWriteDeadline(time.Now().Add(writeTimeout))` setzen.
Aufwand: S · Status: unverifiziert (minor)

### windows/relay/main_test.go

**[Gemischte Sprachen in Testnamen]** (test-quality, readability.md Naming)
Datei: windows/relay/main_test.go:74-209
Warum: Fallnamen und Assertion-Meldungen wechseln innerhalb einer Datei zwischen Deutsch und Englisch.
Vorschlag: Auf Deutsch vereinheitlichen, passend zu `TestPruefeRelayStatus` und der starter/core-Suite.
Aufwand: S · Status: unverifiziert (minor)

### reverse-proxy/caddyfile.go

**[acme-dns-Subdomain ungequotet in der Caddyfile-Site-Adresse]** (security, state.go:21-23)
Datei: reverse-proxy/caddyfile.go:201-221
Warum: `InstallState.valid()` prüft nur auf nicht leer; ein bösartiger acme-dns-Endpunkt kann Caddy-Direktiven injizieren.
Vorschlag: Subdomain in `valid()` gegen `^[a-z0-9-]{1,63}$` prüfen; Quoting allein genügt für ein Site-Adress-Token nicht.
Aufwand: S · Status: bestätigt

**[CSP zweimal wörtlich gepflegt, ohne Guard]** (principle, principles.md DRY)
Datei: reverse-proxy/caddyfile.go:5-8; reverse-proxy/nginx.rocks.conf:142-143
Warum: Nur zwei Prosakommentare halten Demo und Selbsthosting-Installationen synchron; ein Edit driftet still auseinander.
Vorschlag: Test in package main ergänzen, der nginx.rocks.conf liest und `strings.Contains(conf, contentSecurityPolicy)` prüft; die Kommentare ersetzen.
Aufwand: S · Status: bestätigt

**[Historische und tote Verweise in den Kommentaren]** (convention, AGENTS.md Regel 18)
Datei: reverse-proxy/caddyfile.go:6,14,51,69,92
Warum: „früheres nginx-Setup", „prod-nginx" und „Option-2-Ersatz" verweisen auf nichts, was im Repo auffindbar ist.
Vorschlag: Regeln im Ist-Zustand formulieren und als Quelle der Zahlen nginx.rocks.conf:22/147 nennen; Zeile 69 wird zu „halten die Browserwarnung einmalig".
Aufwand: S · Status: unverifiziert (minor)

**[HSTS-Header über Klartext-HTTP]** (correctness, RFC 6797 §7.2)
Datei: reverse-proxy/caddyfile.go:182-196
Warum: RFC 6797 verbietet den Header über unsicheren Transport; der Test zementiert den Verstoß als Sollverhalten.
Vorschlag: `proxySnippet` bei leerem HSTS-Wert die Zeile auslassen lassen und caddyfile_test.go:175 auf Abwesenheit umstellen.
Aufwand: S · Status: unverifiziert (minor)

### reverse-proxy/main.go

**[Leeres JOTTI_DOMAIN startet den Public-Stack im LAN-Modus]** (correctness, docker-compose.prod.yml:140)
Datei: reverse-proxy/main.go:78-86
Warum: Ein VPS liefert dann Zertifikate der internen CA für jede SNI, während der HTTP-Healthcheck weiter grün ist.
Vorschlag: `JOTTI_DOMAIN: ${JOTTI_DOMAIN:?…}` in docker-compose.prod.yml setzen oder LAN-Modus ohne State-Verzeichnis verweigern.
Aufwand: S · Status: bestätigt

**[Status-Seite verspricht Selbstheilung, die es nicht gibt]** (correctness, status.go:55)
Datei: reverse-proxy/main.go:134-158
Warum: `ensureState` läuft nur beim Start; die Seite lädt sich endlos neu und wird nie grün.
Vorschlag: `ensureState` im Hintergrund wiederholen und nach Erfolg neu rendern, oder den Hinweistext auf „jotti neu starten" ändern.
Aufwand: M · Status: bestätigt

**[Caddyfile mit acme-dns-Zugangsdaten als 0o644]** (security, state.go:66)
Datei: reverse-proxy/main.go:111,125,171-173
Warum: Dieselben Zugangsdaten liegen in install.json bewusst als 0o600; der Pfad ist zudem über `PROXY_CADDYFILE_PATH` konfigurierbar.
Vorschlag: Alle drei Schreibaufrufe auf 0o600 umstellen.
Aufwand: S · Status: unverifiziert (minor)

**[Dreifach kopierte Caddyfile-Schreibsequenz]** (principle, principles.md DRY)
Datei: reverse-proxy/main.go:105-130, 164-173
Warum: Pfad, Modus und Fatal-Text stehen dreimal identisch; das Schwesterschritt-Muster `runCaddyOrExit` existiert bereits.
Vorschlag: `writeCaddyfileOrExit(cfg, caddyfile)` extrahieren und aus allen drei Modi aufrufen.
Aufwand: S · Status: unverifiziert (minor)

**[Paketkommentar nennt nur zwei Modi]** (docs-accuracy, main.go:72-76)
Datei: reverse-proxy/main.go:1-14
Warum: Der zuerst geprüfte HTTP-only-Modus fehlt in Paketkommentar und Dockerfile-Header.
Vorschlag: Absatz zu `PROXY_HTTP_ONLY` ergänzen (Klartext-HTTP auf :80, kein TLS/ACME) und reverse-proxy/Dockerfile:8-11 angleichen.
Aufwand: S · Status: unverifiziert (minor)

**[loadConfig ohne Test]** (test-quality, resolver/main_test.go:12)
Datei: reverse-proxy/main.go:53-67
Warum: Der Env-Vertrag inklusive Modus-Umschaltung ist ungepinnt, obwohl das Schwestermodul genau das testet.
Vorschlag: Tabellengetriebenen `TestLoadConfig` nach dem Vorbild von resolver/main_test.go ergänzen, inklusive `parseBool`-Varianten.
Aufwand: S · Status: unverifiziert (minor)

### reverse-proxy/probe.go

**[Rebind-Diagnose verwechselt Lookup-Fehler mit Rebind-Schutz]** (correctness, statuspage.go:157-158)
Datei: reverse-proxy/probe.go:53-67
Warum: Ohne Internet behauptet die Seite „DNS-Rebind-Schutz erkannt" und verlinkt eine dann unerreichbare Anleitung.
Vorschlag: Lookup-Fehler getrennt zurückgeben und nur bei fremder Adresse den Rebind-Hinweis zeigen; sonst neutraler Text.
Aufwand: S · Status: unverifiziert (minor)

### reverse-proxy/nginx.rocks.conf

**[Tote Konfiguration: leere docs_redirect-Map und proxy_cache_bypass]** (code-smell, principles.md YAGNI)
Datei: reverse-proxy/nginx.rocks.conf:28-34, 107-110, 118, 155, 165
Warum: Die Map hat nur `default ""`, kein Proxy-Cache und kein Upgrade-Header existieren; beides kann nie wirken.
Vorschlag: Map samt `if`-Block und die drei `proxy_cache_bypass`-Zeilen löschen.
Aufwand: S · Status: unverifiziert (minor)

**[Einziger deutscher Kommentar in einer englischen Config]** (convention, AGENTS.md Regel 6)
Datei: reverse-proxy/nginx.rocks.conf:142
Warum: Die Datei ist sonst durchgängig englisch kommentiert; docs/language.md legt Englisch für Infrastruktur fest.
Vorschlag: Zeile 142 englisch formulieren und auf den CSP-Gleichheitstest verweisen.
Aufwand: S · Status: unverifiziert (minor)

### reverse-proxy/Dockerfile

**[go mod verify fehlt nur in diesem Image]** (convention, backend/Dockerfile:7)
Datei: reverse-proxy/Dockerfile:27-28
Warum: Backend- und Resolver-Image prüfen die Modulintegrität an derselben Stelle; hier fehlt sie ohne Begründung.
Vorschlag: Zeile 28 zu `RUN go mod download && go mod verify` ändern.
Aufwand: S · Status: unverifiziert (minor)

### resolver/resolve.go

**[Nameserver-Name antwortet nur auf A, sonst REFUSED]** (correctness, resolve.go:72-74)
Datei: resolver/resolve.go:58-62
Warum: Derselbe Name ist für A autoritativ und für AAAA refused; kein Kommentar und kein Test pinnt die Absicht.
Vorschlag: Entweder Zweig streichen oder für andere qtypes `kindNoData` liefern; Testfall mit `dns.TypeAAAA` ergänzen.
Aufwand: S · Status: unverifiziert (minor)

### resolver/main.go

**[Opcode wird nicht geprüft]** (correctness, miekg/dns SetReply)
Datei: resolver/main.go:83-97
Warum: Ein RFC-2136-UPDATE erhält eine autoritative NOERROR-Antwort für eine nie ausgeführte Änderung.
Vorschlag: Neben dem FORMERR-Guard `dns.RcodeNotImplemented` für `req.Opcode != dns.OpcodeQuery` zurückgeben.
Aufwand: S · Status: unverifiziert (minor)

### resolver/main_test.go

**[Drei Namenskonventionen in einem Paket]** (test-quality, readability.md Naming)
Datei: resolver/main_test.go:21-92; resolver/resolve_test.go:35,41,53,71,77,101
Warum: Englische, deutsche und ASCII-transliterierte Testnamen stehen nebeneinander; die nächste Ergänzung erbt den zuletzt gelesenen Stil.
Vorschlag: Auf Deutsch mit Umlauten vereinheitlichen — die im Paket bereits dominierende Form; nur Namen ändern.
Aufwand: S · Status: unverifiziert (minor)

### docs/jotti-rocks-infra.md

**[Falsche TTL im Runbook]** (docs-accuracy, resolver/resolve.go:15-17)
Datei: docs/jotti-rocks-infra.md:17-18
Warum: Nur A-Records nutzen 86400; Challenge-CNAMEs laufen mit 3600, negative Antworten mit 300.
Vorschlag: „(A-Records TTL 86400 …; Challenge-CNAMEs TTL 3600)" schreiben.
Aufwand: S · Status: unverifiziert (minor)

### Makefile

**[make lint kann bei unformatiertem Go nie fehlschlagen]** (correctness, AGENTS.md Regel 15)
Datei: Makefile:78-79
Warum: `goimports -l` beendet sich mit 0; das lokal vorgeschriebene Gate ist schwächer als CI und `make check`.
Vorschlag: Die Guard-Form aus `check-backend` übernehmen: `if [ "$$(goimports -l . | wc -l)" -gt 0 ]; then goimports -l .; exit 1; fi`.
Aufwand: S · Status: bestätigt

**[check-tools prüft migrate und Docker nicht]** (ops, Makefile:296-303)
Datei: Makefile:269-276
Warum: `verify` braucht beides; ohne Prüfung stirbt der Lauf tief in test-integration.sh statt mit der freundlichen Meldung.
Vorschlag: `migrate` und `docker` in die Werkzeugliste von `check-tools` aufnehmen.
Aufwand: S · Status: unverifiziert (minor)

**[rocks-reset-db dupliziert das Reset-Skript ohne Rückfrage]** (code-smell, scripts/reset-and-seed.sh:58)
Datei: Makefile:219-222
Warum: Zwei destruktive Pfade auf dasselbe Volume, einer davon ohne Bestätigung und ohne TLS-Hinweis.
Vorschlag: Target löschen oder auf `./scripts/reset-and-seed.sh rocks` umstellen.
Aufwand: S · Status: unverifiziert (minor)

**[Redundante Prerequisites]** (code-smell, Makefile:32-33)
Datei: Makefile:262-263, 303
Warum: `clean` hängt an `down` und ruft danach `down -v`; `verify` listet `check-tools`, das `check` schon zieht.
Vorschlag: `clean` auf die eine Rezeptzeile reduzieren und `verify: check-full` schreiben.
Aufwand: S · Status: unverifiziert (minor)

**[.env.example im Release-ZIP ohne Konsument]** (convention, packaging/windows/KURZANLEITUNG.md)
Datei: Makefile:143
Warum: Der Starter erzeugt seine eigene .env; die englische Vorlage enthält nur Self-Hosting-Schlüssel.
Vorschlag: `cp .env.example "$(RELEASE_DIR)/"` streichen oder eine deutsche Vier-Schlüssel-Vorlage ausliefern.
Aufwand: S · Status: unverifiziert (minor)

**[Fünffach kopierte Go-Modul-Pipeline in Makefile und CI]** (large-refactor, .github/workflows/ci.yml:220-272)
Datei: Makefile:281-291
Warum: Ein zusätzlicher Lint-Schritt oder ein Versions-Bump erfordert acht Edits; windows-ci zeigt die Matrixform bereits.
Vorschlag: Nur markiert — `define`/`foreach` im Makefile plus Matrix-Job in CI separat planen, nicht im Cleanup-Durchgang.
Aufwand: L · Status: unverifiziert (minor)

### scripts/prod-update.sh

**[Fünf Verweise auf das nicht existierende docs/leitfaden.md]** (docs-accuracy, AGENTS.md Regel 18 / keine toten Links)
Datei: scripts/prod-update.sh:111; scripts/prod-backup.sh:132; scripts/ops-smoke.sh:34; Makefile:231; .env.example:16
Warum: Der Leitfaden ist ein Verzeichnis mit 16 Seiten; zwei der Verweise erscheinen zur Laufzeit in Fehlermeldungen.
Vorschlag: Je konkrete Seite verlinken (aktualisieren-backups.md, datenaufbewahrung.md, self-hosting.md, installation.md), danach greppt `leitfaden\.md` leer.
Aufwand: S · Status: bestätigt

### scripts/prod-backup.sh

**[Dumps mit TSE-Secrets und Passwort-Hashes weltlesbar]** (security, scripts/init-env.sh:84)
Datei: scripts/prod-backup.sh:66-94
Warum: Backups laufen als root mit umask 022; jedes lokale Konto liest `tse_konfiguration.api_secret` und alle `password_hash`.
Vorschlag: `umask 077` vor `mkdir -p "$BACKUP_DIR"` setzen oder Verzeichnis 700 und Dumps 600 chmodden.
Aufwand: S · Status: bestätigt

### scripts/prod-restore.sh

**[Destruktiver Restore ohne Integritätsprüfung]** (correctness, scripts/prod-backup.sh:91-93)
Datei: scripts/prod-restore.sh:99-116
Warum: Die Dumps sind `--clean --if-exists`; ein defektes Archiv droppt die Objekte und bricht dann mitten im Stream ab.
Vorschlag: Vor der Bestätigungsabfrage `gzip -t "$SELECTED"` für `*.gz` ausführen und bei Fehler abbrechen.
Aufwand: S · Status: bestätigt

### scripts/test-integration.sh

**[Testskripte nutzen den gleitenden Tag postgres:17]** (convention, scripts/prod-backup-verify.sh:56-59)
Datei: scripts/test-integration.sh:34; scripts/test-tse-live.sh:51
Warum: Alle Compose-Dateien, beide CI-Services und der Starter pinnen `postgres:17.8`; Tests können gegen eine andere Minor laufen.
Vorschlag: Beide `docker run`-Aufrufe auf `postgres:17.8` pinnen oder den Tag wie prod-backup-verify.sh aus der Compose-Datei lesen.
Aufwand: S · Status: bestätigt

### scripts/test-tse-live.sh

**[Doppeltes Postgres-Harness und abweichende Skriptkonventionen]** (code-smell, principles.md DRY)
Datei: scripts/test-tse-live.sh:1-2, 26-68; scripts/test-integration.sh:1, 9-51
Warum: Startrennen-Workaround, Cleanup-Trap und Migrate-Aufruf stehen doppelt; beide Skripte nutzen `#!/bin/bash` und kein lib.sh.
Vorschlag: `#!/usr/bin/env bash` plus lib.sh übernehmen und `start_throwaway_postgres NAME PORT` gemeinsam nutzen.
Aufwand: M · Status: unverifiziert (minor)

### scripts/lib.sh

**[read_env entfernt keine Anführungszeichen]** (correctness, scripts/prod-backup.sh:55-66)
Datei: scripts/lib.sh:25-27
Warum: Compose strippt Quotes, die Skripte nicht; ein gequotetes `BACKUP_DIR` legt ein Verzeichnis mit Anführungszeichen im Namen an.
Vorschlag: Im sed-Ausdruck eine passende Lage einfacher oder doppelter Anführungszeichen entfernen.
Aufwand: S · Status: unverifiziert (minor)

### scripts/ops-smoke.sh

**[Neun kopierte curl-Blöcke]** (code-smell, principles.md DRY)
Datei: scripts/ops-smoke.sh:264-380
Warum: `http_post_status` kann keinen Authorization-Header tragen, deshalb wiederholt jeder Schritt sieben Zeilen Boilerplate.
Vorschlag: Optionalen dritten Parameter für den Auth-Header ergänzen und einen `post_step`-Wrapper mit Zeitmessung bauen.
Aufwand: M · Status: unverifiziert (minor)

**[Prozess-Notiz im Dateikopf]** (docs-accuracy, AGENTS.md Regel 18)
Datei: scripts/ops-smoke.sh:10-11
Warum: „No real run happens in this phase" beschreibt den Autorenmoment und suggeriert, das Skript sei nicht zum Ausführen gedacht.
Vorschlag: Beide Sätze löschen; der Kopf beschreibt die Modi bereits vollständig.
Aufwand: S · Status: unverifiziert (minor)

### scripts/prod-init.sh

**[Preflight, parse_semver und BACKUP_DIR mehrfach kopiert]** (code-smell, principles.md DRY)
Datei: scripts/prod-init.sh:27-36, 73-91 (plus prod-update, prod-backup, prod-restore, prod-backup-verify, ops-smoke)
Warum: Sechs Kopien der Host-Prüfung und zwei identische `parse_semver` garantieren Drift bei der nächsten Änderung.
Vorschlag: `parse_semver`, `require_docker_stack` und `resolve_backup_dir` nach scripts/lib.sh ziehen.
Aufwand: M · Status: unverifiziert (minor)

### scripts/prod-backup-verify.sh

**[Dump-Auswahl und decompress doppelt gepflegt]** (code-smell, scripts/prod-restore.sh:55-70)
Datei: scripts/prod-backup-verify.sh:64-79, 116-122
Warum: Restore und Verify müssen sich über „neuester Dump gewinnt" einig sein; zwei Kopien können abweichen.
Vorschlag: `select_dump` und `decompress` nach scripts/lib.sh verschieben und aus beiden Skripten aufrufen.
Aufwand: S · Status: unverifiziert (minor)

### scripts/reset-and-seed.sh

**[Mehrstack-Maschinerie für genau einen Stack]** (principle, principles.md YAGNI)
Datei: scripts/reset-and-seed.sh:34-73
Warum: Argument-Parser, Array-Schleife und drei Indirektionen bedienen eine feste Konfiguration.
Vorschlag: Auf `COMPOSE_FILE="docker-compose.rocks.yml"`, direkte Dateiprüfung und `--yes`/`-h` reduzieren; Konstanten inlinen.
Aufwand: S · Status: unverifiziert (minor)

### scripts/rocks-init.sh

**[DNS-Probe und Statusprüfung je mehrfach kopiert]** (code-smell, scripts/prod-init.sh:126-132)
Datei: scripts/rocks-init.sh:82-113, 159-192
Warum: Vier gleiche `host … || dig …`-Proben und drei gleiche curl-Blöcke; prod-init.sh hat dafür längst `resolve_a`.
Vorschlag: `resolve_a` nach lib.sh heben, über die Domains schleifen und `check_http URL LABEL` ergänzen.
Aufwand: M · Status: unverifiziert (minor)

### scripts/setup-dev-tools.sh

**[pnpm als einziges Werkzeug ungepinnt]** (convention, .github/workflows/ci.yml:286)
Datei: scripts/setup-dev-tools.sh:23, 129-143
Warum: CI und alle package.json pinnen `11.6.0`; das Skript sagt nur `11`, obwohl es Drift genau verhindern soll.
Vorschlag: `PNPM_VERSION="11.6.0"` einführen und in Kommentar, corepack-Aufruf und beiden Meldungen verwenden.
Aufwand: S · Status: unverifiziert (minor)

### scripts/generate-spektral-logos.py

**[Docstring beschreibt einen Input, den es nicht mehr gibt]** (code-smell, assets/jotti-symbol.png)
Datei: scripts/generate-spektral-logos.py:2-14
Warum: Die Master in assets/ sind bereits spektral; ein erneuter Lauf remappt nur das grüne Band und korrumpiert still.
Vorschlag: Skript löschen oder als Einmal-Generator kennzeichnen und den Lauf bei bereits spektralem Input verweigern.
Aufwand: S · Status: unverifiziert (minor)

**[Prüfung, die nicht fehlschlagen kann]** (code-smell, code-smells.md Defensive Overkill)
Datei: scripts/generate-spektral-logos.py:243-245
Warum: Die Dateien wurden gerade geschrieben und validiert; ein fehlgeschlagenes `save` würde ohnehin werfen.
Vorschlag: Zeilen 243-245 löschen, `len(MASTER_NAMES)` direkt melden und den Docstring-Punkt streichen.
Aufwand: S · Status: unverifiziert (minor)

### docker-compose.local.yml

**[SECURITY-Kopf widerspricht dem eigenen Dateikopf]** (boundary-consistency, docker-compose.local.yml:5-7)
Datei: docker-compose.local.yml:15-16
Warum: Zeile 15 macht die interne CA zum Sicherheitsmodell, Zeile 5-7 nennt Let's Encrypt als Primärpfad.
Vorschlag: Zeilen 15-16 an 5-7 angleichen und nur den Satz „nie ins öffentliche Internet exponieren" als Warnung behalten.
Aufwand: S · Status: bestätigt

### docker-compose.release.yml

**[Kommentare erklären per „nicht mehr"]** (docs-accuracy, AGENTS.md Regel 18)
Datei: docker-compose.release.yml:70-72; Makefile:145-146
Warum: Beide beschreiben einen entfallenen Schritt statt des geltenden Vertrags.
Vorschlag: Positiv formulieren: Migrationen sind ins jotti-migrate-Image gebacken, deshalb kein Bind-Mount und kein Ordner im ZIP.
Aufwand: S · Status: unverifiziert (minor)

**[Vier handsynchronisierte Kopien desselben Service-Graphen]** (large-refactor, principles.md DRY)
Datei: docker-compose.release.yml:31-172 (plus local, prod, e2e)
Warum: Healthchecks, Netze und depends_on stehen viermal identisch, ohne Test auf Übereinstimmung.
Vorschlag: Nur markiert — Basisdatei plus `include:`/Overrides berührt Projektnamen und Volume-Identitäten und braucht eine eigene Planung.
Aufwand: L · Status: unverifiziert (minor)

### docker-compose.initial-cert.yml

**[Toter Renew-Loop im certbot-Service]** (code-smell, scripts/rocks-init.sh:119-130)
Datei: docker-compose.initial-cert.yml:17-30
Warum: Der einzige Aufrufer startet nur den Proxy und überschreibt entrypoint/command; die 12h-Schleife läuft nie.
Vorschlag: `restart`, `entrypoint` und `command` entfernen und zwei Kopfzeilen ergänzen, dass dieser Stack nur die Erstausstellung macht.
Aufwand: S · Status: unverifiziert (minor)

### .github/workflows/ci.yml

**[shellcheck ohne -x, source-Direktiven wirkungslos]** (ops, scripts/*.sh)
Datei: .github/workflows/ci.yml:342-343
Warum: Zehn `# shellcheck source=scripts/lib.sh`-Direktiven werden nie beachtet, weil externe Quellen deaktiviert sind.
Vorschlag: `shellcheck -x scripts/*.sh` ausführen oder `external-sources=true` in eine .shellcheckrc schreiben.
Aufwand: S · Status: unverifiziert (minor)

**[ci-Filter greift nur für upgrade-path]** (ops, ci.yml:412-414)
Datei: .github/workflows/ci.yml:54-55
Warum: Ein PR, der andere Jobs dieser Datei ändert, führt keinen davon aus; die Begründung bei upgrade-path gilt für alle.
Vorschlag: `|| needs.changes.outputs.ci == 'true'` in jede Job-Bedingung aufnehmen.
Aufwand: S · Status: unverifiziert (minor)

**[Tautologische Bedingung am e2e-Job]** (code-smell, ci.yml:3-7)
Datei: .github/workflows/ci.yml:512-513
Warum: Der Workflow triggert nur auf push und pull_request; die Bedingung kann nie falsch sein.
Vorschlag: Die `if:`-Zeile löschen und den erklärenden Kommentar behalten.
Aufwand: S · Status: unverifiziert (minor)

**[resolver-ci und local-proxy-ci sind identische Jobs]** (code-smell, ci.yml:220-272)
Datei: .github/workflows/ci.yml:126-218
Warum: Acht gleiche Schritte zweimal, während windows-ci dieselbe Aufgabe per `strategy.matrix.module` löst.
Vorschlag: Beide zu einem Matrix-Job über `[resolver, reverse-proxy]` zusammenfassen und die `changes`-Outputs oderverknüpfen.
Aufwand: M · Status: unverifiziert (minor)

### packaging/cron/jotti-backup.cron

**[Begründung verweist auf „the .env fix"]** (docs-accuracy, AGENTS.md Regel 18)
Datei: packaging/cron/jotti-backup.cron:6-10; packaging/systemd/jotti-backup.service:11-12
Warum: Der Betreiber kann die genannte Voraussetzung nicht prüfen; das Verhalten selbst ist unbedingt korrekt.
Vorschlag: „Ein fehlender Schlüssel fällt auf den Standard zurück (./backups, 14), deshalb ist das Setzen optional."
Aufwand: S · Status: unverifiziert (minor)

### cliff.toml

**[Verweis auf einen gelöschten Plan]** (docs-accuracy, AGENTS.md Git-Workflow)
Datei: cliff.toml:1-4
Warum: Abgeschlossene Pläne werden nach dem Merge gelöscht; der einzige Zeiger auf die Begründung läuft ins Leere.
Vorschlag: Den Verweissatz streichen — die Inline-Kommentare tragen die Begründung bereits.
Aufwand: S · Status: unverifiziert (minor)

### database/migrations/README.md

**[Versions-Pin doppelt gepflegt]** (code-smell, .github/workflows/ci.yml:417)
Datei: database/migrations/README.md:44
Warum: „aktuell v0.17.1" muss je Release zusätzlich gebumpt werden und veraltet still.
Vorschlag: Konkrete Version streichen und nur den Ort nennen: `PREVIOUS_VERSION` im Job `upgrade-path`.
Aufwand: S · Status: unverifiziert (minor)

### .github/instructions/frontend.instructions.md

**[Doppelte Doku-Zeiger mit abweichender Kapitelangabe]** (docs-accuracy, backend.instructions.md:6,110)
Datei: .github/instructions/frontend.instructions.md:55-58
Warum: Kopf-Blockquote und Schlussabschnitt zeigen aufs selbe Dokument, nennen aber „§6.3" und „Kap. 6".
Vorschlag: Den Abschnitt „Weiterführende Dokumentation" in beiden Dateien löschen; der Blockquote bleibt kanonisch.
Aufwand: S · Status: unverifiziert (minor)

### .github/instructions/backend.instructions.md

**[Beispielcode überschattet den Paketnamen]** (readability, readability.md Clever Code)
Datei: .github/instructions/backend.instructions.md:162-184
Warum: `produkt, err := produkt.NewProdukt(...)` macht das Paket unbrauchbar, und `else` nach `return` verstößt gegen indent-error-flow.
Vorschlag: Variable in `neuesProdukt` umbenennen und den Fehlerzweig flach ziehen.
Aufwand: S · Status: unverifiziert (minor)

**[Tabs im ansonsten leerzeichenbasierten Verzeichnisbaum]** (convention, readability.md Formatting)
Datei: .github/instructions/backend.instructions.md:26-27
Warum: Zwei Zeilen sitzen visuell auf einer Ebene, die es in der Struktur nicht gibt.
Vorschlag: Führenden Tab durch zwei Leerzeichen ersetzen.
Aufwand: S · Status: unverifiziert (minor)

## Dokumentation

Umfang: anwender- und agentenseitige Dokumentation — `README.md`, `CHANGELOG.md`, Rechtsdokumente, `docs/**`, `.github/instructions/**`.

Befunde: 1 Blocker · 17 Major · 22 Minor (aus 74 Roh-Befunden zusammengeführt).

### docs/leitfaden/tse-einrichten.md

**[Einstieg in den TSE-Assistenten existiert nicht]** (docs-accuracy, AGENTS.md Regel 11)
Datei: docs/leitfaden/tse-einrichten.md:26-32
Warum: Der gesetzlich zwingende TSE-Einrichtungsschritt ist nach Anleitung nicht ausführbar.
Vorschlag: Pfad korrigieren auf „Finanzamt & TSE" → Schrittkarte „2 · TSE aktiv" → „TSE einrichten" (`/admin/tse-einrichtung`); separat die UI-Lücke prüfen, dass `EinrichtungSection` bei `tseOk` keinen Link mehr anbietet.
Aufwand: S · Status: bestätigt

Beleg: Zeile 28 „Im Admin-Bereich „Finanzamt" öffnen, im Kasten „TSE-Anbindung" auf „Einrichten oder ändern" klicken." — `grep -rn "TSE-Anbindung\|Einrichten oder ändern" frontend/src` liefert null Treffer; `frontend/src/admin/finanzamt/FinanzamtPage.tsx:10` `titel="Finanzamt & TSE"`; `EinrichtungSection.tsx:181-198` rendert den Link nur im Zweig `!tseOk`.

### docs/leitfaden/tse-sonderfaelle.md

**[Gleicher toter Einstiegspfad für die Experten-Konfiguration]** (docs-accuracy, cross-layer)
Datei: docs/leitfaden/tse-sonderfaelle.md:47-53
Warum: Der einzige Weg zur manuellen TSE-Konfiguration (TSS-Übernahme, Schlüsselrotation) ist nicht auffindbar.
Vorschlag: Wie oben auf „Finanzamt & TSE" → „TSE einrichten" umschreiben; „Manuelle Konfiguration (Experten)" und „Alle Felder leeren" bleiben.
Aufwand: S · Status: bestätigt

Beleg: Zeile 48-50 nennt „Finanzamt" / „TSE-Anbindung" / „Einrichten oder ändern"; `TSEKonfigurationSection.tsx:302` `<CardTitle>Manuelle Konfiguration (Experten)</CardTitle>`, :194 `Alle Felder leeren`.

### docs/leitfaden/datenaufbewahrung.md

**[Drei erfundene UI-Namen im Archivierungsablauf]** (docs-accuracy, cross-layer)
Datei: docs/leitfaden/datenaufbewahrung.md:15-17
Warum: Die Seite trägt die 10-Jahres-Aufbewahrungspflicht, führt aber zu nicht existierenden Bedienelementen.
Vorschlag: Auf „Nach dem Fest" → „Berichte & Export" → „Archiv herunterladen (ZIP)" umschreiben.
Aufwand: S · Status: bestätigt

Beleg: Doku nennt „Auswertungen", „Kassenberichte", „DSFinV-K-Export"; `AdminSidebar.tsx:146-149` `title: 'Berichte & Export'` in `<NavGroup label="Nach dem Fest">` (:215); `KassenberichtePage.tsx:52` `Archiv herunterladen (ZIP)`, :80 `titel="Berichte & Export"`.

### docs/leitfaden/veranstaltungstag.md

**[Zwei tote Bedienpfade im Tagesablauf]** (docs-accuracy, cross-layer)
Datei: docs/leitfaden/veranstaltungstag.md:10-36
Warum: Die Seite wird am Festtag unter Zeitdruck befolgt; zwei von drei Bedienelementen fehlen.
Vorschlag: Zeilen 10, 17, 36 „Kasse" → „Kassentag"; Zeile 26 „Geldtransit buchen" → Block „Heutige Kassenbewegungen" mit „Geld einlegen" / „Geld entnehmen".
Aufwand: S · Status: bestätigt

Beleg: `AdminSidebar.tsx:113` `title: 'Kassentag'`; `KassensitzungPage.tsx:150-152`; `LaufenderBetriebSection.tsx:107-132` `Geld einlegen` / `Geld entnehmen`; „Geldtransit" ist in `frontend/src` kein sichtbares Label.

**[Entfernter Ausgabe-Schritt weiterhin beschrieben]** (docs-accuracy, ADR 01)
Datei: docs/leitfaden/veranstaltungstag.md:8-9
Warum: Servicekräfte suchen eine Bestätigungsaktion, die vollständig entfernt wurde.
Vorschlag: „geben aus" streichen; die Ausgabe wird über die Arbeitsbons koordiniert.
Aufwand: S · Status: unverifiziert (minor)

Beleg: `docs/adrs/01_ausgabe-bestaetigen.md:44-46` „vollständig entfernt … UI, API-Endpunkt, Event-Typ, Projektion, Contract-Eintrag und Altdatenbestand".

### docs/leitfaden/installation.md

**[Bondruck-Setup nennt Menüpunkt „Druckstationen"]** (docs-accuracy, cross-layer)
Datei: docs/leitfaden/installation.md:65-67
Warum: Ohne konfigurierte Station druckt nichts, und der genannte Menüpunkt existiert nicht.
Vorschlag: „Druckstationen" → „Bondrucker" (Gruppe „Vorbereitung"); Feldnamen „Drucker-IP" und „Bonmodus" bleiben.
Aufwand: S · Status: bestätigt

Beleg: `AdminSidebar.tsx:138-141` `title: 'Bondrucker'`; `DruckstationConfigPage.tsx:551` `titel="Bondrucker"`; „Druckstationen" nur als Lade-/Fehlertext (:515, :518).

### README.md

**[Falscher HTTP-Status im Relay-Schnelltest]** (docs-accuracy, cross-layer)
Datei: README.md:78
Warum: Ein falscher Token liefert `400` mit `{"code":"unauthorized"}`, nicht `401`.
Vorschlag: „`401` bei ungültigem" ersetzen durch „`400` mit `{\"code\":\"unauthorized\"}` bei ungültigem".
Aufwand: S · Status: bestätigt

Beleg: `backend/api/druck/relay/http/handler.go:91` `SendClientError(w, "unauthorized", nil)`; `backend/api/helper/http.go:50-52` `http.StatusBadRequest`; `relay_integration_test.go:384` erwartet 400.

**[Gleiches stale Label »Druckstationen«]** (docs-accuracy, cross-layer)
Datei: README.md:78
Warum: Der Verweis auf die Fehleranzeige für Druckaufträge zeigt auf einen nicht existierenden Menüpunkt.
Vorschlag: »Druckstationen« → »Bondrucker«.
Aufwand: S · Status: unverifiziert (minor)

Beleg: `AdminSidebar.tsx:137` `title: 'Bondrucker'`.

**[Relay-Token nur als Kommandozeilen-Variable dokumentiert]** (security, Secrets)
Datei: README.md:71-76
Warum: Der kopierte Befehl legt den geteilten Relay-Token in der Shell-History ab.
Vorschlag: Die `.env` neben der Binary (oder systemd `EnvironmentFile`) als Hauptweg dokumentieren, die Inline-Variante nur als Ad-hoc-Test.
Aufwand: S · Status: unverifiziert (minor)

Beleg: `windows/relay/env.go:63-75` sucht `.env` in `exeDir` und `wd`; :93-99 lässt echte Umgebungsvariablen gewinnen.

**[Fiskal-Bausteine plus Betreiber-Caveat dreimal wiederholt]** (readability, DRY)
Datei: README.md:11-123
Warum: Drei Kopien derselben Compliance-Aussage driften bei der nächsten Änderung auseinander.
Vorschlag: Vollständige Aussage nur im Compliance-Hinweis (:123) belassen; :11 und :115 auf Claim plus Link zu `docs/compliance.md` kürzen.
Aufwand: S · Status: unverifiziert (minor)

**[SERVICE.md ist nirgends verlinkt]** (docs-accuracy, tote Referenzlücke)
Datei: README.md:139-141
Warum: Die Leistungsbedingungen sind vom Einstiegsdokument nur über TERMS.md § 7 erreichbar.
Vorschlag: In der Fußzeile ergänzen: „Leistungsbedingungen: [SERVICE.md](SERVICE.md)".
Aufwand: S · Status: unverifiziert (minor)

Beleg: `grep -n SERVICE README.md` liefert keinen Treffer; LICENSE, CLA.md und TERMS.md sind verlinkt.

### CLA.md

**[Unwiderruflichkeitsklausel benennt die falsche Partei]** (correctness, LICENSE:112-113)
Datei: CLA.md:36
Warum: Abschnitt 1 räumt die Rechte dem Autor ein; § 2 b) erklärt die Rechte der beitragenden Person für unwiderruflich.
Vorschlag: „Die **dem Autor** gewährten Rechte nach Abschnitt 1 sind unwiderruflich und können nicht zurückgenommen werden."
Aufwand: S · Status: bestätigt

Beleg: CLA.md:25 „gewährt dem Autor hiermit eine unwiderrufliche … Lizenz" vs :36; LICENSE:112-113 „may not revoke the rights granted to the Author".

### docs/handbuch.md

**[Read-Model „Tischdetails" listet zwei nicht existierende Tabs]** (docs-accuracy, AGENTS.md Regel 18)
Datei: docs/handbuch.md:481
Warum: Die Architektur-Referenz widerspricht ihrer eigenen Zeile 419 und dem gelieferten UI.
Vorschlag: Zeile 481 auf „Tabs: Bestellen, Kassieren, Historie" kürzen.
Aufwand: S · Status: bestätigt

Beleg: :481 „Tabs: Übersicht, Bestellen, Bezahlen/Kassieren, Stornieren, Historie" vs :419; `frontend/src/service/TablePage.tsx:231-239` rendert genau drei `TabsTrigger`.

**[Admin-Seitenstruktur unvollständig, Komponentenname falsch]** (docs-accuracy, AGENTS.md Regel 11)
Datei: docs/handbuch.md:417-421
Warum: Fünf von neun Admin-Seiten fehlen und `DruckerConfigPage` existiert nicht.
Vorschlag: Zeile 420 gegen `frontend/src/routes.ts` vervollständigen und in `DruckstationConfigPage` umbenennen; Service-Zeile um Direktverkauf ergänzen.
Aufwand: S · Status: bestätigt

Beleg: :420 „Druckerkonfiguration (`DruckerConfigPage`, …)"; `grep -rn 'DruckerConfigPage' frontend/src` → 0 Treffer; `routes.ts:84-153` listet neun Admin-Kinder.

**[Kassenbeleg-Abschnitt kennt nur zwei von vier Body-Formen]** (docs-accuracy, § 146a Abs. 2 AO)
Datei: docs/handbuch.md:306
Warum: Der Belegpfad für kassenwirksame Warenrücknahmen fehlt in der Architektur-Referenz.
Vorschlag: Alle vier Formen nennen (tischId+zahlungId, tischId+stornierungId, verkaufId, verkaufId+stornierungId); dieselbe Lücke in `docs/compliance.md:280` schließen.
Aufwand: S · Status: bestätigt

Beleg: `backend/api/druck/beleg/http/command_handler.go:49-64` definiert vier Schemas, :108 „eine der vier gültigen Body-Formen".

**[`01_initial.up.sql` als vollständiges Schema bezeichnet]** (docs-accuracy, Freeze-Disziplin)
Datei: docs/handbuch.md:81,265
Warum: Acht additive Migrationen ändern das Schema; `01_initial.up.sql` ist eingefroren.
Vorschlag: Beide Verweise auf „alle `*.up.sql` in `database/migrations/` in Reihenfolge" umstellen; gleiche Korrektur in `docs/verfahrensdokumentation.md:74`.
Aufwand: S · Status: unverifiziert (minor)

Beleg: Migrationen 02-08 fügen `naechster_versuch_ab`, `elster_gemeldet_am`, `reihenfolge`, `pro_stueck` hinzu und entfernen `ausstehende_positionen`.

**[Historisierende Prosa im TSE-Abschnitt]** (convention, AGENTS.md Regel 18)
Datei: docs/handbuch.md:238,244
Warum: „ersetzt das frühere …" und „entfällt deshalb ersatzlos" beschreiben einen Vorzustand statt des geltenden Stands.
Vorschlag: Positiv formulieren: Störungsprotokoll `tse_stoerungen` hält je Störung einen Zeitraum; die Admin-Seite ist reines Monitoring ohne mutierende Aktionen.
Aufwand: S · Status: unverifiziert (minor)

**[Authentifizierungsfläche unvollständig beschrieben]** (docs-accuracy, cross-layer)
Datei: docs/handbuch.md:388-404
Warum: Neben `/auth/*` sind auch `/health`, `/relay/*` und der bedingte Bereich `/test/reset-and-seed` ohne JWT erreichbar.
Vorschlag: Satz auf „außer `/auth/*`, `/relay/*` und `/health`" korrigieren und den Test-Bereich (`JOTTI_ENABLE_TEST_API=1`, ohne JWT) in der Tabelle ergänzen.
Aufwand: S · Status: unverifiziert (minor)

Beleg: `backend/app/routes.go:120-137`; `backend/app/app.go:62-66` „bewusst ohne JWT wie auth/relay".

### docs/compliance.md

**[Widerspruch zur Abrechnungskreis-Regel]** (docs-accuracy, DRY)
Datei: docs/compliance.md:137
Warum: Zeile 137 verspricht einen neuen Abrechnungskreis „Tisch 42-B", § 6.5 (Zeile 355) sagt das Gegenteil, der Exporter erzeugt kein Suffix.
Vorschlag: Den Satz in Zeile 137 streichen; § 6.5 bleibt die einzige Aussage zu diesem Fall.
Aufwand: S · Status: bestätigt

Beleg: `backend/api/fiskal/dsfinvk/mapper.go:528-534` leitet den ABRECHNUNGSKREIS allein aus dem Tischnamen ab.

**[Versionsstring fälschlich als konfigurierbar zugesagt]** (docs-accuracy, AGENTS.md Regel 11)
Datei: docs/compliance.md:296
Warum: Der DSFinV-K-Versionsstring ist eine Compile-Zeit-Konstante ohne Konfigurationspfad.
Vorschlag: Auf „zentral als Konstante `dsfinvk.Version`; eine Umstellung ist ein Codeänderungs-Release" umschreiben und den Kommentar in `dsfinvk.go:19` mitziehen.
Aufwand: S · Status: bestätigt

Beleg: `backend/api/fiskal/dsfinvk/dsfinvk.go:22` `const Version = "2.4"`; kein Config-Key oder Env-Var im Repo.

**[Planungs-Tempus für ausgelieferte Exportfunktionen]** (docs-accuracy, AGENTS.md Regel 18)
Datei: docs/compliance.md:354-376
Warum: „Entscheidung bei Implementierung des Exporters" und „Zielbild ist" beschreiben umgesetztes Verhalten als offen.
Vorschlag: § 6.5 als Ist-Stand formulieren (Direktverkauf-Bons ohne ABRECHNUNGSKREIS), § 6.6 auf die umgesetzte 1:1-Abbildung mit `REF_BON_ID` umstellen, § 6.7 als Umsetzungsbeschreibung führen.
Aufwand: M · Status: unverifiziert (minor)

Beleg: `mapper.go:534`, :752-753, :375, :843-864; `docs/anforderungen.md:110` führt F-04 als umgesetzt.

### .github/instructions/database.instructions.md

**[Selbstwiderspruch über das kanonische Schema]** (docs-accuracy, DRY / Freeze-Disziplin)
Datei: .github/instructions/database.instructions.md:6
Warum: Zeile 6 erklärt `01_initial.up.sql` zum kanonischen Schema, Zeile 26 die Summe aller Migrationen.
Vorschlag: Zeile 6 auf die Formulierung von Zeile 26 ziehen („alle `*.up.sql`-Dateien in Reihenfolge").
Aufwand: S · Status: bestätigt

Beleg: 04 erweitert `bon_art` um `testbon`, 07 fügt `produkte.reihenfolge` hinzu, 03 entfernt `tisch_sessions.ausstehende_positionen`; AGENTS.md:61 friert 01 ein.

### docs/produktbeschreibung.md

**[Raspberry Pi als Plattform genannt, Images sind amd64-only]** (docs-accuracy, cross-layer)
Datei: docs/produktbeschreibung.md:154,184
Warum: Ein Verein kauft nach dieser Aussage ARM-Hardware, auf der die Release-Images nicht starten.
Vorschlag: Entweder Multi-Arch-Images bauen (`--platform linux/amd64,linux/arm64`) oder beide Zeilen auf x86-64-Server einschränken.
Aufwand: S · Status: bestätigt

Beleg: `.github/workflows/release.yml:102` `docker buildx build --load` ohne `--platform`; `docker-compose.prod.yml:68` und `docker-compose.release.yml:79` ziehen genau diese Images.

**[Steuersätze: falsche Ebene, fehlender Enum-Wert]** (docs-accuracy, docs/language.md:489)
Datei: docs/produktbeschreibung.md:138
Warum: Der Steuersatz hängt am Produkt, nicht an der Variante, und `kombi` (70/30) fehlt.
Vorschlag: „19 % (regel), 7 % (ermaessigt), 0 % (befreit), Kombi 70/30 (kombi), konfigurierbar pro Produkt".
Aufwand: S · Status: bestätigt

Beleg: `01_initial.up.sql:66,80` `produkte.steuersatz`; `backend/domain/produkt/product.go:41` — die `Variante` trägt kein solches Feld.

### docs/prds/prd-windows-nativ-ohne-docker.md

**[Aktueller Windows-Standardweg als „früher" bezeichnet, plus tote Rückverweise]** (convention, AGENTS.md Regel 18)
Datei: docs/prds/prd-windows-nativ-ohne-docker.md:3-9,35,76-81
Warum: Die Docker-Verpackung ist der ausgelieferte Standardweg, und „Phase B" sowie die „Vorgänger-PRD" existieren nicht.
Vorschlag: Kopf auf Gegenwart umstellen; Verweise in :35 und :98 entfernen; die TLS-Aussage am Ist-Stand (Caddy interne CA `on_demand`, DNS-01-Wildcard) neu formulieren.
Aufwand: S · Status: bestätigt

Beleg: `docs/leitfaden/installation.md:3,12,17-22` führt `jotti-start.exe` als Standardweg; `reverse-proxy/caddyfile.go:66-69` — keine Selbstsignierung im Entrypoint.

### docs/verfahrensdokumentation.md

**[Bounded-Context-Tabelle abgedriftet, „Ausgeben" als Geschäftsvorfall]** (docs-accuracy, DRY / ADR 01)
Datei: docs/verfahrensdokumentation.md:46-52
Warum: Das Dokument geht an Betriebsprüfer und nennt drei statt sechs Kontexte sowie einen entfernten Vorgang.
Vorschlag: „Ausgeben" streichen, Aufzählung an `docs/handbuch.md:37` angleichen und Fiskalisierung sowie Druck/Ausgabe ergänzen oder die Tabelle ausdrücklich als Teilsicht kennzeichnen.
Aufwand: S · Status: bestätigt

Beleg: `docs/handbuch.md:35-42` listet sechs Kontexte; `database/migrations/03_ausgabe_entfernen.up.sql:19` löscht `ausgabe-bestaetigt:v1`.

### docs/language.md

**[Abschnitt „Geplant" verweist auf eine leere Roadmap]** (docs-accuracy, tote Referenz)
Datei: docs/language.md:516-521
Warum: `docs/anforderungen.md:17` erklärt die Roadmap für leer und kennt weder Stornoquote noch Privatentnahme.
Vorschlag: Entweder beide Begriffe mit ID in `docs/anforderungen.md` aufnehmen oder den Abschnitt streichen und den Geldtransit-Hinweis an dessen Glossareintrag hängen.
Aufwand: S · Status: bestätigt

Beleg: `grep -rni 'stornoquote\|privatentnahme\|privateinlage' docs/` trifft nur `docs/language.md:520-521`.

**[Deprecation-Vermerk im Glossar, Altbegriff weiter in Gebrauch]** (convention, AGENTS.md Regel 18)
Datei: docs/language.md:468
Warum: Das Glossar erklärt „Nachsignierung" für ersetzt, zwei Dokumente verwenden den Begriff weiter.
Vorschlag: Den Satz streichen und `docs/anforderungen.md:117` sowie `docs/plans/guide-manuelle-qa-v1.0.0.md:11,37` auf „Nachsigniert" umstellen.
Aufwand: S · Status: unverifiziert (minor)

### docs/leitfaden/self-hosting.md

**[.env-Beispiel pinnt die fehlerhafte Version v0.14.0]** (docs-accuracy, Template Pollution)
Datei: docs/leitfaden/self-hosting.md:45-55
Warum: Wer das Beispiel abtippt, installiert genau die Version mit leerem DSFinV-K-Pflichtfeld `TSE_SERIAL`.
Vorschlag: Platzhalter setzen: `JOTTI_VERSION=v<aktuelles Tag von der Releases-Seite>`.
Aufwand: S · Status: unverifiziert (minor)

Beleg: `docs/leitfaden/tse-sonderfaelle.md:40-42`; `.env.example:21` hat bewusst keinen Default.

### docs/leitfaden/aktualisieren.md

**[Ganzer Abschnitt fest auf Release 0.17.3 verdrahtet]** (convention, AGENTS.md Regel 18)
Datei: docs/leitfaden/aktualisieren.md:115-127
Warum: Eine Überschrift des Evergreen-Leitfadens behauptet den Inhalt „des Release-ZIPs" und wird mit 1.0.0 falsch.
Vorschlag: Die 0.17.3-Spezifika (:73-86, :58-60, :115) in `CHANGELOG.md` bzw. die Release Notes verschieben; im Leitfaden die versionsunabhängige Regel behalten.
Aufwand: M · Status: unverifiziert (minor)

### docs/leitfaden/betriebsarten.md

**[Ungrammatische und vierte Schreibweise der Helfer-Rolle]** (readability, docs/language.md:56)
Datei: docs/leitfaden/betriebsarten.md:28
Warum: „eine Helfer:in" ist ungrammatisch, und dieselbe Rolle heißt in den Anwenderdokumenten viermal anders.
Vorschlag: „eine Servicekraft bedient das Gerät an der Theke" (Begriff aus `docs/language.md`).
Aufwand: S · Status: unverifiziert (minor)

### CHANGELOG.md

**[Version 1.0.0 ohne Datum, obwohl unveröffentlicht]** (convention, Keep a Changelog)
Datei: CHANGELOG.md:12-14
Warum: Der Eintrag behauptet ein Release, das es nicht gibt (letztes Tag v0.17.3, README:4 nennt Beta).
Vorschlag: Bis zum Tag `## [Unreleased]`, danach `## [1.0.0] - YYYY-MM-DD`.
Aufwand: S · Status: unverifiziert (minor)

**[Zwei ausgelieferte Funktionen fehlen im Funktionsumfang]** (docs-accuracy, cross-layer)
Datei: CHANGELOG.md:44
Warum: Abholbon samt Bonmodus „Pro Stück" und der Versions-Handshake stehen in keinem Anwenderdokument.
Vorschlag: Unter „Küche" den Abholbon und die drei Bonmodi ergänzen, unter „Verwaltung" das automatische Neuladen; `README.md:27` und `docs/leitfaden/installation.md` nachziehen.
Aufwand: M · Status: unverifiziert (minor)

Beleg: `backend/domain/druckstation/druckstation.go:18,47-49`; `DruckstationConfigPage.tsx:84,310`; `grep -rl "Abholbon" docs/leitfaden README.md CHANGELOG.md` → kein Treffer.

**[Doppelte Aussage in der Einleitung]** (readability, AGENTS.md Regel 18)
Datei: CHANGELOG.md:3-5
Warum: „von Hand gepflegt" und „wird manuell fortgeschrieben" sagen dasselbe und suggerieren ein Vorher.
Vorschlag: Zweiten Satz streichen.
Aufwand: S · Status: unverifiziert (minor)

### docs/plans/guide-manuelle-qa-v1.0.0.md

**[Drei Links auf eine nicht existierende Plandatei]** (code-smell, tote Verweise)
Datei: docs/plans/guide-manuelle-qa-v1.0.0.md:14,59,71
Warum: Die Release-Voraussetzung für Block H verweist auf `plan-v1.0-release-blockers.md`, die es nicht gibt.
Vorschlag: Die vier Blocker (zog, G3, G7, `CHANGELOG.md`) direkt im Guide auflisten, die Links entfernen und die „wurde gelöscht"-Einschübe (:3, :14, :72) streichen.
Aufwand: S · Status: unverifiziert (minor)

Beleg: `git ls-files docs/plans` listet nur vier Dateien; `docs/plans/plan-praxis-feedback.md:608-614` führt den toten Verweis bereits als offenen Punkt.

### docs/plans/review-externe-prs.md

**[Mehrfach eingetragene identische Befunde]** (code-smell, DRY)
Datei: docs/plans/review-externe-prs.md:40-49
Warum: Die Kopfzahl „35 bestätigt" zählt einzelne Defekte bis zu dreimal.
Vorschlag: Je Defekt eine Zeile behalten (#109 :40/:41, #110 :104/:105/:109, #111 :159/:160, :162/:166/:169, :163/:168) und die Zählung in :5 und :15-17 korrigieren.
Aufwand: M · Status: unverifiziert (minor)

### docs/plans/plan-orchestrierung.md

**[Fable-Ausnahme deckt den eigenen Rolleneinsatz nicht]** (convention, CLAUDE.md)
Datei: docs/plans/plan-orchestrierung.md:17-26
Warum: Der Plan setzt Fable 5.1 auch als Lead-Session ein, die behauptete Freigabe gilt nur für den Sweep.
Vorschlag: Zeile 17 auf Opus 5 umstellen oder die Ausnahme in `CLAUDE.md` selbst festhalten; ein Plan hebt keine Instruktionsregel auf.
Aufwand: S · Status: unverifiziert (minor)

### docs/adrs/05_spektral-branding-website.md

**[Tote Kontext-Verweise ohne den in ADR 01-04 etablierten Vermerk]** (convention, tote Verweise)
Datei: docs/adrs/05_spektral-branding-website.md:4-6
Warum: ADR 05-08 zeigen auf sieben nicht existierende PRDs, Pläne und ein Handoff-Verzeichnis.
Vorschlag: Eigentümerentscheidung: entweder „(nach Merge gelöscht, siehe Git-Historie)" wie in ADR 01 nachtragen oder die Konvention einmal in `docs/adrs/README.md` festhalten.
Aufwand: S · Status: unverifiziert (minor)

### docs/rechtsquellen/README.md

**[Datierter Änderungseintrag über ein anderes Dokument]** (convention, AGENTS.md Regel 18)
Datei: docs/rechtsquellen/README.md:105
Warum: „(compliance.md wurde entsprechend von 2.5 auf 2.4 korrigiert)" protokolliert eine Bearbeitung statt des Stands.
Vorschlag: Auf die Tatsache kürzen: „v2.4 ist die jüngste auf bzst.de gelistete Fassung; eine 2.5 war dort nicht auffindbar."
Aufwand: S · Status: unverifiziert (minor)

### .github/instructions/backend.instructions.md

**[ASCII-Ersatzschreibung „fuer" in deutscher Instruktionsprosa]** (readability, readability-de)
Datei: .github/instructions/backend.instructions.md:57
Warum: Dieselbe Datei schreibt an allen anderen Stellen „für"; Agenten lesen sie als Stilvorlage.
Vorschlag: „fuer" → „für"; gleiche Korrektur in `.github/instructions/event-sourcing.instructions.md:62`.
Aufwand: S · Status: unverifiziert (minor)

## Cross-Layer-Flüsse

Befunde entlang durchgehender Flüsse: Domain → Repository → HTTP-DTO → Frontend-Schema → Doku → Betrieb.

**0 Blocker · 16 Major · 38 Minor** (68 Rohbefunde, Near-Duplicates zusammengeführt).

### backend/api/druck/bondruck/application/escpos/formatter.go

**[Alle Beleg- und Bon-Zeitstempel werden in UTC gedruckt]** (correctness, KassenSichV § 6 / docs/compliance.md:232)
Datei: backend/api/druck/bondruck/application/escpos/formatter.go:113-305
Warum: Belege und Küchenbons zeigen 1–2 h vor Ortszeit; nahe Mitternacht steht der falsche Kalendertag darauf.
Vorschlag: Europe/Berlin einmal im escpos-Paket laden (tzdata ist über `backend/main.go:10` eingebettet) und alle Zeitpunkte vor `.Format(...)` mit `.In(berlin)` umrechnen; Erwartungen in formatter_test.go:24/54/417-418 nachziehen.
Aufwand: M · Status: bestätigt

**[Unerreichbarer Kombi-Zweig in `steuerMatrixLabel`]** (code-smell)
Datei: backend/api/druck/bondruck/application/escpos/formatter.go:453-454
Warum: `steuer.Steuermatrix` liefert nur Regel-, Ermäßigt- und Befreit-Zeilen, nie einen Kombi-Satz.
Vorschlag: `case steuer.KombiSteuersatz` in `steuerMatrixLabel` streichen; `default: "?"` deckt den unmöglichen Fall ab, `steuerKennzeichenAusSatz` bleibt unberührt.
Aufwand: S · Status: unverifiziert (minor)

**[Kommentar nennt einen „Nachsignier-Worker“, den es nicht gibt]** (convention, docs/language.md:465)
Datei: backend/api/druck/bondruck/application/escpos/formatter.go:60-63
Warum: Die Ubiquitous Language kennt nur den Signatur-Worker; der Leser sucht eine zweite Komponente.
Vorschlag: „(Nachsignier-Worker)“ → „(Signatur-Worker)“ und die Bezeichner in `backend/seed/faketse.go:56-59,200,273-289` auf den Glossarbegriff ziehen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/druck/bondruck/application/arbeitsbon_policy.go

**[Bon-Routing dekodiert eingefrorenes Event-JSON in ein Domain-Struct ohne json-Tags]** (boundary-consistency, AGENTS.md Regel 10)
Datei: backend/api/druck/bondruck/application/arbeitsbon_policy.go:18-21
Warum: Eine Umbenennung in `kasse.Position` bricht das Bon-Routing still — kein Compile-Fehler, kein Test schlägt an.
Vorschlag: `Positionen []kasse.PositionEventData` deklarieren und über `kasse.PositionFromEventData` wandeln, wie `backend/seed/bondruck.go:292-298` es bereits tut.
Aufwand: S · Status: bestätigt

**[Unmarshal-Fehler wird verworfen: kein Druckauftrag, kein Log]** (correctness)
Datei: backend/api/druck/bondruck/application/arbeitsbon_policy.go:186-193
Warum: Der Verkauf gilt als erfolgreich, die Theke bekommt keinen Bon, das Log bleibt leer.
Vorschlag: `CreateArbeitsbonAuftraegeFromEvent` einen `ctx` mitgeben (beide Call-Sites haben ihn) und den Fehler mit Event-ID und Subject per `zerolog.Ctx(ctx).Error()` protokollieren.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/kasse/tischgeschaeft/application/command.go

**[Kassenstatus-Race wird als 500 statt 409 gemeldet]** (correctness)
Datei: backend/api/kasse/tischgeschaeft/application/command.go:135-147
Warum: Servicekräfte sehen „unerwarteter Serverfehler“ statt „Die Kasse wird gerade abgeschlossen“; die Handler-Zweige sind unerreichbar.
Vorschlag: In `persistTischEvent` (:139-143) und `BestellungAufnehmen` (:231-246) vor dem `ErrDatabase`-Fallback `if errors.Is(err, ErrKasseNichtGeoeffnet) { return err }` ergänzen — wie `persistStornoEvents` (:490-493).
Aufwand: S · Status: bestätigt

### backend/api/kasse/kassenfuehrung/application/command.go

**[Doc-Kommentar verortet die Z-Bon-Summen im falschen Kontext]** (docs-accuracy, Regel 18)
Datei: backend/api/kasse/kassenfuehrung/application/command.go:260-261
Warum: Die Summen kommen aus `kasse.ComputeAbschlussSummen` über Sitzungs-Events, nicht aus `GetReporting`.
Vorschlag: Zeile 261 ersetzen: Tagessummen rechnet `kasse.ComputeAbschlussSummen` aus den Sitzungs-Events; Äquivalenz-Guard ist `backend/repository/reporting_repo/summen_abschluss_test.go`.
Aufwand: S · Status: bestätigt

### backend/api/kasse/kassenfuehrung/application/query.go

**[Sitzung im Status `wird_abgeschlossen` ist im Admin unsichtbar]** (boundary-consistency)
Datei: backend/api/kasse/kassenfuehrung/application/query.go:14-24
Warum: Nach einem Absturz zeigt die Kassentag-Seite das Eröffnen-Formular; jede Buchung und jedes Eröffnen scheitert.
Vorschlag: Query auf `GetAktiveKassensitzung` umstellen und in `KassensitzungPage.tsx:200,244-247` bei `wird_abgeschlossen` Schritt 3 mit Hinweis „Abschluss unterbrochen — erneut abschließen“ rendern.
Aufwand: M · Status: bestätigt

### backend/api/kasse/enrichment/enrichment.go

**[Variante wird nie gegen das mitgesendete Produkt geprüft]** (correctness)
Datei: backend/api/kasse/enrichment/enrichment.go:72-99
Warum: Ein Client kann eine fremde Variante buchen; falscher Steuersatz wird als Fat Event eingefroren und TSE-signiert.
Vorschlag: `GetVariantenByIDs` (produkt_repo/batch.go:23) um `produkt_id` erweitern und in `EnrichPositionen` vor dem Aktiv-Check vergleichen; bei Abweichung `ErrProduktNotFound`. Mock (produkt_repo/mock.go:25-28) mitziehen.
Aufwand: S · Status: bestätigt

**[Datierter Änderungseintrag im Package-Kommentar]** (convention, Regel 18)
Datei: backend/api/kasse/enrichment/enrichment.go:1-5
Warum: Provenienz gehört in die Git-Historie und erklärt nichts über das aktuelle Verhalten.
Vorschlag: Die Klammer „(extracted per the 2026-07-17 review, go-code-quality-1)“ streichen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/fiskal/setup/application/command.go

**[TSE-Konfigurationsguard greift im Barrierestatus nicht]** (correctness)
Datei: backend/api/fiskal/setup/application/command.go:17-44
Warum: Während `wird_abgeschlossen` entstehen noch signaturpflichtige Events; ein TSS-Wechsel spaltet den Kassentag.
Vorschlag: In `kassensitzungReader` `GetAktiveKassensitzung` statt `GetOffeneKassensitzung` nutzen und den Kommentar auf „offen oder wird_abgeschlossen“ ziehen.
Aufwand: S · Status: bestätigt

### backend/api/helper/http.go

**[409-Contract verspricht ein Altersfeld, das keine Schicht führt]** (docs-accuracy)
Datei: backend/api/helper/http.go:23-25
Warum: DTO, Domain-Fehler und Zod-Schema tragen ausschließlich `anzahl`; das Alter existiert nirgends.
Vorschlag: „and the age of the oldest“ in http.go:24 und :91 sowie „und Alter“ in docs/handbuch.md:242 streichen — oder das Alter durch alle vier Schichten führen.
Aufwand: S · Status: bestätigt

### backend/api/stammdaten/produkt/application/command.go

**[Namenskonflikt beim Produkt-Update wird 500 statt Feldfehler]** (boundary-consistency)
Datei: backend/api/stammdaten/produkt/application/command.go:76-80
Warum: `EditProductDialog.tsx:65-67` mappt `produkt_already_exists`, den das Update-Endpunkt nie sendet — toter Code.
Vorschlag: Denselben `errors.Is(err, db.ErrAlreadyExists)`-Zweig wie in `CreateProdukt` ergänzen und `ErrProduktAlreadyExists: "produkt_already_exists"` in die Fehlerkarte des Handlers aufnehmen.
Aufwand: S · Status: bestätigt

**[`DeleteVariante` prüft die Produktzugehörigkeit nicht]** (correctness)
Datei: backend/api/stammdaten/produkt/application/command.go:245-278
Warum: Ein falscher `produktId` löscht still die Variante eines anderen Produkts; der Parameter täuscht Absicherung vor.
Vorschlag: `GetVariante` um `produkt_id` erweitern und gegen den Parameter prüfen (`ErrVarianteNotFound` bei Abweichung) — oder `produktId` aus Request, Backend-Klasse und Command entfernen.
Aufwand: S · Status: unverifiziert (minor)

**[Ungenutzte Lesemethoden im Command-Interface]** (code-smell)
Datei: backend/api/stammdaten/produkt/application/command.go:20-21
Warum: `GetAllProdukte` und `GetActiveProdukte` ruft kein Command-Pfad auf; die Lesepfade nutzen `produktQueryRepo`.
Vorschlag: Beide Zeilen aus `produktRepo` entfernen; `produkt_repo.Repository` erfüllt beide Interfaces weiterhin.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/stammdaten/tisch/http/command_handler.go

**[Tisch-Namenskonflikt beim Umbenennen wird 500]** (boundary-consistency)
Datei: backend/api/stammdaten/tisch/http/command_handler.go:78-84
Warum: Die Application liefert korrekt `ErrTischAlreadyExists`, die MapError-Karte des Handlers kennt ihn nicht.
Vorschlag: `application.ErrTischAlreadyExists: "tisch_already_exists"` in die Fehlerkarte aufnehmen — wie in `TischErstellenHandler:49-52`.
Aufwand: S · Status: bestätigt

### backend/api/stammdaten/user/http/command_handler.go

**[Admin kann sich selbst deaktivieren oder herabstufen]** (correctness)
Datei: backend/api/stammdaten/user/http/command_handler.go:166-183
Warum: Der einzige Admin sperrt sich sofort aus; der Bootstrap greift danach nicht mehr.
Vorschlag: Den Self-Guard aus `DeleteUserHandler` (:200-208) auf `DeactivateUserHandler` und den Rollenwechsel in `UpdateUserHandler` ausdehnen; Switch und Rollenfeld im Frontend bei `isSelf` sperren.
Aufwand: S · Status: bestätigt

### backend/domain/user/user.go

**[Passwort wird beim Setzen getrimmt, beim Login nicht]** (boundary-consistency)
Datei: backend/domain/user/user.go:202-228
Warum: Ein mit Leerzeichen gesetztes Passwort ist danach nie wieder eingebbar; die Meldung nennt „ungültige Zugangsdaten“.
Vorschlag: In `backend/api/auth/http/command_handler.go:30` `user.PasswordSchema.Required()` (mindestens `.Trim()`) verwenden und einen Test setzen/prüfen mit umgebenden Leerzeichen ergänzen.
Aufwand: S · Status: bestätigt

### docs/handbuch.md

**[Retry-Schwelle der Druck-Outbox mit drei statt sechs Versuchen dokumentiert]** (docs-accuracy)
Datei: docs/handbuch.md:308
Warum: `MaxDruckversuche = 6`; README, Relay und Migration 02 nennen sechs, zwei kanonische Dokumente drei.
Vorschlag: handbuch.md:308 auf sechs Versuche korrigieren, `naechster_versuch_ab` (5s/15s/30s/60s/180s) ergänzen und docs/language.md:425 mitziehen.
Aufwand: S · Status: bestätigt

**[Stammdaten-Snapshot beim Kassenabschluss existiert nicht]** (docs-accuracy)
Datei: docs/handbuch.md:216
Warum: Der DSFinV-K-Export liest Betreiber, TSE- und Kassenstammdaten live; kein Guard blockiert Steuersatz-Änderungen.
Vorschlag: §3.11 und §2.2 (Zeile 60) auf den Ist-Stand ziehen: Positions-Steuersätze sind per Fat Event eingefroren, Stammdaten werden beim Export gelesen; die Abschlussregel als Betreiberpflicht kennzeichnen.
Aufwand: S · Status: bestätigt

**[Onboarding-Ablauf in §5.2 beschreibt einen nicht existierenden Flow]** (docs-accuracy)
Datei: docs/handbuch.md:364-367
Warum: Neue Benutzer sind `inactive`; der Login mit Einmalpasswort endet mit „Konto ist deaktiviert“, es gibt keine Weiterleitung.
Vorschlag: §5.2 auf vier Schritte umschreiben: anlegen (inactive, Code), „Neues Passwort festlegen“ über `POST /auth/set-password`, Admin aktiviert, dann regulärer Login; Hinweis im `UserCreatedDialog` ergänzen.
Aufwand: M · Status: bestätigt

**[§4.6 kennt nur zwei der vier Belegformen]** (docs-accuracy)
Datei: docs/handbuch.md:302-306
Warum: Handler und Application lösen vier Formen auf, darunter beide Stornobelege; die Doku nennt sie nicht.
Vorschlag: :302 und :306 um die STORNOBELEG-Familie ergänzen (negativer Betrag, „Storno zu Bon-Nr“) und die fehlende vierte Body-Form in docs/compliance.md:275 nachtragen.
Aufwand: S · Status: unverifiziert (minor)

**[§4.1 nennt schwächere Invarianten, als drei Schichten erzwingen]** (docs-accuracy)
Datei: docs/handbuch.md:269-271
Warum: Erzwungen sind Preis 1–99999 Cent und getrimmte Namen mit 3–100 Zeichen, nicht „Preis ≥ 0, nicht leer“.
Vorschlag: §4.1 auf Preis 1–99999 Cent und Name 3–100 Zeichen (getrimmt) ziehen; §4.2 (Zeile 277) analog gegen `domain/tisch/tisch.go:47`.
Aufwand: S · Status: unverifiziert (minor)

**[Regel-18-Prosa in der TSE-Doku]** (convention, Regel 18)
Datei: docs/handbuch.md:238-244
Warum: „ersetzt das frühere Ausfall-Flag“ und „die frühere Signaturauftrags-Verwaltung … entfällt“ halten abgelöste Konzepte am Leben.
Vorschlag: :238 auf „Ein Störungsprotokoll (`tse_stoerungen`) dokumentiert …“ kürzen, den Halbsatz in :244 streichen und in docs/language.md:463/468 dieselben Vorgängerbezüge entfernen.
Aufwand: S · Status: unverifiziert (minor)

### packaging/windows/KURZANLEITUNG.md

**[Dokumentiertes manuelles Backup schlägt auf frischen Installationen fehl]** (docs-accuracy)
Datei: packaging/windows/KURZANLEITUNG.md:85-91
Warum: `%PROGRAMDATA%\jotti\backups` legt nur `mirrorBackupToHost` an — also erst nach dem ersten Update.
Vorschlag: Das Host-Backup-Verzeichnis in `resolveStateDir` mit anlegen (oder dem Befehl `md "%PROGRAMDATA%\jotti\backups" 2>nul` voranstellen) und den Dump-Namen mit Zeitstempel versehen.
Aufwand: S · Status: bestätigt

### backend/domain/tse/signaturstatus.go

**[Nil-Dereferenzierung nur durch eine WHERE-Klausel verhindert]** (correctness)
Datei: backend/domain/tse/signaturstatus.go:59-64
Warum: `Signatur` ist ein optionaler Zeiger; ein Stand ohne Signatur lässt Belegabruf und Abschluss-Gate paniken.
Vorschlag: Im `StatusErledigt`-Zweig `auftrag.Signatur == nil` behandeln und einen Testfall `{Status: StatusErledigt, Signatur: nil}` in signaturstatus_test.go ergänzen.
Aufwand: S · Status: unverifiziert (minor)

### backend/sqlc/queries/tse_signaturauftraege.sql

**[Signierdauer p95 mischt TSE- und Datenbank-Uhr]** (correctness)
Datei: backend/sqlc/queries/tse_signaturauftraege.sql:84
Warum: `log_time_end` kommt sekundengenau von fiskaly, `erstellt_am` von Postgres; die Kennzahl trägt den Uhrenversatz.
Vorschlag: Auf eine Uhr stellen: `EXTRACT(EPOCH FROM (erledigt_am - erstellt_am))` — beide Werte setzt dieselbe DB-Uhr.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/fiskal/dsfinvk/mapper.go

**[Doc-Kommentar von `buildVat` widerspricht dem Code fünf Zeilen tiefer]** (docs-accuracy)
Datei: backend/api/fiskal/dsfinvk/mapper.go:632-639
Warum: Die vat.csv führt bewusst alle sieben amtlichen USt-Schlüssel, nicht nur die verwendeten.
Vorschlag: Doc-Kommentar auf „deklariert alle sieben amtlichen USt-Schlüssel der DSFinV-K-Anlage 2, aufsteigend nach Schlüssel“ ziehen; der innere Kommentar entfällt.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/fiskal/dsfinvk/table.go

**[`LogicalName` und `Description` werden befüllt, aber nie gelesen]** (code-smell)
Datei: backend/api/fiskal/dsfinvk/table.go:40-43
Warum: Die index.xml ist die eingebettete amtliche Datei; `serializeCSV` nutzt nur `Columns` und `Records`.
Vorschlag: Beide Felder aus `Table` entfernen und die Zuweisungen in den 20 `build*`-Funktionen streichen.
Aufwand: M · Status: unverifiziert (minor)

### backend/api/kasse/kassenfuehrung/application/errors.go

**[Fehler-Sentinel ohne Erzeuger und ohne HTTP-Mapping]** (code-smell)
Datei: backend/api/kasse/kassenfuehrung/application/errors.go:21-22
Warum: `ErrKasseAlreadyAbgeschlossen` wird nirgends zurückgegeben; der reale Fall läuft über `ErrKasseNichtGeoeffnet`.
Vorschlag: Deklaration und Kommentar ersatzlos löschen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/kasse/direktverkauf/application/command.go

**[Nil-Guard auf `DruckstationRepo` existiert nur für den Test]** (code-smell)
Datei: backend/api/kasse/direktverkauf/application/command.go:127-135
Warum: Produktionscode ist auf Testbequemlichkeit geformt; eine fehlende Verdrahtung endet in stillem Bon-Verlust.
Vorschlag: Im Integrationstest `druckstation_repo.NewRepository(db)` verdrahten (wie tischgeschaeft) und Guard samt Hilfsmethode entfernen.
Aufwand: S · Status: unverifiziert (minor)

### backend/api/middleware/middleware.go

**[Benutzername stammt aus dem Token, Status und Rolle aus der DB]** (boundary-consistency)
Datei: backend/api/middleware/middleware.go:269-303
Warum: Nach einer Umbenennung schreiben Buchungen bis zu 12 Stunden den alten Namen unveränderlich ins Kassenjournal.
Vorschlag: `u.Username` statt des Claims in den Context legen; `middleware_test.go:31-33,512-540` nachziehen.
Aufwand: S · Status: unverifiziert (minor)

### backend/repository/betreiber_repo/repo.go

**[Domain-Zeitstempel wird auf dem Schreibpfad verworfen]** (code-smell)
Datei: backend/repository/betreiber_repo/repo.go:46-63
Warum: `UpsertBetreiber` reicht `UpdatedAt` nicht durch; das SQL setzt `NOW()`, anders als bei Produkt, Tisch und User.
Vorschlag: `UpdatedAt` als Parameter durchreichen — oder das Feld aus Konstruktor und Pflicht-Schema entfernen und die DB-Herkunft im Kommentar festhalten.
Aufwand: S · Status: unverifiziert (minor)

### backend/app/routes.go

**[Regel-18-Kommentare in Backend und Windows-Starter]** (convention, Regel 18)
Datei: backend/app/routes.go:33-49
Warum: „bisheriges Verhalten“, „frühere imperative Registrierung“ und „wie bisher“ beschreiben abgelöste Stände.
Vorschlag: Auf die Ist-Aussage kürzen; gleiche Behandlung für `backend/config/config.go:34`, `backend/repository/druckauftrag_repo/repo.go:139`, `backend/api/fiskal/dsfinvk/mapper.go:496`, `windows/starter/core/env.go:14,88`, `windows/starter/main.go:89`, `windows/starter/core/diagnose.go:39`, `reverse-proxy/caddyfile.go:6,69`, `packaging/windows/KURZANLEITUNG.md:63,138`.
Aufwand: S · Status: unverifiziert (minor)

### docs/language.md

**[Glossareintrag „Aufholphase“ widerspricht der Statusfunktion]** (docs-accuracy)
Datei: docs/language.md:472
Warum: Bereits signierte Vorgänge tragen in der Aufholphase den Nachsigniert-, nicht den Ausfallvermerk.
Vorschlag: Satz auf die tatsächliche Zurechnung umschreiben, Formulierung wie docs/handbuch.md:240.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/lib/errorMessages.ts

**[Fehlercode `login_throttled` hat keine deutsche Meldung]** (boundary-consistency)
Datei: frontend/src/lib/errorMessages.ts:14-100
Warum: Der 429-Cooldown fällt auf „Anmeldung fehlgeschlagen. Bitte erneut versuchen.“ zurück und fordert zum sofortigen Wiederholen auf.
Vorschlag: `login_throttled: 'Zu viele Fehlversuche. Bitte einen Moment warten und erneut versuchen.'` in `commonErrorMessages` ergänzen — analog `rate_limited`.
Aufwand: S · Status: unverifiziert (minor)

**[Toter Fehlercode `kassensturz_erforderlich`]** (code-smell)
Datei: frontend/src/lib/errorMessages.ts:51-52
Warum: Kein Backend-Pfad sendet den Code; er suggeriert einen separaten Kassensturz-Schritt, den es nicht gibt.
Vorschlag: Eintrag und die Zeile in errorMessages.test.ts:71 ersatzlos löschen.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/service/components/direktverkauf/DirektverkaufAbschluss.tsx

**[`byCode`-Overrides formulieren zentrale Meldungen um]** (code-smell)
Datei: frontend/src/service/components/direktverkauf/DirektverkaufAbschluss.tsx:81-86
Warum: Dieselbe Ursache erscheint je nach Bildschirm in zwei Wortlauten; die zentrale Tabelle verliert ihre Rolle.
Vorschlag: Reine Umformulierungen streichen (`kasse_nicht_geoeffnet` hier und in DirektverkaufStornoDrawer.tsx:65-66, `verkauf_not_found` in StornoDrawer:67 und DirektverkaufHistorie.tsx:72); `stornierung_not_found` nach errorMessages.ts verschieben.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/service/components/ServiceDock.tsx

**[Regel-18-Kommentare in der Frontend-Schicht]** (convention, Regel 18)
Datei: frontend/src/service/components/ServiceDock.tsx:4-7
Warum: „ersetzt die zwei früher schwebenden Leisten“ und die Phase-3-Referenz zeigen auf einen gelöschten Plan.
Vorschlag: Als Präsens-Beschreibung neu schreiben; gleiche Behandlung für `table/Zahlung.tsx:40`, `table/Bestellung.tsx:22`, `direktverkauf/Direktverkauf.tsx:16-18`, `admin/finanzamt/LaeuftAllesSection.tsx:44`, `admin/.../UebersichtStatusZeile.tsx:10`, `GeldtransitDialog.tsx:37-38`, `admin/users/Users.tsx:22-23`, `admin/users/UserRolle.tsx:7-8`.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/admin/users/Users.tsx

**[Kommentare verweisen auf einen nicht existierenden „Design-Handoff 1e“]** (docs-accuracy)
Datei: frontend/src/admin/users/Users.tsx:22-23
Warum: Der in ADR 06 genannte Pfad `docs/prds/design_handoff_spektral_redesign/` liegt nicht im Repo; 17 Kommentare zitieren ihn.
Vorschlag: Herkunftsangabe aus den Kommentaren entfernen und nur die inhaltliche Begründung behalten; die Nennung in ADR 06 bleibt unangetastet.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/admin/tables/Tisch.ts

**[Frontend-Namensschemata trimmen nicht, die zog-Gegenstücke schon]** (convention, Regel 5)
Datei: frontend/src/admin/tables/Tisch.ts:6-9
Warum: „ ab “ besteht die Client-Prüfung und scheitert danach als anonymer `validation_error`-Toast.
Vorschlag: `.trim()` in Tisch.ts:6, `admin/users/User.ts:31` und `lib/identity.ts:9` ergänzen — wie in `admin/products/Produkt.ts:48-52`.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/admin/finanzamt/BetreiberBackend.ts

**[Pflichtfelder ohne clientseitige Mindestlänge]** (convention, Regel 5)
Datei: frontend/src/admin/finanzamt/BetreiberBackend.ts:8-15
Warum: Leere Pflichtfelder laufen bis zum Backend und kommen als generischer Toast ohne Feldmarkierung zurück.
Vorschlag: `.trim().min(1, …)` für vereinsname, strasse, plz und ort ergänzen — passend zum zog-Schema `Min(1).Required()`.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/routes.ts

**[Autorisierungs-Guards des Frontends sind ungetestet]** (test-quality)
Datei: frontend/src/routes.ts:16-44
Warum: Eine vertauschte Rollenabfrage in `AdminGuard` oder `ServiceGuard` würde von keinem Test bemerkt.
Vorschlag: Drei Fälle in routes.test.ts ergänzen: Service-Token an `AdminGuard` → Redirect '/', Admin-Token → undefined, ohne Token an `ServiceGuard` → Redirect '/'.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/components/common/UserDropdown.tsx

**[Abmelden leert den react-query-Cache nicht]** (correctness)
Datei: frontend/src/components/common/UserDropdown.tsx:42-45
Warum: Beim Helferwechsel auf demselben Gerät zeigen benutzerbezogene Ansichten bis zum Refetch die Daten des Vorgängers.
Vorschlag: Gemeinsamen `logout()`-Helfer mit `queryClient.clear()` einführen (oder wie der 401-Pfad voll neu laden) und in UserDropdown.tsx:42-45 sowie AdminSidebar.tsx:186-189 nutzen.
Aufwand: S · Status: unverifiziert (minor)

### frontend/src/components/common/LoginForm.tsx

**[Startseite nach Login wird an zwei Stellen entschieden]** (code-smell)
Datei: frontend/src/components/common/LoginForm.tsx:40-44
Warum: Das Formular kennt nur die Admin-Rolle; Service-Logins laufen über zwei zusätzliche Loader-Hops.
Vorschlag: Die Rollen-Zuordnung nur in routes.ts halten und im Formular nach `validateAndSetToken` auf '/' navigieren.
Aufwand: S · Status: unverifiziert (minor)

### README.md

**[Relay-Schnelltest nennt den falschen Statuscode]** (docs-accuracy)
Datei: README.md:78
Warum: Ein ungültiger Token liefert 400 mit `{"code":"unauthorized"}`, nicht 401.
Vorschlag: „`401` bei ungültigem“ → „`400` mit `{"code":"unauthorized"}` bei ungültigem“; den Handler nicht anfassen.
Aufwand: S · Status: unverifiziert (minor)

### docs/leitfaden/aktualisieren.md

**[Versionsfixierte Änderungsprosa im Evergreen-Leitfaden]** (docs-accuracy, Regel 18)
Datei: docs/leitfaden/aktualisieren.md:58-127
Warum: Die 0.17.1/0.17.3-Aussagen veralten mit jedem Tag; die Relay-Aussage ist nach dem Toolchain-Update bereits falsch.
Vorschlag: Zeitlose Regel formulieren (UI lädt sich bei Versionsdifferenz neu; Relay neu starten, wenn das ZIP eine neue `jotti-relay.exe` enthält) und die 0.17.x-Notizen nach CHANGELOG.md verschieben.
Aufwand: M · Status: unverifiziert (minor)

### docs/leitfaden/self-hosting.md

**[Copy-Paste-Block pinnt `JOTTI_VERSION=v0.14.0`]** (docs-accuracy)
Datei: docs/leitfaden/self-hosting.md:46-54
Warum: Ein gültiges altes Tag installiert stillschweigend eine drei Releases alte Version.
Vorschlag: Platzhalter `JOTTI_VERSION=vX.Y.Z` setzen und den Verweis auf die Releases-Seite behalten.
Aufwand: S · Status: unverifiziert (minor)

### scripts/prod-update.sh

**[Tote Referenz auf `docs/leitfaden.md` an fünf Stellen]** (docs-accuracy)
Datei: scripts/prod-update.sh:111
Warum: Die Datei wurde in `docs/leitfaden/*.md` aufgeteilt; der Verweis erscheint auf den Fehlerpfaden des Updates.
Vorschlag: Je Zielseite verlinken: prod-update.sh:111 und .env.example:16 → `aktualisieren-backups.md`, prod-backup.sh:132 → `datenaufbewahrung.md`, Makefile:231 → `betriebsarten.md`, ops-smoke.sh:34 → `self-hosting.md`; Historien-Halbsatz in `website/src/lib/published-docs.ts:11` streichen.
Aufwand: S · Status: unverifiziert (minor)

### database/migrations/README.md

**[README verspricht ein automatisches Zurückspielen des Backups]** (docs-accuracy)
Datei: database/migrations/README.md:12
Warum: `prod-update.sh` spielt nie zurück; es druckt den Dump-Pfad und zwei manuelle Schritte und bricht ab.
Vorschlag: Satz auf den Ist-Stand ziehen: Das Skript bricht ab und nennt den Dump samt den zwei Schritten für `prod-restore.sh`.
Aufwand: S · Status: unverifiziert (minor)

### .github/workflows/ci.yml

**[Upgrade-Gate pinnt eine zwei Releases alte Vorversion]** (ops)
Datei: .github/workflows/ci.yml:406-417
Warum: `PREVIOUS_VERSION: v0.17.1` prüft nicht den realen Upgrade-Pfad; kein Workflow-Schritt hebt den Pin an.
Vorschlag: Auf `v0.17.3` anheben und „(aktuell `v0.17.1`)“ aus `database/migrations/README.md:44` entfernen, damit die Version an einer Stelle steht.
Aufwand: S · Status: unverifiziert (minor)

### Makefile

**[`.env.example` liegt undokumentiert im Windows-Release-ZIP]** (code-smell)
Datei: Makefile:143
Warum: Der Windows-Stack interpoliert keinen der enthaltenen Schlüssel; die Datei lädt zu wirkungslosen Änderungen ein.
Vorschlag: Die Zeile `cp .env.example "$(RELEASE_DIR)/"` aus `release-windows` entfernen; der Self-Hosting-Pfad bezieht die Datei aus dem Quellarchiv.
Aufwand: S · Status: unverifiziert (minor)

### docker-compose.local.yml

**[Header widerspricht sich selbst und nennt eine nicht existierende Doku-Sektion]** (docs-accuracy)
Datei: docker-compose.local.yml:9-10
Warum: Die Zeilen 3-7 beschreiben denselben Stack korrekt als WLAN-Betrieb mit Bondruck.
Vorschlag: Zeilen 9-10 streichen; wenn ein Verweis gewünscht ist, `docs/leitfaden/installation.md` ohne Sektionsnamen verlinken.
Aufwand: S · Status: unverifiziert (minor)

### windows/starter/core/adminmarker.go

**[Konsolenmeldung schickt in eine Endlos-Neustartschleife]** (correctness)
Datei: windows/starter/core/adminmarker.go:49-53
Warum: Nach abgeschlossener Einrichtung liefert `EnsureInitialAdmin` `ActionSkip` — es erscheint nie wieder ein Code.
Vorschlag: Die Meldung wie die Statusseite konditionieren („Falls die Ersteinrichtung noch offen ist …“) und die fixierte Erwartung in adminmarker_test.go:54 nachziehen.
Aufwand: S · Status: unverifiziert (minor)

### windows/starter/system.go

**[Port-Preflight deckt nur 80 und 443 ab]** (boundary-consistency)
Datei: windows/starter/system.go:112-126
Warum: Der Release-Stack veröffentlicht auch 127.0.0.1:8484; dort scheitert der Start mit einem englischen Docker-Fehler.
Vorschlag: Die Schleife auf 80, 443 und 8484 erweitern; die `reverseProxyRunning`-Abkürzung bleibt.
Aufwand: S · Status: unverifiziert (minor)

## Verworfene Befunde

Diese Behauptungen wurden geprüft und fallengelassen — nicht erneut aufwerfen.

- `backend/api/kasse/tischgeschaeft/application/command.go:89-112` — OCC-Schreibhelfer existiert dreifach (`writeEventOCC`, `writeVersionedEvent`, inline in `kassenfuehrung`), DRY-Verstoß.
- `backend/api/kasse/kassenfuehrung/application/command.go:283-471` — `KasseAbschliessen` als 189-Zeilen-Funktion verletzt Single Responsibility.
- `backend/api/kasse/tischgeschaeft/application/command_test.go:167-169` — rund 2900 Zeilen Unit-Tests werden nie gelintet, 37 `errorlint`-Verstöße unbemerkt.
- `backend/api/kasse/tischgeschaeft/http/command_handler.go:66-71` — `Menge` hat auf keiner Schicht eine Obergrenze, während der Preis daneben `LTE(99999)` trägt.
- `backend/api/fiskal/signatur/tse_signatur_worker.go:205-226` — `(*sql.Conn).Close()` gibt den session-gebundenen Advisory Lock nicht frei, weil pgx einen No-op-`ResetSession` installiert.
- `backend/api/druck/bondruck/application/arbeitsbon_policy.go:90-99` — Bonmodus `pro_stueck` erzeugt einen Druckauftrag je Stück, in derselben Transaktion und ohne Mengenobergrenze.
- `backend/api/auth/http/command_handler.go:51-52` — die Login-Antwort unterscheidet Kontozustände und widerspricht der dokumentierten generischen Fehlermeldung (`docs/handbuch.md:455`).
- `backend/api/auth/http/command_handler.go:48-59` — `user_inactive` und `no_password_set` verraten vor jedem Passwortvergleich, ob das Konto existiert.
- `backend/domain/steuer/steuer.go:44-63` — `Aufteilen` liefert für unbekannte Steuersätze und für `brutto < 0` still `nil`, ohne dokumentierten Vertrag.
- `backend/domain/steuer/steuer.go:26-29` — das exportierte `SteuersatzSchema` wird paketübergreifend in place mutiert (`.Required()` in `domain/produkt` und in der HTTP-Schicht).
- `backend/repository/tse_repo/repo.go:91-102` — `MarkOffeneAlsNichtKonfiguriert` fegt ohne Cut-off und außerhalb jeder Transaktion; dazwischen eingereihte Aufträge werden terminal markiert.
- `backend/repository/tisch_repo/repo.go:156-173` — der Soft-Delete prüft den offenen Saldo nicht innerhalb seiner Transaktion; ein gelöschter Tisch kann `saldo_cents > 0` behalten.
- `backend/app/routes.go:127-140` — `POST /test/reset-and-seed` truncatet unauthentifiziert alle Datentabellen inklusive Kassenjournal, gesperrt nur durch eine Env-Var.
- `backend/bootstrap/bootstrap.go:78-104` — `EnsureInitialAdmin` reaktiviert einen deaktivierten Admin nicht; der dokumentierte Recovery-Pfad greift dort nicht.
- `frontend/src/service/components/direktverkauf/DirektverkaufAbschluss.tsx:60-104` — `verkaufId` rotiert nur beim Übergang leer→gefüllt; ein Retry mit geändertem Warenkorb verwirft die Änderung still.
- `frontend/src/service/table/hooks.ts:74-90` — der Query-Default erfindet Domänendaten, der Hook gibt keinen Fehlerzustand heraus.
- `frontend/src/service/product/Produkt.ts:7-55` — zweite, bereits gedriftete Kopie des Admin-Produktmodells für dasselbe Backend-DTO.
- `frontend/src/admin/tse/TSEEinrichtungWizard.tsx:72-934` — eine Datei trägt 14 Komponenten und die gesamte Setup-Zustandsmaschine, `apiKey`/`apiSecret` über fünf Ebenen gereicht.
- `frontend/src/admin/components/AdminPageHeader.tsx:5-6` — 16 Quellkommentare zitieren ein „Design-Handoff"-Dokument, das es im Repo nicht gibt.
- `frontend/src/admin/users/NewUserDialog.tsx:66-75` — alle vier Anlage-Dialoge erfinden clientseitig vollständige Entitäten (Status, Zeitstempel, Saldo), die niemand liest.
- `frontend/src/admin/settings/DruckstationBackend.ts:91-193` — das HTTP-Client-Modul trägt deutsche UI-Texte und Formatierung und bricht den Model/Backend-Schnitt.
- `frontend/src/admin/products/Produkt.ts:28-33` — ein Backend-Enum (`EntityStatus`) hat vier Frontend-Fassungen, zwei davon in derselben Datei benutzt.
- `frontend/src/components/common/FormFields.tsx:14-19` — die geteilte Komponentenschicht importiert Domänenwerte aus `@/admin/…`.
- `website/nginx.conf:33-43` — `/pagefind/` und `/fonts/` werden als „content-hashed" ewig gecacht, obwohl Manifest und Loader feste Namen tragen.
- `website/src/lib/anfrage-mailto.ts:82-105` — die Annahme-Mail nennt den Sitz der Organisation nicht, den die verbindliche TERMS-Vorlage verlangt.
- `website/src/lib/anfrage-mailto.ts:77-91` — das TERMS-Versionsdatum „7. September 2026" ist ein zweites und drittes Mal hartkodiert, ohne Kopplung an `TERMS.md`.
- `website/src/components/LiveDemo.tsx:196-207` — die Demo zeigt die Kategorie „Speisen", die es nicht gibt, und filtert die Produktliste nicht.
- `e2e/helpers/fehlerpfade.ts:3-6` — Kommentar und Doc versprechen POST-Filterung, `page.route('**/api/**')` filtert nur nach Pfad.
- `e2e/website/browser.mjs:1-19` — der `CHROMIUM_EXECUTABLE`-Fallback stammt aus einem fremden Deployment-Kontext, der Hinweis steht doppelt.
- `windows/relay/main.go:166-183` — die Poll-Query beansprucht Druckaufträge nicht; scheitert `meldeErgebnis`, druckt der nächste Poll dieselben Aufträge erneut.
- `windows/starter/main.go:213-217` — `0o600` schützt die `.env` unter Windows nicht, die ACL erbt von `C:\ProgramData` (gleiches gilt für State- und Dump-Verzeichnis).
- `packaging/windows/KURZANLEITUNG.md:118-125` — der dokumentierte Recovery-Pfad kann eine gescheiterte Migration nicht zurücknehmen, weil die Compose-Tags auf die neue Version zeigen.
- `windows/starter/main.go:238` — die ASCII-Transliterationskonvention ist nur halb angewandt: neun Strings tragen Em-Dashes, einer einen Pfeil, einer ein „ü".
- `docker-compose.rocks.yml:1` — `name: jotti` kollidiert mit dem Prod-Stack; beide teilen Volume `jotti_postgres-data` und die Containernamen.
- `database/migrate/Dockerfile:21` — `POSTGRES_PASSWORD` wird ohne Prozent-Kodierung in die Migrations-URL interpoliert; reservierte Zeichen brechen `migrate`.
- `Makefile:78-87` — `make lint` (nur `go vet` + `goimports -l`) ist schwächer als das Merge-Gate, obwohl Regel 15 darauf verweist.
- `docs/leitfaden/aktualisieren.md:57-60` — die Ausnahme vom automatischen Neuladen ist an die Zielversion 0.17.3 geknüpft statt an die auf dem Gerät geladene Client-Version.
- `docs/leitfaden/veranstaltungstag.md:19-20` — „nimmt keine Zahlungen an" untertreibt: ohne offene Kassensitzung wird gar nichts gebucht, auch keine Bestellung.
- `docs/leitfaden/aktualisieren.md:57-122` — drei Abschnitte sind fest auf Release 0.17.3 verdrahtet und werden mit 1.0.0 falsch.
- `docs/compliance.md:296` — „Versionsstring konfigurierbar" ist falsch, der Wert ist eine Compile-Zeit-Konstante.
- `docs/plans/guide-manuelle-qa-v1.0.0.md:14,59,71` — drei Links zeigen auf das nicht existierende `plan-v1.0-release-blockers.md`, zwei davon tragend.
- `backend/api/druck/beleg/application/kassenbeleg_command.go:297` — die gedruckte „Bon-Nr" ist die Event-ID, während DSFinV-K denselben Vorgang über die Vorgangs-UUID identifiziert.
- `backend/api/fiskal/setup/application/command.go:30-45` — alle drei TSE-Konfigurationspfade prüfen nur „keine offene Kassensitzung", nicht offene Signaturaufträge.
- `.env.example:1-11` — das Windows-ZIP legt Platzhalter-Secrets genau dorthin, wo `core.ResolveEnv` sie ungeprüft adoptiert.
- `database/migrations/03_ausgabe_entfernen.up.sql:5-23` — die Migration leert `tisch_sessions` mit der falschen Begründung, der Backend-Start baue die Projektion neu auf; kein Produktionspfad tut das.
- `.github/workflows/ci.yml:404-417` — `PREVIOUS_VERSION: v0.17.1` ist über zwei Releases gedriftet, die behauptete Release-Mechanik zum Anheben existiert nicht.
- `.github/workflows/ci.yml:36-54` — kein CI-Job deckt `Makefile`, `docker-compose.*.yml`, `packaging/**`, `.env.example` oder `release.yml` ab; `docker-compose.release.yml` läuft nur beim Tag.
