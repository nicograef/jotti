package seed

// AllowSeedEnv ist das Opt-in-Flag, das den Seed-Befehl freischaltet. Das
// seed-Subkommando ist ins Produktions-Binary einkompiliert; ohne dieses
// explizite Flag verweigert es den Lauf, damit ein versehentliches `jotti seed`
// keine Demo-Daten (öffentlich bekannte Passwörter, Fake-TSE-Events) in eine
// echte Installation schreibt. Der Kassenjournal-Guard in writeSeed bleibt als
// zweite Schicht bestehen.
const AllowSeedEnv = "JOTTI_ALLOW_SEED"

// AllowedByEnv meldet, ob das Seeden per JOTTI_ALLOW_SEED=1 ausdrücklich erlaubt
// ist. Nur der exakte Wert "1" schaltet frei; jeder andere Wert (leer, "0",
// "true", …) verweigert.
func AllowedByEnv(getenv func(string) string) bool {
	return getenv(AllowSeedEnv) == "1"
}
