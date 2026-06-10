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

// Schrift
const BoldOn = "\x1B\x45\x01"
const BoldOff = "\x1B\x45\x00"

// Schriftgroesse (GS ! n)
const TextNormal = "\x1D\x21\x00"
const TextDoubleHigh = "\x1D\x21\x01" // Doppelte Hoehe
const TextDoubleAll = "\x1D\x21\x11"  // Doppelte Hoehe und Breite (fuer Tischnummer)

// QR-Code (GS ( k)
const QRCodeStorePrefix = "\x1D\x28\x6B" // GS ( k + pL pH
const QRCodeModel2 = "\x1D\x28\x6B\x04\x00\x31\x41\x32\x00"
const QRCodeModuleSize6 = "\x1D\x28\x6B\x03\x00\x31\x43\x06"
const QRCodeErrorCorrectionM = "\x1D\x28\x6B\x03\x00\x31\x45\x31"
const QRCodePrint = "\x1D\x28\x6B\x03\x00\x31\x51\x30"

// Hardware
const CutPaper = "\x1D\x56\x42\x00" // Partial Cut (GS V B 0)
const Beep = "\x1B\x42\x03\x02"     // 3 Piepser, Dauer 2 (ESC B n1 n2)
