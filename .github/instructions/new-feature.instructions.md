---
description: "Use when adding a new feature, endpoint, API route, page, or implementing a new user story."
---

# Neues Feature hinzufügen

## Backend (neuer Endpunkt)

1. Domain-Modell + zog-Schema in `domain/<domain>/`
2. SQL-Query in `sqlc/queries/<domain>.sql` definieren, dann `sqlc generate` ausführen
3. Repository-Interface + Implementierung in `repository/<domain>_repo/` (wraps sqlc-generierte Funktionen)
4. Application-Service in `api/<domain>/application/`
5. HTTP-Handler in `api/<domain>/http/`
6. Route registrieren in `api/admin.go` oder `api/service.go`
7. Unit-Test mit `//go:build unit` Tag

## Frontend (neue Seite)

1. Zod-Schema + TypeScript-Typen in Feature-Verzeichnis
2. Backend-Client-Klasse (nutzt `BackendClient`-Interface aus `@/lib/Backend`)
3. Custom Hook via `useFetch<T>()` aus `@/lib/useFetch`
4. React-Komponenten
5. Route in `src/routes.ts` registrieren
