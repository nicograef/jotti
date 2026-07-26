// input-otp 1.4.2 startet in einem Effect drei Timer (0/10/50 ms), die einen
// React-State-Update auslösen, und räumt sie beim Unmount nicht ab (dist:
// `ht()` wird ohne Cleanup-Rückgabe aufgerufen). Feuern sie erst, nachdem vitest
// die jsdom-Umgebung der Testdatei abgebaut hat, kippt der State-Update mit
// "window is not defined" als unhandled error — der gesamte Lauf wird rot,
// obwohl jeder Test bestanden hat. 1.4.2 ist die letzte Version des Pakets
// (Stand 2026-07), ein Upgrade behebt es also nicht.
//
// drainInputOtpTimers() läuft nach cleanup() und lässt die Timer noch in der
// lebenden Umgebung auslaufen. Jede Testdatei, die ein OTP-Feld rendert, ruft es
// in afterEach auf.
const MAX_TIMER_MS = 50

export function drainInputOtpTimers(): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, MAX_TIMER_MS + 10))
}
