# Marken-Assets

Dieses Verzeichnis hält die Master aller Marken-Assets. `frontend/public/icons/`,
`website/public/icons/` und `website/src/assets/jotti-symbol.png` sind davon abgeleitete
Laufzeitkopien: Bei einer Änderung am Logo oder an den Icons werden zuerst die Master hier
aktualisiert und danach die benötigten Größen in die Kopien übernommen.

| Datei                                   | Darstellung                                               | Einsatz                                                                                 |
| --------------------------------------- | --------------------------------------------------------- | --------------------------------------------------------------------------------------- |
| `jotti-icon-dark-16.png`                | J-Symbol auf Slate-950, abgerundete Ecken                 | Favicon 16×16 im Dark Mode (`frontend/index.html`)                                      |
| `jotti-icon-dark-32.png`                | J-Symbol auf Slate-950, abgerundete Ecken                 | Favicon 32×32 im Dark Mode (`frontend/index.html`, `website/src/layouts/Landing.astro`) |
| `jotti-icon-dark-64.png`                | J-Symbol auf Slate-950, abgerundete Ecken                 | Master ohne Laufzeitkopie                                                               |
| `jotti-icon-light-16.png`               | J-Symbol auf Slate-50, abgerundete Ecken                  | Favicon 16×16 im Light Mode, PWA-Manifest                                               |
| `jotti-icon-light-32.png`               | J-Symbol auf Slate-50, abgerundete Ecken                  | Favicon 32×32 im Light Mode und Fallback-Favicon in App und Website, PWA-Manifest       |
| `jotti-icon-light-64.png`               | J-Symbol auf Slate-50, abgerundete Ecken                  | PWA-Manifest (64×64)                                                                    |
| `jotti-logo-full-dark.png`              | J und Wortmarke auf Slate-950 (`#020617`)                 | Marketing und Druck; von keinem Build eingebunden                                       |
| `jotti-logo-full-light-transparent.png` | J und Wortmarke, transparenter Hintergrund                | Marketing und Druck; von keinem Build eingebunden                                       |
| `jotti-logo-full-light.png`             | J und Wortmarke auf Slate-50 (`#f8fafc`)                  | Marketing und Druck; von keinem Build eingebunden                                       |
| `jotti-logo-icon-dark.png`              | Quadratisches App-Icon, abgerundet, Hintergrund Slate-950 | PWA-Startbildschirm-Icon (Dark) im Manifest                                             |
| `jotti-logo-icon-light.png`             | Quadratisches App-Icon, abgerundet, Hintergrund Slate-50  | PWA-Startbildschirm-Icon (Light) im Manifest, Apple Touch Icon in App und Website       |
| `jotti-symbol.png`                      | Freigestelltes J, transparenter Hintergrund               | Website (`website/src/assets/`, importiert von `Landing.astro` und `LiveDemo.tsx`)      |

Die Master hier und die Kopien unter `website/` tragen den Spektral-Verlauf, die Kopien in
`frontend/public/icons/` das grüne `J`. Die Vier-Stellen-Regel für Spektral im App-Frontend
([D06](../docs/decisions.md)) deckt Wortmarke, Glows, Skeletons und Hairlines ab, nicht die
Icon-Assets — wer die Master ungeprüft in alle Kopien schiebt, tauscht das App-Favicon aus.

Den Spektral-Verlauf erzeugt `scripts/generate-spektral-logos.py` (Farbmodell und Checks im
Modul-Docstring); die Master werden nie direkt überschrieben.
