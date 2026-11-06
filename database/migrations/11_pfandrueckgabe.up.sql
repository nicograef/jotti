-- Pfandrückgabe: Ein Gast gibt Pfandgut zurück und erhält den Pfandbetrag bar
-- oder verrechnet. Jede Rückgabe ist ein eigener Kassenvorgang
-- (pfand-zurueckgegeben:v1), ihre Stornierung ebenfalls
-- (pfandrueckgabe-storniert:v1).
--
-- Additiv: die neue Spalte hat einen Default, bestehende Zeilen bleiben gültig.
-- Die ersetzte Umsatz-Funktion liefert für alle bisherigen Event-Typen
-- unverändert dasselbe.
BEGIN;

ALTER TABLE produkte ADD COLUMN pfand BOOLEAN NOT NULL DEFAULT false;

COMMENT ON COLUMN produkte.pfand IS 'Pfand-Produkt: seine Varianten bietet die Pfandrückgabe an. Bestellung, Arbeitsbon und Kategorie bleiben davon unberührt.';

-- Idempotenz der Pfandrückgabe über die vom Client vergebene rueckgabeId,
-- type-gescoped wie idx_kassenjournal_verkauf_id: die Stornierung trägt dieselbe ID.
CREATE UNIQUE INDEX idx_kassenjournal_rueckgabe_id
    ON kassenjournal ((data->>'rueckgabeId'))
    WHERE type = 'pfand-zurueckgegeben:v1';

-- Returns the Pfandrückgabe amount in cents for a matching event row, NULL otherwise.
-- Unsigned like kj_extract_direktverkauf_cents; the query sets the sign.
CREATE FUNCTION kj_extract_pfandrueckgabe_cents(type TEXT, data JSONB) RETURNS int AS $$
    SELECT CASE WHEN type = 'pfand-zurueckgegeben:v1' THEN (data->>'gesamtbetragCents')::int END
$$ LANGUAGE sql IMMUTABLE;

-- Returns the Pfandrückgabe-Storno amount in cents for a matching event row, NULL otherwise.
-- Unsigned like kj_extract_direktverkauf_storno_cents; the query sets the sign.
CREATE FUNCTION kj_extract_pfandrueckgabe_storno_cents(type TEXT, data JSONB) RETURNS int AS $$
    SELECT CASE WHEN type = 'pfandrueckgabe-storniert:v1' THEN (data->>'gesamtStornierungCents')::int END
$$ LANGUAGE sql IMMUTABLE;

-- Returns the gross cents per steuersatz from position arrays for Umsatz aggregation.
-- Warenrücknahme (stornierung-erteilt), Direktverkauf-Storno and Pfandrückgabe count
-- as negative Umsatz; the Storno of a Pfandrückgabe counts as positive Umsatz.
CREATE OR REPLACE FUNCTION kj_extract_umsatz_pro_steuersatz(type TEXT, data JSONB)
RETURNS TABLE(steuersatz Steuersatz, brutto_cents int) AS $$
    SELECT
        (position->>'steuersatz')::Steuersatz AS steuersatz,
        ((position->>'einzelpreisCents')::int * (position->>'menge')::int)
            * CASE WHEN type IN ('direktverkauf-storniert:v1', 'stornierung-erteilt:v1', 'pfand-zurueckgegeben:v1') THEN -1 ELSE 1 END AS brutto_cents
    FROM jsonb_array_elements(data->'positionen') AS position
    WHERE type IN ('zahlung-kassiert:v1', 'direktverkauf-getaetigt:v1', 'direktverkauf-storniert:v1', 'stornierung-erteilt:v1', 'pfand-zurueckgegeben:v1', 'pfandrueckgabe-storniert:v1')
$$ LANGUAGE sql IMMUTABLE;

COMMIT;
