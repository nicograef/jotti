// UI-freies Logik-Modul der Live-Demo (#demo) der Landing.
// Kapselt Warenkorb (Menge je Variante), Summenbildung in Cent, den
// Auto-Demo-Ablauf als deterministische Schrittfolge, den permanenten Stopp
// bei manueller Interaktion und den Reset. Kein DOM, keine Timer, keine React-
// Abhängigkeit — die LiveDemo-Island (src/components/LiveDemo.tsx) rendert
// diesen Zustand und liefert Timing (setTimeout) und Viewport-Trigger.
//
// Menü, Preise (in Cent), Auto-Skript und Timings stammen aus dem Handoff-
// Prototyp (PRD docs/prds/prd-website-redesign.md).
//
// Geldregel des Projekts: Beträge sind immer int in Cent, niemals Floats.

export interface DemoVariant {
  id: number
  name: string
  priceCents: number
}

export interface DemoProduct {
  name: string
  variants: DemoVariant[]
}

// Handoff-Menü: Bier 0,5 l / 0,3 l, Weinschorle, Bratwurst, Pommes.
export const demoMenu: readonly DemoProduct[] = [
  {
    name: 'Bier',
    variants: [
      { id: 1, name: '0,5 l', priceCents: 400 },
      { id: 2, name: '0,3 l', priceCents: 300 },
    ],
  },
  {
    name: 'Weinschorle',
    variants: [{ id: 3, name: '0,25 l', priceCents: 350 }],
  },
  {
    name: 'Bratwurst',
    variants: [{ id: 4, name: 'im Weck', priceCents: 350 }],
  },
  {
    name: 'Pommes',
    variants: [{ id: 5, name: 'Portion', priceCents: 300 }],
  },
]

const priceByVariantId = new Map<number, number>(
  demoMenu.flatMap((product) =>
    product.variants.map((variant) => [variant.id, variant.priceCents]),
  ),
)

export function variantPriceCents(id: number): number {
  return priceByVariantId.get(id) ?? 0
}

// ---- Warenkorb (Menge je Variante) ----

export type Cart = Readonly<Record<number, number>>

export const emptyCart: Cart = {}

export function addVariant(cart: Cart, id: number): Cart {
  return { ...cart, [id]: (cart[id] ?? 0) + 1 }
}

export function removeVariant(cart: Cart, id: number): Cart {
  const current = cart[id] ?? 0
  if (current <= 0) return cart
  const next: Record<number, number> = { ...cart }
  if (current === 1) delete next[id]
  else next[id] = current - 1
  return next
}

export function cartTotalCents(cart: Cart): number {
  return Object.entries(cart).reduce(
    (sum, [id, qty]) => sum + variantPriceCents(Number(id)) * qty,
    0,
  )
}

export function cartCount(cart: Cart): number {
  return Object.values(cart).reduce((sum, qty) => sum + qty, 0)
}

// Deutsche Betragsformatierung aus Cent, z. B. 1450 -> "14,50 €" (geschütztes
// Leerzeichen vor dem Euro-Zeichen wie im Prototyp, damit Betrag und Zeichen
// nicht umbrechen).
export function formatEuro(cents: number): string {
  return `${(cents / 100).toFixed(2).replace('.', ',')}\u00A0€`
}

// ---- Auto-Demo als deterministische Schrittfolge ----

export type DemoStep =
  | { action: 'add'; variantId: number }
  | { action: 'pay' }
  | { action: 'reset' }

// Handoff-Ablauf: Bratwurst, 2× Bier 0,5 l, Pommes -> Kassieren -> Reset.
// Endsumme vor dem Kassieren: 350 + 400 + 400 + 300 = 1450 Cent = 14,50 €.
export const demoScript: readonly DemoStep[] = [
  { action: 'add', variantId: 4 },
  { action: 'add', variantId: 1 },
  { action: 'add', variantId: 1 },
  { action: 'add', variantId: 5 },
  { action: 'pay' },
  { action: 'reset' },
]

// Wartezeit vor dem Ausführen des Schritts mit diesem Index (Handoff-Timings).
export function stepDelayMs(index: number): number {
  const step = demoScript[index]
  if (!step) return 0
  if (step.action === 'pay') return 1300
  if (step.action === 'reset') return 1900
  return index === 0 ? 700 : 820
}

// ---- Zustandsautomat der gesamten Demo ----

// idle: noch nicht gestartet · running: Auto-Demo läuft · stopped: dauerhaft
// aus (manuelle Interaktion) · done: Skript vollständig abgespielt.
export type AutoStatus = 'idle' | 'running' | 'stopped' | 'done'

export interface DemoState {
  cart: Cart
  paid: boolean
  autoStatus: AutoStatus
  step: number
}

export const initialDemoState: DemoState = {
  cart: emptyCart,
  paid: false,
  autoStatus: 'idle',
  step: 0,
}

function applyStep(state: DemoState, step: DemoStep): DemoState {
  switch (step.action) {
    case 'add':
      return {
        ...state,
        cart: addVariant(state.cart, step.variantId),
        paid: false,
      }
    case 'pay':
      return { ...state, paid: true }
    case 'reset':
      return { ...state, cart: emptyCart, paid: false }
  }
}

// Startet die Auto-Demo nur aus dem Ruhezustand. Nach manuellem Stopp
// (stopped) oder Abschluss (done) bleibt sie dauerhaft aus — auch wenn die
// Sektion erneut in den Viewport scrollt.
export function startAuto(state: DemoState): DemoState {
  if (state.autoStatus !== 'idle') return state
  return { ...state, autoStatus: 'running', step: 0 }
}

// Führt den nächsten Auto-Schritt aus. No-op, wenn die Auto-Demo nicht läuft
// (spiegelt den `if (!demoAuto) return`-Guard des Prototyps): ein noch
// ausstehender Timer richtet nach einem manuellen Stopp keinen Schaden an.
export function runNextStep(state: DemoState): DemoState {
  if (state.autoStatus !== 'running') return state
  const step = demoScript[state.step]
  if (!step) return { ...state, autoStatus: 'done' }
  const nextStep = state.step + 1
  return {
    ...applyStep(state, step),
    step: nextStep,
    autoStatus: nextStep >= demoScript.length ? 'done' : 'running',
  }
}

// Manuelle Interaktionen stoppen die Auto-Demo dauerhaft.
export function manualAdd(state: DemoState, id: number): DemoState {
  return {
    ...state,
    cart: addVariant(state.cart, id),
    paid: false,
    autoStatus: 'stopped',
  }
}

export function manualRemove(state: DemoState, id: number): DemoState {
  return {
    ...state,
    cart: removeVariant(state.cart, id),
    autoStatus: 'stopped',
  }
}

export function manualPay(state: DemoState): DemoState {
  return { ...state, paid: true, autoStatus: 'stopped' }
}

// „Demo neu abspielen": setzt den Warenkorb zurück und startet die Auto-Demo
// erneut — der einzige Weg zurück in den running-Zustand nach einem Stopp.
export function replay(): DemoState {
  return { cart: emptyCart, paid: false, autoStatus: 'running', step: 0 }
}
