---
description: "Erstellt eine neue Frontend-Seite mit Backend-Anbindung, Routing und allen nötigen Komponenten."
---

# Neue Frontend-Seite anlegen

Erstelle eine neue Seite für das jotti-Frontend. Befolge diese Schritte der Reihe nach:

## Eingabe

- **Bereich**: Admin (`src/admin/`) oder Service (`src/service/`)?
- **Seiten-Name**: Wie heißt die Seite? (z.B. `ProduktListe`)
- **Route**: Welcher Pfad? (z.B. `/admin/produkte`)
- **Datenquelle**: Welcher Backend-Endpunkt wird aufgerufen?

## Schritte

1. **Zod-Schema + TypeScript-Typen** im Feature-Verzeichnis — Response-Typ validieren
2. **Backend-Client-Klasse** — nutzt `BackendClient`-Interface aus `@/lib/Backend`. Nie direkt `fetch()` verwenden.
3. **Custom Hook** via `useFetch<T>()` aus `@/lib/useFetch` — Daten laden
4. **React-Komponenten** — Page-Komponente + ggf. Unter-Komponenten
5. **Route registrieren** in `src/routes.ts` mit passendem Guard (`AdminGuard` oder `ServiceGuard`)

## Konventionen

- shadcn/ui-Komponenten verwenden (Stil: `new-york`)
- Icons: Lucide React
- Fehler-Toasts: `toast.error(...)` via Sonner
- Geldbeträge: `formatCents()` aus `@/lib/utils.ts`
- Styling: Tailwind CSS 4, `cn()` für bedingte Klassen
- Alle Benutzer-sichtbaren Strings auf **Deutsch**
