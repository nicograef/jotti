//go:build unit

package escpos

import (
	"bytes"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/steuer"
)

// FuzzFormatKassenbeleg wirft beliebige (auch bösartige) Freitextinhalte in den
// ESC/POS-Encoder eines fiskalischen Kassenbelegs: Vereinsdaten, Artikeltexte und
// vor allem die TSE-QR-Payload, deren Länge in ein GS ( k Store-Kommando
// hineincodiert wird. Zu haltende Eigenschaften:
//   - kein Panic bei irgendeinem Eingabe-String;
//   - keine kaputte Steuersequenz: jedes GS ( k Store-Kommando deklariert in
//     seinen pL/pH-Bytes exakt die Zahl der nachfolgenden Nutzbytes, und diese
//     Bytes sind auch tatsächlich vorhanden (kein abgeschnittenes Kommando).
//
// Ein längenfehlerhaftes GS ( k würde den Drucker aus dem Tritt bringen und den
// Rest des Belegs als Rohbytes ausgeben.
func FuzzFormatKassenbeleg(f *testing.F) {
	f.Add("Musterverein e.V.", "Cola", "https://finanzamt.example/verify?d=abc", int64(350), false)
	f.Add("", "", "", int64(0), true)
	f.Add("Ünïcödé € \t\n\x1b\x40", "Bier\x00mit\x1dNull", "V0;kasse:1;\x1d\x28\x6b-payload", int64(-1234), false)

	f.Fuzz(func(t *testing.T, vereinsname, artikel, qrData string, preis int64, storno bool) {
		data := KassenbelegData{
			Vereinsname:        vereinsname,
			Strasse:            "Musterstr. 1",
			Plz:                "12345",
			Ort:                "Musterstadt",
			KassenSeriennummer: "KASSE-1",
			Belegnummer:        "42",
			Zeitpunkt:          time.Unix(0, 0).UTC(),
			Positionen: []kasse.Position{{
				ProduktName:      artikel,
				VarianteName:     "Standard",
				Steuersatz:       string(steuer.RegelSteuersatz),
				EinzelpreisCents: int(preis),
				Menge:            1,
			}},
			GesamtbetragCents: int(preis),
			Zahlungsart:       "Bar",
			Stornobeleg:       storno,
			TSE: &TSEAbschnitt{
				TransaktionNr:   1,
				Signaturzaehler: 1,
				TSESeriennummer: "TSE-1",
				ZeitpunktBeginn: time.Unix(0, 0).UTC(),
				ZeitpunktEnde:   time.Unix(0, 0).UTC(),
				Signatur:        "sig",
				QRCodeData:      qrData,
			},
		}

		out := FormatKassenbeleg(data)

		if !bytes.HasPrefix(out, []byte(Init)) {
			t.Fatalf("Beleg beginnt nicht mit Init: %q", out[:min(len(out), 8)])
		}
		assertQRCommandLengths(t, out)
	})
}

// assertQRCommandLengths läuft den ESC/POS-Bytestrom ab und prüft für jedes
// GS ( k Store-Kommando (GS 28 6B, pL, pH, cn, fn, m, <daten>), dass die in
// pL/pH deklarierte Nutzlast vollständig im Puffer liegt. Store-Kommandos
// (Funktion 80/0x50) tragen die eigentliche QR-Payload; ein Längenfehler dort
// verschöbe alle folgenden Bytes.
func assertQRCommandLengths(t *testing.T, out []byte) {
	t.Helper()
	prefix := []byte{0x1D, 0x28, 0x6B} // GS ( k
	for i := 0; i+5 <= len(out); {
		if !bytes.Equal(out[i:i+3], prefix) {
			i++
			continue
		}
		pL := int(out[i+3])
		pH := int(out[i+4])
		bodyLen := pL + pH*256
		if bodyLen < 0 {
			t.Fatalf("GS ( k mit negativer Länge bei Offset %d", i)
		}
		end := i + 5 + bodyLen
		if end > len(out) {
			t.Fatalf("GS ( k bei Offset %d deklariert %d Bytes, aber nur %d verfügbar", i, bodyLen, len(out)-(i+5))
		}
		// Weiter hinter diesem Kommando suchen.
		i = end
	}
}
