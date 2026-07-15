# jotti — Das kostenlose Kassensystem für Vereinsfeste.

> [!NOTE]
> **Beta:** jotti 1.0 wird gerade noch getestet und geprüft.

Ein kostenloses **Gastronomie-Kassensystem (mPOS)** mit einsehbarem Quellcode (Source-Available) für Vereine und gemeinnützige Organisationen — Vereinsfeste, Weihnachtsmärkte, Konzerte, Maihocks, Sommerfeste.

Servicekräfte nehmen auf ihren eigenen Smartphones Bestellungen auf, kassieren und stornieren — alles pro Tisch, alles im Browser. Admins verwalten Produkte, Tische und Benutzer, führen den Kassenbestand und erstellen den Tagesabschluss.

> **Kostenlos. Self-hosted. Auf die KassenSichV ausgelegt.**
> Keine Hardware-Bindung, keine Softwarekosten, kein Cloud-Abo für jotti selbst; allein die gesetzlich vorgeschriebene Cloud-TSE von fiskaly (und optional ein Server) kostet laufend. jotti bringt die fiskalischen Bausteine mit: eine BSI-zertifizierte Cloud-TSE, Belegausgabe nach § 146a AO, ein append-only Kassenjournal (GoBD) und den DSFinV-K-Export (v2.4). Den konformen Betrieb (TSE-Vertrag, Kassenmeldung, Aufbewahrung) verantwortet der Betreiber.

## Was jotti kann

### Kassenbetrieb

- 📱 **Bestellungen** auf Tische buchen — mit Produkten, Varianten, Steuersätzen und Kommentaren
- 💰 **Zahlung** kassieren (Teilzahlungen und Rückgeldberechnung)
- ↩️ **Stornierungen** mit Rollen-Kontrolle (Admin & Serviceleitung) — mit Pflichtkommentar
- 🔄 **Umbuchung** — Bestellungen auf einen anderen Tisch verschieben
- 📋 **Tisch-Übersicht** mit offenem Saldo, Positionen und Bestellhistorie
- ⭐ **Meine Tische** — Favoriten als große Tischkarten auf dem Dashboard, Schnellsuche per Name/Nummer
- 🛒 **Direktverkauf** — bestellen, kassieren und ausgeben in einem Schritt, ohne Tisch (mit Historie und Storno)

### Küche

- 🖨️ **Bon-Druck** — Bestell- und Küchenbons automatisch an zugeordnete Bondrucker senden (pro Kategorie konfigurierbar)

### Kasse

- 📂 **Kassensitzung** — fortlaufend nummerierte Kassensitzungen eröffnen und schließen
- 💵 **Anfangsbestand** — Wechselgeld zu Beginn einer Veranstaltung erfassen
- 📊 **Kassenbestand** — Soll-Bestand jederzeit einsehen, aufgeschlüsselt nach Komponenten
- 🔃 **Kassenbewegungen** — Einlagen und Entnahmen (Geldtransit) buchen
- ✅ **Kassensturz** — Ist-Bestand eingeben, Differenz berechnen, Abweichung automatisch buchen
- 📄 **Tagesabschluss (Z-Bon)** — formaler Abschluss mit fortlaufender Nummer und Stammdaten-Snapshot

### Abrechnung & Reporting

- 📊 **Tagesabrechnung** — Gesamtübersicht aller Umsätze, Zahlungen und offenen Beträge, nach Steuersatz aufgeschlüsselt
- 🧾 **Abrechnung pro Tisch** — detaillierte Aufstellung je Tisch
- 👤 **Abrechnung pro Servicekraft** — Umsatz und Transaktionen pro Person
- 📈 **Produktumsatz-Reporting** — meistverkaufte Varianten, Mengen und Einnahmen pro Produkt
- 📦 **DSFinV-K-Export** — maschinenlesbarer Export für die Finanzverwaltung (ZIP-Archiv nach DSFinV-K v2.4)

### Verwaltung & Sicherheit

- ⚙️ **Admin-Bereich** für Produkte (mit Varianten und Steuersätzen), Tische, Benutzer und Betreiber-Stammdaten
- 🔐 **Rollenmodell** mit `admin`, `serviceleitung` und `service`
- 🔑 **Sicheres Onboarding** per Einmalpasswort, Argon2id-Hashing, JWT-Auth
- 📜 **Event-Sourcing** — lückenlose, unveränderliche Bestellhistorie (GoBD-konform durch Append-only-Architektur)
- 🛡️ **TSE-Anbindung** — integrierte Cloud-TSE von fiskaly mit Signatur jedes Vorgangs
- 🧾 **Belegausgabe** — gesetzeskonforme Belege mit TSE-Signatur, QR-Code, Steuersatz und Betreiberadresse

## Installation für Vereine

Für den Einsatz beim Vereinsfest braucht ihr die Kommandozeile nicht: Ladet das Windows-Release von der [GitHub-Releases-Seite](https://github.com/nicograef/jotti/releases) herunter und startet es per Doppelklick. Die vollständige Anleitung steht im [Leitfaden für Vereine](https://jotti.rocks) und unter [docs/leitfaden/installation.md](docs/leitfaden/installation.md).

## Schnellstart (Entwicklung)

```bash
make init
make dev
# Frontend: http://localhost | API: http://localhost/api
```

### Print-Relay

Das Print-Relay verbindet den jotti-Server mit den ESC/POS-Bondruckern (80 mm, Ethernet, TCP Port 9100; statische IP empfohlen). Es läuft auf einem Rechner im Drucker-Netzwerk:

```bash
make build-relay
RELAY_AUTH_TOKEN="<Token aus .env des Servers>" \
RELAY_BACKEND_URL="https://jotti.meinverein.de/api" \
./windows/relay/relay
```

Optionale Umgebungsvariablen: `RELAY_POLL_SECONDS` (Abfrageintervall, Standard `2`) und `RELAY_TLS_SKIP_VERIFY=1` (Zertifikatsprüfung überspringen). Ohne `RELAY_BACKEND_URL` nutzt das Relay `https://localhost/api` und überspringt die Zertifikatsprüfung automatisch (lokales, selbstsigniertes Setup); bei gültigem Zertifikat `RELAY_TLS_SKIP_VERIFY` **nicht** setzen. Bei nicht erreichbarem Drucker macht das Relay pro Zyklus genau einen kurzen Zustellversuch (TCP-Timeout 2 s) und meldet den Fehlversuch ans Backend; nach sechs gemeldeten Fehlversuchen markiert das Backend den Auftrag als `fehlgeschlagen` (im Admin unter »Druckstationen« sichtbar, dort erneut einreihbar oder verwerfbar). Noch offene Aufträge liefert das Relay mit steigendem Abstand erneut aus (Backoff 5 s, 15 s, 30 s, 60 s, 180 s). Schnelltest gegen den laufenden Stack: `curl -X POST http://localhost/api/relay/poll -H "Content-Type: application/json" -d '{"token":"<RELAY_AUTH_TOKEN>"}'` → `200` mit `auftraege` bei gültigem Token, `401` bei ungültigem.

## Tech-Stack

| Komponente    | Technologie                                           |
| ------------- | ----------------------------------------------------- |
| Frontend      | React 19, Vite, Tailwind CSS 4, shadcn/ui, TypeScript |
| Backend       | Go 1.26, stdlib `net/http`, pgx/v5                    |
| Datenbank     | PostgreSQL 17                                         |
| TSE           | Cloud-TSE via fiskaly (Adapter-Pattern)               |
| Reverse Proxy | Caddy (HTTPS via Let's Encrypt)                       |

Kasse-Operationen (Bestellungen, Zahlungen, Stornierungen, Umbuchungen, Kassensitzungen) werden via **Event Sourcing** im Kassenjournal (append-only) persistiert. Eine synchrone Projektion (`tisch_sessions`) und eine CRUD-Entität (`kassensitzungen`) ermöglichen schnelle Reads. Stammdaten nutzen klassisches CRUD. Alle API-Endpunkte sind ausschließlich `POST`.

## Für wen ist jotti?

✅ **Geeignet für:**

- Eingetragene Vereine (e.V.), gemeinnützige Organisationen, NPOs
- Temporäre Veranstaltungen: Vereinsfeste, Sommerfeste, Weihnachtsmärkte, Maihocks, Konzerte
- Ehrenamtliche Teams (1–30 Helfer:innen)
- Bargeld-Betrieb mit integrierten fiskalischen Bausteinen (TSE-Anbindung, Belegausgabe und DSFinV-K-Export v2.4; den konformen Betrieb verantwortet der Betreiber, siehe [docs/compliance.md](docs/compliance.md))

❌ **Nicht geeignet für:**

- Dauerbetrieb (Restaurants, Cafés)
- Kartenzahlung / NFC / Online-Payment
- Kommerzielle Gastro-Betriebe (ohne separate Lizenz)

> **Compliance-Hinweis:** jotti ist ein elektronisches Aufzeichnungssystem nach § 1 KassenSichV und unterliegt damit der TSE-Pflicht nach § 146a AO. Die fiskalischen Bausteine (TSE-Anbindung, Belegausgabe und DSFinV-K-Export v2.4) sind integriert; eine geprüfte Konformität wird nicht zugesichert. jotti erfüllt die TSE-Pflicht über eine Cloud-TSE von fiskaly — der Betreiber schließt den Vertrag mit fiskaly selbst ab und trägt die API-Schlüssel über den geführten TSE-Assistenten im Admin-Bereich ein (jotti speichert sie in seiner Datenbank). Weitere Informationen: [docs/compliance.md](docs/compliance.md) und der [Leitfaden für Vereine](docs/leitfaden/was-ist-jotti.md).

## Lizenz & Urheberrecht

**Copyright (c) 2025-2026 Nico Gräf. Alle Rechte vorbehalten.**

jotti steht unter einer proprietären Source-Available-Lizenz (siehe [LICENSE](LICENSE)). Der Quellcode ist öffentlich einsehbar, aber es werden keine Nutzungsrechte automatisch gewährt. jotti ist kein Open-Source-Projekt im Sinne der OSI-Definition.

**Was das bedeutet:**

- ✅ **Quellcode lesen:** zum Lernen, Evaluieren oder für Sicherheitsüberprüfungen ohne Vereinbarung erlaubt.
- ✅ **Pull Requests:** an das offizielle Repository willkommen (unter dem [CLA](CLA.md)).
- ✅ **Kostenlose Nutzungsvereinbarung:** Eingetragene Vereine (e.V.), gemeinnützige Stiftungen und NGOs/NPOs können sie mit dem Autor abschließen.
- ❌ **Jede Nutzung erfordert eine Vereinbarung:** Installation, Deployment und Betrieb setzen eine vorherige Nutzungsvereinbarung in Textform (E-Mail) mit dem Autor voraus.
- ❌ **Keine Forks oder Weitergabe:** Forks, Modifikation und Weitergabe sind nicht gestattet (außer PRs an das offizielle Repo).
- 💼 **Kommerzielle Lizenzierung:** Für gewerbliche Nutzung ist eine separate kommerzielle Lizenz vom Urheber erforderlich (Kontakt über GitHub).
- 🤝 **Optionaler Service:** Einrichtung, Hosting-Hilfe, Support und Schulung sind auf Anfrage gegen Entgelt verfügbar, nach gesonderter Absprache, freiwillig und unabhängig von der kostenlosen Nutzung der Software. Die Verantwortung für den konformen Betrieb bleibt beim Verein.

Ausführliche Informationen: [docs/lizenzmodell.md](docs/lizenzmodell.md) · Nutzungsbedingungen & Prozess: [TERMS.md](TERMS.md)
