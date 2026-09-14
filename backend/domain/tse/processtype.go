package tse

// Offizielle processType-Werte nach DSFinV-K Anhang I. Nur Kassenbeleg und
// Bestellung tragen das "-V1"-Suffix, SonstigerVorgang nicht.
const (
	ProcessTypeKassenbelegV1    = "Kassenbeleg-V1"
	ProcessTypeBestellungV1     = "Bestellung-V1"
	ProcessTypeSonstigerVorgang = "SonstigerVorgang"
)
