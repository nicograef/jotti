package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

var ErrDatabase = db.ErrDatabase
var ErrNotFound = db.ErrNotFound
var ErrTSENichtKonfiguriert = errors.New("tse_not_configured")

// ErrTSEKonfigurationKassensitzungOffen: Änderung abgelehnt, weil eine
// Kassensitzung aktiv ist (offen oder wird_abgeschlossen). Das Signaturgerät darf
// nicht mitten im Kassentag wechseln.
var ErrTSEKonfigurationKassensitzungOffen = errors.New("tse_konfiguration_kassensitzung_offen")
var ErrTSEVerbindungFehlgeschlagen = errors.New("tse_connection_failed")
var ErrTSESetupZugangsdaten = errors.New("tse_setup_credentials_invalid")

// ErrTSESetupUmgebungAbweichung zeigt an, dass die vom Admin bestaetigte
// Umgebung nicht der tatsächlichen Umgebung der Zugangsdaten entspricht — der
// Schutz vor einer versehentlichen LIVE-Anlage.
var ErrTSESetupUmgebungAbweichung = errors.New("tse_setup_umgebung_abweichung")

// ErrTSEBereitsEingerichtet zeigt an, dass das Konto bereits eine aktive TSS
// enthält. Die automatische Neuanlage wird dann verweigert; die vorhandene TSS
// lässt sich stattdessen übernehmen (UebernimmTSE).
var ErrTSEBereitsEingerichtet = errors.New("tse_bereits_eingerichtet")

// ErrTSESetupLaeuftBereits: es schreibt bereits jemand an der TSE-Konfiguration,
// der Aufruf startet gar nicht erst. Zwei überlappende Schreiber legten eine
// zweite, bezahlte TSS an bzw. überschrieben einander — siehe einrichtungLaeuft
// in setup.go.
var ErrTSESetupLaeuftBereits = errors.New("tse_setup_laeuft_bereits")

// ErrTSEEinrichtung zeigt einen Fehler während des fiskaly-Lebenszyklus an
// (Anlage, Initialisierung oder Client-Registrierung).
var ErrTSEEinrichtung = errors.New("tse_einrichtung_fehlgeschlagen")

// ErrTSESetupTSSLimitErreicht: das fiskaly-TEST-Konto hat die Obergrenze von fünf
// aktiven TSS erreicht (E_TSS_LIMIT_REACHED). fiskaly bereinigt alte TEST-TSS bei
// Inaktivität; jotti kann sie ohne Admin-PIN nicht stilllegen.
var ErrTSESetupTSSLimitErreicht = errors.New("tse_setup_tss_limit_erreicht")

var ErrTSESetupTSSNichtGefunden = errors.New("tse_setup_tss_nicht_gefunden")

// ErrTSESetupPINErforderlich zeigt an, dass die Übernahme einer bereits
// personalisierten TSS (ab UNINITIALIZED) die vom Admin verwahrte Admin-PIN
// benötigt — sie liegt aber nicht vor.
var ErrTSESetupPINErforderlich = errors.New("tse_setup_pin_erforderlich")

// ErrTSESetupPINUnbekannt zeigt an, dass fiskaly die übergebene Admin-PIN
// abgelehnt hat — der Admin kennt die verwahrte PIN nicht (mehr). Sackgasse mit
// Auswegen (fiskaly-Support oder bewusste Neuanlage), kein technischer Fehler.
var ErrTSESetupPINUnbekannt = errors.New("tse_setup_pin_unbekannt")

// ErrTSESetupUebernahmeNichtMoeglich zeigt an, dass die TSS in einem Zustand ist,
// aus dem keine Wiederaufnahme möglich ist (z. B. DISABLED oder DEFECTIVE).
var ErrTSESetupUebernahmeNichtMoeglich = errors.New("tse_setup_uebernahme_nicht_moeglich")

// ErrTSESetupPUKUnbekannt zeigt an, dass fiskaly den beim PIN-Reset übergebenen
// Admin-PUK abgelehnt hat. Die Zugangsdaten sind zu diesem Zeitpunkt bereits
// bestaetigt, daher ist ein Fehler beim Setzen der PIN praktisch immer ein
// falscher PUK. Sackgasse mit Ausweg (fiskaly-Support), kein technischer Fehler.
var ErrTSESetupPUKUnbekannt = errors.New("tse_setup_puk_unbekannt")
