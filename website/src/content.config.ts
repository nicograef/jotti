import { docsSchema } from '@astrojs/starlight/schema'
import { glob } from 'astro/loaders'
import { defineCollection } from 'astro:content'

import { publishedDocs } from './lib/published-docs'

// Einzige Quelle der Wahrheit: Die veröffentlichten Dokumente werden direkt aus
// dem top-level `docs/`-Verzeichnis gelesen (keine Kopie, kein Sync-Skript).
// Welche Dateien das sind, steht an genau einer Stelle in `published-docs.ts`.
export const collections = {
  docs: defineCollection({
    loader: glob({
      base: '../docs',
      pattern: [...publishedDocs],
      // Alle Doku-Routen unter /docs/ ablegen, damit die Landing `/` frei bleibt.
      // Aus `lizenzmodell.md` wird der Slug `docs/lizenzmodell` → /docs/lizenzmodell/.
      generateId: ({ entry }) => `docs/${entry.replace(/\.[^.]+$/, '')}`,
    }),
    schema: docsSchema(),
  }),
}
