//go:build integration

package kassenjournal_repo

import (
	"context"
	"slices"
	"testing"

	"github.com/nicograef/jotti/backend/domain/kasse"
)

func TestReadPfandrueckgabeEvents_FiltertNachTischID(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	insert := func(eventType, subject string, version int, data map[string]any) int {
		t.Helper()
		id, err := insertEventRaw(repo.db, newTestEvent(userID, eventType, subject, version, data), ksNr)
		if err != nil {
			t.Fatalf("insert %s: %v", eventType, err)
		}
		return id
	}
	rueckgabe := func(rueckgabeID string, tischID int) map[string]any {
		data := validPfandrueckgabeData(rueckgabeID, 200, 1)
		data["tischId"] = tischID
		return data
	}

	insert("direktverkauf-getaetigt:v1", kasse.DirektverkaufSubject(ksNr, "verkauf-1"), 1, validDirektverkaufData("verkauf-1", 500))
	tisch3 := insert("pfand-zurueckgegeben:v1", kasse.PfandrueckgabeSubject(ksNr, "rueckgabe-3"), 1, rueckgabe("rueckgabe-3", 3))
	theke := insert("pfand-zurueckgegeben:v1", kasse.PfandrueckgabeSubject(ksNr, "rueckgabe-0"), 1, rueckgabe("rueckgabe-0", 0))
	insert("pfand-zurueckgegeben:v1", kasse.PfandrueckgabeSubject(ksNr, "rueckgabe-5"), 1, rueckgabe("rueckgabe-5", 5))
	storno := validPfandrueckgabeStornoData("rueckgabe-3", 200)
	storno["tischId"] = 3
	tisch3Storno := insert("pfandrueckgabe-storniert:v1", kasse.PfandrueckgabeSubject(ksNr, "rueckgabe-3"), 2, storno)

	for _, tc := range []struct {
		tischID int
		want    []int
	}{
		{tischID: 3, want: []int{tisch3, tisch3Storno}},
		{tischID: 0, want: []int{theke}},
		{tischID: 9, want: []int{}},
	} {
		events, err := repo.ReadPfandrueckgabeEvents(context.Background(), ksNr, tc.tischID)
		if err != nil {
			t.Fatalf("tisch %d: %v", tc.tischID, err)
		}
		got := make([]int, 0, len(events))
		for _, evt := range events {
			got = append(got, evt.ID)
		}
		if !slices.Equal(got, tc.want) {
			t.Errorf("tisch %d: event ids %v, erwartet %v", tc.tischID, got, tc.want)
		}
	}
}
