// Screenshot-Carousel der Landing („Einblicke").
//
// Bewusst eine statische Datei in public/ (kein gebuendeltes Astro-`<script>`):
// Die Produktiv-CSP erlaubt nur `script-src 'self'` ohne `'unsafe-inline'`, also
// werden Inline-Skripte vom Browser blockiert (siehe theme-init.js).
//
// Das Carousel ist ohne JS voll bedienbar (Scroll-Snap/Wischen). Dieses Skript
// ergaenzt nur die Pfeil-Buttons und die Punkt-Indikatoren als Komfort.
const carousel = document.querySelector('[data-carousel]')

if (carousel) {
  const track = carousel.querySelector('[data-carousel-track]')
  const slides = track ? Array.from(track.children) : []
  const dots = Array.from(carousel.querySelectorAll('[data-carousel-dot]'))
  const prev = carousel.querySelector('[data-carousel-prev]')
  const next = carousel.querySelector('[data-carousel-next]')

  if (track && slides.length > 0) {
    const reduceMotion = window.matchMedia(
      '(prefers-reduced-motion: reduce)',
    ).matches
    const behavior = reduceMotion ? 'auto' : 'smooth'

    const scrollToIndex = (index) => {
      const clamped = Math.max(0, Math.min(index, slides.length - 1))
      const slide = slides[clamped]
      const left =
        slide.offsetLeft - (track.clientWidth - slide.clientWidth) / 2
      track.scrollTo({ left, behavior })
    }

    // Index der Slide, deren Mitte der Track-Mitte am naechsten ist.
    const currentIndex = () => {
      const center = track.scrollLeft + track.clientWidth / 2
      let best = 0
      let bestDistance = Infinity
      slides.forEach((slide, index) => {
        const slideCenter = slide.offsetLeft + slide.clientWidth / 2
        const distance = Math.abs(slideCenter - center)
        if (distance < bestDistance) {
          bestDistance = distance
          best = index
        }
      })
      return best
    }

    const updateState = () => {
      const index = currentIndex()
      dots.forEach((dot, i) =>
        dot.setAttribute('aria-current', String(i === index)),
      )
      if (prev) prev.disabled = index === 0
      if (next) next.disabled = index === slides.length - 1
    }

    dots.forEach((dot, index) =>
      dot.addEventListener('click', () => scrollToIndex(index)),
    )
    if (prev)
      prev.addEventListener('click', () => scrollToIndex(currentIndex() - 1))
    if (next)
      next.addEventListener('click', () => scrollToIndex(currentIndex() + 1))

    let frame = 0
    track.addEventListener('scroll', () => {
      if (frame) return
      frame = window.requestAnimationFrame(() => {
        frame = 0
        updateState()
      })
    })

    updateState()
  }
}
