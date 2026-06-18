// @ts-check
import starlight from '@astrojs/starlight'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath } from 'node:url'
import { defineConfig } from 'astro/config'

import { remarkDocLinks } from './src/lib/remark-doc-links.ts'

// Querverweise: Autoren schreiben repo-relative Markdown-Links in `docs/`; der
// remark-Link-Rewriter bildet sie auf Website-Routen bzw. GitHub-URLs ab.
const docsDir = fileURLToPath(new URL('../docs', import.meta.url))
const repoBaseUrl = 'https://github.com/nicograef/jotti/blob/main'

// https://astro.build/config
export default defineConfig({
  site: 'https://jotti.rocks',
  markdown: {
    // Tupel-Form [attacher, options]: unified ruft remarkDocLinks(options) auf.
    remarkPlugins: [[remarkDocLinks, { docsDir, repoBaseUrl }]],
  },
  integrations: [
    starlight({
      title: 'jotti',
      // Einsprachig deutsch: ein Locale `de` als Root, keine i18n-Routen.
      // Dadurch sind auch die Framework-Texte (Suche, „Auf dieser Seite") deutsch.
      defaultLocale: 'root',
      locales: {
        root: { label: 'Deutsch', lang: 'de' },
      },
      // Marken-Token-Set speist das Doku-Theme.
      customCss: ['./src/styles/starlight.css'],
      // Doku liegt vollständig unter /docs/ (siehe generateId in content.config.ts);
      // die Landing auf / bleibt eine eigene Astro-Seite.
      // Sidebar-Gruppen aus der PRD (ohne Leitfaden, der in Phase 4 folgt).
      sidebar: [
        {
          label: 'Recht und Steuern',
          items: [
            { label: 'Steuerrecht Gastronomie', slug: 'docs/steuerrecht' },
            {
              label: 'Verfahrensdokumentation',
              slug: 'docs/verfahrensdokumentation',
            },
            { label: 'Compliance-Grundlagen', slug: 'docs/compliance' },
          ],
        },
        {
          label: 'Über jotti',
          items: [
            { label: 'Produktbeschreibung', slug: 'docs/produktbeschreibung' },
            { label: 'Lizenzmodell', slug: 'docs/lizenzmodell' },
          ],
        },
      ],
    }),
  ],
  vite: {
    plugins: [tailwindcss()],
  },
})
