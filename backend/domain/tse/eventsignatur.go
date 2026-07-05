package tse

// EventSignatur ist der TSE-Stand eines Events aus der Signaturauftrags-
// Tabelle: der processType-Snapshot des Auftrags plus die Signatur, sobald der
// Signatur-Worker quittiert hat (nil solange unsigniert). Ein Event ohne
// Eintrag ist nicht signaturpflichtig — der Export kennt keine zweite Quelle
// und keine Projektion zur Lesezeit.
type EventSignatur struct {
	ProcessType string
	Signatur    *Signatur
}
