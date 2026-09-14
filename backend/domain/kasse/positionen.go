package kasse

// ValidatePositionRefs rejects duplicate PositionIDs — their quantities would add up unnoticed.
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

// ResolvePositionen returns the requested positions as fat positions (name, category and
// price included, so the event is self-contained) and their total in cents.
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
