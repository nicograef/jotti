// Einzige Quelle der Wahrheit für die Veröffentlichungs-Auswahl.
//
// Diese Liste bestimmt an *einer* Stelle, welche Dateien aus dem top-level
// `docs/` als Website-Doku erscheinen. Sie wird von zwei Stellen gelesen:
//   - der Content-Collection (Glob-Loader in `content.config.ts`),
//   - dem Link-Rewriter (`link-rewriter.ts` / `remark-doc-links.ts`), der
//     repo-relative Markdown-Links auf `/docs/<slug>/`-Routen abbildet.
//
// Pfade sind relativ zu `docs/`.
export const publishedDocs = [
  // Leitfaden (aus `docs/leitfaden.md` aufgeteilt in Schritt-Seiten).
  'leitfaden/was-ist-jotti.md',
  'leitfaden/installation.md',
  'leitfaden/tse-einrichten.md',
  'leitfaden/checkliste.md',
  'leitfaden/pflichten.md',
  'leitfaden/finanzamt-anmelden.md',
  'leitfaden/belege-steuersaetze.md',
  'leitfaden/datenaufbewahrung.md',
  'leitfaden/self-hosting.md',
  'leitfaden/aktualisieren-backups.md',
  'leitfaden/tse-sonderfaelle.md',
  'leitfaden/fehlersuche.md',
  'leitfaden/haeufige-fragen.md',
  // Flache Referenzdokumente.
  'compliance.md',
  'steuerrecht.md',
  'verfahrensdokumentation.md',
  'produktbeschreibung.md',
  'lizenzmodell.md',
] as const
