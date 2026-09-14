import '@testing-library/jest-dom/vitest'

// jsdom implementiert window.matchMedia nicht; useIsMobile ruft es im Effect
// auf. Der Hook liest die Breite über window.innerWidth (jsdom: 1024px ≥ lg →
// Desktop), nicht über `matches`; Tests fürs Handy-Layout mocken useIsMobile
// auf true. `prefers-reduced-motion` wird bewusst als aktiv gemeldet, damit
// useCountUp im Test sofort den Endwert liefert statt zu animieren.
if (typeof window !== 'undefined' && typeof window.matchMedia !== 'function') {
  window.matchMedia = (query: string): MediaQueryList =>
    ({
      matches: query.includes('prefers-reduced-motion'),
      media: query,
      onchange: null,
      addEventListener: () => undefined,
      removeEventListener: () => undefined,
      addListener: () => undefined,
      removeListener: () => undefined,
      dispatchEvent: () => false,
    }) as MediaQueryList
}
