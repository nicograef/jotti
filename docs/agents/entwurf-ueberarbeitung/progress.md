Details siehe plan.md

## Agent Anweisungen

> **Lies diese Anweisungen vollständig, bevor du mit der Arbeit beginnst.**

### Kontext laden (vor jedem Abschnitt)

Bevor du einen Abschnitt beanspruchst, lies **immer** diese Dateien:

1. `plan.md` (im selben Verzeichnis) — Gesamtplan, Kontext und Referenzen
2. Alle in plan.md genannten Referenzdateien, die für den Abschnitt relevant sind
3. Bereits erstellte/geänderte Dateien aus vorherigen Abschnitten (um nahtlos anzuknüpfen)

Diese Dateien werden in jeder neuen Session erneut gelesen — die Kontext-Beschaffung ist kein eigener Abschnitt, sondern Pflicht vor jeder Arbeit.

**Referenzdateien:**

| Dokument            | Pfad                            |
| ------------------- | ------------------------------- |
| Event Storming      | `docs/design/event-storming.md` |
| Anforderungen       | `docs/anforderungen.md`         |
| Produktbeschreibung | `docs/produktbeschreibung.md`   |
| Aktueller Entwurf   | `docs/design/entwurf.md`        |

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

**Alle Abschnitte schreiben in dieselbe Datei** (`docs/design/entwurf.md`). Daher ist **keine Parallelisierung möglich** — die Abschnitte müssen strikt sequentiell bearbeitet werden.

**Abhängigkeiten:**

- Abschnitt 1 → keine Vorgänger (erstellt die Datei mit Grundgerüst)
- Abschnitt 2 → muss nach Abschnitt 1 abgeschlossen sein
- Abschnitt 3 → muss nach Abschnitt 2 abgeschlossen sein
- Abschnitt 4 → muss nach Abschnitt 3 abgeschlossen sein
- Abschnitt 5 → muss nach Abschnitt 4 abgeschlossen sein
- Abschnitt 6 → muss nach Abschnitt 5 abgeschlossen sein
- Abschnitt 7 → muss nach Abschnitt 6 abgeschlossen sein (abschließender Review)

---

## Abschnitt 1: Grundgerüst, Systemvision und Bounded Contexts ✅

Referenzdateien: `docs/produktbeschreibung.md` (Abschnitte 1, 2, 7), `docs/design/event-storming.md` (Abschnitt 2.7, 6), aktueller `docs/design/entwurf.md` (Kapitel 1, 2).

- [x] Datei `docs/design/entwurf.md` mit dem vollständigen Inhaltsverzeichnis (alle 14 Kapitel mit Unterabschnitten als leere Platzhalter-Überschriften) neu erstellen — gemäß Plan Schritt 1
- [x] Kapitel 1.1 — Systemvision schreiben (aus aktuellem Entwurf übernehmen, da bereits sauber formuliert)
- [x] Kapitel 1.2 — Designziele schreiben (7 Designziele aus aktuellem Entwurf übernehmen)
- [x] Kapitel 1.3 — Bewusste Abgrenzung schreiben (9-Punkte-Abgrenzungsliste aus aktuellem Entwurf übernehmen)
- [x] Kapitel 2.1 — Kontextübersicht schreiben (5 Bounded Contexts: Kassenbetrieb, Stammdaten, Ausgabe, Abrechnung, Auth)
- [x] Kapitel 2.2 — Klassifikation schreiben (Core / Supporting / Generic Zuordnung)
- [x] Kapitel 2.3 — Context Map schreiben (Diagramm aus Event Storming Abschnitt 6.3 übernehmen)
- [x] Kapitel 2.4 — Beziehungen zwischen Kontexten schreiben (Beziehungstabelle mit ACL-Erklärung, Auth → alle vier Contexts einzeln)

## Abschnitt 2: Kassenbetrieb (Core Domain) ✅

Referenzdateien: `docs/design/event-storming.md` (Phase 4.1, 5.1, 5.2), `docs/anforderungen.md` (K-01…K-09).

- [x] Kapitel 3.1 — Tisch-Aggregat schreiben: zweistufige Zustandsmodellierung (Tisch → Bestellungen → Positionen) mit allen Feldern gemäß Plan Schritt 4
- [x] Kapitel 3.2 — Invarianten schreiben: Saldo-Formel, Liefer-/Bezahl-/Stornierungsinvarianten, erweiterte Stornierungsinvariante (auch bezahlte Positionen stornierbar, negativer Saldo dokumentieren), Rolleninvariante, Mindestmengen-Invariante
- [x] Kapitel 3.3 — Domain Events schreiben: `BestellungAufgegeben` mit Fat-Event-Daten, opt. Bezeichnung (K-07), opt. Kommentar
- [x] Kapitel 3.3 — Domain Events ergänzen: `ProdukteGeliefert` (Referenz auf Positionen, opt. Kommentar)
- [x] Kapitel 3.3 — Domain Events ergänzen: `ZahlungRegistriert` (bezahlte Positionen, Betrag in Cent, opt. Kommentar)
- [x] Kapitel 3.3 — Domain Events ergänzen: `ProdukteStorniert` (stornierte Positionen, Stornobetrag in Cent, opt. Kommentar)
- [x] Kapitel 3.3 — Domain Events ergänzen: `FreibonAusgestellt` (NEU — freie Bezeichnung, Preis in Cent, opt. Kommentar)
- [x] Kapitel 3.3 — Gemeinsame Event-Metadaten dokumentieren: event_id, tisch_id, benutzer_id, benutzer_name, zeitstempel, version
- [x] Kapitel 3.4 — Zustandsberechnung schreiben: Event-Replay-Algorithmus (Snapshot laden → Events ab Snapshot-Version → sequentiell Apply)
- [x] Kapitel 3.5 — Snapshot-Strategie schreiben: separate technische Speicherung (NICHT im Event Stream), Argumente für Entscheidung, verworfene Alternative (Snapshot als Event) kurz erwähnen
- [x] Kapitel 3.6 — Policies schreiben: Stornierung nur senior_service/admin (K-04), automatischer Bon-Druck nach Kategorie (K-11), Umbuchung (K-08) als Cross-Aggregat-Transaktion mit Verweis auf offene Fragen

## Abschnitt 3: Supporting und Generic Sub-Domains ✅

Referenzdateien: `docs/design/event-storming.md` (Phase 4.2, 4.3, 4.4, 5.1, 5.3), `docs/anforderungen.md` (S-01…S-03, A-01…A-03, K-11…K-13).

- [x] Kapitel 4.1 — Produkt-Aggregat schreiben: id, name, kategorie (food/beverage/other), status, varianten[] mit Invarianten (Name nicht leer, Preis > 0, Soft-Delete)
- [x] Kapitel 4.2 — Tisch-Stammdaten-Aggregat schreiben: id, name, status; Trennung von Tisch-Aggregat im Kassenbetrieb erklären
- [x] Kapitel 4.3 — Benutzer-Aggregat schreiben: id, name, benutzername (unique), passwort_hash, rolle, muss_passwort_setzen, status; Invarianten
- [x] Kapitel 4.4 — Persistenzstrategie (CRUD mit Soft-Delete) schreiben: Begründung, warum kein Event-Sourcing nötig (Fat Events decken historische Stammdaten ab)
- [x] Kapitel 5.1 — Bondruck schreiben: Policy (BestellungAufgegeben → Bon nach Kategorie), Bon-Inhalt, Freibon als Sonderfall, Fire-and-Forget + Retry, Nachdruck
- [x] Kapitel 5.2 — Küchendisplay (KDS) schreiben: Read Model für Echtzeit-Anzeige offener Positionen nach Kategorie/Tisch, offene Architekturentscheidung (Polling vs. SSE vs. WebSockets → H4)
- [x] Kapitel 5.3 — Zubereitungsstatus schreiben: Workflow (offen → in Zubereitung → fertig), offene Modellierungsentscheidung (Events im Tisch-Aggregat vs. eigenes Aggregat vs. transienter State → H5)
- [x] Kapitel 5.4 — Offene Architekturentscheidungen (Ausgabe) schreiben: Hotspots H3, H4, H5 mit Pro/Contra aus Event Storming
- [x] Kapitel 6.1 — Read Models und Projektionen (Abrechnung) schreiben: Tagesabrechnung (R-01), pro Tisch (R-03), pro Servicekraft (R-04), Produktumsatz (R-05)
- [x] Kapitel 6.2 — Aggregationsstrategie schreiben: On-the-fly vs. vorberechnete Projektionen → Hotspot H8, Argumentation für on-the-fly bei 500–2000 Events
- [x] Kapitel 6.3 — Tagesabschluss schreiben: Prozess (offene Tische prüfen → Abschlussbericht → Reset), Verweis auf Veranstaltungskonzept und Hotspot H7
- [x] Kapitel 6.4 — Datenexport schreiben: CSV-Export (Umsätze, Bestellungen, Artikeldaten), Admin-only
- [x] Kapitel 6.5 — Offene Architekturentscheidungen (Abrechnung) schreiben: Hotspots H7, H8
- [x] Kapitel 7.1 — Authentifizierung schreiben: JWT-basiert, Passwort-Hashing (Argon2id), generische Fehlermeldungen, Token-Gültigkeit 12h
- [x] Kapitel 7.2 — Autorisierung schreiben: Drei Rollen (admin, senior_service, service) mit Berechtigungsmatrix aus anforderungen.md
- [x] Kapitel 7.3 — Sicheres Onboarding schreiben: Einmalpasswort (6-stellig) → Eigenes Passwort bei Erstanmeldung

## Abschnitt 4: Read Models und Persistenzstrategie ✅

Referenzdateien: `docs/design/event-storming.md` (Phase 5.1, 5.3).

- [x] Kapitel 8.1 — Service-Ansichten schreiben: Tischübersicht, Tischdetails, Produktkatalog (NEU), Kassenjournal — jeweils mit Datenquelle, Inhalt, Akteure
- [x] Kapitel 8.2 — Admin-Ansichten (Reporting) schreiben: Tagesabrechnung, Abrechnung pro Tisch, pro Servicekraft, Produktumsatz — jeweils mit Datenquelle, Inhalt, Akteure
- [x] Kapitel 8.3 — Ausgabe-Ansichten schreiben: KDS-Ansicht, Zubereitungsstatus — jeweils mit Datenquelle, Inhalt, Akteure
- [x] Kapitel 9.1 — Zwei Strategien, eine Datenbank schreiben: Übersichtstabelle (Kassenbetrieb → Event-Sourcing, Stammdaten/Auth → CRUD)
- [x] Kapitel 9.2 — Event Store schreiben: Append-only-Prinzip, Optimistic Concurrency Control (UNIQUE tisch_id + version), keine SQL-DDL
- [x] Kapitel 9.3 — Stammdaten (CRUD-Prinzipien) schreiben: Soft-Delete, Timestamps, referenzielle Integrität, keine SQL-DDL
- [x] Kapitel 9.4 — Optimistic Concurrency Control schreiben: Version-Mechanismus und Retry-Pattern bei Konflikten erklären
- [x] Kapitel 9 — Snapshot-Speicherung ergänzen: separate Speicherung (nicht im Event Stream), automatisch nach N Events oder auf Admin-Anfrage

## Abschnitt 5: Architekturprinzipien und Infrastruktur ✅

Referenzdateien: aktueller `docs/design/entwurf.md` (Kapitel 5–10), `docs/anforderungen.md` (Q-01…Q-08).

- [x] Kapitel 10.1 — Schichtenarchitektur schreiben: 4-Schichten-Diagramm (HTTP → Application → Domain → Repository/Infra), kurze Beschreibung jeder Schicht, Request-Lebenszyklus-Beispiel (Bestellung aufgeben)
- [x] Kapitel 10.2 — API-Design-Prinzipien schreiben: POST-only, JSON, JWT-Auth, Rollenprüfung in Middleware, Fehlerformat, Bereichsgliederung (Auth/Admin/Service/Senior Service) — keine vollständige Endpunkt-Liste
- [x] Kapitel 10.3 — Frontend-Architektur schreiben: Mobile-first SPA, Guards, Seitenstruktur nach Bereichen, UI-Patterns (Karten, Drawer, Tabs), BackendClient-Interface — keine Framework-Versionen
- [x] Kapitel 10.4 — Validierung schreiben: Frontend-Schema + Backend-Schema, Backend = Single Source of Truth
- [x] Kapitel 10.5 — Geldbeträge schreiben: Immer Cent (Integer), durchgehend DB/Backend/API/Frontend/Events
- [x] Kapitel 10.6 — Mehrbenutzerfähigkeit schreiben: Paralleler Zugriff verschiedene Tische (kein Konflikt), gleicher Tisch (Optimistic Concurrency)
- [x] Kapitel 10.7 — Mobile-first schreiben: ≥360px, Touch-optimiert, Drawer, kein Hover, kein App-Download
- [x] Kapitel 10.8 — Sicherheit schreiben: HTTPS, Rate Limiting, Security Headers, Input-Validierung, Passwort-Hashing, keine Secrets im Code
- [x] Kapitel 11.1 — Architekturüberblick schreiben: Docker-Compose-Diagramm (nginx, Backend, PostgreSQL, Frontend, Migrate-Container)
- [x] Kapitel 11.2 — Deployment-Modell schreiben: Self-hosted, VPS/Raspberry Pi, Let's Encrypt, keine externe Abhängigkeit

## Abschnitt 6: Ubiquitous Language, Priorisierung und Offene Fragen

Referenzdateien: `docs/design/event-storming.md` (Abschnitte 7, 8.2, 8.3).

- [ ] Kapitel 12 — Ubiquitous Language schreiben: Kassenbetrieb-Begriffe (Tisch, Bestellung, Position, Lieferung, Zahlung, Stornierung, Saldo, Kassenjournal, Bezeichnung, Kommentar, Freibon)
- [ ] Kapitel 12 — Ubiquitous Language ergänzen: Stammdaten-Begriffe (Produkt, Variante, Kategorie, Preis, Soft-Delete)
- [ ] Kapitel 12 — Ubiquitous Language ergänzen: Auth-Begriffe (Rolle, Einmalpasswort, Token)
- [ ] Kapitel 12 — Ubiquitous Language ergänzen: Ausgabe-Begriffe (Bon, KDS, Zubereitungsstatus, Ausgabestation)
- [ ] Kapitel 12 — Ubiquitous Language ergänzen: Abrechnung-Begriffe (Tagesabrechnung, Umsatz, Stornoquote, Tagesabschluss, Export)
- [ ] Kapitel 12 — Ubiquitous Language ergänzen: Übergreifende Begriffe (Event-Sourcing, Fat Event, ACL, Append-only, BYOD)
- [ ] Kapitel 13.1 — Stufe 1 (Must-have / MVP) schreiben: Anforderungs-IDs aus Event Storming Abschnitt 8.2, keine Implementierungsreihenfolge
- [ ] Kapitel 13.2 — Stufe 2 (Should-have) schreiben: Anforderungs-IDs zuordnen
- [ ] Kapitel 13.3 — Stufe 3 (Nice-to-have) schreiben: Anforderungs-IDs zuordnen
- [ ] Kapitel 14 — Offene Entwurfsfragen schreiben: Tabelle mit allen 8 Hotspots (H1–H8) — Thema, Kernfrage, Priorität, erweitert um konkrete offene Fragen aus den jeweiligen Hotspot-Abschnitten

## Abschnitt 7: Review und Querverweise

Referenzdateien: `docs/anforderungen.md`, `docs/design/event-storming.md`, fertiger `docs/design/entwurf.md`.

- [ ] Prüfen, dass jede Anforderung (K-01…K-13, S-01…S-03, A-01…A-03, Q-01…Q-08, R-01…R-06) im Entwurf adressiert ist — fehlende Verweise ergänzen
- [ ] Prüfen, dass alle 5 fachlichen Domain Events aus dem Event Storming (BestellungAufgegeben, ProdukteGeliefert, ZahlungRegistriert, ProdukteStorniert, FreibonAusgestellt) vollständig abgebildet sind
- [ ] Prüfen, dass alle Read Models aus dem Event Storming (Phase 5.3) im Kapitel 8 aufgeführt sind
- [ ] Querverweise zwischen Kapiteln auf Konsistenz prüfen (Begriffs-Verwendung, Kapitel-Nummern, Hotspot-Referenzen)
- [ ] Abgleich: Ubiquitous Language (Kapitel 12) enthält alle Begriffe, die im restlichen Dokument verwendet werden
- [ ] Korrektur-Durchgang: Tippfehler, Markdown-Formatierung, Überschriften-Hierarchie, Dokumentlänge prüfen (Ziel: 800–1100 Zeilen)
