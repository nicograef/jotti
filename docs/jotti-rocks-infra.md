# DNS-Infrastruktur für vertrauenswürdiges lokales TLS (jotti.rocks)

Maintainer-Runbook für die zentrale `jotti.rocks`-Infrastruktur, nicht für Vereine: die
DNS-Dienste, über die lokale jotti-Installationen echte Let's-Encrypt-Zertifikate für
`*.<install-id>.lokal.jotti.rocks` beziehen.

## 1. Services

Zwei Services im rocks-Stack (`docker-compose.rocks.yml`):

| Service    | Zone                | Erreichbarkeit                                         |
| ---------- | ------------------- | ------------------------------------------------------ |
| `resolver` | `lokal.jotti.rocks` | öffentlich, Port 53 UDP+TCP (einziger Prozess auf :53) |
| `acme-dns` | `auth.jotti.rocks`  | nur Docker-intern (DNS via resolver, API via Caddy)    |

Der resolver beantwortet A-Records und `_acme-challenge`-CNAMEs zustandslos und rein
rechnerisch aus dem angefragten Namen (Mapping Name → IP unveränderlich).
Anfragen für `auth.jotti.rocks` reicht er Docker-intern an acme-dns weiter.

acme-dns verwaltet die TXT-Records der DNS-01-Challenges. Der Zustand (SQLite) liegt im
Volume `acme-dns-data`.

Seine HTTP-API (`/register`, `/update`, `/health`) läuft hinter Caddy unter
`https://auth.jotti.rocks`. `/register` ist dort streng rate-limitiert (1 Anfrage/Minute je IP).

Reverse-Proxy ist das Caddy-Image aus `reverse-proxy/Dockerfile`, gestartet direkt mit
der statischen `reverse-proxy/Caddyfile.rocks`. `make rocks-up` erstellt den Proxy neu und
übernimmt so Änderungen am Caddyfile.

Caddy holt und erneuert die Zertifikate aller vier Hosts selbst per HTTP-01. Sie liegen im
Volume `caddy-data`.

## 2. Voraussetzungen

1. Port 53 frei. Auf dem VPS prüfen, dass nichts öffentlich auf :53 lauscht
   (systemd-resolved bindet nur Loopback: `127.0.0.53` und `127.0.0.54`, unkritisch):

   ```bash
   sudo ss -lnup 'sport = :53'
   sudo ss -lntp 'sport = :53'
   ```

2. Provider-Firewall (z. B. Hetzner Cloud Firewall) erlaubt eingehend 53/UDP und 53/TCP.
3. DNS-Hoster unterstützt NS-Records für Subdomain-Delegation (Standard).

## 3. DNS-Hoster-Einträge (einmalig, manuell)

Beim DNS-Hoster der Zone `jotti.rocks` anlegen:

| Typ | Name                | Wert                        | Zweck                                 |
| --- | ------------------- | --------------------------- | ------------------------------------- |
| A   | `dns.jotti.rocks`   | `<VPS-IP>`                  | Adresse des eigenen Nameservers       |
| NS  | `lokal.jotti.rocks` | `dns.jotti.rocks`           | Delegation an den resolver            |
| NS  | `auth.jotti.rocks`  | `dns.jotti.rocks`           | Delegation an acme-dns (via resolver) |
| CAA | `jotti.rocks`       | `0 issue "letsencrypt.org"` | nur Let's Encrypt darf ausstellen     |

Der A-Record für `auth.jotti.rocks` selbst kommt aus acme-dns
(`docker-compose.rocks.yml`) und zeigt auf den VPS; beim Hoster ist dafür nichts
einzutragen.

## 4. Deployment

In der `.env` auf dem VPS die öffentliche IPv4 eintragen (der resolver serviert sie als
NS-A-Record, acme-dns als Zone-Apex-A-Record):

```bash
VPS_PUBLIC_IP=<öffentliche IPv4 des VPS>
```

Frischer VPS: `make rocks-init` baut und startet den Stack, wartet auf die Healthchecks
und prüft HTTPS. Danach aktualisiert `make rocks-up` den Stack.

Einmaliger Umstieg eines VPS, auf dem noch der nginx/certbot-Stack läuft: Caddy holt seine
Zertifikate per HTTP-01-Challenge über Port 80.

- Port 80 darf nur der Container `jotti-reverse-proxy` belegen, den `make rocks-up` durch
  Caddy ersetzt.
- Ein anderer Prozess auf Port 80 muss vorher weg (`sudo ss -ltnp 'sport = :80'` zeigt ihn).
- `make rocks-up` entfernt den verwaisten Container `jotti-certbot` selbst
  (`--remove-orphans`).

Danach die alten Zertifikats-Volumes löschen:

```bash
docker volume rm jotti_letsencrypt jotti_certbot-challenges
```

`auth.jotti.rocks` löst erst auf, wenn der Stack läuft und die Delegation (Abschnitt 3)
aktiv ist. Bis dahin scheitert Caddys Zertifikatsanfrage für diesen Host; Caddy
wiederholt sie selbst.

## 5. End-to-End-Verifikation (nach jedem Infra-Setup)

Von außerhalb des VPS prüfen:

```bash
# Delegation + berechneter A-Record
dig +short NS lokal.jotti.rocks                  # → dns.jotti.rocks.
dig +short 10-0-0-1.test.lokal.jotti.rocks A     # → 10.0.0.1

# Berechneter Challenge-CNAME
dig +short CNAME _acme-challenge.test.lokal.jotti.rocks   # → test.auth.jotti.rocks.

# CAA
dig +short CAA jotti.rocks                       # → 0 issue "letsencrypt.org"
```

Registrierung (liefert `username`, `password`, `subdomain`, `fulldomain`, für die
folgenden Schritte aufheben). Ein zweiter Aufruf innerhalb einer Minute muss HTTP 429
liefern:

```bash
curl -s -X POST https://auth.jotti.rocks/register
```

TXT-Update nur mit gültigen Credentials (TXT-Wert exakt 43 Zeichen). Derselbe Aufruf
ohne oder mit falschen Headern muss abgelehnt werden:

```bash
curl -s -X POST https://auth.jotti.rocks/update \
  -H "X-Api-User: <username>" -H "X-Api-Key: <password>" \
  -d '{"subdomain": "<subdomain>", "txt": "0123456789012345678901234567890123456789012"}'

dig +short TXT <subdomain>.auth.jotti.rocks      # → der gesetzte Wert
```

Staging-Wildcard-Zertifikat als Beweis, dass der DNS-01-Pfad komplett steht
(Let's-Encrypt-Staging schont die Rate-Limits der echten Zone). Erwartung:
`Cert success.`, das Zertifikat wird verworfen:

```bash
docker run --rm \
  -e ACMEDNS_BASE_URL=https://auth.jotti.rocks \
  -e ACMEDNS_USERNAME=<username> \
  -e ACMEDNS_PASSWORD=<password> \
  -e ACMEDNS_SUBDOMAIN=<subdomain> \
  neilpang/acme.sh acme.sh --issue --server letsencrypt_test \
  -m graef.nico@gmail.com --dns dns_acmedns \
  -d "*.<subdomain>.lokal.jotti.rocks"
```

## 6. Laufender Betrieb

CT-Log-Monitoring (monatlich): <https://crt.sh/?q=%25.lokal.jotti.rocks> aufrufen und das
Ausstellungsvolumen prüfen.

- Erwartung: einzelne Wildcard-Zertifikate je Install-ID, in der Größenordnung der bekannten
  Installationen.
- Auffällige Spitzen (Massen-Registrierungen) sind ein Missbrauchssignal.

Eskalation bei Missbrauch: Registrierung schließen mit `disable_registration = true` in der
acme-dns-Config (`docker-compose.rocks.yml`), dann `make rocks-up`. Bestehende Installationen
erneuern weiter (Credentials bleiben gültig), nur neue Registrierungen sind blockiert.

Monitoring: Der Betreiber richtet bei einem Uptime-Dienst (z. B. Better Stack) diese
Monitore ein:

- HTTPS-Monitore auf `https://jotti.rocks`, `https://demo.jotti.rocks` und
  `https://auth.jotti.rocks/health`, jeweils mit Alarm vor dem Ablauf des Zertifikats.
- Einen DNS-Monitor auf den resolver: A-Abfrage von `10-0-0-1.test.lokal.jotti.rocks`,
  erwartet `10.0.0.1`.

AVV (Datenschutz): Für den VPS besteht eine Vereinbarung zur Auftragsverarbeitung nach
Art. 28 DSGVO mit netcup (abgeschlossen 2026-07-14). Der Vertragsinhalt ist vertraulich
(Ziff. 11 der Vereinbarung) und gehört nicht ins Repository.

Kopien liegen im netcup-CCP (Stammdaten → Auftragsverarbeitung) und im privaten
Vertragsarchiv.

## 7. Backup der acme-dns-Datenbank

Das Volume `acme-dns-data` enthält die Zuordnung Account ↔ Subdomain. Geht es verloren,
werden die Credentials aller bestehenden Installationen ungültig.

Ihre Zertifikats-Erneuerungen schlagen dann fehl. Abhilfe je Installation: lokalen State
löschen, neu registrieren, neue Install-ID, neue Adresse.

Das Backup läuft manuell vom Laptop aus, zum Beispiel auf eine externe Platte:

```bash
make rocks-backup DEST=/media/<platte>/jotti-rocks
```

Das Skript `scripts/rocks-backup.sh` zieht per SSH ein `sqlite3 .backup` der Datenbank.
Es prüft die Kopie auf dem VPS mit `PRAGMA integrity_check`. Danach kopiert es die Datei
als `acme-dns-<Zeitstempel>.db` per rsync ins Zielverzeichnis.

Auf dem VPS müssen `sqlite3` und `rsync` installiert sein. Der SSH-Zugang braucht
Leserechte auf das Volume.

Host, SSH-Optionen und Datenbankpfad sind über `ROCKS_SSH_HOST`, `ROCKS_SSH_OPTS` und
`ROCKS_ACMEDNS_DB` einstellbar (`./scripts/rocks-backup.sh` ohne Argument zeigt sie).

Wiederherstellungsprobe nach jedem Backup: die Zahl der Registrierungen in der Kopie muss
der auf dem VPS entsprechen.

```bash
sqlite3 /media/<platte>/jotti-rocks/acme-dns-<Zeitstempel>.db \
  "PRAGMA integrity_check; SELECT count(*) FROM records;"
ssh jotti.rocks sqlite3 /var/lib/docker/volumes/jotti_acme-dns-data/_data/acme-dns.db \
  "'SELECT count(*) FROM records;'"
```

Wiederherstellung auf dem VPS:

```bash
docker compose -f docker-compose.rocks.yml stop acme-dns
sudo cp acme-dns-<Zeitstempel>.db /var/lib/docker/volumes/jotti_acme-dns-data/_data/acme-dns.db
docker compose -f docker-compose.rocks.yml start acme-dns
```
