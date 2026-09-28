/**
 * Counts open Vorgänge (anything whose loss would annoy a helper) and notifies subscribers; the forced reload waits while it is non-zero.
 * Register only via `useOffenerVorgang`: a hand-held pair leaks, and a leaked count blocks the reload (docs/handbuch.md §6.8).
 */
class VorgangsRegister {
  private offen = 0
  private interessenten = new Set<() => void>()

  // Alle öffentlichen Methoden sind Pfeilfunktionen: `useSyncExternalStore`
  // nimmt sie ohne `this` entgegen und muss dieselbe Referenz wiedersehen.
  public anmelden = (): void => {
    this.offen += 1
    this.benachrichtigen()
  }

  public abmelden = (): void => {
    this.offen -= 1
    this.benachrichtigen()
  }

  public anzahlOffen = (): number => this.offen

  public abonnieren = (interessent: () => void): (() => void) => {
    this.interessenten.add(interessent)
    return () => {
      this.interessenten.delete(interessent)
    }
  }

  /** Setzt den Zähler zurück — ausschließlich für Tests. */
  public zuruecksetzen = (): void => {
    this.offen = 0
    this.benachrichtigen()
  }

  private benachrichtigen(): void {
    this.interessenten.forEach((interessent) => {
      interessent()
    })
  }
}

export const VorgangsRegisterSingleton = new VorgangsRegister()
