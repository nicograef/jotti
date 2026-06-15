package kasse

import (
	"fmt"
	"strconv"
	"strings"
)

// KassensitzungSubject constructs the subject for a Kassensitzung event stream.
func KassensitzungSubject(zNr int) string {
	return "kassensitzung-" + strconv.Itoa(zNr)
}

// TischSessionSubject constructs the subject for a Tisch-Session event stream.
func TischSessionSubject(zNr int, tischID int) string {
	return KassensitzungSubject(zNr) + "/tisch-" + strconv.Itoa(tischID)
}

// DirektverkaufSubject constructs the subject for a Direktverkauf event stream.
// Each Direktverkauf is its own stream identified by a UUID.
func DirektverkaufSubject(zNr int, verkaufID string) string {
	return KassensitzungSubject(zNr) + "/direktverkauf-" + verkaufID
}

// ParseVerkaufIDFromSubject extracts the verkaufID from a subject like
// "kassensitzung-1/direktverkauf-<uuid>".
func ParseVerkaufIDFromSubject(subject string) (string, error) {
	const marker = "/direktverkauf-"
	idx := strings.LastIndex(subject, marker)
	if idx < 0 {
		return "", fmt.Errorf("invalid direktverkauf subject format: %s", subject)
	}
	return subject[idx+len(marker):], nil
}

// ParseTischIDFromSubject extracts the tischID from a subject like "kassensitzung-1/tisch-42".
func ParseTischIDFromSubject(subject string) (int, error) {
	const marker = "/tisch-"
	idx := strings.LastIndex(subject, marker)
	if idx < 0 {
		return 0, fmt.Errorf("invalid tisch-session subject format: %s", subject)
	}
	id, err := strconv.Atoi(subject[idx+len(marker):])
	if err != nil {
		return 0, fmt.Errorf("invalid tisch ID in subject: %w", err)
	}
	return id, nil
}

// ParseZNrFromSubject extracts the z_nr from a subject like "kassensitzung-1" or "kassensitzung-1/tisch-42".
func ParseZNrFromSubject(subject string) (int, error) {
	const prefix = "kassensitzung-"
	if !strings.HasPrefix(subject, prefix) {
		return 0, fmt.Errorf("invalid subject format: %s", subject)
	}
	rest := subject[len(prefix):]
	// If there's a "/" separator, take only the part before it
	if idx := strings.Index(rest, "/"); idx >= 0 {
		rest = rest[:idx]
	}
	nr, err := strconv.Atoi(rest)
	if err != nil {
		return 0, fmt.Errorf("invalid z_nr in subject: %w", err)
	}
	return nr, nil
}
