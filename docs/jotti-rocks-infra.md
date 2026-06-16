# DNS-Infrastruktur für vertrauenswürdiges lokales TLS (jotti.rocks)

Maintainer-Runbook für die zentrale `jotti.rocks`-Infrastruktur, nicht für Vereine.
Es beschreibt Einrichtung, Verifikation und Betrieb der DNS-Dienste, über die lokale
jotti-Installationen echte Let's-Encrypt-Zertifikate für
`*.<install-id>.lokal.jotti.rocks` beziehen.

## 1. Überblick

Zwei Services im rocks-Stack (`docker-compose.rocks.yml`):

| Service    | Zone                | Erreichbarkeit                                          |
| ---------- | ------------------- | ------------------------------------------------------- |
| `resolver` | `lokal.jotti.rocks` | öffentlich, Port 53 UDP+TCP (einziger Prozess auf :53)   |
| `acme-dns` | `auth.jotti.rocks`  | nur Docker-intern (DNS via resolver, API via nginx)      |

Der resolver beantwortet A-Records und `_acme-challenge`-CNAMEs zustandslos und rein
rechnerisch aus dem angefragten Namen (TTL 86400, Mapping Name → IP unveränderlich).
Anfragen für `auth.jotti.rocks` reicht er Docker-intern an acme-dns weiter.

acme-dns verwaltet die TXT-Records der DNS-01-Challenges. Seine HTTP-API (`/register`,
`/update`, `/health`) läuft hinter nginx unter `https://auth.jotti.rocks`; `/register`
ist dort streng rate-limitiert (1 Anfrage/Minute je IP). Der Zustand (SQLite) liegt im
Volume `acme-dns-data`.

Der rocks-Stack nutzt nginx als Reverse-Proxy, der Prod-Stack
(`docker-compose.prod.yml`) dagegen Caddy. Die nginx-Befehle hier gelten nur für rocks.

## 2. Voraussetzungen

1. Port 53 frei. Auf dem VPS prüfen, dass nichts öffentlich auf :53 lauscht
   (systemd-resolved bindet nur `127.0.0.53`, unkritisch):

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

Bestehendes Deployment nachrüsten:

```bash
git pull
make rocks-up
```

Damit laufen resolver und acme-dns. `auth.jotti.rocks` ist aber erst im Zertifikat,
sobald die DNS-Hoster-Einträge (Abschnitt 3) aktiv sind; dann das Zertifikat erweitern
und nginx neu laden:

```bash
docker compose -f docker-compose.rocks.yml \
  run --rm --entrypoint certbot certbot certonly \
  --webroot -w /var/www/certbot \
  -d jotti.rocks -d www.jotti.rocks -d demo.jotti.rocks -d auth.jotti.rocks \
  --email graef.nico@gmail.com --agree-tos --no-eff-email --expand

docker compose -f docker-compose.rocks.yml \
  exec reverse-proxy nginx -s reload
```

Frischer VPS: `./scripts/rocks-init.sh` wie gehabt. `auth.jotti.rocks` löst erst auf,
wenn der Stack läuft und die Delegation gesetzt ist; bei der Ersteinrichtung überspringt
das Skript die Domain (Warnung) und das Zertifikat wird anschließend wie oben erweitert.

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
folgenden Schritte aufheben). Mehrere schnelle Aufrufe müssen HTTP 429 liefern:

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
Ausstellungsvolumen prüfen. Erwartung: einzelne Wildcard-Zertifikate je Install-ID, in der
Größenordnung der bekannten Installationen. Auffällige Spitzen (Massen-Registrierungen)
sind ein Missbrauchssignal.

Eskalation bei Missbrauch: Registrierung schließen mit `disable_registration = true` in der
acme-dns-Config (`docker-compose.rocks.yml`), dann `make rocks-up`. Bestehende
Installationen erneuern weiter (Credentials bleiben gültig), nur neue Registrierungen sind
blockiert.

Backups: Das Volume `acme-dns-data` enthält die Zuordnung Account ↔ Subdomain. Geht es
verloren, werden die Credentials aller bestehenden Installationen ungültig und ihre
Zertifikats-Erneuerungen schlagen fehl (Abhilfe je Installation: lokalen State löschen, neu
registrieren, neue Install-ID, neue Adresse). Deshalb in die VPS-Backup-Routine aufnehmen.

Logs: `make rocks-logs` zeigt resolver- und acme-dns-Logs mit an.
