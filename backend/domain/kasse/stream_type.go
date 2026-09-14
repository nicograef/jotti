package kasse

type StreamType string

const (
	StreamTypeKassensitzung StreamType = "kassensitzung"
	StreamTypeTischSession  StreamType = "tisch-session"
	StreamTypeDirektverkauf StreamType = "direktverkauf"
)
