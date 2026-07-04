# Audit TSE-Signatur-Outbox: Befunde mit Entscheidungsbedarf

Ergänzung zum [Audit-Bericht](audit-tse-signatur-outbox-bericht.md) (2026-07-04). Diese fünf Punkte sind bewusst
nicht umgesetzt: Sie brauchen eine Produktentscheidung oder wiegen Aufwand gegen einen Randfall ab. Je Punkt:
Befund, Optionen, Empfehlung.

## B1 Verwerfen-Race: Signatur bei fiskaly, Auftrag verworfen

Befund: Verwirft ein Admin einen offenen Auftrag, während der Worker ihn gerade signiert, gewinnt der
Statuswechsel: `QuittiereTSESignaturauftrag` ist durch den Status-Guard (`status='offen'`) ein stiller No-Op
(RowsAffected wird nicht geprüft, `backend/repository/tse_repo/repo.go:105`). Die fertige Signatur existiert dann
nur im TSE-Log (TAR-Export), der Auftrag bleibt `verworfen`, der DSFinV-K-Export schreibt eine
TSE_TA_FEHLER-Zeile. Bei einer Prüfung weichen TSE-Export und Kassen-Export für diesen einen Vorgang ab
(erklärbar über das Verwerfen-Protokoll mit Grund, Benutzer, Zeitpunkt).

Optionen:

1. RowsAffected prüfen und den Fall als Warnung loggen (Sichtbarkeit, keine Verhaltensänderung).
2. Verwerfen nur noch für `fehlgeschlagen` zulassen (offene Aufträge erst nach endgültigem Fehlschlag
   verwerfbar; nimmt dem Admin die Möglichkeit, einen frisch eingereihten Gift-Auftrag sofort zu entsorgen).
3. Belassen und als dokumentierten Randfall akzeptieren.

Empfehlung: Option 1 (kleiner Eingriff, macht den Randfall im Log nachweisbar); Option 2 nur, falls das
Verwerfen offener Aufträge fachlich ohnehin nicht gewollt ist.

## B2 Admin-Aktionen ohne Wirkungs-Rückmeldung

Befund: Zurücksetzen und Verwerfen antworten 200 auch dann, wenn der Status-Guard nichts geändert hat (falscher
Status oder unbekannte ID). Eine UI mit veraltetem Listenstand meldet dann Erfolg ohne Wirkung; erst das
Neuladen der Liste zeigt die Wahrheit (die Hooks laden nach jeder Aktion neu, das Fenster ist also klein).

Optionen:

1. RowsAffected prüfen und bei 0 mit 409 (`statuswechsel_nicht_moeglich`) antworten; UI zeigt eine passende
   Meldung.
2. Belassen (idempotente Semantik: „Zielzustand erreicht oder Aktion nicht mehr anwendbar").

Empfehlung: Option 1, konsistent mit dem übrigen Fehlerbild der Admin-Endpunkte (409-Codes existieren bereits).

## B3 Signaturauftrags-Liste hart auf 200 Einträge begrenzt

Befund: `GetTSESignaturauftraege` liefert die neuesten 200 Aufträge (`tse_signaturauftraege.sql:92`), die UI
zeigt die Kappung nicht an. An einem großen Festtag sind 200 Vorgänge schnell erreicht; ältere fehlgeschlagene
Aufträge könnten aus der Ansicht fallen (das Zurücksetzen-gesamt wirkt unabhängig davon auf alle).

Optionen:

1. Hinweiszeile in der UI, sobald 200 Einträge geliefert wurden („älteste Einträge ausgeblendet").
2. Filter auf nicht-erledigte Status plus Limit (die Verwaltungssicht braucht erledigte Aufträge selten).
3. Paginierung (größter Aufwand, für die Zielgruppe vermutlich überdimensioniert).

Empfehlung: Option 2, ergänzt um Option 1 als Ein-Zeilen-Hinweis.

## B4 AEAO zu § 146a Nr. 2.2.2 („unmittelbar mit Beginn") explizit adressieren

Befund: Das Outbox-Modell startet die TSE-Transaktion nicht „unmittelbar mit Beginn des Vorgangs" im Wortsinn,
sondern Sekunden später durch den Worker (Start und Finish unmittelbar nacheinander). Das ist die fixierte
PRD-Entscheidung; compliance.md §3.8 und verfahrensdokumentation.md §4 beschreiben Mechanismus und Latenz
ehrlich, nennen die AEAO-Textstelle aber nicht ausdrücklich.

Optionen:

1. Ein Absatz in compliance.md §3.8, der die Spannung benennt und die Einordnung begründet (geringe,
   dokumentierte Latenz; keine still unsignierten Vorgänge; Störungsprotokoll; Nachsigniert-Ausweis) — maximale
   Transparenz gegenüber Prüfern und künftigen Maintainer:innen.
2. Belassen (die Konformitätsbedingungen sind beschrieben, nur der Normbezug fehlt).

Empfehlung: Option 1 (ein Absatz, kein Codeaufwand).

## B5 Seed ohne historische Störungszeiträume

Befund: Das Demo-Szenario erzeugt nachsignierte Belege aus aufgelösten Ausfallfenstern, schreibt aber keine
`tse_stoerungen`-Zeilen. Die Ausfalldokumentations-Ansicht zeigt in der Demo „Keine Störungen", obwohl
nachsignierte Vorgänge sichtbar sind — für Vorführungen inkonsistent. (Die offenen Aufträge der laufenden
Sitzung konvergieren nach dem Start über Worker und Watchdog ohnehin live.)

Optionen:

1. Je aufgelöstem Ausfallfenster einen geschlossenen `tse_fehler`-Zeitraum seeden.
2. Belassen (Störungsprotokoll füllt sich nur durch echtes Laufzeitverhalten).

Empfehlung: Option 1 (kleine Ergänzung in `backend/seed`, macht die Demo in sich stimmig).
