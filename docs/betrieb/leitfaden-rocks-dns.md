# Betriebsleitfaden: DNS-Infrastruktur für vertrauenswürdiges lokales TLS

Dieser Leitfaden richtet sich an den Betreiber des jotti.rocks-VPS (nicht an
Vereine). Er beschreibt Einrichtung, Verifikation und laufenden Betrieb der
DNS-Infrastruktur, mit der lokale jotti-Installationen echte
Let's-Encrypt-Zertifikate für `*.<install-id>.lokal.jotti.rocks` beziehen.

## 1. Überblick

Zwei Services im rocks-Stack (`docker-compose.rocks.yml`):

| Service    | Zone                | Erreichbarkeit                                          |
| ---------- | ------------------- | ------------------------------------------------------- |
| `resolver` | `lokal.jotti.rocks` | öffentlich, Port 53 UDP+TCP (einziger Prozess auf :53)   |
| `acme-dns` | `auth.jotti.rocks`  | nur Docker-intern (DNS via resolver, API via nginx)      |

- Der resolver beantwortet A-Records und `_acme-challenge`-CNAMEs rein
  rechnerisch aus dem angefragten Namen (zustandslos, TTL 86400, das Mapping
  Name → IP ist unveränderlich). Anfragen für `auth.jotti.rocks` leitet er
  Docker-intern an acme-dns weiter.
- acme-dns verwaltet die TXT-Records für DNS-01-Challenges. Die HTTP-API
  (`/register`, `/update`, `/health`) läuft hinter nginx unter
  `https://auth.jotti.rocks`; `/register` ist dort streng rate-limitiert
  (1 Anfrage/Minute je IP). Zustand (SQLite) liegt im Volume `acme-dns-data`.

## 2. Voraussetzungen

1. **Port 53 ist frei.** Auf dem VPS prüfen, dass nichts öffentlich auf :53
   lauscht (systemd-resolved bindet nur `127.0.0.53`, unkritisch, Docker
   publiziert auf `0.0.0.0`):

   ```bash
   sudo ss -lnup 'sport = :53'
   sudo ss -lntp 'sport = :53'
   ```

2. Provider-Firewall (z. B. Hetzner Cloud Firewall) erlaubt eingehend
   53/UDP und 53/TCP.

3. DNS-Hoster unterstützt NS-Records für Subdomain-Delegation (Standard).

## 3. DNS-Hoster-Einträge (einmalig, manuell)

Beim DNS-Hoster der Zone `jotti.rocks` anlegen:

| Typ | Name                | Wert                 | Zweck                                  |
| --- | ------------------- | -------------------- | --------------------------------------- |
| A   | `dns.jotti.rocks`   | `<VPS-IP>`           | Adresse des eigenen Nameservers         |
| NS  | `lokal.jotti.rocks` | `dns.jotti.rocks`    | Delegation an den resolver              |
| NS  | `auth.jotti.rocks`  | `dns.jotti.rocks`    | Delegation an acme-dns (via resolver)   |
| CAA | `jotti.rocks`       | `0 issue "letsencrypt.org"` | nur Let's Encrypt darf ausstellen (Defense-in-Depth) |

> Der A-Record für `auth.jotti.rocks` selbst kommt aus acme-dns
> (konfiguriert in `docker-compose.rocks.yml`) und zeigt auf den VPS, beim
> Hoster ist dafür nichts einzutragen.

## 4. Deployment

### .env ergänzen

Auf dem VPS in der `.env` die öffentliche IPv4 des Servers eintragen (wird vom
resolver als NS-A-Record und von acme-dns als Zone-Apex-A-Record serviert):

```bash
VPS_PUBLIC_IP=<öffentliche IPv4 des VPS>
```

### Auf dem bestehenden Deployment nachrüsten

```bash
git pull
make rocks-up
```

Damit laufen resolver und acme-dns; `auth.jotti.rocks` ist aber noch nicht im
Zertifikat. Sobald die DNS-Hoster-Einträge (Abschnitt 3) aktiv sind, das
Zertifikat erweitern und nginx neu laden:

```bash
docker compose -f docker-compose.rocks.yml \
  run --rm --entrypoint certbot certbot certonly \
  --webroot -w /var/www/certbot \
  -d jotti.rocks -d www.jotti.rocks -d demo.jotti.rocks -d auth.jotti.rocks \
  --email graef.nico@gmail.com --agree-tos --no-eff-email --expand

docker compose -f docker-compose.rocks.yml \
  exec reverse-proxy nginx -s reload
```

### Ersteinrichtung (frischer VPS)

`./scripts/rocks-init.sh` wie gehabt. Hinweis: `auth.jotti.rocks` löst erst
auf, wenn der Stack läuft und die Delegation gesetzt ist, bei der
Ersteinrichtung wird die Domain deshalb übersprungen (Warnung im Skript) und
das Zertifikat anschließend wie oben erweitert.

## 5. End-to-End-Verifikation (nach jedem Infra-Setup)

Von außerhalb des VPS (z. B. vom eigenen Rechner):

1. **Delegation + berechneter A-Record:**

   ```bash
   dig +short NS lokal.jotti.rocks        # → dns.jotti.rocks.
   dig +short 10-0-0-1.test.lokal.jotti.rocks A   # → 10.0.0.1
   ```

2. **Berechneter Challenge-CNAME:**

   ```bash
   dig +short CNAME _acme-challenge.test.lokal.jotti.rocks
   # → test.auth.jotti.rocks.
   ```

3. **Registrierung** (liefert `username`, `password`, `subdomain`,
   `fulldomain`, für die folgenden Schritte aufheben):

   ```bash
   curl -s -X POST https://auth.jotti.rocks/register
   ```

   Rate-Limit prüfen: mehrere schnelle Aufrufe hintereinander → HTTP 429.

4. **TXT-Update nur mit Credentials** (der TXT-Wert muss exakt 43 Zeichen
   lang sein):

   ```bash
   curl -s -X POST https://auth.jotti.rocks/update \
     -H "X-Api-User: <username>" -H "X-Api-Key: <password>" \
     -d '{"subdomain": "<subdomain>", "txt": "0123456789012345678901234567890123456789012"}'

   dig +short TXT <subdomain>.auth.jotti.rocks   # → der gesetzte Wert
   ```

   Gegenprobe: derselbe Aufruf ohne/mit falschen Headern muss abgelehnt werden.

5. **CAA-Record:**

   ```bash
   dig +short CAA jotti.rocks   # → 0 issue "letsencrypt.org"
   ```

6. **Staging-Wildcard-Zertifikat:** der Beweis, dass der komplette
   DNS-01-Pfad steht (Let's-Encrypt-Staging, um Rate-Limits auf die echte
   Zone zu vermeiden):

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

   Erwartung: `Cert success.` Das Staging-Zertifikat wird verworfen, es geht
   nur um den Nachweis.

## 6. Laufender Betrieb

- **CT-Log-Monitoring (monatliche Routine):**
  <https://crt.sh/?q=%25.lokal.jotti.rocks> aufrufen und das
  Ausstellungsvolumen prüfen. Erwartung: einzelne Wildcard-Zertifikate je
  Install-ID, Volumen in der Größenordnung der bekannten Installationen.
  Auffällige Spitzen (Massen-Registrierungen) sind ein Missbrauchssignal.
- **Eskalationspfad bei Missbrauch:** Registrierung schließen:
  `disable_registration = true` in der acme-dns-Config
  (`docker-compose.rocks.yml`) setzen und `make rocks-up`. Bestehende
  Installationen erneuern weiter (Credentials bleiben gültig); nur neue
  Registrierungen sind blockiert. Ein dauerhaftes Registrierungs-Gate ist
  eine eigene Produktentscheidung (siehe Plan, „Resolved decisions").
- **Backups:** Das Volume `acme-dns-data` enthält die Zuordnung
  Account ↔ Subdomain. Geht es verloren, werden die Credentials aller
  bestehenden Installationen ungültig und deren Zertifikats-Erneuerungen
  schlagen fehl (Abhilfe je Installation: lokalen State löschen und neu
  registrieren, neue Install-ID, neue Adresse). Deshalb in die bestehende
  VPS-Backup-Routine aufnehmen.
- **Logs:** `make rocks-logs` zeigt resolver- und acme-dns-Logs mit an.
