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

| Dokument                                                   | Inhalt                                                        |
| ---------------------------------------------------------- | ------------------------------------------------------------- |
| [docs/development.md](docs/development.md)                 | Lokale Entwicklung, Tests, Deployment, CI/CD                  |
| [docs/requirements.md](docs/requirements.md)               | Vollständiger Anforderungskatalog (50 Anforderungen)          |
| [docs/implementation-plan.md](docs/implementation-plan.md) | Implementierungsplan für die nächsten Features                |
| [docs/language.md](docs/language.md)                       | Ubiquitous Language: Domain-Begriffe und DDD-Empfehlungen     |
| [docs/database.md](docs/database.md)                       | Datenbank & Persistenz (sqlc, Repository-Layer)               |
| [docs/theory/](docs/theory/README.md)                      | Architektur-Theorie (DDD, Event-Sourcing, CQRS, …)            |
| [docs/adr/orm.md](docs/adr/orm.md)                         | ADR: Bewertung von ORM-Alternativen und Entscheidung für sqlc |
| [AGENTS.md](AGENTS.md)                                     | Instruktionen für KI-Coding-Agenten                           |
