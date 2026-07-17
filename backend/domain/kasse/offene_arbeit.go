package kasse

import "sort"

// EigeneArbeitAnTisch ist die offene eigene Arbeit einer Servicekraft an einem
// einzelnen Tisch: ihre noch unbezahlten Positionen sowie das daraus abgeleitete
// "erledigt"-Kennzeichen. "Offen" bedeutet seit ADR 01 "noch nicht kassiert"
// (unbezahlt).
//
// Reines, DB-freies Deep Module: Eingabe ist eine Tisch-Session, Ausgabe ist die
// berechnete Sicht für genau eine Servicekraft. Der tischweite offene Saldo
// fließt bewusst nicht ein.
type EigeneArbeitAnTisch struct {
	// AnzahlOffen zählt die offenen (= unbezahlten) eigenen Positionen. Je
	// PositionID trägt UnbezahltePositionen höchstens einen Eintrag.
	AnzahlOffen int
	// OffenCents ist der noch offene (unbezahlte) Betrag der eigenen Positionen:
	// Summe aus EinzelpreisCents × Menge.
	OffenCents int
	// Erledigt ist true, wenn keine eigenen unbezahlten Positionen mehr offen sind.
	Erledigt bool
}

// ComputeEigeneArbeitAnTisch berechnet die offene eigene Arbeit der Servicekraft
// userID an der gegebenen Tisch-Session. Schichtübergabe ist implizit: sobald
// eine Kollegin eine eigene Position kassiert, verschwindet sie aus der
// Unbezahlt-Liste und zählt damit nicht mehr als offen.
func ComputeEigeneArbeitAnTisch(session TischSession, userID int) EigeneArbeitAnTisch {
	anzahlOffen := 0
	offenCents := 0
	for _, pos := range session.UnbezahltePositionen {
		if pos.BestellerUserID == userID {
			anzahlOffen++
			offenCents += pos.EinzelpreisCents * pos.Menge
		}
	}

	return EigeneArbeitAnTisch{
		AnzahlOffen: anzahlOffen,
		OffenCents:  offenCents,
		Erledigt:    anzahlOffen == 0,
	}
}

// OffeneArbeitTisch ist die offene eigene Arbeit einer Servicekraft an einem
// einzelnen Tisch, angereichert um die Tisch-ID für die Rollup-Liste.
type OffeneArbeitTisch struct {
	TischID     int
	AnzahlOffen int
	OffenCents  int
}

// OffeneArbeitRollup fasst die offene eigene Arbeit einer Servicekraft über
// mehrere Tisch-Sessions (i. d. R. alle einer offenen Kassensitzung) zusammen.
type OffeneArbeitRollup struct {
	// OffeneTische listet nur Tische mit offener eigener Arbeit (nicht erledigt),
	// aufsteigend nach Tisch-ID.
	OffeneTische []OffeneArbeitTisch
	// Erledigt ist true, wenn an keinem Tisch noch offene eigene Arbeit besteht.
	Erledigt bool
}

// ComputeOffeneArbeitRollup berechnet die offene eigene Arbeit der Servicekraft
// userID über alle gegebenen Tisch-Sessions. Tische, an denen für die Person
// alles erledigt ist, werden ausgelassen. Schichtübergabe ist implizit über
// ComputeEigeneArbeitAnTisch abgedeckt.
func ComputeOffeneArbeitRollup(sessions []TischSession, userID int) OffeneArbeitRollup {
	offeneTische := make([]OffeneArbeitTisch, 0)
	for _, session := range sessions {
		arbeit := ComputeEigeneArbeitAnTisch(session, userID)
		if arbeit.Erledigt {
			continue
		}
		offeneTische = append(offeneTische, OffeneArbeitTisch{
			TischID:     session.TischID,
			AnzahlOffen: arbeit.AnzahlOffen,
			OffenCents:  arbeit.OffenCents,
		})
	}

	sort.Slice(offeneTische, func(i, j int) bool {
		return offeneTische[i].TischID < offeneTische[j].TischID
	})

	return OffeneArbeitRollup{
		OffeneTische: offeneTische,
		Erledigt:     len(offeneTische) == 0,
	}
}

// OffeneArbeitServicekraft ist die offene eigene Arbeit einer Servicekraft über
// mehrere Tisch-Sessions, angereichert um ihre Identität: UserID und den
// eingefrorenen Besteller-Namen aus den Positionen.
type OffeneArbeitServicekraft struct {
	UserID       int
	UserName     string // eingefrorener Besteller-Name aus den Positionen
	OffeneTische []OffeneArbeitTisch
}

// ComputeOffeneArbeitProServicekraft berechnet die offene eigene Arbeit aller
// Servicekräfte, die in den Sessions noch unbezahlte Positionen haben.
// Servicekräfte ohne offene eigene Arbeit erscheinen nicht; das Ergebnis ist
// aufsteigend nach UserID sortiert. Schichtübergabe ist implizit über
// ComputeOffeneArbeitRollup abgedeckt.
func ComputeOffeneArbeitProServicekraft(sessions []TischSession) []OffeneArbeitServicekraft {
	nameByUserID := make(map[int]string)
	for _, session := range sessions {
		for _, pos := range session.UnbezahltePositionen {
			nameByUserID[pos.BestellerUserID] = pos.BestellerName
		}
	}

	servicekraefte := make([]OffeneArbeitServicekraft, 0, len(nameByUserID))
	for userID, name := range nameByUserID {
		rollup := ComputeOffeneArbeitRollup(sessions, userID)
		if rollup.Erledigt {
			continue
		}
		servicekraefte = append(servicekraefte, OffeneArbeitServicekraft{
			UserID:       userID,
			UserName:     name,
			OffeneTische: rollup.OffeneTische,
		})
	}

	sort.Slice(servicekraefte, func(i, j int) bool {
		return servicekraefte[i].UserID < servicekraefte[j].UserID
	})

	return servicekraefte
}
