# jotti — Das kostenlose Kassensystem für Vereinsfeste.

> [!NOTE]
> **Beta:** jotti 1.0 wird gerade noch getestet und geprüft.

Ein kostenloses **Gastronomie-Kassensystem (mPOS)** mit einsehbarem Quellcode (Source-Available) für Vereine und gemeinnützige Organisationen — Vereinsfeste, Weihnachtsmärkte, Konzerte, Maihocks, Sommerfeste.

Servicekräfte nehmen auf ihren eigenen Smartphones Bestellungen auf, kassieren und stornieren — alles pro Tisch, alles im Browser. Admins verwalten Produkte, Tische und Benutzer, führen den Kassenbestand und erstellen den Tagesabschluss.

> **Kostenlos. Self-hosted. Auf die KassenSichV ausgelegt.**
> Keine Hardware-Bindung, keine Softwarekosten, kein Cloud-Abo für jotti selbst; allein die gesetzlich vorgeschriebene Cloud-TSE von fiskaly (und optional ein Server) kostet laufend. jotti bringt die fiskalischen Bausteine mit: eine BSI-zertifizierte Cloud-TSE, Belegausgabe nach § 146a AO, ein append-only Kassenjournal (GoBD) und den DSFinV-K-Export (v2.4). Den konformen Betrieb (TSE-Vertrag, Kassenmeldung, Aufbewahrung) verantwortet der Betreiber.

## Was jotti kann

- **Kassenbetrieb:** Bestellungen auf Tische buchen (Produkte, Varianten, Steuersätze, Kommentare), Zahlungen kassieren (Teilzahlungen, Rückgeldberechnung), stornieren (Admin und Serviceleitung, mit Pflichtkommentar), auf einen anderen Tisch umbuchen; Tisch-Übersicht mit offenem Saldo, Positionen und Bestellhistorie; Favoriten-Tische auf dem Dashboard; Direktverkauf ohne Tisch.
- **Küche:** Bestell- und Küchenbons automatisch an zugeordnete Bondrucker, pro Kategorie konfigurierbar.
- **Kassenführung:** fortlaufend nummerierte Kassensitzungen, Anfangsbestand, Soll-Bestand nach Komponenten, Einlagen und Entnahmen (Geldtransit), Kassensturz mit automatisch gebuchter Differenz, Tagesabschluss (Z-Bon) mit fortlaufender Nummer und Umsatzaggregation.
- **Abrechnung und Reporting:** Tagesabrechnung nach Steuersatz, Abrechnung je Tisch und je Servicekraft, Produktumsatz-Reporting, DSFinV-K-Export als ZIP-Archiv (v2.4).
- **Verwaltung und Sicherheit:** Admin-Bereich für Produkte (mit Varianten und Steuersätzen), Tische, Benutzer und Betreiber-Stammdaten; Rollen `admin`, `serviceleitung`, `service`; Onboarding per Einmalpasswort, Argon2id-Hashing, JWT-Auth.
- **Fiskal:** Event-Sourcing für eine lückenlose, unveränderliche Bestellhistorie; integrierte Cloud-TSE von fiskaly mit Signatur jedes Vorgangs; Belegausgabe mit TSE-Signatur, QR-Code, Steuersatz und Betreiberadresse.

Was jotti bewusst nicht kann: [docs/produktbeschreibung.md](docs/produktbeschreibung.md#62-was-jotti-bewusst-nicht-ist).

## Installation für Vereine

Für den Einsatz beim Vereinsfest braucht ihr die Kommandozeile nicht: Ladet das Windows-Release von der [GitHub-Releases-Seite](https://github.com/nicograef/jotti/releases) herunter und startet es per Doppelklick. Die vollständige Anleitung steht im [Leitfaden für Vereine](https://jotti.rocks) und unter [docs/leitfaden/installation.md](docs/leitfaden/installation.md).

## Schnellstart (Entwicklung)

```bash
make init
make dev
# Frontend: http://localhost | API: http://localhost/api
```

### Print-Relay

Das Print-Relay verbindet den jotti-Server mit den ESC/POS-Bondruckern mit 80 mm Papier, im Netzwerk erreichbar (Ethernet oder WLAN), TCP-Port 9100, feste IP-Adresse empfohlen. Es läuft auf einem Rechner im Drucker-Netzwerk:

```bash
make build-relay
RELAY_AUTH_TOKEN="<Token aus .env des Servers>" \
RELAY_BACKEND_URL="https://jotti.meinverein.de/api" \
./windows/relay/relay
```

Optionale Umgebungsvariablen:

- `RELAY_POLL_SECONDS` — Abfrageintervall, Standard `2`.
- `RELAY_TLS_SKIP_VERIFY=1` — Zertifikatsprüfung überspringen. Ohne `RELAY_BACKEND_URL` nutzt das Relay `https://localhost/api` und überspringt die Zertifikatsprüfung automatisch (lokales, selbstsigniertes Setup); bei gültigem Zertifikat **nicht** setzen.

Bei nicht erreichbarem Drucker:

- Pro Zyklus genau ein kurzer Zustellversuch (TCP-Timeout 2 s); den Fehlversuch meldet das Relay ans Backend.
- Nach sechs gemeldeten Fehlversuchen markiert das Backend den Auftrag als `fehlgeschlagen` (im Admin unter »Bondrucker« sichtbar, dort erneut einreihbar oder verwerfbar).
- Noch offene Aufträge liefert das Relay mit steigendem Abstand erneut aus (Backoff 5 s, 15 s, 30 s, 60 s, 180 s).

Schnelltest gegen den laufenden Stack:

```bash
curl -X POST http://localhost/api/relay/poll -H "Content-Type: application/json" -d '{"token":"<RELAY_AUTH_TOKEN>"}'
# 200 mit `auftraege` bei gültigem Token, 400 mit {"code":"unauthorized"} bei ungültigem
```

## Erster Login (Admin-Zugang)

Beim ersten Start legt das Backend automatisch den Benutzer `admin` an und gibt ein einmaliges, sechsstelliges Anmelde-Passwort in seinem Log aus. Damit meldet ihr euch einmalig an und legt danach ein eigenes Passwort fest (»Neues Passwort festlegen«).

Wo der Code steht, hängt vom Setup ab:

- **Windows-Release:** Der Starter zeigt den Code direkt im Konsolenfenster.
- **Server per `make prod-init`:** Der Code wird am Ende der Ausgabe angezeigt.
- **Manuelles `docker compose` (Entwicklung/Self-Hosting):** aus dem Backend-Log lesen:

  ```bash
  docker compose logs backend | grep ADMIN-EINMALPASSWORT
  ```

Ausführliche Anleitung je nach Setup: [docs/leitfaden/installation.md](docs/leitfaden/installation.md) (Windows) und [docs/leitfaden/self-hosting.md](docs/leitfaden/self-hosting.md) (Server/VPS).

## Architektur

Kasse-Operationen (Bestellungen, Zahlungen, Stornierungen, Umbuchungen, Kassensitzungen) werden via **Event Sourcing** im Kassenjournal (append-only) persistiert, Stammdaten nutzen klassisches CRUD. Alle API-Endpunkte sind `POST`; einzige Ausnahme ist `GET /health`, das für Container-Orchestrierung probebar bleibt.

Bounded Contexts, Aggregate, Invarianten und Design-Entscheidungen: [docs/handbuch.md](docs/handbuch.md).

## Für wen ist jotti?

Eingetragene Vereine (e.V.), gemeinnützige Organisationen und NPOs mit ehrenamtlichen Teams (5–30 Helfer:innen), die temporäre Veranstaltungen im Bargeld-Betrieb abrechnen.

**Nicht geeignet für:**

- Dauerbetrieb (Restaurants, Cafés)
- Kartenzahlung / NFC / Online-Payment
- Kommerzielle Gastro-Betriebe (ohne separate Lizenz)

> **Compliance-Hinweis:** jotti ist ein elektronisches Aufzeichnungssystem nach § 1 KassenSichV und unterliegt damit der TSE-Pflicht nach § 146a AO. Die fiskalischen Bausteine (TSE-Anbindung, Belegausgabe und DSFinV-K-Export v2.4) sind integriert; eine geprüfte Konformität wird nicht zugesichert. jotti erfüllt die TSE-Pflicht über eine Cloud-TSE von fiskaly — der Betreiber schließt den Vertrag mit fiskaly selbst ab und trägt die API-Schlüssel über den geführten TSE-Assistenten im Admin-Bereich ein (jotti speichert sie in seiner Datenbank). Weitere Informationen: [docs/compliance.md](docs/compliance.md) und der [Leitfaden für Vereine](docs/leitfaden/was-ist-jotti.md).

## Lizenz & Urheberrecht

**Copyright (c) 2025-2026 Nico Gräf. Alle Rechte vorbehalten.**

jotti steht unter einer proprietären Source-Available-Lizenz: Der Quellcode ist öffentlich einsehbar, aber jede Nutzung — Installation, Deployment, Betrieb — setzt eine vorherige Nutzungsvereinbarung in Textform (E-Mail) mit dem Autor voraus, gewerbliche Nutzung eine separate kommerzielle Lizenz (graef.nico@gmail.com).

Lizenztext: [LICENSE](LICENSE) · Lizenzmodell: [docs/lizenzmodell.md](docs/lizenzmodell.md) · Nutzungsbedingungen & Prozess: [TERMS.md](TERMS.md)
