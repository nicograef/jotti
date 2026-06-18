package kasse

// EigeneArbeitAnTisch ist die offene eigene Arbeit einer Servicekraft an einem
// einzelnen Tisch: ihre noch ausstehenden (nicht ausgegebenen) und noch
// unbezahlten Positionen sowie das daraus abgeleitete "erledigt"-Kennzeichen.
//
// Reines, DB-freies Deep Module: Eingabe ist eine Tisch-Session, Ausgabe ist die
// berechnete Sicht für genau eine Servicekraft. Die tischweite ausstehende
// Auszahlung (negativer Saldo) fließt bewusst nicht ein.
type EigeneArbeitAnTisch struct {
	// AnzahlAusstehend ist die Anzahl eigener Positionen, die noch nicht ausgegeben sind.
	AnzahlAusstehend int
	// AnzahlUnbezahlt ist die Anzahl eigener Positionen, die noch nicht bezahlt sind.
	AnzahlUnbezahlt int
	// AnzahlOffen zählt die Vereinigung: eigene Positionen, die ausstehend ODER
	// unbezahlt sind, je Position (PositionID) einmal.
	AnzahlOffen int
	// Erledigt ist true, wenn keine eigenen ausstehenden UND keine eigenen
	// unbezahlten Positionen mehr offen sind.
	Erledigt bool
}

// ComputeEigeneArbeitAnTisch berechnet die offene eigene Arbeit der Servicekraft
// userID an der gegebenen Tisch-Session. Schichtübergabe ist implizit: sobald
// eine Kollegin eine eigene Position ausgibt oder kassiert, verschwindet sie aus
// der jeweiligen Liste und zählt damit nicht mehr als offen.
func ComputeEigeneArbeitAnTisch(session TischSession, userID int) EigeneArbeitAnTisch {
	offeneIDs := make(map[string]struct{})

	anzahlAusstehend := 0
	for _, pos := range session.AusstehendePositionen {
		if pos.BestellerUserID == userID {
			anzahlAusstehend++
			offeneIDs[pos.PositionID] = struct{}{}
		}
	}

	anzahlUnbezahlt := 0
	for _, pos := range session.UnbezahltePositionen {
		if pos.BestellerUserID == userID {
			anzahlUnbezahlt++
			offeneIDs[pos.PositionID] = struct{}{}
		}
	}

	return EigeneArbeitAnTisch{
		AnzahlAusstehend: anzahlAusstehend,
		AnzahlUnbezahlt:  anzahlUnbezahlt,
		AnzahlOffen:      len(offeneIDs),
		Erledigt:         anzahlAusstehend == 0 && anzahlUnbezahlt == 0,
	}
}
