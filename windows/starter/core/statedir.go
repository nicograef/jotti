package core

import "slices"

// Wohlbekannte SIDs, denen das Zustandsverzeichnis gehoeren darf.
const (
	SIDSystem         = "S-1-5-18"
	SIDAdministrators = "S-1-5-32-544"
)

// StateDirOwnerAllowed meldet, ob ein vorgefundenes Zustandsverzeichnis mit diesem
// Besitzer uebernommen werden darf: SYSTEM, Administratoren oder einer der aktuellen
// Benutzer. Ein anderer Besitzer koennte eine .env mit ihm bekannten Secrets oder
// einen last-version-Marker untergeschoben haben.
func StateDirOwnerAllowed(owner string, currentUsers []string) bool {
	if owner == "" {
		return false
	}
	return owner == SIDSystem || owner == SIDAdministrators || slices.Contains(currentUsers, owner)
}

// StateDirSDDL ist die geschuetzte, nicht erbende DACL des Zustandsverzeichnisses:
// Vollzugriff fuer SYSTEM und Administratoren, Aendern (0x1301bf) fuer den angemeldeten
// Benutzer, als der das nicht-elevierte jotti-relay und die .cmd-Skripte laufen.
// Vollzugriff fuer die Administratoren haelt WRITE_DAC fuer den naechsten Start offen,
// auch wenn der Ordner einem anderen Konto gehoert.
func StateDirSDDL(userSID string) string {
	return "D:P(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)(A;OICI;0x1301bf;;;" + userSID + ")"
}

// PrivateDirSDDL ist die geschuetzte DACL der Unterordner, deren Inhalt Code als
// Administrator ausfuehren koennte (pgdata, Caddys CA-Schluessel): Vollzugriff nur fuer
// SYSTEM, Administratoren und den Prozessbenutzer, nicht fuer den angemeldeten Benutzer.
// Postgres laeuft mit eingeschraenktem Token, in dem die Administratoren nur noch
// verweigern; es schreibt allein ueber die SID des Prozessbenutzers.
func PrivateDirSDDL(processSID string) string {
	return "D:P(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)(A;OICI;FA;;;" + processSID + ")"
}
