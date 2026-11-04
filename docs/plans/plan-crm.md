# Plan: Privates CRM für Vereinspost und Feldfeedback

> Source PRD: n/a

## Goal

Vereinsmails (Gmail-Label `jotti`) und GitHub-Issues werden mit einer wiederholbaren Routine
eingelesen, einer Organisation zugeordnet, klassifiziert und beantwortet. Personenbezogene Daten
liegen nur im gitignorierten Ordner `.crm/`. Getrackt sind nur Code, Schema, die Routine und
anonymisierte Zahlen. `docs/feldfeedback.md` wird aus dem Register erzeugt statt von Hand gezählt.

Auslöser: Die Handzählung hat beim ersten Abgleich zwei Organisationen aus GitHub-Issues verpasst.
Der Mail-Index enthält 13 Fehlzuordnungen, und 8 Antworten sind unbemerkt offen. Die Einzelfälle
stehen nur in `.crm/`.

Von Nico entschieden: TOML, minimaler Datenumfang, Ordner `.crm/`, Rohkopien nach Abgleich löschen,
Python im Repo-Gate, Datenschutzerklärung in Phase 4 ergänzen, Löschung nur lokal, Backup über
Nicos eigene Sicherung.

## Architectural decisions

- **Kein CRM-Produkt.** Ein SaaS-CRM wäre ein weiterer Auftragsverarbeiter für Vereinsdaten und
  verlangte einen AV-Vertrag sowie eine Änderung der Datenschutzerklärung. Bei rund 40
  Organisationen und einem Maintainer reichen zwei Textdateien und ein Skript.
- **Gmail und GitHub sind die Ablage, `.crm/` ist das Register.** Die Annahme-Mail in Textform liegt
  in Gmail, öffentliche Issues liegen auf GitHub. Das Register hält nur die extrahierten Fakten
  und verweist mit `gmail:<message-id>` oder `gh:<issue-nummer>` auf die Quelle. Lokale `.eml`-Kopien
  braucht es nicht. Gründe:
  - Jede lokale Kopie muss die Löschregel mitmachen; ohne Kopien bleibt nur das Register.
  - Rohbytes liefert der Gmail-Connector nur über das Session-Transkript. Zum Lesen genügt
    `get_thread` mit `PLAIN_TEXT`.
- **Format TOML, Werkzeug Python-Stdlib.** `tomllib` liest TOML ohne Paket-Abhängigkeit
  (Python ≥ 3.11). Der CI-Job `repo-checks` nutzt das System-Python von `ubuntu-latest` (3.12).
  YAML bräuchte PyYAML. Das Skript schreibt das Register nie. Claude oder Nico pflegen es per
  Editor, das Skript prüft und wertet aus.
- **`CRM_DIR`** ist der Ort des Registers. Vorgabe:
  `$(git rev-parse --path-format=absolute --git-common-dir)/../.crm`. So finden auch Worktrees unter
  `../jotti-wt/` das Register des Haupt-Checkouts.
- **Datensparsamkeit im Register.** Gespeichert werden die Kategorien der Datenschutzerklärung:
  Organisation, Art, Sitz, Ansprechperson mit E-Mail, Nachrichteninhalt als Kurzfassung. Dazu
  kommen Rolle der Ansprechperson, Land und die Einwilligung zur Namensnennung. Phase 4 ergänzt die
  Datenschutzerklärung darum und um GitHub-Issues als Kontaktweg. Nicht gespeichert werden Telefon,
  Postanschrift, Registernummern und Vorstandslisten aus Signaturen; der Zweck braucht sie nicht.
- **Zwei Ebenen der Klassifikation.**
  - Eine *Meldung* ist ein Anliegen einer Organisation aus einer Quelle. Jede eingelesene Nachricht
    erzeugt mindestens eine Meldung; die Annahme-Mail erzeugt eine Meldung der Art `verwaltung`.
    Die Quellen aller Meldungen sind zugleich die Liste bekannter Nachrichten (Duplikatprüfung).
  - Ein *Thema* bündelt gleichartige Meldungen und entspricht einer Zeile in `docs/feldfeedback.md`.
  - Die Zahl „Vereine" ist die Zahl verschiedener Organisationen unter den Meldungen eines Themas.
- **Schlüssel englisch, Werte deutsch** (Fachbegriffe):
  - `org.kind`: `verein`, `person`, `dienstleister`. Dienstleister fallen aus dem PII-Gate, weil
    Doku sie legitim nennt.
  - `org.status`: `anfrage`, `vereinbart`, `zurueckgetreten`, `bestritten`, `erledigt`.
    `vereinbart` verlangt `agreed_on`, `terms_version`, `agreement_source`.
  - `topic.kind`: `wunsch`, `fehler`, `frage`, `einsatzbericht`, `verwaltung`. Ideen von
    Nicht-Vereinen sind `wunsch`, Lob ist `einsatzbericht`. `verwaltung` erscheint nicht im Render.
  - `topic.status`: `offen`, `umgesetzt`, `abgelehnt`, `beantwortet`. `beantwortet` speist
    „Wiederkehrende Fragen", die übrigen die gleichnamigen Abschnitte.
  - `topic.ref` ist freies Markdown (Link auf `Dnn`, `backlog.md`, Leitfaden oder Text);
    `check-links` prüft es nach dem Render.
  - `mention.answered`: `false`, solange der Verein auf eine Antwort wartet. Einziges
    Offen-Kennzeichen.
- **Löschregel nach Datenschutzerklärung.** `crm.py retention` listet Organisationen, deren
  Korrespondenz zu löschen ist:
  - `zurueckgetreten`, `bestritten`, `erledigt`: sofort;
  - `anfrage` nur, wenn alle Meldungen `answered = true` sind und Nico sie auf `erledigt` setzt;
  - `vereinbart`: erst nach Beendigung nach TERMS § 8.
  Gelöscht werden der Eintrag im Register und lokale Kopien. Gmail bleibt außen vor (Nicos
  Entscheidung).
- **Erfassungsregeln** aus der Durchsicht aller Mails, Anhänge und Issues:
  - Eine Vereinbarung gilt nur mit Annahmesatz in Textform. Signierte PDFs, freie Briefe und
    Formularmails ohne Satz bleiben `anfrage`.
  - `terms_version` muss eine Fassung aus `[terms] versions` in `register.toml` sein. Zitiert
    wurden drei Fassungen, eine davon eine Woche nach Erscheinen der nächsten.
  - Die Annahme steht oft in einem Thread mit anderem Thema. `agreement_source` darf jede
    Nachricht sein, und dieselbe Nachricht trägt weitere Meldungen.
  - Eine Nachricht, die nur als Zitat in einer späteren Mail vorliegt, bekommt deren Quelle.
  - Rechtsform und Land sind Angaben des Vereins. Vereine aus Österreich und der Schweiz wählen
    im Formular „e.V."; `country` folgt dem Sitz, Zweifel stehen in `notes`.
  - Kontakt-E-Mail und Absender weichen oft ab, einer schreibt im Auftrag eines anderen. Ein
    Kontakt hat deshalb eine Liste `emails`. Internationalisierte Domains stehen in Unicode.
  - Anhänge kommen nie ins Register. Registerauszüge und Signaturen nennen Dritte mit
    Geburtsdatum oder Amt; sie bleiben in Gmail. Archive aus Mails sind fremder Code: nur in
    einem leeren Temp-Ordner entpacken, nie ausführen, danach löschen. Die Erkenntnis geht als
    Kurzfassung in die Meldung.
- **PII-Grenze als Gate.** `scripts/check-private-data.sh` bekommt die Suchbegriffe von
  `crm.py pii-terms`: E-Mail-Adressen, Kontaktnamen, Organisationsnamen ohne `public_name_ok`,
  ohne Dienstleister, jeweils mindestens 5 Zeichen. GitHub-Logins fehlen, sie sind auf den Issues
  ohnehin öffentlich. Es läuft über `check-repo` im Pre-Push-Hook. Ohne Register (CI) endet es
  mit 0 und einem Hinweis.
- **Der Renderer gibt nur Whitelist-Felder aus:** Thema-Titel, Status, Ref, Anzahl, Summen je
  Land und Status, dazu die PII-frei formulierten Lage-Sätze aus `[lage]` in `topics.toml`. Das
  Gate fängt Fehler in diesen Texten ab.
- **Routine als Projekt-Skill** `.claude/skills/crm/SKILL.md`.
  - Quellen: Gmail-Connector und `gh`. Cursor: `[sync]` in `register.toml`.
  - Keine autonomen Antworten und keine Issue-Labels ohne Nicos Zustimmung; Entwürfe und
    Label-Vorschläge bleiben im Chat.
  - IMAP mit App-Passwort verworfen: voller Postfachzugriff für ein Skript. Das Klassifizieren
    braucht ohnehin eine Claude-Session.

### Dateien in `.crm/` (gitignoriert)

Alle Werte unten sind erfunden.

```toml
# register.toml
[sync]
gmail_through = 2026-01-31        # Datum der neuesten eingelesenen Nachricht
gh_through = 2026-01-31

[terms]
versions = [2026-06-03, 2026-07-14, 2026-09-07]   # Fassungen von TERMS.md

[[org]]
id = "beispielverein"             # slug, eindeutig
kind = "verein"
name = "Beispielverein e.V."      # wie in der Annahme-Mail
legal_form = "e.V."               # wie angegeben; Zweifel in notes
seat = "Musterstadt"
country = "DE"
status = "vereinbart"
agreed_on = 2026-01-15
terms_version = 2026-09-07
agreement_source = "gmail:0000000000000001"
public_name_ok = false
gh = []
notes = ""

  [[org.contact]]
  name = "Erika Mustermann"
  role = "Vorsitzende"
  emails = ["vorstand@example.org"]
```

```toml
# topics.toml
[lage]
notes = ["Als Wechselgrund nennen die Vereine teure Miet-Kassensysteme."]

[[topic]]
id = "glaeserpfand"
title = "Gläserpfand: Rückgabe und Verrechnung"
kind = "wunsch"
status = "offen"
ref = "[backlog.md](backlog.md)"

  [[topic.mention]]
  org = "beispielverein"
  source = "gmail:0000000000000002"
  on = 2026-01-20
  answered = false
```

## Inventory

- `.mails/README.md` — Index, Anfragen je Verein, Datenschutzregel. Wird zu `.crm/README.md`.
- `.mails/raw/*.eml` — 49 Mails und 5 Anhänge, Quelle der Erstbefüllung.
- `.mails/extract_gmail_raw.py` — Rohbytes aus dem Transkript. Wird mit den Rohkopien gelöscht.
- `.gitignore` — Zeile `.mails/`.
- `docs/feldfeedback.md` — heute Handzählung, Prettier-formatierte Tabellen. Wird erzeugt.
- `docs/backlog.md` — nennt Vereinszahlen in Prosa. Das doppelt die Zählung.
- `website/src/pages/datenschutz.astro` — Datenkategorien und Speicherdauer.
- `website/src/lib/anfrage-mailto.ts` — `buildMailtoUrl()`. Liefert die festen Felder der Annahme-Mail.
- `TERMS.md` — Prozess und § 8 Beendigung. Grundlage für `org.status`.
- `Makefile` — `check-repo` führt jedes `scripts/check-*.sh` aus. `.githooks/pre-push` ruft es auf.
- `scripts/lib.sh` — gemeinsame Helfer der Check-Skripte.
- `scripts/setup-dev-tools.sh` — Entwicklerwerkzeuge. Bekommt die Prüfung auf Python ≥ 3.11.
- GitHub-Issues #55, #58, #144, #145, #146 — externe Meldungen, in keinem Index.

## Open questions / Risks

- **Python im Repo-Gate** ist freigegeben. macOS-System-Python 3.9 reicht nicht; die Skripte
  brechen dort mit klarer Meldung ab.
- **Lokale Rohkopien löschen** ist freigegeben, als Handschritt nach Phase 1, kein Phase-Kriterium:
  1. In einer Session jede `.eml` in `.crm/mail/` per `rfc822msgid:` in Gmail suchen und ihre
     Quelle im Register bestätigen.
  2. Danach `.crm/mail/` und `.crm/extract_gmail_raw.py` löschen.
  3. Den Abschnitt zu Rohkopien aus `.crm/README.md` streichen.
- **Gelöschte Fälle.** Die lokalen Kopien eines zurückgetretenen Vereins und eines Absenders, der
  die Mail abgestritten hat, sind gelöscht. Das Register führt sie nur als Zahl.
- **Backup von `.crm/`** läuft über Nicos eigene Sicherung des Home-Verzeichnisses. `.crm/README.md`
  nennt das. Fehlt sie, lässt sich das Register aus Gmail und GitHub neu aufbauen.
- **Fälle mit „Ein Verein"** in `docs/backlog.md` beschreiben den Anlass eines Eintrags, keine
  Zählung. Sie bleiben. Gestrichen werden Mehrzahl-Zählungen.

## Phase 1: Register, Validator und Erstbefüllung

**Depends on**: none

### Context

- `.mails/README.md` — Index und Anfragetabelle. Der Mail-Digest fand dort 13 Fehlzuordnungen:
  Verträge ohne Annahmesatz als Vereinbarung geführt, Support als Fehler geführt, Themen Vereinen
  zugeschrieben, die sie nicht genannt haben.
- `.mails/raw/*.eml` — Quelle. Gmail-IDs liefert `search_threads label:jotti`; eine Datei findet
  ihre ID über `rfc822msgid:<Message-ID>`.
- GitHub-Issues #55, #58, #144–#146 — fehlen in Index und Zählung.
- `.gitignore` — Zeile `.mails/`.

### What to build

- Umzug in dieser Reihenfolge, damit kein privater Ordner ungeschützt ist:
  1. `.crm/` in `.gitignore` ergänzen, `.mails/` stehen lassen.
  2. `.mails/` nach `.crm/` verschieben, `raw/` wird `.crm/mail/`.
  3. Erst wenn `.mails/` nicht mehr existiert, die Zeile `.mails/` entfernen.
  4. Den Verweis im Kopf von `docs/feldfeedback.md` anpassen.
- `scripts/crm/crm.py validate` prüft mit `tomllib`:
  - eindeutige `id`, Enum-Werte, Pflichtfelder je `status`
  - Quellformat `gmail:[0-9a-f]{16}` oder `gh:[0-9]+`, jede Quelle höchstens einmal je Thema
  - jede `mention.org` existiert
  - `terms_version` steht in `[terms] versions`
  - jeder Kontakt hat mindestens eine E-Mail in `emails`
  - kein Feld außerhalb des Schemas: `phone`, `address` und Unbekanntes werden abgewiesen
  - `CRM_DIR` nach der Architekturentscheidung, überschreibbar per Umgebung
- `crm.py` bricht unter Python 3.11 mit klarer Meldung ab. `scripts/setup-dev-tools.sh` prüft
  die Version.
- `.crm/register.toml` und `.crm/topics.toml` aus allen 49 Mails und 5 Issues befüllen. Die zwei
  gelöschten Fälle gehen nur als Zahl ein: `[lage]` bekommt einen Zähler `deleted = 2`. Die
  Klassifikation folgt den Architekturentscheidungen.
- `scripts/crm/test_crm.py` (stdlib `unittest`) mit erfundenen Fixtures: gültiges Register
  besteht, Schema-Verstöße werden abgewiesen.
- `scripts/check-crm.sh` führt die Tests in `check-repo` aus.
- `.crm/README.md` beschreibt Layout, Schema-Kurzform, Löschregel und Datenschutzregel. Index- und
  Anfragetabellen entfallen, das Register ersetzt sie.

### Acceptance criteria

- [x] `python3 scripts/crm/crm.py validate` endet mit 0 und meldet die Zahl der Organisationen und Themen.
- [x] `python3 -m unittest scripts/crm/test_crm.py` grün, inklusive Abweisung eines `phone`-Felds, einer unbekannten `mention.org` und einer unbekannten `terms_version`.
- [x] `git check-ignore -q .crm/register.toml && test ! -e .mails` endet mit 0.
- [x] `make check-repo` grün.

## Phase 2: PII-Gate

**Depends on**: 1

### Context

- `Makefile` — `check-repo`, `.githooks/pre-push`.
- `scripts/lib.sh` — Helfer der Check-Skripte.
- `scripts/crm/crm.py` — bekommt `pii-terms`.

### What to build

- `crm.py pii-terms` gibt eine Zeile je Suchbegriff aus, nach der Regel der
  Architekturentscheidung.
- `scripts/check-private-data.sh`:
  - Kein Register oder keine Begriffe: Hinweis, Ende mit 0.
  - Sonst `git grep -n -I -F -i -w -f -` über alle getrackten Dateien, Ausgabe durch
    `cut -d: -f1,2`. So erscheinen nur Datei und Zeile, nie der Begriff.
  - Rückgabe 1 von `git grep` heißt kein Treffer, also bestanden. 0 heißt Treffer, also Fehlschlag.
    Jeder andere Code ist ein Fehler.
- Tests in `test_crm.py`:
  - `pii-terms` liefert für das Fixture genau die erwarteten Begriffe, ohne öffentlichen Namen,
    ohne Dienstleister, ohne Begriffe unter 5 Zeichen.
  - Das Skript in einem temporären Git-Repo: Eine eingeschleuste Kontakt-E-Mail endet mit 1 und
    druckt die E-Mail nicht.

### Acceptance criteria

- [x] `bash scripts/check-private-data.sh` endet mit 0 auf dem aktuellen Stand und meldet die Zahl der geprüften Begriffe größer 0.
- [x] `python3 -m unittest scripts/crm/test_crm.py` grün, inklusive des Falls mit eingeschleuster Kontakt-E-Mail.
- [x] `CRM_DIR=/nonexistent bash scripts/check-private-data.sh` endet mit 0 und dem Hinweis.
- [x] `make check-repo` grün.

## Phase 3: `docs/feldfeedback.md` aus dem Register

**Depends on**: 2

### Context

- `docs/feldfeedback.md` — Abschnitte Lage, Umgesetzt, Abgelehnt, Offen, Wiederkehrende Fragen;
  Tabellen in Prettier-Ausrichtung.
- `docs/backlog.md` — Vereinszahlen in Prosa.
- `scripts/crm/crm.py` — bekommt `render`.

### What to build

- `crm.py render` schreibt `docs/feldfeedback.md` vollständig:
  - fester Kopftext im Skript, der `check-prose` besteht (Sätze ≤ 20 Wörter, Absätze ≤ 3 Zeilen);
  - Lage aus Summen je `org.status` und `country`, dazu die Sätze aus `[lage]`;
  - eine Tabelle je `topic.status`, sortiert nach Anzahl absteigend, dann Titel.
- `crm.py render --check` rendert in eine temporäre Datei, formatiert sie mit Prettier und
  vergleicht mit der Datei. Ende mit 1 bei Abweichung.
- `make crm` ruft `validate`, `render`, Prettier auf `docs/feldfeedback.md` und
  `check-private-data.sh` nacheinander auf.
- `docs/backlog.md`: Mehrzahl-Zählungen von Vereinen aus der Prosa streichen. Die Zahl steht nur
  noch in `feldfeedback.md`.
- Test: `render` auf dem Fixture ergibt den erwarteten, Prettier-formatierten Text als Literal. Ein
  Kontaktname und ein `verwaltung`-Thema aus dem Fixture erscheinen nicht.

### Acceptance criteria

- [x] `make crm` endet mit 0.
- [x] `python3 scripts/crm/crm.py render --check` endet mit 0 direkt nach `make crm`.
- [x] `git grep -n -i -E '\b(zwei|drei|vier|fünf|[0-9]+) Vereine\b' docs/backlog.md` findet nichts.
- [x] `python3 -m unittest scripts/crm/test_crm.py` grün.
- [x] `make check-repo check-format` grün.

## Phase 4: Sync-Routine, offene Antworten, Löschregel

**Depends on**: 3

### Context

- Gmail-Connector: `search_threads`, `get_thread` mit `PLAIN_TEXT`.
- `gh issue list --state all --json`, `gh issue view --comments`.
- `website/src/pages/datenschutz.astro` — Abschnitt „E-Mail-Anfragen" und „Speicherdauer".
- `AGENTS.md` — Regelabschnitt.

### What to build

- `crm.py open` listet Meldungen mit `answered = false`, ältestes zuerst. `crm.py retention`
  listet Organisationen nach der Löschregel. Beide geben nur `id`, Quelle und Datum aus.
- Tests für `open` und `retention` auf dem Fixture, je ein Fall pro Status.
- `.claude/skills/crm/SKILL.md`, Schritte:
  1. Gmail `label:jotti after:<gmail_through>` und `gh issue list` seit `gh_through` holen,
     dazu neue Kommentare auf offenen Issues. Eine Quelle, die schon in einer Meldung steht,
     wird übersprungen.
  2. Je neue Nachricht: Organisation anlegen oder zuordnen, Meldungen einem Thema zuordnen oder ein
     Thema anlegen, `answered` setzen.
  3. Cursor setzen, `make crm` laufen lassen.
  4. `crm.py open` und `crm.py retention` berichten. Antwortentwürfe und Label-Vorschläge
     (`bug`, `enhancement`, `question`) nur im Chat, ausgeführt erst nach Nicos Zustimmung.
- `AGENTS.md` bekommt eine Regel: Vereinsdaten nur in `.crm/`, Feldfeedback über den Skill `crm`.
- `datenschutz.astro`: GitHub-Issues als Kontaktweg und die Kategorien Rolle, Land und
  Einwilligung zur Namensnennung ergänzen. Die Speicherdauer bleibt, die Löschregel folgt ihr.

### Acceptance criteria

- [x] `python3 scripts/crm/crm.py open` und `python3 scripts/crm/crm.py retention` enden mit 0 und geben je mindestens eine Zeile aus.
- [x] `python3 -m unittest scripts/crm/test_crm.py` grün, inklusive der Fälle für `open` und `retention`.
- [x] `test -f .claude/skills/crm/SKILL.md && grep -q '\.crm/' AGENTS.md` endet mit 0.
- [x] `git grep -n -i 'GitHub' website/src/pages/datenschutz.astro` findet den neuen Kontaktweg.
- [x] `make check-repo check-format` und `make website-check` grün.

