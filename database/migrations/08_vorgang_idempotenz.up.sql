-- Client-gelieferte Idempotenz-Schlüssel ALLER buchenden Vorgänge: Bestellung,
-- Zahlung, Stornierung, Umbuchung, Direktverkauf, Direktverkauf-Stornierung und
-- Geldtransit. Ein Vorgang umfasst je nach Art ein Event (Zahlung,
-- Direktverkauf, Geldtransit, Direktverkauf-Stornierung), zwei Events
-- (Umbuchung) oder n Events (Stornierung) — deshalb eine eigene Tabelle statt
-- eines Event-Feldes; die Event-JSON-Contracts bleiben unberührt.
--
-- Die Zeile wird in der Transaktion des Vorgangs VOR den Event-Inserts
-- geschrieben. Damit bleibt ein UNIQUE(subject, version)-Konflikt im
-- kassenjournal eindeutig ein echter OCC-Konflikt, und die partiellen
-- Unique-Indexe auf dem Event-JSON (idx_kassenjournal_bestellung_id,
-- idx_kassenjournal_verkauf_id, idx_kassenjournal_geldtransit_id in
-- 01_initial.up.sql) bleiben nur noch als zweite Absicherung stehen: Für jeden
-- ab dieser Migration geschriebenen Vorgang können sie nicht mehr auslösen,
-- weil die Idempotenz-Zeile vor den Events entsteht.
--
-- Für Bestandsdaten gilt das nicht. Zu Events, die vor dieser Migration
-- geschrieben wurden, gibt es keine Idempotenz-Zeile; diese Migration legt
-- keine nach. Ein Wiederholversuch, der genau das Upgrade überspannt — erste
-- Einreichung davor gebucht, Wiederholung danach —, findet deshalb keinen
-- Schlüssel, bucht regulär und läuft in den alten Index. Er endet in HTTP 409
-- statt in der stillen Erfolgsantwort. Doppelt gebucht wird dabei nichts; das
-- Fenster ist auf Wiederholversuche über das Upgrade hinweg beschränkt.
--
-- payload_hash bindet den Schlüssel an die Nutzdaten des Vorgangs. Ohne diese
-- Bindung entscheidet allein der Client, was „derselbe Vorgang" ist, und beide
-- denkbaren Client-Strategien sind falsch: Ein bei Nutzdaten-Änderung
-- rotierender Schlüssel bucht doppelt, ein stabiler Schlüssel ohne Serverprüfung
-- verschluckt eine geänderte Einreichung. Aus Schlüssel und Hash folgen drei
-- Ausgänge:
--   * kein Schlüssel gespeichert          → neuer Vorgang, regulär buchen
--   * gleicher Schlüssel, gleicher Hash   → Duplikat, stille Erfolgsantwort
--   * gleicher Schlüssel, anderer Hash    → abweichende Nutzdaten, HTTP 409
BEGIN;

CREATE TABLE vorgang_idempotenz (
    vorgang_id UUID PRIMARY KEY,
    art TEXT NOT NULL CHECK (art IN ('bestellung', 'zahlung', 'stornierung', 'umbuchung', 'direktverkauf', 'direktverkauf-stornierung', 'geldtransit')),
    user_id INT NOT NULL REFERENCES users(id),
    payload_hash BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

COMMENT ON TABLE vorgang_idempotenz IS 'Client-gelieferte Idempotenz-Schlüssel buchender Vorgänge, an die Nutzdaten des Vorgangs gebunden; Zeile entsteht im selben Commit wie die Events des Vorgangs, vor deren Insert.';
COMMENT ON COLUMN vorgang_idempotenz.art IS 'Vorgangstyp für die Nachvollziehbarkeit im Support.';
COMMENT ON COLUMN vorgang_idempotenz.payload_hash IS 'SHA-256 über die Nutzdaten des Vorgangs. Gleicher Schlüssel mit gleichem Hash ist eine Duplikat-Einreichung (stille Erfolgsantwort), gleicher Schlüssel mit anderem Hash sind abweichende Nutzdaten (HTTP 409).';

COMMIT;
