package kasse

import (
	"cmp"
	"slices"
)

// HatEigeneOffenePositionen: "offen" bedeutet "noch nicht kassiert" (siehe docs/decisions.md D01);
// der tischweite offene Saldo fließt bewusst nicht ein. Schichtübergabe ist implizit — kassiert
// eine Kollegin eine eigene Position, verschwindet sie aus der Unbezahlt-Liste.
func HatEigeneOffenePositionen(session TischSession, userID int) bool {
	return slices.ContainsFunc(session.UnbezahltePositionen, func(pos Position) bool {
		return pos.BestellerUserID == userID
	})
}

type OffeneArbeitTisch struct {
	TischID int
	// OffenCents ist die Summe aus EinzelpreisCents × Menge der eigenen unbezahlten Positionen.
	OffenCents int
}

type OffeneArbeitServicekraft struct {
	UserID   int
	UserName string // eingefrorener Besteller-Name aus den Positionen
	// OffeneTische listet nur Tische mit offener eigener Arbeit, aufsteigend nach Tisch-ID.
	OffeneTische []OffeneArbeitTisch
}

// ComputeOffeneArbeitProServicekraft liefert nur Servicekräfte mit offener eigener Arbeit, aufsteigend nach UserID.
func ComputeOffeneArbeitProServicekraft(sessions []TischSession) []OffeneArbeitServicekraft {
	type userTisch struct{ userID, tischID int }
	offenCents := map[userTisch]int{}
	namen := map[int]string{}
	for _, session := range sessions {
		for _, pos := range session.UnbezahltePositionen {
			namen[pos.BestellerUserID] = pos.BestellerName
			offenCents[userTisch{pos.BestellerUserID, session.TischID}] += pos.EinzelpreisCents * pos.Menge
		}
	}

	tischeJeUser := map[int][]OffeneArbeitTisch{}
	for schluessel, cents := range offenCents {
		tischeJeUser[schluessel.userID] = append(tischeJeUser[schluessel.userID], OffeneArbeitTisch{TischID: schluessel.tischID, OffenCents: cents})
	}

	servicekraefte := make([]OffeneArbeitServicekraft, 0, len(tischeJeUser))
	for userID, tische := range tischeJeUser {
		slices.SortFunc(tische, func(a, b OffeneArbeitTisch) int { return cmp.Compare(a.TischID, b.TischID) })
		servicekraefte = append(servicekraefte, OffeneArbeitServicekraft{UserID: userID, UserName: namen[userID], OffeneTische: tische})
	}
	slices.SortFunc(servicekraefte, func(a, b OffeneArbeitServicekraft) int { return cmp.Compare(a.UserID, b.UserID) })

	return servicekraefte
}
