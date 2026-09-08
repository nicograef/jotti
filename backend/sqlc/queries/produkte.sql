-- name: GetProdukt :one
SELECT
    p.id,
    p.name,
    p.kategorie,
    p.steuersatz,
    p.status,
    p.created_at,
    p.updated_at,
    COALESCE(
        (SELECT json_agg(
            json_build_object(
                'id', pv.id,
                'name', pv.name,
                'preisCents', pv.preis_cents,
                'status', pv.status,
                'createdAt', pv.created_at,
                'updatedAt', pv.updated_at
            )
            ORDER BY pv.reihenfolge, pv.id
        )
        FROM produkt_varianten pv
        WHERE pv.produkt_id = p.id AND pv.status != 'deleted'),
        '[]'
    )::json AS varianten
FROM produkte p
WHERE p.id = $1 AND p.status != 'deleted';

-- name: GetAlleProdukte :many
WITH varianten_json AS (
    SELECT
        produkt_id,
        json_agg(
            json_build_object(
                'id', id,
                'name', name,
                'preisCents', preis_cents,
                'status', status,
                'createdAt', created_at,
                'updatedAt', updated_at
            )
            ORDER BY reihenfolge, id
        ) AS varianten
    FROM produkt_varianten
    WHERE status != 'deleted'
    GROUP BY produkt_id
)
SELECT
    p.id,
    p.name,
    p.kategorie,
    p.steuersatz,
    p.status,
    p.created_at,
    p.updated_at,
    COALESCE(vj.varianten, '[]')::json AS varianten
FROM produkte p
LEFT JOIN varianten_json vj ON vj.produkt_id = p.id
WHERE p.status != 'deleted'
ORDER BY p.kategorie, p.reihenfolge, p.id;

-- name: GetAktiveProdukte :many
-- Bestelliste für den Service: nur aktive Produkte mit mindestens einer aktiven Variante.
-- Der INNER JOIN blendet aktive Produkte ohne aktive (bepreiste) Variante bewusst aus,
-- da sie nicht bestellbar sind. Die Admin-Sicht (GetAlleProdukte) zeigt sie via LEFT JOIN.
WITH varianten_json AS (
    SELECT
        produkt_id,
        json_agg(
            json_build_object(
                'id', id,
                'name', name,
                'preisCents', preis_cents,
                'status', status,
                'createdAt', created_at,
                'updatedAt', updated_at
            )
            ORDER BY reihenfolge, id
        ) AS varianten
    FROM produkt_varianten
    WHERE status = 'active'
    GROUP BY produkt_id
)
SELECT
    p.id,
    p.name,
    p.kategorie,
    p.steuersatz,
    p.status,
    p.created_at,
    p.updated_at,
    vj.varianten::json AS varianten
FROM produkte p
INNER JOIN varianten_json vj ON vj.produkt_id = p.id
WHERE p.status = 'active'
ORDER BY p.kategorie, p.reihenfolge, p.id;

-- name: CreateProdukt :one
-- Neue Produkte landen ans Ende ihrer Kategorie. Die Reihenfolge wird in der
-- Datenbank berechnet, damit zwischen Lesen und Schreiben keine Lücke entsteht.
INSERT INTO produkte (name, kategorie, steuersatz, status, created_at, updated_at, reihenfolge)
VALUES ($1, $2, $3, $4, $5, $6,
    COALESCE((SELECT MAX(reihenfolge) + 1 FROM produkte WHERE kategorie = $2), 1))
RETURNING id;

-- name: UpdateProdukt :execresult
-- Wechselt das Produkt die Kategorie, hängt es sich ans Ende der neuen an:
-- die alte Reihenfolge gilt dort nicht und träfe womöglich auf den Wert
-- einer bestehenden Zeile. Bleibt die Kategorie, bleibt auch die Position.
UPDATE produkte p SET
    name = $1,
    kategorie = $2,
    steuersatz = $3,
    status = $4,
    updated_at = $5,
    reihenfolge = CASE
        WHEN p.kategorie = $2 THEN p.reihenfolge
        ELSE COALESCE((SELECT MAX(zk.reihenfolge) + 1 FROM produkte zk WHERE zk.kategorie = $2), 1)
    END
WHERE p.id = $6;

-- name: GetProduktReihenfolge :one
SELECT id, kategorie, reihenfolge FROM produkte WHERE id = $1 AND status != 'deleted';

-- name: GetProduktVorgaenger :one
-- Das in der Sortierung (reihenfolge, id) unmittelbar davor liegende Produkt
-- derselben Kategorie. Kein Treffer bedeutet: das Produkt steht bereits oben.
SELECT id, reihenfolge FROM produkte
WHERE kategorie = sqlc.arg(kategorie)
  AND status != 'deleted'
  AND (reihenfolge < sqlc.arg(reihenfolge)
       OR (reihenfolge = sqlc.arg(reihenfolge) AND id < sqlc.arg(id)))
ORDER BY reihenfolge DESC, id DESC
LIMIT 1;

-- name: GetProduktNachfolger :one
SELECT id, reihenfolge FROM produkte
WHERE kategorie = sqlc.arg(kategorie)
  AND status != 'deleted'
  AND (reihenfolge > sqlc.arg(reihenfolge)
       OR (reihenfolge = sqlc.arg(reihenfolge) AND id > sqlc.arg(id)))
ORDER BY reihenfolge ASC, id ASC
LIMIT 1;

-- name: NormalisiereProduktReihenfolge :exec
-- Vergibt die Reihenfolge aller Produkte einer Kategorie dicht neu (1..N) in
-- der bestehenden Sortierung. Das läuft vor jedem Tausch, weil zwei Zeilen mit
-- demselben Wert sonst denselben Wert zurückgeschrieben bekämen und das
-- Verschieben wirkungslos bliebe. updated_at bleibt unberührt: die
-- Normalisierung verändert die sichtbare Reihenfolge nicht.
UPDATE produkte p
SET reihenfolge = neu.rang
FROM (
    SELECT pk.id, (row_number() OVER (ORDER BY pk.reihenfolge, pk.id))::int AS rang
    FROM produkte pk
    WHERE pk.kategorie = sqlc.arg(kategorie) AND pk.status != 'deleted'
) neu
WHERE p.id = neu.id;

-- name: SetProduktReihenfolge :execresult
UPDATE produkte SET reihenfolge = sqlc.arg(reihenfolge), updated_at = sqlc.arg(updated_at) WHERE id = sqlc.arg(id);

-- name: GetVariante :one
SELECT id, name, preis_cents, status, created_at, updated_at
FROM produkt_varianten WHERE id = $1 AND status != 'deleted';

-- name: CreateVariante :one
-- Neue Varianten landen ans Ende ihres Produkts (siehe CreateProdukt).
INSERT INTO produkt_varianten (produkt_id, name, preis_cents, status, created_at, updated_at, reihenfolge)
VALUES ($1, $2, $3, $4, $5, $6,
    COALESCE((SELECT MAX(reihenfolge) + 1 FROM produkt_varianten WHERE produkt_id = $1), 1))
RETURNING id;

-- name: UpdateVariante :execresult
UPDATE produkt_varianten SET name = $1, preis_cents = $2, status = $3, updated_at = $4 WHERE id = $5;

-- name: GetVarianteReihenfolge :one
SELECT id, produkt_id, reihenfolge FROM produkt_varianten WHERE id = $1 AND status != 'deleted';

-- name: GetVarianteVorgaenger :one
-- Die in der Sortierung (reihenfolge, id) unmittelbar davor liegende Variante
-- desselben Produkts. Kein Treffer bedeutet: die Variante steht bereits oben.
SELECT id, reihenfolge FROM produkt_varianten
WHERE produkt_id = sqlc.arg(produkt_id)
  AND status != 'deleted'
  AND (reihenfolge < sqlc.arg(reihenfolge)
       OR (reihenfolge = sqlc.arg(reihenfolge) AND id < sqlc.arg(id)))
ORDER BY reihenfolge DESC, id DESC
LIMIT 1;

-- name: GetVarianteNachfolger :one
SELECT id, reihenfolge FROM produkt_varianten
WHERE produkt_id = sqlc.arg(produkt_id)
  AND status != 'deleted'
  AND (reihenfolge > sqlc.arg(reihenfolge)
       OR (reihenfolge = sqlc.arg(reihenfolge) AND id > sqlc.arg(id)))
ORDER BY reihenfolge ASC, id ASC
LIMIT 1;

-- name: NormalisiereVarianteReihenfolge :exec
-- Dichte Neunummerierung der Varianten eines Produkts, siehe
-- NormalisiereProduktReihenfolge.
UPDATE produkt_varianten v
SET reihenfolge = neu.rang
FROM (
    SELECT pv.id, (row_number() OVER (ORDER BY pv.reihenfolge, pv.id))::int AS rang
    FROM produkt_varianten pv
    WHERE pv.produkt_id = sqlc.arg(produkt_id) AND pv.status != 'deleted'
) neu
WHERE v.id = neu.id;

-- name: SetVarianteReihenfolge :execresult
UPDATE produkt_varianten SET reihenfolge = sqlc.arg(reihenfolge), updated_at = sqlc.arg(updated_at) WHERE id = sqlc.arg(id);

-- name: SortiereVariantenAlphabetisch :exec
-- Vergibt die Reihenfolge der Varianten eines Produkts neu, alphabetisch nach
-- Namen. COLLATE "de-DE-x-icu" hält die Sortierung deterministisch deutsch,
-- unabhängig von der Locale des Clusters: Umlaute und Akzente reihen sich bei
-- ihrem Grundbuchstaben ein, statt (etwa unter der C-Locale) hinter allen
-- anderen Buchstaben zu landen. Gelöschte Varianten bleiben unberührt; ihre
-- alten Werte stören nicht, weil sie überall herausgefiltert werden.
UPDATE produkt_varianten v
SET reihenfolge = neu.rang, updated_at = sqlc.arg(updated_at)
FROM (
    SELECT id, (row_number() OVER (ORDER BY name COLLATE "de-DE-x-icu", id))::int AS rang
    FROM produkt_varianten pv
    WHERE pv.produkt_id = sqlc.arg(produkt_id) AND pv.status != 'deleted'
) neu
WHERE v.id = neu.id;
