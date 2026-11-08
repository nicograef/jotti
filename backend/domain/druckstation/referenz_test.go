package druckstation

import "testing"

func TestReferenzAnzeige(t *testing.T) {
	cases := []struct{ referenz, want string }{
		{EventReferenz(ReferenzBestellung, 86), "Bestellung Nr. 86"},
		{EventReferenz(ReferenzZahlung, 12), "Zahlung Nr. 12"},
		{EventReferenz(ReferenzDirektverkauf, 3), "Direktverkauf Nr. 3"},
		{EventReferenz(ReferenzDirektverkaufStorno, 7), "Direktverkauf-Storno Nr. 7"},
		{EventReferenz(ReferenzStornierung, 41), "Stornierung Nr. 41"},
		{EventReferenz(ReferenzPfandrueckgabe, 42), "Pfandrückgabe Nr. 42"},
		{EventReferenz(ReferenzPfandrueckgabeStorno, 43), "Pfandrückgabe-Storno Nr. 43"},
		{"bestellung-aufgenommen:86", "Bestellung Nr. 86"},
		{TestdruckReferenz(KategorieGetraenk), "Testbon Getränk"},
		{"testdruck:kassenbeleg", "Testbon Kassenbeleg"},
		{"testdruck:foo", "Testbon foo"},
		{"unbekanntes-format:99", "unbekanntes-format:99"},
		{"ohne-doppelpunkt", "ohne-doppelpunkt"},
		{"zahlung-kassiert:", "zahlung-kassiert:"},
	}
	for _, tc := range cases {
		if got := ReferenzAnzeige(tc.referenz); got != tc.want {
			t.Errorf("ReferenzAnzeige(%q) = %q, want %q", tc.referenz, got, tc.want)
		}
	}
}

func TestDruckfehlerMeldung(t *testing.T) {
	cases := []struct{ fehler, want string }{
		{"drucker 192.168.1.5: nicht erreichbar: dial tcp", "Drucker nicht erreichbar"},
		{"drucker 192.168.1.5: papier leer (status=0x72)", "Papier leer"},
		{"drucker 192.168.1.5: senden fehlgeschlagen: EOF", "Übertragung fehlgeschlagen"},
		{"ungueltiges Base64: illegal data", "Druckfehler"},
	}
	for _, tc := range cases {
		if got := DruckfehlerMeldung(tc.fehler); got != tc.want {
			t.Errorf("DruckfehlerMeldung(%q) = %q, want %q", tc.fehler, got, tc.want)
		}
	}
}
