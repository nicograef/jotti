import { useEffect, useRef, useState } from 'react'
import { Check, Euro, Minus, Plus, RotateCcw } from 'lucide-react'

import jottiSymbol from '../assets/jotti-symbol.png'
import {
  cartCount,
  cartTotalCents,
  demoMenu,
  formatEuro,
  initialDemoState,
  manualAdd,
  manualPay,
  manualRemove,
  replay,
  runNextStep,
  startAuto,
  stepDelayMs,
  type DemoState,
} from '../lib/live-demo'

// Live-Demo-Island (Handoff-Prototyp docs/prds/design_handoff_jotti_website, #demo).
// Der einzige UI-Nachbau der App auf der ganzen Seite. Die gesamte Logik
// (Warenkorb, Summen in Cent, Auto-Skript, Stopp, Reset) liegt UI-frei in
// src/lib/live-demo.ts; diese Komponente rendert nur den Zustand und liefert
// Timing (setTimeout) und Viewport-Trigger (IntersectionObserver, 25 %).
//
// Auto-Demo: startet einmalig beim Hereinscrollen, läuft mit den Handoff-Timings
// und stoppt dauerhaft bei jeder manuellen Interaktion. Unter
// prefers-reduced-motion startet keine Auto-Demo und es laufen keine
// Animationen — die manuelle Bedienung bleibt voll funktionsfähig.
//
// Phone-Rahmen: wiederverwendete Phase-3-Klassen (.hero-phone / -notch /
// -screen); .demo-phone hält das Telefon ruhig (siehe landing.css).

// Nummerierte Erklärschritte der linken Spalte (rein statisch, dekorativ).
const explainSteps: { accent: string; title: string; text: string }[] = [
  {
    accent: 'var(--sp-red)',
    title: 'Produkt antippen',
    text: 'Varianten mit einem Tipp auf die Bestellung buchen.',
  },
  {
    accent: 'var(--sp-teal)',
    title: 'Summe wächst live mit',
    text: 'Der offene Betrag steht immer unten in der Leiste.',
  },
  {
    accent: 'var(--sp-violet)',
    title: 'Kassieren mit Rückgeld',
    text: 'Ein Tipp auf »Kassieren« — TSE-Signatur inklusive.',
  },
]

function usePrefersReducedMotion(): boolean {
  const [reduced, setReduced] = useState(false)
  useEffect(() => {
    const query = window.matchMedia('(prefers-reduced-motion: reduce)')
    setReduced(query.matches)
    const onChange = (event: MediaQueryListEvent) => setReduced(event.matches)
    query.addEventListener('change', onChange)
    return () => query.removeEventListener('change', onChange)
  }, [])
  return reduced
}

export default function LiveDemo() {
  const [state, setState] = useState<DemoState>(initialDemoState)
  const reducedMotion = usePrefersReducedMotion()
  const phoneRef = useRef<HTMLDivElement>(null)
  const startedRef = useRef(false)

  // Auto-Demo startet einmalig, sobald das Telefon zu 25 % im Viewport ist.
  // Unter reduced-motion wird kein Observer angelegt — keine Auto-Demo.
  useEffect(() => {
    if (reducedMotion) return
    const element = phoneRef.current
    if (!element || typeof IntersectionObserver === 'undefined') return
    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting && !startedRef.current) {
            startedRef.current = true
            setState((current) => startAuto(current))
          }
        }
      },
      { threshold: 0.25 },
    )
    observer.observe(element)
    return () => observer.disconnect()
  }, [reducedMotion])

  // Treibt den laufenden Auto-Ablauf: pro (autoStatus, step) genau ein Timer mit
  // der Handoff-Wartezeit; runNextStep ist no-op, falls inzwischen gestoppt.
  useEffect(() => {
    if (reducedMotion || state.autoStatus !== 'running') return
    const timer = setTimeout(() => {
      setState((current) => runNextStep(current))
    }, stepDelayMs(state.step))
    return () => clearTimeout(timer)
  }, [state.autoStatus, state.step, reducedMotion])

  function onReplay() {
    startedRef.current = true
    // Unter reduced-motion nur den Warenkorb leeren (kein Auto-Ablauf).
    setState(reducedMotion ? initialDemoState : replay())
  }

  const total = cartTotalCents(state.cart)
  const totalStr = formatEuro(total)
  const hasCart = cartCount(state.cart) > 0

  return (
    <div className="grid items-center gap-14 nav:grid-cols-2">
      {/* Textspalte */}
      <div>
        <p className="inline-flex items-center gap-2.5 text-[0.78rem] font-bold tracking-[0.14em] text-[var(--sp-teal)] uppercase">
          <span
            className="h-[3px] w-[22px] rounded-sm"
            style={{ background: 'var(--spectral)' }}
            aria-hidden="true"
          ></span>
          Live-Demo
        </p>
        <h2 className="font-brand mt-3 text-[length:var(--fs-h2)] leading-[1.06] font-bold tracking-[-0.025em]">
          Bestellen, kassieren,
          <br />
          fertig.
        </h2>
        <p className="mt-4 max-w-[30em] text-[1.0625rem] leading-relaxed text-muted">
          Genau so sieht der Bestellvorgang auf dem Handy deiner Servicekraft
          aus. Tippe selbst Produkte an — oder lehn dich zurück und schau der
          Demo zu.
        </p>

        <ol className="mt-8 flex flex-col gap-4">
          {explainSteps.map((step, index) => (
            <li key={step.title} className="flex items-start gap-3.5">
              <span
                className="font-brand flex h-[30px] w-[30px] shrink-0 items-center justify-center rounded-[9px] text-sm font-bold"
                style={{
                  background: `color-mix(in srgb, ${step.accent} 15%, transparent)`,
                  color: step.accent,
                }}
                aria-hidden="true"
              >
                {index + 1}
              </span>
              <span>
                <span className="block text-[15.5px] font-semibold">
                  {step.title}
                </span>
                <span className="block text-sm text-muted">{step.text}</span>
              </span>
            </li>
          ))}
        </ol>

        <button
          type="button"
          onClick={onReplay}
          className="btn btn-ghost btn-sm mt-7"
        >
          <RotateCcw size={17} aria-hidden="true" />
          Demo neu abspielen
        </button>
      </div>

      {/* Telefonspalte */}
      <div className="flex justify-center">
        <div ref={phoneRef} className="hero-phone demo-phone">
          <div className="hero-phone-notch" aria-hidden="true"></div>
          <div className="hero-phone-screen flex flex-col">
            {/* App-Kopf */}
            <div className="flex shrink-0 items-center justify-between gap-2.5 border-b border-card-border bg-background px-[18px] pt-4 pb-3">
              <div className="flex min-w-0 items-center gap-2.5">
                <img
                  src={jottiSymbol.src}
                  alt=""
                  width={16}
                  height={25}
                  className="block shrink-0"
                />
                <div className="min-w-0">
                  <div className="font-brand text-[15px] leading-tight font-bold whitespace-nowrap">
                    Tisch 12
                  </div>
                  <div className="text-[11.5px] leading-tight text-muted whitespace-nowrap">
                    Biergarten
                  </div>
                </div>
              </div>
              <div className="shrink-0 text-right">
                <div className="text-[10.5px] tracking-wide text-muted uppercase">
                  Offen
                </div>
                <div className="text-[15px] leading-tight font-bold whitespace-nowrap tabular-nums">
                  {totalStr}
                </div>
              </div>
            </div>

            {/* Kategorie-Pills (dekorativ) */}
            <div
              className="flex shrink-0 gap-2 px-[18px] pt-3 pb-1.5"
              aria-hidden="true"
            >
              <span className="inline-flex h-8 items-center rounded-full bg-foreground px-[15px] text-[13px] font-semibold text-background">
                Getränke
              </span>
              <span className="inline-flex h-8 items-center rounded-full border border-card-border px-[15px] text-[13px] font-medium">
                Speisen
              </span>
            </div>

            {/* Produktliste */}
            <div className="flex-1 overflow-y-auto px-[18px] pt-2 pb-3.5">
              {demoMenu.map((product) => (
                <div key={product.name} className="mt-3">
                  <div className="mb-[7px] text-xs font-semibold text-muted">
                    {product.name}
                  </div>
                  <div className="flex flex-col gap-[7px]">
                    {product.variants.map((variant) => {
                      const qty = state.cart[variant.id] ?? 0
                      const active = qty > 0
                      return (
                        <div
                          key={variant.id}
                          className="flex items-center justify-between rounded-[11px] border py-[9px] pr-[11px] pl-[13px]"
                          style={{
                            borderColor: active
                              ? 'color-mix(in srgb, var(--primary) 55%, var(--border))'
                              : 'var(--border)',
                            background: active
                              ? 'color-mix(in srgb, var(--primary) 8%, var(--card))'
                              : 'var(--card)',
                          }}
                        >
                          <div className="flex items-baseline gap-2.5">
                            <span className="text-[14.5px] font-medium">
                              {variant.name}
                            </span>
                            <span className="text-[13px] font-bold tabular-nums">
                              {formatEuro(variant.priceCents)}
                            </span>
                          </div>
                          <div className="flex items-center gap-2">
                            <button
                              type="button"
                              onClick={() =>
                                setState((s) => manualRemove(s, variant.id))
                              }
                              disabled={!active}
                              aria-label={`${product.name} ${variant.name} entfernen`}
                              className="flex h-[34px] w-[34px] items-center justify-center rounded-full border border-card-border text-foreground transition-opacity disabled:opacity-40"
                            >
                              <Minus size={15} aria-hidden="true" />
                            </button>
                            <span className="w-[18px] text-center text-[15px] font-bold tabular-nums">
                              {qty}
                            </span>
                            <button
                              type="button"
                              onClick={() =>
                                setState((s) => manualAdd(s, variant.id))
                              }
                              aria-label={`${product.name} ${variant.name} hinzufügen`}
                              className="flex h-[34px] w-[34px] items-center justify-center rounded-full bg-brand-solid text-white"
                            >
                              <Plus size={15} aria-hidden="true" />
                            </button>
                          </div>
                        </div>
                      )
                    })}
                  </div>
                </div>
              ))}
            </div>

            {/* Warenkorb-Leiste */}
            <div className="shrink-0 border-t border-card-border bg-background px-4 pt-3 pb-4">
              <button
                type="button"
                onClick={() => setState((s) => manualPay(s))}
                disabled={!hasCart}
                className="flex h-[50px] w-full items-center justify-between rounded-[13px] bg-brand-solid px-[18px] text-[15.5px] font-bold text-white transition-opacity disabled:opacity-45"
              >
                <span className="inline-flex items-center gap-2.5">
                  <Euro size={18} aria-hidden="true" />
                  Kassieren
                </span>
                <span className="tabular-nums">{totalStr}</span>
              </button>
            </div>

            {/* Kassieren-Overlay mit Erfolgsanimation */}
            {state.paid && (
              <div
                role="status"
                className="absolute inset-0 z-30 flex flex-col items-center justify-center gap-4 p-10 text-center backdrop-blur-sm"
                style={{
                  background: 'color-mix(in srgb, var(--bg) 92%, transparent)',
                }}
              >
                <div
                  className="demo-pop flex h-[88px] w-[88px] items-center justify-center rounded-full text-brand"
                  style={{
                    background:
                      'color-mix(in srgb, var(--primary) 16%, transparent)',
                  }}
                >
                  <Check size={46} strokeWidth={2.4} aria-hidden="true" />
                </div>
                <div>
                  <div className="font-brand text-[22px] font-bold">
                    Zahlung erfolgreich
                  </div>
                  <div className="mt-1.5 text-sm text-muted">
                    Beleg signiert · TSE bestätigt
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
