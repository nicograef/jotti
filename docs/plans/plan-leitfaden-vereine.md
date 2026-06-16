# Plan: Ein einziger Leitfaden für Vereine (Doku-Konsolidierung)

> Source PRD: [docs/prds/prd-leitfaden-vereine.md](../prds/prd-leitfaden-vereine.md)

## Goal

Aus vier verstreuten Betriebs-Leitfäden und zwei Website-Seiten wird genau ein
kanonischer Leitfaden `docs/leitfaden.md` (progressive disclosure: Standardfall
zuerst, Expertenkram unten), inhaltlich korrekt, in einfacher Sprache, ohne
KI-Slop. Alles Umliegende (README, Produktbeschreibung, AGENTS, Website, Code-
und Querverweise) wird mit ihm faktisch in Deckung gebracht oder zu einem dünnen
Verweis darauf. Am Ende ist `docs/betrieb/` gelöscht und kein interner Link tot.

## Architectural decisions

Durable Entscheidungen für alle Phasen:

- **Kanonische Datei:** `docs/leitfaden.md` ist die einzige Quelle der Wahrheit.
  Sie ersetzt `docs/betrieb/leitfaden-betreiber.md`,
  `docs/betrieb/leitfaden-hosting.md`, `docs/betrieb/leitfaden-tse-einrichtung.md`
  und `docs/betrieb/dns-rebind-schutz.md`; der Ordner `docs/betrieb/` wird zuletzt
  gelöscht.
- **Reihenfolge (keine toten Links zu keinem Zeitpunkt):** erst `leitfaden.md`
  schreiben → dann README/Produktbeschreibung/AGENTS angleichen → dann Website
  spiegeln → zuletzt Code-/Querverweise umbiegen und `docs/betrieb/` löschen. Die
  Löschung ist die letzte Handlung, nach allem Repointing.
- **Korrektheits-Baseline (verbindlicher Faktenstand überall identisch):**
  - TSE-Anbindung: vorhanden. Belegausgabe: vorhanden.
  - DSFinV-K-Export: in Entwicklung, Zielformat v2.5 (nicht v2.4).
  - Hash-Chain: existiert nicht und ist nicht geplant; jede Erwähnung in den
    angefassten Texten entfällt.
  - TSE-Anbieter: fiskaly ist der einzige unterstützte Anbieter. Kanonische
    Formulierung „Cloud-TSE von fiskaly"; kein „z. B. fiskaly", keine
    Alternativen (D-Trust). Bring Your Own fiskaly-Konto bleibt.
  - Betriebs-Reverse-Proxy: Caddy. nginx nur noch im `jotti.rocks`-Demo-Stack,
    nicht mehr als Betriebs-Proxy genannt.
- **Ein Name pro Hosting-Weg, durchgängig identisch in Leitfaden und Website:**
  „Standardweg" (Einzelgerät im WLAN) und „Experten-Weg" (eigener Server).
  Ersetzt „Weg A"/„Weg B" und „Standard-Weg".
- **Anker-Vertrag:** Die vom Code referenzierten Abschnitte (Fehlersuche/
  DNS-Rebind) bekommen stabile Heading-Anker, weil `reverse-proxy/statuspage.go`
  und sein Test eine URL-Konstante (Datei + `#anker`) festschreiben.
- **Website-Identität:** Die eine verbleibende Vereins-Leitfaden-Seite bleibt
  unter `/leitfaden-fuer-vereine/`. `/jotti-selbst-betreiben/` wird per
  301-Redirect in `reverse-proxy/nginx.rocks.conf` darauf umgeleitet; der Ordner
  `website/jotti-selbst-betreiben/` entfällt. `sitemap.xml` führt nur noch zwei
  URLs (Startseite + Leitfaden).
- **Mirror-Prinzip:** Markdown ist kanonisch; die Website ist die bewusste,
  von Hand gepflegte Spiegelung. Keine Build-Pipeline Markdown→HTML.

## Inventory

Altbestand (Quelle für Phase 1, Löschung in Phase 4):

- `docs/betrieb/leitfaden-betreiber.md` (11 KB) — Pflichten, Betreiber-Setup
- `docs/betrieb/leitfaden-hosting.md` (16 KB) — Weg A/B, WLAN vs. Server
- `docs/betrieb/leitfaden-tse-einrichtung.md` (16 KB) — fiskaly, Wizard, PUK/PIN
- `docs/betrieb/dns-rebind-schutz.md` (6 KB) — Router-Anleitung, Fallback

Faktisch anzugleichende Referenzdokumente:

- `README.md:8` Intro-Status; `:25-26,43-45` „in Entwicklung"-Marker (DSFinV-K
  v2.4 → v2.5); `:53` Hash-Chain-Zeile (entfällt); `:54` TSE „in Entwicklung"
  (→ vorhanden); `:85-86` Tech-Tabelle (Reverse Proxy nginx → Caddy); `:97`
  Status; `:105` Compliance-Hinweis mit „z. B. fiskaly" und Link auf
  `docs/betrieb/leitfaden-betreiber.md`
- `docs/produktbeschreibung.md:97,140` DSFinV-K v2.5 (bereits korrekt); `:186`
  „Cloud-TSE (z. B. fiskaly)" → fiskaly; kein nginx-/Reverse-Proxy-Vorkommen
  (Reverse-Proxy-Angleich ist hier ein No-op)
- `AGENTS.md:38` Infrastruktur-Tabelle „nginx Reverse Proxy" → Caddy
- `docs/language.md:383` listet „Hash-Chain … offen" → widerspricht der Baseline
  („nicht geplant"); einzeilige Faktenkorrektur (Hash-Chain aus der Liste nehmen)

Website (Phase 3):

- `website/leitfaden-fuer-vereine/index.html` — verbleibende Seite (Ziel des
  Merge); Querlinks auf `/jotti-selbst-betreiben/` bei `:50,470`
- `website/jotti-selbst-betreiben/index.html` — wird eingeschmolzen; „Weg A/B"
  bei `:68,77,108-109,162,303,…`; Glossar bei `:478-479`
- `website/index.html:424,428` DSFinV-K v2.4 → v2.5; `:483,626,665,716` „z. B.
  fiskaly"; `:675` „Docker Compose, nginx, Let's Encrypt" → Caddy
- `website/partials/header.html:15-16` zwei Nav-Links („Leitfaden", „Hosting") →
  ein Eintrag; `website/partials/footer.html:17,22`;
  `website/partials/mobile-nav.html:5-6`
- `website/sitemap.xml:10,16` (zwei Leitfaden-URLs → eine)
- `website/404.html` (vorhanden); `reverse-proxy/nginx.rocks.conf` (SSI-Auslieferung,
  nutzt bereits `return 301` — Einhängepunkt für den Redirect)

Code- und Querverweise (Phase 4):

- `reverse-proxy/statuspage.go:24` `rebindGuideURL`-Konstante (→ Fehlersuche-Anker
  in `docs/leitfaden.md`); `:135` Verwendung; `reverse-proxy/statuspage_test.go:78`
  asserted die Konstante im Body
- `frontend/src/admin/finanzamt/DokumenteUndPflichtenSection.tsx:17` Eintrag
  „Betreiber-Leitfaden" → `docs/leitfaden.md`; Compliance-Link bei `:21` bleibt
- `docs/compliance.md:395` Links auf `betrieb/leitfaden-betreiber.md` und
  `betrieb/leitfaden-tse-einrichtung.md`; `:413` „fiskaly oder D-Trust" → fiskaly
  und Link auf `betrieb/leitfaden-tse-einrichtung.md`
- `packaging/windows/KURZANLEITUNG.md:45` Online-Link auf `dns-rebind-schutz.md`
  (ausgeliefertes Windows-Dokument) → Fehlersuche-Anker
- `docker-compose.local.yml:10` Kommentar „… leitfaden-hosting.md → Weg A";
  `docker-compose.prod.yml:9` Kommentar „… leitfaden-hosting.md → Weg B"
- `docs/plans/plan-tse-setup-wizard.md:45,185` betrieb-Verweise (Scrub)
- `docs/anforderungen.md:246` Checklistenzeile auf nie erstelltes
  `docs/betrieb/elster-meldung.md` (Scrub)

Prior Art / Muster:

- `docs/plans/plan-tse-setup-wizard.md` — Plan-Format und -Stil
- `reverse-proxy/nginx.rocks.conf:24-37` — bestehendes `return 301`-Muster

## Resolved decisions

Aus dem Klärungslauf (2026-06-16, mit User abgestimmt):

- **Website-URL:** Die verbleibende Seite bleibt `/leitfaden-fuer-vereine/`
  (höhere Priorität, bestehende Indizierung); `/jotti-selbst-betreiben/` wird
  darauf zusammengeführt.
- **Alte Pfade:** 301-Redirect in `reverse-proxy/nginx.rocks.conf` (kein
  404-Fallback), da PRD den Redirect als wünschenswert nennt und nginx.rocks.conf
  ihn trivial unterstützt.
- **Sweep-Umfang:** repo-weit alles scrubben — auch `docs/plans/*` und die
  unangehakte `docs/anforderungen.md`-Checklistenzeile (`elster-meldung.md`),
  sodass ein repo-weiter grep nach `docs/betrieb` null Treffer liefert. Begründung
  für die Ausnahme zum PRD-Out-of-Scope: ausdrücklicher User-Wunsch nach
  Null-Treffern; die Änderung an `anforderungen.md` beschränkt sich auf das
  Umbiegen/Entfernen dieser einen Pfadangabe, keine inhaltliche Überarbeitung.
- **language.md:** Die eine widersprüchliche Hash-Chain-Zeile (`:383`) wird als
  minimale Faktenkorrektur in Phase 2 mitgenommen, obwohl das PRD `language.md`
  nicht namentlich nennt (über die genannten Dateien hinaus, im Sinne der
  Korrektheits-Baseline).

## Open questions / Risks

- **Anker-Stabilität:** Der Fehlersuche-/DNS-Rebind-Anker in `docs/leitfaden.md`
  muss vor dem Umbiegen der `statuspage.go`-Konstante feststehen. Phase 1 legt das
  Heading fest; Phase 4 verwendet exakt dessen GitHub-Anker. Risiko bei
  späterem Umbenennen des Headings → Code-Link bricht. Mitigation: Heading-Wortlaut
  in Phase 1 als bewusst gewählt markieren.
- **`anforderungen.md` / `handbuch.md` Hash-Chain-Stellen** (`anforderungen.md:304,313`)
  sind mit der Baseline konsistent (sie begründen, warum *keine* Hash-Chain gebaut
  wird) und bleiben unangetastet. Nur die `betrieb/`-Pfadzeile (`:246`) wird
  gescrubbt.

---

## Phase 1: Kanonischen `docs/leitfaden.md` schreiben

**User Stories:** 1, 2, 3, 4, 5, 6, 7, 8, 9, 12, 21 (sowie der Leitfaden-Anteil
von 10, 11, 23)

### Context

- `docs/betrieb/leitfaden-betreiber.md`, `docs/betrieb/leitfaden-hosting.md`,
  `docs/betrieb/leitfaden-tse-einrichtung.md`, `docs/betrieb/dns-rebind-schutz.md`
  — sachlicher Altbestand, der dedupliziert und gekürzt übernommen wird
- `docs/language.md` — verbindliche Fachbegriffe (Kassensitzung, Tagesabschluss/
  Z-Bon, Servicekraft, Serviceleitung, Admin, Direktverkauf)
- `docs/compliance.md`, `docs/steuerrecht.md` — bleiben die juristische
  Tiefen-Referenz hinter dem Leitfaden (nicht hierher kopieren, nur verweisen)

### What to build

Eine einzelne Markdown-Datei `docs/leitfaden.md`, oben sichtbar als kanonisch
markiert, mit der Abschnittsfolge nach progressive disclosure:

1. In 60 Sekunden: Was ist jotti, welche drei Pflichten, was kostet es.
2. Schnellstart Standardweg: Einzelgerät im WLAN, Doppelklick-Start, Handys per
   QR-Code, grünes Schloss als Normalfall, Fallback-Adresse als Auffanglösung.
3. TSE einrichten: fiskaly-Konto und API-Key, geführter Assistent, PUK/PIN
   verwahren, TEST→LIVE inkl. grober Kostenangabe.
4. Pflichten erfüllen: ELSTER-Meldung mit der Seriennummer aus dem Admin-Bereich,
   Belege und Steuersätze, zehn Jahre aufbewahren mit Backups.
5. Checkliste: einmalig vor dem Fest, laufend (abhakbar).
6. Experten-Weg (optional, klar markiert): eigener Server (VPS), Domain, HTTPS,
   Update, Backup, Härten.
7. TSE-Sonderfälle: PIN per PUK zurücksetzen, vorhandene TSS übernehmen, manuelle
   Konfiguration, Test-Limit, Wiederaufnahme nach Abbruch.
8. Fehlersuche: grüne Adresse lädt nicht (DNS-Rebind), Router-Hinweise, Fallback
   (Heading mit bewusst gewähltem, stabilem Anker für den Code-Link).
9. Häufige Fragen.

Inhalt aus dem Altbestand zusammenführen, doppelte Erklärungen verschmelzen,
seltene Detailfälle knapp halten. Einfache Sprache, kurze aktive Sätze,
Vereins-Ansprache („ihr"), keine großen Tabellen, kein Glossar, kein KI-Slop.
Faktenstand strikt nach Korrektheits-Baseline; ein Name pro Hosting-Weg; „Cloud-
TSE von fiskaly" als der Anbieter.

### Acceptance criteria

- [ ] `docs/leitfaden.md` existiert, enthält alle neun Abschnitte in dieser
      Reihenfolge und ist oben sichtbar als kanonische Quelle markiert.
- [ ] Standardweg steht vollständig vor jedem Server-/Domain-/Kommandozeilen-Thema;
      Experten-Weg und TSE-Sonderfälle stehen erkennbar als optional weiter unten.
- [ ] Faktenstand korrekt: TSE/Beleg vorhanden, DSFinV-K in Entwicklung (v2.5),
      keine Hash-Chain, Betriebs-Proxy Caddy, „Cloud-TSE von fiskaly" (kein
      „z. B.", keine Alternativen).
- [ ] Verbindliche Fachbegriffe gemäß `docs/language.md` durchgängig verwendet.
- [ ] Der Fehlersuche-/DNS-Rebind-Abschnitt hat ein stabiles Heading mit klar
      bestimmbarem GitHub-Anker (Vertragsbasis für Phase 4).
- [ ] Stil-Gegenprobe: Wortstrom-Diff gegen den Altbestand zeigt keinen neuen
      KI-Slop (keine Gedankenstrich-Manier, kein liberales Bold, keine Floskeln;
      Run-in-Labels erlaubt).
- [ ] Keine inhaltliche Erweiterung über den Altbestand hinaus (Ziel:
      Konsolidierung, Korrektur, Kürzung).

---

## Phase 2: README, Produktbeschreibung & AGENTS faktisch angleichen

**User Stories:** 15, 20 (sowie 10, 11, 23 für diese Dateien)

### Context

- `README.md:8,25-26,43-45,53-54,85-86,97,105` — Status, Marker, Hash-Chain,
  Tech-Tabelle, Compliance-Hinweis und Leitfaden-Link
- `docs/produktbeschreibung.md:97,140,186` — DSFinV-K-Version, fiskaly-Wording
- `AGENTS.md:38` — Infrastruktur-Tabelle (nginx → Caddy)
- `docs/language.md:383` — Hash-Chain-Zeile (Faktenkorrektur)
- `docs/leitfaden.md` (aus Phase 1) — Ziel des README-Verweises

### What to build

Alle vier Dateien auf die Korrektheits-Baseline ziehen, jede in ihrer Rolle:

- **README** neu schreiben, aber entwickler-/GitHub-orientiert: korrekter
  Funktionsstatus (TSE/Beleg vorhanden, DSFinV-K in Entwicklung v2.5, Hash-Chain
  entfernt), korrekter Tech-Stack (Caddy statt nginx), Schnellstart, genau ein
  Verweis auf `docs/leitfaden.md` (ersetzt den `betrieb/`-Link), „Cloud-TSE von
  fiskaly" im Compliance-Hinweis, Lizenz. Echt-in-Entwicklung-Features (KDS,
  Ausgabestationen, CSV/Reporting) behalten ihren Marker.
- **produktbeschreibung.md** punktuell angleichen: „z. B. fiskaly" → fiskaly;
  Status bleibt sonst (bereits weitgehend korrekt); Reverse-Proxy-Angleich ist
  ein No-op (kein Vorkommen).
- **AGENTS.md** Tech-Tabelle: nginx → Caddy (Betriebspfad).
- **language.md:383**: „Hash-Chain" aus der Liste der offenen Begriffe entfernen.

### Acceptance criteria

- [ ] README nennt TSE/Beleg als vorhanden, DSFinV-K als in Entwicklung (v2.5),
      enthält keine Hash-Chain-Zeile mehr und keine veralteten „in Entwicklung"-
      Marker auf TSE.
- [ ] README-Tech-Tabelle nennt Caddy als Reverse Proxy; AGENTS.md ebenso.
- [ ] README verweist auf genau eine Datei `docs/leitfaden.md` (kein
      `docs/betrieb/`-Link mehr); Link ist gültig (Datei existiert seit Phase 1).
- [ ] Produktbeschreibung und README nennen „Cloud-TSE von fiskaly" ohne „z. B."
      und ohne Alternativ-Anbieter.
- [ ] `language.md:383` widerspricht der Baseline nicht mehr.
- [ ] Faktenabgleich gegen die PRD-Statusliste bestanden (README-Status =
      Anforderungs-Status, keine nginx-Reste im Betriebspfad, v2.5, keine
      Hash-Chain).

---

## Phase 3: Website spiegeln (zwei Seiten → eine)

**User Stories:** 13, 14, 11, 19, 23

### Context

- `website/leitfaden-fuer-vereine/index.html` — verbleibende Seite (Merge-Ziel)
- `website/jotti-selbst-betreiben/index.html` — Quelle des Merge; „Weg A/B",
  Glossar `:478-479`
- `website/index.html:424,428,483,626,665,675,716` — Startseiten-Fakten
- `website/partials/header.html:15-16`, `footer.html:17,22`,
  `mobile-nav.html:5-6` — Navigation
- `website/sitemap.xml`, `website/404.html`
- `reverse-proxy/nginx.rocks.conf:24-37` — `return 301`-Muster für den Redirect
- `docs/leitfaden.md` — kanonische Quelle, auf die die Seite sichtbar verweist

### What to build

Die zwei Seiten zu einer Vereins-Leitfaden-Seite unter `/leitfaden-fuer-vereine/`
zusammenführen, die den Markdown-Leitfaden kompakt spiegelt und sichtbar als
Spiegelung darauf verweist. Glossar entfällt. „Weg A/Weg B" → „Standardweg/
Experten-Weg". Startseite `index.html` faktisch angleichen (DSFinV-K v2.5, nginx
→ Caddy in der Tech-Liste, fiskaly-Wording, keine Hash-Chain). Navigation
(Header, Mobile-Nav, Footer) auf einen Leitfaden-Eintrag reduzieren. `sitemap.xml`
auf zwei URLs (Startseite + Leitfaden). 301-Redirect `/jotti-selbst-betreiben/`
→ `/leitfaden-fuer-vereine/` in `nginx.rocks.conf`. Ordner
`website/jotti-selbst-betreiben/` entfernen. Kein Redesign, nur Inhalt/Struktur/
Konsistenz.

### Acceptance criteria

- [ ] Genau eine Vereins-Leitfaden-Seite unter `/leitfaden-fuer-vereine/`; kein
      Glossar; Wegbenennung „Standardweg/Experten-Weg" identisch zum Leitfaden.
- [ ] Seite verweist sichtbar auf `docs/leitfaden.md` als kanonische Quelle.
- [ ] Startseite zeigt denselben Funktionsstatus wie Leitfaden und README
      (DSFinV-K v2.5/in Entwicklung, keine Hash-Chain) und nennt Caddy, nicht
      nginx; fiskaly ohne „z. B.".
- [ ] Header, Mobile-Nav und Footer enthalten genau einen Leitfaden-Link;
      `sitemap.xml` listet nur Startseite + Leitfaden.
- [ ] `/jotti-selbst-betreiben/` liefert 301 auf `/leitfaden-fuer-vereine/`
      (in `nginx.rocks.conf`); der Ordner `website/jotti-selbst-betreiben/` ist
      entfernt.
- [ ] Kein interner Website-Link zeigt mehr auf `/jotti-selbst-betreiben/`.

---

## Phase 4: Code-/Querverweise umbiegen, `docs/betrieb/` löschen, Sweep

**User Stories:** 16, 17, 18, 19, 22 (sowie 20, 23 für compliance.md)

### Context

- `reverse-proxy/statuspage.go:24,135` + `reverse-proxy/statuspage_test.go:78`
  — DNS-Rebind-URL-Konstante und ihr Test
- `frontend/src/admin/finanzamt/DokumenteUndPflichtenSection.tsx:17,21`
- `docs/compliance.md:395,413` — Leitfaden-Links + „fiskaly oder D-Trust"
- `packaging/windows/KURZANLEITUNG.md:45` — ausgelieferter Online-Link
- `docker-compose.local.yml:10`, `docker-compose.prod.yml:9` — Kommentare
- `docs/plans/plan-tse-setup-wizard.md:45,185`, `docs/anforderungen.md:246`
- `docs/leitfaden.md` (Phase 1) — Ziel aller Verweise; Fehlersuche-Anker
- Reihenfolge: Löschung von `docs/betrieb/` erst nach allem Repointing.

### What to build

Alle verbliebenen Verweise auf `docs/leitfaden.md` umbiegen, dann den Altbestand
löschen, dann verifizieren:

- **statuspage.go:** `rebindGuideURL` auf `docs/leitfaden.md` + Fehlersuche-Anker
  (URL aus Phase 1) setzen; `statuspage_test.go` auf die neue Konstante ziehen.
- **Finanzamt-Komponente:** Eintrag „Betreiber-Leitfaden" → `docs/leitfaden.md`
  (URL-Konstante); Compliance-Link bleibt.
- **compliance.md:** beide Leitfaden-Links (`:395,413`) auf `docs/leitfaden.md`
  (Datei + Anker); „fiskaly oder D-Trust" → „fiskaly" (einzeilige Faktenkorrektur,
  sonst keine Überarbeitung); Hardware-TSE-Erklärung (Swissbit/Epson) unangetastet.
- **KURZANLEITUNG.md:** Online-Link auf den Fehlersuche-Anker von
  `docs/leitfaden.md`.
- **docker-compose-Kommentare:** auf `docs/leitfaden.md` und die neuen Wegnamen
  (Standardweg/Experten-Weg) umschreiben.
- **Scrub:** `plan-tse-setup-wizard.md` und die `anforderungen.md`-Checklistenzeile
  (`elster-meldung.md`) so anpassen, dass kein `docs/betrieb`-Pfad übrig bleibt.
- **Löschen:** Ordner `docs/betrieb/` komplett entfernen.
- **Sweep:** repo-weiter Suchlauf nach `docs/betrieb`, `leitfaden-betreiber`,
  `leitfaden-hosting`, `leitfaden-tse-einrichtung`, `dns-rebind-schutz` ergibt
  null Treffer (außer dem PRD/diesem Plan als Historie).

### Acceptance criteria

- [ ] `statuspage.go` zeigt auf `docs/leitfaden.md` + Fehlersuche-Anker;
      `statuspage_test.go` prüft die neue Konstante und ist grün (`go test ./...`
      im reverse-proxy-Modul).
- [ ] Finanzamt-Komponente, `compliance.md`, `KURZANLEITUNG.md` und die
      docker-compose-Kommentare verweisen auf `docs/leitfaden.md`.
- [ ] `compliance.md` nennt nur noch „fiskaly" (kein „oder D-Trust"); die
      Hardware-TSE-Erklärung bleibt unverändert.
- [ ] Ordner `docs/betrieb/` existiert nicht mehr.
- [ ] Repo-weiter grep nach `docs/betrieb` und den vier alten Dateinamen liefert
      null lebende Treffer (Docs, Code, Website, Packaging, Plans, anforderungen.md).
- [ ] Kein interner Link (Docs, Code, Website) ist tot.
