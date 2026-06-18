---
title: Eigener Server (Ersteinrichtung)
description: 'Experten-Weg: jotti auf einem eigenen VPS mit Domain und HTTPS betreiben. Voraussetzungen und Ersteinrichtung Schritt für Schritt.'
---

> Dieser Abschnitt ist nur für den Betrieb über das Internet gedacht und setzt
> Grundkenntnisse mit Linux und Kommandozeile voraus. Wer alle Helfer beim Fest vor
> Ort im selben WLAN hat, bleibt beim [Standardweg](installation.md).

Wer jotti auch außerhalb des lokalen WLANs erreichen will (über das Internet, an
mehreren Standorten oder ohne einen Rechner vor Ort), betreibt es auf einem eigenen
kleinen Server, einem VPS. Dann erreichen alle Helfer jotti über eine
Internet-Adresse (Domain) mit Verschlüsselung (HTTPS).

jotti ist genügsam: Schon der kleinste VPS (1 vCPU, 2 GB RAM, 20 GB SSD, Linux)
reicht für ein durchschnittliches Vereinsfest. Typisches Angebot: netcup VPS 200
oder vergleichbar (circa 5€/Monat). Zusätzlich braucht ihr eine Domain, die per
DNS auf den Server zeigt, sowie ein TLS-Zertifikat. Das Zertifikat holt jotti
automatisch: Die Produktions-Konfiguration (`docker-compose.prod.yml`) bringt einen
Caddy-Reverse-Proxy mit, der es beim ersten Start selbst bei Let's Encrypt
anfordert und danach erneuert.

> ⚠️ Ohne HTTPS dürft ihr jotti nicht über das offene Internet betreiben:
> Anmeldedaten und Bestellungen würden sonst unverschlüsselt übertragen.

## Ersteinrichtung

1. **Docker installieren.** Auf dem Server Docker Engine samt Compose-Plugin
   einrichten (<https://docs.docker.com/engine/install/>).
2. **Projektdateien holen.** Das aktuelle Release als ZIP entpacken oder das
   Repository klonen, dann in den Projektordner wechseln.
3. **`.env` anlegen** mit `make init` (erzeugt sichere Zufallswerte für die
   Geheimnisse).
4. **Domain, E-Mail und Version eintragen** in der `.env`:

   ```bash
   JOTTI_DOMAIN=kasse-musterverein.de
   LETSENCRYPT_EMAIL=vorstand@musterverein.de
   JOTTI_VERSION=v0.2.0
   ```

5. **DNS auf den Server zeigen lassen.** Beim Domain-Anbieter einen A-Record (und
   bei IPv6 einen AAAA-Record) auf die öffentliche IP des Servers setzen. Erst wenn
   die Domain auf den Server zeigt, kann Let's Encrypt ein Zertifikat ausstellen.
6. **Stack starten** mit `make prod-init`. Das Skript prüft Docker und die
   DNS-Auflösung, zieht die gepinnten Images, startet den Stack und wartet, bis
   Backend und HTTPS gesund antworten.

Danach ist jotti unter `https://<eure-domain>` erreichbar; HTTP leitet automatisch
auf HTTPS um.
