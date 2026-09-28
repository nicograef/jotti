package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

var ErrDatabase = db.ErrDatabase
var ErrNotFound = db.ErrNotFound
var ErrTSENichtKonfiguriert = errors.New("tse_not_configured")

// ErrTSEKonfigurationKassensitzungOffen rejects a change while a Kassensitzung is offen or
// wird_abgeschlossen.
var ErrTSEKonfigurationKassensitzungOffen = errors.New("tse_konfiguration_kassensitzung_offen")
var ErrTSEVerbindungFehlgeschlagen = errors.New("tse_connection_failed")
var ErrTSESetupZugangsdaten = errors.New("tse_setup_credentials_invalid")

// ErrTSESetupUmgebungAbweichung is the LIVE guard: the confirmed environment differs from the
// credentials' actual one.
var ErrTSESetupUmgebungAbweichung = errors.New("tse_setup_umgebung_abweichung")

// ErrTSEBereitsEingerichtet refuses a new TSS on an account with an active one; UebernimmTSE takes
// it over instead.
var ErrTSEBereitsEingerichtet = errors.New("tse_bereits_eingerichtet")

// ErrTSESetupLaeuftBereits means another writer holds the TSE configuration lock (einrichtungLaeuft).
var ErrTSESetupLaeuftBereits = errors.New("tse_setup_laeuft_bereits")

// ErrTSEEinrichtung is a failed step of the fiskaly lifecycle.
var ErrTSEEinrichtung = errors.New("tse_einrichtung_fehlgeschlagen")

// ErrTSESetupTSSLimitErreicht maps fiskaly's TEST limit of five active TSS (E_TSS_LIMIT_REACHED).
// fiskaly purges idle TEST TSS itself; jotti cannot disable them without their admin PIN.
var ErrTSESetupTSSLimitErreicht = errors.New("tse_setup_tss_limit_erreicht")

var ErrTSESetupTSSNichtGefunden = errors.New("tse_setup_tss_nicht_gefunden")

// ErrTSESetupPINErforderlich means taking over a TSS from UNINITIALIZED needs the stored admin PIN.
var ErrTSESetupPINErforderlich = errors.New("tse_setup_pin_erforderlich")

// ErrTSESetupPINUnbekannt means fiskaly rejected the admin PIN.
// It is a user-facing dead end (fiskaly support or a deliberate new TSS), not a technical error.
var ErrTSESetupPINUnbekannt = errors.New("tse_setup_pin_unbekannt")

// ErrTSESetupUebernahmeNichtMoeglich means the TSS state allows no resumption, e.g. DISABLED or
// DEFECTIVE.
var ErrTSESetupUebernahmeNichtMoeglich = errors.New("tse_setup_uebernahme_nicht_moeglich")

// ErrTSESetupPUKUnbekannt means fiskaly rejected the admin PUK during a PIN reset.
// It is a user-facing dead end (fiskaly support), not a technical error.
var ErrTSESetupPUKUnbekannt = errors.New("tse_setup_puk_unbekannt")
