# Plan: Lokale Transportverschlüsselung — selbstsigniertes TLS

> Source PRD: `docs/prds/prd-lokale-tls-selbstsigniert.md`

## Goal

Der lokale LAN-Stack (`docker-compose.local.yml`) serviert HTTPS mit einem
automatisch erzeugten selbstsignierten Zertifikat. Passives Mitlesen von
Passwörtern und JWTs im WLAN wird ausgeschlossen. HTTP leitet auf HTTPS um.
Die Browserwarnung beim ersten Zugriff ist ein bewusster Kompromiss.

## Architectural decisions

- **Zertifikatserzeugung**: Custom-Entrypoint-Script im nginx-Container. Das
  Script erkennt die LAN-IP automatisch (Default-Route-Gateway), erzeugt das
  Zertifikat per `openssl` falls fehlend oder IP geändert, dann `exec nginx`.
- **SAN**: `IP:<detected-LAN-IP>` + `IP:127.0.0.1` + `DNS:localhost`. Bei
  IP-Wechsel wird das Zertifikat automatisch neu erzeugt.
- **Validität**: 365 Tage.
- **Speicherung**: Named Docker Volume (`tls-certs`), persistiert über Restarts.
- **Ports**: 80 (HTTP → HTTPS-Redirect) + 443 (HTTPS).
- **In-Place-Änderung**: `docker-compose.local.yml` und `nginx.local.conf`
  werden direkt modifiziert. Kein HTTP-only-Fallback.

## Inventory

- `docker-compose.local.yml:1-125` — lokaler Stack, aktuell HTTP-only auf Port 80
- `reverse-proxy/nginx.local.conf:1-48` — HTTP-only nginx-Konfiguration
- `reverse-proxy/nginx.conf:1-84` — Prod-HTTPS-Konfiguration (Muster für TLS-Blöcke, Headers, Ciphers)
- `scripts/init-env.sh:1-60` — Env-Initialisierung (Muster für idempotente Bash-Scripte)
- `docs/betrieb/leitfaden-hosting.md:57-130` — Weg-A-Anleitung (aktuell HTTP)
- `docs/anforderungen.md:143` — Q-06 (nur Let's Encrypt erwähnt)
- `Makefile:152-155` — `local-up`/`local-down` Targets

## Resolved decisions

- Bestehende Dateien werden in-place geändert (kein Overlay-Ansatz).
- Zertifikat wird im nginx-Entrypoint erzeugt (kein separater Init-Container).
- LAN-IP wird automatisch erkannt; bei IP-Wechsel wird das Zertifikat regeneriert.
- Zertifikat ist 365 Tage gültig.

---

## Phase 1: TLS-Infrastruktur

**User stories**: 1, 2, 3

### Context

- `docker-compose.local.yml:99-115` — reverse-proxy Service-Definition (Port 80, kein Volume für Certs)
- `reverse-proxy/nginx.local.conf:16-48` — HTTP-Server-Block
- `reverse-proxy/nginx.conf:45-84` — Prod-HTTPS-Server als Muster (TLS 1.2/1.3, HSTS, Cipher-Config)

### What to build

Ein Entrypoint-Script (`reverse-proxy/local-entrypoint.sh`) das:

1. Die LAN-IP des Containers ermittelt (Default-Route → Gateway-Interface → IP).
2. Prüft ob `/etc/nginx/certs/selfsigned.crt` existiert UND die gespeicherte IP
   (`/etc/nginx/certs/.lan-ip`) mit der aktuellen übereinstimmt.
3. Falls nicht: ein neues selbstsigniertes Zertifikat mit `openssl req -x509`
   erzeugt (SAN: erkannte LAN-IP + 127.0.0.1 + localhost, 365 Tage, RSA 2048).
4. `exec nginx -g 'daemon off;'` startet.

Die nginx-Konfiguration (`nginx.local.conf`) wird auf dual-server umgebaut:

- Port 80: `return 301 https://$host$request_uri`
- Port 443: HTTPS mit dem selbstsignierten Zertifikat, TLS 1.2+, Security-Headers
  inkl. HSTS, Proxy-Regeln wie bisher.

`docker-compose.local.yml` wird angepasst:

- Port 443 zusätzlich exponiert.
- Named Volume `tls-certs` für Zertifikat-Persistenz.
- Custom Entrypoint auf das Script zeigend.
- Das Script als Volume-Mount (read-only).

### Acceptance criteria

- [x] `docker compose -f docker-compose.local.yml up -d --build` erzeugt automatisch ein Zertifikat und serviert HTTPS auf Port 443
- [x] `curl -k https://localhost/api/health` liefert 200
- [x] `curl -I http://localhost` liefert 301-Redirect auf `https://`
- [x] Das Zertifikat enthält die LAN-IP und `localhost` als SANs
- [x] Bei erneutem `up` ohne IP-Änderung wird das vorhandene Zertifikat wiederverwendet (Log: „Certificate exists…")
- [x] Bei IP-Änderung (simuliert durch Löschen der `.lan-ip`-Datei) wird das Zertifikat neu erzeugt
- [x] Das Entrypoint-Script ist POSIX-kompatibel (`/bin/sh`, da Alpine)

---

## Phase 2: Dokumentation und Ripple-Updates

**User stories**: 4, 5

### Context

- `docs/betrieb/leitfaden-hosting.md:57-130` — Weg-A-Anleitung (HTTP-Referenzen, Sicherheitshinweis, Schritt-für-Schritt)
- `docs/anforderungen.md:143` — Q-06 beschreibt nur Let's Encrypt
- `Makefile:152-153` — `local-up` Help-Text sagt „HTTP, ohne TLS"
- `docker-compose.local.yml:1-14` — Header-Kommentar sagt „HTTP only, no TLS"

### What to build

Alle Dokumentation und Kommentare, die den lokalen Stack als „HTTP-only"
beschreiben, werden auf den neuen HTTPS-Zustand aktualisiert:

- **Leitfaden-Hosting.md, Weg A**: Sicherheitshinweis ersetzt „unverschlüsselt
  über HTTP" durch „selbstsigniertes HTTPS, einmalige Browserwarnung". Schritt 5
  zeigt `https://` statt `http://`. Neuer Absatz: Restrisiko (aktiver MITM),
  Verweis auf Option 3. Hinweis auf Einzeltheke/localhost-Ausnahme.
- **Anforderungen.md, Q-06**: Beschreibung ergänzen: „Lokal: selbstsigniertes
  Zertifikat, automatisch erzeugt. Prod: Let's Encrypt."
- **Makefile**: `local-up` Help-Text von „HTTP, ohne TLS" auf
  „HTTPS, selbstsigniertes Zertifikat" ändern.
- **docker-compose.local.yml**: Header-Kommentar von „HTTP only" auf
  „self-signed TLS" aktualisieren.

### Acceptance criteria

- [x] `docs/betrieb/leitfaden-hosting.md` Weg A beschreibt HTTPS-Zugriff mit Browserwarnung und Restrisiko-Hinweis
- [x] `docs/anforderungen.md` Q-06 erwähnt den lokalen selbstsignierten Betrieb
- [x] Keine Referenz auf „HTTP-only" / „unverschlüsselt" im Kontext des lokalen Stacks verbleibt
- [x] `make help` zeigt aktualisierte Beschreibung für `local-up`
