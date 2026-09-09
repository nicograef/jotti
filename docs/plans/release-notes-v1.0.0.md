# Release-Notes-Entwurf v1.0.0

Vorlage für den Text des GitHub-Releases; `release.yml` veröffentlicht das Release mit den
git-cliff-Notes aus den Commits, dieser Entwurf ist der Text davor. Nach dem Release löschen.

## Text

jotti 1.0.0 ist die erste stabile Version des kostenlosen Mobile-Kassensystems für Vereinsfeste:
Servicekräfte nehmen auf ihren eigenen Smartphones Bestellungen auf, kassieren und stornieren,
Admins verwalten Produkte, Tische und Benutzer. Alle Vorgänge werden von einer BSI-zertifizierten
Cloud-TSE signiert, das Kassenjournal ist append-only und der DSFinV-K-Export liegt bereit.

Alle Funktionen und Änderungen im Detail: Abschnitt `[1.0.0]` in `CHANGELOG.md`.

### Enthaltene Laufzeit-Versionen

| Laufzeit   | Version           | Quelle                                                       |
| ---------- | ----------------- | ------------------------------------------------------------ |
| Go         | 1.27.1            | `backend/go.mod`, `backend/Dockerfile` (`golang:1.27.1-alpine`) |
| Node       | 24 (Alpine)       | `frontend/Dockerfile` (`node:24-alpine`)                     |
| pnpm       | 11.6.0            | `frontend/package.json` (`packageManager`)                   |
| PostgreSQL | 17.8              | `docker-compose.prod.yml` (`postgres:17.8`)                  |
| Caddy      | 2.11.4            | `reverse-proxy/Dockerfile` (`caddy:2.11.4`)                  |

Ein Gate hält jede dieser Versionen über alle Stacks hinweg auf einem Wert: `scripts/check-pins.sh`,
Teil von `make check-repo`.

### Aktualisieren

- **Windows-Rechner (Standardweg):** `jotti-stop.cmd`, neues Release-ZIP entpacken,
  `jotti-start.exe` starten. Der Starter sichert die Datenbank automatisch, bevor er
  aktualisiert. Ablauf, Rauchtest und der Weg zurück: `docs/leitfaden/aktualisieren.md`.
- **Eigener Server (Self-Hosting):** `JOTTI_VERSION=v1.0.0` in `.env` setzen und `make prod-update`
  ausführen. Das Skript zieht vor der Aktualisierung ein Backup. Ablauf und Backup-Strategie:
  `docs/leitfaden/aktualisieren-backups.md`.

## Schritt nach dem Tag (Eigentümer)

Nach dem Tag wird `PREVIOUS_VERSION` an beiden Stellen auf `v1.0.0` gehoben —
`.github/workflows/ci.yml` (Job `upgrade-path`) und `database/migrations/README.md`
(Abschnitt „Vorversions-Pinning"). Der Bump ist ein eigener Commit nach dem Tag: vorher gibt
es die `v1.0.0`-Images noch nicht.
