# Audit TSE-Signatur-Outbox: Befunde mit Entscheidungsbedarf

Ergänzung zum [Audit-Bericht](audit-tse-signatur-outbox-bericht.md). Stand nach Besprechung am 2026-07-04:
B4 und B5 sind entschieden und umgesetzt; die Einzelbefunde B1, B2 und B3 sind durch eine Richtungsfrage
ersetzt, die der Entwickler aufgeworfen hat: die TSE-Admin-Funktionalität grundsätzlich zu verkleinern
(YAGNI, KISS) statt sie zu härten. Dazu unten die Evaluation mit Zielbild und offenen Fragen.

## Erledigt

- B4 AEAO zu § 146a Nr. 2.2.2: Einordnungs-Absatz in compliance.md §3.8 ergänzt (benennt die Normstelle,
  begründet die asynchrone Signierung, verweist auf die Herstellerdokumentation).
- B5 Seed: Aufgelöste Ausfallfenster werden als geschlossene `tse_fehler`-Störungszeiträume geseedet
  (`stoerungszeitraeumeAus` in `backend/seed/faketse.go`, Insert in `writer.go`); die
  Ausfalldokumentations-Ansicht der Demo passt damit zu den nachsignierten Belegen. Das offene Fenster der
  laufenden Sitzung materialisiert weiterhin live über Worker und Watchdog.

## Richtungsentscheidung: TSE-Admin auf Monitoring reduzieren

> Entschieden am 2026-07-04: Zielbild bestätigt (Diagnose minimal, Warnfenster bis zum Abschluss,
> `tse_nicht_konfiguriert` endgültig, Umsetzung auf diesem Branch). Quelle ist jetzt
> [prd-tse-admin-vereinfachung.md](../prds/prd-tse-admin-vereinfachung.md) mit
> [plan-tse-admin-vereinfachung.md](plan-tse-admin-vereinfachung.md); die Abschnitte unten bleiben als
> Herleitung stehen.

Impuls des Entwicklers zu B1/B2/B3: Statt Verwerfen-Race, No-Op-Antworten und Listen-Limit zu härten, die
Admin-UI für TSE kritisch prüfen; möglicherweise keine Einzelauftrags-Verwaltung, kein Zurücksetzen, kein
Verwerfen, nur Monitoring und Status.

### Evaluation: Wer braucht welche Funktion wirklich?

Durch die Fehlertaxonomie (Phase 3) gibt es genau drei Wege in einen Problemzustand:

1. `fehlgeschlagen`: nur auftragsspezifische Ablehnungen durch fiskaly (400/409/422). Praktisch ist das ein
   jotti-Bug (fehlerhafte processData) oder ein von fiskaly abgelehnter Einzelvorgang. Ein TSE-Ausfall führt
   nie zu `fehlgeschlagen`; die Aufträge bleiben offen und heilen automatisch.
2. `tse_nicht_konfiguriert`: Buchen ohne TSE-Konfiguration, endgültig markiert; der Tag wird beim Abschluss
   deutlich „ohne TSE" ausgewiesen.
3. Offen bei Störung: heilt vollautomatisch (Störungszustand, Half-Open, Watchdog). Keine Admin-Aktion nötig
   oder möglich.

Wert der heutigen Verwaltungsfunktionen für die Zielgruppe (nicht-technische Vereinshelfer):

- Zurücksetzen (einzeln/gesamt) hilft nur, wenn sich zwischen den Versuchen etwas geändert hat: ein
  Software-Fix (Gift-Auftrag) oder eine späte TSE-Einrichtung (`tse_nicht_konfiguriert`). Beides sind
  Entwickler- bzw. Sonderszenarien. Ein Helfer, der auf einen deterministischen Fehler „erneut versuchen"
  klickt, produziert nur drei neue Fehlversuche. Retroaktives Nachsignieren vor-konfigurationeller Vorgänge
  ist zudem fiskalisch fragwürdig (TSE-Zeiten liegen dann Tage nach dem Vorgang; der Tag wurde bereits ohne
  TSE ausgewiesen und abgeschlossen).
- Verwerfen mit Begründung ist faktisch eine Quittierfunktion: Der fiskalische Zustand bleibt Ausfall, es
  ändert sich nur, dass das Dashboard Ruhe gibt und ein Protokolleintrag entsteht. Dafür kostet es das
  Race (Audit-Befund B1), die No-Op-Frage (B2), Formular, Guards, Endpunkt und Tests.
- Die Einzelauftrags-Liste (B3) hat für Helfer ohne Handlungsoptionen keinen Wert. Ihr realer Nutzen ist
  Diagnose: den letzten Fehlertext ablesen und an Support/Entwickler melden. Dafür braucht es keine
  200er-Liste aller Status.

### Zielbild (Vorschlag)

1. Entfernen: Signaturauftrags-Verwaltung komplett — Einzelauftrags-Liste, Zurücksetzen einzeln/gesamt,
   Verwerfen; die drei Endpunkte, die UI-Sektion, der Status `verworfen` samt Spalten
   (`verworfen_grund/von/am`) und Guards. B1, B2 und B3 entfallen damit strukturell.
2. Behalten: Dashboard-Warnungen, Queue-Zustand (offene Aufträge, Rückstand, Signaturen/Minute, p95),
   Störungsprotokoll-Ansicht, TSE-Status.
3. Ergänzen (klein): Diagnose ohne Verwaltung — im Queue-Zustand zusätzlich die Anzahl fehlgeschlagener
   Aufträge seit dem letzten Abschluss und der letzte Fehlertext; alternativ eine read-only-Liste nur der
   fehlgeschlagenen Aufträge.
4. Warnlogik: fehlgeschlagene Aufträge warnen bis zum nächsten Kassenabschluss (der sie in der
   Abschlussmeldung ausweist und damit quittiert), danach Ruhe. Das ersetzt die Quittierfunktion des
   Verwerfens ohne eigenen Endpunkt.
5. Entwicklerpfad statt UI: Reparatur nach einem Bugfix per dokumentiertem SQL (Runbook im Handbuch),
   z. B. `UPDATE tse_signaturauftraege SET status = 'offen', versuche = 0 WHERE status = 'fehlgeschlagen'`.

Konsequenzen: Schema-CHECK ohne `verworfen` und drei Spalten weniger; Signaturstatus-Funktion, Gate-Ausweis
und Doku (language.md, handbuch.md, compliance.md, verfahrensdokumentation.md) verlieren einen Endstatus;
PRD-Userstories 6, 9 und 12 werden revidiert (7 Monitoring und 8 Ausfalldokumentation bleiben); die
Seed-Dramaturgie „genau ein verworfener Auftrag" entfällt.

### Offene Fragen vor der Umsetzung

1. Diagnose-Tiefe: reicht Anzahl plus letzter Fehlertext (aggregiert), oder read-only-Liste der
   fehlgeschlagenen Aufträge?
2. Warnfenster: fehlgeschlagen-Warnung endet mit dem Kassenabschluss — einverstanden? (Ohne Verwerfen würde
   ein Gift-Auftrag sonst dauerhaft warnen.)
3. `tse_nicht_konfiguriert` bleibt ohne Wiedereinreihungs-UI endgültig — einverstanden? (Eine späte
   Einrichtung signiert Altbestand dann nie nach; der Tag wurde ohne TSE ausgewiesen.)
4. Prozess: eigenes Mini-PRD und Folge-Branch nach dem Merge dieses Branches (Empfehlung: der Branch ist
   auditiert und plankonform, der Umbau revidiert PRD-Entscheidungen), oder noch auf diesem Branch?
