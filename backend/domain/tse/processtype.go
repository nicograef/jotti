package tse

// Offizielle processType-Werte nach DSFinV-K Anhang I.
// Nur Kassenbeleg und Bestellung tragen das "-V1"-Suffix;
// der dritte Typ heißt "SonstigerVorgang" (ohne Suffix).
const (
	ProcessTypeKassenbelegV1    = "Kassenbeleg-V1"
	ProcessTypeBestellungV1     = "Bestellung-V1"
	ProcessTypeSonstigerVorgang = "SonstigerVorgang"
)
