package core

import "testing"

func TestParsePortOwnersSingleObject(t *testing.T) {
	in := []byte(`{ "LocalPort": 80, "OwningProcess": 4, "ProcessName": "System" }`)
	owners, err := ParsePortOwners(in)
	if err != nil {
		t.Fatalf("unerwarteter Fehler: %v", err)
	}
	if len(owners) != 1 {
		t.Fatalf("Anzahl: got %d, want 1", len(owners))
	}
	want := PortOwner{LocalPort: 80, PID: 4, ProcessName: "System"}
	if owners[0] != want {
		t.Fatalf("got %+v, want %+v", owners[0], want)
	}
}

func TestParsePortOwnersArray(t *testing.T) {
	in := []byte(`[
		{ "LocalPort": 443, "OwningProcess": 1234, "ProcessName": "vmware-hostd" },
		{ "LocalPort": 80, "OwningProcess": 4, "ProcessName": "System" }
	]`)
	owners, err := ParsePortOwners(in)
	if err != nil {
		t.Fatalf("unerwarteter Fehler: %v", err)
	}
	if len(owners) != 2 {
		t.Fatalf("Anzahl: got %d, want 2", len(owners))
	}
	want := PortOwner{LocalPort: 443, PID: 1234, ProcessName: "vmware-hostd"}
	if owners[0] != want {
		t.Fatalf("erster Eintrag: got %+v, want %+v", owners[0], want)
	}
}

func TestParsePortOwnersMissingProcessName(t *testing.T) {
	in := []byte(`{ "LocalPort": 80, "OwningProcess": 4, "ProcessName": null }`)
	owners, err := ParsePortOwners(in)
	if err != nil {
		t.Fatalf("unerwarteter Fehler: %v", err)
	}
	if len(owners) != 1 || owners[0].ProcessName != "" {
		t.Fatalf("nicht aufloesbarer ProcessName sollte als \"\" toleriert werden: %+v", owners)
	}
}

func TestParsePortOwnersEmptyIsNoError(t *testing.T) {
	owners, err := ParsePortOwners([]byte("   \n"))
	if err != nil {
		t.Fatalf("leere Ausgabe sollte keinen Fehler liefern: %v", err)
	}
	if len(owners) != 0 {
		t.Fatalf("Anzahl: got %d, want 0", len(owners))
	}
}

func TestParsePortOwnersBrokenJSON(t *testing.T) {
	if _, err := ParsePortOwners([]byte("{ kaputt ")); err == nil {
		t.Fatal("kaputtes JSON sollte einen Fehler liefern (-> generischer Fallback)")
	}
}

func TestParsePortOwnersToleratesBOM(t *testing.T) {
	in := append([]byte("\xef\xbb\xbf"),
		[]byte(`{ "LocalPort": 80, "OwningProcess": 4, "ProcessName": "System" }`)...)
	owners, err := ParsePortOwners(in)
	if err != nil {
		t.Fatalf("UTF-8-BOM sollte toleriert werden: %v", err)
	}
	if len(owners) != 1 {
		t.Fatalf("Anzahl: got %d, want 1", len(owners))
	}
}
