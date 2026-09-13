---
description: 'Use when working on React frontend code, components, pages, hooks, styling, or TypeScript types.'
applyTo: 'frontend/**'
---

Repo-weite Regeln und Guardrails: `AGENTS.md`. Frontend-Architektur: [docs/handbuch.md](../../docs/handbuch.md) §6.3. Namenskonventionen und UI-Labels: [docs/language.md](../../docs/language.md).

- **401:** `Backend.post()` erkennt 401, loggt aus und leitet zu `/login` weiter — kein manuelles 401-Handling nötig.
- **Server-State:** `@tanstack/react-query`, Client über `createQueryClient` aus `src/lib/queryClient.ts`. Schreibvorgänge laufen über `useActionSubmit` (`src/hooks/use-action-submit.ts`), nicht über `useMutation`.
- **Styling:** Tailwind CSS 4 via `@tailwindcss/vite` — es gibt keine `tailwind.config.js`.
- **Drawer:** `src/components/ui/drawer.tsx` auf Radix Dialog, kein vaul (siehe [docs/decisions.md](../../docs/decisions.md), D03 und D08).
