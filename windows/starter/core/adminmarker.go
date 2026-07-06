package core

import (
	"fmt"
	"regexp"
	"strings"
)

// AdminMarkerPrefix ist das feste ASCII-Literal am Anfang der maschinen-greifbaren
// Backend-Logzeile mit dem Initial-Admin-Code. Es muss mit
// backend/bootstrap.MarkerPrefix uebereinstimmen (Phase 2) — der Starter grept
// exakt danach. windows/starter ist ein eigenstaendiges Go-Modul und kann das
// Backend-Package nicht importieren, daher die lokale Kopie.
const AdminMarkerPrefix = "ADMIN-EINMALPASSWORT"

// AdminUsername ist der Benutzername des generierten Initial-Admins (muss mit
// backend/bootstrap.AdminUsername uebereinstimmen).
const AdminUsername = "admin"

// adminCodePattern extrahiert den 6-stelligen Klartext-Code aus der Markerzeile.
// Der zerolog-ConsoleWriter faerbt die Zeile mit ANSI-Escapes ein; die Ziffern
// selbst bleiben unberuehrt, daher greift das Muster unabhaengig von der Faerbung.
var adminCodePattern = regexp.MustCompile(`code=([0-9]{6})`)

// ParseAdminOTP durchsucht das Log-Blob nach der Markerzeile des Initial-Admins und
// liefert den 6-stelligen Klartext-Code der JUENGSTEN passenden Zeile (mehrere
// Marker: der neueste gewinnt). Robust gegen ANSI-umschlossene Zeilen: geprueft wird
// auf das Vorkommen des Praefix, nicht auf den Zeilenanfang. Fehlt der Marker oder
// steht kein 6-Ziffern-Code darin, ist found=false.
func ParseAdminOTP(logs string) (code string, found bool) {
	lines := strings.Split(logs, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if !strings.Contains(lines[i], AdminMarkerPrefix) {
			continue
		}
		if m := adminCodePattern.FindStringSubmatch(lines[i]); m != nil {
			return m[1], true
		}
	}
	return "", false
}

// AdminCodeHinweis liefert die deutsche, ASCII-transliterierte Konsolen-Anleitung
// zum Initial-Admin-Code (Stil wie diagnose.go: ae/oe/ue statt Umlauten). Bei found
// nennt sie Benutzer admin, den 6-stelligen Code und den Eingabeort in der App. Ohne
// Code (Einrichtung abgeschlossen oder kein Marker im aktuellen Boot) ist die Meldung
// non-fatal und verweist nur auf einen Neustart — kein Hinweis auf Logs oder Docker.
func AdminCodeHinweis(code string, found bool) string {
	if !found {
		return "Einrichtung ist abgeschlossen oder es liegt kein Code vor. " +
			"jotti neu starten, dann wird ein neuer Code angezeigt."
	}
	return fmt.Sprintf(
		"Ersteinrichtung: In der jotti-App \"Neues Passwort festlegen\" waehlen und als Benutzer "+
			"\"%s\" mit dem Einmalpasswort %s anmelden, dann ein eigenes Passwort festlegen.",
		AdminUsername, code)
}
