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
// (Freitag/Samstag abgeschlossen, Sonntag offen) und die zugehörigen Events — die fiskalischen
// davon mit Fake-TSE-Daten samt Ausfallfenster und Nachsignier-Aufträgen — plus die
// Druckauftrags-Historie zu Bestellungen, Direktverkäufen und Kassenbelegen. Alles wird in
// einer Transaktion geschrieben; anschließend wird die Tisch-Session-Projektion neu aufgebaut.
// Ein Guard verhindert das Überschreiben einer Datenbank, die bereits Kassenjournal-Events enthält.
func Run(ctx context.Context, database *sql.DB) error {
	jetzt := time.Now().UTC()
	s := demoSzenario()

	daten, err := buildSeedDaten(s, jetzt)
	if err != nil {
		return fmt.Errorf("seed-daten aufbauen: %w", err)
	}

	tseDaten, err := signiereEvents(daten.Events, ausfallFensterAus(s, jetzt))
	if err != nil {
		return fmt.Errorf("fake-tse signieren: %w", err)
	}

	druckauftraege, err := baueDruckauftraege(s, daten.Events, jetzt)
	if err != nil {
		return fmt.Errorf("druckaufträge aufbauen: %w", err)
	}

	if err := schreibeSeed(ctx, database, s, daten, tseDaten, druckauftraege, jetzt); err != nil {
		return err
	}

	repo := kassenjournal_repo.NewRepository(database)
	if _, err := repo.RebuildAllProjections(ctx); err != nil {
		return fmt.Errorf("projektionen neu aufbauen: %w", err)
	}

	return nil
}

func schreibeSeed(ctx context.Context, database *sql.DB, s szenario, daten seedDaten, tseDaten tseSeitentabellen, druckauftraege []druckauftragZeile, jetzt time.Time) error {
	q := dbgen.New(database)

	// Guard: niemals eine Datenbank überschreiben, die bereits Kassenjournal-Events enthält.
	// Die Prüfung läuft ohne Schreibzugriff vor dem Transaktionsbeginn.
	anzahl, err := q.SeedCountKassenjournal(ctx)
	if err != nil {
		return fmt.Errorf("kassenjournal prüfen: %w", err)
	}
	if anzahl > 0 {
		return fmt.Errorf("datenbank enthält bereits %d kassenjournal-event(s) — seed abgebrochen ohne Schreibzugriff; zum Zurücksetzen: make clean && make dev, danach make seed", anzahl)
	}

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return db.Error(err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	qtx := q.WithTx(tx)

	if err := schreibeStammdaten(ctx, qtx, s, jetzt); err != nil {
		return err
	}
	if err := schreibeSitzungen(ctx, qtx, daten.Kassensitzungen); err != nil {
		return err
	}
	if err := schreibeEvents(ctx, qtx, daten.Events); err != nil {
		return err
	}
	if err := schreibeTSESeitentabellen(ctx, qtx, tseDaten); err != nil {
		return err
	}
	if err := schreibeDruckauftraege(ctx, qtx, druckauftraege); err != nil {
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

func schreibeStammdaten(ctx context.Context, qtx *dbgen.Queries, s szenario, jetzt time.Time) error {
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

	return nil
}

func schreibeSitzungen(ctx context.Context, qtx *dbgen.Queries, sitzungen []kassensitzungZeile) error {
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

// schreibeEvents persistiert die Events mit den von der Engine vergebenen IDs — Bondruck-
// Referenzen und Belegnummern verweisen darauf, deshalb läuft der Insert nicht über die
// Identity-Spalte (Sequenzkorrektur in korrigiereSequenzen).
func schreibeEvents(ctx context.Context, qtx *dbgen.Queries, events []seedEvent) error {
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

func schreibeDruckauftraege(ctx context.Context, qtx *dbgen.Queries, auftraege []druckauftragZeile) error {
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

func schreibeTSESeitentabellen(ctx context.Context, qtx *dbgen.Queries, daten tseSeitentabellen) error {
	for i := range daten.Auftraege {
		a := &daten.Auftraege[i]
		err := qtx.SeedInsertTSENachsignierAuftrag(ctx, dbgen.SeedInsertTSENachsignierAuftragParams{
			TxID:               a.TxID,
			ProcessType:        a.ProcessType,
			ProcessData:        a.ProcessData,
			Status:             a.Status,
			Versuche:           a.Versuche,
			LetzterFehler:      nullString(a.LetzterFehler),
			NaechsterVersuchAm: a.NaechsterVersuchAm,
			ErstelltAm:         a.ErstelltAm,
			ErledigtAm:         nullTime(a.ErledigtAm),
		})
		if err != nil {
			return fmt.Errorf("nachsignier-auftrag %s einfügen: %w", a.TxID, err)
		}
	}

	for i := range daten.Signaturen {
		sig := &daten.Signaturen[i]
		err := qtx.SeedInsertTSESignatur(ctx, dbgen.SeedInsertTSESignaturParams{
			TxID:              sig.TxID,
			TransaktionNummer: sig.TransaktionNummer,
			SignaturZaehler:   sig.SignaturZaehler,
			TseSeriennummer:   sig.TSESeriennummer,
			LogTimeStart:      sig.LogTimeStart,
			LogTimeEnd:        sig.LogTimeEnd,
			Signatur:          sig.Signatur,
			QrCodeData:        sig.QRCodeData,
			ErstelltAm:        sig.ErstelltAm,
		})
		if err != nil {
			return fmt.Errorf("tse-signatur %s einfügen: %w", sig.TxID, err)
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
