# jotti — Das kostenlose Kassensystem für Vereinsfeste.

Ein kostenloses, quelloffenes **Gastronomie-Kassensystem (mPOS)** für Vereine und gemeinnützige Organisationen — Vereinsfeste, Weihnachtsmärkte, Konzerte, Maihocks, Sommerfeste.

Servicekräfte nehmen auf ihren eigenen Smartphones Bestellungen auf, bestätigen die Ausgabe, kassieren und stornieren — alles pro Tisch, alles im Browser. Admins verwalten Produkte, Tische und Benutzer, führen den Kassenbestand und erstellen den Tagesabschluss.

> **Kostenlos. Self-hosted. Auf dem Weg zur Fiskalkonformität.**
> Keine Hardware-Bindung. Keine laufenden Kosten. Kein Cloud-Abo. Auf KassenSichV-Konformität ausgelegt — TSE-Anbindung und Belegausgabe sind integriert, der DSFinV-K-Export ist in Entwicklung (siehe [Status](#was-jotti-kann) und [docs/compliance.md](docs/compliance.md)).

## Was jotti kann

### Kassenbetrieb

- 📱 **Bestellungen** auf Tische buchen — mit Produkten, Varianten, Steuersätzen und Kommentaren
- 🚚 **Ausgabe** bestätigen
- 💰 **Zahlung** kassieren (Teilzahlungen und Rückgeldberechnung)
- ↩️ **Stornierungen** mit Rollen-Kontrolle (Admin & Serviceleitung) — mit Pflichtkommentar
- 💸 **Auszahlung** leisten — negativen Saldo ausgleichen (z. B. nach Stornierung bereits kassierter Positionen)
- 🔄 **Umbuchung** — Bestellungen auf einen anderen Tisch verschieben
- 📋 **Tisch-Übersicht** mit offenem Saldo, Positionen und Bestellhistorie
- ⭐ **Meine Tische** — Favoriten als Rich Cards auf dem Dashboard, Schnellsuche per Name/Nummer

### Küche & Ausgabe

- 🖥️ **Küchendisplay (KDS)** _(in Entwicklung)_ — eingehende Bestellungen in Echtzeit auf einem Bildschirm in Küche oder Ausgabe
- 🍳 **Ausgabestationen** _(in Entwicklung)_ — Zubereitungsstatus verwalten, Servicekräfte sehen, wann Positionen abholbereit sind
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
- 📈 **Produktumsatz-Reporting** _(in Entwicklung)_ — meistverkaufte Varianten, Mengen und Einnahmen pro Produkt
- 📥 **Datenexport (CSV)** _(in Entwicklung)_ — Umsätze, Bestellungen und Artikeldaten für die Vereinsbuchhaltung
- 📦 **DSFinV-K-Export** — maschinenlesbarer Export für die Finanzverwaltung (ZIP-Archiv nach DSFinV-K v2.5)

### Verwaltung & Sicherheit

- ⚙️ **Admin-Bereich** für Produkte (mit Varianten und Steuersätzen), Tische, Benutzer und Betreiber-Stammdaten
- 🔐 **Rollenmodell** mit `admin`, `serviceleitung` und `service`
- 🔑 **Sicheres Onboarding** per Einmalpasswort, Argon2id-Hashing, JWT-Auth
- 📜 **Event-Sourcing** — lückenlose, unveränderliche Bestellhistorie (GoBD-konform durch Append-only-Architektur)
- 🛡️ **TSE-Anbindung** — integrierte Cloud-TSE von fiskaly mit Signatur jedes Vorgangs
- 🧾 **Belegausgabe** — gesetzeskonforme Belege mit TSE-Signatur, QR-Code, Steuersatz und Betreiberadresse

## Schnellstart

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
./windows/relay/jotti-relay
```

Optionale Umgebungsvariablen: `RELAY_POLL_SECONDS` (Abfrageintervall, Standard `2`) und `RELAY_TLS_SKIP_VERIFY=1` (Zertifikatsprüfung überspringen). Ohne `RELAY_BACKEND_URL` nutzt das Relay `https://localhost/api` und überspringt die Zertifikatsprüfung automatisch (lokales, selbstsigniertes Setup); bei gültigem Zertifikat `RELAY_TLS_SKIP_VERIFY` **nicht** setzen. Bei nicht erreichbarem Drucker macht das Relay pro Zyklus genau einen kurzen Zustellversuch (TCP-Timeout 2 s) und meldet den Fehlversuch ans Backend; nach drei gemeldeten Fehlversuchen markiert das Backend den Auftrag als `fehlgeschlagen` (im Admin unter »Druckstationen« sichtbar, dort erneut einreihbar oder verwerfbar). Noch offene Aufträge liefert der nächste Poll erneut. Schnelltest: `curl -X POST http://localhost:3000/relay/poll -H "Content-Type: application/json" -d '{"token":"<RELAY_AUTH_TOKEN>"}'` → `200` mit `auftraege` bei gültigem Token, `401` bei ungültigem.

## Tech-Stack

| Komponente    | Technologie                                           |
| ------------- | ----------------------------------------------------- |
| Frontend      | React 19, Vite, Tailwind CSS 4, shadcn/ui, TypeScript |
| Backend       | Go 1.26, stdlib `net/http`, pgx/v5                    |
| Datenbank     | PostgreSQL 17                                         |
| TSE           | Cloud-TSE via fiskaly (Adapter-Pattern)               |
| Reverse Proxy | Caddy (HTTPS via Let's Encrypt)                       |

Kasse-Operationen (Bestellungen, Ausgaben, Zahlungen, Stornierungen, Auszahlungen, Kassensitzungen) werden via **Event Sourcing** im Kassenjournal (append-only) persistiert. Eine synchrone Projektion (`tisch_sessions`) und eine CRUD-Entität (`kassensitzungen`) ermöglichen schnelle Reads. Stammdaten nutzen klassisches CRUD. Alle API-Endpunkte sind ausschließlich `POST`.

## Für wen ist jotti?

✅ **Geeignet für:**

- Eingetragene Vereine (e.V.), gemeinnützige Organisationen, NPOs
- Temporäre Veranstaltungen: Vereinsfeste, Sommerfeste, Weihnachtsmärkte, Maihocks, Konzerte
- Ehrenamtliche Teams (1–30 Helfer:innen)
- Bargeld-Betrieb mit dem Ziel voller Fiskalkonformität (TSE und Belegausgabe vorhanden, DSFinV-K-Export in Entwicklung, siehe [docs/compliance.md](docs/compliance.md))

❌ **Nicht geeignet für:**

- Dauerbetrieb (Restaurants, Cafés)
- Kartenzahlung / NFC / Online-Payment
- Kommerzielle Gastro-Betriebe (ohne separate Lizenz)

> **Compliance-Hinweis:** jotti ist ein elektronisches Aufzeichnungssystem nach § 1 KassenSichV und unterliegt damit der TSE-Pflicht nach § 146a AO. Die TSE-Anbindung und die Belegausgabe sind integriert; der DSFinV-K-Export ist in Entwicklung. jotti erfüllt die TSE-Pflicht über eine Cloud-TSE von fiskaly — der Betreiber schließt den Vertrag mit fiskaly selbst ab und konfiguriert die API-Schlüssel über die `.env`-Datei. Weitere Informationen: [docs/compliance.md](docs/compliance.md) und der [Leitfaden für Vereine](docs/leitfaden.md).

## Lizenz & Urheberrecht

**Copyright (c) 2025-2026 Nico Gräf. Alle Rechte vorbehalten.**

jotti steht unter einer **proprietären Source-Available-Lizenz** — siehe [LICENSE](LICENSE). Der Quellcode ist öffentlich einsehbar, aber **es werden keine Nutzungsrechte automatisch gewährt**. jotti ist kein Open-Source-Projekt im Sinne der OSI-Definition.

**Was das bedeutet:**

- ✅ **Quellcode lesen** — zum Lernen, Evaluieren oder für Sicherheitsüberprüfungen — ist ohne Vereinbarung erlaubt.
- ✅ **Pull Requests** an das offizielle Repository sind willkommen (unter dem [CLA](CLA.md)).
- ✅ **Eingetragene Vereine (e.V.), gemeinnützige Stiftungen und NGOs/NPOs** können eine **kostenlose Nutzungsvereinbarung** mit dem Autor abschließen.
- ❌ **Jede Nutzung** (Installation, Deployment, Betrieb) **erfordert eine vorherige Nutzungsvereinbarung in Textform (E-Mail)** mit dem Autor.
- ❌ **Forks, Modifikation und Weitergabe** sind nicht gestattet (außer PRs an das offizielle Repo).
- 💼 **Kommerzielle Lizenzierung:** Für gewerbliche Nutzung ist eine separate kommerzielle Lizenz vom Urheber erforderlich — Kontakt über GitHub.

Ausführliche Informationen: [docs/lizenzmodell.md](docs/lizenzmodell.md) · Nutzungsbedingungen & Prozess: [TERMS.md](TERMS.md)
