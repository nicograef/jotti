package escpos

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/nicograef/jotti/backend/domain/kasse"
)

const lineWidth = 48 // Font A, 12x24 Dots bei 576 dots/line -> 48 Zeichen

type KassenbelegData struct {
	Vereinsname        string
	Strasse            string
	Plz                string
	Ort                string
	KassenSeriennummer string
	Belegnummer        string
	Zeitpunkt          time.Time
	Positionen         []kasse.Position
	GesamtbetragCents  int
	Zahlungsart        string
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

	// Tisch - gross und fett, zentriert
	buf.WriteString(AlignCenter)
	buf.WriteString(TextDoubleAll)
	buf.WriteString(BoldOn)
	fmt.Fprintf(&buf, "== %s ==\n", tischName)
	buf.WriteString(BoldOff)
	buf.WriteString(TextNormal)
	buf.WriteString("\n")

	// Position - doppelte Hoehe, fett, zentriert
	buf.WriteString(TextDoubleHigh)
	buf.WriteString(BoldOn)
	fmt.Fprintf(&buf, "%dx %s (%s)\n", pos.Menge, pos.ProduktName, pos.VarianteName)
	buf.WriteString(BoldOff)
	buf.WriteString(TextNormal)

	// Kommentar (optional) - fett, linksbuendig
	if kommentar != "" {
		buf.WriteString("\n")
		buf.WriteString(AlignLeft)
		buf.WriteString(BoldOn)
		buf.WriteString(wrapLine(kommentar, lineWidth))
		buf.WriteByte('\n')
		buf.WriteString(BoldOff)
	}

	// Trennlinie + Metadaten
	buf.WriteString(AlignLeft)
	buf.WriteString(strings.Repeat("-", lineWidth))
	buf.WriteByte('\n')
	fmt.Fprintf(&buf, "  %s  Bedienung: %s\n", zeitpunkt.Format("15:04"), truncate(userName, 24))

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

	// Tisch - gross und fett, zentriert
	buf.WriteString(AlignCenter)
	buf.WriteString(TextDoubleAll)
	buf.WriteString(BoldOn)
	fmt.Fprintf(&buf, "== %s ==\n", tischName)
	buf.WriteString(BoldOff)
	buf.WriteString(TextNormal)
	buf.WriteString("\n")

	// Positionen - doppelte Hoehe, fett, linksbuendig
	buf.WriteString(AlignLeft)
	buf.WriteString(TextDoubleHigh)
	buf.WriteString(BoldOn)
	for _, pos := range positionen {
		artikel := fmt.Sprintf("%dx %s (%s)", pos.Menge, pos.ProduktName, pos.VarianteName)
		buf.WriteString(artikel)
		buf.WriteByte('\n')
	}
	buf.WriteString(BoldOff)
	buf.WriteString(TextNormal)

	// Kommentar (optional) - fett
	if kommentar != "" {
		buf.WriteString("\n")
		buf.WriteString(BoldOn)
		buf.WriteString(wrapLine(kommentar, lineWidth))
		buf.WriteByte('\n')
		buf.WriteString(BoldOff)
	}

	// Trennlinie + Metadaten
	buf.WriteString(strings.Repeat("-", lineWidth))
	buf.WriteByte('\n')
	fmt.Fprintf(&buf, "  %s  Bedienung: %s\n",
		zeitpunkt.Format("15:04"),
		truncate(userName, 24),
	)

	// 5 Leerzeilen vor dem Schnitt
	buf.WriteString(strings.Repeat("\n", 5))
	buf.WriteString(CutPaper)

	return buf.Bytes()
}

// FormatKassenbeleg generiert einen fiskalischen Kassenbeleg (Basisstand ohne Steuer/TSE).
func FormatKassenbeleg(data KassenbelegData) []byte {
	var buf bytes.Buffer

	buf.WriteString(Init)

	buf.WriteString(AlignCenter)
	buf.WriteString(BoldOn)
	buf.WriteString("KASSENBELEG\n")
	buf.WriteString(BoldOff)
	buf.WriteString("\n")

	buf.WriteString(wrapLine(data.Vereinsname, lineWidth))
	buf.WriteByte('\n')
	buf.WriteString(wrapLine(data.Strasse, lineWidth))
	buf.WriteByte('\n')
	buf.WriteString(wrapLine(strings.TrimSpace(data.Plz+" "+data.Ort), lineWidth))
	buf.WriteByte('\n')
	buf.WriteString("\n")

	buf.WriteString(AlignLeft)
	fmt.Fprintf(&buf, "Datum: %s\n", data.Zeitpunkt.Format("02.01.2006 15:04"))
	fmt.Fprintf(&buf, "Bon-Nr: %s\n", data.Belegnummer)
	fmt.Fprintf(&buf, "Kassen-ID: %s\n", data.KassenSeriennummer)
	buf.WriteString(strings.Repeat("-", lineWidth))
	buf.WriteByte('\n')

	for _, pos := range data.Positionen {
		artikel := fmt.Sprintf("%dx %s (%s)", pos.Menge, pos.ProduktName, pos.VarianteName)
		buf.WriteString(wrapLine(artikel, lineWidth))
		buf.WriteByte('\n')
		fmt.Fprintf(&buf, "  %s x %d = %s EUR\n", formatCents(pos.Einzelpreis), pos.Menge, formatCents(pos.Einzelpreis*pos.Menge))
	}

	buf.WriteString(strings.Repeat("-", lineWidth))
	buf.WriteByte('\n')
	buf.WriteString(BoldOn)
	fmt.Fprintf(&buf, "GESAMT: %s EUR\n", formatCents(data.GesamtbetragCents))
	buf.WriteString(BoldOff)
	fmt.Fprintf(&buf, "Zahlungsart: %s\n", data.Zahlungsart)

	// TODO(F-07): Steueraufteilung sobald Positionen einen Steuersatz enthalten.
	// TODO(F-02): TSE-Signatur und processType nach TSE-Integration ergaenzen.

	buf.WriteString("\n")
	buf.WriteString(AlignCenter)
	buf.WriteString("Vielen Dank!\n")

	buf.WriteString(strings.Repeat("\n", 5))
	buf.WriteString(CutPaper)

	return buf.Bytes()
}

// truncate kuerzt einen String auf maxLen Zeichen.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

// wrapLine bricht einen langen String an Wortgrenzen um.
func wrapLine(s string, width int) string {
	if len(s) <= width {
		return s
	}
	var result strings.Builder
	words := strings.Fields(s)
	line := ""
	for _, w := range words {
		if len(line)+1+len(w) > width && line != "" {
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
	euros := cents / 100
	rest := cents % 100
	if rest < 0 {
		rest = -rest
	}
	return fmt.Sprintf("%d,%02d", euros, rest)
}
