# Plan: Praxis-Feedback der Vereine umsetzen

> Source PRD: n/a (Praxis-Feedback aus E-Mails von Vereinen, April bis September 2026;
> die Auswertung selbst liegt bewusst außerhalb des Repos)

## Goal

Die Rückmeldungen aus den ersten produktiven Einsätzen vollständig in Doku, Website und
Software einarbeiten und danach v1.0.0 veröffentlichen: bestätigte Bedienprobleme beheben,
Lücken in den Anleitungen schließen, Erwartungen an den Nutzungsprozess auf der Website
richtigstellen, die zugrunde liegenden Anforderungen hinter den Wünschen mit eigenen
Lösungen abdecken und festhalten, was jotti bewusst nicht tut.

## Architectural decisions

- **Kategorien bleiben ein festes Enum** (`essen`, `getraenk`, `sonstiges`). Sie steuern die
  Druckstationen-Zuordnung; freie Kategorien würden Druck-Konfiguration und DSFinV-K nach
  sich ziehen.
- **Bonmodus** bekommt den dritten Wert `pro_stueck`, zulässig nur für die Druckstation
  `abholbon`. Arbeitsbons der Produktstationen kennen ihn nicht.
- **Schema-Änderungen nur additiv** als neue Migration (Freeze-Disziplin); die
  CHECK-Bedingung auf `druckstationen.bonmodus` wird in einer neuen Migration ersetzt,
  `01_initial.up.sql` bleibt unverändert.
- **TSE bleibt Cloud-TSE (fiskaly)**, bis die ADR aus Phase 8 etwas anderes entscheidet.
- **Belegausgabe bleibt Papier** über den Kassenbeleg-Drucker oder die Befreiung nach
  § 146a Abs. 2 Satz 2 AO; ein elektronischer Beleg ist Nicht-Ziel.
- **Website und Leitfaden bleiben eine Quelle**: Leitfaden-Seiten werden aus `docs/`
  gelesen (`website/src/content.config.ts`), Änderungen passieren nur in `docs/`.
- **Release-Reihenfolge**: erst alle Phasen, dann v1.0.0. Vereine mit Einsatz im September
  arbeiten mit v0.17.3, das produktionsreif ist.

## Inventory

- `docs/leitfaden/installation.md — ## Bondruck einrichten (optional)` — nur generische
  ESC/POS-Angabe, keine Modelle, kein USB-Hinweis
- `packaging/windows/KURZANLEITUNG.md` — Absatz „Danach läuft jotti auch ohne Internet"
  widerspricht der TSE-Internetpflicht in `docs/leitfaden/haeufige-fragen.md`
- `docs/leitfaden/fehlersuche.md — ## Router-Hinweise` — FRITZ!Box-Rezept ohne den
  Hinweis, dass die Ausnahme nach Router-Neustart einmal Internet braucht
- `docs/leitfaden/haeufige-fragen.md` — keine Antwort zu Stromausfall, Geräteanzahl,
  Helfer-Verzehr, Drucker-Notwendigkeit, Größenordnung der TSE-Kosten
- `docs/leitfaden/belege-steuersaetze.md` — Befreiung von der Aushändigung beschrieben
- `website/src/pages/fuer-vereine.astro` + `website/src/components/AnfrageFormular.tsx`
  — Formular öffnet `mailto:` per JS-Navigation; Erfolgs-State ohne Installationslink;
  kein Kopier-Fallback, wenn kein Mailprogramm reagiert
- `website/src/lib/anfrage-mailto.ts — buildMailtoUrl()` mit Tests in
  `website/src/lib/anfrage-mailto.test.ts`
- `TERMS.md` — deutsches Recht, kein Hinweis auf Österreich/Schweiz
- `frontend/src/components/common/VariantNamePreis.tsx — VariantNamePreis()` —
  einzeiliges `truncate`; ähnliche Variantennamen werden auf dem Handy identisch
- `frontend/src/service/components/table/ProductList.tsx — ProductList()` —
  Kategorie-Pills (seit vor v0.17.2 vorhanden), danach eine lange Liste ohne Sortierung
- `frontend/src/service/components/ServiceSplitLayout.tsx` — ab 1024 px feste Spalte
  für den Abschluss
- `backend/api/druck/bondruck/application/arbeitsbon_policy.go —
createAbholbonAuftraege()` — ein Auftrag je Position oder je Bestellung
- `backend/api/druck/bondruck/application/escpos/formatter.go — FormatPositionBon()`
  — druckt `Nx Bezeichnung`
- `backend/domain/druckstation/druckstation.go` — Bonmodus-Typ und Validierung
- `backend/api/druck/station/http/handler.go` — Druckstationen-Konfiguration
- `frontend/src/admin/settings/DruckstationBackend.ts — BonmodusSchema` und
  `frontend/src/admin/settings/DruckstationConfigPage.tsx` — Auswahl der Bonmodi
- `database/migrations/01_initial.up.sql` — CHECK auf `druckstationen.bonmodus`
- `docs/language.md — #### Abholbon` und `#### Bonmodus`; `docs/handbuch.md —
  ### 4.6 Bondruck: Arbeitsbon und Kassenbeleg (K-12)`
- `docs/compliance.md — ### 3.5 TSE-Varianten und Anbieter-Entscheidung` — Hardware-TSE
  als ausgeschlossen begründet
- `docs/anforderungen.md` — Funktionsumfang-Tabelle und Nicht-Ziele-Tabelle
- `docs/plans/plan-bondruck-ursachenklaerung.md` — offen
- `docs/plans/guide-manuelle-qa-v1.0.0.md` — offen; verweist auf die nicht existierende
  Datei `plan-v1.0-release-blockers.md`
- Offene externe PRs: #109 (Reihenfolge für Produkte/Varianten), #110 (Variantenname auf
  eigener Zeile), #111 (Produktebene über der Variantenliste)
- `e2e/playwright.config.ts` — Projekt mit `devices['Pixel 7']` für Handy-Viewport

## Resolved decisions

- **Vollständig umsetzen, dann releasen.** Alle Phasen landen vor v1.0.0; kein Datum
  gegenüber Vereinen.
- **Scroll-Problem** wird über Reihenfolge (#109) gelöst. Die Produktebene (#111) wird
  erst nach Feldeinsatz von #109 entschieden und vorher gegen das aktuelle Design geprüft:
  die Kategorie-Pills existieren bereits, der PR darf sie nicht duplizieren.
- **Mobile-Abschneiden** wird über Umbruch statt Kürzung gelöst; #110 ist der Kandidat.
- **Bon pro Stück** wird nicht 1:1 übernommen. Die Anforderung (mehrere Einheiten auf
  einmal kaufen, einzeln an der Theke einlösen) deckt der Abholbon-Modus `pro_stueck`.
- **Bon per E-Mail** wird nicht gebaut. Die Anforderung (Beleg ohne eigenen Drucker)
  beantwortet die FAQ mit Befreiung oder Theken-Drucker; elektronischer Beleg ist
  Nicht-Ziel, weil das Gast-Handy nicht im Vereins-LAN ist.
- **Vereinslogo auf dem Bon** und **Helferdeckel** sind Nicht-Ziele.
- **USB-/Hardware-TSE** bekommt einen Kosten- und Anbieter-Spike mit ADR, keine
  Implementierung. Das Nicht-Ziel in `docs/compliance.md` bleibt bis zur ADR.
- **Druckerliste** nennt im Einsatz bestätigte Modelle (Epson TM-T20IV per Ethernet,
  Sam4s H-Cube per WLAN) plus eine recherchierte Kaufempfehlung mit Preisklasse; USB-Drucker
  sind nicht unterstützt und werden so benannt.
- **Österreich/Schweiz**: Nutzung erlaubt, Klartext auf Website und in TERMS, dass nur die
  deutsche KassenSichV abgebildet ist.
- **Formular-Fallback** bleibt serverlos: der Mailtext wird auf der Seite gezeigt und ist
  kopierbar.

## Open questions / Risks

- Zwei Vereine setzen jotti in der zweiten Septemberhälfte produktiv ein, einer davon
  zweitägig mit zwei Druckern. Phase 6 sollte vorher abgeschlossen sein.
- Die externen PRs sind nach dem Repo-Stand vom 04.08. entstanden; Konflikte mit
  späteren Änderungen sind beim Review zu prüfen.

---

## Phase 1: Leitfaden-Korrekturen

**Depends on**: none

### Context

- `packaging/windows/KURZANLEITUNG.md` — falscher Offline-Satz
- `docs/leitfaden/fehlersuche.md — ## Router-Hinweise` — Router-Neustart
- `docs/leitfaden/installation.md — ## Bondruck einrichten (optional)` — Drucker
- `docs/leitfaden/haeufige-fragen.md` — neue Fragen
- `docs/leitfaden/belege-steuersaetze.md` — Befreiung von der Aushändigung

### What to build

Doku-Änderungen, die Fragen aus dem Feld ohne Rückfrage beantworten. Kurzanleitung:
„Zertifikat und Handy-Zugang laufen danach ohne Internet; die TSE braucht beim Fest
Internet." Fehlersuche: FRITZ!Box wendet die Rebind-Ausnahme nach Neustart erst mit
Internet an; ohne Internet die Fallback-Adresse nutzen. Installation: Abschnitt
„Bondrucker" mit den zwei bestätigten Modellen und Anschlussart, einer recherchierten
Kaufempfehlung (Modell, Anschluss, Preisklasse, Stand der Recherche) und dem Satz, dass
USB-Drucker nicht unterstützt werden. FAQ: Stromausfall (Server aus, Kasse aus, Daten
bleiben; Server und Router an eine USV oder Powerbank), wie viele Handys (bis 30 Helfer,
ein Handy pro Servicekraft), Helfer-Verzehr (ein Tisch pro Helfer, am Abend kassieren
oder stornieren), „Braucht ihr einen Drucker?" (Befreiung beantragen oder ein
Theken-Drucker für Belege auf Anforderung; Bon per E-Mail gibt es nicht).

### Acceptance criteria

- [ ] Kurzanleitung und FAQ widersprechen sich nicht mehr zur Internetfrage
- [ ] Fehlersuche nennt den Neustart-Fall und verweist auf die Fallback-Adresse
- [ ] Installation listet die zwei bestätigten Modelle, eine Kaufempfehlung mit
      Preisklasse und Recherchedatum, und schließt USB aus
- [ ] FAQ beantwortet Stromausfall, Geräteanzahl, Helfer-Verzehr, Drucker-Notwendigkeit
- [ ] `make check` grün, Website-Build (`website/`) rendert die geänderten Seiten

---

## Phase 2: Website — Prozess und Ausland klarstellen

**Depends on**: none

### Context

- `website/src/pages/fuer-vereine.astro` — Seite mit Formular
- `website/src/components/AnfrageFormular.tsx` — Erfolgs-State
- `TERMS.md` — Geltungsbereich

### What to build

Auf `/fuer-vereine` und im Erfolgs-State: „Es gibt keine Freigabe und keine Zugangsdaten.
Mit dem Absenden könnt ihr installieren" mit direktem Link auf die Installationsseite und
dem Hinweis, den Spam-Ordner zu prüfen. Auf `/fuer-vereine` und in `TERMS.md` ein Absatz:
jotti setzt die deutsche KassenSichV um; für Österreich (RKSV) und die Schweiz ist nichts
vorgesehen, die Nutzung ist erlaubt, die Konformität Sache des Vereins. Die
Rechtsform-Auswahl bekommt keinen Länder-Eintrag; der Absatz reicht.

### Acceptance criteria

- [ ] Erfolgs-State enthält Installationslink und Spam-Hinweis
- [ ] `/fuer-vereine` und `TERMS.md` tragen den Absatz zu Österreich/Schweiz
- [ ] TERMS-Fassungsdatum und der Fassungsbezug in `buildMailtoUrl()` sind identisch
- [ ] Website-Build grün

---

## Phase 3: Website — Formular-Fallback ohne Mailprogramm

**Depends on**: none

### Context

- `website/src/lib/anfrage-mailto.ts — buildMailtoUrl()` — erzeugt Betreff und Text
- `website/src/components/AnfrageFormular.tsx` — öffnet `mailto:` per Navigation
- `website/src/lib/anfrage-mailto.test.ts` — bestehende Tests

### What to build

Nach dem Absenden zeigt die Seite zusätzlich Empfänger, Betreff und den vollständigen
Mailtext in einem Textfeld mit „Kopieren"-Button. Geräte ohne Mailprogramm (iOS ohne
Mail-App, Browser ohne Handler) können den Text so in jedes Postfach einfügen. Ein
reines Logik-Modul liefert die drei Teile getrennt, damit der Test sie ohne DOM prüft.

### Acceptance criteria

- [ ] Nach dem Absenden sind Empfänger, Betreff und Text sichtbar und kopierbar
- [ ] Test: die getrennten Teile entsprechen dem Inhalt der `mailto:`-URL
- [ ] Der bisherige `mailto:`-Weg bleibt unverändert

---

## Phase 4: Service-UI — Variantennamen auf dem Handy lesbar

**Depends on**: none

### Context

- `frontend/src/components/common/VariantNamePreis.tsx — VariantNamePreis()` — `truncate`
- `frontend/src/service/components/table/ProductList.tsx — VariantRow()` — Zeile mit
  Stepper
- PR #110 — Kandidat
- `e2e/playwright.config.ts` — Projekt `Pixel 7`

### What to build

Variantennamen dürfen nie so gekürzt werden, dass zwei Varianten desselben Produkts
gleich aussehen. Der Name bricht auf zwei Zeilen um; der Preis steht darunter oder
rechts, der Stepper behält seine Breite. Gilt für Tisch-Bestellung und Direktverkauf.
PR #110 wird dagegen reviewt: passt er, wird er übernommen; sonst wird der Umbruch
direkt in `VariantNamePreis` umgesetzt.

### Acceptance criteria

- [ ] E2E-Test im Projekt `Pixel 7`: zwei Varianten mit langem gemeinsamem Präfix sind
      vollständig lesbar
- [ ] Hoch- und Querformat des Handy-Viewports geprüft (Playwright `setViewportSize`)
- [ ] `make check` grün

---

## Phase 5: Service-UI — Reihenfolge von Produkten und Varianten

**Depends on**: none

### Context

- PR #109 — Migration `reihenfolge`, Endpunkte zum Verschieben, Admin-UI
- `database/migrations/README.md` — Regeln für additive Migrationen
- `docs/language.md` — Begriff „Reihenfolge" aufnehmen
- `docs/handbuch.md` — Sortierregel der Produktliste

### What to build

Admins legen die Reihenfolge von Produkten und ihren Varianten fest; die Service-Liste
sortiert innerhalb der Kategorie danach. Häufig bestellte Varianten stehen oben, das
Scrollen sinkt ohne neue Interaktionsebene. PR #109 wird reviewt und, wenn Migration
(additiv, forward-only), POST-only-Endpunkte und Validierung passen, übernommen.

### Acceptance criteria

- [ ] Neue Migration ist additiv und nummeriert, `01_initial.up.sql` unverändert
- [ ] Service-Liste sortiert nach (Kategorie, Reihenfolge, ID)
- [ ] `docs/language.md` und `docs/handbuch.md` beschreiben die Reihenfolge
- [ ] `make verify` grün

---

## Phase 6: Bondruck-Ursachenklärung abschließen

**Depends on**: none

### Context

- `docs/plans/plan-bondruck-ursachenklaerung.md` — bestehender Plan, alle Punkte offen

### What to build

Den bestehenden Plan abarbeiten oder bewusst mit „Ursache nicht bestimmbar" schließen.
Nichts wird hier dupliziert; diese Phase ist nur die Frist: vor den produktiven
Einsätzen Ende September.

### Acceptance criteria

- [ ] Bestehender Plan hat keine offene Checkbox mehr oder ist gelöscht

---

## Phase 7: Abholbon pro Stück im Direktverkauf

**Depends on**: none

### Context

- `backend/api/druck/bondruck/application/arbeitsbon_policy.go —
createAbholbonAuftraege()` — Ableitung der Abholbon-Aufträge
- `backend/api/druck/bondruck/application/escpos/formatter.go — FormatPositionBon()`
- `backend/domain/druckstation/druckstation.go` — Bonmodus-Typ
- `backend/api/druck/station/http/handler.go` — Validierung der Konfiguration
- `frontend/src/admin/settings/DruckstationBackend.ts — BonmodusSchema`
- `frontend/src/admin/settings/DruckstationConfigPage.tsx` — Auswahl
- `database/migrations/01_initial.up.sql` — CHECK auf `bonmodus`
- `docs/language.md — #### Bonmodus`, `docs/handbuch.md — ### 4.6`,
  `docs/anforderungen.md` — Funktionsumfang

### What to build

Gäste und Gruppen kaufen im Direktverkauf mehrere Einheiten auf einmal und lösen sie
einzeln an der Theke ein. Die Druckstation `abholbon` bekommt den dritten Bonmodus
`pro_stueck`: je Einheit einer Position ein eigener Abholbon mit `1x Bezeichnung`. Für
die Produktstationen (`essen`, `getraenk`, `sonstiges`) bleibt der Wert unzulässig;
Backend (zog) und Frontend (Zod) lehnen ihn dort ab. Neue Migration ersetzt die
CHECK-Bedingung. Kassenbeleg und TSE sind nicht betroffen; der Abholbon bleibt
nicht-fiskalisch.

### Acceptance criteria

- [ ] Migration `07_abholbon_pro_stueck.up.sql` erlaubt `pro_stueck` nur für `abholbon`
- [ ] Direktverkauf mit `3x Bier` und Modus `pro_stueck` erzeugt drei Druckaufträge mit
      `1x Bier`
- [ ] Modus `pro_stueck` an einer Produktstation wird von Backend und Admin-UI abgelehnt
- [ ] Bestehende Konfigurationen behalten ihr Verhalten (Standard bleibt `pro_position`)
- [ ] `docs/language.md`, `docs/handbuch.md`, `docs/anforderungen.md` beschreiben den Modus
- [ ] `make verify` grün

---

## Phase 8: TSE-Kosten und Hardware-TSE als ADR

**Depends on**: none

### Context

- `docs/compliance.md — ### 3.5 TSE-Varianten und Anbieter-Entscheidung`
- `docs/leitfaden/haeufige-fragen.md` — Frage „Was kostet der Betrieb?"
- `docs/leitfaden/tse-einrichten.md` — Abschnitt zu Kosten

### What to build

Ein Verein berichtet für die fiskaly-Cloud-TSE eine Mindestabnahme von mehreren Kassen
und eine Mindestlaufzeit; für zwei Feste im Jahr ist das die eigentliche Hürde. Der
Spike verifiziert die aktuellen fiskaly-Konditionen für eine Kasse, prüft alternative
Cloud-TSE-Anbieter mit Kurzzeitlizenz und schätzt, was eine Hardware-TSE am
Windows-Starter über das `TSEClient`-Adapter-Interface bedeuten würde. Ergebnis ist eine
ADR und eine ehrliche Kostenaussage in der FAQ.

### Acceptance criteria

- [ ] ADR `09_tse-kosten-und-hardware-tse.md` mit Konditionen, Alternativen, Entscheidung
- [ ] FAQ nennt Größenordnung und Vertragsbindung der TSE mit Datum der Recherche
- [ ] `docs/compliance.md` Abschnitt 3.5 stimmt mit der ADR überein

---

## Phase 9: Nicht-Ziele aus dem Feedback festhalten

**Depends on**: none

### Context

- `docs/anforderungen.md` — Nicht-Ziele-Tabelle
- `docs/produktbeschreibung.md` — Abgrenzung

### What to build

Drei Punkte werden als Nicht-Ziele mit Begründung eingetragen: Vereinslogo auf dem Bon
(kosmetisch, Raster-Druck und Logo-Upload ohne Kernnutzen; der Vereinsname steht im
Kopf), elektronischer Beleg per E-Mail oder Link (Gast-Handy ist nicht im Vereins-LAN,
Mailversand vom Vereins-Server und Adress-Erfassung am Tisch; Befreiung oder
Theken-Drucker decken den Bedarf), Helferdeckel (ein Tisch pro Helfer deckt es ab). Die
Tabelle bleibt die einzige Stelle; die FAQ aus Phase 1 verweist darauf.

### Acceptance criteria

- [ ] Drei Zeilen in der Nicht-Ziele-Tabelle mit je einer Begründung
- [ ] `docs/produktbeschreibung.md` Abgrenzung stimmt damit überein

---

## Phase 10: Produktebene bewerten

**Depends on**: 5

### Context

- PR #111 — Produkt-Kacheln vor der Variantenliste
- `frontend/src/service/components/table/ProductList.tsx — ProductList()` — vorhandene
  Kategorie-Pills
- `docs/adrs/08_service-split-screen.md` — bestehende Entscheidung zum Service-Layout

### What to build

Zuerst prüfen, auf welchem Stand der PR entstanden ist und ob er die vorhandenen
Kategorie-Pills berücksichtigt oder ersetzt. Dann nach mindestens einem Fest mit
sortierter Liste entscheiden, ob eine Produktebene den zusätzlichen Tap pro Bestellung
rechtfertigt. Ergebnis ist eine ADR: angenommen (dann #111 auf den aktuellen Stand
bringen und übernehmen) oder abgelehnt (dann #111 mit Begründung schließen).

### Acceptance criteria

- [ ] Prüfergebnis zum PR-Stand gegenüber den Kategorie-Pills liegt in der ADR
- [ ] ADR `10_produktebene-service.md` mit Status und Begründung
- [ ] PR #111 gemerged oder mit Verweis auf die ADR geschlossen

---

## Phase 11: Release v1.0.0

**Depends on**: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10

### Context

- `docs/plans/guide-manuelle-qa-v1.0.0.md` — Rest-Guide, toter Verweis auf
  `plan-v1.0-release-blockers.md`
- `CHANGELOG.md` — Abschnitt `[1.0.0]` vorhanden

### What to build

Den QA-Guide durchlaufen, den toten Verweis entfernen, den Abholbon-Modus `pro_stueck` und
die Reihenfolge in den `[1.0.0]`-Abschnitt aufnehmen, Version und Tag `v1.0.0` setzen.
Vereine mit Einsatz vor dem Release arbeiten mit v0.17.3 und erhalten nach dem Release
eine Nachricht; ein Update ist ein ZIP-Tausch mit automatischem Backup.

### Acceptance criteria

- [ ] QA-Guide ohne offene Checkbox, Datei gelöscht
- [ ] Tag `v1.0.0` und GitHub-Release vorhanden
- [ ] `CHANGELOG.md` `[1.0.0]` trägt Release-Datum, Reihenfolge und `pro_stueck`
