package kasse

// ValidatePositionRefs checks that every requested PositionRef exists in the available
// positions, that no PositionID is referenced more than once (duplicates would add up
// unnoticed), and that the requested Menge does not exceed the available Menge.
func ValidatePositionRefs(available []Position, requested []PositionRef) bool {
	seen := make(map[string]bool, len(requested))
	for _, ref := range requested {
		if seen[ref.PositionID] {
			return false
		}
		seen[ref.PositionID] = true
		found := false
		for _, pos := range available {
			if pos.PositionID == ref.PositionID {
				if ref.Menge > pos.Menge {
					return false
				}
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// ResolvePositionen resolves the requested PositionRefs to fat Positions using the
// available positions, returning the resolved positions and their total in cents.
// The resolved positions carry name, category and price so the resulting event is
// self-contained (fat).
func ResolvePositionen(available []Position, requested []PositionRef) ([]Position, int) {
	resolved := make([]Position, 0, len(requested))
	totalCents := 0
	for _, ref := range requested {
		for _, pos := range available {
			if pos.PositionID == ref.PositionID {
				resolved = append(resolved, Position{
					PositionID:       pos.PositionID,
					VarianteID:       pos.VarianteID,
					ProduktName:      pos.ProduktName,
					VarianteName:     pos.VarianteName,
					Kategorie:        pos.Kategorie,
					Steuersatz:       pos.Steuersatz,
					EinzelpreisCents: pos.EinzelpreisCents,
					Menge:            ref.Menge,
				})
				totalCents += pos.EinzelpreisCents * ref.Menge
				break
			}
		}
	}
	return resolved, totalCents
}
