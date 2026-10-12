import '@testing-library/jest-dom/vitest'

// jsdom lacks window.matchMedia, which useIsMobile subscribes to; the hook reads
// window.innerWidth instead (set it with setViewportWidth). Reduced motion reports
// as active, so useCountUp returns its end value at once.
class FakeMediaQueryList extends EventTarget implements MediaQueryList {
  readonly matches: boolean
  readonly media: string
  onchange: MediaQueryList['onchange'] = null

  constructor(query: string) {
    super()
    this.media = query
    this.matches = query.includes('prefers-reduced-motion')
  }

  // The fake never changes, so a legacy listener is never called.
  addListener(): void {
    return undefined
  }

  removeListener(): void {
    return undefined
  }
}

if (typeof window !== 'undefined' && typeof window.matchMedia !== 'function') {
  window.matchMedia = (query: string): MediaQueryList =>
    new FakeMediaQueryList(query)
}
