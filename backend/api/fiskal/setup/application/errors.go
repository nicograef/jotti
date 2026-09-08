package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

var ErrDatabase = db.ErrDatabase
var ErrNotFound = db.ErrNotFound
var ErrTSENichtKonfiguriert = errors.New("tse_not_configured")

// ErrTSEKonfigurationKassensitzungOffen zeigt an, dass eine Änderung der
// TSE-Konfiguration abgelehnt wurde, weil eine Kassensitzung offen ist. Das
// Signaturgeraet darf nicht mitten in einem laufenden Kassentag wechseln — der
// Admin schließt die Kassensitzung und wiederholt die Änderung.
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

// ErrTSESetupLaeuftBereits zeigt an, dass bereits jemand an der
// TSE-Konfiguration schreibt (Neuanlage, Übernahme oder manueller
// Zugangsdaten-Wechsel) und der Aufruf deshalb gar nicht erst gestartet wurde.
// Zwei überlappende Schreiber würden eine zweite, bezahlte TSS anlegen bzw.
// die Konfiguration des jeweils anderen überschreiben — siehe
// einrichtungLaeuft in setup.go.
var ErrTSESetupLaeuftBereits = errors.New("tse_setup_laeuft_bereits")

// ErrTSEEinrichtung zeigt einen Fehler während des fiskaly-Lebenszyklus an
// (Anlage, Initialisierung oder Client-Registrierung).
var ErrTSEEinrichtung = errors.New("tse_einrichtung_fehlgeschlagen")

// ErrTSESetupTSSLimitErreicht zeigt an, dass das fiskaly-TEST-Konto die
// Obergrenze von fünf aktiven TSS erreicht hat (E_TSS_LIMIT_REACHED). Alte
// TEST-TSS werden von fiskaly bei Inaktivität automatisch bereinigt; jotti
// kann sie ohne Admin-PIN nicht stilllegen. Verständliche Meldung statt
// technischem Fehler.
var ErrTSESetupTSSLimitErreicht = errors.New("tse_setup_tss_limit_erreicht")

// ErrTSESetupTSSNichtGefunden zeigt an, dass die zur Übernahme gewählte TSS im
// fiskaly-Konto nicht (mehr) existiert.
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
