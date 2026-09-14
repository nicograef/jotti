//go:build unit

package escpos

import (
	"bytes"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/steuer"
)

// FuzzFormatKassenbeleg wirft beliebige Freitexte in den ESC/POS-Encoder, vor
// allem die TSE-QR-Payload, deren Länge in ein GS ( k Store-Kommando codiert
// wird. Zu halten: kein Panic bei irgendeinem Eingabe-String, und jedes GS ( k
// Store-Kommando deklariert in pL/pH exakt die Zahl der nachfolgenden, auch
// tatsächlich vorhandenen Nutzbytes. Ein längenfehlerhaftes GS ( k brächte den
// Drucker aus dem Tritt und gäbe den Rest des Belegs als Rohbytes aus.
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

// Eine QR-Payload, die selbst mit den GS ( k Prefix-Bytes beginnt, darf die
// Längenprüfung nicht fehlleiten: Der Beleg ist strukturell korrekt und muss
// assertQRCommandLengths ohne Fehlalarm passieren.
func TestAssertQRCommandLengths_PayloadMitPrefix(t *testing.T) {
	// Payload beginnt mit GS ( k und enthält den Prefix auch in der Mitte.
	qr := "\x1D\x28\x6B" + "V0;kasse:1;" + "\x1D\x28\x6B" + "x"
	data := KassenbelegData{
		Vereinsname:        "Musterverein e.V.",
		Strasse:            "Musterstr. 1",
		Plz:                "12345",
		Ort:                "Musterstadt",
		KassenSeriennummer: "KASSE-1",
		Belegnummer:        "42",
		Zeitpunkt:          time.Unix(0, 0).UTC(),
		Positionen: []kasse.Position{{
			ProduktName:      "Cola",
			VarianteName:     "Standard",
			Steuersatz:       string(steuer.RegelSteuersatz),
			EinzelpreisCents: 350,
			Menge:            1,
		}},
		GesamtbetragCents: 350,
		Zahlungsart:       "Bar",
		TSE: &TSEAbschnitt{
			TransaktionNr:   1,
			Signaturzaehler: 1,
			TSESeriennummer: "TSE-1",
			ZeitpunktBeginn: time.Unix(0, 0).UTC(),
			ZeitpunktEnde:   time.Unix(0, 0).UTC(),
			Signatur:        "sig",
			QRCodeData:      qr,
		},
	}

	out := FormatKassenbeleg(data)
	assertQRCommandLengths(t, out)

	// Gegenprobe: Ein manuell verstümmeltes Store-Kommando (überlange Längenangabe)
	// muss von assertQRCommandLengths erkannt werden — sonst prüft die Probe nichts.
	corrupt := verstuemmeleStoreLaenge(out)
	rec := &fatalRecorder{}
	assertQRCommandLengths(rec, corrupt)
	if !rec.failed {
		t.Fatal("verstümmeltes Store-Kommando wurde nicht erkannt")
	}
}

// qrAsserter ist die von assertQRCommandLengths genutzte Teilmenge von
// testing.TB (*testing.T erfüllt sie). Über diese Schnittstelle kann die
// Gegenprobe Fatalf-Aufrufe abfangen, ohne den Test zu beenden.
type qrAsserter interface {
	Helper()
	Fatalf(format string, args ...any)
}

type fatalRecorder struct {
	failed bool
}

func (r *fatalRecorder) Helper()               {}
func (r *fatalRecorder) Fatalf(string, ...any) { r.failed = true }

// verstuemmeleStoreLaenge sucht das erste echte GS ( k Store-Kommando (cn=0x31,
// fn=0x50) und setzt seine deklarierte Länge (pL/pH) auf einen zu großen Wert,
// sodass die Nutzlast über das Pufferende hinausragt.
func verstuemmeleStoreLaenge(out []byte) []byte {
	prefix := []byte{0x1D, 0x28, 0x6B}
	corrupt := make([]byte, len(out))
	copy(corrupt, out)
	for i := 0; i+7 <= len(corrupt); i++ {
		if bytes.Equal(corrupt[i:i+3], prefix) && corrupt[i+5] == 0x31 && corrupt[i+6] == 0x50 {
			corrupt[i+3] = 0xFF
			corrupt[i+4] = 0xFF
			break
		}
	}
	return corrupt
}

// assertQRCommandLengths läuft den ESC/POS-Bytestrom ab und prüft für jedes
// GS ( k Kommando (GS 28 6B, pL, pH, cn, fn, ...), dass die in pL/pH deklarierte
// Nutzlast vollständig im Puffer liegt; ein Längenfehler im Store-Kommando
// (Funktion 0x50, es trägt die QR-Payload) verschöbe alle folgenden Bytes.
func assertQRCommandLengths(t qrAsserter, out []byte) {
	t.Helper()
	prefix := []byte{0x1D, 0x28, 0x6B} // GS ( k
	const qrFunctionClass = 0x31       // cn=49; alle QR-Kommandos (GS ( k) tragen dies
	for i := 0; i+6 <= len(out); {
		if !bytes.Equal(out[i:i+3], prefix) {
			i++
			continue
		}
		// cn (out[i+5]) unterscheidet ein echtes QR-Kommando von Prefix-Bytes in
		// der Nutzlast. Nur bei cn=0x31 wird die Längenangabe interpretiert.
		if out[i+5] != qrFunctionClass {
			i++
			continue
		}
		pL := int(out[i+3])
		pH := int(out[i+4])
		bodyLen := pL + pH*256
		end := i + 5 + bodyLen
		if end > len(out) {
			t.Fatalf("GS ( k bei Offset %d deklariert %d Bytes, aber nur %d verfügbar", i, bodyLen, len(out)-(i+5))
		}
		// Über das komplette Kommando (inkl. Store-Nutzlast) springen, damit
		// Prefix-Bytes innerhalb der QR-Payload nicht als Kommando fehlgelesen werden.
		i = end
	}
}
