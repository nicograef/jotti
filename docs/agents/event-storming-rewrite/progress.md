Details siehe plan.md

## Agent Anweisungen

> **Lies diese Anweisungen vollständig, bevor du mit der Arbeit beginnst.**

### Abschnitt beanspruchen

1. **Lies die gesamte progress.md** — insbesondere den Parallelisierungs-Abschnitt und alle Abschnitts-Überschriften.
2. **Finde den nächsten verfügbaren Abschnitt.** Ein Abschnitt ist verfügbar, wenn:
   - Er offene Tasks hat (`- [ ]`)
   - Er **nicht** mit 🔒 oder ✅ markiert ist
   - Seine Abhängigkeiten erfüllt sind (alle Vorgänger-Abschnitte sind ✅)
3. **Beanspruche den Abschnitt sofort**, indem du `🔒` an die Überschrift anhängst (`## Abschnitt N: Titel` → `## Abschnitt N: Titel 🔒`). Erst danach mit der Arbeit beginnen.
4. **Falls kein verfügbarer Abschnitt existiert: Stoppe sofort, ohne Änderungen vorzunehmen.** Erkläre dem User: welche Abschnitte noch offen sind, warum sie nicht bearbeitet werden können (🔒 = anderer Agent arbeitet daran, oder Abhängigkeiten noch nicht ✅), und welche Vorgänger-Abschnitte zuerst abgeschlossen werden müssen. **Führe keine Änderungen an Dateien durch.**

### Abschnitt abarbeiten

1. **Ein Task nach dem anderen.** Arbeite Tasks innerhalb des Abschnitts sequentiell ab — von oben nach unten.
2. **Sofort abhaken.** Ändere `- [ ]` zu `- [x]` in dieser Datei **unmittelbar** nachdem ein Task erfolgreich erledigt ist. Nicht erst am Ende des Abschnitts, nicht gebündelt — **nach jedem einzelnen Task**. Verwende beim Abhaken immer die **Abschnitts-Überschrift + den vollständigen Task-Text** als Kontext, damit die Ersetzung eindeutig ist.
3. **Abschnitt abschließen.** Wenn du an Code gearbeitet hast: Wenn alle Tasks eines Abschnitts `[x]` sind, führe die wichtigsten Dev-Scripte und CI-Steps lokal aus: compilation, build, linting, formatting, testing. Stelle sicher, dass es keine Fehler oder Warnings gibt. Erst dann ist der Abschnitt fertig. Wenn du an Dokumentation gearbeitet hast: Lese Korrektur, stelle sicher, dass alle Links funktionieren, und dass die Formatierung korrekt ist.
4. **✅ setzen.** Ersetze `🔒` durch `✅` in der Abschnitts-Überschrift (`## Abschnitt N: Titel 🔒` → `## Abschnitt N: Titel ✅`).
5. **Stoppen.** Nach Abschluss eines Abschnitts: **stopp**. Beginne nicht den nächsten Abschnitt, sondern melde, dass der Abschnitt abgeschlossen ist.
6. **Conventional Commit Message schreiben.** Wenn du an Code gearbeitet hast: Schreibe zu deinen Änderungen bzw. dem Abschnitt eine Conventional Commit Message. Führe kein Commit selbst durch, schreibe nur die Message in den Chat, sodass diese kopiert werden kann. Wenn du an Dokumentation gearbeitet hast: Schreibe eine passende Commit Message für die Dokumentationsänderungen.

---

## Parallelisierung

**Keine Parallelisierung möglich.** Alle Abschnitte schreiben in dieselbe Datei (`docs/event-storming-neu.md`) und der Inhalt jedes Abschnitts baut auf den vorherigen auf.

**Abhängigkeiten (strikt sequentiell):**

- Abschnitt 1 → muss zuerst abgeschlossen sein (Referenzdokumente lesen, Datei anlegen)
- Abschnitt 2 → nach Abschnitt 1 (Header, Einleitung, Teilnehmer)
- Abschnitt 3 → nach Abschnitt 2 (Phase 1 — Events einführen)
- Abschnitt 4 → nach Abschnitt 3 (Phase 2 — Events clustern)
- Abschnitt 5 → nach Abschnitt 4 (Phase 3 — Prozesse modellieren)
- Abschnitt 6 → nach Abschnitt 5 (Phase 4 — Software Design)
- Abschnitt 7 → nach Abschnitt 6 (Bounded Contexts)
- Abschnitt 8 → nach Abschnitt 7 (Hotspots)
- Abschnitt 9 → nach Abschnitt 8 (Ergebnisse)
- Abschnitt 10 → nach Abschnitt 9 (Anhänge)
- Abschnitt 11 → nach Abschnitt 10 (Validierung und Finalisierung — muss ganz am Ende kommen)

---

## Abschnitt 1: Vorbereitung

- [ ] `docs/anforderungen.md` lesen — alle Anforderungs-IDs notieren (K-01 bis K-13, S-01 bis S-03, A-01 bis A-03, Q-01 bis Q-08, R-01 bis R-06) sowie die bewusste Abgrenzung (§7)
- [ ] `docs/produktbeschreibung.md` lesen — Positionierung, Personas, Kernfeatures, Abgrenzung erfassen
- [ ] `docs/lizenz-und-nutzung.md` lesen — Lizenzmodell (AGPL-3.0 + Non-Commercial) und berechtigte Nutzer erfassen
- [ ] `docs/event-storming.md` lesen — als Strukturreferenz für Stil, Aufbau und Dialogform (nicht inhaltlich übernehmen)
- [ ] Leere Datei `docs/event-storming-neu.md` anlegen

## Abschnitt 2: Header, Einleitung und Setup

- [ ] Dokumenttitel und Einleitung schreiben (Event Storming nach Brandolini, Zweck der Simulation)
- [ ] Inhaltsverzeichnis erstellen (8 Hauptabschnitte + 2 Anhänge, analog zur Strukturreferenz)
- [ ] Setup-Abschnitt schreiben: Datum (Samstag, 11. März 2026), Ort (Vereinsheim Sportverein Grüntal), Dauer (~5 Stunden)
- [ ] Notation/Stickies-Legende schreiben (6 Farben: Event, Command, Aggregate, Policy, Read Model, Hotspot)
- [ ] Teilnehmertabelle schreiben (10 Personen mit Kürzel, Rolle, Person, Hintergrund): FAC (Lisa), DEV1 (Anna), DEV2 (Tim), DOM (Rudi), SRV1 (Jonas), SRV2 (Maria), SRL (Felix), ADM (Thomas), KAS (Eva), VER1 (Klaus)
- [ ] Prüfen: Keine entfernten Teilnehmer (SCR, ARC, POS, VER2) erwähnt

## Abschnitt 3: Phase 1 — Big Picture

- [ ] Phasen-Einleitung schreiben: FAC eröffnet Phase, erklärt Regeln und Timebox für chaotische Exploration
- [ ] Kassenbetrieb-Events K-01 bis K-06 im Dialog schreiben: Bestellung aufgeben, Zahlung registrieren, Lieferung bestätigen, Stornierung, Tischübersicht, Kassenjournal
- [ ] Kassenbetrieb-Events K-07 bis K-10 im Dialog schreiben: Bezeichnung pro Bestellung, Bestellungen umbuchen, Rückgeldberechnung, Tisch-Schnellsuche
- [ ] Bondruck- und Ausgabestations-Events im Dialog schreiben: Bondruck inkl. Freibon (K-11), Küchendisplay/KDS (K-12), Zubereitungsstatus (K-13)
- [ ] Stammdaten-Events im Dialog schreiben: Produktverwaltung (S-01), Tischverwaltung (S-02), Benutzerverwaltung (S-03)
- [ ] Auth-Events im Dialog schreiben: Login (A-01), Passwort setzen (A-02), Logout (A-03)
- [ ] Reporting-Events im Dialog schreiben: Tagesabrechnung (R-01), Datenexport (R-02), Abrechnung pro Tisch (R-03), pro Servicekraft (R-04), Produktumsatz (R-05), Tagesabschluss (R-06)
- [ ] Querschnitt-Events im Dialog schreiben: Q-01 bis Q-08 (z.B. Mobile-first, Offline, Datenschutz etc.)
- [ ] Abgrenzungs-Diskussion im Dialog schreiben: Kartenzahlung, TSE/KassenSichV, Reservierungen, Warenwirtschaft, Gast-Benachrichtigung etc.
- [ ] Prüfen: Alle 6 neuen Events aus dem Plan enthalten — KDS angezeigt, Position „in Zubereitung", Position „fertig", Produktumsatz ausgewertet, Abrechnung pro Tisch erstellt, Abrechnung pro Servicekraft erstellt
- [ ] Prüfen: Diskussion über technische vs. fachliche Events enthalten (Auth/Sessions nicht mit Domänen-Events vermischen)

## Abschnitt 4: Phase 2 — Clustering und Pivot Points

- [ ] Phasen-Einleitung schreiben: FAC leitet Sortierung und Gruppierung ein
- [ ] Timeline schreiben: 3 Phasen (Vorbereitung → Betrieb → Abschluss) — Events entlang der Zeitachse ordnen
- [ ] Pivot Points identifizieren und schreiben: Veranstaltung eröffnet, Bestellung aufgegeben, Tischkonto ausgeglichen, Tagesabschluss
- [ ] Cluster A (Stammdaten) schreiben: Produkte, Tische, Benutzer, Auth
- [ ] Cluster B (Kassenbetrieb) schreiben: K-01 bis K-06 plus K-07 (Bezeichnung), K-09 (Rückgeld), K-10 (Schnellsuche)
- [ ] Cluster C (Stornierung & Umbuchung) schreiben: K-04, K-08
- [ ] Cluster D (Ausgabestationen & Bons) schreiben: K-11 (Bondruck/Freibon), K-12 (KDS), K-13 (Zubereitungsstatus)
- [ ] Cluster E (Abrechnung/Reporting) schreiben: R-01 bis R-06

## Abschnitt 5: Phase 3 — Process Modelling

- [ ] Phasen-Einleitung schreiben: FAC erklärt Command-Event-Policy-Notation, Teilnehmer modellieren Prozesse
- [ ] 4.1 Bestellung aufnehmen (K-01): Command → Event-Flow im Dialog, inkl. Akteur (SRV1/SRV2), Bezeichnung pro Bestellung (K-07)
- [ ] 4.1 Lieferung bestätigen (K-03): Command → Event-Flow im Dialog, inkl. Akteur und Abschluss-Logik
- [ ] 4.1 Zahlung registrieren (K-02): Command → Event-Flow im Dialog, Teilzahlung hervorheben, Rückgeldberechnung (K-09)
- [ ] 4.1 Stornierung (K-04): Command → Event-Flow im Dialog, SRL-Berechtigung, Policy-Diskussion (wer darf stornieren, wann)
- [ ] 4.2 Produktverwaltung (S-01): CRUD-Flow im Dialog (Anlegen, Bearbeiten, Deaktivieren — Soft-Delete)
- [ ] 4.2 Tischverwaltung (S-02): CRUD-Flow im Dialog
- [ ] 4.2 Benutzerverwaltung (S-03): CRUD-Flow im Dialog (Rollen: admin, senior_service, service)
- [ ] 4.3 Tagesabrechnung / Umsatzübersicht (R-01): Flow im Dialog schreiben
- [ ] 4.3 Datenexport (R-02): Flow im Dialog schreiben
- [ ] 4.3 Abrechnung pro Tisch (R-03): Flow im Dialog schreiben
- [ ] 4.3 Abrechnung pro Servicekraft (R-04): Flow im Dialog schreiben
- [ ] 4.3 Produktumsatz-Reporting (R-05): Flow im Dialog schreiben
- [ ] 4.3 Tagesabschluss (R-06): Flow im Dialog schreiben — Diskussion über offene Tische und Voraussetzungen
- [ ] 4.4 Bon-Druck (K-11): Flow im Dialog — automatisch nach Bestellung, getrennt nach Kategorie (Küche/Getränke)
- [ ] 4.4 Freibon (Teil von K-11): Flow im Dialog — freie Preiseingabe, Bon ohne reguläres Produkt
- [ ] 4.4 KDS / Küchendisplay (K-12): Flow im Dialog — Echtzeit-Anzeige offener Bestellungen, gruppiert nach Tisch
- [ ] 4.4 Zubereitungsstatus (K-13): Flow im Dialog — Positionen als „in Zubereitung" / „fertig" markieren, VER1 (Klaus) bringt Ausgabe-Perspektive ein

## Abschnitt 6: Phase 4 — Software Design

- [ ] Phasen-Einleitung schreiben: FAC leitet Software-Design-Phase ein, Teilnehmer identifizieren Aggregate und Read Models
- [ ] 5.1 Aggregate identifizieren: Dialog — aus Clustern Aggregate ableiten, Invarianten benennen, Events zuordnen
- [ ] 5.1 Event-Sourcing vs. CRUD: Dialog — Diskussion, was sich für welchen Bereich eignet (z.B. Tisch-Operationen vs. Stammdaten)
- [ ] 5.2 Event-Typen benennen: Dialog — Teilnehmer legen Konventionen fest (Sprache, Format, Vergangenheitsform), benennen alle Event-Typen selbst
- [ ] 5.2 Fachliche vs. technische Events: Dialog — klare Unterscheidung, welche Events zur Domäne gehören und welche zur Infrastruktur
- [ ] 5.3 Read Models / Projektionen — Tischübersicht, Tischdetails, Kassenjournal: Dialog über benötigte Ansichten für Servicekräfte
- [ ] 5.3 Read Models — Reporting-Ansichten (R-01 bis R-06): Dialog über benötigte Ansichten für KAS (Eva)
- [ ] 5.3 Read Models — Küchen-/Ausgabe-Ansichten (K-12, K-13): Dialog über benötigte Ansichten für VER1 (Klaus)
- [ ] 5.3 Diskussion: Wie Read Models aus Events aufgebaut werden (Projektionen, Event-Replay)

## Abschnitt 7: Bounded Contexts und Domain Map

- [ ] Phasen-Einleitung schreiben: FAC leitet Context-Identifikation ein
- [ ] Context-Identifikation: Dialog — welche fachlichen Bereiche haben eigene Sprache und eigene Regeln?
- [ ] Context-Klassifikation: Dialog — Core, Supporting oder Generic für jeden identifizierten Context
- [ ] Context Map: Dialog — wie kommunizieren die Contexts miteinander? (Upstream/Downstream, Shared Kernel, ACL etc.)
- [ ] ASCII-Diagramm der Context Map erstellen
- [ ] Prüfen: Contexts emergieren aus der Session, kein Bezug auf bestehende Architektur (`docs/design.md`, `docs/cqrs.md`)

## Abschnitt 8: Hotspots und offene Fragen

- [ ] Hotspot-Einleitung schreiben: FAC sammelt ungelöste Fragen aus den vorherigen Phasen
- [ ] Beizubehaltende Hotspots schreiben: Tischumbuchung (K-08, Nice-to-have), Freibon (K-11, Should-have), Bondruck/Drucker-Integration (K-11, Should-have), Offline-Fähigkeit (Q-05, Nice-to-have), Tagesabschluss mit offenen Tischen (R-06, Nice-to-have)
- [ ] Neue Hotspots schreiben: KDS-Architektur (K-12) — Echtzeit-Datenfluss zur Küche; Zubereitungsstatus-Modellierung (K-13) — eigene Events oder UI-State; Reporting-Aggregation (R-01 bis R-05) — wie Auswertungen aus Rohdaten berechnet werden
- [ ] Prüfen: Rückgeldberechnung (K-09) ist kein Hotspot (nur bei Phase 3 als Anzeige-Aspekt erwähnt)
- [ ] Prüfen: Keine alten Anforderungs-Nummern (#23–#40) im Abschnitt

## Abschnitt 9: Ergebnisse und nächste Schritte

- [ ] 5 Kernaussagen als gemeinsames Verständnis formulieren (Zusammenfassung der wichtigsten Erkenntnisse der Session)
- [ ] Priorisierte Erkenntnisse als Tabelle schreiben: Anforderungs-IDs, Beschreibung, Priorität (Must/Should/Nice-to-have)
- [ ] Ubiquitous Language dokumentieren: Alle Fachbegriffe, die in der Session entwickelt wurden, als eigenständige Terminologie-Liste auflisten
- [ ] Feedback der Teilnehmer schreiben: Kurze Zitate der 10 Teilnehmer (FAC, DEV1, DEV2, DOM, SRV1, SRV2, SRL, ADM, KAS, VER1)
- [ ] Prüfen: Keine Zitate von entfernten Teilnehmern (SCR, ARC, POS, VER2)

## Abschnitt 10: Anhänge

- [ ] Anhang A — Vollständige Event-Liste schreiben: Alle Domain Events aus der Session, mit zugehöriger Anforderungs-ID, gruppiert nach Cluster/Bounded Context
- [ ] Anhang B — Stickies-Legende schreiben: 6 Sticky-Farben mit Bedeutung (analog zur Strukturreferenz)

## Abschnitt 11: Validierung und Finalisierung

- [ ] Prüfen: Alle Anforderungs-IDs mindestens einmal referenziert — K-01 bis K-13, S-01 bis S-03, A-01 bis A-03, Q-01 bis Q-08, R-01 bis R-06
- [ ] Prüfen: Keine alten Anforderungs-Nummern (#23, #24, #25, … #40) im gesamten Dokument
- [ ] Prüfen: Keine entfernten Teilnehmer (SCR, ARC, POS, VER2) haben Dialogzeilen im gesamten Dokument
- [ ] Prüfen: Kein Bezug auf `docs/language.md`, `docs/design.md` oder `docs/cqrs.md`
- [ ] Prüfen: Kein Implementierungsstatus (✅/🔲) erwähnt
- [ ] Prüfen: Event-Typ-Namen, Aggregate und Bounded Contexts emergieren aus der Session (keine vorgegebenen Namen aus bestehenden Dokumenten)
- [ ] Prüfen: Bewusste Abgrenzung (§7 aus `docs/anforderungen.md`) vollständig und korrekt reflektiert
- [ ] Prüfen: Lizenzmodell korrekt referenziert (AGPL-3.0-or-later + Non-Commercial), falls im Dokument erwähnt
- [ ] Alte Datei `docs/event-storming.md` löschen
- [ ] `docs/event-storming-neu.md` umbenennen zu `docs/event-storming.md`
