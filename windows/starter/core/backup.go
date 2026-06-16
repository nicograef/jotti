package core

import "sort"

// devVersion ist der Versionswert eines nicht-releasten Entwickler-Builds (der
// Default von main.version). Ein Dev-Build loest nie ein Backup aus, damit der
// lokale Repo-Dev-Lauf (go run) inert bleibt und keine fremden Volumes anfasst.
const devVersion = "dev"

// ShouldBackup entscheidet, ob vor dem vollen Hochfahren des Stacks (inkl. der
// schemaveraendernden Migrationen) ein automatischer pg_dump gezogen werden soll.
// Gesichert wird, sobald Daten existieren (postgres-data-Volume vorhanden) und die
// laufende Version nicht nachweislich gleich der zuletzt gesund gestarteten ist —
// also bei einem echten Versionswechsel ebenso wie bei fehlendem last-version-Marker
// (Erst-Upgrade von einer Vor-Phase-3-Version, die noch keinen Marker schrieb). Ohne
// Daten gibt es nichts zu sichern (echte Erstinstallation); ein Dev-Build sichert
// nie (lokaler Repo-Dev-Lauf bleibt inert).
func ShouldBackup(lastVersion, currentVersion string, postgresDataExists bool) bool {
	if !postgresDataExists || currentVersion == devVersion {
		return false
	}
	return lastVersion != currentVersion
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
