package core

import "sort"

// ShouldBackup entscheidet, ob vor dem vollen Hochfahren des Stacks (inkl. der
// schemaveraendernden Migrationen) ein automatischer pg_dump gezogen werden soll.
// Ausloeser ist ausschliesslich ein echter Versionswechsel — lastVersion ist
// gesetzt und weicht von currentVersion ab — bei gleichzeitig vorhandenen Daten
// (postgres-data-Volume existiert). Ein leerer lastVersion (Erststart, noch kein
// Marker) und eine unveraenderte Version erzeugen bewusst kein Backup; ohne Daten
// gibt es nichts zu sichern.
func ShouldBackup(lastVersion, currentVersion string, postgresDataExists bool) bool {
	if lastVersion == "" || lastVersion == currentVersion {
		return false
	}
	return postgresDataExists
}

// DumpsToDelete liefert die zu loeschenden Dump-Dateinamen, sodass nur die
// neuesten keep Backups erhalten bleiben. Die Namen sind zeitgestempelt
// (jotti-YYYYMMDD-HHMMSS.sql) und damit lexikografisch == chronologisch
// sortierbar; rotiert werden die aeltesten ueber keep hinaus. Bei keep <= 0 wird
// defensiv nichts geloescht — eine Fehlkonfiguration darf nie alle Backups
// entfernen.
func DumpsToDelete(names []string, keep int) []string {
	if keep <= 0 || len(names) <= keep {
		return nil
	}
	sorted := append([]string(nil), names...)
	sort.Strings(sorted)
	return sorted[:len(sorted)-keep]
}
