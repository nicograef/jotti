---
title: Aktualisieren und Backups
description: 'Den selbst gehosteten jotti-Server sicher aktualisieren und regelmäßige Datenbank-Backups einrichten.'
---

## Aktualisieren

Tragt unter `JOTTI_VERSION` die gewünschte Release-Version ein und führt
`make prod-update` aus. Das Skript sichert die Datenbank automatisch, bevor es die
neuen Images zieht und die Migrationen ausführt, und prüft danach die Gesundheit.
Bleibt der Stack ungesund, bricht es ab und zeigt, wie ihr mit dem eben erstellten
Backup zurückkehrt.

> 🔁 Nur vorwärts, kein Downgrade. Updates verändern die Datenbank und lassen sich
> nicht zurücknehmen; eine ältere Version kann mit den neuen Daten nicht mehr
> starten. Wollt ihr zurück, spielt ein Backup ein (`make prod-restore`).

## Backups

jotti speichert alle Daten in einem Docker-Volume. Macht regelmäßige Backups, schon
wegen der gesetzlichen 10-Jahre-Aufbewahrung.

- **Sichern:** `make prod-backup` zieht einen komprimierten `pg_dump` in den Ordner
  `BACKUP_DIR` (Standard `./backups`) und behält die neuesten `BACKUP_KEEP` Stück.
- **Wiederherstellen:** `make prod-restore` listet die Backups, fragt eine
  Bestätigung ab und spielt das gewählte zurück. Das überschreibt die aktuelle
  Datenbank. Die Bestätigung nennt Dump und Compose-Datei; standardmäßig
  `docker-compose.prod.yml`, über `COMPOSE_FILE` auf einen anderen Stack
  umstellbar (gilt ebenso für `make prod-backup`).
- **Prüfen (gelegentlich):** `make prod-backup-verify` spielt das neueste Backup
  in einen Wegwerf-Postgres ein und meldet die Tabellenzahl. So wisst ihr, dass
  ein Backup wirklich wiederherstellbar ist, ohne den laufenden Betrieb
  anzufassen. Einen bestimmten Dump prüft ihr per Argument:
  `./scripts/prod-backup-verify.sh <datei>`.
- **Täglich automatisch:** Für einen täglichen Dump liegen Vorlagen im Repository
  (systemd-Timer unter `packaging/systemd/` oder cron unter `packaging/cron/`); die
  Installationsschritte stehen als Kommentar in den Dateien.
- **Überwachung (optional):** Tragt unter `BACKUP_PING_URL` die URL eines
  Überwachungsdienstes ein (z. B. healthchecks.io). Nach jedem erfolgreichen Backup
  ruft das Skript sie einmal auf. Bleibt der Aufruf aus, schlägt der Dienst Alarm
  (Dead-Man-Switch). So merkt ihr, wenn ein automatischer Lauf still ausfällt.
  Welchen Dienst ihr nutzt, ist frei wählbar; ein Fehlschlag des Aufrufs gefährdet
  das Backup nicht.
- **Extern kopieren:** Holt die Dumps auf einen anderen Rechner, damit sie einen
  Serverausfall überstehen. Ein Einzeiler dafür, vom Zielrechner aus (Pfad und
  Zugang anpassen):

  ```sh
  rsync -avz admin@SERVER:/pfad/zu/jotti/backups/ ./jotti-backups/
  ```

> 💾 Ein Backup, das nur auf demselben Server liegt, hilft bei dessen Ausfall
> nicht. Kopiert die Dumps regelmäßig weg.
