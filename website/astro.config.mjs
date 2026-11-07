// @ts-check
import { unified } from '@astrojs/markdown-remark'
import react from '@astrojs/react'
import starlight from '@astrojs/starlight'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath } from 'node:url'
import { defineConfig } from 'astro/config'

import { docGroups, docSlug } from './src/lib/published-docs.ts'
import { remarkDocLinks } from './src/lib/remark-doc-links.ts'
import { externalizeInlineScripts } from './src/lib/externalize-inline-scripts.ts'

const docsDir = fileURLToPath(new URL('../docs', import.meta.url))
const repoBaseUrl = 'https://github.com/nicograef/jotti/blob/main'

export default defineConfig({
  site: 'https://jotti.rocks',
  markdown: {
    processor: unified({
      remarkPlugins: [[remarkDocLinks, { docsDir, repoBaseUrl }]],
    }),
  },
  integrations: [
    react(),
    starlight({
      title: 'jotti',
      // Ersetzt Starlights Inline-Theme-Init durch das externe /theme-init.js
      // (CSP). Siehe src/components/ThemeProvider.astro.
      components: {
        ThemeProvider: './src/components/ThemeProvider.astro',
      },
      // Explizit, weil Starlights Default auf ein fehlendes /favicon.svg zeigt.
      favicon: '/icons/jotti-icon-light-32.png',
      // Locale `de` als Root statt i18n-Routen: so sind auch die Framework-Texte
      // (Suche, „Auf dieser Seite") deutsch.
      defaultLocale: 'root',
      locales: {
        root: { label: 'Deutsch', lang: 'de' },
      },
      customCss: ['./src/styles/starlight.css'],
      // Slugs liegen unter /docs/ (docSlug in src/lib/published-docs.ts); / bleibt die
      // Landing.
      sidebar: docGroups.map((group) => ({
        label: group.label,
        items: group.items.map((item) => ({
          label: item.label,
          slug: docSlug(item.path),
        })),
      })),
    }),
    // Externalisiert nach dem Build Starlights verbliebene is:inline-Skripte
    // (Produktiv-CSP `script-src 'self'`).
    externalizeInlineScripts(),
  ],
  vite: {
    plugins: [tailwindcss()],
    // 0 erzwingt externe Skriptdateien, auch für die React-Hydration: die
    // Produktiv-CSP (`script-src 'self'`) verbietet Inline-Skripte.
    build: { assetsInlineLimit: 0 },
  },
})
