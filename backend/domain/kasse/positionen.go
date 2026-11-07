package kasse

import (
	"slices"

	"github.com/google/uuid"
)

// summeCents is the gross total EinzelpreisCents × Menge.
func summeCents(positionen []Position) int {
	summe := 0
	for _, pos := range positionen {
		summe += pos.EinzelpreisCents * pos.Menge
	}
	return summe
}

// mitNeuenPositionIDs stamps a fresh PositionID on a copy of each position; the caller's slice stays untouched.
func mitNeuenPositionIDs(positionen []Position) []Position {
	out := slices.Clone(positionen)
	for i := range out {
		out[i].PositionID = uuid.New().String()
	}
	return out
}

// ResolvePositionen returns the requested positions as fat positions (name, category and
// price included, so the event is self-contained) and their total in cents. It returns false
// for an unknown PositionID, a Menge above the available one, or a repeated PositionID, whose
// quantities would otherwise add up unnoticed.
func ResolvePositionen(available []Position, requested []PositionRef) ([]Position, int, bool) {
	resolved := make([]Position, 0, len(requested))
	totalCents := 0
	seen := make(map[string]bool, len(requested))
	for _, ref := range requested {
		if seen[ref.PositionID] {
			return nil, 0, false
		}
		seen[ref.PositionID] = true
		idx := slices.IndexFunc(available, func(pos Position) bool { return pos.PositionID == ref.PositionID })
		if idx < 0 || ref.Menge > available[idx].Menge {
			return nil, 0, false
		}
		resolved = append(resolved, mitMenge(available[idx], ref.Menge))
		totalCents += available[idx].EinzelpreisCents * ref.Menge
	}
	return resolved, totalCents, true
}
