package escpos

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/nicograef/jotti/backend/domain/kasse"
)

const lineWidth = 48 // Font A, 12×24 Dots bei 576 dots/line → 48 Zeichen

// FormatPositionBon generiert einen Bon für eine einzelne Position (Standard-Bonmodus).
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

	// Tisch — groß und fett, zentriert
	buf.WriteString(AlignCenter)
	buf.WriteString(TextDoubleAll)
	buf.WriteString(BoldOn)
	fmt.Fprintf(&buf, "== %s ==\n", tischName)
	buf.WriteString(BoldOff)
	buf.WriteString(TextNormal)
	buf.WriteString("\n")

	// Position — doppelte Höhe, fett, zentriert
	buf.WriteString(TextDoubleHigh)
	buf.WriteString(BoldOn)
	fmt.Fprintf(&buf, "%dx %s (%s)\n", pos.Menge, pos.ProduktName, pos.VarianteName)
	buf.WriteString(BoldOff)
	buf.WriteString(TextNormal)

	// Kommentar (optional) — fett, linksbündig
	if kommentar != "" {
		buf.WriteString("\n")
		buf.WriteString(AlignLeft)
		buf.WriteString(BoldOn)
		buf.WriteString(wrapLine(kommentar, lineWidth) + "\n")
		buf.WriteString(BoldOff)
	}

	// Trennlinie + Metadaten
	buf.WriteString(AlignLeft)
	buf.WriteString(strings.Repeat("-", lineWidth) + "\n")
	fmt.Fprintf(&buf, "  %s  Bedienung: %s\n", zeitpunkt.Format("15:04"), truncate(userName, 24))

	// 5 Leerzeilen vor dem Schnitt (Messer sitzt ~3mm über dem Druckkopf)
	buf.WriteString(strings.Repeat("\n", 5))
	buf.WriteString(CutPaper)

	return buf.Bytes()
}

// FormatSammelBon generiert einen Bon für alle Positionen einer Kategorie (optionaler Bonmodus).
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

	// Tisch — groß und fett, zentriert
	buf.WriteString(AlignCenter)
	buf.WriteString(TextDoubleAll)
	buf.WriteString(BoldOn)
	fmt.Fprintf(&buf, "== %s ==\n", tischName)
	buf.WriteString(BoldOff)
	buf.WriteString(TextNormal)
	buf.WriteString("\n")

	// Positionen — doppelte Höhe, fett, linksbündig
	buf.WriteString(AlignLeft)
	buf.WriteString(TextDoubleHigh)
	buf.WriteString(BoldOn)
	for _, pos := range positionen {
		artikel := fmt.Sprintf("%dx %s (%s)", pos.Menge, pos.ProduktName, pos.VarianteName)
		buf.WriteString(artikel + "\n")
	}
	buf.WriteString(BoldOff)
	buf.WriteString(TextNormal)

	// Kommentar (optional) — fett
	if kommentar != "" {
		buf.WriteString("\n")
		buf.WriteString(BoldOn)
		buf.WriteString(wrapLine(kommentar, lineWidth) + "\n")
		buf.WriteString(BoldOff)
	}

	// Trennlinie + Metadaten
	buf.WriteString(strings.Repeat("-", lineWidth) + "\n")
	fmt.Fprintf(&buf, "  %s  Bedienung: %s\n",
		zeitpunkt.Format("15:04"),
		truncate(userName, 24),
	)

	// 5 Leerzeilen vor dem Schnitt
	buf.WriteString(strings.Repeat("\n", 5))
	buf.WriteString(CutPaper)

	return buf.Bytes()
}

// truncate kürzt einen String auf maxLen Zeichen.
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
			result.WriteString(line + "\n")
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
