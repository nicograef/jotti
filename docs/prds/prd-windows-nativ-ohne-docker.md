# PRD: Native Windows-Verpackung ohne Docker (Phase C, Ziel)

> Vorgänger: `docs/prds/prd-windows-verpackung.md` (Docker-basiertes
> Release-ZIP mit Starter/Relay). Diese PRD ersetzt langfristig dessen
> **Laufzeitbasis** (Docker Desktop) für den lokalen Windows-Betrieb und
> übernimmt dessen Bedien-Erkenntnisse (`.env`-Vertrag, Health-Check,
> Zugriffs-URL, Relay per Doppelklick, kein Autostart).
> **Status: Ziel-Architektur, spätere Ausarbeitung.** Es gibt noch keinen
> Umsetzungsplan; User Stories und Entscheidungen sind bewusst grob.

## Problem Statement

Auch nach der klickbaren Windows-Verpackung bleibt **Docker Desktop die größte
Fragilitäts- und Hürdenquelle** des lokalen Betriebs. Administratorrechte
machen es bedienbarer (Selbststart, Engine-Switch, Firewall), aber nicht
robust:

- **Installation:** eigener Download, WSL2-Aktivierung, ggf. Virtualisierung
  im BIOS, Neustarts — die mit Abstand längste und fehleranfälligste Stelle
  der Kurzanleitung, und genau der Schritt, den ein nicht-technischer
  Vereinshelfer allein bewältigen muss.
- **Laufzeit:** RAM-Bedarf auf Altgeräten, WSL2-/Docker-Desktop-Updates,
  Kaltstart im Minutenbereich, 120-s-Health-Timeouts.
- **Folgekomplexität im Starter:** LAN-IP-Heuristik gegen vEthernet-/
  WSL-Adapter, Engine-Modus-Checks, Image-Pull beim Erststart
  (Internet-Pflicht), eigene Release-Compose-Datei (Drift-Risiko zur lokalen
  Compose-Datei).

## Solution (Skizze)

Eine einzige native `jotti.exe` plus Installer — ganz ohne Docker:

1. **Eine Binärdatei:** Das Go-Backend bettet das gebaute Frontend per
   `go:embed` ein und übernimmt TLS-Terminierung und statische Auslieferung
   selbst — nginx entfällt im lokalen Modus (die frühere „Phase B" geht hierin
   auf).
2. **Gebündelte PostgreSQL:** Die offiziellen PostgreSQL-Windows-Binaries
   liegen im Installationsverzeichnis; jotti startet die Datenbank als
   Kindprozess über `pg_ctl` und beendet sie beim Herunterfahren sauber.
   (`pg_ctl` erzeugt unter Windows automatisch ein restricted token —
   funktioniert also auch aus Admin-Kontext; `postgres.exe` verweigert sonst
   den Start mit Admin-Rechten.)
3. **Migrationen in-process:** golang-migrate läuft heute als CLI im
   migrate-Container und wird stattdessen als Go-Library eingebunden — kein
   separater Migrationsschritt, Migrationen laufen beim Start.
4. **Installer statt ZIP:** Ein Inno-Setup-Installer (auf GitHub-Runnern
   vorinstalliert) installiert nach `C:\Program Files\jotti`, legt Startmenü-/
   Desktop-Verknüpfungen an, setzt die Firewall-Regel (eingehend, TCP 80/443,
   nur lokales Subnetz) **einmalig bei der Installation** und bringt einen
   Uninstaller mit. Administratorrechte braucht nur die Installation — der
   tägliche Doppelklick läuft ohne erhöhte Rechte (Windows kennt keine
   Beschränkung für Ports < 1024; die Firewall-Regel existiert bereits).

Ergebnis: Start in Sekunden statt Minuten, kein Internet beim Erststart, keine
Docker-Preflights, keine Engine-Zustände, keine Image-Pulls, keine
Compose-Drift, keine vEthernet-Adapter, kein UAC-Dialog im Tagesbetrieb.

## User Stories (grob)

1. Als Vereins-Admin möchte ich einen Installer herunterladen und mit
   „Weiter → Fertig" installieren, ohne Docker oder WSL2 zu kennen.
2. Als Vereins-Admin möchte ich jotti über eine Desktop-Verknüpfung starten
   und in wenigen Sekunden die Zugriffs-URL für die Helfer-Smartphones sehen.
3. Als Vereins-Admin möchte ich jotti komplett ohne Internetverbindung
   betreiben können — auch beim allerersten Start auf dem Fest.
4. Als Vereins-Admin möchte ich, dass meine Daten Updates überstehen
   (Datenverzeichnis getrennt vom Programmverzeichnis, z. B. unter
   `%ProgramData%\jotti`).
5. Als Vereins-Admin möchte ich jotti sauber beenden können, wobei die
   Datenbank kontrolliert mit herunterfährt.
6. Als Vereins-Admin möchte ich den Bondruck weiterhin per Doppelklick starten
   (Relay unverändert, vom Installer mit installiert).

## Open Questions

- **TLS ohne nginx:** Option 2 (selbstsigniert) erzeugt das Zertifikat heute im
  reverse-proxy-Entrypoint — diese Logik wandert in die `jotti.exe`. Synergie
  mit Option 3 (`docs/prds/prd-lokale-tls-vertrauenswuerdig.md`): deren
  Caddy-Baustein ließe sich durch die CertMagic-Go-Library (Caddys
  ACME-Engine) direkt im Backend ersetzen — acme-dns/DNS-01 in-process, ein
  Baustein weniger.
- **PostgreSQL-Lebenszyklus:** Ort des Datenverzeichnisses, `initdb` beim
  ersten Start, Major-Upgrades (pg_upgrade beim jotti-Update?), Backups —
  Verantwortung wandert vom Docker-Volume zu jotti selbst.
- **Parität der zwei Laufzeitpfade:** Server-Deployments (rocks/prod) bleiben
  Compose — der native Modus muss funktional identisch bleiben (gleiche
  Events, gleiche Migrationen). Was sichert die Parität — ein Smoke-Test auf
  einem Windows-Runner im Release-Workflow?
- **Update-Mechanik:** Installer über bestehende Installation; genügt der
  Migrationslauf beim ersten Start nach dem Update?
- **Code-Signing:** Bleibt es ohne Signatur (SmartScreen-Klickweg), oder lohnt
  ein Zertifikat, sobald es einen Installer gibt?

## Out of Scope

- Ablösung von Docker/Compose auf den Server-Deployments (rocks/prod).
- macOS-/Linux-Pakete.
- **Autostart / Windows-Dienst** — wie in der Vorgänger-PRD bewusst abgelehnt
  (eintägige Feste, lange Pausen; täglicher manueller Start gewollt).
- Änderungen an POST-only, Event-Sourcing oder Datenmodell.
- Wechsel des Datenbanksystems (es bleibt PostgreSQL, nur gebündelt statt
  containerisiert).
