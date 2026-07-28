-- name: InsertVorgangIdempotenz :exec
-- Hält Schlüssel und Nutzdaten-Hash eines buchenden Vorgangs fest — im selben
-- Commit wie dessen Events und VOR den Event-Inserts, damit ein
-- Primärschlüssel-Konflikt eindeutig eine Zweiteinreichung desselben Schlüssels
-- ist (Duplikat oder abweichende Nutzdaten, entschieden über payload_hash).
INSERT INTO vorgang_idempotenz (vorgang_id, art, user_id, payload_hash, created_at)
VALUES ($1, $2, $3, $4, now());

-- name: GetVorgangPayloadHash :one
-- Liefert den gespeicherten Nutzdaten-Hash zum Schlüssel; sql.ErrNoRows bedeutet
-- „Schlüssel unbekannt" und damit einen neuen Vorgang. Der Vergleich mit dem Hash
-- der eingereichten Nutzdaten trennt die Duplikat-Einreichung (stille
-- Erfolgsantwort) von abweichenden Nutzdaten (HTTP 409).
SELECT payload_hash FROM vorgang_idempotenz WHERE vorgang_id = $1;
