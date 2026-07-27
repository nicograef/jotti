-- Client-gelieferte Idempotenz-Schlüssel der buchenden Vorgänge, die ihren
-- Schlüssel nicht bereits über einen partiellen Unique-Index auf dem Event-JSON
-- tragen: Zahlung, Stornierung, Umbuchung, Direktverkauf-Stornierung. Ein
-- Vorgang umfasst je nach Art ein Event (Zahlung, Direktverkauf-Stornierung),
-- zwei Events (Umbuchung) oder n Events (Stornierung) — deshalb eine eigene
-- Tabelle statt eines Event-Feldes; die Event-JSON-Contracts bleiben unberührt.
--
-- Die Zeile wird in der Transaktion des Vorgangs VOR den Event-Inserts
-- geschrieben: Ein Primärschlüssel-Konflikt ist damit eindeutig eine
-- Duplikat-Einreichung (stille Erfolgsantwort ohne zweite Buchung), ein
-- UNIQUE(subject, version)-Konflikt bleibt eindeutig ein echter OCC-Konflikt.
BEGIN;

CREATE TABLE vorgang_idempotenz (
    vorgang_id UUID PRIMARY KEY,
    art TEXT NOT NULL CHECK (art IN ('zahlung', 'stornierung', 'umbuchung', 'direktverkauf-stornierung')),
    user_id INT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL
);

COMMENT ON TABLE vorgang_idempotenz IS 'Client-gelieferte Idempotenz-Schlüssel buchender Vorgänge; Zeile entsteht im selben Commit wie die Events des Vorgangs, vor deren Insert.';
COMMENT ON COLUMN vorgang_idempotenz.art IS 'Vorgangstyp für die Nachvollziehbarkeit im Support.';

COMMIT;
