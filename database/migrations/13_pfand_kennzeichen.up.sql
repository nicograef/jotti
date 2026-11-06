-- Pfand-Kennzeichen je Verkaufsposition: zahlung-kassiert:v2 und
-- direktverkauf-getaetigt:v2 sind v1 plus "pfand" je Position. Der
-- DSFinV-K-Export bucht damit Pfand-Zeilen unter GV_TYP Pfand; v1-Events
-- bleiben Umsatz.
--
-- Additiv: die Funktionen liefern für alle bisherigen Event-Typen unverändert
-- dasselbe, der neue Index deckt nur die neue Version mit ab.
BEGIN;

-- Idempotenz des Direktverkaufs über die verkaufId, versionsübergreifend.
CREATE UNIQUE INDEX idx_kassenjournal_verkauf_id_v2
    ON kassenjournal ((data->>'verkaufId'))
    WHERE type IN ('direktverkauf-getaetigt:v1', 'direktverkauf-getaetigt:v2');

-- Returns the Zahlung amount in cents for a matching event row, NULL otherwise.
CREATE OR REPLACE FUNCTION kj_extract_zahlung_cents(type TEXT, data JSONB) RETURNS int AS $$
    SELECT CASE WHEN type IN ('zahlung-kassiert:v1', 'zahlung-kassiert:v2') THEN (data->>'gesamtZahlungCents')::int END
$$ LANGUAGE sql IMMUTABLE;

-- Returns the Direktverkauf amount in cents for a matching event row, NULL otherwise.
CREATE OR REPLACE FUNCTION kj_extract_direktverkauf_cents(type TEXT, data JSONB) RETURNS int AS $$
    SELECT CASE WHEN type IN ('direktverkauf-getaetigt:v1', 'direktverkauf-getaetigt:v2') THEN (data->>'gesamtbetragCents')::int END
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
    WHERE type IN ('zahlung-kassiert:v1', 'zahlung-kassiert:v2', 'direktverkauf-getaetigt:v1', 'direktverkauf-getaetigt:v2', 'direktverkauf-storniert:v1', 'stornierung-erteilt:v1', 'pfand-zurueckgegeben:v1', 'pfandrueckgabe-storniert:v1')
$$ LANGUAGE sql IMMUTABLE;

COMMIT;
