# jotti

Ein leichtgewichtiges **Gastronomie-Kassensystem (POS)** für Vereine und Non-Profit-Organisationen bei Veranstaltungen — Vereinsfeste, Weihnachtsmärkte, Konzerte, Maihocks, Sommerfeste.

Servicekräfte nehmen auf Smartphones Bestellungen auf, liefern aus, kassieren und stornieren — alles pro Tisch. Admins verwalten Produkte, Tische und Benutzer über einen eigenen Admin-Bereich.

> **Keine Hardware-Bindung. Keine laufenden Kosten. Kein Cloud-Abo. Self-hosted, Open Source, Mobile-first.**

## Was jotti kann

- **Bestellungen** auf Tische buchen — mit Produkten, Varianten und Kommentaren
- **Lieferungen** als ausgeliefert markieren
- **Zahlungen** registrieren (Teilzahlungen möglich)
- **Stornierungen** mit Rollen-Kontrolle (Admin & Serviceleitung)
- **Tisch-Übersicht** mit offenem Saldo, Positionen und Bestellhistorie
- **Admin-Bereich** für Produkte, Tische und Benutzer
- **Rollenmodell** mit `admin`, `senior_service` und `service`
- **Sicheres Onboarding** per Einmalpasswort, Argon2id-Hashing, JWT-Auth

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

Tisch-Operationen (Bestellungen, Zahlungen, Lieferungen, Stornierungen) werden via **Event Sourcing** (append-only) persistiert. Stammdaten nutzen klassisches CRUD. Alle API-Endpunkte sind ausschließlich `POST`.

## Dokumentation

| Dokument                                                   | Inhalt                                                             |
| ---------------------------------------------------------- | ------------------------------------------------------------------ |
| [docs/development.md](docs/development.md)                 | Lokale Entwicklung, Tests, Deployment, CI/CD                       |
| [docs/requirements.md](docs/requirements.md)               | Vollständiger Anforderungskatalog (50 Anforderungen)               |
| [docs/implementation-plan.md](docs/implementation-plan.md) | Implementierungsplan für die nächsten Features                     |
| [docs/language.md](docs/language.md)                       | Ubiquitous Language: Domain-Begriffe und DDD-Empfehlungen          |
| [docs/event-storming.md](docs/event-storming.md)           | Event-Storming-Session: Domain Events, Aggregate, Bounded Contexts |
| [docs/database.md](docs/database.md)                       | Datenbank & Persistenz (sqlc, Repository-Layer)                    |
| [docs/theory/](docs/theory/README.md)                      | Architektur-Theorie (DDD, Event-Sourcing, CQRS, …)                 |
| [docs/adr/orm.md](docs/adr/orm.md)                         | ADR: Bewertung von ORM-Alternativen und Entscheidung für sqlc      |
| [docs/lizenz-und-nutzung.md](docs/lizenz-und-nutzung.md)   | Lizenz, Nutzungsvereinbarung, IP, DSGVO, Kommerzialisierung        |
| [AGENTS.md](AGENTS.md)                                     | Instruktionen für KI-Coding-Agenten                                |

## Lizenz & Urheberrecht

**Copyright (c) 2025 Nico Gräf. Alle Rechte vorbehalten.**

jotti ist lizenziert unter der [AGPL-3.0-or-later](LICENSE) (GNU Affero General Public License v3).

**Was das bedeutet:**

- ✅ **Vereine und Non-Profits** dürfen jotti kostenlos nutzen, installieren und betreiben (Self-Hosted).
- ✅ **Quellcode einsehen, modifizieren und beitragen** ist erlaubt und erwünscht.
- ⚠️ **Wer jotti modifiziert und als Netzwerkservice (SaaS) anbietet**, muss den vollständigen Quellcode aller Änderungen unter AGPL-3.0 veröffentlichen.
- ❌ **Proprietäre Abspaltungen sind nicht erlaubt.** Niemand darf jotti oder Teile davon in ein geschlossenes kommerzielles Produkt überführen, ohne die AGPL-Pflichten zu erfüllen.
- 💼 **Kommerzielle Lizenzierung:** Für die Nutzung ohne AGPL-Pflichten (z.B. proprietäres SaaS, White-Label) ist eine kommerzielle Lizenz vom Urheber erforderlich — Kontakt über GitHub.

Ausführliche Informationen: [docs/lizenz-und-nutzung.md](docs/lizenz-und-nutzung.md)
