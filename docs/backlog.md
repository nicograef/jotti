# Backlog

Offene Produktentscheidungen und Folgekandidaten ohne Release-Bezug.

## Produktentscheidungen

- [ ] Kassenabschluss-Wiederanlauf nach einer Geldbewegung: Steht ein Kassensturz und bewegt sich
      danach Geld (Tischzahlung, Direktverkauf, Storno mit Rückgabe, Geldtransit), bricht jeder
      Wiederanlauf mit `ErrBuchungenNachKassensturz` ab; kein Befehl hebt einen Kassensturz auf,
      die Kassensitzung bleibt offen, das Frontend sagt „Bitte den Administrator kontaktieren".
      Reparaturweg (z. B. zweiter Kassensturz beim Wiederanlauf) bauen oder Admin-Ablauf
      dokumentieren.
- [ ] `GetEigeneUebersicht` im Barrierestatus: liest weiter `GetOffeneKassensitzungNr`; die eigene
      Übersicht der Servicekraft zeigt Nullen, solange die Kassensitzung `wird_abgeschlossen` ist.
      Auf `GetAktiveKassensitzung` umstellen oder die Nullen hinnehmen.
- [ ] Zeichen gegen Bytes an Kommentar- und Tischnamen-Feldern: Zod zählt Code-Units, die
      zog-Schemas im Backend Bytes — ein Umlaut-Kommentar mit 100 Zeichen passiert das
      Frontend und bekommt 400. Betreiber-Felder zählen Zeichen. Frontend zählt Bytes
      (`TextEncoder`) oder die UI nennt die Einheit.
- [ ] Login-Formular: `frontend/src/lib/AuthBackend.ts` nutzt für den Login das volle
      `PasswordSchema` (min 6/max 72) und verrät damit die Passwort-Policy; der Endpunkt tut es
      nicht mehr. Eigenes Login-Schema (trim + min 1).
- [ ] Windows: `jotti-restore.cmd` liest nur `jotti-*.sql` aus dem Volume; für `manuell-*.sql` auf
      dem Host gibt es keinen dokumentierten Weg. Restore-Weg dokumentieren oder das Skript
      erweitern.
- [ ] `JOTTI_DOMAIN` bei jedem Compose-Befehl des Public-Stacks: Folge der Entscheidung Option A —
      jeder Compose-Befehl mit `docker-compose.prod.yml` (auch `make prod-down`, `make prod-logs`)
      verlangt die Variable. Option A (hinnehmen) laut statt still bestätigen.
- [ ] Laufende Bewirtung im Vereinsheim: Ein Verein setzt jotti auch dafür ein. Das ist nicht die
      Zielgruppe (2–3 Feste pro Jahr), wird aber nicht verhindert. Ob das ein Nicht-Ziel wird, ist
      offen.
- [ ] Kontingent-Funktion für den Bondruck: Ein Verein mit Lehrgängen und einem Turnier wünscht sie
      an der Kasse (Teilnehmer erhalten ein festes Kontingent). Anforderung noch unklar, Rückfrage
      läuft. Warenwirtschaft ist Nicht-Ziel; ein Tisch pro Teilnehmer oder Abholbons könnten
      reichen.

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
- [ ] Status-Seite: eine alte `install.json` mit einer Subdomain, die den Subdomain-Regex verletzt,
      lässt den Proxy nur mit der internen CA weiterlaufen; die Status-Seite rät dann zu einem
      Neustart, der nichts ändert (acme-dns vergibt Kleinbuchstaben-UUIDs, daher theoretisch).
