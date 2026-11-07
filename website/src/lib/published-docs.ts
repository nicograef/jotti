// Die veröffentlichten Docs in Sidebar-Reihenfolge; Pfade relativ zu `docs/`.
// astro.config.mjs baut daraus die Sidebar, der Glob-Loader (`content.config.ts`)
// und der Link-Rewriter (`link-rewriter.ts`) lesen `publishedDocs`.
export interface DocGroup {
  readonly label: string
  readonly items: readonly { readonly path: string; readonly label: string }[]
}

export const docGroups: readonly DocGroup[] = [
  {
    label: 'Erste Schritte',
    items: [
      { path: 'leitfaden/was-ist-jotti.md', label: 'Was ist jotti?' },
      { path: 'leitfaden/betriebsarten.md', label: 'Welcher Modus passt?' },
    ],
  },
  {
    label: 'Vereinsbetrieb (Standardweg)',
    items: [
      { path: 'leitfaden/installation.md', label: 'Installation und Start' },
      {
        path: 'leitfaden/tse-einrichten.md',
        label: 'TSE einrichten (fiskaly)',
      },
      {
        path: 'leitfaden/veranstaltungstag.md',
        label: 'Der Veranstaltungstag',
      },
      { path: 'leitfaden/aktualisieren.md', label: 'Aktualisieren' },
      { path: 'leitfaden/checkliste.md', label: 'Checkliste' },
    ],
  },
  {
    label: 'Recht und Steuern',
    items: [
      { path: 'leitfaden/pflichten.md', label: 'Pflichten im Überblick' },
      {
        path: 'leitfaden/finanzamt-anmelden.md',
        label: 'Kasse beim Finanzamt anmelden',
      },
      {
        path: 'leitfaden/belege-steuersaetze.md',
        label: 'Belege und Steuersätze',
      },
      { path: 'leitfaden/datenaufbewahrung.md', label: 'Datenaufbewahrung' },
      { path: 'steuerrecht.md', label: 'Steuerrecht Gastronomie' },
      {
        path: 'verfahrensdokumentation.md',
        label: 'Muster-Verfahrensdokumentation',
      },
      { path: 'compliance.md', label: 'Compliance-Anforderungen' },
    ],
  },
  {
    label: 'Self-Hosting (Experten-Weg)',
    items: [
      {
        path: 'leitfaden/self-hosting.md',
        label: 'Eigener Server (Ersteinrichtung)',
      },
      {
        path: 'leitfaden/aktualisieren-backups.md',
        label: 'Server aktualisieren und Backups',
      },
    ],
  },
  {
    label: 'Hilfe',
    items: [
      { path: 'leitfaden/fehlersuche.md', label: 'Fehlersuche' },
      { path: 'leitfaden/tse-sonderfaelle.md', label: 'TSE-Sonderfälle' },
      { path: 'leitfaden/haeufige-fragen.md', label: 'Häufige Fragen' },
    ],
  },
  {
    label: 'Über jotti',
    items: [
      { path: 'produktbeschreibung.md', label: 'Produktbeschreibung' },
      { path: 'lizenzmodell.md', label: 'Lizenzmodell' },
    ],
  },
]

export const publishedDocs: readonly string[] = docGroups.flatMap((group) =>
  group.items.map((item) => item.path),
)

// docSlug bildet einen Pfad auf seinen Slug ab, wie generateId in content.config.ts.
export const docSlug = (path: string): string =>
  `docs/${path.replace(/\.[^.]+$/, '')}`
