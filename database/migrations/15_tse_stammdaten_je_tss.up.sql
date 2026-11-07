-- TSE-Stammdaten je TSS: tse.csv muss zu jeder Signatur eines Exports den
-- Public Key und das Zertifikat der TSS liefern, die sie erzeugt hat. Die
-- Singleton-Zeile tse_stammdaten hält nur die aktuell eingerichtete TSS; ein
-- TSS-Wechsel überschrieb dort die Stammdaten der alten.
--
-- Additiv: eine neue Tabelle, befüllt aus der Singleton-Zeile, sofern dort
-- bereits eine TSS eingerichtet ist. Die Singleton-Zeile bleibt unverändert.
BEGIN;

CREATE TABLE tse_stammdaten_je_tss (
    seriennummer         TEXT PRIMARY KEY CHECK (seriennummer <> ''),
    signatur_algorithmus TEXT NOT NULL,
    public_key           TEXT NOT NULL,
    zertifikat           TEXT NOT NULL,
    log_time_format      TEXT NOT NULL,
    updated_at           TIMESTAMPTZ NOT NULL
);

-- Ohne die Stammdaten einer TSS lassen sich ihre Signaturen nicht mehr prüfen.
CREATE TRIGGER tse_stammdaten_je_tss_no_delete
    BEFORE DELETE ON tse_stammdaten_je_tss FOR EACH ROW
    EXECUTE FUNCTION prevent_table_mutation();

CREATE TRIGGER tse_stammdaten_je_tss_no_truncate
    BEFORE TRUNCATE ON tse_stammdaten_je_tss FOR EACH STATEMENT
    EXECUTE FUNCTION prevent_table_mutation();

COMMENT ON TABLE tse_stammdaten_je_tss IS 'TSE-Stammdaten jeder je eingerichteten TSS, Schlüssel ist die Seriennummer. Quelle der tse.csv im DSFinV-K-Export; die Signaturaufträge verweisen über tse_seriennummer darauf. Aufbewahrungspflichtig, kein DELETE.';
COMMENT ON COLUMN tse_stammdaten_je_tss.seriennummer IS 'Seriennummer der TSS (tss_serial_number aus fiskaly; SHA-256 des Public Key, hex-kodiert). DSFinV-K-Feld TSE_SERIAL.';
COMMENT ON COLUMN tse_stammdaten_je_tss.updated_at IS 'Letzter Abruf der Stammdaten bei fiskaly (UTC)';
COMMENT ON TABLE tse_stammdaten IS 'TSE-Stammdaten der aktuell eingerichteten TSS (Singleton); tse_stammdaten_je_tss hält die jeder TSS.';

INSERT INTO tse_stammdaten_je_tss (seriennummer, signatur_algorithmus, public_key, zertifikat, log_time_format, updated_at)
SELECT seriennummer, signatur_algorithmus, public_key, zertifikat, log_time_format, updated_at
FROM tse_stammdaten
WHERE id = 1 AND seriennummer <> '';

COMMIT;
