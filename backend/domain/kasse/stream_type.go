package kasse

// StreamType identifies the type of event stream for routing writes to the correct projection.
type StreamType string

const (
	StreamTypeKassensitzung StreamType = "kassensitzung"
	StreamTypeTischSession  StreamType = "tisch-session"
)
