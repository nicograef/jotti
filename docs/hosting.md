# Hosting-Anforderungen

jotti ist für temporäre Vereinsfest-Gastronomie konzipiert: 2–3 Veranstaltungen pro Jahr, wenige Tage, ehrenamtliche Helfer mit eigenen Smartphones. Die Hosting-Anforderungen sind entsprechend gering.

## Lokaler Betrieb ohne Server (Einzelgerät im WLAN)

Für das kleinste Einsatzszenario braucht jotti **weder Server noch Domain noch laufende Kosten**: eine einzige Direktverkaufs-/Selbstbedienungskasse, an **einem Gerät** von **einer Person** bedient, bar, **ohne Bondruck** — nur Eintragen und Kassieren. Das passt genau zum **Direktverkauf** („Theke"): Der Gast bestellt, zahlt sofort bar, fertig — kein offener Saldo, keine spätere Ausgabe-Bestätigung.

jotti läuft dabei auf einem vorhandenen Rechner (Windows-Laptop, Mac oder Linux-PC), und ein **Tablet oder Smartphone im selben WLAN** bedient die Kasse im Browser über die lokale IP-Adresse des Rechners.

> **Sicherheitshinweis:** Der lokale Betrieb läuft unverschlüsselt über HTTP und ist ausschließlich für das **eigene, vertrauenswürdige WLAN** gedacht. Den Rechner **niemals** per Port-Weiterleitung ins Internet öffnen.

### Voraussetzungen

- Ein Rechner mit **Docker Desktop** (Windows oder macOS) bzw. **Docker Engine + Compose-Plugin** (Linux) — Download: <https://www.docker.com/products/docker-desktop/>
- Rechner und Tablet hängen am **selben Router/WLAN**.
- Die jotti-Projektdateien liegen auf dem Rechner (ZIP entpackt oder per `git clone`).

### Schritt für Schritt

1. **`.env` anlegen.** Im Projektordner die Vorlage kopieren:

   ```bash
   cp .env.example .env
   ```

   Darin `POSTGRES_PASSWORD` und `JWT_SECRET` durch eigene Werte ersetzen. Ein sicheres Secret erzeugt z. B. `openssl rand -base64 32` (unter Windows ersatzweise eine lange, zufällige Zeichenfolge eintragen).

2. **jotti starten.** Im Projektordner:

   ```bash
   docker compose -f docker-compose.local.yml up -d --build
   ```

   Der erste Start dauert einige Minuten (die Images werden gebaut). Danach laufen Datenbank, Backend, Frontend und ein nginx-Reverse-Proxy auf **Port 80**. (Mit installiertem `make` alternativ: `make local-up`.)

3. **Lokal testen.** Auf dem Rechner `http://localhost` im Browser öffnen — die Anmeldemaske erscheint.

4. **Lokale IP-Adresse des Rechners ermitteln:**

   | System  | Befehl                   | Beispiel-Ausgabe             |
   | ------- | ------------------------ | ---------------------------- |
   | Windows | `ipconfig`               | `IPv4-Adresse: 192.168.1.50` |
   | Linux   | `hostname -I`            | `192.168.1.50`               |
   | macOS   | `ipconfig getifaddr en0` | `192.168.1.50`               |

   Gesucht ist die Adresse im Heimnetz, meist `192.168.x.x` oder `10.x.x.x`.

5. **Vom Tablet verbinden.** Tablet ins gleiche WLAN bringen und im Browser die IP-Adresse öffnen, z. B. `http://192.168.1.50`. Anmelden — fertig. Über „Zum Startbildschirm hinzufügen" lässt sich jotti wie eine App ablegen.

### Beenden

```bash
docker compose -f docker-compose.local.yml down
```

Die Daten bleiben im Docker-Volume erhalten und stehen beim nächsten Start wieder bereit. (Alternativ: `make local-down`.)

### Hinweise

- **Windows-Firewall:** Beim ersten Start fragt Windows ggf., ob der Zugriff erlaubt werden soll. Für **private Netzwerke** zulassen, damit das Tablet den Rechner über Port 80 erreicht.
- **Rechner muss laufen.** Während des Betriebs muss der Rechner eingeschaltet und im WLAN sein; Energiespar-/Ruhezustand vorher deaktivieren.
- **Hardware genügt locker.** Dieses Einzelgerät-Szenario liegt deutlich unter dem „Durchschnitt"-Szenario unten — die [minimalen Anforderungen](#minimale-anforderungen) reichen mehr als aus.
- **Für größere Feste** (mehrere Helfer, viele Tische, mehrtägig) ist ein Server-Deployment mit Domain und HTTPS sinnvoller — siehe die folgenden Abschnitte.

## Szenarien

### Durchschnittliches Vereinsfest

| Parameter               | Wert                                       |
| ----------------------- | ------------------------------------------ |
| Tische                  | 20                                         |
| Helfer (gleichzeitig)   | 10                                         |
| Betriebsdauer           | 3 Tage × 8 Stunden                         |
| Events pro Tisch/Stunde | ~5 (2 Bestellungen, 2 Ausgaben, 1 Zahlung) |
| **Events gesamt**       | **2.400**                                  |
| Datenbankgröße (Events) | ~1,2 MB                                    |
| Peak-Requests/s         | ~10                                        |

### Maximalszenario (Großes Mehrtages-Fest)

| Parameter               | Wert                                                       |
| ----------------------- | ---------------------------------------------------------- |
| Tische                  | 50                                                         |
| Helfer (gleichzeitig)   | 30                                                         |
| Betriebsdauer           | 7 Tage × 12 Stunden                                        |
| Events pro Tisch/Stunde | ~10 (4 Bestellungen, 4 Ausgaben, 1 Zahlung, 1 Stornierung) |
| **Events gesamt**       | **42.000**                                                 |
| Datenbankgröße (Events) | ~21 MB                                                     |
| Peak-Requests/s         | ~30                                                        |

## Reporting-Performance

Die Reporting-Queries aggregieren direkt über die Events-Tabelle (Conditional Aggregation auf JSONB). PostgreSQL scannt die Events einmal und berechnet alle Kennzahlen in einem Durchlauf.

| Szenario     | Events  | Geschätzte Dauer (Warm Cache) | Geschätzte Dauer (Cold Cache, SSD) |
| ------------ | ------- | ----------------------------- | ---------------------------------- |
| Durchschnitt | 2.400   | < 5 ms                        | < 50 ms                            |
| Maximum      | 42.000  | 20–80 ms                      | 100–300 ms                         |
| Extremwert   | 100.000 | 100–300 ms                    | 500 ms – 1,5 s                     |

Bei 42.000 Events passen alle Daten (~21 MB) problemlos in den PostgreSQL-Cache. Nach der ersten Abfrage sind alle folgenden Warm-Cache-Abfragen.

## VPS-Empfehlungen

### Minimale Anforderungen

Für das durchschnittliche Szenario (20 Tische, 10 Helfer, 3 Tage).

| Ressource | Minimum               |
| --------- | --------------------- |
| CPU       | 1 vCPU                |
| RAM       | 2 GB                  |
| Speicher  | 20 GB SSD             |
| Netzwerk  | 100 Mbit/s            |
| OS        | Linux (Debian/Ubuntu) |

Typische VPS-Angebote: netcup VPS 200 oder vergleichbar (~3–4 €/Monat).

### Empfohlene Anforderungen

Für das Maximalszenario (50 Tische, 30 Helfer, 7 Tage × 12h) mit komfortablem Headroom.

| Ressource | Empfohlen             |
| --------- | --------------------- |
| CPU       | 2–4 vCPUs             |
| RAM       | 4 GB                  |
| Speicher  | 40+ GB SSD            |
| Netzwerk  | 200+ Mbit/s           |
| OS        | Linux (Debian/Ubuntu) |

Typische VPS-Angebote: netcup VPS 500 oder vergleichbar (~5–8 €/Monat).

### Wichtige Hinweise

- **SSD ist Pflicht.** HDD-basierte VPS sind 5–10× langsamer bei Datenbank-Queries (Random I/O).
- **RAM ist wichtiger als CPU.** Mehr RAM = mehr PostgreSQL-Cache = schnellere Queries. CPU ist selten der Engpass, weil PostgreSQL pro Query nur 1 Core nutzt und Go-HTTP-Requests in < 20 ms abgearbeitet werden.
- **Speicherplatz ist unkritisch.** Selbst bei jahrelanger Nutzung wächst die Datenbank nur auf wenige hundert MB. Docker-Images (~500 MB) und PostgreSQL-WAL (~100 MB) sind die größten Posten.

## Docker-Compose Resource-Limits (Produktion)

Empfohlene Limits für das Maximalszenario auf einem 4 vCPU / 4 GB RAM VPS:

| Container        | RAM   | CPU  |
| ---------------- | ----- | ---- |
| PostgreSQL       | 1536M | 2.0  |
| Backend (Go)     | 384M  | 1.5  |
| Frontend (nginx) | 128M  | 0.5  |
| Reverse-proxy    | 128M  | 0.5  |
| Certbot          | 64M   | 0.25 |

CPU-Limits dürfen in Summe über die verfügbaren vCPUs gehen — Docker-CPU-Limits sind Burst-fähig, nicht exklusiv reserviert. Nicht alle Container lasten gleichzeitig aus.
