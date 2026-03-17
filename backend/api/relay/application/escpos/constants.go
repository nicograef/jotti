package escpos

// Initialisierung
const Init = "\x1B\x40"

// Ausrichtung
const AlignLeft = "\x1B\x61\x00"
const AlignCenter = "\x1B\x61\x01"
const AlignRight = "\x1B\x61\x02"

// Schrift
const BoldOn = "\x1B\x45\x01"
const BoldOff = "\x1B\x45\x00"

// Schriftgröße (GS ! n)
const TextNormal = "\x1D\x21\x00"
const TextDoubleHigh = "\x1D\x21\x01"  // Doppelte Höhe
const TextDoubleWidth = "\x1D\x21\x10" // Doppelte Breite
const TextDoubleAll = "\x1D\x21\x11"   // Doppelte Höhe und Breite (für Tischnummer)

// Hardware
const CutPaper = "\x1D\x56\x42\x00" // Partial Cut (GS V B 0)
const Beep = "\x1B\x42\x03\x02"     // 3 Piepser, Dauer 2 (ESC B n1 n2)

// Hardware-Statusabfrage (wird im Relay verwendet, nicht im Backend)
const StatusPaper = "\x10\x04\x04" // DLE EOT 4 — liefert 1 Byte zurück
// Antwortbyte: Bit 5 (0x20) = Papier fast leer, Bit 6 (0x40) = Papier leer
// Drucker "bereit" wenn: (antwort & 0x60) == 0
