package tse

// EventSignatur has a nil Signatur while unsigned; an event without one is not subject to signing.
// See docs/handbuch.md §3.13 (Ein Leseweg).
type EventSignatur struct {
	ProcessType string
	Signatur    *Signatur
}
