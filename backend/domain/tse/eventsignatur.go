package tse

// EventSignatur has a nil Signatur while unsigned; an event without one is not subject to signing.
// Status is the Signaturauftrag status. See docs/handbuch.md §3.13 (Ein Leseweg).
type EventSignatur struct {
	ProcessType string
	Status      string
	Signatur    *Signatur
}
