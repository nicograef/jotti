# jotti — Das kostenlose Kassensystem für Vereinsfeste.

Ein kostenloses, quelloffenes **Gastronomie-Kassensystem (mPOS)** für Vereine und gemeinnützige Organisationen — Vereinsfeste, Weihnachtsmärkte, Konzerte, Maihocks, Sommerfeste.

Servicekräfte nehmen auf ihren eigenen Smartphones Bestellungen auf, bestätigen die Ausgabe, kassieren und stornieren — alles pro Tisch, alles im Browser. Admins verwalten Produkte, Tische und Benutzer über einen eigenen Admin-Bereich.

> **Kostenlos. Self-hosted. Open Source.**
> Keine Hardware-Bindung. Keine laufenden Kosten. Kein Cloud-Abo. Kein Zahlungsgateway.

## Was jotti kann

### Kassenbetrieb

- 📱 **Bestellungen** auf Tische buchen — mit Produkten, Varianten und Kommentaren
- 🚚 **Ausgabe** bestätigen
- 💰 **Zahlung** kassieren (Teilzahlungen möglich)
- ↩️ **Stornierungen** mit Rollen-Kontrolle (Admin & Serviceleitung) — mit Pflichtkommentar
- 💸 **Auszahlung leisten** — negativen Saldo ausgleichen (z. B. nach Stornierung bereits kassierter Positionen)
- 📋 **Tisch-Übersicht** mit offenem Saldo, Positionen und Bestellhistorie

### Küche & Ausgabe

- 🖥️ **Küchendisplay (KDS)** — eingehende Bestellungen in Echtzeit auf einem Bildschirm in Küche oder Ausgabe
- 🖨️ **Bon-Druck** — Bestell- und Küchenbons direkt an einen Bondrucker senden

### Abrechnung & Reporting

- 📊 **Tagesabrechnung** — Gesamtübersicht aller Umsätze, Zahlungen und offenen Beträge
- 🧾 **Abrechnung pro Tisch** — detaillierte Aufstellung je Tisch
- 👤 **Abrechnung pro Servicekraft** — Umsatz und Transaktionen pro Person
- 📈 **Reporting** — Produktumsätze, meistverkaufte Varianten, Gesamteinnahmen

### Verwaltung & Sicherheit

- ⚙️ **Admin-Bereich** für Produkte (mit Varianten), Tische und Benutzer
- 🔐 **Rollenmodell** mit `admin`, `serviceleitung` und `service`
- 🔑 **Sicheres Onboarding** per Einmalpasswort, Argon2id-Hashing, JWT-Auth
- 📜 **Event-Sourcing** — lückenlose, unveränderliche Bestellhistorie (GoBD-Grundsätze durch Append-only-Architektur)

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
| Reverse Proxy | nginx (HTTPS via Let's Encrypt)                       |

Tisch-Operationen (Bestellungen, Ausgaben, Zahlungen, Stornierungen, Auszahlungen) werden via **Event Sourcing** (append-only) persistiert. Stammdaten nutzen klassisches CRUD. Alle API-Endpunkte sind ausschließlich `POST`.

## Für wen ist jotti?

✅ **Geeignet für:**

- Eingetragene Vereine (e.V.), gemeinnützige Organisationen, NPOs
- Temporäre Veranstaltungen: Vereinsfeste, Sommerfeste, Weihnachtsmärkte, Maihocks, Konzerte
- Ehrenamtliche Teams (5–30 Servicekräfte)
- Bargeld-Betrieb mit schrittweiser Compliance-Roadmap (TSE/KassenSichV)

❌ **Nicht geeignet für:**

- Dauerbetrieb (Restaurants, Cafés)
- Betriebe, die **sofort** eine vollständig zertifizierte TSE-Lösung benötigen
- Kartenzahlung / NFC / Online-Payment

> **Compliance-Hinweis:** jotti ist ein elektronisches Aufzeichnungssystem nach § 1 KassenSichV und unterliegt der TSE-Pflicht nach § 146a AO. Die TSE-Integration wird schrittweise implementiert — der Betreiber ist verpflichtet, eine Cloud-TSE (BYOT) einzubinden, sobald die entsprechende Phase der Roadmap umgesetzt ist. Siehe [docs/roadmap.md](docs/roadmap.md) und [docs/compliance.md](docs/compliance.md).

## Lizenz & Urheberrecht

**Copyright (c) 2025 Nico Gräf. Alle Rechte vorbehalten.**

jotti ist lizenziert unter der **AGPL-3.0-or-later mit Zusatzbedingungen** (Source-Available, Non-Commercial) — siehe [LICENSE](LICENSE). Das vollständige Lizenzmodell besteht aus dem AGPL-3.0-Text **und** den verbindlichen Additional Conditions in derselben Datei.

**Was das bedeutet:**

- ✅ **Eingetragene Vereine (e.V.), gemeinnützige Stiftungen und NGOs/NPOs** dürfen jotti kostenlos nutzen, installieren und betreiben.
- ✅ **Nicht-kommerzielle Open-Source-Projekte** dürfen jotti forken, modifizieren und einsetzen — aber das Ergebnis muss unter **denselben Lizenzbedingungen** (AGPL-3.0 + Zusatzbedingungen) veröffentlicht und betrieben werden.
- ⚠️ **Wer jotti modifiziert und als Netzwerkservice (SaaS) anbietet**, muss den vollständigen Quellcode aller Änderungen unter denselben Lizenzbedingungen veröffentlichen.
- ❌ **Kommerzielle Nutzung ist ohne separate Lizenz nicht erlaubt** — auch nicht, wenn der Quellcode unter AGPL offengelegt wird. Niemand darf jotti oder Ableitungen davon gewerblich verwerten, ohne eine kommerzielle Lizenz des Urhebers zu besitzen.
- ❌ **Proprietäre Abspaltungen sind nicht erlaubt.** Ableitungen dürfen nicht unter restriktiveren oder permissiveren Bedingungen veröffentlicht werden, die die Nicht-Kommerziell-Einschränkung oder das AGPL-Copyleft aufheben.
- 💼 **Kommerzielle Lizenzierung:** Für gewerbliche Nutzung (z.B. kostenpflichtiges SaaS, White-Label, Integration in kommerzielle Produkte) ist eine separate kommerzielle Lizenz vom Urheber erforderlich — Kontakt über GitHub.

Ausführliche Informationen: [docs/lizenz-und-nutzung.md](docs/lizenz-und-nutzung.md)
