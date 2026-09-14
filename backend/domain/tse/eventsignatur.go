package tse

// EventSignatur ist der TSE-Stand eines Events aus der Signaturauftrags-Tabelle:
// processType-Snapshot plus Signatur, sobald der Worker quittiert hat (nil solange
// unsigniert). Ein Event ohne Eintrag ist nicht signaturpflichtig — der Export
// kennt keine zweite Quelle.
type EventSignatur struct {
	ProcessType string
	Signatur    *Signatur
}
