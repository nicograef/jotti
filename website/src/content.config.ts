import { docsSchema } from '@astrojs/starlight/schema'
import { glob } from 'astro/loaders'
import { defineCollection } from 'astro:content'

import { publishedDocs } from './lib/published-docs'

// Liest direkt aus dem top-level `docs/` — keine Kopie, kein Sync-Skript.
export const collections = {
  docs: defineCollection({
    loader: glob({
      base: '../docs',
      pattern: [...publishedDocs],
      // Alle Doku-Routen unter /docs/ ablegen, damit die Landing `/` frei bleibt.
      generateId: ({ entry }) => `docs/${entry.replace(/\.[^.]+$/, '')}`,
    }),
    schema: docsSchema(),
  }),
}
