package kasse

import "sort"

// EigeneArbeitAnTisch ist die offene eigene Arbeit einer Servicekraft an einem einzelnen
// Tisch. "Offen" bedeutet "noch nicht kassiert" (siehe docs/decisions.md D01); der
// tischweite offene Saldo fließt bewusst nicht ein.
type EigeneArbeitAnTisch struct {
	// AnzahlOffen zählt die eigenen unbezahlten Positionen (je PositionID höchstens eine).
	AnzahlOffen int
	// OffenCents ist die Summe aus EinzelpreisCents × Menge der eigenen unbezahlten Positionen.
	OffenCents int
	Erledigt   bool
}

// ComputeEigeneArbeitAnTisch: Schichtübergabe ist implizit — kassiert eine Kollegin eine
// eigene Position, verschwindet sie aus der Unbezahlt-Liste und zählt nicht mehr als offen.
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

type OffeneArbeitTisch struct {
	TischID     int
	AnzahlOffen int
	OffenCents  int
}

type OffeneArbeitRollup struct {
	// OffeneTische listet nur Tische mit offener eigener Arbeit, aufsteigend nach Tisch-ID.
	OffeneTische []OffeneArbeitTisch
	Erledigt     bool
}

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

type OffeneArbeitServicekraft struct {
	UserID       int
	UserName     string // eingefrorener Besteller-Name aus den Positionen
	OffeneTische []OffeneArbeitTisch
}

// ComputeOffeneArbeitProServicekraft liefert nur Servicekräfte mit offener eigener Arbeit, aufsteigend nach UserID.
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
