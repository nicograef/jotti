-- name: InsertVorgangIdempotenz :exec
-- Hält den client-gelieferten Idempotenz-Schlüssel eines buchenden Vorgangs
-- fest — im selben Commit wie dessen Events und VOR den Event-Inserts, damit
-- ein Primärschlüssel-Konflikt eindeutig eine Duplikat-Einreichung ist.
INSERT INTO vorgang_idempotenz (vorgang_id, art, user_id, created_at)
VALUES ($1, $2, $3, now());

-- name: ExistsVorgangIdempotenz :one
-- Duplikat-Vorprüfung der Commands VOR der fachlichen Validierung: Ein
-- Wiederholversuch nach erfolgreicher Buchung darf nicht an inzwischen
-- geänderten Invarianten scheitern, sondern wiederholt die Erfolgsantwort.
SELECT EXISTS(SELECT 1 FROM vorgang_idempotenz WHERE vorgang_id = $1);
