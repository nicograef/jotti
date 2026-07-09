package seed

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

// Run spielt das Demo-Szenario „3-Tage-Sommerfest TSV Musterstadt e.V." in die Datenbank ein:
// Stammdaten mit Favoriten und Druckstations-Konfiguration, drei Kassensitzungen
// (Freitag/Samstag abgeschlossen, Sonntag offen) und die zugehörigen Events — jedes
// fiskalische davon mit genau einem Signaturauftrag (quittiert, nachsigniert oder offen,
// je nach Ausfallfenster) — plus die Druckauftrags-Historie zu Bestellungen,
// Direktverkäufen und Kassenbelegen. Alles wird in einer Transaktion geschrieben;
// anschließend wird die Tisch-Session-Projektion neu aufgebaut. Ein Guard verhindert das
// Überschreiben einer Datenbank, die bereits Kassenjournal-Events enthält.
func Run(ctx context.Context, database *sql.DB) error {
	return seedInTransaction(ctx, database, false)
}

// ResetAndSeed leert alle Daten-Tabellen und schreibt anschließend den
// Demo-Zustand neu — beides in einer Transaktion. Anders als Run überspringt es
// den Kassenjournal-Guard (das Leeren ist gewollt) und ist ausschließlich für
// den Test-Reset-Endpoint (POST /test/reset-and-seed) gedacht, der wie das
// seed-Subkommando nur bei JOTTI_ALLOW_SEED=1 registriert wird.
func ResetAndSeed(ctx context.Context, database *sql.DB) error {
	return seedInTransaction(ctx, database, true)
}

// seedInTransaction baut den Demo-Zustand auf und schreibt ihn in einer
// Transaktion. Bei reset=true werden zuvor (in derselben Transaktion) alle
// Daten-Tabellen geleert und der Kassenjournal-Guard entfällt; bei reset=false
// gilt der Guard, der ein Überschreiben bestehender Kassenjournal-Events
// verhindert. Anschließend werden die Tisch-Session-Projektionen neu aufgebaut.
func seedInTransaction(ctx context.Context, database *sql.DB, reset bool) error {
	jetzt := time.Now().UTC()
	s := demoSzenario()

	daten, err := buildSeedDaten(s, jetzt)
	if err != nil {
		return fmt.Errorf("seed-daten aufbauen: %w", err)
	}

	fenster := ausfallFensterAus(s, jetzt)
	auftraege, err := buildSignaturauftraege(daten.Events, fenster)
	if err != nil {
		return fmt.Errorf("fake-tse signieren: %w", err)
	}
	stoerungen := stoerungszeitraeumeAus(fenster)

	druckauftraege, err := buildDruckauftraege(s, daten.Events, signaturenNachEventID(auftraege), jetzt)
	if err != nil {
		return fmt.Errorf("druckaufträge aufbauen: %w", err)
	}

	if err := writeSeed(ctx, database, s, daten, auftraege, stoerungen, druckauftraege, jetzt, reset); err != nil {
		return err
	}

	repo := kassenjournal_repo.NewRepository(database)
	if _, err := repo.RebuildAllProjections(ctx); err != nil {
		return fmt.Errorf("projektionen neu aufbauen: %w", err)
	}

	return nil
}

func writeSeed(ctx context.Context, database *sql.DB, s szenario, daten seedDaten, auftraege []signaturauftragZeile, stoerungen []stoerungZeile, druckauftraege []druckauftragZeile, jetzt time.Time, reset bool) error {
	q := dbgen.New(database)

	// Guard: niemals eine Datenbank überschreiben, die bereits Kassenjournal-Events enthält.
	// Die Prüfung läuft ohne Schreibzugriff vor dem Transaktionsbeginn. Beim Test-Reset
	// entfällt sie, weil das Leeren der Tabellen ausdrücklich gewollt ist.
	if !reset {
		anzahl, err := q.SeedCountKassenjournal(ctx)
		if err != nil {
			return fmt.Errorf("kassenjournal prüfen: %w", err)
		}
		if anzahl > 0 {
			return fmt.Errorf("datenbank enthält bereits %d kassenjournal-event(s) — seed abgebrochen ohne Schreibzugriff; zum Zurücksetzen: make clean && make dev, danach make seed", anzahl)
		}
	}

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return db.Error(err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	qtx := q.WithTx(tx)

	// Beim Test-Reset in derselben Transaktion zuerst leeren, dann neu schreiben.
	if reset {
		// Das Kassenjournal ist per Trigger append-only und lässt sich sonst nicht
		// leeren. SET LOCAL session_replication_role = replica deaktiviert die
		// User-Trigger nur für DIESE Transaktion (pool-sicher, setzt sich beim
		// Commit/Rollback selbst zurück). Der Endpoint existiert ohnehin nur bei
		// JOTTI_ALLOW_SEED=1 (Test-/Demo-Umgebung), nie in Produktion.
		// kassenidentitaet ist von SeedTruncateAll bewusst ausgenommen (Install-
		// Identität, insert-once) und bleibt daher unangetastet.
		if _, err := tx.ExecContext(ctx, "SET LOCAL session_replication_role = replica"); err != nil {
			return fmt.Errorf("append-only-schutz für den reset lösen: %w", db.Error(err))
		}
		if err := qtx.SeedTruncateAll(ctx); err != nil {
			return fmt.Errorf("tabellen leeren: %w", err)
		}
		// Die leere tse_konfiguration-Singleton-Zeile wiederherstellen, die die
		// Migration beim Erstlauf anlegt und das Truncate mitgelöscht hat. Ohne
		// sie ist der Ausgangszustand nach einem Reset ein anderer als nach der
		// Erstmigration, und ein Folge-Reseed liefe nicht mehr deterministisch.
		if err := qtx.SeedInsertLeereTSEKonfiguration(ctx); err != nil {
			return fmt.Errorf("leere tse-konfiguration wiederherstellen: %w", err)
		}
	}

	if err := writeStammdaten(ctx, qtx, s, jetzt); err != nil {
		return err
	}
	if err := writeSitzungen(ctx, qtx, daten.Kassensitzungen); err != nil {
		return err
	}
	if err := writeEvents(ctx, qtx, daten.Events); err != nil {
		return err
	}
	if err := writeSignaturauftraege(ctx, qtx, auftraege); err != nil {
		return err
	}
	if err := writeStoerungen(ctx, qtx, stoerungen); err != nil {
		return err
	}
	if err := writeDruckauftraege(ctx, qtx, druckauftraege); err != nil {
		return err
	}
	if err := korrigiereSequenzen(ctx, qtx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return db.Error(err)
	}

	return nil
}

func writeStammdaten(ctx context.Context, qtx *dbgen.Queries, s szenario, jetzt time.Time) error {
	for _, b := range s.Benutzer {
		err := qtx.SeedInsertUser(ctx, dbgen.SeedInsertUserParams{
			ID:           b.ID,
			Name:         b.Name,
			Username:     b.Username,
			PasswordHash: sql.NullString{String: demoArgon2idHash, Valid: true},
			Role:         dbgen.Userrole(b.Rolle),
			Status:       dbgen.Entitystatus(b.Status),
			CreatedAt:    jetzt,
			UpdatedAt:    jetzt,
		})
		if err != nil {
			return fmt.Errorf("benutzer %d einfügen: %w", b.ID, err)
		}
	}

	for _, t := range s.Tische {
		err := qtx.SeedInsertTisch(ctx, dbgen.SeedInsertTischParams{
			ID:        t.ID,
			Name:      t.Name,
			Status:    dbgen.Entitystatus(t.Status),
			CreatedAt: jetzt,
			UpdatedAt: jetzt,
		})
		if err != nil {
			return fmt.Errorf("tisch %d einfügen: %w", t.ID, err)
		}
	}

	for _, p := range s.Produkte {
		err := qtx.SeedInsertProdukt(ctx, dbgen.SeedInsertProduktParams{
			ID:         p.ID,
			Name:       p.Name,
			Kategorie:  dbgen.Produktkategorie(p.Kategorie),
			Steuersatz: dbgen.Steuersatz(p.Steuersatz),
			Status:     dbgen.Entitystatus(p.Status),
			CreatedAt:  jetzt,
			UpdatedAt:  jetzt,
		})
		if err != nil {
			return fmt.Errorf("produkt %d einfügen: %w", p.ID, err)
		}

		for _, v := range p.Varianten {
			err := qtx.SeedInsertVariante(ctx, dbgen.SeedInsertVarianteParams{
				ID:         v.ID,
				ProduktID:  p.ID,
				Name:       v.Name,
				PreisCents: v.PreisCents,
				Status:     dbgen.Entitystatus(v.Status),
				CreatedAt:  jetzt,
				UpdatedAt:  jetzt,
			})
			if err != nil {
				return fmt.Errorf("variante %d einfügen: %w", v.ID, err)
			}
		}
	}

	for _, f := range s.Favoriten {
		err := qtx.AddFavorit(ctx, dbgen.AddFavoritParams{
			UserID:  f.UserID,
			TischID: f.TischID,
		})
		if err != nil {
			return fmt.Errorf("favorit (benutzer %d, tisch %d) einfügen: %w", f.UserID, f.TischID, err)
		}
	}

	err := qtx.UpsertBetreiber(ctx, dbgen.UpsertBetreiberParams{
		Vereinsname:  s.Betreiber.Vereinsname,
		Strasse:      s.Betreiber.Strasse,
		Plz:          s.Betreiber.Plz,
		Ort:          s.Betreiber.Ort,
		Steuernummer: nullString(s.Betreiber.Steuernummer),
		UstID:        nullString(s.Betreiber.UstID),
	})
	if err != nil {
		return fmt.Errorf("betreiber einfügen: %w", err)
	}

	for _, st := range s.Druckstationen {
		err := qtx.UpsertDruckstation(ctx, dbgen.UpsertDruckstationParams{
			Kategorie: dbgen.Druckstationkategorie(st.Kategorie),
			DruckerIp: st.DruckerIP,
			Bonmodus:  sql.NullString{String: string(st.Bonmodus), Valid: st.Bonmodus != ""},
		})
		if err != nil {
			return fmt.Errorf("druckstation %s konfigurieren: %w", st.Kategorie, err)
		}
	}

	// TSE-Stammdaten wie eine echte fiskaly-Einrichtung schreiben, damit die tse.csv
	// des Demo-Exports vollständig ist (sonst bliebe die Singleton-Zeile beim
	// Migrations-Default leer).
	stammdaten := fakeTSEStammdaten()
	if err := qtx.UpsertTSEStammdaten(ctx, dbgen.UpsertTSEStammdatenParams{
		Seriennummer:        stammdaten.Seriennummer,
		SignaturAlgorithmus: stammdaten.SignaturAlgorithmus,
		PublicKey:           stammdaten.PublicKey,
		Zertifikat:          stammdaten.Zertifikat,
		LogTimeFormat:       stammdaten.LogTimeFormat,
	}); err != nil {
		return fmt.Errorf("tse-stammdaten schreiben: %w", err)
	}

	return nil
}

func writeSitzungen(ctx context.Context, qtx *dbgen.Queries, sitzungen []kassensitzungZeile) error {
	for _, k := range sitzungen {
		err := qtx.SeedInsertKassensitzung(ctx, dbgen.SeedInsertKassensitzungParams{
			ZNr:         k.ZNr,
			Datum:       k.Datum,
			Bezeichnung: k.Bezeichnung,
			Status:      string(k.Status),
			CreatedAt:   k.CreatedAt,
			UpdatedAt:   k.UpdatedAt,
		})
		if err != nil {
			return fmt.Errorf("kassensitzung %d einfügen: %w", k.ZNr, err)
		}
	}

	return nil
}

// writeEvents persistiert die Events mit den von der Engine vergebenen IDs — Bondruck-
// Referenzen und Belegnummern verweisen darauf, deshalb läuft der Insert nicht über die
// Identity-Spalte (Sequenzkorrektur in korrigiereSequenzen).
func writeEvents(ctx context.Context, qtx *dbgen.Queries, events []seedEvent) error {
	for i := range events {
		ev := &events[i]
		err := qtx.SeedInsertEvent(ctx, dbgen.SeedInsertEventParams{
			ID:              ev.event.ID,
			UserID:          ev.event.UserID,
			UserName:        ev.event.UserName,
			Type:            ev.event.Type,
			Subject:         ev.event.Subject,
			Version:         ev.event.Version,
			Data:            ev.event.Data,
			Timestamp:       ev.event.Time,
			KassensitzungNr: ev.kassensitzungNr,
		})
		if err != nil {
			return fmt.Errorf("event %s v%d schreiben: %w", ev.event.Subject, ev.event.Version, err)
		}
	}

	return nil
}

func writeDruckauftraege(ctx context.Context, qtx *dbgen.Queries, auftraege []druckauftragZeile) error {
	for i := range auftraege {
		a := &auftraege[i]
		err := qtx.SeedInsertDruckauftrag(ctx, dbgen.SeedInsertDruckauftragParams{
			ZielIp:        a.ZielIP,
			Payload:       a.Payload,
			Status:        a.Status,
			BonArt:        a.BonArt,
			Referenz:      a.Referenz,
			Versuche:      a.Versuche,
			LetzterFehler: nullString(a.LetzterFehler),
			ErstelltAm:    a.ErstelltAm,
			GedrucktAm:    nullTime(a.GedrucktAm),
		})
		if err != nil {
			return fmt.Errorf("druckauftrag %s einfügen: %w", a.Referenz, err)
		}
	}

	return nil
}

// writeSignaturauftraege persistiert die Signaturaufträge; die Signaturspalten sind nur
// bei quittierten (erledigten) Aufträgen gefüllt.
func writeSignaturauftraege(ctx context.Context, qtx *dbgen.Queries, auftraege []signaturauftragZeile) error {
	for i := range auftraege {
		a := &auftraege[i]
		params := dbgen.SeedInsertTSESignaturauftragParams{
			EventID:            a.EventID,
			TxID:               a.TxID,
			ProcessType:        a.ProcessType,
			ProcessData:        a.ProcessData,
			Status:             a.Status,
			Versuche:           a.Versuche,
			LetzterFehler:      nullString(a.LetzterFehler),
			NaechsterVersuchAm: a.NaechsterVersuchAm,
			ErstelltAm:         a.ErstelltAm,
			ErledigtAm:         nullTime(a.ErledigtAm),
		}
		if sig := a.Signatur; sig != nil {
			params.TransaktionNummer = sql.NullInt64{Int64: int64(sig.TransaktionNummer), Valid: true}
			params.SignaturZaehler = sql.NullInt64{Int64: int64(sig.SignaturZaehler), Valid: true}
			params.TseSeriennummer = sql.NullString{String: sig.TSESeriennummer, Valid: true}
			params.LogTimeStart = sql.NullTime{Time: sig.LogTimeStart, Valid: true}
			params.LogTimeEnd = sql.NullTime{Time: sig.LogTimeEnd, Valid: true}
			params.Signatur = sql.NullString{String: sig.Signatur, Valid: true}
			params.QrCodeData = sql.NullString{String: sig.QRCodeData, Valid: true}
		}
		if err := qtx.SeedInsertTSESignaturauftrag(ctx, params); err != nil {
			return fmt.Errorf("signaturauftrag %s einfügen: %w", a.TxID, err)
		}
	}

	return nil
}

// writeStoerungen persistiert die geschlossenen Störungszeiträume der aufgelösten
// Ausfallfenster (Störungsprotokoll / Ausfalldokumentation).
func writeStoerungen(ctx context.Context, qtx *dbgen.Queries, stoerungen []stoerungZeile) error {
	for _, st := range stoerungen {
		err := qtx.SeedInsertTSEStoerung(ctx, dbgen.SeedInsertTSEStoerungParams{
			Beginn:     st.Beginn,
			Ende:       sql.NullTime{Time: st.Ende, Valid: true},
			GrundArt:   st.GrundArt,
			Fehlertext: st.Fehlertext,
		})
		if err != nil {
			return fmt.Errorf("störungszeitraum ab %s einfügen: %w", st.Beginn, err)
		}
	}

	return nil
}

// korrigiereSequenzen zieht die IDENTITY-Sequenzen auf den höchsten manuell vergebenen Wert nach,
// damit anschließend per Anwendung erzeugte Datensätze keine Primärschlüssel-Kollision auslösen.
func korrigiereSequenzen(ctx context.Context, qtx *dbgen.Queries) error {
	if err := qtx.SeedResetUsersSeq(ctx); err != nil {
		return fmt.Errorf("users-sequenz nachziehen: %w", err)
	}
	if err := qtx.SeedResetTischeSeq(ctx); err != nil {
		return fmt.Errorf("tische-sequenz nachziehen: %w", err)
	}
	if err := qtx.SeedResetProdukteSeq(ctx); err != nil {
		return fmt.Errorf("produkte-sequenz nachziehen: %w", err)
	}
	if err := qtx.SeedResetVariantenSeq(ctx); err != nil {
		return fmt.Errorf("varianten-sequenz nachziehen: %w", err)
	}
	if err := qtx.SeedResetKassensitzungenSeq(ctx); err != nil {
		return fmt.Errorf("kassensitzungen-sequenz nachziehen: %w", err)
	}
	if err := qtx.SeedResetKassenjournalSeq(ctx); err != nil {
		return fmt.Errorf("kassenjournal-sequenz nachziehen: %w", err)
	}

	return nil
}

func nullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
