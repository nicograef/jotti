# Backlog

Offene Produktentscheidungen und Folgekandidaten ohne Release-Bezug.

## Produktentscheidungen

- [ ] Kassenabschluss-Wiederanlauf nach einer Geldbewegung: Steht ein Kassensturz und bewegt sich
      danach Geld, bricht jeder Wiederanlauf mit `ErrBuchungenNachKassensturz` ab. Geldbewegungen
      sind Tischzahlung, Direktverkauf, Storno mit Rückgabe und Geldtransit. Kein Befehl hebt einen
      Kassensturz auf. Die Kassensitzung bleibt offen, das Frontend sagt „Bitte den Administrator kontaktieren".
      Reparaturweg (z. B. zweiter Kassensturz beim Wiederanlauf) bauen oder Admin-Ablauf
      dokumentieren.
- [ ] `GetEigeneUebersicht` im Barrierestatus: liest weiter `GetOffeneKassensitzungNr`; die eigene
      Übersicht der Servicekraft zeigt Nullen, solange die Kassensitzung `wird_abgeschlossen` ist.
      Auf `GetAktiveKassensitzung` umstellen oder die Nullen hinnehmen.
- [ ] Zeichen gegen Bytes an Kommentar- und Tischnamen-Feldern: Zod zählt Code-Units, die
      zog-Schemas im Backend Bytes. Ein Umlaut-Kommentar mit 100 Zeichen passiert daher das
      Frontend und bekommt 400. Betreiber-Felder zählen Zeichen. Frontend zählt Bytes
      (`TextEncoder`) oder die UI nennt die Einheit.
- [ ] Login-Formular: `frontend/src/lib/AuthBackend.ts` prüft den Login mit dem vollen
      `PasswordSchema` (min 6/max 72). Das verrät die Passwort-Policy; der Login-Endpunkt prüft
      nur trim + min 1. Eigenes Login-Schema (trim + min 1).
- [ ] Windows: `jotti-restore.cmd` liest nur `jotti-*.sql` aus dem Volume; für `manuell-*.sql` auf
      dem Host gibt es keinen dokumentierten Weg. Restore-Weg dokumentieren oder das Skript
      erweitern.
- [ ] `JOTTI_DOMAIN` bei jedem Compose-Befehl des Public-Stacks: Folge der Entscheidung Option A —
      jeder Compose-Befehl mit `docker-compose.prod.yml` (auch `make prod-down`, `make prod-logs`)
      verlangt die Variable. Option A (hinnehmen) laut statt still bestätigen.
- [ ] Laufende Bewirtung im Vereinsheim: Ein Verein setzt jotti auch dafür ein. Das ist nicht die
      Zielgruppe (2–3 Feste pro Jahr), wird aber nicht verhindert. Ob das ein Nicht-Ziel wird, ist
      offen.
- [ ] Höchstmenge je Produkt: Ein Verein verkauft Abholbons und will je Produkt eine Höchstmenge
      hinterlegen; ist sie erreicht, sperrt der Verkauf. So gehen nicht mehr Bons raus, als Ware da
      ist. Heute schaltet der Admin eine Variante von Hand ausverkauft; eine automatische Grenze
      fehlt. Warenwirtschaft ist Nicht-Ziel; offen ist, ob ein Verkaufsstopp je Variante darunter fällt.
- [ ] Gläserpfand, Rabatt und Gutschrift: Drei Vereine fragen nach Pfandrückgabe, einer davon auch
      nach Rabatt und Gutschrift am Tisch. Ein Pfand-Produkt mit positivem Preis geht heute. Rückgabe,
      Rabatt und Verrechnung mit neuen Positionen gehen nicht, weil kein Preis negativ sein darf. Der
      DSFinV-K-Export schreibt Preisfindung (`itemamounts.csv`) und Zusatzinfo (`subitems.csv`) nur
      als Kopfzeile. Weg wählen (Rückgabe-Vorgang, Negativ-Variante oder Preisfindung je Position).
      Die Umsatzsteuer von Rückgabe und Rabatt vorher gegen `docs/rechtsquellen/` prüfen.
- [ ] Offene Tische automatisch unter „Meine Tische": Ein Verein wünscht sich Tische mit eigenen
      offenen Bestellungen ohne Stern auf dem Dashboard.
- [ ] Letzter Admin ausgesperrt: Ein Verein hat das Passwort seines einzigen Admins vergessen. Andere
      Konten setzt ein Admin per Einmalpasswort zurück; für den letzten Admin gibt es keinen Weg.
      Wiederherstellung über die Startkonsole bauen oder einen Weg dokumentieren.
- [ ] Installation ohne Docker: Docker Desktop und Netzwerk-Know-how sind die größte Hürde der
      Vereine. Vorschläge: ein Binary mit SQLite und eingebettetem Frontend oder ein gehosteter Betrieb
      mit Freischaltung je Verein. Beides berührt D13, D15 und die Freeze-Regeln. Bekannte Kosten von
      SQLite: Das eingefrorene `01_initial.up.sql` ist reines Postgres (plpgsql-Trigger, `REVOKE`,
      `GENERATED … AS IDENTITY`, JSONB). SQLite bräuchte eine zweite Schema- und Migrationslinie mit
      nachgebauten Schutz-Triggern. sqlc generiert für die Postgres-Engine, die Abfragen nutzen
      `LATERAL`, JSONB und Zeilensperren. Die Abschluss-Barriere (`FOR SHARE` gegen `UPDATE`) und die
      Passwort-Sperre (`FOR UPDATE`) müssten auf SQLites einzigen Schreiber umziehen. Ein
      Replikations-Backup mit einer Woche Aufbewahrung ersetzt nicht die zehn Jahre nach § 147 AO.
      Ein Binary mit eigenem TLS ersetzt Caddy (D18). HTTP-01 scheitert im LAN, deshalb holt jotti
      Zertifikate per DNS-01 über acme-dns. Der gehostete Betrieb widerspricht TERMS § 1 (2) und
      § 3 (2): kein SaaS, kein Betrieb für Dritte.
- [ ] Bestellkommentar an fremden Stationen (#145): Der Kommentar hängt an der Bestellung, nicht
      an der Position. Die Arbeitsbon-Policy druckt ihn deshalb an jede Station, bei
      `pro_position` auf jeden Einzelbon. Bei einem Verein landeten so 44 % der Kommentare auf
      mindestens zwei Druckern. Wege: auf dem Bon als Hinweis zur ganzen Bestellung kennzeichnen,
      ohne Datenänderung. Oder ein Kommentar je Position als neue Event-Version.
- [ ] Ausgabestationen unabhängig von Kategorien (#146): `druckstationen` hat eine Zeile je
      Kategorie, also höchstens drei Produktstationen. Ein Verein legte deshalb ein Gericht unter
      Getränke ab, was die Produktstatistik verfälscht. Ein Direktverkauf ohne Bon lässt sich
      nicht einstellen, und nur „Essen" piept. Vorschlag: Station als eigener Datensatz, Produkt
      schlägt Kategorie. Kleinster Teil: Aus-Schalter für den Direktverkauf.
- [ ] Berechtigung in Grenzfällen: TERMS § 2 (1) verlangt eine eingetragene, nicht-gewinnorientierte
      Organisation. Eine Untergliederung eines Verbands ohne eigene Eintragung kannte ihre Rechtsform
      nicht. Schweizer Vereine sind auch ohne Eintrag rechtsfähig (Annahme, gegen das ZGB prüfen). Ein
      Verein fragt, ob ein Partnerverein seine Installation mitnutzen darf. § 3 erlaubt nur eigene
      Veranstaltungen und schließt den Betrieb für Dritte aus. Technisch trägt eine Installation genau
      einen Betreiber mit einer Steuernummer auf allen Belegen. Klarstellung in TERMS oder FAQ
      entscheiden; eine TERMS-Änderung ist eine neue Fassung.
- [ ] Probelauf auf dem späteren Kassenrechner: Das Kassenjournal ist append-only, Kassensitzungs-
      und Z-Nummern zählen fortlaufend (K-16). Eine Probe-Kassensitzung auf dem Rechner, der danach
      live kassiert, bleibt vermutlich im Export. Für Windows gibt es keinen dokumentierten Weg zum
      sauberen Neubeginn. Prüfen, dann Leitfaden-Hinweis oder Reset vor dem ersten Echtbetrieb.

## Folgekandidaten

- [ ] `export.go` liest die Kassensitzung mit `GetOffeneKassensitzung` und `//nolint:forbidigo`;
      wegen der Sortierung von `GetAllKassensitzungen` liefert `GetAktiveKassensitzung` dieselbe
      Sitzung — die Ausnahme kauft kein Verhalten. Auflösen und ersetzen.
- [ ] `scripts/check-ui-labels.sh` prüft das Vorkommen zitierter Bedienelemente in `frontend/src`;
      ein Zitat, das nur als Bezeichner oder Kommentar vorkommt, läuft durch. Treffer in
      Bezeichnern und Kommentaren ausschließen — schärfer ginge es nur gegen gerenderte Texte.
- [ ] `dsfinvkpruefung` prüft Typen und Dezimalformat, nicht `MaxLength`; die Feldlängen sichern
      allein die Mapper-Tests. `MaxLength` ergänzen.
- [ ] `backend/repository/repotest/kassenjournal.go` kann keine Summe der Differenzbuchungen ≠ 0
      abbilden; den unterscheidenden Fall deckt nur der Integrationstest. Abbildbar machen.
- [ ] `tisch_sessions.unbezahlte_positionen` persistiert `[]kasse.Position` (ohne JSON-Tags,
      PascalCase-Schlüssel) als Projektion; außerhalb des Event-Vertragsgates, ein Umbenennen der
      Felder wäre ein Bruch persistierter Daten. Guard ergänzen oder in `docs/decisions.md`
      festhalten.
- [ ] `api/auth/http/command_handler.go` ruft `user.PasswordSchema.Required()` auf dem geteilten
      Paket-Schema auf (zog mutiert in place) — latente Falle, heute ohne Auswirkung. Eine Kopie
      verwenden.
- [ ] Status-Seite: Eine alte `install.json` mit einer Subdomain, die den Subdomain-Regex verletzt,
      lässt den Proxy nur mit der internen CA weiterlaufen. Die Status-Seite rät dann zu einem
      Neustart, der nichts ändert. Der Fall ist theoretisch, weil acme-dns Kleinbuchstaben-UUIDs
      vergibt.
- [ ] Bondrucker hängt still: Im Einsatz druckte ein WLAN-Bondrucker zweimal nach 1–1,5 h nichts
      mehr, bis er neu startete. Relay und Kasse meldeten keinen Fehler. Das Relay fragt vor dem Druck
      nur den Papiersensor ab, eine fehlende Antwort gilt als OK. Kandidat: Offline-Status abfragen
      und Schweigen als Fehler werten.
- [ ] Umlaute auf Epson-Bondruckern (#144): `SetCodepageWPC1252` sendet `ESC t 6`. Am MUNBYN ist
      das WPC1252, am Epson TM-T20IV Hiragana; Umlaute und ß werden dort Zeichensalat. Das trifft
      Arbeitsbon, Testbon und Kassenbeleg. Kandidat ohne Einstellung: PC437 per `ESC t 0`, denn
      die Bons drucken „EUR" statt €. Auf beiden Druckern prüfen; der Testbon druckt eine
      Umlautzeile mit.
- [ ] Migration vorab gegen eine Kopie prüfen: `make prod-update` sichert vor der Migration und nennt
      bei einem Fehlschlag den Restore-Weg. Eine Migration, die an Bestandsdaten scheitert, fällt aber
      erst im Live-Stack auf. Dort bleibt `schema_migrations` auf `dirty = true`, und der Stack fährt
      nicht hoch. `scripts/prod-backup-verify.sh` spielt den Dump schon in ein Wegwerf-Postgres. Vor
      `up -d` dort das `jotti-migrate`-Image der Zielversion laufen lassen. Dessen `CMD` erwartet den
      Host `postgres`; der Windows-Starter zieht nach.
- [ ] Anfrage-Mail ohne Zeilenumbrüche: Zwei Annahme-Mails kamen als ein einziger Absatz an.
      `buildMailtoUrl` in `website/src/lib/anfrage-mailto.ts` trennt die Zeilen mit `\n`, kodiert als
      `%0A`. RFC 6068 verlangt `%0D%0A` für Zeilenumbrüche im Body. Für die URL mit `\r\n` verbinden;
      die Anzeige zum Kopieren behält `\n`.
- [ ] Fassungsdatum der Nutzungsbedingungen von Hand an drei Stellen: `TERMS.md` nennt es im Stand
      und in der E-Mail-Vorlage, `buildAnfrageMail` im Annahmesatz. Der Test pinnt nur das Literal,
      eine neue Fassung ohne Formular-Update bleibt grün. Ein Verein nannte das Datum einer älteren
      Fassung. Ein Test liest den Stand aus `TERMS.md` und vergleicht ihn mit dem Annahmesatz.
- [ ] Rechtsform-Auswahl im Anfrage-Formular: `artOptionen` bietet e.V., Stiftung, „NGO / NPO" und
      „Sonstige". Vereine aus Österreich und der Schweiz wählen deshalb „Eingetragener Verein (e.V.)".
      gGmbH und gUG aus TERMS § 2 (1) fehlen. Optionen an § 2 (1) angleichen und einen Verein nach
      österreichischem oder Schweizer Recht anbieten.
- [ ] Lasttest fehlt: Die FAQ legt jotti auf Teams bis 30 Helfer aus, der Self-Hosting-Leitfaden
      nennt einen kleinen VPS. Keine Messung stützt beides, das Repo hat keinen Lasttest. Ein Verein
      fragt nach der Grenze unter hoher Last. Einmal messen: 30 Servicekräfte buchen und kassieren
      parallel gegen den Prod-Stack mit fiskaly-Testumgebung. Das Ergebnis als Grenze in FAQ und
      Leitfaden eintragen.
- [ ] WLAN mit mehreren Access Points: Ein Verein hatte drei Router unter derselben SSID. Jedes Handy
      musste nach einem Wechsel die Adresse neu eingeben; mit einem Router lief es. Das WLAN hatte kein Internet,
      mobile Daten waren aus. Ursache klären und in die Fehlersuche des Leitfadens aufnehmen.
