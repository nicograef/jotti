# Hosting-Leitfaden: jotti selbst betreiben

## Inhalt

1. [Das Wichtigste in 60 Sekunden](#1-das-wichtigste-in-60-sekunden)
2. [Welcher Weg passt zu uns?](#2-welcher-weg-passt-zu-uns)
3. [Weg A: Einzelgerät im WLAN (ohne Server)](#3-weg-a-einzelgerät-im-wlan-ohne-server)
   - [Voraussetzungen](#voraussetzungen)
   - [Schritt für Schritt](#schritt-für-schritt)
   - [Beenden](#beenden)
   - [Gut zu wissen](#gut-zu-wissen)
4. [Weg B: Eigener Server (für größere Feste)](#4-weg-b-eigener-server-für-größere-feste)
   - [Welchen Server mieten?](#welchen-server-mieten)
   - [Worauf es bei der Hardware ankommt](#worauf-es-bei-der-hardware-ankommt)
   - [Domain und Verschlüsselung (HTTPS)](#domain-und-verschlüsselung-https)
5. [Häufige Fragen (FAQ)](#5-häufige-fragen-faq)
6. [Glossar](#6-glossar)

---

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
- **Was kostet das?** Weg A: nichts außer vorhandener Hardware. Weg B: ein kleiner VPS ab etwa
  **3–8 € pro Monat**.
- **Worum es hier _nicht_ geht:** die rechtlichen Pflichten rund um die Kasse (TSE, Finanzamt,
  10 Jahre Aufbewahrung). Das steht im [Betreiber-Leitfaden](./leitfaden-betreiber.md).

---

## 2. Welcher Weg passt zu uns?

| Frage                 | Weg A: Einzelgerät im WLAN       | Weg B: Eigener Server                   |
| --------------------- | -------------------------------- | --------------------------------------- |
| Typisches Fest        | Eine Theke, eine Person kassiert | Mehrere Helfer, viele Tische, mehrtägig |
| Geräte                | 1 Laptop + 1 Tablet/Handy        | Beliebig viele Handys der Helfer        |
| Internet nötig?       | Nein — nur lokales WLAN          | Ja                                      |
| Domain & HTTPS nötig? | Nein                             | Ja                                      |
| Bondruck              | Nein                             | Möglich _(in Entwicklung)_              |
| Laufende Kosten       | Keine                            | ~3–8 €/Monat (VPS)                      |
| Einrichtungsaufwand   | Sehr gering                      | Etwas höher (einmalig)                  |

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

> 🔒 **Sicherheitshinweis:** Der lokale Betrieb läuft **unverschlüsselt** über HTTP und ist
> ausschließlich für das **eigene, vertrauenswürdige WLAN** gedacht. Öffnet den Rechner **niemals**
> per Port-Weiterleitung ins Internet.

### Voraussetzungen

- Ein Rechner mit **Docker Desktop** (Windows oder macOS) bzw. **Docker Engine + Compose-Plugin**
  (Linux) — Download: <https://www.docker.com/products/docker-desktop/>
- Rechner und Tablet hängen am **selben Router/WLAN**.
- Die jotti-Projektdateien liegen auf dem Rechner (ZIP entpackt oder per `git clone`).

### Schritt für Schritt

1. **`.env` anlegen.** Im Projektordner ausfuehren:

   ```bash
   make init
   ```

Das Kommando erzeugt eine vollstaendige `.env` mit sicheren Zufallswerten fuer
`POSTGRES_PASSWORD`, `JWT_SECRET` und `RELAY_AUTH_TOKEN`.

2. **jotti starten.** Im Projektordner:

   ```bash
   docker compose -f docker-compose.local.yml up -d --build
   ```

   Der erste Start dauert einige Minuten (die „Container" werden gebaut). Danach laufen Datenbank,
   Backend, Frontend und ein nginx-Reverse-Proxy auf **Port 80**. Mit installiertem `make`
   alternativ: `make local-up`.

3. **Lokal testen.** Auf dem Rechner `http://localhost` im Browser öffnen — die Anmeldemaske
   erscheint.

4. **Lokale IP-Adresse des Rechners ermitteln:**

   | System  | Befehl                   | Beispiel-Ausgabe             |
   | ------- | ------------------------ | ---------------------------- |
   | Windows | `ipconfig`               | `IPv4-Adresse: 192.168.1.50` |
   | Linux   | `hostname -I`            | `192.168.1.50`               |
   | macOS   | `ipconfig getifaddr en0` | `192.168.1.50`               |

   Gesucht ist die Adresse im Heimnetz, meist `192.168.x.x` oder `10.x.x.x`.

5. **Vom Tablet verbinden.** Tablet ins gleiche WLAN bringen und im Browser die IP-Adresse öffnen,
   z. B. `http://192.168.1.50`. Anmelden — fertig. Über „Zum Startbildschirm hinzufügen" lässt sich
   jotti wie eine App ablegen.

### Beenden

```bash
docker compose -f docker-compose.local.yml down
```

Die Daten bleiben im Docker-Volume erhalten und stehen beim nächsten Start wieder bereit.
Alternativ: `make local-down`.

### Gut zu wissen

- **Windows-Firewall:** Beim ersten Start fragt Windows ggf., ob der Zugriff erlaubt werden soll.
  Für **private Netzwerke** zulassen, damit das Tablet den Rechner über Port 80 erreicht.
- **Rechner muss laufen.** Während des Betriebs muss der Rechner eingeschaltet und im WLAN sein;
  Energiespar-/Ruhezustand vorher deaktivieren.
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

---

## 5. Häufige Fragen (FAQ)

**Brauchen wir Internet auf dem Fest?**
Bei Weg A nein — es genügt ein lokales WLAN, in dem Rechner und Tablet hängen. Bei Weg B ja, denn
die Handys der Helfer erreichen jotti über das Internet.

**Können mehrere Helfer gleichzeitig kassieren?**
Bei Weg A bedient **ein** Gerät die Theke. Sobald mehrere Servicekräfte mit eigenen Handys an
Tischen aufnehmen sollen, ist Weg B die richtige Wahl — dort arbeiten beliebig viele Geräte
gleichzeitig.

**Müssen wir uns mit Linux auskennen?**
Grundkenntnisse im Umgang mit der Kommandozeile reichen. Alle nötigen Befehle stehen in diesem
Leitfaden zum Kopieren bereit.

**Wie sichern wir unsere Daten?**
jotti speichert alles in einer Datenbank (in einem „Docker-Volume"). Macht **regelmäßige Backups**
— besonders wichtig wegen der gesetzlichen 10-Jahre-Aufbewahrung. Details dazu im
[Betreiber-Leitfaden, Schritt 4](./leitfaden-betreiber.md#schritt-4-daten-10-jahre-aufbewahren).

**Was kostet uns der Betrieb?**
jotti selbst ist für berechtigte Vereine kostenlos. Weg A verursacht keine laufenden Kosten; bei
Weg B fällt nur die VPS-Miete an (ca. 3–8 €/Monat), ggf. plus die Cloud-TSE-Gebühr (siehe
Betreiber-Leitfaden).

---

## 6. Glossar

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
| **Reverse Proxy (nginx)**   | Vermittler, der Anfragen aus dem Internet sicher an jotti weiterleitet.                    |
| **Docker-Volume**           | Der Speicherort, an dem Docker eure Daten dauerhaft aufbewahrt.                            |
| **SSD**                     | Schnelle Festplatte (Flash-Speicher) — Pflicht für flüssigen Betrieb.                      |
| **Direktverkauf („Theke")** | Verkauf, bei dem sofort bar kassiert wird — kein offener Tisch-Saldo.                      |
