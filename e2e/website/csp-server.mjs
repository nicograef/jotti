// Dependency-free static file server for the built website (`website/dist`).
//
// Serves the artefact with the EXACT production Content-Security-Policy header of
// the jotti.rocks block from `reverse-proxy/nginx.rocks.conf`. The CSP is parsed
// from that file at runtime, so it can never silently drift from production.
//
// Shared by two consumers (see `docs/plans/plan-website-redesign.md`):
//   - the per-phase CSP verification (`csp-check.mjs`),
//   - the Phase-9 OG-image screenshot mode.
//
// No external dependencies — plain `node:http`/`node:fs`, runnable with any Node.

import { createServer } from 'node:http'
import { readFile, stat } from 'node:fs/promises'
import { extname, join, normalize, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const repoRoot = resolve(fileURLToPath(new URL('../..', import.meta.url)))
const nginxConf = join(repoRoot, 'reverse-proxy', 'nginx.rocks.conf')

const CONTENT_TYPES = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.mjs': 'text/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.xml': 'application/xml; charset=utf-8',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.webp': 'image/webp',
  '.woff2': 'font/woff2',
  '.ico': 'image/x-icon',
  '.txt': 'text/plain; charset=utf-8',
  '.wasm': 'application/wasm',
  '.map': 'application/json; charset=utf-8',
}

// Extract the production CSP verbatim. The jotti.rocks landing+docs block is the
// only CSP header carrying the Pagefind `'wasm-unsafe-eval'` exception, which
// uniquely identifies it among the server blocks in the file.
export async function readProductionCsp() {
  const conf = await readFile(nginxConf, 'utf8')
  const matches = [...conf.matchAll(/add_header\s+Content-Security-Policy\s+"([^"]+)"/g)]
  const csp = matches.map((m) => m[1]).find((value) => value.includes("'wasm-unsafe-eval'"))
  if (!csp) {
    throw new Error(`jotti.rocks CSP not found in ${nginxConf}`)
  }
  return csp
}

// Start a static server for `distDir`. Resolves with { url, port, close() }.
export async function startStaticServer(distDir, { csp, host = '127.0.0.1' } = {}) {
  const root = resolve(distDir)
  const cspHeader = csp ?? (await readProductionCsp())

  const server = createServer(async (req, res) => {
    try {
      const urlPath = decodeURIComponent(new URL(req.url, 'http://localhost').pathname)
      let filePath = normalize(join(root, urlPath))
      if (!filePath.startsWith(root)) {
        res.writeHead(403).end('Forbidden')
        return
      }
      let info = await stat(filePath).catch(() => null)
      if (info?.isDirectory()) {
        filePath = join(filePath, 'index.html')
        info = await stat(filePath).catch(() => null)
      }
      if (!info) {
        // Astro emits pretty-URL directories; fall back to `<path>/index.html`.
        const fallback = join(root, urlPath, 'index.html')
        info = await stat(fallback).catch(() => null)
        if (info) filePath = fallback
      }
      if (!info) {
        res.writeHead(404, { 'Content-Security-Policy': cspHeader }).end('Not found')
        return
      }
      const body = await readFile(filePath)
      res.writeHead(200, {
        'Content-Type': CONTENT_TYPES[extname(filePath)] ?? 'application/octet-stream',
        'Content-Security-Policy': cspHeader,
      })
      res.end(body)
    } catch (err) {
      res.writeHead(500).end(String(err))
    }
  })

  await new Promise((res) => server.listen(0, host, res))
  const { port } = server.address()
  return {
    url: `http://${host}:${port}`,
    port,
    close: () => new Promise((res) => server.close(res)),
  }
}
