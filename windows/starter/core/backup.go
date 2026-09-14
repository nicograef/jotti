package core

import (
	"slices"
	"sort"
)

// devVersion ist der Default von main.version; ein Dev-Build loest nie ein Backup
// aus, damit der lokale Repo-Dev-Lauf keine fremden Volumes anfasst.
const devVersion = "dev"

// ShouldBackup entscheidet, ob vor dem vollen Hochfahren (inkl. der
// schemaveraendernden Migrationen) ein automatischer pg_dump noetig ist: nur wenn
// Daten existieren (postgres-data-Volume) und die laufende Version nicht
// nachweislich der zuletzt gesund gestarteten entspricht — ein fehlender
// last-version-Marker zaehlt als Abweichung. Ein Dev-Build sichert nie.
func ShouldBackup(lastVersion, currentVersion string, postgresDataExists bool) bool {
	if !postgresDataExists || currentVersion == devVersion {
		return false
	}
	return lastVersion != currentVersion
}

// DumpsToDelete liefert die zu loeschenden Dumps, sodass nur die neuesten keep
// bleiben. Die zeitgestempelten Namen (jotti-YYYYMMDD-HHMMSS.sql) sind
// lexikografisch == chronologisch sortierbar. Bei keep <= 0 wird nichts geloescht —
// eine Fehlkonfiguration darf nie alle Backups entfernen.
func DumpsToDelete(names []string, keep int) []string {
	if keep <= 0 || len(names) <= keep {
		return nil
	}
	sorted := append([]string(nil), names...)
	sort.Strings(sorted)
	return sorted[:len(sorted)-keep]
}

// MirrorPlan: Copy ist leer, wenn die Datei auf dem Host schon liegt; Delete nennt
// die danach zu rotierenden Host-Dateien.
type MirrorPlan struct {
	Copy   string
	Delete []string
}

// PlanBackupMirror entscheidet rein, wie ein neuer Dump gespiegelt wird: kopiert
// wird nur, wenn newDump auf dem Host fehlt; geloescht wird nach derselben Regel wie
// in DumpsToDelete, sodass Volume und Host-Spiegel dieselbe Aufbewahrung teilen.
func PlanBackupMirror(newDump string, hostNames []string, keep int) MirrorPlan {
	plan := MirrorPlan{}
	all := append([]string(nil), hostNames...)
	if !slices.Contains(hostNames, newDump) {
		plan.Copy = newDump
		all = append(all, newDump)
	}
	plan.Delete = DumpsToDelete(all, keep)
	return plan
}
