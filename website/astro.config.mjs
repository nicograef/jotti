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
      sidebar: [
        {
          label: 'Über jotti',
          items: [{ label: 'Lizenzmodell', slug: 'docs/lizenzmodell' }],
        },
      ],
    }),
  ],
  vite: {
    plugins: [tailwindcss()],
  },
})
