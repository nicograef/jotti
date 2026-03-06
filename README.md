# jotti

Ein einfaches Bestell- und Kassensystem für Vereine und Nonprofit-Organisationen (z. B. Vereinsfeste, Weihnachtsmärkte).

Servicekräfte nehmen auf Smartphones Bestellungen auf, liefern aus, kassieren und stornieren — alles pro Tisch. Admins verwalten Produkte, Tische und Benutzer über einen eigenen Admin-Bereich.

## Features

### ✅ Umgesetzt (23 von 50 Anforderungen)

- **Bestellungen** auf Tische buchen (Produkte mit Varianten, Mengen, Kategorien)
- **Lieferungen** als ausgeliefert markieren
- **Bezahlungen** registrieren (Teilzahlung möglich)
- **Stornierungen** (nur Admin und Serviceleitung)
- **Tisch-Übersicht**: offener Saldo, unbezahlte/ungelieferte Positionen, Verlauf
- **Admin-Bereich**: Produkte, Tische und Benutzer verwalten
- **Rollen**: `admin`, `senior_service` (Service + Stornierung), `service`
- **Authentifizierung**: JWT (12h), Einmalpasswort-Onboarding, Argon2id-Hashing
- **Kommentar/Notiz** pro Bestellvorgang (max. 100 Zeichen)

### ❌ Nächste Schritte

Siehe [Implementierungsplan](docs/implementation-plan.md) und vollständigen [Anforderungskatalog](docs/requirements.md) (50 Anforderungen).

## Architektur

Single-Tenant, deployed via Docker Compose auf einer VM.

| Komponente    | Technologie                                           |
| ------------- | ----------------------------------------------------- |
| Frontend      | React 19, Vite, Tailwind CSS 4, shadcn/ui, TypeScript |
| Backend       | Go 1.26, stdlib `net/http`, pgx/v5                    |
| Datenbank     | PostgreSQL 17                                         |
| Reverse Proxy | nginx (HTTPS via Let's Encrypt)                       |

- Stammdaten (Benutzer, Produkte, Tische) → relationale Tabellen (CRUD)
- Tisch-Operationen (Bestellungen, Bezahlungen, Lieferungen, Stornierungen) → **Event Sourcing** (append-only)
- Alle API-Endpunkte sind ausschließlich `POST`
- Frontend validiert Request/Response mit Zod-Schemas

## Projektstruktur

```
backend/          Go-Backend (HTTP → Application → Domain → Repository)
frontend/         React-SPA (admin/, service/, components/, lib/)
database/         SQL-Migrationen (golang-migrate)
reverse-proxy/    nginx-Konfigurationen (dev, staging, production)
docs/             Dokumentation
```

## Schnellstart

```bash
cp .env.example .env
make dev
# Frontend: http://localhost | API: http://localhost/api
```

Ausführliche Anleitungen: [Entwicklung & Deployment](docs/development.md)

## Dokumentation

| Dokument | Inhalt |
| --- | --- |
| [docs/development.md](docs/development.md) | Lokale Entwicklung, Tests, Deployment, CI/CD |
| [docs/requirements.md](docs/requirements.md) | Vollständiger Anforderungskatalog (50 Anforderungen) |
| [docs/implementation-plan.md](docs/implementation-plan.md) | Implementierungsplan für die nächsten Features |
| [AGENTS.md](AGENTS.md) | Instruktionen für KI-Coding-Agenten |

## Offene Fragen

- Steuern/Mehrwertsteuer: mehrere Sätze pro Produkt? Brutto/Netto-Preise, Rundung?
- Dynamische Preise (Happy Hour, Event-spezifisch), Rabatte/Aktionen, Gratisartikel?
- Ausverkauft: manuell gesetzt vs. Bestandsführung?
