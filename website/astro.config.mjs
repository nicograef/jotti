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
      // Explizites Favicon: schließt die /favicon.svg-Lücke des Starlight-Defaults
      // und zeigt in der Doku dieselbe Marke wie die Landing (Kopie in public/).
      favicon: '/icons/jotti-icon-light-32.png',
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
      // Sidebar-Gruppen aus der PRD. Der Leitfaden (`docs/leitfaden/`) trennt
      // Standardweg vom Experten-Weg und Technik vom Recht über die Gruppierung.
      sidebar: [
        {
          label: 'Erste Schritte',
          items: [
            { label: 'Was ist jotti?', slug: 'docs/leitfaden/was-ist-jotti' },
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
              label: 'Verfahrensdokumentation',
              slug: 'docs/verfahrensdokumentation',
            },
            { label: 'Compliance-Grundlagen', slug: 'docs/compliance' },
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
              label: 'Aktualisieren und Backups',
              slug: 'docs/leitfaden/aktualisieren-backups',
            },
            {
              label: 'TSE-Sonderfälle',
              slug: 'docs/leitfaden/tse-sonderfaelle',
            },
          ],
        },
        {
          label: 'Hilfe',
          items: [
            { label: 'Fehlersuche', slug: 'docs/leitfaden/fehlersuche' },
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
  ],
  vite: {
    plugins: [tailwindcss()],
  },
})
