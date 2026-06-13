# Hosting-Leitfaden: jotti selbst betreiben

## 1. Das Wichtigste in 60 Sekunden

- jotti ist **self-hosted**: Es läuft nicht in einer fremden Cloud, sondern auf **eurer eigenen**
  Hardware. Gestartet wird es mit **Docker** — einem Programm, das jotti samt Datenbank in fertigen
  „Containern" startet, ohne dass ihr irgendetwas von Hand installieren müsst.
- Es gibt **zwei Wege**, je nach Größe des Fests:
  1. **Weg A — Einzelgerät im WLAN:** jotti läuft auf **einem Laptop/PC**, bedient von **einem
     Tablet oder Handy** im selben WLAN. **Kein Server, keine Domain, keine laufenden Kosten.**
     Ideal für eine einzelne Theke (Direktverkauf, bar, ohne Bondruck).
  2. **Weg B — Eigener Server:** jotti läuft auf einem kleinen gemieteten Server (einem „VPS") mit
     eigener Internet-Adresse (Domain) und Verschlüsselung (HTTPS). Für **größere Feste** mit
     mehreren Helfern, vielen Tischen und über mehrere Tage.
- **Was kostet das?** jotti selbst ist für berechtigte Vereine **kostenlos**. Weg A verursacht
  sonst keine laufenden Kosten (nur vorhandene Hardware); bei Weg B kommt ein kleiner VPS ab etwa
  **3–8 € pro Monat** hinzu (ggf. plus Cloud-TSE-Gebühr, siehe Betreiber-Leitfaden).
- **Technisches Vorwissen?** Grundkenntnisse im Umgang mit der Kommandozeile genügen — alle
  nötigen Befehle stehen in diesem Leitfaden zum Kopieren bereit.
- **Worum es hier _nicht_ geht:** die rechtlichen Pflichten rund um die Kasse (TSE, Finanzamt,
  10 Jahre Aufbewahrung). Das steht im [Betreiber-Leitfaden](./leitfaden-betreiber.md).

---

## 2. Welcher Weg passt zu uns?

| Frage                 | Weg A: Einzelgerät im WLAN               | Weg B: Eigener Server                   |
| --------------------- | ---------------------------------------- | --------------------------------------- |
| Typisches Fest        | Eine Theke, eine Person kassiert         | Mehrere Helfer, viele Tische, mehrtägig |
| Geräte                | 1 Laptop + 1 Tablet/Handy                | Beliebig viele Handys der Helfer        |
| Internet nötig?       | Nein — nur lokales WLAN                  | Ja                                      |
| Domain & HTTPS nötig? | Domain: nein, HTTPS: ja (echtes Zertifikat + Fallback) | Ja                         |
| Bondruck              | Nein                                     | Möglich _(in Entwicklung)_              |
| Laufende Kosten       | Keine                                    | ~3–8 €/Monat (VPS)                      |
| Einrichtungsaufwand   | Sehr gering                              | Etwas höher (einmalig)                  |

> **Faustregel:** Eine kleine Theke an einem Tag? → **Weg A.** Ein richtiges Fest, bei dem
> Servicekräfte mit ihren eigenen Handys an den Tischen bestellen? → **Weg B.**

---

## 3. Weg A: Einzelgerät im WLAN (ohne Server)

Für das kleinste Einsatzszenario braucht jotti **weder Server noch Domain noch laufende Kosten**:
eine einzige Direktverkaufs-Kasse („Theke"), an **einem Gerät** von **einer Person** bedient, bar,
**ohne Bondruck** — nur Eintragen und Kassieren. Der Gast bestellt, zahlt sofort bar, fertig: kein
offener Saldo, keine spätere Ausgabe-Bestätigung.

jotti läuft dabei auf einem vorhandenen Rechner (Windows-Laptop, Mac oder Linux-PC). Ein **Tablet
oder Smartphone im selben WLAN** bedient die Kasse im Browser über die lokale Adresse des Rechners.

> 🔒 **Grünes Schloss als Normalfall:** Der lokale Betrieb nutzt ein **echtes, vom Browser
> anerkanntes Zertifikat** (grünes Schloss, **keine** Warnung) über die Adresse
> `…lokal.jotti.rocks`. jotti holt es beim ersten Start mit Internet automatisch und erneuert es
> selbst — **kein** CA-Rollout, **keine** Einrichtung pro Gerät. Welche Adresse gerade gilt und ein
> QR-Code dazu stehen auf der **Status-Seite** `http://localhost:8484` am Kassenrechner.
>
> Zusätzlich gibt es einen **Fallback** `https://<LAN-IP>` mit selbstsigniertem Zertifikat (einmalige
> Browserwarnung pro Gerät) — er greift vor der ersten Ausstellung, ohne Internet oder wenn der Router
> die grüne Adresse blockiert ([DNS-Rebind-Schutz](./dns-rebind-schutz.md)).
>
> **Restrisiko nur beim Fallback:** Über die grüne Adresse scheitert ein aktiver Angreifer im WLAN
> (MITM) hart. Nur solange ihr den selbstsignierten Fallback nutzt, bleibt — wie bei reinem
> selbstsigniertem Betrieb — ein aktiver MITM möglich. Hintergrund:
> `docs/prds/prd-lokale-tls-vertrauenswuerdig.md`.
>
> ℹ️ **Einzeltheke-/localhost-Ausnahme:** Wenn jotti nur direkt am selben Rechner über `localhost`
> bedient wird, entsteht kein WLAN-Transport.

### Voraussetzungen

- Ein Rechner mit **Docker Desktop** (Windows oder macOS) bzw. **Docker Engine + Compose-Plugin**
  (Linux) — Download: <https://www.docker.com/products/docker-desktop/>
- Rechner und Tablet hängen am **selben Router/WLAN**.
- Die jotti-Projektdateien liegen auf dem Rechner (ZIP entpackt oder per `git clone`).

### Schritt für Schritt

1. **`.env` anlegen.** Im Projektordner ausführen:

   ```bash
   make init
   ```

Das Kommando erzeugt eine vollständige `.env` mit sicheren Zufallswerten für
`POSTGRES_PASSWORD`, `JWT_SECRET` und `RELAY_AUTH_TOKEN`.

2. **jotti starten.** Im Projektordner:

   ```bash
   docker compose -f docker-compose.local.yml up -d --build
   ```

Der erste Start dauert einige Minuten (die „Container" werden gebaut). Danach laufen Datenbank,
Backend, Frontend und ein Caddy-Reverse-Proxy auf **Port 443** (Port 80 leitet auf HTTPS um); Caddy
holt und erneuert das vertrauenswürdige Zertifikat selbst. Mit installiertem `make` alternativ:
`make local-up` — es nennt am Ende die Status-URL.

3. **Status-Seite öffnen.** Am Kassenrechner `http://localhost:8484` im Browser öffnen. Die
   Status-Seite zeigt den aktuellen Zustand, die **Zugangsadresse(n)** und einen **QR-Code**. Direkt
   nach dem ersten Start steht dort zunächst die Fallback-Adresse; sobald das Zertifikat ausgestellt
   ist (wenige Sekunden bis Minuten, braucht Internet), wechselt sie automatisch auf die **grüne
   Adresse**.

4. **Vom Smartphone/Tablet verbinden.** Das Gerät ins **Vereins-WLAN** bringen (nicht Mobilfunk,
   **kein Gastnetz** — die [DNS-Rebind-Anleitung](./dns-rebind-schutz.md) erklärt, warum). Dann den
   **QR-Code** von der Status-Seite scannen oder die grüne Adresse eintippen → **grünes Schloss, keine
   Warnung**, anmelden. Über „Zum Startbildschirm hinzufügen" lässt sich jotti wie eine App ablegen.

5. **Falls die grüne Adresse (noch) nicht geht.** Die Status-Seite nennt dann die **Fallback-Adresse**
   `https://<LAN-IP>` (z. B. `https://192.168.1.50`) — beim ersten Zugriff pro Gerät die einmalige
   Browserwarnung bestätigen, danach anmelden. Lädt die grüne Adresse auf den Handys gar nicht,
   blockiert vermutlich der Router: dann die [DNS-Rebind-Anleitung](./dns-rebind-schutz.md) befolgen
   (die Status-Seite verlinkt sie ebenfalls).

### Beenden

```bash
docker compose -f docker-compose.local.yml down
```

Die Daten bleiben im Docker-Volume erhalten und stehen beim nächsten Start wieder bereit.
Alternativ: `make local-down`.

### Gut zu wissen

- **Windows-Firewall:** Beim ersten Start fragt Windows ggf., ob der Zugriff erlaubt werden soll.
  Für **private Netzwerke** zulassen, damit das Tablet den Rechner über Port 443 erreicht
  (Port 80 wird nur für Redirect genutzt).
- **Rechner muss laufen.** Während des Betriebs muss der Rechner eingeschaltet und im WLAN sein;
  Energiespar-/Ruhezustand vorher deaktivieren.
- **Erster Start braucht Internet.** Für das vertrauenswürdige Zertifikat braucht der erste Start
  (und der Start nach längerer Pause, wenn das Zertifikat abgelaufen ist) **Internet** — am besten
  vorab zuhause testen. Ohne Internet läuft der Verkauf über den Fallback weiter.
- **Stabile Adresse (optional).** Bekommt der Rechner per DHCP eine neue LAN-IP, ändert sich der
  Hostname — aber **dasselbe** Zertifikat gilt weiter (keine neue Warnung) und die Status-Seite zeigt
  die neue Adresse. Wer eine feste Adresse möchte, richtet im Router eine **DHCP-Reservierung** für
  den Kassenrechner ein.
- **Hardware genügt locker.** Jeder halbwegs aktuelle Laptop bewältigt eine Theke mühelos.
- **Für größere Feste** (mehrere Helfer, viele Tische, mehrtägig) ist Weg B mit Domain und HTTPS
  die bessere Wahl.

---

## 4. Weg B: Eigener Server (für größere Feste)

Für größere Feste — mehrere Helfer mit eigenen Handys, viele Tische, mehrere Tage — lohnt sich ein
eigener kleiner Server im Internet, ein **VPS** („Virtual Private Server"). Dann erreichen **alle
Helfer** jotti über eine **Internet-Adresse (Domain)** mit **Verschlüsselung (HTTPS)** — nicht nur
im lokalen WLAN vor Ort.

### Welchen Server mieten?

jotti ist **sehr genügsam** — schon der kleinste VPS reicht für ein durchschnittliches Vereinsfest.
Zwei Richtwerte:

**Minimal** — für ein durchschnittliches Fest (etwa 20 Tische, 10 Helfer gleichzeitig, 2–3 Tage):

| Ressource      | Minimum               |
| -------------- | --------------------- |
| CPU            | 1 vCPU                |
| RAM            | 2 GB                  |
| Speicher       | 20 GB SSD             |
| Netzwerk       | 100 Mbit/s            |
| Betriebssystem | Linux (Debian/Ubuntu) |

Typisches Angebot: netcup VPS 200 oder vergleichbar (~3–4 €/Monat).

**Empfohlen** — für große, mehrtägige Feste (etwa 50 Tische, 30 Helfer gleichzeitig) mit Reserve:

| Ressource      | Empfohlen             |
| -------------- | --------------------- |
| CPU            | 2–4 vCPUs             |
| RAM            | 4 GB                  |
| Speicher       | 40+ GB SSD            |
| Netzwerk       | 200+ Mbit/s           |
| Betriebssystem | Linux (Debian/Ubuntu) |

Typisches Angebot: netcup VPS 500 oder vergleichbar (~5–8 €/Monat).

### Worauf es bei der Hardware ankommt

- **SSD ist Pflicht.** Server mit alter Festplatte (HDD) sind um ein Vielfaches langsamer bei
  Datenbankzugriffen.
- **RAM ist wichtiger als CPU.** Mehr Arbeitsspeicher bedeutet schnellere Auswertungen. Die
  Rechenleistung (CPU) ist selten der Engpass.
- **Speicherplatz ist unkritisch.** Selbst nach Jahren wächst die Datenbank nur auf wenige hundert
  Megabyte.

### Domain und Verschlüsselung (HTTPS)

Für den Server-Betrieb braucht ihr zusätzlich zwei Dinge:

- eine **Domain** (z. B. `kasse-musterverein.de`), die auf euren Server zeigt, und
- ein **TLS-Zertifikat** für HTTPS — kostenlos über **Let's Encrypt**.

jotti bringt dafür eine fertige Produktions-Konfiguration mit (`docker-compose.prod.yml`),
inklusive nginx-Reverse-Proxy und automatischer Zertifikats-Erneuerung. Die einmalige
Ersteinrichtung (erstes Zertifikat anfordern, Stack starten) übernimmt das Skript
`scripts/prod-init.sh` — Domain und E-Mail-Adresse darin vorher anpassen.

> ⚠️ Ohne HTTPS dürft ihr jotti **nicht** über das offene Internet betreiben: Anmeldedaten und
> Bestellungen würden sonst unverschlüsselt übertragen.

> 💾 **Backups nicht vergessen.** jotti speichert alle Daten in einem **Docker-Volume**. Macht
> **regelmäßige Backups** der Datenbank — besonders wegen der gesetzlichen 10-Jahre-Aufbewahrung
> (Details im [Betreiber-Leitfaden, Schritt 4](./leitfaden-betreiber.md#schritt-4-daten-10-jahre-aufbewahren)).

---

## 5. Glossar

| Begriff                     | Einfache Erklärung                                                                         |
| --------------------------- | ------------------------------------------------------------------------------------------ |
| **Self-hosted**             | Ihr betreibt die Software auf **eurer eigenen** Hardware, nicht in einer fremden Cloud.    |
| **Docker**                  | Programm, das jotti samt Datenbank in fertigen „Containern" startet — ohne Installieren.   |
| **Container**               | Ein abgeschlossenes Paket mit einem laufenden Programmteil (Datenbank, Backend, Frontend). |
| **VPS**                     | „Virtual Private Server" — ein kleiner, gemieteter Server im Internet.                     |
| **WLAN / LAN**              | Euer lokales (Funk-)Netzwerk vor Ort.                                                      |
| **IP-Adresse**              | Die „Hausnummer" eines Geräts im Netzwerk, z. B. `192.168.1.50`.                           |
| **Domain**                  | Eine Internet-Adresse wie `kasse-musterverein.de`.                                         |
| **HTTPS / TLS**             | Verschlüsselte Verbindung im Internet (Schloss-Symbol im Browser).                         |
| **Let's Encrypt**           | Kostenlose Stelle, die HTTPS-Zertifikate ausstellt.                                        |
| **`.env`**                  | Konfigurationsdatei mit Passwörtern und Geheimnissen — niemals öffentlich teilen.          |
| **Reverse Proxy (nginx / Caddy)** | Vermittler, der Anfragen sicher an jotti weiterleitet (lokal: Caddy, Server: nginx).  |
| **Docker-Volume**           | Der Speicherort, an dem Docker eure Daten dauerhaft aufbewahrt.                            |
| **SSD**                     | Schnelle Festplatte (Flash-Speicher) — Pflicht für flüssigen Betrieb.                      |
| **Direktverkauf („Theke")** | Verkauf, bei dem sofort bar kassiert wird — kein offener Tisch-Saldo.                      |
