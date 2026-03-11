---
description: "Use when working on React frontend code, components, pages, hooks, styling, or TypeScript types."
applyTo: "frontend/**"
---

# Frontend-Konventionen

## Verzeichnisstruktur

```
frontend/
  src/routes.ts                 # Alle Routen + Guards
  src/App.tsx                   # Root-Komponente
  src/lib/                      # Auth, Backend-Client, useFetch-Hook, Utilities
  src/admin/                    # Admin-Bereich (Produkte, Tische, Benutzer)
  src/service/                  # Service-Bereich (Tisch-Workflow)
  src/pages/                    # Login, Passwort setzen
  src/components/ui/            # shadcn/ui-Komponenten
  src/components/common/        # Gemeinsame Komponenten
```

## UI-Bibliotheken

- **shadcn/ui** (Stil: `new-york`, Radix-basiert)
- **Lucide React** (Icons)
- **Sonner** (Toasts) — alle mutativen Aktionen zeigen `toast.error(...)` bei Fehlern
- **Vaul** (Drawers)

## Patterns

- **401-Interceptor**: `Backend.post()` erkennt 401, loggt aus und leitet zu `/login` weiter — kein manuelles 401-Handling nötig
- **Drawer-Pattern**: Bestellen, Bezahlen, Stornieren, Liefern öffnen Bottom-Sheet-Drawer mit Zusammenfassung. Hilfsfunktionen (`selectPositionen`, `calculateTotalPrice`) in `src/service/components/table/drawerUtils.ts`
- **Geldbeträge anzeigen**: `formatCents()` aus `src/lib/utils.ts` — nie inline formatieren

## Styling

- Tailwind CSS 4 via `@tailwindcss/vite` (keine `tailwind.config.js`)
- CSS-Variablen in `src/index.css` (Violet/Indigo-Schema, Dark Mode via `.dark`-Klasse)
- `cn()` Utility aus `src/lib/utils.ts` (`clsx` + `tailwind-merge`)
- Path-Alias: `@/` → `./src/`
