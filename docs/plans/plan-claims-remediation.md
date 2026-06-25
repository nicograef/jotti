---
title: Plan — Claims-Remediation (Fiskalkonformität & Haftung)
description: Entschärfung der öffentlichen Konformitätsaussagen (Website, README, Produktbeschreibung) auf belegbares, tätigkeitsbeschreibendes Wording plus Vereinheitlichung des DSFinV-K-Status — zur Reduktion von UWG- und § 379-AO-Risiko.
---

# Plan: Claims-Remediation (Fiskalkonformität & Haftung)

> Source PRD: n/a (aus Risikoeinschätzung in der Konversation abgeleitet)

## Goal

Die öffentlichen Aussagen von jotti so umstellen, dass sie **belegbar und wahr**
sind und den eigenen Haftungsausschluss (TERMS.md §§ 7–10) nicht mehr
untergraben. Konkret: absolute Konformitätsversprechen
(„Finanzamtssicher. Ab Werk.", „Fiskalkonform") durch tätigkeitsbeschreibende
Aussagen ersetzen, den widersprüchlichen DSFinV-K-Status vereinheitlichen und
die tatsächlich erzeugte DSFinV-K-Version (2.4) im Code korrigieren.

**Warum:** Software-Konformität ist nicht zertifizierbar (nur die TSE wird
BSI-zertifiziert). Absolute Claims sind damit unbelegbar; bei objektiver
Unrichtigkeit drohen irreführende Werbung (§ 5 UWG, Abmahnung) und ein Bußgeld
bis 25.000 € nach § 379 Abs. 1 S. 1 Nr. 6 i.V.m. Abs. 6 AO (Bewerben/
Inverkehrbringen nicht-konformer Systeme, unabhängig von tatsächlicher
Steuerverkürzung). Zudem kann eine öffentlich beworbene Eigenschaft als
vereinbarte Beschaffenheit/Garantie gelten und den AGB-Haftungsausschluss
aushebeln (§ 305b BGB).

## Architectural decisions

Durchgängige Entscheidungen (in der Klärungsrunde festgelegt):

- **Lizenz bleibt unverändert.** Source-Available + TERMS.md (§ 521 BGB
  Schenkungsprivileg, Freistellung, aktive Annahme) sind ein Haftungs-Asset.
  LICENSE, TERMS.md, docs/lizenzmodell.md werden in diesem Plan **nicht**
  angefasst.
- **Tonlage: konservativ/tätigkeitsbeschreibend.** Nie absolutes
  „fiskalkonform"/„finanzamtssicher". Erlaubte, belegbare Bausteine:
  „nutzt eine BSI-zertifizierte Cloud-TSE (fiskaly)", „erstellt Belege nach
  § 146a AO", „append-only Kassenjournal (unterstützt GoBD-Anforderungen)",
  „DSFinV-K-Export (v2.4)", stets gekoppelt an einen Hinweis auf die
  Betreiberverantwortung.
- **Realer Funktionsstand: implementiert, aber nicht praxis-/prüfungsvalidiert.**
  TSE, Beleg und DSFinV-K-Export sind im Code vorhanden und unit-getestet, aber
  noch nicht in echter Betriebsprüfung/IDEA-Prüfung oder produktivem Festbetrieb
  bestätigt. Das Wording darf Existenz behaupten, aber keine geprüfte Konformität.
- **DSFinV-K-Zielversion: 2.4** (verbindliche Fassung, Stand Dez. 2023). Der Code
  deklariert fälschlich `Version = "2.5"`.
- **Eine kanonische Status-Formulierung** wird überall wortgleich verwendet
  (Website-Intro, README, Produktbeschreibung). Festgelegter Wortlaut → siehe
  Resolved decisions.
- **Beta-Kennzeichnung bleibt** und wird konsistent gehalten (README-Note,
  Hero-/Eyebrow-Badge). Sie senkt die berechtigte Erwartung, heilt aber keine
  Falschaussage — daher zusätzlich zu, nicht statt korrekter Claims.

## Inventory

**Öffentliche Claim-Oberflächen (in Scope):**

- `website/src/pages/index.astro` — Landing Page, höchstes Risiko:
  - `:12` Meta-/OG-Description „Fiskalkonform mit TSE-Anbindung …"
  - `:101-102` Hero-Badge „Fiskalkonform"
  - `:187` „Alle Funktionen inklusive Fiskalkonformität."
  - `:419` Tabellen-Zeilenlabel „Fiskalkonformität" (Topic, unkritisch)
  - `:422` jotti-Zelle „KassenSichV-konform (TSE + DSFinV-K)" — Konformitätsurteil
  - `:984-985` Kommentar + `id="fiskalkonform"` (Anchor — bleibt)
  - `:987` Eyebrow „Rechtssicher"
  - `:991` „Finanzamtssicher. Ab Werk." — schärfster Claim
  - `:997` „… Fiskalkonform, ohne Zusatzkosten für die Software."
  - `:1225` „… mit voller Fiskalkonformität …"
  - `:1530` „kostenlos, fiskalkonform und in Minuten einsatzbereit"
- `website/src/layouts/Landing.astro`:
  - `:143-145`, `:226-228` Nav-Label „Fiskalkonform" (Anchor `/#fiskalkonform` bleibt)
  - `:267` Footer „Das kostenlose, fiskalkonforme Kassensystem …"
- `README.md`:
  - `:10` „Auf dem Weg zur Fiskalkonformität." (bereits ehrlich)
  - `:11`, `:99` „… DSFinV-K-Export in Entwicklung" (ehrlich)
  - `:48` DSFinV-K-Export als fertiges Feature gelistet + „v2.5" (Widerspruch)
  - `:107` „Compliance-Hinweis"-Blockquote (gut, bleibt als Vorbild)
- `docs/produktbeschreibung.md`:
  - `:14` „Auf KassenSichV-Konformität ausgelegt …" (gut)
  - `:22` „… einfaches und fiskalkonformes Kassensystem …" (absolut)
  - `:34` „… voller Fiskalkonformität als Zielbild." (vertretbar)
  - `:62`, `:71`, `:120`, `:131`, `:172` weitere Vorkommen (überwiegend
    „ausgelegt"/„in Entwicklung", einzeln prüfen)

**Wahrheits-Backbone (in Scope):**

- `backend/domain/dsfinvk/dsfinvk.go:20` — `const Version = "2.5"` → `"2.4"`
- `backend/domain/dsfinvk/index.go:29` ff. — schreibt `<Version>` in index.xml
  (nutzt die Konstante; verifizieren)
- `backend/domain/dsfinvk/index_test.go:26` — erwartet `"<Version>2.5</Version>"`
- `backend/domain/dsfinvk/mapper_test.go:124` — Testdaten enthalten `"2.5"`
- `docs/compliance.md` §2.5 / §6.1 — sagt bereits korrekt „verbindlich v2.4"
  (Referenz, bleibt; auf interne Konsistenz prüfen)

**Bewusst NICHT in Scope (interne Domänensprache, kein öffentliches Versprechen):**

- `docs/language.md:387` „### Fiskalkonformität (Compliance Sub-Domain)" —
  Ubiquitous-Language-Begriff, keine Marketingaussage
- `docs/anforderungen.md:98` „### Fiskalkonformität" — interne Anforderungs-Sektion
- `LICENSE`, `TERMS.md`, `docs/lizenzmodell.md` — Lizenz bleibt unverändert

## Resolved decisions

- Lizenzmodell unverändert (Source-Available + TERMS).
- Konservatives, tätigkeitsbeschreibendes Wording statt absoluter Konformität.
- Funktionsstand-Narrativ: „implementiert, aber noch nicht praxis-/prüfungs-
  validiert".
- DSFinV-K-Zielversion ist **2.4** (Code-Konstante `"2.5"` ist ein Fehler).
- Organisatorische Maßnahmen (Anwaltsprüfung, UG-Gründung, IT-Vermögensschaden-
  haftpflicht) sind **nicht** Teil dieses Plans.
- Interne Domänendokumente (language.md, anforderungen.md) bleiben unangetastet.

**Festgelegte Wortlaute (Klärungsrunde 2):**

- **Kanonischer Statussatz** (wortgleich auf Website-Intro, README,
  Produktbeschreibung):

  > jotti bringt die fiskalischen Bausteine mit: eine BSI-zertifizierte
  > Cloud-TSE, Belegausgabe nach § 146a AO, ein append-only Kassenjournal (GoBD)
  > und den DSFinV-K-Export (v2.4). Eine geprüfte Konformität wird nicht
  > zugesichert; den konformen Betrieb (TSE-Vertrag, Kassenmeldung, Aufbewahrung)
  > verantwortet der Betreiber.

- **Compliance-Abschnitt (index.astro):** Eyebrow „Compliance", H2 „Auf die
  KassenSichV ausgelegt." (ersetzt Eyebrow „Rechtssicher" + „Finanzamtssicher.
  Ab Werk.").
- **Hero-Trust-Badge:** „TSE-Anbindung" (ersetzt „Fiskalkonform"; steht neben
  „Kostenlos" / „Self-hosted").
- **Disclaimer-Form:** eigene, sichtbar abgesetzte Callout-Box im
  Compliance-Abschnitt (analog README-„Compliance-Hinweis"), Kerntext „jotti
  sichert keine Konformität zu; die Verantwortung liegt beim Betreiber", mit
  Links auf `docs/compliance.md` und den Leitfaden für Vereine.
- **Beta:** kein zusätzlicher Beta-Satz am Compliance-Abschnitt — das globale
  Beta-Badge (Hero) und die README-Beta-Note genügen.
- **Vergleichstabellen-Zelle (index.astro:422):** „TSE, Beleg & DSFinV-K
  integriert" (deskriptiv, kein Konformitätsurteil gegenüber der Konkurrenz).

## Open questions / Risks

- **CSV-Struktur vs. Versionsstring:** Die DSFinV-K-Tabellenstruktur ist seit
  v2.0 stabil; der Wechsel des deklarierten Strings 2.5 → 2.4 sollte ohne
  Strukturänderung gültig bleiben. Vor dem Merge kurz gegen die Spezifikation
  v2.4 prüfen, dass der erzeugte Export wirklich der deklarierten Version
  entspricht (kein 2.5-spezifisches Feld in Verwendung).
- **Anchor-Stabilität:** Sichtbare Labels ändern sich, die Sektion behält
  `id="fiskalkonform"` und die internen Links `/#fiskalkonform` (keine toten
  Anker, kein SEO-Bruch).

---

## Phase 1: Landing Page entschärfen

### Context

- `website/src/pages/index.astro:12,101-102,187,419-422,984-997,1225,1530` —
  alle öffentlichen Konformitätsaussagen der Hauptseite (inkl. SEO-Meta, das aus
  der `description`-Konstante in `Landing.astro` gespeist wird)
- `website/src/layouts/Landing.astro:143-145,226-228,267` — Nav-Labels und Footer
- `docs/compliance.md`, `docs/leitfaden/was-ist-jotti.md` — Ziel der neuen
  Disclaimer-/Mehr-erfahren-Verlinkung

### What to build

Die Landing Page durchgängig von absoluten Konformitätsversprechen auf
tätigkeitsbeschreibendes Wording umstellen:

- Eyebrow „Rechtssicher" → „Compliance"; H2 „Finanzamtssicher. Ab Werk." →
  „Auf die KassenSichV ausgelegt." (festgelegt).
- Hero-Trust-Badge „Fiskalkonform" → „TSE-Anbindung" (festgelegt).
- Meta-/OG-Description (`:12`), „voller Fiskalkonformität" (`:1225`),
  „kostenlos, fiskalkonform" (`:1530`), „Alle Funktionen inklusive
  Fiskalkonformität" (`:187`), Abschnitts-Schlusssatz (`:997`) auf belegbare
  Bausteinaussagen umstellen, abgeleitet aus dem kanonischen Statussatz.
- Vergleichstabellen-Zelle (`:422`) „KassenSichV-konform (TSE + DSFinV-K)" →
  „TSE, Beleg & DSFinV-K integriert" (festgelegt; kein Konformitätsurteil).
- Nav-/Footer-Labels „Fiskalkonform" (Landing.astro) deskriptiv machen; Anchor
  `id="fiskalkonform"` und Links `/#fiskalkonform` bleiben.
- **Neu:** eine sichtbar abgesetzte Callout-Box im Compliance-Abschnitt
  (festgelegt), Kerntext „jotti sichert keine Konformität zu; die Verantwortung
  liegt beim Betreiber", mit Links auf `docs/compliance.md` und den Leitfaden —
  analog zur README-„Compliance-Hinweis"-Box, die es auf der Website bisher
  nicht gibt.
- Globales Beta-Badge im Hero bleibt; **kein** zusätzlicher Beta-Satz am
  Compliance-Abschnitt.

Aussagen über Wettbewerber (Excel „keine Fiskalkonformität" `:346`,
kommerzielles POS „Fiskalkonform und leistungsfähig" `:369`) bleiben — sie sind
keine jotti-Selbstaussagen.

### Acceptance criteria

- [x] Auf der gerenderten Landing Page (inkl. `<meta name="description">` und
      OG-Tags) kommt kein absolutes „fiskalkonform"/„finanzamtssicher" als
      jotti-Selbstaussage mehr vor.
- [x] Die Vergleichstabelle behauptet keine KassenSichV-Konformität als Urteil,
      sondern beschreibt die integrierten Bausteine.
- [x] Eine sichtbar abgesetzte Callout-Box zur Betreiberverantwortung ist im
      Compliance-Abschnitt vorhanden und verlinkt auf compliance.md und den
      Leitfaden.
- [x] Alle internen Anker (`/#fiskalkonform`) funktionieren weiterhin (kein
      toter Link).
- [x] Beta-Kennzeichnung weiterhin sichtbar; Astro-Build läuft fehlerfrei
      (`make`-Build der Website grün).

---

## Phase 2: README & Produktbeschreibung angleichen

### Context

- `README.md:10-11,48,99,107` — Statusaussagen teils ehrlich, teils
  widersprüchlich (`:48` listet DSFinV-K-Export ohne WIP-Tag und nennt „v2.5")
- `docs/produktbeschreibung.md:14,22,34,62,71,120,131,172` — Mischung aus
  bereits korrektem („ausgelegt"/„in Entwicklung") und absolutem Wording (`:22`)
- Output von Phase 1 — dieselbe kanonische Status-Formulierung wiederverwenden

### What to build

Den festgelegten kanonischen Statussatz (siehe Resolved decisions) in README und
Produktbeschreibung **wortgleich** einsetzen. Verbleibende absolute Formulierungen (insb. produktbeschreibung
`:22` „fiskalkonformes Kassensystem") auf Bausteinaussagen umstellen.
`README.md:48` so anpassen, dass der DSFinV-K-Export konsistent mit `:11`/`:99`
dargestellt wird und die Version 2.4 nennt (statt v2.5). Die README-„Compliance-
Hinweis"-Box bleibt als gutes Vorbild erhalten.

### Acceptance criteria

- [x] README und Produktbeschreibung verwenden eine identische, belegbare
      Status-Formulierung; kein interner Widerspruch mehr zwischen „Feature
      vorhanden" und „in Entwicklung".
- [x] Kein absolutes „fiskalkonform" als jotti-Selbstaussage außerhalb klar als
      Zielbild markierter Stellen.
- [x] Jede DSFinV-K-Versionsangabe in README nennt 2.4.
- [x] Beta-Note in der README bleibt erhalten.

---

## Phase 3: Wahrheits-Backbone — DSFinV-K-Version korrigieren

### Context

- `backend/domain/dsfinvk/dsfinvk.go:18-20` — `const Version = "2.5"` (Kommentar
  behauptet „aktuell verbindlich ist v2.5", was falsch ist)
- `backend/domain/dsfinvk/index.go:29` ff. — `<Version>`-Ausgabe in index.xml
- `backend/domain/dsfinvk/index_test.go:26`, `mapper_test.go:124` — Tests mit
  hartkodiertem „2.5"
- `docs/compliance.md` §2.5/§6.1 — nennt bereits korrekt v2.4 (Konsistenz prüfen)

### What to build

Die Versionskonstante auf `2.4` korrigieren und den irreführenden Kommentar
richtigstellen. Alle abhängigen Tests anpassen, sodass der erzeugte Export und
die `index.xml` Version 2.4 deklarieren. Vor Abschluss verifizieren, dass kein
2.5-spezifisches Strukturmerkmal verwendet wird (Spezifikation seit v2.0 stabil,
Strukturänderung nicht erwartet).

### Acceptance criteria

- [x] `dsfinvk.go` deklariert `Version = "2.4"`; der Kommentar nennt die korrekte
      verbindliche Fassung.
- [x] Erzeugte `index.xml` enthält `<Version>2.4</Version>`.
- [x] `go test ./backend/domain/dsfinvk/...` ist grün (alle „2.5"-Erwartungen
      aktualisiert).
- [x] compliance.md, README und der erzeugte Export nennen durchgängig dieselbe
      Version (2.4) — kein Widerspruch mehr zwischen Code, Doku und Marketing.
