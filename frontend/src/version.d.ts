// Zur Bauzeit eingebrannt per `define` in vite.config.ts (Default `dev`) und
// vitest.config.ts (feste Release-Version). Ohne diese Deklaration scheitert `tsc -b`; sie liegt unter
// src/, weil tsconfig.app.json nur "src" inkludiert.
declare const __CLIENT_VERSION__: string
