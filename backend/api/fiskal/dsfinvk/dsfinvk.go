// Package dsfinvk turns a Kassensitzung's events and master data into a DSFinV-K
// archive (CSVs, index.xml, DTD) without I/O; the orchestrator loads the data.
package dsfinvk

import (
	"fmt"
	"strconv"
	"time"

	"github.com/nicograef/jotti/backend/domain/betreiber"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/domain/tse"
)

// Version is the declared DSFinV-K version, kept in one place for a future spec
// version (docs/compliance.md §6.1).
const Version = "2.4"

// Snapshot is the master data the export needs besides the events.
type Snapshot struct {
	// KasseSeriennummer feeds Z_KASSE_ID and KASSE_SERIENNR (the Kasse UUID).
	KasseSeriennummer string
	// Erstellung is Z_ERSTELLUNG: the Tagesabschluss time of a closed session, the
	// export time of an open one.
	Erstellung time.Time
	// KassensitzungNr is the closing's Z_NR.
	KassensitzungNr int
	Betreiber       betreiber.Betreiber
	TSEStammdaten   tse.Stammdaten
	// SoftwareVersion is the jotti build version (KASSE_SW_VERSION), set via ldflags
	// ("dev" in development).
	SoftwareVersion string
	// Tischnamen maps table IDs to names (source of ABRECHNUNGSKREIS), deleted tables
	// included; for a missing table the mapper synthesises "Tisch N".
	Tischnamen map[int]string
}

func itoa(n int) string { return strconv.Itoa(n) }

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func derefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// stornoNein is the DSFinV-K flag "0" for BON_STORNO and P_STORNO; jotti always sets
// it, as partial returns are shown negative (docs/compliance.md §6.6).
const stornoNein = "0"

// formatAmount renders cents with a decimal comma and two decimals, e.g. -150 -> "-1,50".
// The official index.xml sets DecimalSymbol ","; two decimals are the DSFinV-K norm
// (up to five allowed).
func formatAmount(cents int) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%d,%02d", sign, cents/100, cents%100)
}

// formatQuantity renders a quantity with three decimals (MENGE).
func formatQuantity(menge int) string {
	return fmt.Sprintf("%d,000", menge)
}

// ustNichtSteuerbar is DSFinV-K VAT key 5 (Anlage 2) for non-taxable cash movements
// (Anfangsbestand, Geldtransit, Kassendifferenz).
const ustNichtSteuerbar = 5

// ustSchluessel maps a jotti rate to its DSFinV-K Anlage 2 key: 1 = 19 %, 2 = 7 %,
// 6 = exempt (0 %, e.g. § 19 UStG). Key 7 (not determinable) only serves receivables,
// which revenue-at-payment never has; kombi is split before it arrives here.
func ustSchluessel(satz steuer.Steuersatz) int {
	switch satz {
	case steuer.RegelSteuersatz:
		return 1
	case steuer.ErmaessigtSteuersatz:
		return 2
	case steuer.BefreitSteuersatz:
		return 6
	default:
		return 0
	}
}
