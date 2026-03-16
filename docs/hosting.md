# Hosting-Anforderungen

jotti ist für temporäre Vereinsfest-Gastronomie konzipiert: 2–3 Veranstaltungen pro Jahr, wenige Tage, ehrenamtliche Helfer mit eigenen Smartphones. Die Hosting-Anforderungen sind entsprechend gering.

## Szenarien

### Durchschnittliches Vereinsfest

| Parameter              | Wert                        |
| ---------------------- | --------------------------- |
| Tische                 | 20                          |
| Helfer (gleichzeitig)  | 10                          |
| Betriebsdauer          | 3 Tage × 8 Stunden         |
| Events pro Tisch/Stunde| ~5 (2 Bestellungen, 2 Ausgaben, 1 Zahlung) |
| **Events gesamt**      | **2.400**                   |
| Datenbankgröße (Events)| ~1,2 MB                     |
| Peak-Requests/s        | ~10                         |

### Maximalszenario (Großes Mehrtages-Fest)

| Parameter              | Wert                        |
| ---------------------- | --------------------------- |
| Tische                 | 50                          |
| Helfer (gleichzeitig)  | 30                          |
| Betriebsdauer          | 7 Tage × 12 Stunden        |
| Events pro Tisch/Stunde| ~10 (4 Bestellungen, 4 Ausgaben, 1 Zahlung, 1 Stornierung) |
| **Events gesamt**      | **42.000**                  |
| Datenbankgröße (Events)| ~21 MB                      |
| Peak-Requests/s        | ~30                         |

## Reporting-Performance

Die Reporting-Queries aggregieren direkt über die Events-Tabelle (Conditional Aggregation auf JSONB). PostgreSQL scannt die Events einmal und berechnet alle Kennzahlen in einem Durchlauf.

| Szenario       | Events  | Geschätzte Dauer (Warm Cache) | Geschätzte Dauer (Cold Cache, SSD) |
| -------------- | ------- | ----------------------------- | ---------------------------------- |
| Durchschnitt   | 2.400   | < 5 ms                       | < 50 ms                           |
| Maximum        | 42.000  | 20–80 ms                     | 100–300 ms                        |
| Extremwert     | 100.000 | 100–300 ms                   | 500 ms – 1,5 s                    |

Bei 42.000 Events passen alle Daten (~21 MB) problemlos in den PostgreSQL-Cache. Nach der ersten Abfrage sind alle folgenden Warm-Cache-Abfragen.

## VPS-Empfehlungen

### Minimale Anforderungen

Für das durchschnittliche Szenario (20 Tische, 10 Helfer, 3 Tage).

| Ressource   | Minimum       |
| ----------- | ------------- |
| CPU         | 1 vCPU        |
| RAM         | 2 GB          |
| Speicher    | 20 GB SSD     |
| Netzwerk    | 100 Mbit/s    |
| OS          | Linux (Debian/Ubuntu) |

Typische VPS-Angebote: netcup VPS 200 oder vergleichbar (~3–4 €/Monat).

### Empfohlene Anforderungen

Für das Maximalszenario (50 Tische, 30 Helfer, 7 Tage × 12h) mit komfortablem Headroom.

| Ressource   | Empfohlen     |
| ----------- | ------------- |
| CPU         | 2–4 vCPUs     |
| RAM         | 4 GB          |
| Speicher    | 40+ GB SSD    |
| Netzwerk    | 200+ Mbit/s   |
| OS          | Linux (Debian/Ubuntu) |

Typische VPS-Angebote: netcup VPS 500 oder vergleichbar (~5–8 €/Monat).

### Wichtige Hinweise

- **SSD ist Pflicht.** HDD-basierte VPS sind 5–10× langsamer bei Datenbank-Queries (Random I/O).
- **RAM ist wichtiger als CPU.** Mehr RAM = mehr PostgreSQL-Cache = schnellere Queries. CPU ist selten der Engpass, weil PostgreSQL pro Query nur 1 Core nutzt und Go-HTTP-Requests in < 20 ms abgearbeitet werden.
- **Speicherplatz ist unkritisch.** Selbst bei jahrelanger Nutzung wächst die Datenbank nur auf wenige hundert MB. Docker-Images (~500 MB) und PostgreSQL-WAL (~100 MB) sind die größten Posten.

## Docker-Compose Resource-Limits (Produktion)

Empfohlene Limits für das Maximalszenario auf einem 4 vCPU / 4 GB RAM VPS:

| Container      | RAM    | CPU   |
| -------------- | ------ | ----- |
| PostgreSQL     | 1536M  | 2.0   |
| Backend (Go)   | 384M   | 1.5   |
| Frontend (nginx)| 128M  | 0.5   |
| Reverse-proxy  | 128M   | 0.5   |
| Certbot        | 64M    | 0.25  |

CPU-Limits dürfen in Summe über die verfügbaren vCPUs gehen — Docker-CPU-Limits sind Burst-fähig, nicht exklusiv reserviert. Nicht alle Container lasten gleichzeitig aus.
