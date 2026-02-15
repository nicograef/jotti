# jotti

jotti ist ein einfaches Bestell- und Kassensystem für Vereine und Nonprofit-Organisationen (z. B. Vereinsfeste, Weihnachtsmärkte).

## Was ist jotti?

Eine Webapp, die auf jedem Smartphone läuft. Servicekräfte nehmen Bestellungen auf, liefern Produkte aus, kassieren und stornieren — alles pro Tisch. Administratoren verwalten Produkte, Tische und Benutzer über einen eigenen Admin-Bereich.

## Features

### ✅ Umgesetzt (22 von 50 Anforderungen)

- **Bestellungen** auf Tische buchen (Produkte mit Varianten und Mengen auswählen)
- **Lieferungen** als ausgeliefert markieren
- **Bezahlungen** für offene Positionen registrieren (Teilzahlung möglich)
- **Stornierungen** von Positionen (aktuell für beide Rollen, soll nur Admin sein)
- **Tisch-Übersicht**: offener Saldo, unbezahlte und ungelieferte Positionen, Verlauf aller Aktionen
- **Admin-Bereich**: Produkte (mit Varianten, Preisen, Kategorien), Tische und Benutzer verwalten
- **Rollen**: `admin` (Vollzugriff) und `service` (Bestell-/Kassierbetrieb)
- **Authentifizierung**: JWT (12h Gültigkeit), Einmalpasswort-Onboarding, Argon2id-Hashing
- **Kommentar/Notiz** pro Bestellvorgang (max. 100 Zeichen)
- **Produktvarianten** (z. B. Pommes mit Ketchup/Soße)
- **Produktkategorien** in der Bestellansicht (gruppiert nach Essen, Getränke, Sonstiges)

### 🔧 Teilweise umgesetzt (1)

- **Stornierung nur für Admins (#22)**: Stornierung steht beiden Rollen offen, Rollenprüfung fehlt

### ❌ Nächste offene Must-haves

| #   | Anforderung                              |
| --- | ---------------------------------------- |
| 23  | Tisch-Schnellsuche per Shortcut          |
| 24  | Übersicht eigene Bestellungen mit Status |
| 25  | Bestellungen auf anderen Tisch umbuchen  |
| 26  | Umsatz pro Bediener (Tagesabrechnung)    |
| 27  | Bons drucken (formatiert)                |
| 31  | Freibon mit freier Preiseingabe          |
| 33  | Offline-Fähigkeit                        |

### Nice-to-haves

- Rückgeldberechnung (#37)
- Freitext-Notiz pro Position (#42)
- Bezeichnung/Name pro Bestellung (#36)
- Reporting und Datenexport (#38–40)
- Zubereitungsstatus einsehen (#35)

Der vollständige Anforderungskatalog (50 Anforderungen) mit Implementierungsvorschlägen liegt in `ANFORDERUNGEN.md`.

## Architektur

Single-Tenant, deployed via Docker Compose auf einer VM.

| Komponente    | Technologie                                           |
| ------------- | ----------------------------------------------------- |
| Frontend      | React 19, Vite, Tailwind CSS 4, shadcn/ui, TypeScript |
| Backend       | Go 1.25, stdlib `net/http`, pgx/v5                    |
| Datenbank     | PostgreSQL 17                                         |
| Reverse Proxy | nginx (HTTPS via Let's Encrypt)                       |

- Benutzer, Produkte und Tische → relationale Tabellen (CRUD)
- Bestellungen, Bezahlungen, Lieferungen, Stornierungen → **Event Sourcing** (append-only Event-Log mit Snapshots)
- Alle API-Endpunkte sind ausschließlich `POST`
- Frontend-API-Aufrufe validieren Request und Response mit Zod-Schemas

## Projektstruktur

```
backend/          Go-Backend (Layered Architecture: HTTP → Application → Domain → Repository)
frontend/         React-SPA (Feature-basiert: admin/, service/, components/, lib/)
database/         SQL-Migrationen (golang-migrate)
reverse-proxy/    nginx-Konfigurationen (dev, staging, production)
```

- Der Server wird in Go geschrieben.
  - Der Server stellt eine HTTP API zur Verfügung.
  - Die API ist im Command- und Query-Pattern aufgebaut und verwendet JSON für die Datenübertragung.
  - Der Server implementiert Event Sourcing für Bestellungen und Bezahlungen.
- Das Frontend ist eine React SPA Webapp.
  - Die Webapp kommuniziert mit dem Server via HTTP API.
  - Die Webapp ist responsive und funktioniert auf Smartphones und Tablets.
  - Die Webapp hat einen Admin-Bereich für Administratoren.
- Als Webserver wird nginx eingesetzt, inklusive SSL Verschlüsselung via Let's Encrypt.
- Alle Services (Datenbank, Server und Webserver) laufen in Docker Containern und werden via Docker-Compose orchestriert.
- On-Premise Installation: Das System kann lokal auf einem Rechner oder Server installiert werden, ohne Cloud-Anbindung.
- Anmeldung via Benutzername und Passwort
  - Passwort-Hashing mit Argon2id [owasp-cheatsheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
  - Sessions werden via JSON Web Tokens (JWT) realisiert (12 Stunden Gültigkeit).

## Offene Fragen

- Steuern/Mehrwertsteuer: mehrere Sätze pro Produkt? Brutto/Netto-Preise, Rundung?
- Dynamische Preise (Happy Hour, Event-spezifisch), Rabatte/Aktionen, Gratisartikel?
- Ausverkauft: manuell gesetzt vs. Bestandsführung?
