---
description: "Auditiert den jotti-Codebase auf Cross-Layer-Konsistenz, Vereinfachungsmöglichkeiten und führt die Repo-Verifikation durch."
argument-hint: "Optionaler Fokus, z.B. auth, admin CRUD, service flow, reporting, ganzes Repo"
agent: "agent"
---

# Code-Qualitäts-Audit

Auditiere den jotti-Codebase in drei Schritten. Führe alle drei durch, sofern der User keinen Fokus einschränkt.

## Schritt 1: Cross-Layer-Konsistenz

Stimmen die Schichten noch überein?

- Frontend Request-Bodies vs. Backend Handler Request-Structs
- Frontend Response-Parsing vs. Backend Response JSON-Shape
- TypeScript-Typen und Zod-Schemas vs. Go-Structs und JSON-Payloads
- SQL-Queries vs. Schema (Spalten, Nullability, Defaults, Status-Werte)
- sqlc-generierte Erwartungen vs. Repository-Mapping
- Geldbeträge überall in Cent/Integer
- Ubiquitous-Language-Konsistenz (Tisch, Bestellung, Ausgabe, Zahlung, Stornierung)
- Validierungsregeln auf beiden Seiten konsistent

Trace repräsentative Flows end-to-end: Frontend Call → Backend-Klasse → HTTP-Handler → Application Service → Repository → SQL.

## Schritt 2: Vereinfachung (Readability-First)

Was ist schwerer zu lesen als nötig?

- Lange, verschachtelte Logik die vereinfacht werden kann
- Interfaces mit nur einer Implementierung die Indirektion ohne Wert hinzufügen
- Wrapper-Funktionen die nur weiterleiten
- Stale Patterns aus früheren Architekturphasen (vor Event-Sourcing, vor table_state)
- Ungenutzter Code, tote Exports, Endpoints die nichts aufruft
- Inkonsistenter Coding-Stil über ähnliche Module hinweg
- SQL-Queries die unnötig komplex sind
- Repository-Methoden die nur sqlc-Calls forwarden ohne Domain-Wert

## Schritt 3: Repo-Verifikation

```bash
make verify
```

Falls Fehler auftreten: Step, Typ (Lint/Test/Build), betroffene Dateien, Ursache und nächsten Debugging-Schritt berichten.

## Guardrails

- POST-only Endpoints nicht beanstanden
- Backend ist Single Source of Truth für Filterung — nicht als Problem melden
- Event-Sourcing und synchrone `tisch_sessions`-Projektion sind gewollt
- Trennung admin/service/serviceleitung nicht in Frage stellen
- Frontend-API-Calls über Backend-Klassen sind Pflicht
- Breaking Changes sind erlaubt — keine Backwards-Kompatibilität vorschlagen

## Output

Pro Finding: **Was** → **Wo** (Datei:Zeilen) → **Impact** → **Vorschlag** → **Aufwand** (S/M/L).

Am Ende: priorisierte Empfehlungen (Korrektheitsfehler > Quick Wins > Größere Refactors).
