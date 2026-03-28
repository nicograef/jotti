---
description: "Review des jotti-Frontends auf Mobile-UX, UI-Konsistenz, Workflow-Reibung und Terminologie-Drift."
argument-hint: "Optionaler Fokus, z.B. login, table page, admin product flow, service checkout, ganzes Frontend"
agent: "agent"
---

# Mobile-UX-Review

Finde UX-Probleme die ehrenamtliche Helfer auf Smartphones während einer Veranstaltung ausbremsen.

## Prüfbereiche

### Workflow-Reibung

- Zu viele Taps für häufige Service-Aktionen
- Versteckte oder unklare nächste Schritte
- Fehlende Rückmeldung nach Speichern, Bestellen, Ausgeben, Kassieren, Stornieren
- Schwer zu behebende Fehlerzustände

### Mobile-First-Qualität

- Komponenten die auf schmalen Screens brechen
- Dichte Tabellen/Formulare ohne Mobile-Fallback
- Zu kleine oder zu nah beieinander liegende Touch-Targets
- Wichtige Aktionen unterhalb des sichtbaren Bereichs

### UI-Konsistenz

- Gleiches Konzept unterschiedlich beschriftet auf verschiedenen Screens
- Ähnliche Aktionen mit unterschiedlichen Button-Labels oder Platzierungen
- Inkonsistente Loading-, Empty- und Error-States

### Domain-Sprache

- Terminologie-Drift in UI-Labels vs. Backend/Domain-Begriffen
- Labels die technisch korrekt aber für Ehrenamtliche unklar sind

## Output

Pro Finding: **Kategorie** → **Was** → **Wo** (Datei:Zeilen) → **User-Impact** → **Vorschlag** → **Aufwand** (S/M/L).

Am Ende: Quick Wins zuerst, dann Konsistenz-Fixes zum Batchen.
