# jotti

Ein leichtgewichtiges **Gastronomie-Kassensystem (POS)** für Vereine und Non-Profit-Organisationen bei Veranstaltungen — Vereinsfeste, Weihnachtsmärkte, Konzerte, Maihocks, Sommerfeste.

Servicekräfte nehmen auf Smartphones Bestellungen auf, liefern aus, kassieren und stornieren — alles pro Tisch. Admins verwalten Produkte, Tische und Benutzer über einen eigenen Admin-Bereich.

> **Keine Hardware-Bindung. Keine laufenden Kosten. Kein Cloud-Abo. Self-hosted, Source-Available, Mobile-first.**

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

jotti ist lizenziert unter der **AGPL-3.0-or-later mit Zusatzbedingungen** (Source-Available, Non-Commercial) — siehe [LICENSE](LICENSE). Das vollständige Lizenzmodell besteht aus dem AGPL-3.0-Text **und** den verbindlichen Additional Conditions in derselben Datei.

**Was das bedeutet:**

- ✅ **Eingetragene Vereine (e.V.), gemeinnützige Stiftungen und NGOs/NPOs** dürfen jotti kostenlos nutzen, installieren und betreiben.
- ✅ **Nicht-kommerzielle Open-Source-Projekte** dürfen jotti forken, modifizieren und einsetzen — aber das Ergebnis muss unter **denselben Lizenzbedingungen** (AGPL-3.0 + Zusatzbedingungen) veröffentlicht und betrieben werden.
- ⚠️ **Wer jotti modifiziert und als Netzwerkservice (SaaS) anbietet**, muss den vollständigen Quellcode aller Änderungen unter denselben Lizenzbedingungen veröffentlichen.
- ❌ **Kommerzielle Nutzung ist ohne separate Lizenz nicht erlaubt** — auch nicht, wenn der Quellcode unter AGPL offengelegt wird. Niemand darf jotti oder Ableitungen davon gewerblich verwerten, ohne eine kommerzielle Lizenz des Urhebers zu besitzen.
- ❌ **Proprietäre Abspaltungen sind nicht erlaubt.** Ableitungen dürfen nicht unter restriktiveren oder permissiveren Bedingungen veröffentlicht werden, die die Nicht-Kommerziell-Einschränkung oder das AGPL-Copyleft aufheben.
- 💼 **Kommerzielle Lizenzierung:** Für gewerbliche Nutzung (z.B. kostenpflichtiges SaaS, White-Label, Integration in kommerzielle Produkte) ist eine separate kommerzielle Lizenz vom Urheber erforderlich — Kontakt über GitHub.

Ausführliche Informationen: [docs/lizenz-und-nutzung.md](docs/lizenz-und-nutzung.md)
