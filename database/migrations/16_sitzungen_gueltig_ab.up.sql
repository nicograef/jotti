-- Ein zurückgesetztes oder neu gesetztes Passwort beendet die laufenden
-- Sitzungen des Benutzers: ein Token, das vor sitzungen_gueltig_ab ausgestellt
-- wurde, gilt nicht mehr. So sperrt ein Passwort-Reset ein verlorenes Telefon aus.
--
-- Additiv: eine neue, leere Spalte; NULL heißt, keine Sitzung wurde je beendet.
BEGIN;

ALTER TABLE users ADD COLUMN sitzungen_gueltig_ab TIMESTAMPTZ;

COMMENT ON COLUMN users.sitzungen_gueltig_ab IS 'Frühester gültiger Ausstellungszeitpunkt (iat) eines Tokens, auf volle Sekunden gekürzt wie iat; gesetzt bei Passwort-Reset und Passwortwechsel. NULL = keine Einschränkung.';

COMMIT;
