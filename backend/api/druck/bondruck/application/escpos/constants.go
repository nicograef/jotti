package escpos

const Init = "\x1B\x40"

// Zeichentabelle (ESC t n): Codepage 6 ist am MUNBYN ITPP047P WPC1252
// (Windows-1252, deutsche Umlaute und Euro-Zeichen); die MUNBYN-Nummerierung
// folgt NICHT Epson (dort wäre WPC1252 = 16), die Liste steht auf der
// Selbsttest-Seite des Druckers. ESC @ (Init) setzt die Codepage zurück — daher
// immer NACH Init senden.
const SetCodepageWPC1252 = "\x1B\x74\x06" // ESC t 6

const AlignLeft = "\x1B\x61\x00"
const AlignCenter = "\x1B\x61\x01"

const BoldOn = "\x1B\x45\x01"
const BoldOff = "\x1B\x45\x00"

// Schriftgröße (GS ! n)
const TextNormal = "\x1D\x21\x00"
const TextDoubleHigh = "\x1D\x21\x01" // Doppelte Höhe
const TextDoubleAll = "\x1D\x21\x11"  // Doppelte Höhe und Breite (für Tischnummer)

// QR-Code (GS ( k)
const QRCodeStorePrefix = "\x1D\x28\x6B" // GS ( k + pL pH
const QRCodeModel2 = "\x1D\x28\x6B\x04\x00\x31\x41\x32\x00"

// QRCodeModuleSizeCmdPrefix is the GS ( k command prefix for setting the QR
// module size; the module-size byte n (1-16) must be appended immediately.
const QRCodeModuleSizeCmdPrefix = "\x1D\x28\x6B\x03\x00\x31\x43"
const QRCodeErrorCorrectionM = "\x1D\x28\x6B\x03\x00\x31\x45\x31"
const QRCodePrint = "\x1D\x28\x6B\x03\x00\x31\x51\x30"

const CutPaper = "\x1D\x56\x42\x00" // Partial Cut (GS V B 0)
const Beep = "\x1B\x42\x03\x02"     // 3 Piepser, Dauer 2 (ESC B n1 n2)
