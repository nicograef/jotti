-- Die JSON-Darstellung einer Variante, wie sie das Backend in den
-- Produkt-Antworten (varianten-Array) ausliefert. Die drei Lese-Queries in
-- backend/sqlc/queries/produkte.sql aggregieren diese Einträge mit je eigenem
-- Status-Filter; die Feldliste steht damit nur hier und muss mit jsonVariante in
-- backend/repository/produkt_repo/types.go übereinstimmen.
BEGIN;

CREATE VIEW produkt_varianten_json AS
SELECT
    id,
    produkt_id,
    status,
    reihenfolge,
    json_build_object(
        'id', id,
        'name', name,
        'preisCents', preis_cents,
        'status', status,
        'createdAt', created_at,
        'updatedAt', updated_at
    ) AS eintrag
FROM produkt_varianten;

COMMIT;
