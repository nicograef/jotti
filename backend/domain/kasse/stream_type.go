package kasse

type StreamType string

const (
	StreamTypeKassensitzung StreamType = "kassensitzung"
	StreamTypeTischSession  StreamType = "tisch-session"
	StreamTypeDirektverkauf StreamType = "direktverkauf"
	// StreamTypePfandrueckgabe has no projection: one stream per Rückgabe, version 1 is the
	// Rückgabe, each Stornierung a following version.
	StreamTypePfandrueckgabe StreamType = "pfandrueckgabe"
)
