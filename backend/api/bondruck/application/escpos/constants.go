package escpos

// Initialisierung
const Init = "\x1B\x40"

// Zeichentabelle (ESC t n)
// SetCodepageCP858 waehlt Codepage 19 (PC858: Euro) gemaess Epson-ESC/POS-Standard.
// CP858 deckt die deutschen Umlaute (ae/oe/ue/Ae/Oe/Ue/ss) und das Euro-Zeichen ab.
// Wird von ESC @ (Init) zurueckgesetzt und muss daher nach Init gesendet werden.
// HINWEIS: Der exakte Index ist druckerabhaengig und am Zielgeraet
// (MUNBYN ITPP047P-UE) zu verifizieren.
const SetCodepageCP858 = "\x1B\x74\x13" // ESC t 19

// Ausrichtung
const AlignLeft = "\x1B\x61\x00"
const AlignCenter = "\x1B\x61\x01"
const AlignRight = "\x1B\x61\x02"

// Schrift
const BoldOn = "\x1B\x45\x01"
const BoldOff = "\x1B\x45\x00"

// Schriftgroesse (GS ! n)
const TextNormal = "\x1D\x21\x00"
const TextDoubleHigh = "\x1D\x21\x01"  // Doppelte Hoehe
const TextDoubleWidth = "\x1D\x21\x10" // Doppelte Breite
const TextDoubleAll = "\x1D\x21\x11"   // Doppelte Hoehe und Breite (fuer Tischnummer)

// Hardware
const CutPaper = "\x1D\x56\x42\x00" // Partial Cut (GS V B 0)
const Beep = "\x1B\x42\x03\x02"     // 3 Piepser, Dauer 2 (ESC B n1 n2)

// Hardware-Statusabfrage (wird im Relay verwendet, nicht im Backend)
const StatusPaper = "\x10\x04\x04" // DLE EOT 4 — liefert 1 Byte zurueck
// Antwortbyte: Bit 5 (0x20) = Papier fast leer, Bit 6 (0x40) = Papier leer
// Drucker "bereit" wenn: (antwort & 0x60) == 0
