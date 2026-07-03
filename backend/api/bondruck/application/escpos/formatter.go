package escpos

import (
	"bytes"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

const lineWidth = 48 // Font A, 12x24 Dots bei 576 dots/line -> 48 Zeichen

type KassenbelegData struct {
	Vereinsname              string
	Strasse                  string
	Plz                      string
	Ort                      string
	KassenSeriennummer       string
	Belegnummer              string
	Zeitpunkt                time.Time
	ErsteBestellungZeitpunkt *time.Time
	Positionen               []kasse.Position
	Steuermatrix             []steuer.Aufteilung
	TSE                      *TSEAbschnitt
	TSEAusfallvermerk        bool
	GesamtbetragCents        int
	Zahlungsart              string
	// Stornobeleg switches the title to STORNOBELEG; amounts are expected pre-negated.
	Stornobeleg         bool
	StornoZuBelegnummer string
}

type TSEAbschnitt struct {
	TransaktionNr   int
	Signaturzaehler int
	TSESeriennummer string
	ZeitpunktBeginn time.Time
	ZeitpunktEnde   time.Time
	Signatur        string
	QRCodeData      string
	// Nachsigniert: die Signatur entstand nach einem TSE-Ausfall nachträglich
	// (Nachsignier-Worker). Der Beleg weist das aus, weil die TSE-Zeitpunkte
	// dann sichtbar vom Belegdatum abweichen.
	Nachsigniert bool
}

// FormatPositionBon generiert einen Bon fuer eine einzelne Position (Standard-Bonmodus).
func FormatPositionBon(
	pos kasse.Position,
	tischName string,
	userName string,
	zeitpunkt time.Time,
	kommentar string,
	withBeep bool,
) []byte {
	var buf bytes.Buffer

	if withBeep {
		buf.WriteString(Beep)
	}
	buf.WriteString(Init)
	buf.WriteString(SetCodepageWPC1252)

	// Tisch - gross und fett, zentriert
	buf.WriteString(AlignCenter)
	buf.WriteString(TextDoubleAll)
	buf.WriteString(BoldOn)
	buf.WriteString(toWPC1252(fmt.Sprintf("== %s ==\n", tischName)))
	buf.WriteString(BoldOff)
	buf.WriteString(TextNormal)
	buf.WriteString("\n")

	// Position - doppelte Hoehe, fett, zentriert
	buf.WriteString(TextDoubleHigh)
	buf.WriteString(BoldOn)
	buf.WriteString(toWPC1252(fmt.Sprintf("%dx %s\n", pos.Menge, pos.Bezeichnung())))
	buf.WriteString(BoldOff)
	buf.WriteString(TextNormal)

	// Kommentar (optional) - fett, linksbuendig
	if kommentar != "" {
		buf.WriteString("\n")
		buf.WriteString(AlignLeft)
		buf.WriteString(BoldOn)
		buf.WriteString(toWPC1252(wrapLine(kommentar, lineWidth)))
		buf.WriteByte('\n')
		buf.WriteString(BoldOff)
	}

	// Trennlinie + Metadaten
	buf.WriteString(AlignLeft)
	buf.WriteString(strings.Repeat("-", lineWidth))
	buf.WriteByte('\n')
	buf.WriteString(toWPC1252(fmt.Sprintf("  %s  Bedienung: %s\n", zeitpunkt.Format("15:04"), truncate(userName, 24))))

	// 5 Leerzeilen vor dem Schnitt (Messer sitzt ~3mm ueber dem Druckkopf)
	buf.WriteString(strings.Repeat("\n", 5))
	buf.WriteString(CutPaper)

	return buf.Bytes()
}

// FormatSammelBon generiert einen Bon fuer alle Positionen einer Kategorie (optionaler Bonmodus).
func FormatSammelBon(
	positionen []kasse.Position,
	tischName string,
	userName string,
	zeitpunkt time.Time,
	kommentar string,
	withBeep bool,
) []byte {
	var buf bytes.Buffer

	if withBeep {
		buf.WriteString(Beep)
	}
	buf.WriteString(Init)
	buf.WriteString(SetCodepageWPC1252)

	// Tisch - gross und fett, zentriert
	buf.WriteString(AlignCenter)
	buf.WriteString(TextDoubleAll)
	buf.WriteString(BoldOn)
	buf.WriteString(toWPC1252(fmt.Sprintf("== %s ==\n", tischName)))
	buf.WriteString(BoldOff)
	buf.WriteString(TextNormal)
	buf.WriteString("\n")

	// Positionen - doppelte Hoehe, fett, linksbuendig
	buf.WriteString(AlignLeft)
	buf.WriteString(TextDoubleHigh)
	buf.WriteString(BoldOn)
	for _, pos := range positionen {
		artikel := fmt.Sprintf("%dx %s", pos.Menge, pos.Bezeichnung())
		buf.WriteString(toWPC1252(artikel))
		buf.WriteByte('\n')
	}
	buf.WriteString(BoldOff)
	buf.WriteString(TextNormal)

	// Kommentar (optional) - fett
	if kommentar != "" {
		buf.WriteString("\n")
		buf.WriteString(BoldOn)
		buf.WriteString(toWPC1252(wrapLine(kommentar, lineWidth)))
		buf.WriteByte('\n')
		buf.WriteString(BoldOff)
	}

	// Trennlinie + Metadaten
	buf.WriteString(strings.Repeat("-", lineWidth))
	buf.WriteByte('\n')
	buf.WriteString(toWPC1252(fmt.Sprintf("  %s  Bedienung: %s\n",
		zeitpunkt.Format("15:04"),
		truncate(userName, 24),
	)))

	// 5 Leerzeilen vor dem Schnitt
	buf.WriteString(strings.Repeat("\n", 5))
	buf.WriteString(CutPaper)

	return buf.Bytes()
}

// FormatDirektverkaufAbholbon generates a single combined pickup ticket for a Direktverkauf.
// The header label is fixed to "Direktverkauf" and, like all Arbeitsbons, contains no prices.
func FormatDirektverkaufAbholbon(
	positionen []kasse.Position,
	userName string,
	zeitpunkt time.Time,
	kommentar string,
) []byte {
	return FormatSammelBon(positionen, "Direktverkauf", userName, zeitpunkt, kommentar, false)
}

// FormatKassenbeleg generiert einen fiskalischen Kassenbeleg.
func FormatKassenbeleg(data KassenbelegData) []byte {
	var buf bytes.Buffer

	buf.WriteString(Init)
	buf.WriteString(SetCodepageWPC1252)

	buf.WriteString(AlignCenter)
	buf.WriteString(BoldOn)
	if data.Stornobeleg {
		buf.WriteString("STORNOBELEG\n")
	} else {
		buf.WriteString("KASSENBELEG\n")
	}
	buf.WriteString(BoldOff)
	buf.WriteString("\n")

	buf.WriteString(toWPC1252(wrapLine(data.Vereinsname, lineWidth)))
	buf.WriteByte('\n')
	buf.WriteString(toWPC1252(wrapLine(data.Strasse, lineWidth)))
	buf.WriteByte('\n')
	buf.WriteString(toWPC1252(wrapLine(strings.TrimSpace(data.Plz+" "+data.Ort), lineWidth)))
	buf.WriteByte('\n')
	buf.WriteString("\n")

	buf.WriteString(AlignLeft)
	fmt.Fprintf(&buf, "Datum: %s\n", data.Zeitpunkt.Format("02.01.2006 15:04"))
	fmt.Fprintf(&buf, "Bon-Nr: %s\n", data.Belegnummer)
	if data.StornoZuBelegnummer != "" {
		fmt.Fprintf(&buf, "Storno zu Bon-Nr: %s\n", data.StornoZuBelegnummer)
	}
	fmt.Fprintf(&buf, "Kassen-ID: %s\n", data.KassenSeriennummer)
	if data.ErsteBestellungZeitpunkt != nil {
		fmt.Fprintf(&buf, "Erste Bestellung: %s\n", data.ErsteBestellungZeitpunkt.Format("02.01.2006 15:04:05"))
	}
	buf.WriteString(strings.Repeat("-", lineWidth))
	buf.WriteByte('\n')

	for _, pos := range data.Positionen {
		artikel := fmt.Sprintf("%dx %s", pos.Menge, pos.Bezeichnung())
		buf.WriteString(toWPC1252(wrapLine(artikel, lineWidth)))
		buf.WriteByte('\n')
		fmt.Fprintf(&buf, "  %s x %d = %s EUR (%s)\n", formatCents(pos.Einzelpreis), pos.Menge, formatCents(pos.Einzelpreis*pos.Menge), steuerKennzeichenAusPosition(pos.Steuersatz))
	}

	buf.WriteString(strings.Repeat("-", lineWidth))
	buf.WriteByte('\n')
	buf.WriteString(BoldOn)
	fmt.Fprintf(&buf, "GESAMT: %s EUR\n", formatCents(data.GesamtbetragCents))
	buf.WriteString(BoldOff)
	buf.WriteString(toWPC1252(fmt.Sprintf("Zahlungsart: %s\n", data.Zahlungsart)))

	if len(data.Steuermatrix) > 0 {
		buf.WriteByte('\n')
		buf.WriteString("Steueraufteilung:\n")
		for _, zeile := range data.Steuermatrix {
			fmt.Fprintf(
				&buf,
				"  %s: Netto %s EUR, Steuer %s EUR, Brutto %s EUR\n",
				steuerKennzeichenAusSatz(zeile.Satz),
				formatCents(zeile.Netto),
				formatCents(zeile.Steuer),
				formatCents(zeile.Brutto),
			)
		}
	}

	if data.TSE != nil {
		buf.WriteByte('\n')
		buf.WriteString("TSE-Daten:\n")
		fmt.Fprintf(&buf, "  TSE-Transaktion: %d\n", data.TSE.TransaktionNr)
		fmt.Fprintf(&buf, "  Signaturzaehler: %d\n", data.TSE.Signaturzaehler)
		fmt.Fprintf(&buf, "  TSE-Seriennummer: %s\n", data.TSE.TSESeriennummer)
		fmt.Fprintf(&buf, "  TSE-Start: %s\n", data.TSE.ZeitpunktBeginn.Format("02.01.2006 15:04:05"))
		fmt.Fprintf(&buf, "  TSE-Ende: %s\n", data.TSE.ZeitpunktEnde.Format("02.01.2006 15:04:05"))
		buf.WriteString(toWPC1252("  Signatur: "))
		buf.WriteString(toWPC1252(wrapLine(data.TSE.Signatur, lineWidth-2)))
		buf.WriteByte('\n')
		if data.TSE.Nachsigniert {
			fmt.Fprintf(&buf, toWPC1252("  Nachsigniert am %s\n"), data.TSE.ZeitpunktEnde.Format("02.01.2006 15:04:05"))
			buf.WriteString(toWPC1252("  (TSE war bei der Erfassung nicht erreichbar)\n"))
		}

		appendNativeQRCode(&buf, data.TSE.QRCodeData)
	} else if data.TSEAusfallvermerk {
		buf.WriteByte('\n')
		buf.WriteString("TSE-Hinweis:\n")
		buf.WriteString(toWPC1252("  TSE voruebergehend nicht erreichbar.\n"))
		buf.WriteString(toWPC1252("  Dieser Vorgang wird automatisch nachsigniert.\n"))
	}

	buf.WriteString("\n")
	buf.WriteString(AlignCenter)
	buf.WriteString("Vielen Dank!\n")

	buf.WriteString(strings.Repeat("\n", 5))
	buf.WriteString(CutPaper)

	return buf.Bytes()
}

// appendNativeQRCode writes a printer-rendered QR code using ESC/POS GS ( k.
func appendNativeQRCode(buf *bytes.Buffer, qrCodeData string) {
	data := strings.TrimSpace(qrCodeData)
	if data == "" {
		return
	}

	buf.WriteByte('\n')
	buf.WriteString(AlignCenter)
	buf.WriteString(QRCodeModel2)
	buf.WriteString(QRCodeModuleSize6)
	buf.WriteString(QRCodeErrorCorrectionM)

	payload := []byte(data)
	commandLen := len(payload) + 3
	pL := byte(commandLen % 256)
	pH := byte(commandLen / 256)

	buf.WriteString(QRCodeStorePrefix)
	buf.WriteByte(pL)
	buf.WriteByte(pH)
	buf.Write([]byte{0x31, 0x50, 0x30})
	buf.Write(payload)
	buf.WriteString(QRCodePrint)
	buf.WriteByte('\n')
	buf.WriteString(AlignLeft)
}

// toWPC1252 transkodiert sichtbaren Text von UTF-8 in die Drucker-Codepage WPC1252
// (Windows-1252), passend zu SetCodepageWPC1252 (ESC t 6 am MUNBYN ITPP047P).
// Nicht abbildbare Zeichen werden ersetzt, damit ein einzelnes Sonderzeichen nicht
// den ganzen Bon verwirft. ESC/POS-Steuerbytes werden nie hierdurch geschickt.
func toWPC1252(s string) string {
	encoded, err := encoding.ReplaceUnsupported(charmap.Windows1252.NewEncoder()).String(s)
	if err != nil {
		return s
	}
	return encoded
}

// truncate kuerzt einen String auf maxLen Runen (inkl. Auslassungszeichen)
// und schneidet dabei nie mitten in einer Rune.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}

// wrapLine bricht einen langen String an Wortgrenzen um (runenbasierte Breite).
func wrapLine(s string, width int) string {
	if utf8.RuneCountInString(s) <= width {
		return s
	}
	var result strings.Builder
	words := strings.Fields(s)
	line := ""
	for _, w := range words {
		if utf8.RuneCountInString(line)+1+utf8.RuneCountInString(w) > width && line != "" {
			result.WriteString(line)
			result.WriteByte('\n')
			line = w
		} else {
			if line == "" {
				line = w
			} else {
				line += " " + w
			}
		}
	}
	if line != "" {
		result.WriteString(line)
	}
	return result.String()
}

func formatCents(cents int) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%d,%02d", sign, cents/100, cents%100)
}

func steuerKennzeichenAusPosition(satz string) string {
	return steuerKennzeichenAusSatz(steuer.Steuersatz(satz))
}

func steuerKennzeichenAusSatz(satz steuer.Steuersatz) string {
	switch satz {
	case steuer.RegelSteuersatz:
		return "A"
	case steuer.ErmaessigtSteuersatz:
		return "B"
	case steuer.BefreitSteuersatz:
		return "C"
	case steuer.KombiSteuersatz:
		return "A/B"
	default:
		return "?"
	}
}
