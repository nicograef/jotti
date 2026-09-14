// Astro-Integration: löst nach dem Build alle ausführbaren Inline-Skripte im
// gebauten HTML in externe Dateien auf. Die Produktiv-CSP
// (reverse-proxy/nginx.rocks.conf, jotti.rocks-Block) erlaubt nur
// `script-src 'self'` ohne `'unsafe-inline'`; Starlight liefert für
// Theme-Picker, Suche und Sidebar-Persistenz `is:inline`-Skripte aus. Die
// Integration externalisiert sie generisch — Inhalt, Attribute und Reihenfolge
// bleiben erhalten — statt jede Starlight-Komponente zu überschreiben.
//
// Unangetastet bleiben bereits externe Skripte, leere Tags und nicht-ausführbare
// Skripttypen (`application/json`, `application/ld+json`, `importmap`,
// `speculationrules`): kein JS, von der CSP ohnehin nicht blockiert.

import type { AstroIntegration } from 'astro'
import { createHash } from 'node:crypto'
import { readFile, writeFile, readdir, mkdir } from 'node:fs/promises'
import { fileURLToPath } from 'node:url'
import { join } from 'node:path'

// Kommentar ODER Skript-Element. Der Kommentar steht zuerst, damit ein
// wörtliches `<script>` in einem HTML-Kommentar als Ganzes verschluckt wird.
// Skript-Körper können laut HTML kein `</script>` enthalten — der nicht-gierige
// Body ist daher sicher.
const COMMENT_OR_SCRIPT_RE =
  /<!--[\s\S]*?-->|<script\b([^>]*)>([\s\S]*?)<\/script>/gi

// Nur klassische und Modul-Skripte sind ausführbar und CSP-relevant. Fehlt das
// type-Attribut oder ist es leer, gilt das Skript als klassisches JS.
const EXECUTABLE_TYPES = new Set([
  'module',
  'text/javascript',
  'application/javascript',
  'text/ecmascript',
  'application/ecmascript',
])

function isExecutableScript(attrs: string): boolean {
  const match = /\btype\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s>]+))/i.exec(attrs)
  if (!match) return true
  const type = (match[1] ?? match[2] ?? match[3] ?? '').trim().toLowerCase()
  return type === '' || EXECUTABLE_TYPES.has(type)
}

async function htmlFiles(dir: string): Promise<string[]> {
  const entries = await readdir(dir, { withFileTypes: true })
  const files: string[] = []
  for (const entry of entries) {
    const full = join(dir, entry.name)
    if (entry.isDirectory()) files.push(...(await htmlFiles(full)))
    else if (entry.name.endsWith('.html')) files.push(full)
  }
  return files
}

export function externalizeInlineScripts(): AstroIntegration {
  return {
    name: 'externalize-inline-scripts',
    hooks: {
      'astro:build:done': async ({ dir, logger }) => {
        const distDir = fileURLToPath(dir)
        const assetsDir = join(distDir, '_astro')
        await mkdir(assetsDir, { recursive: true })

        // hash → Skriptinhalt; über alle Seiten dedupliziert.
        const scripts = new Map<string, string>()
        let externalized = 0

        for (const file of await htmlFiles(distDir)) {
          const html = await readFile(file, 'utf8')
          let changed = false

          const rewritten = html.replace(
            COMMENT_OR_SCRIPT_RE,
            (match, attrs?: string, body?: string) => {
              // Kommentar-Alternative: unverändert lassen (attrs/body undefined).
              if (attrs === undefined) return match
              if (/\ssrc\s*=/.test(attrs)) return match
              if (body === undefined || body.trim() === '') return match
              if (!isExecutableScript(attrs)) return match

              const hash = createHash('sha256')
                .update(body)
                .digest('hex')
                .slice(0, 16)
              scripts.set(hash, body)
              changed = true
              externalized++
              return `<script${attrs} src="/_astro/inline-${hash}.js"></script>`
            },
          )

          if (changed) await writeFile(file, rewritten)
        }

        await Promise.all(
          [...scripts].map(([hash, body]) =>
            writeFile(join(assetsDir, `inline-${hash}.js`), body),
          ),
        )

        logger.info(
          `${externalized} Inline-Skript(e) externalisiert (CSP: kein inline script).`,
        )
      },
    },
  }
}
