/**
 * Zählt die gerade offenen Vorgänge und benachrichtigt Interessenten.
 *
 * „Offen" heißt: etwas, dessen Verlust eine Helferin ärgern würde — ein
 * gefüllter Korb, eine getroffene Auswahl, eine laufende Buchung, ein
 * angefangenes Formular. Der erzwungene Reload wartet, solange der Zähler nicht
 * null ist. Angemeldet wird ausschließlich über `useOffenerVorgang`: Ein von
 * Hand gehaltenes Paar leckt, und ein geleckter Zähler blockiert den Reload.
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
