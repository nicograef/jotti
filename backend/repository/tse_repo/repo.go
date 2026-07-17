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
		TransaktionNummer: sql.NullInt64{Int64: int64(signatur.TransaktionNummer), Valid: true},
		SignaturZaehler:   sql.NullInt64{Int64: int64(signatur.SignaturZaehler), Valid: true},
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

// MarkOffeneAlsNichtKonfiguriert markiert alle offenen Auftraege endgueltig
// als tse_nicht_konfiguriert und liefert die Anzahl markierter Auftraege. Ohne
// vorhandene TSE-Konfiguration gibt es keine Signatur; ein Nachsignieren ist
// ausgeschlossen (keine Fehlversuche, keine automatische Wiederaufnahme).
// Bereits endgueltig markierte Auftraege bleiben unberuehrt.
func (r Repository) MarkOffeneAlsNichtKonfiguriert(ctx context.Context) (int64, error) {
	n, err := r.q.MarkOffeneTSESignaturauftraegeNichtKonfiguriert(ctx)
	if err != nil {
		return 0, db.Error(err)
	}
	return n, nil
}

// GetTSESignaturQueueZustand liefert den on demand berechneten Zustand der
// Signatur-Queue fuer das Admin-Monitoring.
func (r Repository) GetTSESignaturQueueZustand(ctx context.Context) (tse.SignaturQueueZustand, error) {
	row, err := r.q.GetTSESignaturQueueZustand(ctx)
	if err != nil {
		return tse.SignaturQueueZustand{}, db.Error(err)
	}
	return tse.SignaturQueueZustand{
		OffeneAuftraege:          row.OffeneAuftraege,
		FehlgeschlageneAuftraege: row.FehlgeschlageneAuftraege,
		LetzterFehler:            row.LetzterFehler,
		RueckstandSekunden:       row.RueckstandSekunden,
		SignaturenProMinute:      row.SignaturenProMinute,
		SignierdauerP95Sekunden:  row.SignierdauerP95Sekunden,
	}, nil
}

// GetAlleTSEStoerungen liefert das Stoerungsprotokoll (Ausfalldokumentation):
// alle Stoerungszeitraeume, neueste zuerst.
func (r Repository) GetAlleTSEStoerungen(ctx context.Context) ([]tse.Stoerungszeitraum, error) {
	rows, err := r.q.GetAlleTSEStoerungen(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	result := make([]tse.Stoerungszeitraum, 0, len(rows))
	for i := range rows {
		row := &rows[i]
		zeitraum := tse.Stoerungszeitraum{
			ID:         row.ID,
			Beginn:     row.Beginn,
			GrundArt:   row.GrundArt,
			Fehlertext: row.Fehlertext,
		}
		if row.Ende.Valid {
			ende := row.Ende.Time
			zeitraum.Ende = &ende
		}
		result = append(result, zeitraum)
	}

	return result, nil
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
			TransaktionNummer: int(row.TransaktionNummer.Int64),
			SignaturZaehler:   int(row.SignaturZaehler.Int64),
			TSESeriennummer:   row.TseSeriennummer.String,
			LogTimeStart:      row.LogTimeStart.Time,
			LogTimeEnd:        row.LogTimeEnd.Time,
			Signatur:          row.Signatur.String,
			QRCodeData:        row.QrCodeData.String,
		}
	}
	return stand, nil
}

// GetOffeneSignaturauftragStaendeFuerKassensitzung liefert die Signatur-Staende
// aller noch nicht erledigten Signaturauftraege der Kassensitzung — die
// Grundlage des Kassenabschluss-Gates. Erledigte Auftraege bleiben aussen vor
// (bereits signiert); das Gate ordnet die Staende ueber DetermineSignaturstatus
// in ausstehend bzw. Ausfall ein.
func (r Repository) GetOffeneSignaturauftragStaendeFuerKassensitzung(ctx context.Context, kassensitzungNr int) ([]tse.SignaturauftragStand, error) {
	rows, err := r.q.GetOffeneSignaturauftragStaendeFuerKassensitzung(ctx, kassensitzungNr)
	if err != nil {
		return nil, db.Error(err)
	}

	result := make([]tse.SignaturauftragStand, 0, len(rows))
	for _, row := range rows {
		result = append(result, tse.SignaturauftragStand{Status: row.Status, ErstelltAm: row.ErstelltAm})
	}
	return result, nil
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

// OpenTSEStoerung oeffnet einen Stoerungszeitraum im Stoerungsprotokoll.
// Idempotent: Solange irgendein Zeitraum aktiv ist, ist das Oeffnen ein No-Op
// (hoechstens ein aktiver Zeitraum, DB-seitig per partiellem Unique-Index).
func (r Repository) OpenTSEStoerung(ctx context.Context, grundArt string, fehlertext string) error {
	return db.Error(r.q.OpenTSEStoerung(ctx, dbgen.OpenTSEStoerungParams{
		GrundArt:   grundArt,
		Fehlertext: fehlertext,
	}))
}

// CloseTSEStoerung beendet den aktiven Stoerungszeitraum der Grund-Art;
// jeder Schreiber schliesst nur Zeitraeume seiner Grund-Art. Idempotent: Ohne
// aktiven Zeitraum der Art ein No-Op.
func (r Repository) CloseTSEStoerung(ctx context.Context, grundArt string) error {
	return db.Error(r.q.CloseTSEStoerung(ctx, grundArt))
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
