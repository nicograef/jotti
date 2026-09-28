package seed

// AllowSeedEnv opts in to the seed subcommand, which ships in the production binary.
// Without it an accidental `jotti seed` would write demo data (public passwords, fake-TSE events) into a real installation.
const AllowSeedEnv = "JOTTI_ALLOW_SEED"

// AllowedByEnv meldet, ob das Seeden per JOTTI_ALLOW_SEED=1 ausdrücklich erlaubt
// ist. Nur der exakte Wert "1" schaltet frei; jeder andere Wert (leer, "0",
// "true", …) verweigert.
func AllowedByEnv(getenv func(string) string) bool {
	return getenv(AllowSeedEnv) == "1"
}
