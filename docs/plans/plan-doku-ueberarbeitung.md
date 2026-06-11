# Plan: Dokumentations-Überarbeitung (docs/ ohne plans/ und prds/)

> Source PRD: n/a (Task-Beschreibung + Klärungsrunden vom 2026-06-11)

## Goal

Alle zehn Dokumente unter `docs/` (ohne `plans/` und `prds/`) verschlanken, entdoppeln und konsistent machen: weniger Text, eine einzige Quelle pro Fakt in den Entwickler-Dokumenten, eigenständig lesbare Betreiber-Leitfäden, korrekte Implementierungsstatus-Angaben. Zielgröße: von ~4.220 Zeilen auf ~2.650–2.800 Zeilen (≈ 35 %; das ursprüngliche 40–50 %-Ziel ist durch die eingefrorenen §§-Texte und die Arbeitsartefakte anforderungen.md/language.md begrenzt — Akzeptanzkriterien sind inhaltlich, Zeilenziele indikativ).

## Architectural decisions

Dauerhafte Strukturentscheidungen, die für alle Phasen gelten:

- **Flache Struktur:** Entwickler-Dokumente bleiben flach unter `docs/` (Tooling-Deep-Links in AGENTS.md und `.github/instructions/*` bleiben stabil). `docs/betrieb/` bleibt unverändert. **Kein** `docs/recht/`-Ordner: Mit den Nutzungsbedingungen in `TERMS.md` bliebe dort nur eine Datei übrig.
- **Nutzungsbedingungen → `TERMS.md` (Repo-Root):** Der vollständige Inhalt von `docs/nutzungsbedingungen.md` ersetzt den bisherigen 7-Zeilen-Stub. Root-Datei = dauerhaft stabile URL. Alte Deep-Links auf `docs/nutzungsbedingungen.md` werden 404 — akzeptiert (Stand-Datum pinnt die Fassung, Git-History bewahrt sie). Kein Redirect-Stub.
- **`lizenz-und-nutzung.md` → `docs/lizenzmodell.md`:** Stark gekürzt; erklärt nur noch das Lizenzmodell (Eigentum, Source-Available, Berechtigte, Forks, Dual Licensing, CLA). Alles Vertragliche lebt in `TERMS.md`, dorthin wird nur verlinkt.
- **Dedup-Modell:** Entwickler-Dokumente haben pro Fakt genau eine Quelle + Querverweise. Die beiden Leitfäden in `betrieb/` bleiben eigenständig lesbar (Laien-Zielgruppe) inkl. eigener Mini-Glossare — bewusste, begrenzte Redundanz.
- **Kanonische Zuständigkeiten:**
  - `handbuch.md` — Architektur, Aggregate, Events, Invarianten, Berechtigungsmatrix
  - `language.md` — Begriff↔Code-Mappings, Namenskonventionen (keine Architektur-Erklärungen)
  - `anforderungen.md` — Anforderungs-IDs, Status, Akzeptanzkriterien
  - `compliance.md` — Rechtsnormen, Entwickler-/Betreiberpflichten (normativ), TSE/DSFinV-K/ELSTER-Fachregeln
  - `steuerrecht.md` — USt-Sätze und -Regeln
  - `produktbeschreibung.md` — Produktidentität (mit ehrlichem Status)
  - `betrieb/leitfaden-*.md` — Laienfassung, standalone
- **Keine TOCs:** Inhaltsverzeichnisse in allen Dokumenten entfernen (GitHub rendert die Outline automatisch). Ausnahme: keine — auch in `TERMS.md` ist das TOC kein Bestandteil der §§.
- **Handbuch = Architektur-Referenz:** Bounded Contexts, Aggregat-Grenzen, Event-Sourcing-Modell, Invarianten, Design-Entscheidungen, Berechtigungsmatrix. Keine Code-Listings, Feld-Schemata oder Betriebsanleitungen — kanonische Quelle dafür sind Code (`backend/domain/`), Migrationen (`database/migrations/`) bzw. README/`betrieb/`. Kapitelnummern dürfen sich ändern; referenzierende Deep-Links (`.github/instructions/*`, AGENTS.md) werden in derselben Phase angepasst.
- **`git mv` für Verschiebungen** (History-Erhalt). Für `TERMS.md`: erst Stub löschen, dann `git mv docs/nutzungsbedingungen.md TERMS.md`, dann anpassen.

## Inventory

In-Scope-Dokumente (Stand vor Überarbeitung):

| Datei | Zeilen | Befund |
| --- | ---: | --- |
| `docs/handbuch.md` | 923 | Kern-Referenz, dicht. TOC 44 Zeilen; Relay-Betriebsdetails in §4.6 (Z. 699–734) überdimensioniert; §3.10/§3.11 enthalten Rechts-Grundlagen, die zu compliance.md gehören; §5.2 nennt „min. 8 Zeichen" — Code sagt 6 (`backend/domain/user/password.go:15`) |
| `docs/language.md` | 579 | Veraltete Anmerkungen: `druckauftraege` „(geplant)" (Z. 479) — Tabelle existiert (`database/migrations/01_initial.up.sql`); Relay-Cursor-Hinweis (Z. 483) widerspricht §-Abweichungen-Abschnitt (Z. 37); Event `kassenbewegung-gebucht:v1` (Z. 303–323) — Code: `geldtransit-gebucht:v1` (`backend/domain/kasse/kassensitzung_events.go:14`). Vereinswesen-/Fiskal-Definitionen überlappen steuerrecht.md §8 und compliance.md §2 |
| `docs/anforderungen.md` | 343 | Bereits straff; A-02 „min. 6 Zeichen" ist korrekt (Code bestätigt) |
| `docs/compliance.md` | 700 | Größtes Dedup-Potenzial: Betreiberpflichten 3× (§2.7b Z. 94–100, §7.6 Z. 622–648, §8 Z. 652–667); Maihock-Beispiel 2× (§3.6 Z. 224–235, §6.5 Z. 486–495); datierte Grundsatzentscheidungs-Tabelle (Z. 27–39) dupliziert normative Aussagen; Beleg-Beispiel §5.3 |
| `docs/steuerrecht.md` | 178 | Fokussiert; kürzbar: Historischer Hintergrund (§1.1), Jahreswechsel 2025/2026 (§7.1); §6 Belegausweis überlappt compliance.md §5.2 |
| `docs/produktbeschreibung.md` | 281 | §6 Kernfeatures und §7.3 behaupten offene Features als vorhanden (TSE, DSFinV-K, Hash-Chain, KDS, Zubereitungsstatus, CSV-Export — alle laut anforderungen.md offen); Personas/Problem/Lösung redundant erzählt |
| `docs/lizenz-und-nutzung.md` | 341 | §§ 6–10 sind reine Zusammenfassungen mit „Verbindlich geregelt in"-Verweis; Lizenz-Historie (Z. 70–78); Anhänge A–C (Z. 291–341) = private Entwickler-Notizen |
| `docs/nutzungsbedingungen.md` | 257 | Quasi-vertraglicher §§-Text — **materiell unangetastet**; E-Mail-Vorlage (Z. 231) enthält GitHub-URL auf alten Pfad |
| `docs/betrieb/leitfaden-betreiber.md` | 271 | Gut; FAQ (§7) wiederholt §1/§2/Schritt-2-Inhalte; Link auf `../lizenz-und-nutzung.md` (Z. 213) |
| `docs/betrieb/leitfaden-hosting.md` | 250 | Gut; FAQ (§5) wiederholt §2-Tabelle; Umlaut-lose Schreibweisen („ausfuehren", „vollstaendige", Z. 85–101) |

Referenzierende Dateien (müssen bei Moves/Renames mitgezogen werden):

- `README.md:8,84,92,109` — Links auf compliance.md, leitfaden-betreiber.md, lizenz-und-nutzung.md, nutzungsbedingungen.md
- `TERMS.md:5` — Stub-Link auf docs/nutzungsbedingungen.md (wird ersetzt)
- `AGENTS.md:9,23–28` — Compliance-Hinweis + Dokumenten-Tabelle (Pfade bleiben gültig, Beschreibungen in Phase 7 auffrischen)
- `.github/instructions/{database,event-sourcing,frontend,backend}.instructions.md` — Deep-Links auf handbuch.md §3/§4/§6 und language.md
- `docker-compose.local.yml:8` — Kommentar-Verweis auf leitfaden-hosting.md (bleibt gültig)
- `docs/produktbeschreibung.md:272` — Links auf lizenz-und-nutzung.md + nutzungsbedingungen.md
- `docs/lizenz-und-nutzung.md` / `docs/nutzungsbedingungen.md` — wechselseitige Links

## Resolved decisions

Aus den Klärungsrunden (2026-06-11):

- Entwickler-Docs flach, `betrieb/` unverändert; Nutzungsbedingungen → `TERMS.md` (Root), kein `docs/recht/`, kein Redirect-Stub, 404 alter Links akzeptiert.
- §§-Text der Nutzungsbedingungen bleibt materiell unangetastet; nur Umfeld (TOC, Template-URL) wird angepasst.
- Aggressive Kürzung: Historie-Abschnitte, TOCs, redundante Begründungen/Beispiele/FAQs entfallen; tragende rechtliche Begründungen bleiben.
- Leitfäden bleiben standalone inkl. Mini-Glossaren; FAQ-Antworten werden in den Fließtext gefaltet, Rest entfällt.
- anforderungen.md und language.md behalten Rolle und Struktur (IDs, Status, Glossar, Ist/Soll) — nur Dedup + Straffung.
- produktbeschreibung.md: Zielbild bleibt, nicht implementierte Features werden inline markiert („in Entwicklung"), konsistent mit README und anforderungen.md.
- 7 Phasen, Ausführung per `/implement-plan`, ein Commit pro Phase.

## Open questions / Risks

- **Alte E-Mail-Links:** Bereits versendete Nutzungsvereinbarungs-Mails verweisen auf `docs/nutzungsbedingungen.md` → nach Phase 1 404. Bewusst akzeptiert (Entscheidung Nico, 2026-06-11).
- **Inhaltsverlust-Risiko bei aggressiver Kürzung:** Jede Phase listet im Commit-Body, welche Abschnitte ersatzlos entfallen sind, damit das Review gezielt prüfen kann.
- **Stand-Datum der Nutzungsbedingungen:** Der Move ändert den §§-Text nicht — das Stand-Datum (3. Juni 2026) bleibt unverändert. Nur die Template-URL wird aktualisiert.

---

## Phase 1: Recht-Cluster — TERMS.md und lizenzmodell.md

### Context

- `docs/nutzungsbedingungen.md` — vollständiger §§-Text + Prozess + E-Mail-Vorlagen (URL in Z. 231)
- `TERMS.md` — 7-Zeilen-Stub, wird ersetzt
- `docs/lizenz-und-nutzung.md` — 341 Zeilen, §§ 6–10 nur Zusammenfassungen, Anhänge A–C
- Referenzen: `README.md:109`, `docs/produktbeschreibung.md:272`, `docs/betrieb/leitfaden-betreiber.md:213`

### What to build

`docs/nutzungsbedingungen.md` per `git mv` nach `TERMS.md` verschieben (alten Stub vorher entfernen); §§-Text unverändert lassen, nur: TOC entfernen, GitHub-URL in der E-Mail-Vorlage auf `https://github.com/nicograef/jotti/blob/main/TERMS.md` ändern, interne Links anpassen. `docs/lizenz-und-nutzung.md` per `git mv` nach `docs/lizenzmodell.md` verschieben und auf den nicht-vertraglichen Kern kürzen: Eigentum, Source-Available-Erklärung (eine Tabelle), Berechtigte/Nicht-Berechtigte, Fork-/Weitergabe-Verbot, Dual Licensing, CLA. Entfallen: Lizenz-Historie, §§ 6–10-Zusammenfassungen (ersetzt durch einen Verweisblock auf TERMS.md), „Warum"-Begründungsabschnitte auf je 1–2 Sätze verdichtet, Anhänge A–C auf einen kompakten „Hinweise für den Entwickler"-Abschnitt (~15 Zeilen) reduziert. Alle Links repo-weit aktualisieren (README, produktbeschreibung, leitfaden-betreiber).

### Acceptance criteria

- [x] `TERMS.md` enthält die vollständigen Nutzungsbedingungen; §§ 1–13 textidentisch zur Vorversion (Diff zeigt nur Move, TOC-Entfall, URL/Link-Anpassungen)
- [x] `docs/nutzungsbedingungen.md` existiert nicht mehr
- [x] `docs/lizenzmodell.md` ≤ ~150 Zeilen; jede Aussage steht entweder dort oder in TERMS.md, nicht in beiden
- [x] `grep -rn "nutzungsbedingungen.md\|lizenz-und-nutzung.md" .` (ohne .git, docs/plans, docs/prds) liefert keine Treffer mehr
- [x] README §Lizenz verlinkt auf `docs/lizenzmodell.md` und `TERMS.md`

---

## Phase 2: compliance.md und steuerrecht.md

### Context

- `docs/compliance.md` — Betreiberpflichten 3× (§2.7b, §7.6-Tabelle, §8); Maihock-Beispiel 2× (§3.6, §6.5); Grundsatzentscheidungen (Z. 27–39); Beleg-Beispiel §5.3
- `docs/steuerrecht.md` — §1.1 Historie, §7.1 Jahreswechsel, §6 überlappt compliance §5.2
- `docs/handbuch.md §3.13` — Gegenstück der Architektur-Querverweise

### What to build

compliance.md auf ~400–450 Zeilen konsolidieren: Betreiberpflichten in **einem** Abschnitt (§8 als normative Quelle mit der Verantwortlichkeits-Tabelle; §2.7b und §7.6 werden zu Kurzverweisen). Grundsatzentscheidungs-Tabelle auflösen — noch gültige Entscheidungen in die jeweiligen Fachabschnitte einarbeiten, Datierung entfällt. Maihock-/Festzelt-Beispiel nur einmal (in §3.6, §6.5 verweist darauf). Beleg-Beispiel §5.3 behalten (tragend für Durchbedienen-Pflicht), aber §5.5/§5.6 straffen. TOC entfernen. Quellenverzeichnis behalten. steuerrecht.md auf ~140–150 Zeilen: Historischer Hintergrund und §7.1 Jahreswechsel 2025/2026 entfernen (Gutschein-Übergangsregel in §5.1 bleibt — sie wirkt fort); §6 Belegausweis auf die steuerlichen Pflichtangaben beschränken und für Beleg-Struktur auf compliance.md verweisen (keine doppelte Pflichtangaben-Liste).

### Acceptance criteria

- [x] Betreiberpflichten stehen vollständig in genau einem compliance-Abschnitt; die anderen beiden Stellen sind ≤ 3-Zeilen-Verweise
- [x] Kein Inhalt doppelt zwischen steuerrecht.md §6 und compliance.md §5.2
- [x] Maihock-Beispieltabelle existiert genau einmal
- [x] compliance.md ≤ 450 Zeilen, steuerrecht.md ≤ 150 Zeilen
- [x] Alle Querverweise (anforderungen.md F-03/F-04, handbuch.md §3.13, leitfäden) zeigen weiterhin auf existierende Anker

---

## Phase 3: handbuch.md

### Context

- `docs/handbuch.md` — 923 Zeilen. TOC Z. 3–47; §3.6 Event-Feldtabellen (Z. 170–300) duplizieren die Go-Structs; §3.10 „rechtliche Grundlagen" (Z. 398–405) und §3.11 „Betreiber-Ablauf" (Z. 415–421) = Rechts-/Betreiberinhalte; §3.13 enthält Go-Code (TSEClient-Interface, TSEData-Struct) und zwei ASCII-Diagramme; §4.1–§4.4 Entitäts-Bäume duplizieren die Migrationen; §4.6 enthält Tabellen-Schemata und eine Relay-Betriebsanleitung (Start-Kommando, Env-Tabelle, TLS, Schnelltest, Z. 699–734); §5.2 Passwort „min. 8 Zeichen" (Z. 788) — Code: min. 6 (`backend/domain/user/password.go:15`)
- Zweck (→ Architectural decisions): Architektur-Referenz — übergeordnete Erklärungen, Architektur-Ansatz, Design-Entscheidungen, Invarianten. Was 1:1 im Code/in Migrationen steht, wird dort nachgeschlagen, nicht hier dupliziert.
- `.github/instructions/*.instructions.md` — Deep-Links auf §3/§3.4/§4/§6/§6.3; `database.instructions.md:6` verweist für das Event-Store-Schema bereits heute fälschlich auf §3.4 (richtig: §3.2)

### What to build

handbuch.md auf die Architektur-Referenz reduzieren (~500–550 Zeilen). Kapitelnummern dürfen sich ändern — alle Deep-Links werden im selben Commit nachgezogen. Im Einzelnen:

- **TOC entfernen** (GitHub-Outline genügt).
- **Code-/Schema-Duplikate entfernen:** Go-Blöcke (TSEClient, TSEData), ASCII-Diagramme (TSE-Integration, DSFinV-K-Exporter) und Feld-Bäume/Spalten-Tabellen, die 1:1 im Code stehen (§3.2 Kassenjournal-Baum, §3.8 Projektions-Spalten, §4.1–§4.4 Entitäts-Bäume, §4.6 druckauftraege/druckstationen-Schemata), durch kurze Prosa mit den architektonisch tragenden Eigenschaften ersetzen (z. B. `UNIQUE(subject, version)` für OCC, `kassensitzung_nr` für Cross-Stream-Queries) und auf Code/Migrationen verweisen.
- **§3.6 Domain Events:** Feldtabellen entfernen; kompakte Event-Übersicht bleibt (Event-Typ, Subject, Semantik, tragende Constraints wie Pflicht-Kommentare) mit Verweis auf `backend/domain/kasse/*_events.go` als kanonische Schema-Quelle.
- **§3.10/§3.11:** Rechtliche Grundlagen und Betreiber-Ablauf durch Verweise auf compliance.md ersetzen; technische Invarianten (Zwei-Event-Muster, z_nr-Regeln, Tisch-Saldo-Sperre) bleiben — ggf. zu einem Abschnitt zusammenführen.
- **§3.13:** Mapping-Tabelle jotti-Vorgang → TSE-Transaktion (Atomares Modell) bleibt als Design-Entscheidung; Rest auf Prosa + Verweise (compliance.md §3, Code) verdichten.
- **§4.6:** Relay-Betriebsanleitung (Start-Kommando, Env-Tabelle, TLS-Hinweise, Schnelltest, Drucker-Hardware) ersatzlos streichen (README deckt den Start ab); bleiben: Zwei-Familien-Tabelle (Arbeitsbon/Kassenbeleg), Outbox-Konzept, „Relay = Transport" in 2–3 Sätzen.
- **§5.2:** Passwort-Mindestlänge auf 6 Zeichen korrigieren.
- **Dedup:** Fat Events und OCC je einmal kanonisch erklären (§2.2 bzw. §6.6), übrige Stellen verweisen nur.
- **Deep-Links nachziehen:** `.github/instructions/*` und AGENTS.md gegen die neuen Kapitelnummern prüfen und korrigieren (inkl. des bereits falschen §3.4-Verweises in database.instructions.md).

### Acceptance criteria

- [x] Keine Go-Code-Blöcke, keine Spalten-/Feld-Schemata und keine ASCII-Architektur-Diagramme mehr im Handbuch; Schema-Fragen verweisen auf Code/Migrationen
- [x] §5.2 nennt min. 6 Zeichen (konsistent mit Code und anforderungen.md A-02)
- [x] Keine Rechts-Grundlagen-Aufzählungen und kein Betreiber-Ablauf mehr (nur Verweise auf compliance.md)
- [x] Alle Deep-Links aus `.github/instructions/*` und AGENTS.md zeigen auf existierende Kapitel des neuen Stands
- [x] Berechtigungsmatrix (§5.1), Invarianten und Event-Übersicht (Typ, Subject, Semantik) bleiben vollständig erhalten
- [x] handbuch.md ≤ 550 Zeilen (Ergebnis: 480)

---

## Phase 4: language.md und anforderungen.md

### Context

- `docs/language.md` — veraltete Stellen: Z. 303–323 (`kassenbewegung-gebucht:v1` → Code: `geldtransit-gebucht:v1`, nur `einlage`/`entnahme`), Z. 479 (`druckauftraege` „geplant" → existiert), Z. 483 (Relay-Cursor-Hinweis → Relay ist Transport), Z. 37 (Widerspruch zum Cursor-Hinweis); Vereinswesen-Definitionen (Z. 56–65) überlappen steuerrecht.md §8; Fiskal-Begriffe (Z. 506–514) überlappen compliance.md §2
- `docs/anforderungen.md` — bereits straff; Status-Symbole und IDs werden extern referenziert
- Code-Referenzen: `backend/domain/kasse/kassensitzung_events.go:14`, `database/migrations/01_initial.up.sql`
- Nachtrag aus Phase 3: language.md Z. 116 und Z. 269 behaupten, die Event-Feldschemata seien kanonisch in handbuch.md §3.6 — kanonisch ist seit Phase 3 der Code (`backend/domain/kasse/*_events.go`); Verweise entsprechend umstellen

### What to build

language.md auf ~430–470 Zeilen bringen: Alle veralteten Anmerkungen gegen den Code verifizieren und korrigieren (Geldtransit-Event-Name und -Felder, Druckauftrag-Persistenz, Relay-Status; dabei prüfen, ob Privatentnahme/Privateinlage im Code existieren oder als „geplant" zu markieren sind). Begriffseinträge, die nur handbuch-Inhalte nacherzählen, auf Mapping + 1-Satz-Definition + Verweis reduzieren (z. B. Direktverkauf-Stornierung Z. 147 dupliziert handbuch §3.3). Vereinswesen-Sphären auf Kurzdefinitionen mit Verweis auf steuerrecht.md §8 kürzen; Gesetzes-Definitionen (AO, KassenSichV, GoBD, BSI) auf je einen Satz mit Verweis auf compliance.md. anforderungen.md nur leicht anfassen: Querverweis-Konsistenz (F-03/F-02-Verweise auf compliance/handbuch), keine ID- oder Strukturänderungen; Beschreibungstexte straffen, wo sie handbuch-Sätze wiederholen.

### Acceptance criteria

- [x] Alle Code-Mappings in language.md stimmen mit dem Code überein (Stichproben: Geldtransit-Event, druckauftraege, Relay, TSEClient-Status)
- [x] Kein Begriffseintrag erklärt Architektur, die handbuch.md bereits beschreibt — nur Mapping + Verweis
- [x] Widerspruch zwischen Abweichungs-Abschnitt und Einzeleinträgen beseitigt
- [x] anforderungen.md: alle IDs und Akzeptanzkriterien unverändert vorhanden _(bewusste Abweichung: F-02- und F-03-Status nach Code-Verifikation auf ✅ korrigiert — TSE-Integration inkl. fiskaly-Adapter, Beleg-TSE-Feldern, Admin-Warnung und Nachsignier-Outbox ist implementiert; die Plan-Prämisse „F-02 offen" war veraltet)_
- [x] language.md ≤ 470 Zeilen (Ergebnis: 470)

---

## Phase 5: produktbeschreibung.md

### Context

- `docs/produktbeschreibung.md` — §6 Kernfeatures-Tabelle (Z. 140–182) und §7.3 (Z. 222–237) nennen offene Features als vorhanden; Abgleich: anforderungen.md (K-13, K-15, R-02, R-05, F-04, F-08 offen); README.md:8 als Vorbild für Status-Formulierung
- Nachtrag aus Phase 4: F-02 (TSE) und F-03 (Kassenbeleg) sind im Code umgesetzt und in anforderungen.md jetzt ✅ — TSE und Beleg **nicht** mehr als „in Entwicklung" markieren. Offen bleiben: K-13, K-15, K-23, Q-05, R-02, R-05, F-04, F-05, F-08, F-09.

### What to build

produktbeschreibung.md auf ~160–190 Zeilen kürzen und ehrlich machen: In §6 und §7.1/§7.3 jede Aussage gegen anforderungen.md abgleichen und offene Features inline markieren („_(in Entwicklung)_"), implementierte unmarkiert lassen. Kernfeatures-Tabelle thematisch gruppieren und von 40 auf ~25 Zeilen verdichten (Einzelfeatures wie Tisch-Schnellsuche unter Sammelzeilen). Personas auf je Name + 2 Zeilen kürzen; §4 Problemstellung und §5 Lösung zu einem Abschnitt zusammenführen (Problem→Lösung-Tabelle, die 4.1/4.2/5.1 heute dreifach erzählen). §7.1-Vergleichstabelle auf die differenzierenden Zeilen kürzen. TOC entfernen. §9 Links auf `lizenzmodell.md` und `TERMS.md` (aus Phase 1) prüfen.

### Acceptance criteria

- [ ] Jedes in §6/§7 genannte offene Feature trägt eine Status-Markierung; keine Behauptung widerspricht anforderungen.md
- [ ] Problem und Lösung stehen in einem Abschnitt ohne inhaltliche Wiederholung
- [ ] produktbeschreibung.md ≤ 190 Zeilen
- [ ] Positioning Statement (§2.1) und USPs (§8) bleiben erhalten (Kernidentität)

---

## Phase 6: Betriebs-Leitfäden

### Context

- `docs/betrieb/leitfaden-betreiber.md` — FAQ §7 (Z. 230–254) wiederholt §1/§2/Schritt 2; Glossar §8 bleibt
- `docs/betrieb/leitfaden-hosting.md` — FAQ §5 (Z. 206–230) wiederholt §2-Tabelle; Umlaut-lose Schreibweisen Z. 85–101 („ausfuehren", „vollstaendige", „Energiespar-"); Glossar §6 bleibt

### What to build

Beide Leitfäden bleiben vollständig standalone (keine Verweise in Entwickler-Dokumente nötig, um sie zu verstehen). FAQs auflösen: Antworten mit neuem Informationsgehalt in den passenden Abschnitt falten (z. B. „Smartphones nicht melden" steht schon in Schritt 2 — entfällt; „Was kostet uns das?" → in „Das Wichtigste in 60 Sekunden"), rein wiederholende Fragen ersatzlos streichen. TOCs entfernen. Im Hosting-Leitfaden Umlaut-Schreibweisen normalisieren („ausführen", „vollständige"). Glossare und der ⏳-Status-Hinweis (Betreiber-Leitfaden §1) bleiben unverändert erhalten.

### Acceptance criteria

- [ ] Keine FAQ-Abschnitte mehr; kein Informationsverlust (jede vorher nur im FAQ stehende Aussage hat eine neue Stelle)
- [ ] Beide Leitfäden ohne Öffnen anderer Dokumente verständlich; Mini-Glossare vorhanden
- [ ] Keine Umlaut-losen Schreibweisen mehr im Hosting-Leitfaden
- [ ] leitfaden-betreiber.md ≤ 230 Zeilen, leitfaden-hosting.md ≤ 215 Zeilen
- [ ] Der „noch nicht fiskalkonform"-Warnhinweis (§1) bleibt unverändert

---

## Phase 7: Abschlusspass — Referenzen, Links, Bilanz

### Context

- `AGENTS.md:23–28` — Dokumenten-Tabelle („Wann lesen")
- `README.md`, `.github/instructions/*.instructions.md`, `docker-compose.local.yml:8`
- Alle in Phasen 1–6 geänderten Dokumente

### What to build

AGENTS.md-Dokumententabelle auffrischen: Pfade/Beschreibungen an den neuen Zustand anpassen (lizenzmodell.md/TERMS.md ergänzen, falls für Agenten relevant; Beschreibungs-Spalte gegen die gekürzten Inhalte prüfen). Repo-weiter Link-Check: alle relativen Markdown-Links in `docs/`, `README.md`, `TERMS.md`, `AGENTS.md`, `.github/` auf existierende Dateien **und Anker** prüfen (Anker ändern sich durch entfernte TOCs nicht, aber durch entfallene Abschnitte). Abschlussbilanz im Commit-Body: Zeilen vorher/nachher pro Datei und gesamt.

### Acceptance criteria

- [ ] AGENTS.md-Tabelle entspricht dem Ist-Zustand der Dokumente
- [ ] Kein toter relativer Link und kein toter Anker in README.md, TERMS.md, AGENTS.md, docs/, .github/
- [ ] `grep -rn "docs/nutzungsbedingungen\|lizenz-und-nutzung\|docs/recht" .` (ohne .git und docs/plans, docs/prds) ist leer
- [ ] Gesamtbilanz dokumentiert: Ausgangswert 4.221 Zeilen (in-scope), Zielkorridor ≤ ~2.800 erreicht oder Abweichung begründet
