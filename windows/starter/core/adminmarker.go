package core

import (
	"fmt"
	"regexp"
	"strings"
)

// AdminMarkerPrefix muss mit backend/bootstrap.MarkerPrefix uebereinstimmen — der
// Starter grept exakt danach. windows/starter ist ein eigenstaendiges Go-Modul und
// kann das Backend-Package nicht importieren, daher die lokale Kopie.
const AdminMarkerPrefix = "ADMIN-EINMALPASSWORT"

// AdminUsername ist der Benutzername des generierten Initial-Admins (muss mit
// backend/bootstrap.AdminUsername uebereinstimmen).
const AdminUsername = "admin"

// adminCodePattern extrahiert den 6-stelligen Code. Der zerolog-ConsoleWriter
// faerbt die Zeile mit ANSI-Escapes ein; die Ziffern bleiben unberuehrt.
var adminCodePattern = regexp.MustCompile(`code=([0-9]{6})`)

// ParseAdminOTP liefert den Code der JUENGSTEN Markerzeile im Log-Blob. Geprueft
// wird auf das Vorkommen des Praefix, nicht auf den Zeilenanfang (ANSI-umschlossene
// Zeilen); ohne Marker oder Code ist found=false.
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

// AdminCodeHinweis liefert die Konsolen-Anleitung zum Initial-Admin-Code, deutsch
// und ASCII-transliteriert wie die uebrigen Konsolen-Strings. Ohne Code verweist sie
// nur auf einen Neustart — kein Hinweis auf Logs oder Docker.
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
