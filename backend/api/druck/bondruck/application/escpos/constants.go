package escpos

const Init = "\x1B\x40"

// Codepage 0 is PC437 on Epson, MUNBYN and most clones; higher indices differ per brand.
// ESC @ (Init) resets the codepage to the printer's default, so send this after Init.
const SetCodepagePC437 = "\x1B\x74\x00" // ESC t 0

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
