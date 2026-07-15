import '@testing-library/jest-dom/vitest'

// jsdom implementiert window.matchMedia nicht. useIsMobile (Service-Split,
// ADR 07/08) ruft es im Effect auf; ohne Polyfill würde jede Komponente, die
// den Hook nutzt, im Test werfen. useIsMobile liest die Breite über
// window.innerWidth (jsdom: 1024px ≥ lg → Desktop), nicht über `matches`;
// Tests fürs schmale Handy-Layout mocken useIsMobile explizit auf true.
// `prefers-reduced-motion` wird bewusst als aktiv gemeldet, damit bewegungs-
// abhängige Hooks (useCountUp) im Test weiterhin sofort den Endwert liefern
// statt zu animieren.
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
