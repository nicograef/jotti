package escpos

import (
	"bytes"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/internal/zeit"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/unicode/norm"
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
	TSEVermerk               TSEBelegvermerk
	GesamtbetragCents        int
	Zahlungsart              string
	// Stornobeleg switches the title to STORNOBELEG; the amounts arrive with their final sign.
	Stornobeleg         bool
	StornoZuBelegnummer string
}

// TSEBelegvermerk ist der TSE-Hinweis auf einem Beleg ohne Signaturdaten.
type TSEBelegvermerk int

const (
	// KeinTSEVermerk: der Vorgang ist signiert oder nicht signaturpflichtig.
	KeinTSEVermerk TSEBelegvermerk = iota
	// TSEVermerkVoruebergehend: die TSE war bei der Erfassung nicht erreichbar;
	// der Vorgang wird automatisch nachsigniert.
	TSEVermerkVoruebergehend
	// TSEVermerkKeineKonfiguration: für den Vorgang war keine TSE konfiguriert;
	// er wird nicht nachsigniert.
	TSEVermerkKeineKonfiguration
	// TSEVermerkSignaturFehlgeschlagen: die Signatur ist endgültig fehlgeschlagen;
	// der Vorgang wird nicht automatisch nachsigniert.
	TSEVermerkSignaturFehlgeschlagen
)

type TSEAbschnitt struct {
	TransaktionNr   int
	Signaturzaehler int
	TSESeriennummer string
	ZeitpunktBeginn time.Time
	ZeitpunktEnde   time.Time
	Signatur        string
	QRCodeData      string
	// Nachsigniert: die Signatur entstand nach einem TSE-Ausfall nachträglich
	// (Signatur-Worker). Der Beleg weist das aus, weil die TSE-Zeitpunkte
	// dann sichtbar vom Belegdatum abweichen.
	Nachsigniert bool
}

func FormatPositionBon(
	pos kasse.Position,
	tischName string,
	userName string,
	zeitpunkt time.Time,
	kommentar string,
) []byte {
	var buf bytes.Buffer

	writeArbeitsbonKopf(&buf, tischName)

	buf.WriteString(TextDoubleHigh)
	buf.WriteString(BoldOn)
	writeLine(&buf, fmt.Sprintf("%dx %s", pos.Menge, pos.Bezeichnung()))
	buf.WriteString(BoldOff)
	buf.WriteString(TextNormal)

	if pos.Kommentar != "" {
		buf.WriteString(AlignLeft)
		buf.WriteString(BoldOn)
		writePositionsKommentar(&buf, pos.Kommentar)
		buf.WriteString(BoldOff)
	}

	if kommentar != "" {
		buf.WriteString("\n")
		buf.WriteString(AlignLeft)
		buf.WriteString(BoldOn)
		buf.WriteString(wrapCP437(kommentar, lineWidth))
		buf.WriteByte('\n')
		buf.WriteString(BoldOff)
	}

	buf.WriteString(AlignLeft)
	writeArbeitsbonFuss(&buf, zeitpunkt, userName)

	return buf.Bytes()
}

func FormatSammelBon(
	positionen []kasse.Position,
	tischName string,
	userName string,
	zeitpunkt time.Time,
	kommentar string,
) []byte {
	var buf bytes.Buffer

	writeArbeitsbonKopf(&buf, tischName)

	buf.WriteString(AlignLeft)
	buf.WriteString(TextDoubleHigh)
	buf.WriteString(BoldOn)
	for _, pos := range positionen {
		artikel := fmt.Sprintf("%dx %s", pos.Menge, pos.Bezeichnung())
		writeLine(&buf, artikel)
		if pos.Kommentar != "" {
			buf.WriteString(TextNormal)
			writePositionsKommentar(&buf, pos.Kommentar)
			buf.WriteString(TextDoubleHigh)
		}
	}
	buf.WriteString(BoldOff)
	buf.WriteString(TextNormal)

	if kommentar != "" {
		buf.WriteString("\n")
		buf.WriteString(BoldOn)
		buf.WriteString(wrapCP437(kommentar, lineWidth))
		buf.WriteByte('\n')
		buf.WriteString(BoldOff)
	}

	writeArbeitsbonFuss(&buf, zeitpunkt, userName)

	return buf.Bytes()
}

// writePositionsKommentar writes the Positionskommentar under its position line, each wrapped
// line indented by two spaces; the caller sets size, weight and alignment.
func writePositionsKommentar(buf *bytes.Buffer, kommentar string) {
	for zeile := range strings.SplitSeq(wrapCP437(kommentar, lineWidth-2), "\n") {
		buf.WriteString("  ")
		buf.WriteString(zeile)
		buf.WriteByte('\n')
	}
}

// writeArbeitsbonKopf leaves the alignment centered; each Arbeitsbon sets its own for the positions.
func writeArbeitsbonKopf(buf *bytes.Buffer, tischName string) {
	buf.WriteString(Init)
	buf.WriteString(SetCodepagePC437)

	buf.WriteString(AlignCenter)
	buf.WriteString(TextDoubleAll)
	buf.WriteString(BoldOn)
	writeLine(buf, fmt.Sprintf("== %s ==", tischName))
	buf.WriteString(BoldOff)
	buf.WriteString(TextNormal)
	buf.WriteString("\n")
}

// writeArbeitsbonFuss writes no alignment: the caller has set AlignLeft already.
func writeArbeitsbonFuss(buf *bytes.Buffer, zeitpunkt time.Time, userName string) {
	buf.WriteString(strings.Repeat("-", lineWidth))
	buf.WriteByte('\n')
	writeLine(buf, fmt.Sprintf("  %s  Bedienung: %s", zeitpunkt.In(zeit.Berlin).Format("15:04"), truncate(userName, 24)))

	// 5 Leerzeilen vor dem Schnitt (Messer sitzt ~3mm über dem Druckkopf)
	buf.WriteString(strings.Repeat("\n", 5))
	buf.WriteString(CutPaper)
}

// FormatDirektverkaufAbholbon carries no prices, like every Arbeitsbon; the
// header label is fixed to "Direktverkauf".
func FormatDirektverkaufAbholbon(
	positionen []kasse.Position,
	userName string,
	zeitpunkt time.Time,
	kommentar string,
) []byte {
	return FormatSammelBon(positionen, "Direktverkauf", userName, zeitpunkt, kommentar)
}

// FormatTestbon trägt Stationsname, Zeitstempel und eine Umlautzeile, an der
// der Admin die Codepage prüft — kein fiskalischer Inhalt, keine TSE-Daten.
func FormatTestbon(stationsName string, zeitpunkt time.Time) []byte {
	var buf bytes.Buffer

	buf.WriteString(Init)
	buf.WriteString(SetCodepagePC437)

	buf.WriteString(AlignCenter)
	buf.WriteString(TextDoubleAll)
	buf.WriteString(BoldOn)
	writeLine(&buf, "TESTBON")
	buf.WriteString(BoldOff)
	buf.WriteString(TextNormal)
	buf.WriteString("\n")

	writeLine(&buf, fmt.Sprintf("Station: %s", stationsName))
	writeLine(&buf, zeitpunkt.In(zeit.Berlin).Format("02.01.2006 15:04:05"))
	buf.WriteString("\n")
	writeLine(&buf, "Drucker und Netzwerk funktionieren.")
	writeLine(&buf, "Umlaute: äöü ÄÖÜ ß")

	buf.WriteString(AlignLeft)
	buf.WriteString(strings.Repeat("\n", 5))
	buf.WriteString(CutPaper)

	return buf.Bytes()
}

func FormatKassenbeleg(data KassenbelegData) []byte {
	var buf bytes.Buffer

	writeBelegKopf(&buf, data)
	writeBelegPositionen(&buf, data.Positionen)
	writeBelegSumme(&buf, data)
	writeSteuermatrix(&buf, data.Steuermatrix)
	if data.TSE != nil {
		writeTSEAbschnitt(&buf, *data.TSE)
	} else {
		writeTSEVermerk(&buf, data.TSEVermerk)
	}

	buf.WriteString("\n")
	buf.WriteString(AlignCenter)
	buf.WriteString("Vielen Dank!\n")

	buf.WriteString(strings.Repeat("\n", 5))
	buf.WriteString(CutPaper)

	return buf.Bytes()
}

// writeBelegKopf writes the title, the Betreiber address and the Beleg identifiers (§ 6 KassenSichV).
func writeBelegKopf(buf *bytes.Buffer, data KassenbelegData) {
	buf.WriteString(Init)
	buf.WriteString(SetCodepagePC437)

	buf.WriteString(AlignCenter)
	buf.WriteString(BoldOn)
	if data.Stornobeleg {
		buf.WriteString("STORNOBELEG\n")
	} else {
		buf.WriteString("KASSENBELEG\n")
	}
	buf.WriteString(BoldOff)
	buf.WriteString("\n")

	buf.WriteString(wrapCP437(data.Vereinsname, lineWidth))
	buf.WriteByte('\n')
	buf.WriteString(wrapCP437(data.Strasse, lineWidth))
	buf.WriteByte('\n')
	buf.WriteString(wrapCP437(strings.TrimSpace(data.Plz+" "+data.Ort), lineWidth))
	buf.WriteByte('\n')
	buf.WriteString("\n")

	buf.WriteString(AlignLeft)
	fmt.Fprintf(buf, "Datum: %s\n", data.Zeitpunkt.In(zeit.Berlin).Format("02.01.2006 15:04"))
	fmt.Fprintf(buf, "Bon-Nr: %s\n", data.Belegnummer)
	if data.StornoZuBelegnummer != "" {
		fmt.Fprintf(buf, "Storno zu Bon-Nr: %s\n", data.StornoZuBelegnummer)
	}
	fmt.Fprintf(buf, "Kassen-ID: %s\n", data.KassenSeriennummer)
	if data.ErsteBestellungZeitpunkt != nil {
		fmt.Fprintf(buf, "Erste Bestellung: %s\n", data.ErsteBestellungZeitpunkt.In(zeit.Berlin).Format("02.01.2006 15:04:05"))
	}
	buf.WriteString(strings.Repeat("-", lineWidth))
	buf.WriteByte('\n')
}

func writeBelegPositionen(buf *bytes.Buffer, positionen []kasse.Position) {
	for _, pos := range positionen {
		artikel := fmt.Sprintf("%dx %s", pos.Menge, pos.Bezeichnung())
		buf.WriteString(wrapCP437(artikel, lineWidth))
		buf.WriteByte('\n')
		kennzeichen := steuerKennzeichenAusSatz(steuer.Steuersatz(pos.Steuersatz))
		fmt.Fprintf(buf, "  %s x %d = %s EUR (%s)\n", formatCents(pos.EinzelpreisCents), pos.Menge, formatCents(pos.EinzelpreisCents*pos.Menge), kennzeichen)
	}
}

func writeBelegSumme(buf *bytes.Buffer, data KassenbelegData) {
	buf.WriteString(strings.Repeat("-", lineWidth))
	buf.WriteByte('\n')
	buf.WriteString(BoldOn)
	fmt.Fprintf(buf, "GESAMT: %s EUR\n", formatCents(data.GesamtbetragCents))
	buf.WriteString(BoldOff)
	writeLine(buf, fmt.Sprintf("Zahlungsart: %s", data.Zahlungsart))
}

func writeSteuermatrix(buf *bytes.Buffer, steuermatrix []steuer.Aufteilung) {
	if len(steuermatrix) == 0 {
		return
	}
	buf.WriteByte('\n')
	buf.WriteString("Steueraufteilung:\n")
	for _, zeile := range steuermatrix {
		fmt.Fprintf(
			buf,
			"  %s: Netto %s EUR, Steuer %s EUR, Brutto %s EUR\n",
			steuerMatrixLabel(zeile.Satz),
			formatCents(zeile.Netto),
			formatCents(zeile.Steuer),
			formatCents(zeile.Brutto),
		)
	}
}

func writeTSEAbschnitt(buf *bytes.Buffer, tse TSEAbschnitt) {
	buf.WriteByte('\n')
	buf.WriteString("TSE-Daten:\n")
	fmt.Fprintf(buf, "  TSE-Transaktion: %d\n", tse.TransaktionNr)
	fmt.Fprintf(buf, "  Signaturzaehler: %d\n", tse.Signaturzaehler)
	fmt.Fprintf(buf, "  TSE-Seriennummer: %s\n", tse.TSESeriennummer)
	fmt.Fprintf(buf, "  TSE-Start: %s\n", tse.ZeitpunktBeginn.In(zeit.Berlin).Format("02.01.2006 15:04:05"))
	fmt.Fprintf(buf, "  TSE-Ende: %s\n", tse.ZeitpunktEnde.In(zeit.Berlin).Format("02.01.2006 15:04:05"))
	buf.WriteString(toCP437("  Signatur: "))
	buf.WriteString(wrapCP437(tse.Signatur, lineWidth-2))
	buf.WriteByte('\n')
	if tse.Nachsigniert {
		writeLine(buf, fmt.Sprintf("  Nachsigniert am %s", tse.ZeitpunktEnde.In(zeit.Berlin).Format("02.01.2006 15:04:05")))
		writeLine(buf, "  (TSE war bei der Erfassung nicht erreichbar)")
	}

	appendNativeQRCode(buf, tse.QRCodeData)
}

func writeTSEVermerk(buf *bytes.Buffer, vermerk TSEBelegvermerk) {
	switch vermerk {
	case TSEVermerkVoruebergehend:
		buf.WriteByte('\n')
		buf.WriteString("TSE-Hinweis:\n")
		writeLine(buf, "  TSE voruebergehend nicht erreichbar.")
		writeLine(buf, "  Dieser Vorgang wird automatisch nachsigniert.")
	case TSEVermerkKeineKonfiguration:
		buf.WriteByte('\n')
		buf.WriteString("TSE-Hinweis:\n")
		writeLine(buf, "  Fuer diesen Vorgang war keine TSE konfiguriert.")
	case TSEVermerkSignaturFehlgeschlagen:
		buf.WriteByte('\n')
		buf.WriteString("TSE-Hinweis:\n")
		writeLine(buf, "  Signatur fehlgeschlagen.")
		writeLine(buf, "  Vorgang ohne TSE-Signatur erfasst.")
	case KeinTSEVermerk:
	}
}

func appendNativeQRCode(buf *bytes.Buffer, qrCodeData string) {
	data := strings.TrimSpace(qrCodeData)
	if data == "" {
		return
	}

	buf.WriteByte('\n')
	buf.WriteString(AlignCenter)
	buf.WriteString(QRCodeModel2)
	buf.WriteString(QRCodeModuleSizeCmdPrefix)
	buf.WriteByte(qrModuleSizeByte(len(data)))
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

// asciiErsatz bildet Zeichen, die CP437 fehlen und die Handy-Tastaturen oder
// truncate erzeugen, auf lesbares ASCII ab.
var asciiErsatz = map[rune]string{
	'„': `"`, '“': `"`, '”': `"`, '″': `"`,
	'‚': "'", '‘': "'", '’': "'", '′': "'",
	'‐': "-", '‑': "-", '‒': "-", '–': "-", '—': "-", '−': "-",
	'…': "...",
	'€': "EUR",
	'×': "x",
	'•': "*",
	'→': "->",
	'§': "Par.",
	'³': "3",
	'©': "(c)", '®': "(R)", '™': "TM",
	'\u2007': " ", '\u2009': " ", '\u202F': " ",
}

// printable reduces UTF-8 text to runes CP437 can encode, so callers count print columns
// in runes. Accented letters lose only their accent; "?" marks a rune with no readable form.
// Control bytes would reach the printer as ESC/POS commands: whitespace becomes a space,
// every other C0 control and DEL is dropped.
func printable(s string) string {
	var out strings.Builder
	for _, r := range norm.NFC.String(s) {
		if r < 0x20 || r == 0x7F {
			if unicode.IsSpace(r) {
				out.WriteByte(' ')
			}
			continue
		}
		if _, ok := charmap.CodePage437.EncodeRune(r); ok {
			out.WriteRune(r)
			continue
		}
		if ersatz, ok := asciiErsatz[r]; ok {
			out.WriteString(ersatz)
			continue
		}
		if unicode.In(r, unicode.Mn, unicode.Cf) {
			continue
		}
		basis, _ := utf8.DecodeRuneInString(norm.NFD.String(string(r)))
		if _, ok := charmap.CodePage437.EncodeRune(basis); ok {
			out.WriteRune(basis)
			continue
		}
		out.WriteByte('?')
	}
	return out.String()
}

// toCP437 encodes visible text for the printer codepage set by SetCodepagePC437.
// ESC/POS control sequences never pass through here; a line break in s prints as a space.
func toCP437(s string) string {
	return encodeCP437(printable(s))
}

func writeLine(buf *bytes.Buffer, s string) {
	buf.WriteString(toCP437(s))
	buf.WriteByte('\n')
}

// wrapCP437 wraps after the transliteration, which can lengthen a line ("…" → "...").
func wrapCP437(s string, width int) string {
	return encodeCP437(wrapLine(printable(s), width))
}

// encodeCP437 expects the output of printable, plus the line breaks of wrapLine.
func encodeCP437(s string) string {
	var out strings.Builder
	for _, r := range s {
		b, _ := charmap.CodePage437.EncodeRune(r)
		out.WriteByte(b)
	}
	return out.String()
}

// truncate kürzt einen String auf maxLen Runen (inkl. Auslassungszeichen)
// und schneidet dabei nie mitten in einer Rune.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}

func wrapLine(s string, width int) string {
	if utf8.RuneCountInString(s) <= width {
		return s
	}
	var result strings.Builder
	words := strings.Fields(s)
	line := ""
	for _, w := range words {
		if line == "" {
			line = w
			continue
		}
		if utf8.RuneCountInString(line)+1+utf8.RuneCountInString(w) > width {
			result.WriteString(line)
			result.WriteByte('\n')
			line = w
			continue
		}
		line += " " + w
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

// steuerMatrixLabel prints the rate or the exemption notice (§ 6 Satz 1 Nr. 5 KassenSichV).
// steuer.Steuermatrix() already splits Kombi positions, so no Kombi row reaches it.
func steuerMatrixLabel(satz steuer.Steuersatz) string {
	switch satz {
	case steuer.RegelSteuersatz:
		return "A (19 %)"
	case steuer.ErmaessigtSteuersatz:
		return "B (7 %)"
	case steuer.BefreitSteuersatz:
		return "C (umsatzsteuerfrei)"
	case steuer.KombiSteuersatz:
		fallthrough
	default:
		return "?"
	}
}

// qrVersionForLengthM returns the smallest QR version (1-40) whose level-M byte-mode capacity
// holds payloadLen bytes. Data codewords for level M from ISO/IEC 18004, minus the mode indicator
// and character count: 2 bytes for V1-V9, 3 bytes for V10-V40.
func qrVersionForLengthM(payloadLen int) int {
	codewords := [40]int{
		16, 28, 44, 64, 86, 108, 124, 154, 182, 216, // V1-V10
		254, 290, 334, 365, 415, 453, 507, 563, 627, 669, // V11-V20
		714, 782, 860, 914, 1000, 1062, 1128, 1193, 1267, 1373, // V21-V30
		1455, 1541, 1631, 1725, 1812, 1914, 1992, 2102, 2216, 2334, // V31-V40
	}
	for i, n := range codewords {
		overhead := 2
		if i >= 9 {
			overhead = 3
		}
		if n-overhead >= payloadLen {
			return i + 1
		}
	}
	return 40
}

// qrModuleSizeByte returns the ESC/POS module-size byte (n in GS ( k 1 67 n)
// that keeps the QR symbol plus its 4-module quiet zone within 576 printable
// dots on an 80 mm thermal printer. Error correction level M is assumed.
func qrModuleSizeByte(payloadLen int) byte {
	v := qrVersionForLengthM(payloadLen)
	// QR matrix: (4*v + 17) modules. Quiet zone: 4 modules on each side.
	totalModules := 4*v + 17 + 8
	switch {
	case 6*totalModules <= 576:
		return 6
	case 5*totalModules <= 576:
		return 5
	case 4*totalModules <= 576:
		return 4
	default:
		return 3
	}
}
