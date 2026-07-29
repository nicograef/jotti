// Zur Bauzeit eingebrannte Clientversion — gesetzt per `define` in
// vite.config.ts und vitest.config.ts (Default `dev`). Ohne diese Deklaration
// scheitert bereits `tsc -b` vor dem Vite-Build. Die Datei liegt unter src/,
// weil tsconfig.app.json nur "src" inkludiert.
declare const __CLIENT_VERSION__: string
