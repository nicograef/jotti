package core

import "strings"

// envContent erzeugt den vollstaendigen .env-Inhalt mit frisch erzeugten
// Secrets. POSTGRES_USER ist fest "admin" (wie .env.example); die drei Secrets
// stammen aus GenerateSecret. Der Kommentar-Header haelt die erste Zeile frei
// von einem Key: Schreibt Notepad spaeter ein UTF-8-BOM in die Datei, landet es
// so vor einem Kommentar statt vor POSTGRES_USER und beschaedigt keinen Key.
func envContent() string {
	lines := []string{
		"# Diese Datei wurde automatisch von jotti erzeugt. Hier muss nichts geaendert werden.",
		"POSTGRES_USER=admin",
		"POSTGRES_PASSWORD=" + GenerateSecret(),
		"JWT_SECRET=" + GenerateSecret(),
		"RELAY_AUTH_TOKEN=" + GenerateSecret(),
		"",
	}
	return strings.Join(lines, "\n")
}

// MaterializeEnv schreibt die .env nach path, falls sie noch nicht existiert,
// und meldet ueber created, ob geschrieben wurde. Eine vorhandene Datei wird
// nie ueberschrieben (idempotent wie scripts/init-env.sh) — die Secrets werden
// dann gar nicht erst erzeugt. exists und write kapseln die Dateizugriffe, damit
// die Funktion ohne echtes Dateisystem testbar ist.
func MaterializeEnv(path string, exists func(string) (bool, error), write func(string, []byte) error) (created bool, err error) {
	present, err := exists(path)
	if err != nil {
		return false, err
	}
	if present {
		return false, nil
	}
	if err := write(path, []byte(envContent())); err != nil {
		return false, err
	}
	return true, nil
}
