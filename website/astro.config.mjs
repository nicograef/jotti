// @ts-check
import { unified } from '@astrojs/markdown-remark'
import react from '@astrojs/react'
import starlight from '@astrojs/starlight'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath } from 'node:url'
import { defineConfig } from 'astro/config'

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
      // Slugs liegen unter /docs/ (generateId in content.config.ts); / bleibt die
      // Landing.
      sidebar: [
        {
          label: 'Erste Schritte',
          items: [
            { label: 'Was ist jotti?', slug: 'docs/leitfaden/was-ist-jotti' },
            {
              label: 'Welcher Modus passt?',
              slug: 'docs/leitfaden/betriebsarten',
            },
          ],
        },
        {
          label: 'Vereinsbetrieb (Standardweg)',
          items: [
            {
              label: 'Installation und Start',
              slug: 'docs/leitfaden/installation',
            },
            {
              label: 'TSE einrichten (fiskaly)',
              slug: 'docs/leitfaden/tse-einrichten',
            },
            {
              label: 'Der Veranstaltungstag',
              slug: 'docs/leitfaden/veranstaltungstag',
            },
            { label: 'Aktualisieren', slug: 'docs/leitfaden/aktualisieren' },
            { label: 'Checkliste', slug: 'docs/leitfaden/checkliste' },
          ],
        },
        {
          label: 'Recht und Steuern',
          items: [
            {
              label: 'Pflichten im Überblick',
              slug: 'docs/leitfaden/pflichten',
            },
            {
              label: 'Kasse beim Finanzamt anmelden',
              slug: 'docs/leitfaden/finanzamt-anmelden',
            },
            {
              label: 'Belege und Steuersätze',
              slug: 'docs/leitfaden/belege-steuersaetze',
            },
            {
              label: 'Datenaufbewahrung',
              slug: 'docs/leitfaden/datenaufbewahrung',
            },
            { label: 'Steuerrecht Gastronomie', slug: 'docs/steuerrecht' },
            {
              label: 'Muster-Verfahrensdokumentation',
              slug: 'docs/verfahrensdokumentation',
            },
            { label: 'Compliance-Anforderungen', slug: 'docs/compliance' },
          ],
        },
        {
          label: 'Self-Hosting (Experten-Weg)',
          items: [
            {
              label: 'Eigener Server (Ersteinrichtung)',
              slug: 'docs/leitfaden/self-hosting',
            },
            {
              label: 'Server aktualisieren und Backups',
              slug: 'docs/leitfaden/aktualisieren-backups',
            },
          ],
        },
        {
          label: 'Hilfe',
          items: [
            { label: 'Fehlersuche', slug: 'docs/leitfaden/fehlersuche' },
            {
              label: 'TSE-Sonderfälle',
              slug: 'docs/leitfaden/tse-sonderfaelle',
            },
            { label: 'Häufige Fragen', slug: 'docs/leitfaden/haeufige-fragen' },
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
