# Backlog

Offene Produktentscheidungen und Folgekandidaten ohne Release-Bezug.

## Produktentscheidungen

- [ ] Kassenabschluss-Wiederanlauf nach einer Geldbewegung: Steht ein Kassensturz und bewegt sich
      danach Geld, bricht jeder Wiederanlauf mit `ErrBuchungenNachKassensturz` ab. Geldbewegungen
      sind Tischzahlung, Direktverkauf, Storno mit Rückgabe, Pfandrückgabe samt Storno und
      Geldtransit. Kein Befehl hebt einen
      Kassensturz auf. Die Kassensitzung bleibt offen, das Frontend sagt „Bitte den Administrator kontaktieren".
      Reparaturweg (z. B. zweiter Kassensturz beim Wiederanlauf) bauen oder Admin-Ablauf
      dokumentieren.
- [ ] `GetEigeneUebersicht` im Barrierestatus: liest weiter `GetOffeneKassensitzungNr`; die eigene
      Übersicht der Servicekraft zeigt Nullen, solange die Kassensitzung `wird_abgeschlossen` ist.
      Auf `GetAktiveKassensitzung` umstellen oder die Nullen hinnehmen.
- [ ] Login-Formular: `frontend/src/lib/AuthBackend.ts` prüft den Login mit dem vollen
      `PasswordSchema` (min 6/max 72). Das verrät die Passwort-Policy; der Login-Endpunkt prüft
      nur trim + min 1. Eigenes Login-Schema (trim + min 1).
- [ ] Windows: `jotti-restore.cmd` liest nur `jotti-*.sql` aus dem Volume; für `manuell-*.sql` auf
      dem Host gibt es keinen dokumentierten Weg. Restore-Weg dokumentieren oder das Skript
      erweitern.
- [ ] `JOTTI_DOMAIN` bei jedem Compose-Befehl des Public-Stacks: Jeder Compose-Befehl
      mit `docker-compose.prod.yml` (auch `make prod-down`, `make prod-logs`) verlangt die
      Variable. Die Pflicht hinnehmen und das ausdrücklich festhalten.
- [ ] Rabatt und Gutschrift: Ein Verein fragt danach. Eine Gutschrift für ein reklamiertes Getränk deckt
      die Stornierung ab. Ein Rabatt bräuchte eine negative Positionszeile vom Typ „Rabatt" je Steuersatz
      (DSFinV-K Tz. 4.2.4) und Folgeversionen der Zahlungs-Events. Offen: nur bei weiterer Nachfrage.
- [ ] Offene Tische automatisch unter „Meine Tische": Ein Verein wünscht sich Tische mit eigenen
      offenen Bestellungen ohne Stern auf dem Dashboard.
- [ ] Letzter Admin ausgesperrt: Ein Verein hat das Passwort seines einzigen Admins vergessen. Andere
      Konten setzt ein Admin per Einmalpasswort zurück; für den letzten Admin gibt es keinen Weg.
      Wiederherstellung über die Startkonsole bauen oder einen Weg dokumentieren.

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
- [ ] Umlaute auf Epson-Bondruckern (#144): jotti druckt in PC437 (`ESC t 0`), der Testbon zeigt
      eine Umlautzeile. Offen: Druckprobe auf Epson TM-T20IV und Sam4s Hcube, danach #144 schließen.
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
- [ ] Lasttest fehlt: Die FAQ legt jotti auf Teams bis 30 Helfer aus, der Self-Hosting-Leitfaden
      nennt einen kleinen VPS. Keine Messung stützt beides, das Repo hat keinen Lasttest. Ein Verein
      fragt nach der Grenze unter hoher Last. Einmal messen: 30 Servicekräfte buchen und kassieren
      parallel gegen den Prod-Stack mit fiskaly-Testumgebung. Das Ergebnis als Grenze in FAQ und
      Leitfaden eintragen.
- [ ] WLAN mit mehreren Access Points: Ein Verein hatte drei Router unter derselben SSID. Jedes Handy
      musste nach einem Wechsel die Adresse neu eingeben; mit einem Router lief es. Das WLAN hatte kein Internet,
      mobile Daten waren aus. Ursache klären und in die Fehlersuche des Leitfadens aufnehmen.
- [ ] DSFinV-K 3.0: Der Export folgt der DSFinV-K 2.4. Version 3.0 gilt für Aufzeichnungen nach dem
      31.12.2028 und ist nach dem 31.12.2027 wahlweise anwendbar (BMF 30.10.2026). 2.4 gegen 3.0
      abgleichen, Einstieg ist der Änderungsnachweis. Danach ein PRD für die Migration schreiben.
