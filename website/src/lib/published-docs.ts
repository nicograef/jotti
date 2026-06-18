// Einzige Quelle der Wahrheit für die Veröffentlichungs-Auswahl.
//
// Diese Liste bestimmt an *einer* Stelle, welche Dateien aus dem top-level
// `docs/` als Website-Doku erscheinen. Sie wird von zwei Stellen gelesen:
//   - der Content-Collection (Glob-Loader in `content.config.ts`),
//   - dem Link-Rewriter (`link-rewriter.ts` / `remark-doc-links.ts`), der
//     repo-relative Markdown-Links auf `/docs/<slug>/`-Routen abbildet.
//
// Pfade sind relativ zu `docs/`. Der Leitfaden (`docs/leitfaden/`) folgt in
// Phase 4.
export const publishedDocs = [
  'compliance.md',
  'steuerrecht.md',
  'verfahrensdokumentation.md',
  'produktbeschreibung.md',
  'lizenzmodell.md',
] as const
