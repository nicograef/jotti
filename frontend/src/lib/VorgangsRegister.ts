/**
 * Zählt die gerade offenen Vorgänge und benachrichtigt Interessenten bei jeder
 * Änderung.
 *
 * „Offen" heißt: etwas, dessen Verlust eine Helferin ärgern würde — ein
 * gefüllter Korb, eine getroffene Auswahl, eine laufende Buchung, ein
 * angefangenes Formular. Der erzwungene Reload wartet, solange der Zähler nicht
 * null ist.
 *
 * Angemeldet wird ausschließlich über `useOffenerVorgang`, das sich im
 * Effekt-Cleanup wieder abmeldet; gelesen wird über `useAnzahlOffeneVorgaenge`.
 * Ein von Hand gehaltenes Paar aus An- und Abmeldung leckt früher oder später,
 * und ein geleckter Zähler blockiert den Reload dauerhaft, ohne dass es
 * jemandem auffällt.
 */
class VorgangsRegister {
  private offen = 0
  private interessenten = new Set<() => void>()

  // Alle öffentlichen Methoden sind Pfeilfunktionen: `useSyncExternalStore`
  // nimmt `abonnieren` und `anzahlOffen` als lose Referenzen entgegen (ohne
  // `this`) und muss bei jedem Rendern dieselbe Referenz sehen.
  public anmelden = (): void => {
    this.offen += 1
    this.benachrichtigen()
  }

  public abmelden = (): void => {
    this.offen -= 1
    this.benachrichtigen()
  }

  public anzahlOffen = (): number => this.offen

  /** Meldet jede Änderung des Zählers; das Ergebnis beendet das Abo. */
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
