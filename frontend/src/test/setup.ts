import '@testing-library/jest-dom/vitest'

// jsdom lacks window.matchMedia, which useIsMobile subscribes to; the hook reads
// window.innerWidth instead (set it with setViewportWidth). Reduced motion reports
// as active, so useCountUp returns its end value at once.
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
