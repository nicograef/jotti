// Fehlermeldungen der fiskaly-Einrichtung, geteilt von den Wizard-Schritten.
export const ZUGANGSDATEN_FEHLER = {
  tse_setup_zugangsdaten_ungueltig:
    'API-Key oder API-Secret ist ungültig. Bitte Zugangsdaten prüfen.',
}
export const SETUP_FEHLER = {
  ...ZUGANGSDATEN_FEHLER,
  tse_setup_umgebung_abweichung:
    'Die Umgebung der Zugangsdaten hat sich geändert. Bitte das Konto erneut prüfen.',
  tse_konfiguration_kassensitzung_offen:
    'Die TSE kann nicht geändert werden, solange eine Kassensitzung aktiv ist (offen oder Abschluss unterbrochen). Bitte zuerst den Kassenabschluss durchführen bzw. wiederholen.',
}
