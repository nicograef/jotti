// Set at build time by `define` in vite.config.ts (default `dev`) and vitest.config.ts (a fixed
// release version). `tsc -b` fails without this declaration; it lives in src/ because
// tsconfig.app.json only includes "src".
declare const __CLIENT_VERSION__: string
