-- Eine Vorgangs-ID ist im DSFinV-K-Export die BON_ID und darf deshalb über alle
-- Event-Typen nur einmal vorkommen. Die type-gescopten Idempotenz-Indexe sehen
-- jeden Typ einzeln; dieser Index schließt die Lücke zwischen den Typen. Ein
-- Client, der etwa die rueckgabeId einer Rückgabe als stornierungId einer
-- anderen schickt, bucht damit nichts.
--
-- Erfasst ist jede ID, die als BON_ID exportiert wird und die ein Client senden
-- oder aus der Historie kennen kann. Die umbuchungId fehlt: Abgang und Zugang
-- tragen sie gemeinsam, und kein Client bekommt sie zu sehen.
--
-- Additiv: nur ein neuer Index, keine Zeile wird geändert.
BEGIN;

CREATE UNIQUE INDEX idx_kassenjournal_bon_id
    ON kassenjournal ((CASE type
        WHEN 'bestellung-aufgenommen:v1' THEN data->>'bestellungId'
        WHEN 'zahlung-kassiert:v1' THEN data->>'zahlungId'
        WHEN 'zahlung-kassiert:v2' THEN data->>'zahlungId'
        WHEN 'stornierung-erteilt:v1' THEN data->>'stornierungId'
        WHEN 'bestellung-korrigiert:v1' THEN data->>'korrekturId'
        WHEN 'direktverkauf-getaetigt:v1' THEN data->>'verkaufId'
        WHEN 'direktverkauf-getaetigt:v2' THEN data->>'verkaufId'
        WHEN 'direktverkauf-storniert:v1' THEN data->>'stornierungId'
        WHEN 'pfand-zurueckgegeben:v1' THEN data->>'rueckgabeId'
        WHEN 'pfandrueckgabe-storniert:v1' THEN data->>'stornierungId'
        WHEN 'geldtransit-gebucht:v1' THEN data->>'geldtransitId'
    END))
    WHERE type IN (
        'bestellung-aufgenommen:v1',
        'zahlung-kassiert:v1',
        'zahlung-kassiert:v2',
        'stornierung-erteilt:v1',
        'bestellung-korrigiert:v1',
        'direktverkauf-getaetigt:v1',
        'direktverkauf-getaetigt:v2',
        'direktverkauf-storniert:v1',
        'pfand-zurueckgegeben:v1',
        'pfandrueckgabe-storniert:v1',
        'geldtransit-gebucht:v1'
    );

COMMIT;
