package tse_repo

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

// MaxSignaturVersuche ist die Anzahl auftragsspezifischer Fehlversuche, nach
// der ein Signaturauftrag endgueltig fehlgeschlagen ist. Solche Fehler (von
// fiskaly abgelehnte processData, tse.AuftragsFehler) sind fast immer
// deterministisch — mit dem Sekunden-Backoff (5, 15, 45 s) endet die Kurve
// nach unter einer Minute und damit bewusst unter tse.RueckstandSchwelle:
// Ein Gift-Auftrag schlaegt endgueltig fehl, bevor der Watchdog ihn als
// Rueckstand dokumentiert. TSE-weite Fehler zaehlen nie auf den Auftrag,
// sondern schalten den Signatur-Worker in den Stoerungszustand.
const MaxSignaturVersuche = 3

// OffenerSignaturauftrag ist die Worker-Sicht eines faelligen Auftrags.
type OffenerSignaturauftrag struct {
	ID          int
	TxID        string
	ProcessType string
	ProcessData string
}

// Signaturauftrag ist die Admin-Sicht eines Signaturauftrags. Sie dient
// zugleich als TSE-Ausfalldokumentation (AEAO zu § 146a, 1.14.1):
// ErstelltAm = Beginn, ErledigtAm = Ende, LetzterFehler = Grund.
type Signaturauftrag struct {
	ID            int
	TxID          string
	ProcessType   string
	Status        string
	Versuche      int
	LetzterFehler string
	ErstelltAm    time.Time
	ErledigtAm    *time.Time
}

type Repository struct {
	db *sql.DB
	q  *dbgen.Queries
}

func NewRepository(database *sql.DB) Repository {
	return Repository{db: database, q: dbgen.New(database)}
}

func (r Repository) GetOffeneTSESignaturauftraege(ctx context.Context, limit int) ([]OffenerSignaturauftrag, error) {
	rows, err := r.q.GetOffeneTSESignaturauftraege(ctx, int32(limit))
	if err != nil {
		return nil, db.Error(err)
	}

	result := make([]OffenerSignaturauftrag, 0, len(rows))
	for _, row := range rows {
		result = append(result, OffenerSignaturauftrag{
			ID:          row.ID,
			TxID:        row.TxID,
			ProcessType: row.ProcessType,
			ProcessData: row.ProcessData,
		})
	}

	return result, nil
}

// QuittiereTSESignaturauftrag schreibt die Signatur als einzelnes Update an den
// Auftrag: Signaturspalten fuellen, Status erledigt. Der Status-Guard (offen)
// macht die Quittierung idempotent — die Signaturspalten werden genau einmal
// beschrieben.
func (r Repository) QuittiereTSESignaturauftrag(ctx context.Context, auftragID int, signatur tse.Signatur) error {
	return db.Error(r.q.QuittiereTSESignaturauftrag(ctx, dbgen.QuittiereTSESignaturauftragParams{
		ID:                auftragID,
		TransaktionNummer: sql.NullInt32{Int32: int32(signatur.TransaktionNummer), Valid: true},
		SignaturZaehler:   sql.NullInt32{Int32: int32(signatur.SignaturZaehler), Valid: true},
		TseSeriennummer:   sql.NullString{String: strings.TrimSpace(signatur.TSESeriennummer), Valid: true},
		LogTimeStart:      sql.NullTime{Time: signatur.LogTimeStart.UTC(), Valid: true},
		LogTimeEnd:        sql.NullTime{Time: signatur.LogTimeEnd.UTC(), Valid: true},
		Signatur:          sql.NullString{String: strings.TrimSpace(signatur.Signatur), Valid: true},
		QrCodeData:        sql.NullString{String: strings.TrimSpace(signatur.QRCodeData), Valid: true},
	}))
}

// TSESignaturauftragFehlversuch verbucht einen auftragsspezifischen
// Fehlversuch: Zaehler hoch, Fehlertext speichern, naechster Versuch mit
// Sekunden-Backoff (5, 15, 45 s). Beim MaxSignaturVersuche-ten Fehlversuch
// wechselt der Auftrag auf fehlgeschlagen (Backoff-Logik liegt in der
// SQL-Query).
func (r Repository) TSESignaturauftragFehlversuch(ctx context.Context, auftragID int, fehler string) error {
	return db.Error(r.q.TSESignaturauftragFehlversuch(ctx, dbgen.TSESignaturauftragFehlversuchParams{
		ID:            auftragID,
		LetzterFehler: sql.NullString{String: fehler, Valid: true},
		MaxVersuche:   MaxSignaturVersuche,
	}))
}

// MarkiereOffeneAlsNichtKonfiguriert markiert alle offenen Auftraege endgueltig
// als tse_nicht_konfiguriert und liefert die Anzahl markierter Auftraege. Ohne
// vorhandene TSE-Konfiguration gibt es keine Signatur; ein Nachsignieren ist
// ausgeschlossen (keine Fehlversuche, keine automatische Wiederaufnahme).
// Bereits endgueltig markierte Auftraege bleiben unberuehrt.
func (r Repository) MarkiereOffeneAlsNichtKonfiguriert(ctx context.Context) (int64, error) {
	n, err := r.q.MarkiereOffeneTSESignaturauftraegeNichtKonfiguriert(ctx)
	if err != nil {
		return 0, db.Error(err)
	}
	return n, nil
}

// GetTSESignaturauftraege liefert die Signaturauftraege fuer die
// Admin-Verwaltung und die Ausfalldokumentation, neueste zuerst.
func (r Repository) GetTSESignaturauftraege(ctx context.Context) ([]Signaturauftrag, error) {
	rows, err := r.q.GetTSESignaturauftraege(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	result := make([]Signaturauftrag, 0, len(rows))
	for i := range rows {
		row := &rows[i]
		auftrag := Signaturauftrag{
			ID:            row.ID,
			TxID:          row.TxID,
			ProcessType:   row.ProcessType,
			Status:        row.Status,
			Versuche:      row.Versuche,
			LetzterFehler: row.LetzterFehler.String,
			ErstelltAm:    row.ErstelltAm,
		}
		if row.ErledigtAm.Valid {
			erledigtAm := row.ErledigtAm.Time
			auftrag.ErledigtAm = &erledigtAm
		}
		result = append(result, auftrag)
	}

	return result, nil
}

// TSESignaturauftragZuruecksetzen reiht einen endgueltig markierten Auftrag
// wieder ein (fehlgeschlagen oder tse_nicht_konfiguriert -> offen, Zaehler und
// Fehler zurueckgesetzt). Der Status-Guard wirkt nur auf diese beiden Status;
// so signiert der Worker nach einer spaeten Einrichtung zurueckgesetzte
// Auftraege nach.
func (r Repository) TSESignaturauftragZuruecksetzen(ctx context.Context, auftragID int) error {
	return db.Error(r.q.TSESignaturauftragZuruecksetzen(ctx, auftragID))
}

// TSESignaturauftragVerwerfen markiert einen fehlgeschlagenen Auftrag als
// verworfen. Der Eintrag bleibt fuer die Ausfalldokumentation erhalten; der
// Status-Guard wirkt nur auf fehlgeschlagene Auftraege.
func (r Repository) TSESignaturauftragVerwerfen(ctx context.Context, auftragID int) error {
	return db.Error(r.q.TSESignaturauftragVerwerfen(ctx, auftragID))
}

func (r Repository) CountOffeneTSESignaturauftraege(ctx context.Context) (int, error) {
	count, err := r.q.CountOffeneTSESignaturauftraege(ctx)
	if err != nil {
		return 0, db.Error(err)
	}
	return count, nil
}

// GetSignaturauftragZuEvent liefert den Signatur-Stand eines Events fuer den
// Beleg-Abruf. db.ErrNotFound heisst: kein Auftrag, das Event ist nicht
// signaturpflichtig.
func (r Repository) GetSignaturauftragZuEvent(ctx context.Context, eventID int) (tse.SignaturauftragStand, error) {
	row, err := r.q.GetTSESignaturauftragZuEvent(ctx, eventID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tse.SignaturauftragStand{}, db.ErrNotFound
		}
		return tse.SignaturauftragStand{}, db.Error(err)
	}

	stand := tse.SignaturauftragStand{Status: row.Status, ErstelltAm: row.ErstelltAm}
	if row.Status == tse.StatusErledigt {
		stand.Signatur = &tse.Signatur{
			TransaktionNummer: int(row.TransaktionNummer.Int32),
			SignaturZaehler:   int(row.SignaturZaehler.Int32),
			TSESeriennummer:   row.TseSeriennummer.String,
			LogTimeStart:      row.LogTimeStart.Time,
			LogTimeEnd:        row.LogTimeEnd.Time,
			Signatur:          row.Signatur.String,
			QRCodeData:        row.QrCodeData.String,
		}
	}
	return stand, nil
}

// GetAeltesterOffenerTSESignaturauftrag liefert den Erstellungszeitpunkt des
// aeltesten offenen Signaturauftrags; nil, wenn kein Auftrag offen ist. Der
// Rueckstands-Watchdog bemisst daran den Signatur-Rueckstand.
func (r Repository) GetAeltesterOffenerTSESignaturauftrag(ctx context.Context) (*time.Time, error) {
	erstelltAm, err := r.q.GetAeltesterOffenerTSESignaturauftrag(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, db.Error(err)
	}
	return &erstelltAm, nil
}

// OeffneTSEStoerung oeffnet einen Stoerungszeitraum im Stoerungsprotokoll.
// Idempotent: Solange irgendein Zeitraum aktiv ist, ist das Oeffnen ein No-Op
// (hoechstens ein aktiver Zeitraum, DB-seitig per partiellem Unique-Index).
func (r Repository) OeffneTSEStoerung(ctx context.Context, grundArt string, fehlertext string) error {
	return db.Error(r.q.OeffneTSEStoerung(ctx, dbgen.OeffneTSEStoerungParams{
		GrundArt:   grundArt,
		Fehlertext: fehlertext,
	}))
}

// SchliesseTSEStoerung beendet den aktiven Stoerungszeitraum der Grund-Art;
// jeder Schreiber schliesst nur Zeitraeume seiner Grund-Art. Idempotent: Ohne
// aktiven Zeitraum der Art ein No-Op.
func (r Repository) SchliesseTSEStoerung(ctx context.Context, grundArt string) error {
	return db.Error(r.q.SchliesseTSEStoerung(ctx, grundArt))
}

// GetAktiveTSEStoerung liefert den aktiven Stoerungszeitraum; nil, wenn keine
// Stoerung aktiv ist.
func (r Repository) GetAktiveTSEStoerung(ctx context.Context) (*tse.Stoerung, error) {
	row, err := r.q.GetAktiveTSEStoerung(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, db.Error(err)
	}
	return &tse.Stoerung{Beginn: row.Beginn, GrundArt: row.GrundArt, Fehlertext: row.Fehlertext}, nil
}
