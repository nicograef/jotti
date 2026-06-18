# PRD: Website-Neubau mit Astro und Starlight

## Problem Statement

Die heutige `jotti.rocks`-Website ist ein statisches HTML/CSS-Projekt ohne
Build-Schritt: handgeschriebene Seiten, ein CSS-Mini-Design-System, SSI-Partials
für die Wiederverwendung und ein bash-basiertes Prüfskript (`scripts/check-website.sh`),
das Links, Assets und CSS-Klassen gegenprüft. Das funktioniert, hat aber drei
strukturelle Schwächen:

- **Doppelte Wahrheit beim Leitfaden.** Die Seite
  `website/leitfaden-fuer-vereine/index.html` ist eine handgebaute HTML-Spiegelung
  von `docs/leitfaden.md`. Beide Quellen müssen manuell synchron gehalten werden.
  Driften sie auseinander, lesen Website-Besucher etwas anderes als Agenten und
  Contributor im Repo.
- **Der Leitfaden ist als Webseite nicht gelungen.** Er liest sich als eine lange
  Marketing-Seite, nicht als Dokumentation. Es fehlt eine echte
  Schritt-für-Schritt-Führung, und der einfache Standardweg, der optionale
  Experten-Weg (eigener Server) sowie technische und rechtliche Pflichten sind
  vermischt statt sauber getrennt.
- **Keine tragfähige Doku-Plattform.** Die rechtlich und fachlich dichten
  Dokumente (`compliance.md`, `steuerrecht.md`, `verfahrensdokumentation.md`,
  `produktbeschreibung.md`, `lizenzmodell.md`) sind heute gar nicht als
  durchsuchbare, navigierbare Website verfügbar. Sie existieren nur als
  Markdown im Repo.

Gleichzeitig darf der Repo-Charakter der `docs/` nicht verloren gehen: Agenten,
Contributor und die `AGENTS.md`-Referenztabelle greifen direkt auf
`docs/*.md` zu. Eine Lösung, die die Inhalte in ein Website-Format wegsperrt,
ist nicht akzeptabel.

## Solution

Die Website wird auf **Astro mit Starlight** (dem offiziellen Doku-Framework von
Astro) neu gebaut, als eigenes pnpm-Workspace-Paket `website/` neben `frontend/`.
Eine Toolchain liefert sowohl die neu gestaltete Marketing-Landing als auch eine
professionelle, durchsuchbare Dokumentation im Doku-Look.

Die zentrale Entscheidung ist die einzige Quelle der Wahrheit: **`docs/` bleibt
der kanonische Ort des Markdowns.** Starlight liest die zu veröffentlichenden
Dateien direkt aus `docs/`. Es gibt keine zweite Kopie und kein Sync-Skript.
Agenten und Contributor arbeiten weiter mit denselben Dateien an denselben Pfaden;
die Website rendert exakt diese Dateien.

Damit Querverweise in beiden Welten gültig bleiben, schreiben Autoren und Agenten
weiterhin normale repo-relative Markdown-Links (`[compliance.md](compliance.md)`,
auch mit `#anker`). Ein remark-Plugin im Build wandelt diese Links in
Website-Routen um. So funktioniert die Link-Vorschau auf GitHub und im Editor
genauso wie die Navigation auf der Website.

Der Leitfaden wird aus einer langen Datei in eine Ordnerstruktur einzelner
Schritt-Seiten aufgeteilt und nach Zielgruppe und Thema gegliedert: einfacher
Standardweg getrennt vom optionalen Experten-Weg, Technik getrennt von Recht.

Das Setup folgt KISS und YAGNI. Bewusst ausgelassen: Analytics, E2E-Tests,
atomare Deploys, automatisiertes Deployment, Mehrsprachigkeit. Diese lassen sich
später ergänzen, ohne den Aufbau umzuwerfen.

## User Stories

### Lesen und Navigieren (Website-Besucher)

1. Als Vereinshelfer möchte ich die Doku als zusammenhängende, navigierbare
   Website mit Seitenleiste lesen, damit ich mich zurechtfinde, ohne lange
   Markdown-Dateien zu scrollen.
2. Als Vereinshelfer möchte ich eine Volltextsuche über die gesamte Doku,
   damit ich eine Antwort finde, ohne die Struktur zu kennen.
3. Als Vereinshelfer möchte ich die Doku auf dem Smartphone gut lesen können,
   damit ich sie auch beim Aufbau am Festplatz nutzen kann.
4. Als Besucher möchte ich einen Dunkelmodus, der meiner System-Einstellung
   folgt, damit das Lesen angenehm ist.
5. Als Besucher möchte ich pro Seite eine Gliederung („Auf dieser Seite"),
   damit ich innerhalb einer langen Seite schnell springen kann.

### Leitfaden im Doku-Look (Verein)

6. Als nicht-technischer Vereinshelfer möchte ich eine klare
   Schritt-für-Schritt-Anleitung für den Standardweg (Computer im Vereinsheim),
   damit ich jottis ohne Vorwissen in Betrieb nehmen kann.
7. Als Vereinshelfer möchte ich den einfachen Standardweg getrennt vom
   optionalen Experten-Weg (eigener Server) sehen, damit mich der komplexe Weg
   nicht verwirrt, wenn ich ihn nicht brauche.
8. Als Vereinsvorstand möchte ich die rechtlichen und steuerlichen Pflichten
   getrennt von der technischen Einrichtung finden, damit ich gezielt das Thema
   nachlesen kann, das mich gerade betrifft.
9. Als Vereinsvorstand möchte ich eine Checkliste der einmaligen und laufenden
   Pflichten, damit ich nichts vergesse.
10. Als Betreiber möchte ich die Muster-Verfahrensdokumentation als eigene
    Doku-Seite finden, damit ich sie für meine Instanz anpassen und bei einer
    Kassen-Nachschau vorlegen kann.
11. Als technisch versierter Betreiber möchte ich die Self-Hosting-Schritte
    (Ersteinrichtung, Aktualisieren, Backups, TSE-Sonderfälle) als eigenen
    Bereich, damit ich den Experten-Weg gehen kann, ohne den Standardweg zu
    durchsuchen.
12. Als Vereinshelfer möchte ich einen Fehlersuche- und FAQ-Bereich, damit ich
    häufige Stolpersteine selbst lösen kann.

### Produkt und Recht (Website-Besucher)

13. Als Interessent möchte ich Produktbeschreibung und Lizenzmodell als
    Doku-Seiten lesen, damit ich Positionierung, Zielgruppe und das
    Source-Available-Modell verstehe.
14. Als Verein in der Entscheidungsphase möchte ich die fiskalischen Grundlagen
    (KassenSichV, GoBD, TSE) und das aktuelle Steuerrecht nachlesen können,
    damit ich die Rechtslage einschätzen kann.

### Marketing-Landing (Website-Besucher)

15. Als Interessent möchte ich eine moderne, ansprechende Landing-Page mit
    klarer Value-Proposition, damit ich auf einen Blick verstehe, was jotti ist
    und für wen.
16. Als Interessent möchte ich von der Landing klar in den Leitfaden bzw. die
    Doku geführt werden, damit ich den nächsten Schritt finde.
17. Als Besucher möchte ich, dass die Marke (Grün, Montserrat, Logo) und die
    Botschaft der heutigen Seite erhalten bleiben, damit jotti wiedererkennbar
    bleibt.

### Eine Quelle der Wahrheit (Autor, Agent, Contributor)

18. Als Agent möchte ich die Doku-Inhalte weiterhin unter `docs/*.md` an
    stabilen Pfaden im Repo lesen, damit meine Arbeitsweise und die
    `AGENTS.md`-Referenzen unverändert funktionieren.
19. Als Autor möchte ich einen Doku-Inhalt an genau einer Stelle pflegen, damit
    Repo und Website nie auseinanderdriften.
20. Als Autor möchte ich beim Schreiben normale repo-relative Markdown-Links
    setzen, damit die Vorschau auf GitHub und im Editor funktioniert und ich
    keine Website-Routen kennen muss.
21. Als Leser auf der Website möchte ich, dass ein Link auf ein nicht
    veröffentlichtes Dokument (z. B. `handbuch.md`) auf die Repo-Quelle auf
    GitHub zeigt, damit der Link nicht ins Leere läuft.
22. Als Maintainer möchte ich an einer Stelle festlegen, welche `docs/`-Dateien
    veröffentlicht werden, damit interne Dokumente (Architektur, PRDs, Pläne,
    Maintainer-Runbooks) privat bleiben.
23. Als Maintainer möchte ich, dass der Build fehlschlägt, wenn ein Doku-Link
    auf ein nicht existierendes Ziel oder einen fehlenden Anker zeigt, damit
    tote Links nicht unbemerkt online gehen.

### Entwickeln und Betreiben (Maintainer)

24. Als Entwickler möchte ich die Seite lokal mit einem einzigen Befehl starten
    und live bearbeiten, damit ich Änderungen sofort sehe.
25. Als Entwickler möchte ich die Website-Abhängigkeiten über denselben
    pnpm-Workspace wie das Frontend verwalten, damit es eine Toolchain und ein
    Lockfile gibt.
26. Als Betreiber möchte ich die fertige Seite über das bestehende
    Docker-Compose- und nginx-Setup ausliefern, gebaut in einer Build-Stage,
    damit es kein `dist/` im Repo und keinen Host-seitigen Build braucht.
27. Als Betreiber möchte ich die Auslieferung manuell anstoßen wie bisher, damit
    ich kein automatisches Deployment pflegen muss.
28. Als Maintainer möchte ich, dass das alte SSI-, CSS-Mini-Design-System- und
    `check-website.sh`-Setup vollständig entfernt wird, damit es kein
    Parallelsystem gibt.

### Qualität und SEO

29. Als Betreiber möchte ich, dass jede veröffentlichte Doku-Seite Titel und
    Beschreibung für SEO und Open Graph hat, damit Suchmaschinen und
    Social-Previews funktionieren.
30. Als Betreiber möchte ich `sitemap.xml` und `robots.txt` automatisch bzw.
    gepflegt, damit die neuen Doku-Seiten indexiert werden.

## Implementation Decisions

### Technologie und Projektstruktur

- **Astro + Starlight**, ein neues pnpm-Workspace-Paket `website/` neben
  `frontend/`. Eine `pnpm install`, ein Lockfile.
- **Styling.** Tailwind CSS 4 für die Landing-Page. Die Doku nutzt das
  Starlight-Theme, gebrandet über dessen CSS-Custom-Properties (Markenfarbe,
  Schrift Montserrat self-hosted). Ein gemeinsames Token-Set definiert die Marke
  einmal und speist Landing und Doku, damit beide visuell konsistent sind.
- **Kein Client-Framework als Island**, solange keine echte Interaktivität es
  erfordert (YAGNI). Astro-Komponenten reichen; Starlight bringt Navigation,
  Suche und Mobile-Menü mit.
- **Suche.** Starlights eingebaute, vollständig statische Volltextsuche
  (Pagefind). Kein externer Dienst.
- **Sprache.** Die Website ist fest deutsch. Keine i18n-Mechanik: keine
  Locale-Präfixe in den Routen, kein Sprachumschalter, keine
  Übersetzungsdateien, keine doppelten Inhalte. Starlight wird einsprachig
  konfiguriert (ein Locale `de` als Root), wodurch auch die
  Oberflächen-Texte des Frameworks (Suche, „Auf dieser Seite") deutsch sind.
  Die Landing besteht aus reinen deutschen Astro-Seiten ohne i18n-Routing.
  jotti ist auf deutsches Recht und Finanzamt zugeschnitten, Zielgruppe sind
  deutsche Vereine; eine andere Sprache ist kein Ziel.
- **Bewusst ausgelassen (Phase 1):** Analytics, E2E-Tests, atomare Deploys,
  automatisiertes Deployment, ein Git-CMS für Nicht-Entwickler. (Deutsch ist
  keine Phasen-Entscheidung, sondern dauerhaft, siehe Out of Scope.)

### Eine Quelle der Wahrheit: docs/ bleibt kanonisch

- Das kanonische Markdown bleibt unter top-level `docs/`. Starlight liest die
  veröffentlichten Dateien direkt von dort (Astro-Content-Collection mit einem
  Glob-Loader, dessen Basis auf das Repo-`docs/`-Verzeichnis zeigt). Es gibt
  keine Kopie und kein Sync-Skript.
- **Veröffentlichte Inhalte (Phase 1):** der Leitfaden (siehe Restrukturierung),
  `compliance.md`, `steuerrecht.md`, `verfahrensdokumentation.md`,
  `produktbeschreibung.md`, `lizenzmodell.md`.
- **Privat (nicht veröffentlicht):** `handbuch.md`, `language.md`,
  `anforderungen.md`, `jotti-rocks-infra.md`, `docs/plans/`, `docs/prds/`. Diese
  bleiben rein interne Repo-Dokumente.
- **Veröffentlichungs-Auswahl an genau einer Stelle.** Welche Dateien publiziert
  werden, steht explizit in der Website-Konfiguration (Glob-Muster plus
  Sidebar-Definition), nicht in einem pro-Datei-Flag. KISS und an einem Ort
  überschaubar.
- **Frontmatter für veröffentlichte Dateien.** Veröffentlichte Dokumente
  erhalten minimales Frontmatter (`title`, `description`, optional Sidebar-Label
  und Reihenfolge). Die bisherige H1 entfällt zugunsten des aus dem Frontmatter
  gerenderten Titels, um doppelte Überschriften zu vermeiden. Tradeoff: GitHub
  rendert YAML-Frontmatter als kleine Tabelle am Seitenanfang; das ist
  akzeptabel. Private Dokumente bleiben unverändert ohne Frontmatter.

### Querverweise: remark-Link-Rewriter (testbares, tiefes Modul)

- Autoren und Agenten schreiben weiter repo-relative Markdown-Links
  (`[x.md](x.md)`, inklusive `#anker`). Ein remark-Plugin im Astro-Build
  übernimmt die Transformation.
- **Verhalten:**
  - Link auf ein veröffentlichtes Dokument wird zur entsprechenden
    Website-Route (inkl. Anker als Slug der Zielüberschrift).
  - Link auf ein nicht veröffentlichtes Dokument oder eine Repo-Datei außerhalb
    von `docs/` (z. B. `../TERMS.md`, `../LICENSE`, `../README.md`, `handbuch.md`)
    wird zu einer absoluten GitHub-URL auf die Quelle, damit der Verweis gültig
    bleibt.
  - Externe Links (`http`, `https`, `mailto`) bleiben unverändert.
- **Schnittstelle (rein, isoliert testbar):** Aus (Link-Ziel, Pfad des
  Quelldokuments, Map veröffentlichter Dokumente, Repo-Basis-URL) wird das
  korrekte `href` berechnet. Diese Funktion ist die deklarierte Kernlogik und
  wird isoliert getestet.
- **Link-Validierung.** Der Build bricht ab, wenn ein Link auf ein
  veröffentlichtes Ziel oder einen Anker zeigt, der nicht existiert (Starlights
  Link-Validierung bzw. eine entsprechende Build-Prüfung). Das ersetzt die
  Link-/Anker-Checks aus `scripts/check-website.sh`.

### Leitfaden-Restrukturierung

- `docs/leitfaden.md` wird in einen Ordner `docs/leitfaden/` aus einzelnen
  Schritt-Seiten aufgeteilt, gegliedert nach Zielgruppe und Thema. Damit ändert
  sich der Pfad des Leitfadens im Repo; betroffene Querverweise und die
  `AGENTS.md`-Referenztabelle werden entsprechend aktualisiert.
- **Vorgeschlagene Sidebar-Struktur (Doku):**
  - Erste Schritte: Was ist jotti?, Schnellstart (Standardweg)
  - Vereinsbetrieb (Standardweg): Installation und Start, TSE einrichten
    (fiskaly), Täglicher Betrieb, Checkliste
  - Recht und Steuern: Pflichten im Überblick, Kasse beim Finanzamt anmelden,
    Belege und Steuersätze, Steuerrecht Gastronomie, Datenaufbewahrung,
    Verfahrensdokumentation, Compliance-Grundlagen
  - Self-Hosting (Experten-Weg): Eigener Server (Ersteinrichtung), Aktualisieren
    und Backups, TSE-Sonderfälle
  - Hilfe: Fehlersuche, Häufige Fragen
  - Über jotti: Produktbeschreibung, Lizenzmodell
- Der Inhalt wird beim Aufteilen nicht neu erfunden, sondern umgegliedert und
  in Schritt-für-Schritt-Form gebracht. Die Trennung Standardweg/Experten-Weg
  und Technik/Recht wird über die Gruppierung hergestellt.

### Marketing-Landing

- Marke (Grün, Montserrat, Logo) und Botschaft der heutigen Seite (Hero,
  Vertrauensmerkmale, Features, Vergleich, CTA) bleiben inhaltlich erhalten und
  werden mit Astro und Tailwind modern neu aufgebaut. Kein Content-Workshop in
  Phase 1.
- Bilder über Astros Bild-Optimierung (moderne Formate, responsive Größen) statt
  manuell gepflegter Varianten.

### Build, Auslieferung und Aufräumen

- **Build und Serve.** Eine Multi-Stage-Build-Stage im bestehenden
  Compose-Setup baut die statische Seite; nginx serviert das Ergebnis. Kein
  `dist/` im Repo, kein Host-seitiger Build. Die Änderungen an
  `docker-compose.rocks.yml` und der nginx-Konfiguration sind Teil dieser Arbeit
  (die Auslieferung von `./website` als Volume wird durch das Build-Ergebnis
  ersetzt).
- **Deployment.** Manuell wie bisher. Kein automatisiertes Deployment in Phase 1.
- **Makefile.** Die Website-Targets werden ersetzt: lokaler Dev-Server
  (Astro dev), Build und der Qualitäts-Check (Typprüfung plus Build inkl.
  Link-Validierung) statt `check-website.sh`. Prettier-Formatierung bleibt über
  die bestehende Frontend-Installation.
- **Entfernen.** Das alte Setup wird vollständig zurückgebaut: SSI-Partials und
  ihre nginx-Konfiguration, das CSS-Mini-Design-System (`website/css/base.css`),
  `scripts/check-website.sh`, die handgebaute Leitfaden-HTML und die alte
  `nginx.website-dev.conf`. Es bleibt kein Parallelsystem.

### Auswirkungen auf bestehende Repo-Konventionen

- Die `AGENTS.md`-Referenztabelle und Querverweise, die auf `docs/leitfaden.md`
  zeigen, werden auf die neue `docs/leitfaden/`-Struktur aktualisiert.
- Das in `base.css` und der `make`-Hilfe beschriebene „statische Seite, kein
  Build-Schritt"-Prinzip entfällt; die Projektdoku zum Website-Workflow wird
  entsprechend angepasst.

## Testing Decisions

- **Was ein guter Test hier ist.** Getestet wird externes, stabiles Verhalten,
  nicht Implementierungsdetails. Für eine überwiegend statische Inhaltsseite
  heißt das: die wenige echte Logik isoliert prüfen und ansonsten den Build als
  Korrektheits-Gate nutzen.
- **Einziges Unit-getestetes Modul: der Link-Rewriter.** Seine reine Funktion
  (Link-Ziel + Quellpfad + Veröffentlichungs-Map -> `href`) wird mit Vitest
  isoliert getestet, über repräsentative Fälle: Link auf veröffentlichtes
  Dokument (mit und ohne Anker), Link auf privates Dokument (GitHub-URL), Link
  auf Repo-Datei außerhalb `docs/` (GitHub-URL), externer Link (unverändert).
  Vitest ist im Frontend bereits etabliert und dient als Vorbild.
- **Build als Gate statt zusätzlicher Tests.** Tote Links und fehlende Anker auf
  veröffentlichte Ziele, fehlendes Frontmatter und Typfehler werden durch
  `astro check` und den Build erkannt. Das ersetzt die selbstgebauten Prüfungen
  aus `scripts/check-website.sh`.
- **Bewusst keine E2E-Tests** (KISS, YAGNI). Visuelle Regression und
  Browser-Tests sind spätere Erweiterungen.

## Out of Scope

- Analytics jeder Art (auch self-hosted).
- E2E-Tests, visuelle Regressionstests, Lighthouse-CI-Gates.
- Atomare Deploys und Rollback-Mechanik.
- Automatisiertes Deployment (CI/CD-Pipeline für die Website).
- Mehrsprachigkeit, Übersetzungen und i18n-Routing. Die Website ist dauerhaft
  deutsch.
- Ein Git-CMS (z. B. Keystatic) für die Inhaltspflege durch Nicht-Entwickler.
- Veröffentlichung der internen Entwickler-/Maintainer-Dokumente
  (`handbuch.md`, `language.md`, `anforderungen.md`, `jotti-rocks-infra.md`,
  `docs/plans/`, `docs/prds/`).
- Inhaltliche Neukonzeption der Marketing-Botschaft; Marke und Aussage bleiben.

## Further Notes

- **Warum Starlight statt Docusaurus oder GitBook.** Starlight ist Teil der
  Astro-Toolchain und teilt sich Build, Komponenten und Design-Token mit der
  Landing. Docusaurus wäre ein zweiter React-Build mit eigenem Theme und doppelter
  Pflege. GitBook ist gehostetes SaaS mit Lock-in und Inhalten außer Haus, was dem
  self-hosted-, datensouveränen Charakter des Projekts widerspricht.
- **Erweiterbarkeit.** Spätere Phasen können ohne Umbau ergänzen: ein Git-CMS
  für nicht-technische Redakteure, CI-Qualitäts-Gates und automatisiertes
  Deployment.
- **Risiko Glob-Loader auf externes Verzeichnis.** Starlight aus dem
  `website/`-Paket auf das top-level `docs/` zeigen zu lassen, ist der eine
  technische Knackpunkt (Pfad-Basis des Loaders, Datei-Watching im Dev-Modus).
  Falls sich das als unzuverlässig erweist, ist der Fallback ein im Build
  erzeugter Symlink vom Content-Verzeichnis auf `docs/`. Die
  Quelle-der-Wahrheit-Garantie (eine Datei, in `docs/`) bleibt in beiden Fällen
  erhalten.
