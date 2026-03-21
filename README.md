# jotti — Das kostenlose Kassensystem für Vereinsfeste.

Ein kostenloses, quelloffenes **Gastronomie-Kassensystem (mPOS)** für Vereine und gemeinnützige Organisationen — Vereinsfeste, Weihnachtsmärkte, Konzerte, Maihocks, Sommerfeste.

Servicekräfte nehmen auf ihren eigenen Smartphones Bestellungen auf, bestätigen die Ausgabe, kassieren und stornieren — alles pro Tisch, alles im Browser. Admins verwalten Produkte, Tische und Benutzer, führen den Kassenbestand und erstellen den Tagesabschluss.

> **Kostenlos. Self-hosted. Fiskalkonform.**
> Keine Hardware-Bindung. Keine laufenden Kosten. Kein Cloud-Abo. KassenSichV-konform mit TSE-Anbindung und DSFinV-K-Export.

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

- 🖥️ **Küchendisplay (KDS)** — eingehende Bestellungen in Echtzeit auf einem Bildschirm in Küche oder Ausgabe
- 🍳 **Ausgabestationen** — Zubereitungsstatus verwalten, Servicekräfte sehen, wann Positionen abholbereit sind
- 🖨️ **Bon-Druck** — Bestell- und Küchenbons automatisch an zugeordnete Bondrucker senden (pro Kategorie konfigurierbar)

### Kassenführung

- 📂 **Abrechnungskreis** — fortlaufend nummerierte Kassensitzungen eröffnen und schließen
- 💵 **Anfangsbestand** — Wechselgeld zu Beginn einer Veranstaltung erfassen
- 📊 **Kassenbestand** — Soll-Bestand jederzeit einsehen, aufgeschlüsselt nach Komponenten
- 🔃 **Kassenbewegungen** — Geldtransit, Privatentnahmen und Privateinlagen buchen
- ✅ **Kassensturz** — Ist-Bestand eingeben, Differenz berechnen, Abweichung automatisch buchen
- 📄 **Tagesabschluss (Z-Bon)** — formaler Abschluss mit fortlaufender Nummer und Stammdaten-Snapshot

### Abrechnung & Reporting

- 📊 **Tagesabrechnung** — Gesamtübersicht aller Umsätze, Zahlungen und offenen Beträge, nach Steuersatz aufgeschlüsselt
- 🧾 **Abrechnung pro Tisch** — detaillierte Aufstellung je Tisch
- 👤 **Abrechnung pro Servicekraft** — Umsatz und Transaktionen pro Person
- 📈 **Produktumsatz-Reporting** — meistverkaufte Varianten, Mengen und Einnahmen pro Produkt
- 📥 **Datenexport (CSV)** — Umsätze, Bestellungen und Artikeldaten für die Vereinsbuchhaltung
- 📦 **DSFinV-K-Export** — maschinenlesbarer Export für die Finanzverwaltung (ZIP-Archiv nach DSFinV-K v2.4)

### Verwaltung & Sicherheit

- ⚙️ **Admin-Bereich** für Produkte (mit Varianten und Steuersätzen), Tische, Benutzer und Betreiber-Stammdaten
- 🔐 **Rollenmodell** mit `admin`, `serviceleitung` und `service`
- 🔑 **Sicheres Onboarding** per Einmalpasswort, Argon2id-Hashing, JWT-Auth
- 📜 **Event-Sourcing** — lückenlose, unveränderliche Bestellhistorie (GoBD-konform durch Append-only-Architektur)
- 🔗 **Kryptografische Hash-Chain** — SHA-256-Verkettung aller Events, nachträgliche Manipulation nachweisbar
- 🛡️ **TSE-Anbindung** — integrierte Cloud-TSE-Schnittstelle (fiskaly) mit Signatur jedes Vorgangs
- 🧾 **Belegausgabe** — gesetzeskonforme Belege mit TSE-Signatur, QR-Code, Steuersatz und Betreiberadresse

## Schnellstart

```bash
cp .env.example .env
make dev
# Frontend: http://localhost | API: http://localhost/api
```

## Tech-Stack

| Komponente    | Technologie                                           |
| ------------- | ----------------------------------------------------- |
| Frontend      | React 19, Vite, Tailwind CSS 4, shadcn/ui, TypeScript |
| Backend       | Go 1.26, stdlib `net/http`, pgx/v5                    |
| Datenbank     | PostgreSQL 17                                         |
| TSE           | Cloud-TSE via fiskaly (Adapter-Pattern, BYOT)         |
| Reverse Proxy | nginx (HTTPS via Let's Encrypt)                       |

Tisch-Operationen (Bestellungen, Ausgaben, Zahlungen, Stornierungen, Auszahlungen) werden via **Event Sourcing** (append-only) persistiert. Stammdaten und Kassenführung nutzen immutable Records bzw. klassisches CRUD. Alle API-Endpunkte sind ausschließlich `POST`.

## Für wen ist jotti?

✅ **Geeignet für:**

- Eingetragene Vereine (e.V.), gemeinnützige Organisationen, NPOs
- Temporäre Veranstaltungen: Vereinsfeste, Sommerfeste, Weihnachtsmärkte, Maihocks, Konzerte
- Ehrenamtliche Teams (5–30 Servicekräfte)
- Bargeld-Betrieb mit voller Fiskalkonformität (TSE, DSFinV-K, Belegausgabe)

❌ **Nicht geeignet für:**

- Dauerbetrieb (Restaurants, Cafés)
- Kartenzahlung / NFC / Online-Payment
- Kommerzielle Gastro-Betriebe (ohne separate Lizenz)

> **Compliance-Hinweis:** jotti ist ein elektronisches Aufzeichnungssystem nach § 1 KassenSichV und erfüllt die TSE-Pflicht nach § 146a AO. Die TSE-Anbindung erfolgt über eine Cloud-TSE (BYOT-Modell) — der Betreiber schließt selbst einen Vertrag mit einem TSE-Anbieter (z. B. fiskaly) ab und konfiguriert die API-Schlüssel über die `.env`-Datei. Weitere Informationen: [docs/compliance.md](docs/compliance.md).

## Lizenz & Urheberrecht

**Copyright (c) 2025 Nico Gräf. Alle Rechte vorbehalten.**

jotti ist lizenziert unter der **AGPL-3.0-or-later mit Zusatzbedingungen** (Source-Available, Non-Commercial) — siehe [LICENSE](LICENSE). Das vollständige Lizenzmodell besteht aus dem AGPL-3.0-Text **und** den verbindlichen Additional Conditions in derselben Datei.

**Was das bedeutet:**

- ✅ **Eingetragene Vereine (e.V.), gemeinnützige Stiftungen und NGOs/NPOs** dürfen jotti kostenlos nutzen, installieren und betreiben.
- ✅ **Nicht-kommerzielle Projekte** dürfen jotti forken, modifizieren und einsetzen — aber das Ergebnis muss unter **denselben Lizenzbedingungen** (AGPL-3.0 + Zusatzbedingungen) veröffentlicht und betrieben werden.
- ⚠️ **Wer jotti modifiziert und als Netzwerkservice (SaaS) anbietet**, muss den vollständigen Quellcode aller Änderungen unter denselben Lizenzbedingungen veröffentlichen.
- ❌ **Kommerzielle Nutzung ist ohne separate Lizenz nicht erlaubt** — auch nicht, wenn der Quellcode unter AGPL offengelegt wird.
- ❌ **Proprietäre Abspaltungen sind nicht erlaubt.**
- 💼 **Kommerzielle Lizenzierung:** Für gewerbliche Nutzung ist eine separate kommerzielle Lizenz vom Urheber erforderlich — Kontakt über GitHub.

Ausführliche Informationen: [docs/lizenz-und-nutzung.md](docs/lizenz-und-nutzung.md)
