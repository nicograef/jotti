package kasse

import (
	"fmt"
	"strconv"
	"strings"
)

func KassensitzungSubject(zNr int) string {
	return "kassensitzung-" + strconv.Itoa(zNr)
}

func TischSessionSubject(zNr int, tischID int) string {
	return KassensitzungSubject(zNr) + "/tisch-" + strconv.Itoa(tischID)
}

func DirektverkaufSubject(zNr int, verkaufID string) string {
	return KassensitzungSubject(zNr) + "/direktverkauf-" + verkaufID
}

func ParseVerkaufIDFromSubject(subject string) (string, error) {
	const marker = "/direktverkauf-"
	idx := strings.LastIndex(subject, marker)
	if idx < 0 {
		return "", fmt.Errorf("invalid direktverkauf subject format: %s", subject)
	}
	return subject[idx+len(marker):], nil
}

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

func ParseZNrFromSubject(subject string) (int, error) {
	const prefix = "kassensitzung-"
	if !strings.HasPrefix(subject, prefix) {
		return 0, fmt.Errorf("invalid subject format: %s", subject)
	}
	rest := subject[len(prefix):]
	if idx := strings.Index(rest, "/"); idx >= 0 {
		rest = rest[:idx]
	}
	nr, err := strconv.Atoi(rest)
	if err != nil {
		return 0, fmt.Errorf("invalid z_nr in subject: %w", err)
	}
	return nr, nil
}
