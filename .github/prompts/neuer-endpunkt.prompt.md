---
description: "Erstellt einen neuen Backend-API-Endpunkt mit allen Schichten (Domain, Repository, Application, HTTP-Handler, Route)."
---

# Neuen Endpunkt anlegen

Erstelle einen neuen API-Endpunkt für das jotti-Backend. Befolge diese Schritte der Reihe nach:

## Eingabe

- **Domain**: In welchem Bereich? (z.B. product, table, user)
- **Endpunkt-Name**: Wie heißt die Operation? (z.B. `produkt-erstellen`)
- **Bereich**: Admin (`api/admin.go`) oder Service (`api/service.go`)?
- **Request-/Response-Felder**: Welche Daten werden gesendet/zurückgegeben?

## Schritte

1. **Domain-Modell** in `backend/domain/{domain}/` — zog-Validierungsschema für Request
2. **SQL-Query** in `backend/sqlc/queries/{domain}.sql` — dann `sqlc generate`
3. **Repository** in `backend/repository/{domain}_repo/` — Interface + Implementierung
4. **Application-Service** in `backend/api/{domain}/application/` — Business-Logik
5. **HTTP-Handler** in `backend/api/{domain}/http/` — Request parsen, validieren, Service aufrufen, Response schreiben
6. **Route registrieren** in `backend/api/admin.go` oder `backend/api/service.go`
7. **Unit-Test** mit `//go:build unit` Tag

## Konventionen

- Alle Endpunkte sind **POST-only**
- Fehler-Response: `{"code": "<string>", "details": "<optional>"}`
- Geldbeträge in **Cent (int)**
- Fachbegriffe auf **Deutsch** (Bestellung, Zahlung, etc.)
