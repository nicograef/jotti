# PRD: Windows-Einzeldatei

## Problem Statement

Ein Verein, der jotti auf dem Vereinsrechner betreibt, braucht dafür heute Docker Desktop.
Docker Desktop und Netzwerk-Know-how sind die größte Hürde der Vereine.

- Docker Desktop ist eine eigene Installation mit eigenen Updates und eigenen Fehlermeldungen.
- Es braucht WSL 2 und eingeschaltete Virtualisierung, auf manchen Rechnern erst im BIOS.
- Der Erststart lädt mehrere Container-Images und dauert einige Minuten.
- Die WSL-2-VM belegt dauerhaft Arbeitsspeicher, auf alten Vereinsrechnern spürbar (Annahme, nicht gemessen).
- Meldet Docker einen Fehler, kann ein Helfer ihn weder lesen noch beheben.
- Das ZIP enthält Starter, Relay und drei Hilfsskripte; der Verein muss wissen, welche Datei was tut.
- Ein Verein fragte ausdrücklich nach einer Installation ohne Docker.

Entschieden: jotti bettet Postgres ein, nicht SQLite. Das eingefrorene Schema bleibt so unverändert.

## Solution

Der Verein lädt eine einzige Datei `jotti.exe` herunter und doppelklickt sie.
Mehr braucht der Kassenrechner nicht: kein Docker Desktop, kein ZIP, keine Hilfsskripte.
Der Abschnitt „Rückfall B" hält die verworfene SQLite-Variante zum Vergleich fest.

- SmartScreen und UAC bestätigt der Verein wie heute einmal.
- Der erste Doppelklick richtet jotti ein, gibt die Firewall frei und startet die Kasse.
- Die Startkonsole zeigt den Anmelde-Code für den Admin und die Status-Seite.
- Status-Seite, grüne Adresse, Fallback-Adresse und QR-Code bleiben, wie sie sind.
- Das Fenster darf danach zu; jotti läuft weiter, bis jemand es beendet.
- Ein Doppelklick bei laufendem jotti zeigt den Status und bietet „Beenden", „Backup zurückspielen" und „Entfernen".
- Der Bondruck ist eingebaut; ein zweites Programm für das Relay entfällt.
- Für ein Update lädt der Verein die nächste `jotti.exe` und doppelklickt sie; jotti sichert vorher die Daten.
- Läuft auf dem Rechner schon eine Docker-Installation, bietet der erste Start die Übernahme an.
- Die Übernahme bringt alle Daten, Zugangsdaten und dieselbe grüne Adresse mit.
- Der Betrieb auf einem Linux-Server bleibt unverändert bei Docker Compose.

## User Stories

1. Als Kassenwart möchte ich jotti ohne Docker Desktop installieren, damit die Einrichtung ohne IT-Kenntnisse gelingt.
2. Als Kassenwart möchte ich genau eine Datei herunterladen, damit ich nichts entpacken und nichts zuordnen muss.
3. Als Kassenwart möchte ich nach dem Doppelklick in unter einer Minute kassierbereit sein, damit der Probelauf nicht stockt.
4. Als Kassenwart möchte ich die grüne Adresse wie heute nutzen, damit die Handys ohne Browserwarnung arbeiten.
5. Als Kassenwart möchte ich das Konsolenfenster schließen können, ohne dass die Kasse stoppt, damit ein Fehlklick nichts anrichtet.
6. Als Kassenwart möchte ich jotti mit einem Doppelklick beenden, damit der Vereinsrechner nach dem Fest wieder frei ist.
7. Als Kassenwart möchte ich mit einer neueren Datei aktualisieren, damit vorher automatisch ein Backup entsteht.
8. Als Kassenwart möchte ich ein misslungenes Update zurückrollen, damit das Fest mit der Vorversion weiterläuft.
9. Als Kassenwart möchte ich meine Docker-Installation übernehmen, damit Kassenjournal, Seriennummer und TSE-Anbindung erhalten bleiben.
10. Als Kassenwart möchte ich nach der Übernahme dieselbe grüne Adresse behalten, damit gespeicherte Lesezeichen und QR-Codes weiter gelten.
11. Als Kassenwart möchte ich Bons ohne zweites Programm drucken, damit am Festtag nichts vergessen wird.
12. Als Kassenwart möchte ich jotti entfernen können, ohne Kassendaten zu verlieren, damit die Aufbewahrungspflicht erfüllt bleibt.
13. Als Admin eines Linux-Servers möchte ich meinen Compose-Betrieb unverändert behalten, damit kein Umzug nötig wird.
14. Als Maintainer möchte ich nur eine Schema- und Abfragelinie pflegen, damit jede Änderung genau einmal gebaut und getestet wird.

## Implementation Decisions

### Abgewogene Ansätze

| Ansatz                                                         | Aufwand (Schätzung)                                                                  | Risiko                                                                                                     | Umkehrbarkeit                                                                                          |
| -------------------------------------------------------------- | ------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------ |
| A: Postgres 17 als eingebetteter Teilprozess der Einzeldatei   | 6–9 Wochen eines Entwicklers, einmalig; Schema, sqlc, Abfragen und Tests unverändert | mittel und betrieblich: Prozessaufsicht, Dienstkonto, Virenscanner, Dateigröße                             | voll: ein Dump wandert jederzeit zurück in eine Docker- oder Server-Installation                       |
| B: SQLite als zweite Schema- und Abfragelinie nur für Windows  | 4–6 Monate einmalig, danach doppelte Arbeit an jeder Migration und jeder Abfrage     | hoch und fiskalisch: zwei SQL-Umsetzungen von Reporting und DSFinV-K-Export können still auseinanderlaufen | schlecht: die SQLite-Linie friert mit dem ersten Echtbetrieb ein und braucht für immer einen Konverter |
| C: SQLite für alle Installationen, Postgres-Bestände ziehen um | wie B plus Umzug jedes Servers                                                       | sehr hoch: verändert das Datenformat aller Bestandsinstallationen unter Aufbewahrungspflicht               | keine: der Rückweg wäre ein zweiter Umzug                                                              |

Gewählt ist A. Die Owner-Entscheidung nennt SQLite; dieses PRD empfiehlt davon abzuweichen.

- Das Nutzerziel ist eine Datei ohne Docker Desktop; das erreicht A vollständig.
- Für den Verein ist die Datenbank unsichtbar; SQLite bringt ihm keinen Vorteil, den er bemerkt.
- A lässt das eingefrorene Schema, die Schutz-Trigger, REVOKE, LATERAL, JSONB und die Zeilensperren unangetastet.
- Die Übernahme einer Docker-Installation ist bei A ein Dump und Restore innerhalb derselben Postgres-Version.
- B verdoppelt den kassenwirksamen Teil des Codes und damit den Prüfaufwand jeder fiskalischen Änderung.
- B gewinnt eine kleinere Datei und keinen zweiten Prozess; das wiegt die Doppelpflege nicht auf.
- C scheidet aus, weil er persistierte Daten aller Bestandsinstallationen umformt.
- Die verworfene SQLite-Variante beschreibt der Abschnitt „Rückfall B" unten.

### Aufbau der Einzeldatei

Die Einzeldatei ist ein Go-Programm für Windows, das alle Teile von heute in sich trägt.

| Teil                   | Herkunft                                                     | Laufform in der Einzeldatei                                              |
| ---------------------- | ------------------------------------------------------------ | ------------------------------------------------------------------------ |
| Aufsicht und Bedienung | Windows-Starter mit seiner reinen Logik                      | Konsole beim Doppelklick, Windows-Dienst im Betrieb                      |
| Postgres 17            | offizielle Windows-Binaries von Server, initdb und pg_dump   | Teilprozess des Dienstes, lauscht nur auf Loopback                       |
| Migrationen            | dieselben Migrationsdateien wie im Migrate-Image             | golang-migrate als Bibliothek, vor dem Backend-Start                     |
| Backend                | unverändertes Backend-Modul                                  | im Prozess, lauscht nur auf Loopback                                     |
| TSE-Signatur-Worker    | Teil des Backends                                            | unverändert, der Advisory Lock funktioniert wie im Container             |
| Caddy                  | dieselben Plugins wie das Proxy-Image (acme-dns, Rate-Limit) | als Go-Bibliothek im Prozess, Konfiguration aus derselben Erzeugung      |
| App-Build              | derselbe statische Build wie im Proxy-Image                  | beim Einrichten in den Programmordner entpackt, Caddy liefert ihn aus    |
| Status-Seite           | LAN-Modus des Reverse-Proxy-Programms                        | unverändert auf Loopback-Port 8484                                       |
| Print-Relay            | Relay-Programm                                               | im Prozess, Token automatisch, Transport weiter über die Relay-Endpunkte |

- Das Backend bietet seinen Start als aufrufbare Funktion an; das Container-Binary ruft dieselbe Funktion.
- Die Einzeldatei setzt die Konfiguration im Prozess statt über Umgebungsvariablen aus Compose.
- Der LAN-Modus des Reverse-Proxy-Programms wird importierbar; Container und Einzeldatei erzeugen dieselbe Caddy-Konfiguration.
- Einziger Unterschied der Konfiguration ist das Upstream-Ziel: Loopback statt Container-Name.
- Das Relay bleibt ein reiner Transport und spricht das Backend über dieselben Relay-Endpunkte an.
- D18 bleibt wahr: Caddy terminiert TLS, leitet die API weiter und liefert die App aus.
- Programmordner und Datenverzeichnis sind getrennt; ein Update ersetzt nur den Programmordner.
- Das Datenverzeichnis ist dasselbe, das heute Zugangsdaten und Backups hält.

### Datenbank

- Die Einzeldatei trägt die Postgres-Binaries komprimiert und entpackt sie je Version in den Programmordner.
- Der erste Start legt den Cluster mit initdb im Datenverzeichnis an, mit UTF-8 und Passwort-Authentifizierung.
- Postgres lauscht nur auf Loopback, auf einem festen Port abseits von 5432; die Port-Prüfung meldet Konflikte.
- Datenbankname, Benutzer und generiertes Passwort folgen dem Compose-Betrieb.
- Der Dienst läuft unter einem virtuellen Dienstkonto ohne Adminrechte; nur es schreibt ins Datenverzeichnis.
- Grund ist das Prinzip der geringsten Rechte: Ein Fehler im Dienst erreicht nur das Datenverzeichnis, nicht den Rechner.
- Postgres verweigert unter Windows ohnehin den Start mit Administratorrechten.
- Migrationen laufen beim Start mit golang-migrate als Bibliothek, aus denselben eingebetteten Dateien und derselben Versionstabelle.
- Schema, Freeze-Regeln, Schutz-Trigger, REVOKE, sqlc, Abfragen und Sperren bleiben genau, wie sie sind.
- Append-only erzwingen dieselben Trigger wie auf jedem Server; es gibt keine zweite Schutzlogik.
- Die Postgres-Version ist dieselbe 17.x wie in allen Compose-Stacks; die Pin-Prüfung liest auch sie (D15).

### TLS und grüne Adresse

- Der Mechanismus bleibt: einmalige Registrierung bei acme-dns, Install-State, Wildcard-Zertifikat per DNS-01.
- Der Hostname bildet sich wie heute aus LAN-IP und Install-ID; der zentrale Resolver ändert sich nicht.
- Die LAN-IP liest die Einzeldatei direkt aus der Default-Route, weil keine Docker-Bridge dazwischen liegt.
- Der Fallback auf die LAN-IP mit interner CA und einmaliger Browserwarnung bleibt.
- Install-State und Zertifikatsspeicher liegen im Datenverzeichnis und überleben jedes Update.
- Die Ports 80 und 443 prüft die Einzeldatei vor dem Start wie der Starter heute.
- Die Firewall-Regel gilt für das Programm im Programmordner, nicht für Docker.

### Dienst und Bedienung

Aus dem Starter wird die Einzeldatei selbst; ihr Manifest verlangt wie heute Administratorrechte.

- Erster Doppelklick: kopiert sich in den Programmordner, entpackt die Teile, registriert den Dienst „jotti".
- Danach setzt sie die Firewall-Regel, startet den Dienst, wartet auf Gesundheit und zeigt Anmelde-Code und Status-Seite.
- Den Anmelde-Code liest die Konsole aus dem Dienst-Log seit dem aktuellen Dienststart, wie heute aus dem Container-Log.
- Der Dienst startet manuell, nicht beim Hochfahren; ein Vereinsrechner dient im Alltag anderen Zwecken.
- Der Dienst überwacht den Postgres-Teilprozess und startet ihn bei einem Absturz neu.
- Beim Beenden hält er zuerst Relay und Backend an und fährt danach Postgres sauber herunter.
- Logs schreibt der Dienst rotierend ins Datenverzeichnis.
- Doppelklick bei laufendem Dienst: Status, dann „Beenden", „Backup zurückspielen", „Entfernen" oder Fenster schließen.
- Es gibt kein Tray-Symbol und keinen grafischen Installer; Konsole und Status-Seite reichen.
- Die reine Starter-Logik bleibt: Versionsvergleich, Downgrade-Sperre, Backup-Rotation, Port-Prüfung, LAN-IP, Geheimnisse, Admin-Marker.
- Docker-Prüfung, Compose-Aufrufe und Volume-Zugriffe entfallen bis auf das Lesen bei der Übernahme.

### Update, Backup und Wiederherstellen

- Eine neuere Einzeldatei erkennt den Versionswechsel und fragt vor dem Update einmal nach.
- Vor jeder Migration zieht sie mit pg_dump ein Backup ins Datenverzeichnis und behält die letzten fünf.
- Ein fehlgeschlagenes Backup bricht das Update ab, wie heute.
- Der Programmordner behält die aktuelle und die vorige Version.
- „Backup zurückspielen" legt einen leeren Cluster an, spielt das jüngste Backup ein und startet die Version, die es schrieb.
- Die Downgrade-Sperre folgt dem zurückgespielten Stand.
- Der Update-Hinweis nach dem Start bleibt ein reiner Hinweis; eine laufende Kasse aktualisiert sich nie selbst.
- „Entfernen" löscht Dienst, Firewall-Regel und Programmordner, nie das Datenverzeichnis, und sagt das ausdrücklich.
- Das Archiv für die zehn Jahre nach § 147 AO bleibt der DSFinV-K-Export je Kassensitzung.
- Die Datenbank-Backups bleiben Sicherheitsnetz; außer Haus kopiert sie der Verein wie heute selbst.

### Übernahme einer Docker-Installation

- Der erste Start erkennt ein vorhandenes Datenvolume der lokalen Docker-Installation und bietet die Übernahme an.
- Die Übernahme verweigert den Start, solange eine Kassensitzung offen ist; ein Fest zieht nicht mitten im Betrieb um.
- Sie startet den Postgres-Container, zieht einen Dump und hält den Docker-Stack an.
- Der Dump geht in den eingebetteten Cluster derselben Postgres-Version, also verlustfrei samt Versionstabelle.
- Install-State, Zertifikatsspeicher und Geheimnisse kommen mit; die grüne Adresse bleibt damit dieselbe.
- Danach vergleicht sie Schema-Version, Zeilenzahl je Tabelle, höchste Journal-ID und Kassen-Seriennummer.
- Weicht ein Wert ab, bricht sie ab, verwirft den eingebetteten Cluster und lässt Docker unangetastet.
- Die Docker-Volumes bleiben in jedem Fall stehen; der Verein löscht Docker Desktop erst, wenn alles läuft.
- Die Seriennummer bleibt; eine Abmeldung beim Finanzamt und eine erneute fiskaly-Anbindung entfallen.
- Offene TSE-Signaturaufträge und Druckaufträge wandern mit ihren Tabellen und laufen danach weiter.

### Linux und Self-Hosting

- Der Server-Betrieb bleibt Docker Compose mit denselben Images, Skripten und Postgres-Containern.
- Das Docker-ZIP für Windows läuft eine Übergangszeit parallel und entfällt danach.
- Mit ihm entfallen Starter, die Release-Compose-Datei und die Windows-Hilfsskripte.
- `docker-compose.local.yml` bleibt für den LAN-Betrieb unter Linux mit `make local-up`.
- Das eigenständige Relay bleibt für Server und für Drucker in einem anderen Netz.
- Die Einzeldatei bekommt einen eigenen Prüf-Job oder tritt der Windows-Matrix bei, wie D13 es verlangt.

### Build und Release

- Der Release-Workflow baut die Einzeldatei per Cross-Compile, wie heute Starter und Relay.
- Er lädt die Postgres-Binaries mit festgeschriebener Prüfsumme, wie das Migrate-Image golang-migrate.
- Die Datei wird geschätzt 60–100 MB groß; Stufe 1 misst das.
- Eine unsignierte Datei, die einen Dienst anlegt und die Firewall öffnet, fällt Virenscannern eher auf.
- Ob jotti eine Code-Signatur kauft, ist eine offene Entscheidung.

### Stufen

| Stufe | Inhalt                                                                                     | Ergebnis                                                    | Schätzung  |
| ----- | ------------------------------------------------------------------------------------------ | ----------------------------------------------------------- | ---------- |
| 1     | Durchstich: Einzeldatei entpackt Postgres, migriert, startet Backend und App auf localhost | Messwerte zu Größe, Startzeit und Defender; Go oder Abbruch | 1 Woche    |
| 2     | Caddy im Prozess, grüne Adresse, Fallback, Status-Seite, Firewall                          | Erstinstallation im LAN nutzbar, als Beta                   | 1–2 Wochen |
| 3     | Dienst, Menü, Update, Backup, Wiederherstellen, Entfernen, Relay im Prozess                | Gleichstand mit dem Docker-Weg für Erstinstallationen       | 2–3 Wochen |
| 4     | Übernahme einer Docker-Installation mit Abgleich                                           | Bestandsvereine können umziehen                             | 1–2 Wochen |
| 5     | Leitfaden, Kurzanleitung und Verfahrensdokumentation; Docker-ZIP läuft aus                 | ein Windows-Weg                                             | 1 Woche    |

Jede Stufe endet lieferbar; nach Stufe 1 entscheidet der Owner über den Weiterbau.

### Rückfall B: SQLite-Linie

Dieser Abschnitt gilt nur, wenn der Owner bei SQLite bleibt.

- Eine zweite Migrationslinie mit eigener Nummerierung übersetzt den Gesamtstand der Postgres-Migrationen in SQLite.
- Mit dem ersten Echtbetrieb friert diese Linie ebenso ein wie die Postgres-Linie.
- Jede künftige Postgres-Migration bekommt im selben Change ihren SQLite-Zwilling.
- Eine CI-Prüfung vergleicht beide Schemata nach Tabellen, Spalten, Constraints und Indexen.
- IDENTITY wird INTEGER PRIMARY KEY mit AUTOINCREMENT; die Z-Nummer steigt so streng und wird nie wiederverwendet.
- ENUM-Typen werden TEXT mit CHECK; JSONB wird TEXT mit einer JSON-Gültigkeitsprüfung.
- Zeitstempel speichert die Linie als UTC-Text mit fester Länge, damit Sortierung und Vergleich stimmen.
- Append-only erzwingen BEFORE-UPDATE- und BEFORE-DELETE-Trigger mit Abbruch auf dem Kassenjournal.
- Die Kassenidentität sperren dieselben Trigger nach ihrer ersten Zeile, auch für INSERT; TRUNCATE gibt es in SQLite nicht.
- SQLite kennt kein REVOKE; der Schutz hält gegen Anwendungs- und SQL-Fehler, nicht gegen Schreibzugriff auf die Datei.
- Praktisch gleicht das dem Tabellen-Owner unter Postgres, der Trigger ebenfalls entfernen kann.
- Ein zweiter Abfragesatz für die SQLite-Engine von sqlc ersetzt die 122 Postgres-Abfragen.
- Zehn LATERAL-Joins werden korrelierte Unterabfragen, JSONB-Funktionen werden die JSON-Funktionen von SQLite.
- Die SQL-Hilfsfunktionen des Reportings werden im Treiber registrierte Go-Funktionen oder ausgeschriebene Ausdrücke.
- sqlc erzeugt zwei Pakete mit eigenen Typen; jedes Repository bekommt eine Schnittstelle mit zwei Umsetzungen.
- Alle schreibenden Transaktionen beginnen sofort exklusiv, also mit BEGIN IMMEDIATE und einem Busy-Timeout.
- Das ersetzt die Abschluss-Barriere aus FOR SHARE gegen UPDATE und die Passwort-Sperre aus FOR UPDATE.
- Den Advisory Lock des TSE-Workers ersetzt eine Sperre im Prozess plus eine Einzelinstanz-Sperre auf die Datei.
- Die Fehlerabbildung lernt die Constraint-Codes von SQLite neben denen von Postgres.
- Der Treiber ist reines Go ohne cgo, damit der Cross-Compile bleibt.
- JSONB normalisiert Schlüsselreihenfolge und Leerzeichen, SQLite speichert Text wie geschrieben.
- Code, der rohes Event-JSON vergleicht, muss beide Formen vertragen.
- Die Übernahme braucht einen Konverter je Tabelle in ID-Reihenfolge, die Kassenidentität vor den Schutz-Triggern.
- Migrationen laufen mit golang-migrate und dessen SQLite-Treiber ohne cgo, aus der zweiten Migrationslinie.
- Das Backup vor jeder Migration zieht `VACUUM INTO` statt pg_dump; „Backup zurückspielen" ersetzt die Datenbankdatei.
- Es gibt keinen Postgres-Teilprozess; Prozessaufsicht und Postgres-Binaries entfallen.
- Das virtuelle Dienstkonto bleibt, aus demselben Grund der geringsten Rechte.

## Testing Decisions

Getestet wird äußeres Verhalten an öffentlichen Schnittstellen; erwartete Werte sind Literale.

- Schema, Abfragen, Projektionen, Reporting und DSFinV-K: keine, abgedeckt durch die bestehenden Tests, weil die Engine bleibt.
- Caddy-Konfiguration der Einzeldatei: Upstream auf Loopback, sonst identisch zum LAN-Modus. Vorbild ist `caddyfile_test.go`.
- Reine Logik von Aufsicht und Bedienung: Versionswechsel, Downgrade, Backup-Rotation, Ports. Erweitert werden die Tests der Starter-Logik.
- Übernahme-Abgleich als reine Funktion: abweichende Zeilenzahl, Journal-ID oder Seriennummer ergibt Abbruch.
- Übernahme gegen echtes Postgres: Demo-Szenario dumpen, einspielen, Abgleich grün. Vorbild ist der CI-Job `upgrade-path` mit `prod-backup-verify`.
- Rauchtest auf einem Windows-Runner: einrichten, Gesundheit, Login, Buchung, Update mit Backup, Zurückspielen. Vorbild ist `ops-smoke.sh`.
- Der Windows-Runner ist eine eigene CI-Ausgabe; ohne ihn läuft der Rauchtest von Hand je Release.
- Bei Rückfall B: jeder Repository- und Integrationstest gegen beide Engines.
- Bei Rückfall B zusätzlich: derselbe Event-Strom in beide Engines, Reporting und DSFinV-K-Export byteweise gleich.

## Out of Scope

- Gehosteter Betrieb für Vereine; TERMS § 1 (2) und § 3 (2) schließen ihn aus.
- Eine Einzeldatei für Linux oder macOS; Ansatz A macht sie später möglich.
- Automatische Updates einer laufenden Kasse.
- Tray-Symbol, MSI-Installer und Einträge im Startmenü.
- Autostart beim Hochfahren des Rechners.
- Ein Upgrade auf einen höheren Postgres-Major; D15 gilt unverändert.
- Laufende Backups außer Haus aus der Einzeldatei.
- Hardware-TSE; D09 gilt unverändert.
- Eine grüne Adresse ohne Internet; der Resolver bleibt zentral.

## Further Notes

Vorgeschlagene Einträge für `docs/language.md`:

| Begriff     | Bedeutung                                                                                                                                   | Zu vermeiden                                            |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------- |
| Einzeldatei | Die Windows-Programmdatei, die Datenbank, Proxy, Backend, App und Relay enthält und den Kassenrechner ohne Docker einrichtet und betreibt.  | Single Binary, All-in-One, Standalone, Portable-Version |
| Übernahme   | Der einmalige Umzug einer Docker-Installation auf demselben Kassenrechner in die Einzeldatei, mit allen Daten und derselben grünen Adresse. | Migration, Import, Umstieg                              |

„Migration" bleibt der Schemaänderung der Datenbank vorbehalten.

Bei Umsetzung ändern sich Leitfaden Installation, Aktualisieren, Datenaufbewahrung und Fehlersuche sowie die Kurzanleitung.
Ebenso ändern sich Systemüberblick der Verfahrensdokumentation und Betriebsteil des Handbuchs.

Entfällt das Docker-ZIP, wird die Zahl der Compose-Dateien in D13 falsch und braucht eine Umformulierung.
D15 erstreckt sich dann auch auf die eingebettete Postgres-Version.

Offene Entscheidungen für den Owner:

1. Datenbank: entschieden für eingebettetes Postgres.
2. Zusätzliche Go-Abhängigkeiten: Caddy als Bibliothek, golang-migrate als Bibliothek und die Postgres-Windows-Binaries.
3. Code-Signatur der Einzeldatei: kaufen (empfohlen vor Stufe 5) oder unsigniert mit SmartScreen-Hinweis bleiben.
4. Windows-Runner in der CI für den Rauchtest (empfohlen) oder Rauchtest von Hand je Release.
5. Übergangszeit des Docker-ZIPs: empfohlen zwei Releases nach Stufe 4, dann entfernen.
6. Docker-Konfiguration ändern (AGENTS.md: vorher fragen): Wegfall von `docker-compose.release.yml` und des Docker-ZIPs.
7. Reverse-Proxy-Konfiguration ändern (AGENTS.md: vorher fragen): LAN-Modus importierbar, Upstream auf Loopback.
