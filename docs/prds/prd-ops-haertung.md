# PRD: Ops-Härtung Runde 1 (Backup-Verlässlichkeit, CI-Lücken, Betriebsrobustheit)

> Quelle: Ops-Analyse vom 2026-07-05 (Codebase, Docker, CI, Hosting,
> Windows-Wrapper) plus die Backup-Test-Findings F1 bis F4 aus
> `docs/plans/findings-phase4-backup-test.md`. Scope-Abwägung mit dem
> Maintainer: umgesetzt wird, was jetzt notwendig ist; bewusst verschoben
> wird alles unter „Out of Scope".

## Problem Statement

jotti soll einfach, robust und stabil auf Linux-VPS und Windows-Rechnern
laufen. Das Fundament steht (Healthchecks, Pre-Update-Backups mit
Downgrade-Schutz, Release-Smoke-Test), aber der Backup-Test gegen die
Live-Demo und eine anschließende Gesamtanalyse haben Defekte und Lücken
gezeigt, die genau die Kernversprechen untergraben:

1. Das Backup-System hat den gefährlichsten Fehlermodus überhaupt, den
   stillen Totalausfall: `prod-backup.sh` stirbt ohne `BACKUP_DIR`/
   `BACKUP_KEEP` in der `.env` lautlos (F2), die mitgelieferten
   systemd-/cron-Vorlagen laufen dadurch wie ausgeliefert ins Leere (F3),
   und niemand würde es merken, weil kein Erfolgssignal existiert.
2. Ein erzeugter Dump wird nie auf Integrität oder Wiederherstellbarkeit
   geprüft; das Restore-Skript ist gegen den rocks-Stack unsicher (F4).
3. Die Windows-Module (`windows/starter`, `windows/relay`) haben keinerlei
   CI: der zuständige Job filtert auf ein Verzeichnis, das nicht mehr
   existiert. Windows-Releases werden ungetestet veröffentlicht.
4. Erstinstallationen laufen in eine Falle: die Vorlagen verweisen auf das
   Image-Tag `latest`, das auf GHCR nie veröffentlicht wird; `make
   prod-init` scheitert dann mit einem kryptischen Pull-Fehler.
5. Auf dem Demo-VPS erneuert certbot Zertifikate, aber nginx lädt sie nie
   nach; ohne Deploy innerhalb von rund 30 Tagen nach der Erneuerung
   serviert die Demo ein abgelaufenes Zertifikat.
6. Kein Compose-Stack begrenzt seine Container-Logs; auf dem VPS musste das
   nachträglich per Host-Konfiguration gelöst werden, jeder Self-Hoster und
   jeder Windows-Kassenrechner hat das Problem weiterhin.
7. Kleinere Robustheitslücken: Backend und Migrate laufen als root, das
   Backend crash-loopt beim Start, wenn Postgres noch nicht bereit ist,
   Frontend/Proxy/Website haben keine Healthchecks, und das Backend kann
   nicht sagen, welche Version es ist (erschwert Support aus der Ferne).

## Solution

Ein Bündel gezielter, überwiegend kleiner Eingriffe, die zusammen das
Backup-System vertrauenswürdig machen, die CI-Lücken schließen und den
Betrieb auf allen drei Wegen (Self-Hosting, Windows-Kasse, Demo-VPS)
härten:

- **Backups, die nachweislich funktionieren:** F2-Fix in allen Skripten,
  Integritätsprüfung nach jedem Dump, ein Verify-Skript, das den neuesten
  Dump probeweise in eine Wegwerf-Datenbank einspielt, ein opt-in
  Erfolgs-Ping (Dead-Man-Switch) und korrigierte Timer-Vorlagen.
- **Restore, das überall sicher ist:** Backup- und Restore-Skript werden
  auf die Compose-Datei parametrisierbar, damit sie auch gegen den
  rocks-Stack korrekt arbeiten (F4).
- **CI, die das Release deckt:** reparierter Windows-CI-Job, shellcheck
  für die Skripte, Dependabot, und der Release-Smoke-Test fährt zusätzlich
  Backup plus Verify gegen den Release-Stack, womit die Ops-Skripte bei
  jedem Release Ende-zu-Ende geprüft sind.
- **Fail-fast statt kryptischer Fehler:** prod-init und prod-update
  verlangen ein gepinntes Release-Tag und brechen sonst mit klarer
  Meldung ab; die `.env`-Vorlage führt nicht mehr auf `latest`.
- **Betriebsrobustheit:** nginx-Reload-Loop auf dem Demo-VPS,
  Log-Rotation in allen Compose-Dateien, non-root Backend/Migrate,
  Healthchecks für die restlichen Services, Start-Retry gegen Postgres,
  Versionsauskunft im Health-Endpoint.
- **Windows:** Pre-Update-Dumps werden zusätzlich ins Windows-Dateisystem
  gespiegelt, damit ein `docker compose down -v` nicht Daten und Backups
  zugleich vernichtet; die Anleitung erklärt das manuelle Sichern.

## User Stories

1. Als Self-Hosting-Admin möchte ich, dass `make prod-backup` auch ohne
   `BACKUP_DIR`/`BACKUP_KEEP` in der `.env` einen Dump erzeugt, damit das
   Backup nicht lautlos ausfällt (F2).
2. Als Self-Hosting-Admin möchte ich, dass alle prod-Skripte fehlende
   `.env`-Schlüssel tolerieren (derselbe `read_env`-Fehlermodus steckt
   auch in restore, update und init), damit kein Skript mitten im Lauf
   stumm abbricht.
3. Als Self-Hosting-Admin möchte ich, dass der mitgelieferte
   systemd-Timer und die cron-Vorlage nach Anleitung funktionieren
   (vollständige Pfad-Hinweise inklusive `ExecStart`), damit nächtliche
   Backups wirklich entstehen (F3).
4. Als Self-Hosting-Admin möchte ich, dass jeder Dump direkt nach dem
   Schreiben auf Archiv-Integrität geprüft wird, damit eine korrupte Datei
   nie als Backup zählt.
5. Als Self-Hosting-Admin möchte ich optional eine Ping-URL hinterlegen
   können, die nur nach erfolgreichem Backup aufgerufen wird, damit ein
   ausbleibendes Backup bei einem Dienst meiner Wahl Alarm auslöst statt
   unbemerkt zu bleiben.
6. Als Self-Hosting-Admin möchte ich ein Verify-Skript, das den neuesten
   Dump in eine Wegwerf-Datenbank einspielt und das Ergebnis prüft, damit
   ich weiß, dass meine Backups wiederherstellbar sind, ohne den
   Live-Betrieb anzufassen.
7. Als Self-Hosting-Admin möchte ich bei leerer oder ungültiger
   `JOTTI_VERSION` (auch `latest`) eine klare Abbruchmeldung von
   prod-init und prod-update, damit ich nicht an einem kryptischen
   Pull-Fehler rätsle.
8. Als Self-Hosting-Admin möchte ich eine `.env`-Vorlage, die mich aktiv
   zum Eintragen eines Release-Tags auffordert, damit der geführte Weg
   gar nicht erst in die `latest`-Falle führt.
9. Als Self-Hosting-Admin möchte ich beim Backup keine verwirrenden
   Warnungen über nicht gesetzte Variablen sehen (F1), damit ich der
   Skript-Ausgabe vertrauen kann.
10. Als Self-Hosting-Admin möchte ich im Leitfaden einen konkreten,
    kopierbaren Befehl zum Wegkopieren der Backups auf einen anderen
    Rechner, damit die gesetzliche Aufbewahrung praktikabel wird.
11. Als Demo-Betreiber möchte ich das Restore-Skript gefahrlos gegen den
    rocks-Stack ausführen können (Compose-Datei parametrisierbar, am Ende
    Reverse-Proxy-Neuerstellung), damit ein Restore nie den laufenden
    Stack nach Prod-Konfiguration umbaut (F4).
12. Als Demo-Betreiber möchte ich, dass nginx erneuerte Zertifikate
    automatisch übernimmt, damit die Demo nicht nach Wochen ohne Deploy
    mit abgelaufenem Zertifikat ausfällt.
13. Als Kassen-Verantwortlicher (Windows) möchte ich, dass die
    automatischen Pre-Update-Dumps zusätzlich im Windows-Dateisystem
    liegen, damit ein versehentliches Entfernen der Docker-Volumes nicht
    Daten und Backups zugleich vernichtet.
14. Als Kassen-Verantwortlicher möchte ich in der Anleitung einen kurzen
    Abschnitt zum manuellen Sichern (ein Befehl, Ergebnis z. B. auf einen
    USB-Stick kopieren), damit ich nach dem Fest eine externe Kopie habe.
15. Als Betreiber (alle Wege) möchte ich, dass Container-Logs überall
    rotiert werden, damit weder VPS noch Kassenrechner an vollgelaufenen
    Log-Dateien scheitern.
16. Als Betreiber möchte ich, dass Backend und Migrate nicht als root
    laufen, damit ein kompromittierter Prozess weniger Schaden anrichten
    kann.
17. Als Betreiber möchte ich Healthchecks für Frontend, Reverse-Proxy und
    Website, damit `docker ps` und die Update-Skripte den echten Zustand
    aller Services zeigen.
18. Als Betreiber möchte ich, dass das Backend beim Start begrenzt auf die
    Datenbank wartet statt sofort zu sterben, damit Boot-Reihenfolgen
    (etwa nach Stromausfall) keine Crash-Loops und Fehler-Logs erzeugen.
19. Als Support-Gebender möchte ich, dass der Health-Endpoint die laufende
    Version meldet, damit ich am Telefon oder per Ferndiagnose sofort
    weiß, welche Version ein Verein betreibt.
20. Als Maintainer möchte ich, dass Änderungen unter `windows/` in der CI
    formatiert, gelintet, getestet und gebaut werden, damit keine
    ungetesteten Windows-Exes mehr veröffentlicht werden.
21. Als Maintainer möchte ich, dass der Release-Smoke-Test nach dem
    Health-Check zusätzlich das Backup-Skript und das Verify-Skript gegen
    den Release-Stack fährt, damit die Ops-Skripte bei jedem Release
    Ende-zu-Ende geprüft sind und ein F2-artiger Bug nie wieder in ein
    Release gelangt.
22. Als Maintainer möchte ich shellcheck über alle Shell-Skripte in der
    CI, damit Standard-Shell-Fehler früh auffallen.
23. Als Maintainer möchte ich Dependabot für GitHub Actions, Go-Module,
    npm-Pakete und Docker-Basis-Images (monatlich, gebündelt), damit
    Abhängigkeiten nicht still veralten.
24. Als Maintainer möchte ich einen aufgeräumten website-CI-Job
    (Installation im richtigen Ordner, korrekter Cache-Pfad, keine
    Referenz auf gelöschte Skripte), damit die CI-Konfiguration den
    tatsächlichen Ablauf widerspiegelt.

## Implementation Decisions

**Backup-Skripte.** Der `read_env`-Helfer wird in allen vier prod-Skripten
identisch gefixt (grep-No-Match wird geschluckt, echte Fehler nicht); die
Skripte bleiben bewusst eigenständig kopiert statt eine gemeinsame Datei zu
sourcen, weil sie einzeln kopierbar bleiben sollen. Nach jedem Dump prüft
das Backup-Skript das Archiv mit gzip; bei Fehler wird die Datei verworfen
und das Skript endet mit Fehlerstatus. Neu ist die optionale Variable
`BACKUP_PING_URL` (Umgebung oder `.env`): ist sie gesetzt, ruft das Skript
sie nach erfolgreichem Dump per curl auf (kurzer Timeout, Fehlschlag des
Pings ist eine Warnung, kein Skriptfehler). Die systemd- und cron-Vorlagen
erwähnen in ihren Kommentaren alle anzupassenden Pfade inklusive
`ExecStart`; nach dem F2-Fix funktionieren sie ohne gesetzte
Environment-Zeilen.

**Compose-Parametrisierung (F4).** Backup- und Restore-Skript lesen die
Compose-Datei aus einer Umgebungsvariablen mit Default auf die
prod-Compose-Datei. Das Restore-Skript erzeugt den Reverse-Proxy nach dem
abschließenden Start neu, damit die nginx-502-Falle des rocks-Stacks nicht
greift (für Caddy-Stacks ist der Schritt ein harmloser No-op). Die
interaktive Bestätigung des Restore-Skripts bleibt erhalten. Die
prod-Compose-Datei gibt den nur dort genutzten Variablen (Domain,
Zertifikats-E-Mail) leere Defaults, damit Skriptläufe gegen andere Stacks
keine Warnungen mehr ausgeben (F1).

**Verify-Skript.** Ein neues Skript spielt einen Dump (Argument oder
neuester im Backup-Verzeichnis) in einen Wegwerf-Postgres-Container ein
(gleiche gepinnte Postgres-Version wie der Stack, ON_ERROR_STOP) und
prüft das Ergebnis über die Tabellenanzahl größer null; Ausgabe ist eine
kurze Zusammenfassung plus Exit-Code. Es berührt weder den laufenden Stack
noch dessen Volumes. Der Leitfaden empfiehlt einen gelegentlichen
manuellen Lauf.

**Release-Smoke-Test.** Nach dem bestehenden Health-Check ruft der
Workflow das Backup-Skript (mit der Release-Compose-Datei) und danach das
Verify-Skript auf den erzeugten Dump auf. Damit sind Backup-Pfad und
Verify-Pfad pro Release getestet; der destruktive Restore-Pfad wird
bewusst nicht automatisiert (interaktive Bestätigung bleibt), sein
psql-Kern ist mit dem Verify-Pfad identisch.

**Versions-Pinning.** prod-init und prod-update validieren
`JOTTI_VERSION` als Release-Tag im Format vMAJOR.MINOR.PATCH und brechen
sonst mit einer Meldung ab, die den Fix nennt (Release-Tag in `.env`
eintragen). Die `.env`-Vorlage liefert den Schlüssel leer mit
aufforderndem Kommentar aus. Der Compose-Default bleibt als Fallback
bestehen, wird über den geführten Weg aber nicht mehr erreicht.

**Demo-VPS (rocks).** Der Reverse-Proxy-Service erhält einen periodischen
nginx-Reload (Größenordnung alle 12 Stunden) neben dem Hauptprozess.
Das übernimmt erneuerte Zertifikate und löst die Upstream-Adressen neu
auf, was die dokumentierte 502-Falle entschärft. Die eigentliche
Caddy-Migration des rocks-Stacks ist eine eigene, spätere PRD.

**Log-Rotation.** Alle Compose-Dateien mit dauerhaft laufenden Services
begrenzen ihre Logs einheitlich (json-file, kleine Maximalgröße, wenige
Dateien), einmal pro Datei definiert und je Service referenziert. Damit
wirkt die Begrenzung auf jedem Host ohne Docker-Daemon-Konfiguration.

**Container-Härtung.** Backend- und Migrate-Image legen einen
unprivilegierten Benutzer an und laufen mit ihm; beide Prozesse brauchen
keine Schreibrechte im Dateisystem. Frontend, Website und Reverse-Proxy
erhalten einfache HTTP-Healthchecks in den Compose-Dateien.

**Backend-Start.** Der Verbindungsaufbau zur Datenbank bekommt einen
begrenzten Retry (Größenordnung 30 Sekunden mit kurzen Abständen) statt
des sofortigen Fatals beim ersten fehlgeschlagenen Ping. Danach gilt
weiterhin: ohne Datenbank kein Start.

**Version im Backend.** Die Version wird wie beim Windows-Starter per
ldflags einkompiliert (Docker-Build-Argument, vom Release-Workflow mit dem
Tag befüllt, Default „dev") und im Health-Endpoint als zusätzliches Feld
ausgegeben. Die Response bleibt abwärtskompatibel (nur ein neues Feld).

**Windows-Host-Spiegel.** Der Starter kopiert jeden Pre-Update-Dump nach
dem Erzeugen zusätzlich in ein Backup-Verzeichnis unter dem bestehenden
Zustandsordner im Windows-Dateisystem und rotiert dort mit derselben
Aufbewahrungszahl wie im Volume. Ein Fehlschlag des Spiegels ist eine
Warnung, kein Startabbruch (der Dump im Volume existiert bereits). Die
Entscheidungslogik (welche Dateien kopieren, welche rotieren) liegt als
reine Funktion im core-Paket, die Seiteneffekte im dünnen Wrapper, analog
zur bestehenden Struktur.

**CI.** Der Windows-Job filtert auf `windows/**` und arbeitet in den
tatsächlichen Modulverzeichnissen. Der website-Job installiert die
Abhängigkeiten im website-Ordner und cached dessen Lockfile; die Referenz
auf das gelöschte Website-Skript im Pfadfilter entfällt. Neu: ein
shellcheck-Job mit Pfadfilter auf die Skripte und eine
Dependabot-Konfiguration für Actions, alle Go-Module, beide npm-Projekte
und die Docker-Basis-Images, monatlich und gruppiert.

**Doku.** Der Leitfaden-Abschnitt zu Backups erklärt Ping-URL,
Verify-Skript und einen konkreten Offsite-Kopierbefehl. Die
Windows-Kurzanleitung bekommt einen Absatz zum manuellen Sichern nach dem
Fest (ein Befehl, extern kopieren). Die Restore-Doku erwähnt die
Compose-Parametrisierung.

## Testing Decisions

Ein guter Test prüft von außen sichtbares Verhalten, nicht die
Implementierung: für Skripte heißt das „Dump entsteht, Verify meldet
Erfolg, Health antwortet", nicht „welche Befehle intern laufen".

- **Shell-Skripte:** keine Unit-Test-Infrastruktur (kein bats); Abdeckung
  Ende-zu-Ende über den Release-Smoke-Test (Backup plus Verify gegen den
  echten Release-Stack) und statisch über shellcheck.
- **Windows-Starter (Host-Spiegel):** Unit-Tests für die reine
  Kopier-/Rotationslogik im core-Paket; Vorbild sind die bestehenden
  core-Tests (Backup-, Update-, Env-Logik).
- **Health-Endpoint:** bestehender Handler-Test wird um das Versionsfeld
  erweitert (Unit-Tag, Vorbild vorhandene health- und app-Tests).
- **Postgres-Retry:** kleiner Unit-Test für den Retry-Helfer
  (Abbruchbedingungen, Erfolgsfall), Seiteneffekte gemockt.
- **Compose-/Dockerfile-Änderungen:** implizit über den Release-Smoke-Test
  (Stack wird mit non-root Images, Healthchecks und Log-Optionen
  hochgefahren und muss gesund werden).

## Out of Scope

- **Multi-Arch-Images (arm64):** amd64 genügt aktuell; die große Mehrheit
  installiert auf Standard-Windows-64-bit, ein Raspberry Pi kann den rohen
  Linux-Weg nutzen.
- **Externes Uptime-Monitoring** (inklusive Cert-Expiry-Überwachung) für
  jotti.rocks und Self-Hoster.
- **Offsite-/USB-Backup-Skripte** (rclone, Storage Box, USB-Automatik):
  bleibt manuell, dafür der Doku-Hinweis (User Stories 10 und 14).
- **Caddy-Migration des rocks-Stacks:** eigene PRD zu einem späteren
  Zeitpunkt; diese PRD liefert nur den Reload-Loop als Interims-Fix.
- **rocks auf GHCR-Images umstellen** (pull statt build auf dem VPS).
- **Strukturierte JSON-Logs / Log-Aggregation:** die Zielgruppe liest
  `docker logs`, der ConsoleWriter bleibt.
- **Image-Vulnerability-Scanning, SBOM, Signierung.**
- **Windows-Autostart nach Reboot:** der dokumentierte Weg (Starter erneut
  doppelklicken) bleibt.

## Further Notes

- Die Findings F1 bis F4 samt Mechanismus, Evidenz und verifizierten
  Fix-Diffs stehen in `docs/plans/findings-phase4-backup-test.md`; die
  Zeilenreferenzen dort gelten für Commit 21d8c4b.
- Zur Einordnung des Windows-Anteils: `prd-windows-nativ-ohne-docker.md`
  will die Docker-Basis unter Windows langfristig ablösen. Der
  Host-Spiegel hier ist bewusst klein gehalten und lohnt sich trotzdem,
  weil die Docker-Verpackung bis dahin der ausgelieferte Weg ist.
- Der Erfolgs-Ping ist bewusst ein Dead-Man-Switch: alarmiert wird beim
  Ausbleiben des Pings durch den externen Dienst, nicht durch jotti
  selbst. Ein Fehlschlag-Ping (etwa ein /fail-Endpunkt) ist nicht Teil
  dieser PRD.
- Reihenfolge-Empfehlung für den späteren Plan: zuerst die reinen
  Skript-Fixes (F2, F4, gzip, Ping), dann Verify-Skript plus
  Smoke-Test-Erweiterung (baut auf F4 auf), dann CI-Reparaturen, dann
  Compose-/Image-Härtung, zuletzt Starter-Spiegel und Doku.
