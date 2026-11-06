-- Idempotenz der Pfandrückgabe-Stornierung über die vom Client vergebene
-- stornierungId, type-gescoped wie idx_kassenjournal_rueckgabe_id. Die ID wird
-- BON_ID im DSFinV-K-Export und muss deshalb über alle Rückgaben eindeutig sein.
--
-- Additiv: nur ein neuer Index, keine Zeile wird geändert.
BEGIN;

CREATE UNIQUE INDEX idx_kassenjournal_pfand_stornierung_id
    ON kassenjournal ((data->>'stornierungId'))
    WHERE type = 'pfandrueckgabe-storniert:v1';

COMMIT;
