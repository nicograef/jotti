package main

// Die Identitaet des LAN-Stacks aus docker-compose.local.yml und
// docker-compose.release.yml. stack_contract_test.go prueft sie gegen beide
// Compose-Dateien und die Skripte in packaging/windows.
const (
	composeProject = "jotti-local"

	// postgresContainer ist der feste container_name des postgres-Service. pg_dump
	// laeuft per `docker exec` darin: lokale Socket-Verbindungen sind dort
	// trust-authentifiziert, also braucht das Backup kein Passwort.
	postgresContainer = "jotti-postgres-local"

	// Compose stellt benannten Volumes den Projektnamen voran.
	postgresDataVolumeKey = "postgres-data"
	postgresDataVolume    = composeProject + "_" + postgresDataVolumeKey
	configVolumeKey       = "jotti-config"
	configVolume          = composeProject + "_" + configVolumeKey

	// backupDir ist der Mountpunkt des jotti-backups-Volumes im postgres-Container.
	backupDir = "/jotti-backups"
)
