# Guide: Manuelle QA vor v1.0.0

Handlungsleitfaden für Nico. Konsolidiert Audit-Abschnitt F und die manuellen Teile der Gates 2, 3 und 6 aus dem [Release-Guide](plan-v1.0.0-release.md) an einer Stelle; Checkboxen werden nur hier geführt. Die Blöcke sind so geordnet, dass ein Durchlauf auf einem Test-Server alles abdeckt. Zeitlich unkritisch bis kurz vor dem v1.0.0-Tag; die Erstinstallation des ersten Vereins hängt nicht daran.

## Vorbereitung

- [ ] fiskaly-TEST-Konto angelegt (Zugangsdaten griffbereit)
- [ ] Frischer Test-Server oder VM (nicht der Build-Host, nicht der Dev-Rechner)
- [ ] 80-mm-Bondrucker (ESC/POS) angeschlossen
- [ ] Zwei Endgeräte (Handys/Tablets) für den Parallelzugriffstest
- [ ] Ein echter Windows-Rechner für den Starter-Smoke-Test
- [ ] Referenzmaterial: `docs/rechtsquellen/technik-spezifikationen/DSFinV-K-2.4/` (inkl. `02_index.xml/` und `gdpdu-01-09-2004.dtd`), `docs/compliance.md`

## Block 1: Frische Installation (Gate 3)

- [ ] `make prod-init` auf leerem Server bis zum ersten Admin-Login. Erwartet: Skript prüft Secrets (A2-Fix), gibt am Ende den Einmalpasswort-Code aus (C22-Fix), Login klappt.
- [ ] Windows-Pfad: `make release-windows VERSION=…` baut `jotti-start.exe` + `jotti-relay.exe` + Release-ZIP unter `dist/`; ZIP auf dem Windows-Rechner entpacken und bis zum ersten Login smoke-testen (Gate 3/6).

## Block 2: TSE-Inbetriebnahme (Gate 2.1)

- [ ] Setup-Wizard im Admin-Bereich real durchlaufen: TSS und Client von jotti anlegen lassen, nicht im fiskaly-Dashboard.
- [ ] TEST→LIVE-Umschaltung im Wizard geprüft; PUK/PIN-Verwahrung dokumentiert (Betreiber-Leitfaden).
- [ ] TSE-Stammdaten wurden beim Setup persistiert (Signaturalgorithmus, Public Key, Zertifikat, Seriennummer) — prüft den Phase-3-Fix (C6); Voraussetzung für `tse.csv` im Export.

## Block 3: Signaturbetrieb im Normalfall (Gate 2.2, C7)

- [ ] Je einen realen Fall erzeugen und die Signatur prüfen: Bestellung, Teil-/Vollzahlung, Direktverkauf, Direktverkauf-Storno, Storno bezahlter Positionen (Warenrücknahme), geldneutrale Korrektur, Umbuchung, Geldtransit, Kassendifferenz, Tagesabschluss.
- [ ] Jeder Signaturauftrag wird vom Worker abgeschlossen; p95-Latenz unter Burst messen (Zusage p95 < 5 s). Fällt die Messung durch: Retrieve-first zu Start-first optimieren oder die Zusage in der Verfahrensdoku anpassen (C7).
- [ ] processType-Mapping stichprobenhaft gegen compliance §3.3 (Bestellung-V1 / Kassenbeleg-V1 / SonstigerVorgang).

## Block 4: Ausfall und Nachsignierung (Gate 2.3)

- [ ] TSE künstlich stören (falscher Key / Netzwerk blockiert): Vorgänge bleiben buchbar, Buchen wartet nicht auf die TSE.
- [ ] Störungsprotokoll (`tse_stoerungen`) erfasst den Zeitraum mit Grund.
- [ ] Nach Wiederherstellung: automatische Nachsignierung; Beleg trägt „Nachsigniert am …".
- [ ] Vorgang ohne konfigurierte TSE trägt „keine TSE konfiguriert"; im Export `TSE_TA_FEHLER`-Zeile.
- [ ] Kassenabschluss-Verhalten: 409 mit Anzahl/Alter bei frisch ausstehenden Signaturen; erlaubt bei dokumentiertem Ausfall / fehlender Konfiguration, mit Ausweisung in der Abschlussmeldung.

## Block 5: Beleg und Hardware (Gate 2.4; Rest aus A1/A5)

- [ ] Kassenbeleg enthält alle Pflichtangaben (§5.2): Vereinsname/Anschrift, Datum, Positionen mit Steuerkennzeichen, Steuermatrix je Satz mit Prozentsatz und Befreiungshinweis (A1-Fix sichtprüfen), Bon-Nr.
- [ ] TSE-Block gedruckt: Start-/Endzeit, TSE- und Kassen-Seriennummer, Transaktionsnummer, Signaturzähler, Signatur.
- [ ] QR-Code auf echtem 80-mm-Drucker gedruckt und mit dem Handy scannbar — prüft den A5-Fix (dynamische Modulgröße) mit realer fiskaly-Payload (~350–470 Byte).
- [ ] Durchbedienen: Tisch-Beleg druckt den Startzeitpunkt der ersten Bestellung in Klarschrift (§5.3).
- [ ] Kombi 70/30 korrekt ausgewiesen (7-%- und 19-%-Anteil).
- [ ] Reihenfolge: TSE-`FinishTransaction` vor Druckbefehl (§5.6 Variante A) — im Log nachvollziehen.

## Block 6: DSFinV-K-Export (Gate 2.5; prüft A4)

- [ ] Export-ZIP je Kassensitzung wird erzeugt.
- [ ] Struktur gegen die Spezifikation in `docs/rechtsquellen/technik-spezifikationen/DSFinV-K-2.4/` geprüft.
- [ ] Dateinamen englisch und kleingeschrieben (`transactions.csv`, `lines.csv`, `cashregister.csv`, `tse.csv`, …).
- [ ] CSV-Regeln: Semikolon, CRLF, Komma-Dezimaltrennung, Header-Zeile, Spaltenreihenfolge exakt.
- [ ] `index.xml` deklariert nur vorhandene Tabellen; DTD beiliegend.
- [ ] Storno-Abbildung: neue Datensätze mit Negativbetrag, `REF_BON_ID` in `references.csv`, `BON_STORNO = 0` (compliance §6.6).
- [ ] Abrechnungskreis pro Tisch/Kassensitzung in `allocation_groups.csv`.
- [ ] kombi-Aufteilung in `lines_vat.csv` und `transactions_vat.csv` (zwei Steueranteile).
- [ ] `tse.csv` gefüllt (Zertifikats-ID, Algorithmus, TSE-Seriennummer, Public Key).
- [ ] `BEDIENER_NAME` = eingefrorener Benutzername, `BEDIENER_ID` = user_id (§6.4).
- [ ] Tagesabschluss-Zeile (`AVSonstige`) trägt BON_NAME „Tagesabschluss" (A4-Fix).
- [ ] Möglichst mit IDEA oder fiskaly-Prüftooling gegenlesen; mindestens gegen die DSFinV-K-2.4-Beispiele plausibilisieren.
- [ ] Steuersätze (Gate 2.6): Steuerbeträge auf Beleg und im Export stimmen mit `domain/steuer` überein; Rundung geprüft (Cent).

## Block 7: Ops-Roundtrips (Gate 3)

- [ ] `make prod-update` Roundtrip: Pre-Update-Backup → Images ziehen → Migration → Health-Check → Rollback-Anleitung greift.
- [ ] `make prod-backup` + `make prod-backup-verify`: Backup lässt sich in Wegwerf-Postgres einspielen; Backup enthält die Kassen-Seriennummer (§3.7).
- [ ] `make prod-restore` destruktiv einmal geprüft (mit Bestätigung).
- [ ] TLS/Let's Encrypt live grün (lokale LAN-Infra bereits E2E verifiziert, Regressionscheck genügt).
- [ ] Rate Limiting am Login (HTTP 429) und Security Headers (CSP, HSTS, X-Frame-Options, X-Content-Type-Options) geprüft.
- [ ] Parallelzugriffstest: zwei Servicekräfte, gleicher Tisch, ohne Dateninkonsistenz.

## Block 8: Release-Smoke v1.0.0 (Gate 6)

- [ ] Gepinnte 1.0.0-Images auf frischem Server (nicht der Build-Host): `prod-init`, erster Login, ein Verkauf, ein Beleg, ein Export.
