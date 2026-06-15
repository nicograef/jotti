package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

var ErrDatabase = db.ErrDatabase
var ErrNotFound = db.ErrNotFound
var ErrTSENichtKonfiguriert = errors.New("tse_not_configured")
var ErrTSEVerbindungFehlgeschlagen = errors.New("tse_connection_failed")
var ErrTSESetupZugangsdaten = errors.New("tse_setup_credentials_invalid")

// ErrTSESetupUmgebungAbweichung zeigt an, dass die vom Admin bestaetigte
// Umgebung nicht der tatsaechlichen Umgebung der Zugangsdaten entspricht — der
// Schutz vor einer versehentlichen LIVE-Anlage.
var ErrTSESetupUmgebungAbweichung = errors.New("tse_setup_umgebung_abweichung")

// ErrTSEBereitsEingerichtet zeigt an, dass das Konto bereits eine aktive TSS
// enthaelt. Die automatische Neuanlage wird dann verweigert (die Uebernahme
// einer vorhandenen TSS folgt in einer spaeteren Phase).
var ErrTSEBereitsEingerichtet = errors.New("tse_bereits_eingerichtet")

// ErrTSEEinrichtung zeigt einen Fehler waehrend des fiskaly-Lebenszyklus an
// (Anlage, Initialisierung oder Client-Registrierung).
var ErrTSEEinrichtung = errors.New("tse_einrichtung_fehlgeschlagen")

// ErrTSESetupTSSNichtGefunden zeigt an, dass die zur Uebernahme gewaehlte TSS im
// fiskaly-Konto nicht (mehr) existiert.
var ErrTSESetupTSSNichtGefunden = errors.New("tse_setup_tss_nicht_gefunden")

// ErrTSESetupPINErforderlich zeigt an, dass die Uebernahme einer bereits
// personalisierten TSS (ab UNINITIALIZED) die vom Admin verwahrte Admin-PIN
// benoetigt — sie liegt aber nicht vor.
var ErrTSESetupPINErforderlich = errors.New("tse_setup_pin_erforderlich")

// ErrTSESetupPINUnbekannt zeigt an, dass fiskaly die uebergebene Admin-PIN
// abgelehnt hat — der Admin kennt die verwahrte PIN nicht (mehr). Sackgasse mit
// Auswegen (fiskaly-Support oder bewusste Neuanlage), kein technischer Fehler.
var ErrTSESetupPINUnbekannt = errors.New("tse_setup_pin_unbekannt")

// ErrTSESetupUebernahmeNichtMoeglich zeigt an, dass die TSS in einem Zustand ist,
// aus dem keine Wiederaufnahme moeglich ist (z. B. DISABLED oder DEFECTIVE).
var ErrTSESetupUebernahmeNichtMoeglich = errors.New("tse_setup_uebernahme_nicht_moeglich")
