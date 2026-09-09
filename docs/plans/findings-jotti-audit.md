# Offene Reste aus dem Vollreview (Stand nach den Audit-Fixes)

Die 97 bestätigten Befunde des Vollreviews sind über `plan-jotti-audit-fixes.md` umgesetzt
(Git-Historie). Hier stehen nur die Punkte, die bewusst offen bleiben: Produktentscheidungen
des Eigentümers und Folgekandidaten ohne Release-Bezug.

## Produktentscheidungen

| Punkt | Stand | Entscheidung |
| --- | --- | --- |
| Kassenabschluss-Wiederanlauf nach Geldbewegung | Steht ein Kassensturz und bewegt sich danach Geld (Tischzahlung, Direktverkauf, Storno mit Rückgabe, Geldtransit), bricht jeder Wiederanlauf mit `ErrBuchungenNachKassensturz` ab; kein Befehl hebt einen Kassensturz auf, die Kassensitzung bleibt offen. Frontend: „Bitte den Administrator kontaktieren". | Reparaturweg (z. B. zweiter Kassensturz beim Wiederanlauf) oder dokumentierter Admin-Ablauf |
| `GetEigeneUebersicht` im Barrierestatus | Liest weiter `GetOffeneKassensitzungNr`; die eigene Übersicht der Servicekraft zeigt Nullen, solange die Kassensitzung `wird_abgeschlossen` ist (Kriterium 7.1 hat die `…Nr`-Variante ausgenommen). | Auf `GetAktiveKassensitzung` umstellen oder so lassen |
| Zeichen vs. Bytes an Kommentar- und Tischnamen-Feldern | Zod zählt Code-Units, die eingefrorenen Event-Schemas Bytes: ein Umlaut-Kommentar mit 100 Zeichen passiert das Frontend und bekommt 400. Betreiber-Felder zählen seit Phase 6 Zeichen. | Frontend zählt Bytes (`TextEncoder`) oder die UI nennt die Einheit |
| Login-Formular verrät die Passwort-Policy | `frontend/src/lib/AuthBackend.ts` nutzt für den Login das volle `PasswordSchema` (min 6/max 72); der Endpunkt tut es seit 8.1 nicht mehr. | Eigenes Login-Schema (trim + min 1) |
| Windows: manuelle Sicherung wiederherstellen | `jotti-restore.cmd` liest nur `jotti-*.sql` aus dem Volume; für `manuell-*.sql` auf dem Host gibt es keinen dokumentierten Weg. | Dokumentieren oder Skript erweitern |
| `JOTTI_DOMAIN` im Public-Stack | Folge der Entscheidung Option A: jeder Compose-Befehl mit `docker-compose.prod.yml` (auch `make prod-down`, `make prod-logs`) verlangt die Variable. | Hinnehmen (laut statt still) |

## Folgekandidaten

- `reverse-proxy/caddyfile.go` steht auf der Ausschlussliste von `scripts/check-prose.sh`; die Zeilen 14, 51, 69 und 92 tragen Historien-Prosa. Umschreiben, dann den Ausschluss streichen.
- `export.go` liest die Kassensitzung mit `GetOffeneKassensitzung` und `//nolint:forbidigo`; wegen der Sortierung von `GetAllKassensitzungen` lieferte `GetAktiveKassensitzung` dieselbe Sitzung — die Ausnahme kauft kein Verhalten.
- `scripts/check-pins.sh` prüft Compose, Dockerfiles und `packageManager`; die Postgres-Pins in `scripts/test-integration.sh`, `scripts/test-tse-live.sh` und den CI-Services liegen außerhalb.
- `scripts/check-ui-labels.sh` prüft das Vorkommen zitierter Bedienelemente in `frontend/src`; ein Zitat, das nur als Bezeichner oder Kommentar vorkommt, läuft durch. Schärfer ginge es nur gegen gerenderte Texte.
- `dsfinvkpruefung` prüft Typen und Dezimalformat, nicht `MaxLength`; die Feldlängen sichern allein die Mapper-Tests.
- `backend/repository/kassenjournal_repo/mock.go` kann keine Summe der Differenzbuchungen ≠ 0 abbilden; den unterscheidenden Fall deckt nur der Integrationstest.
- `tisch_sessions.unbezahlte_positionen` persistiert `[]kasse.Position` (ohne JSON-Tags, PascalCase-Schlüssel) als Projektion; außerhalb des Event-Vertragsgates, ein Umbenennen der Felder wäre ein Bruch persistierter Daten.
- `api/auth/http/command_handler.go` ruft `user.PasswordSchema.Required()` auf dem geteilten Paket-Schema auf (zog mutiert in place) — latente Falle, heute ohne Auswirkung.
- Alt-`install.json` mit einer Subdomain, die den Regex aus 11.3 verletzt: der Proxy läuft nur mit der internen CA weiter und die Status-Seite rät zu einem Neustart, der nichts ändert (acme-dns vergibt Kleinbuchstaben-UUIDs, daher theoretisch).
