import { docsSchema } from '@astrojs/starlight/schema'
import { glob } from 'astro/loaders'
import { defineCollection } from 'astro:content'

// Einzige Quelle der Wahrheit: Die veröffentlichten Dokumente werden direkt aus
// dem top-level `docs/`-Verzeichnis gelesen (keine Kopie, kein Sync-Skript).
//
// Diese Liste ist die *eine* Stelle, an der die Veröffentlichungs-Auswahl steht.
// Pfade sind relativ zu `docs/`.
const publishedDocs = ['lizenzmodell.md']

export const collections = {
  docs: defineCollection({
    loader: glob({
      base: '../docs',
      pattern: publishedDocs,
      // Alle Doku-Routen unter /docs/ ablegen, damit die Landing `/` frei bleibt.
      // Aus `lizenzmodell.md` wird der Slug `docs/lizenzmodell` → /docs/lizenzmodell/.
      generateId: ({ entry }) => `docs/${entry.replace(/\.[^.]+$/, '')}`,
    }),
    schema: docsSchema(),
  }),
}
