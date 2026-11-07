-- Pfand-Verknüpfung: Eine Variante zeigt auf die Pfand-Variante, die der
-- Warenkorb je Stück dazulegt. Die Feldliste der View produkt_varianten_json
-- muss mit jsonVariante in backend/repository/produkt_repo/types.go
-- übereinstimmen.
--
-- Additiv: die Spalte ist nullable, bestehende Varianten bleiben ohne Pfand.
BEGIN;

ALTER TABLE produkt_varianten
    ADD COLUMN pfand_variante_id INT REFERENCES produkt_varianten(id);

COMMENT ON COLUMN produkt_varianten.pfand_variante_id IS
    'Pfand-Variante, die der Warenkorb je Stück dieser Variante dazulegt; NULL = kein Pfand. Gleicher Steuersatz, Ziel ist eine Variante eines Pfand-Produkts (geprüft im Backend).';

CREATE OR REPLACE VIEW produkt_varianten_json AS
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
        'pfandVarianteId', pfand_variante_id,
        'createdAt', created_at,
        'updatedAt', updated_at
    ) AS eintrag
FROM produkt_varianten;

COMMIT;
