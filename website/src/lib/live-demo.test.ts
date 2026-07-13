import { describe, expect, it } from 'vitest'

import {
  addVariant,
  cartCount,
  cartTotalCents,
  demoScript,
  emptyCart,
  formatEuro,
  initialDemoState,
  manualAdd,
  manualPay,
  manualRemove,
  removeVariant,
  replay,
  runNextStep,
  startAuto,
  stepDelayMs,
} from './live-demo'

describe('Warenkorb', () => {
  it('erhöht die Menge einer Variante', () => {
    const cart = addVariant(addVariant(emptyCart, 1), 1)
    expect(cart[1]).toBe(2)
  })

  it('verringert die Menge und entfernt die Variante bei 0', () => {
    const cart = addVariant(emptyCart, 1)
    const removed = removeVariant(cart, 1)
    expect(removed[1]).toBeUndefined()
    expect(Object.keys(removed)).toHaveLength(0)
  })

  it('ignoriert das Verringern einer nicht enthaltenen Variante', () => {
    expect(removeVariant(emptyCart, 1)).toEqual(emptyCart)
  })

  it('behandelt Warenkörbe unveränderlich', () => {
    const cart = addVariant(emptyCart, 1)
    addVariant(cart, 1)
    expect(cart[1]).toBe(1)
  })
})

describe('Summenbildung in Cent', () => {
  it('summiert Preise mal Menge', () => {
    // 2× Bier 0,5 l (400) + 1× Pommes (300) = 1100
    const cart = addVariant(addVariant(addVariant(emptyCart, 1), 1), 5)
    expect(cartTotalCents(cart)).toBe(1100)
    expect(cartCount(cart)).toBe(3)
  })

  it('ist für den leeren Warenkorb 0', () => {
    expect(cartTotalCents(emptyCart)).toBe(0)
    expect(cartCount(emptyCart)).toBe(0)
  })
})

describe('formatEuro', () => {
  it('formatiert Cent deutsch', () => {
    expect(formatEuro(1450)).toBe('14,50\u00A0€')
    expect(formatEuro(400)).toBe('4,00\u00A0€')
    expect(formatEuro(0)).toBe('0,00\u00A0€')
  })
})

describe('Auto-Skriptablauf', () => {
  it('durchläuft das Handoff-Skript bis zur Endsumme 14,50\u00A0€ und schließt ab', () => {
    let state = startAuto(initialDemoState)
    expect(state.autoStatus).toBe('running')

    // 1. Bratwurst (350)
    state = runNextStep(state)
    expect(cartTotalCents(state.cart)).toBe(350)
    // 2. Bier 0,5 l (400)
    state = runNextStep(state)
    expect(cartTotalCents(state.cart)).toBe(750)
    // 3. Bier 0,5 l (400)
    state = runNextStep(state)
    expect(cartTotalCents(state.cart)).toBe(1150)
    // 4. Pommes (300) -> Endsumme
    state = runNextStep(state)
    expect(cartTotalCents(state.cart)).toBe(1450)
    expect(formatEuro(cartTotalCents(state.cart))).toBe('14,50\u00A0€')
    expect(state.paid).toBe(false)

    // 5. Kassieren -> Overlay, Warenkorb bleibt unter dem Overlay bestehen
    state = runNextStep(state)
    expect(state.paid).toBe(true)
    expect(cartTotalCents(state.cart)).toBe(1450)

    // 6. Reset -> leerer Warenkorb, Auto-Demo abgeschlossen
    state = runNextStep(state)
    expect(state.cart).toEqual({})
    expect(state.paid).toBe(false)
    expect(state.autoStatus).toBe('done')
  })

  it('liefert die Handoff-Timings je Schritt', () => {
    expect(demoScript).toHaveLength(6)
    expect(stepDelayMs(0)).toBe(700)
    expect(stepDelayMs(1)).toBe(820)
    expect(stepDelayMs(2)).toBe(820)
    expect(stepDelayMs(3)).toBe(820)
    expect(stepDelayMs(4)).toBe(1300) // pay
    expect(stepDelayMs(5)).toBe(1900) // reset
  })

  it('startet die Auto-Demo nur aus dem Ruhezustand', () => {
    const running = startAuto(initialDemoState)
    // erneutes Starten im running-Zustand ändert nichts
    expect(startAuto(running)).toBe(running)
  })

  it('ignoriert runNextStep, wenn die Auto-Demo nicht läuft', () => {
    expect(runNextStep(initialDemoState)).toBe(initialDemoState)
  })
})

describe('Permanenter Stopp bei manueller Interaktion', () => {
  it('stoppt die laufende Auto-Demo und nimmt sie nicht wieder auf', () => {
    let state = startAuto(initialDemoState)
    state = runNextStep(state) // ein Auto-Schritt gelaufen
    expect(state.autoStatus).toBe('running')

    // manuelle Interaktion
    state = manualAdd(state, 2)
    expect(state.autoStatus).toBe('stopped')

    // erneuter Viewport-Trigger darf nicht fortsetzen
    expect(startAuto(state).autoStatus).toBe('stopped')
    // ein ausstehender Auto-Timer richtet keinen Schaden an
    expect(runNextStep(state)).toBe(state)
  })

  it('stoppt auch bei manuellem Entfernen und Kassieren', () => {
    const withItem = runNextStep(startAuto(initialDemoState)) // ein Produkt gebucht
    expect(withItem.autoStatus).toBe('running')
    expect(manualRemove(withItem, 4).autoStatus).toBe('stopped')
    expect(manualPay(withItem).autoStatus).toBe('stopped')
  })

  it('bucht manuell hinzugefügte Produkte korrekt', () => {
    const state = manualAdd(manualAdd(initialDemoState, 1), 3)
    expect(cartTotalCents(state.cart)).toBe(750) // 400 + 350
  })
})

describe('Reset und Replay', () => {
  it('replay leert den Warenkorb und startet die Auto-Demo neu', () => {
    let state = manualAdd(initialDemoState, 1)
    state = manualPay(state)
    expect(state.autoStatus).toBe('stopped')

    const replayed = replay()
    expect(replayed.cart).toEqual({})
    expect(replayed.paid).toBe(false)
    expect(replayed.autoStatus).toBe('running')
    expect(replayed.step).toBe(0)
  })
})
