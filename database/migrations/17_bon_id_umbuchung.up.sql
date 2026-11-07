-- Die umbuchungId wählt der Client als Idempotenz-Schlüssel, und der Abgang
-- exportiert sie als BON_ID. Sie gehört deshalb in den BON_ID-Index aus
-- Migration 14: Eine umbuchungId, die schon eine andere Vorgangs-ID ist oder
-- auf einem anderen Quelltisch gebucht wurde, bucht nichts.
--
-- Erfasst ist nur der Abgang, das Event im Stream des Quelltischs. Der Zugang
-- trägt dieselbe umbuchungId; für ihn liefert der Ausdruck NULL, und NULL ist
-- im eindeutigen Index nie ein Duplikat.
--
-- Additiv: der Index wird mit einem Typ mehr neu angelegt, keine Zeile ändert sich.
BEGIN;

DROP INDEX idx_kassenjournal_bon_id;

CREATE UNIQUE INDEX idx_kassenjournal_bon_id
    ON kassenjournal ((CASE type
        WHEN 'bestellung-aufgenommen:v1' THEN data->>'bestellungId'
        WHEN 'zahlung-kassiert:v1' THEN data->>'zahlungId'
        WHEN 'zahlung-kassiert:v2' THEN data->>'zahlungId'
        WHEN 'stornierung-erteilt:v1' THEN data->>'stornierungId'
        WHEN 'bestellung-korrigiert:v1' THEN data->>'korrekturId'
        WHEN 'bestellung-umgebucht:v1' THEN CASE
            WHEN split_part(subject, '/tisch-', 2) = data->>'quellTischId'
            THEN data->>'umbuchungId'
        END
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
        'bestellung-umgebucht:v1',
        'direktverkauf-getaetigt:v1',
        'direktverkauf-getaetigt:v2',
        'direktverkauf-storniert:v1',
        'pfand-zurueckgegeben:v1',
        'pfandrueckgabe-storniert:v1',
        'geldtransit-gebucht:v1'
    );

COMMIT;
