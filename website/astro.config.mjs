// @ts-check
import starlight from '@astrojs/starlight'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig } from 'astro/config'

// https://astro.build/config
export default defineConfig({
  site: 'https://jotti.rocks',
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
