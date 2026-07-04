# PRD: TSE-Admin auf Monitoring vereinfachen

Mini-PRD, entstanden aus dem Audit des Outbox-Branches (2026-07-04, siehe
[Audit-Bericht](../plans/audit-tse-signatur-outbox-bericht.md) und
[Entscheidungs-Doku](../plans/audit-tse-signatur-outbox-entscheidungen.md)). Revidiert die User Stories 6, 9,
12 und den zweiten Teil von 13 aus [prd-tse-signatur-outbox.md](prd-tse-signatur-outbox.md).

## Problem Statement

Die Signaturauftrags-Verwaltung (Einzelauftrags-Liste, Zurücksetzen einzeln/gesamt, Verwerfen mit Begründung)
bietet der Zielgruppe keinen echten Handlungswert. Durch die Fehlertaxonomie heilen TSE-Ausfälle vollautomatisch;
`fehlgeschlagen` entsteht praktisch nur durch einen jotti-Bug, den kein Vereinshelfer per Klick beheben kann.
Verwerfen ist faktisch nur eine Quittierfunktion, und Zurücksetzen nützt nur Entwicklern nach einem Bugfix.
Gleichzeitig erzeugt die Verwaltung reale Kosten: ein Race (Signatur bei fiskaly, Auftrag verworfen), stille
No-Op-Antworten, ein Listen-Limit ohne Hinweis, drei Endpunkte, Formular, Guards und Tests (Audit-Befunde
B1, B2, B3). YAGNI: weniger Funktion ist hier die korrektere Funktion.

## Solution

Die TSE-Admin-Oberfläche wird auf Monitoring und Status reduziert. Es gibt keine mutierenden Admin-Aktionen auf
Signaturaufträgen mehr; der Status `verworfen` entfällt vollständig. Die Diagnose ist minimalistisch: eine grobe
Information, dass etwas nicht stimmt, plus der letzte Fehlertext. Fehlgeschlagene Aufträge warnen nur bis zum
nächsten Kassenabschluss; der Abschluss weist sie aus und quittiert damit. Reparatur nach einem Bugfix ist ein
dokumentierter Entwicklerpfad (SQL-Runbook), keine UI.

## User Stories

1. Als Vereins-Admin möchte ich im Dashboard und in der Finanzamt-Ansicht sehen, ob die TSE-Signierung gesund
   ist (offene Aufträge, Rückstand, Durchsatz), und bei Problemen eine grobe Warnung mit dem letzten Fehlertext,
   damit ich weiß, wann ich Hilfe holen muss, ohne selbst eingreifen zu können oder zu müssen.
2. Als Vereins-Admin möchte ich, dass eine Warnung über endgültig fehlgeschlagene Signaturen mit dem
   Kassenabschluss endet (der Abschluss weist die Reste aus), damit ein alter Vorfall nicht dauerhaft warnt.
3. Als Vereins-Admin akzeptiere ich, dass Vorgänge aus einem Zeitraum ohne TSE-Konfiguration endgültig
   unsigniert bleiben: keine TSE konfiguriert heißt keine TSE für diesen Zeitraum.
4. Als Entwickler möchte ich fehlgeschlagene Aufträge nach einem Bugfix über ein dokumentiertes SQL wieder
   einreihen können, damit die Reparatur möglich bleibt, ohne eine Admin-UI zu pflegen.

Revisionen gegenüber prd-tse-signatur-outbox.md: US 6 (Liste aller Aufträge) wird zur Minimal-Diagnose; US 9
(Zurücksetzen/Verwerfen) entfällt; US 12 (Dashboard-Warnung) bleibt, die fehlgeschlagen-Warnung wird
sitzungsbezogen; US 13 verliert den zweiten Halbsatz (bewusstes Wiedereinreihen nach später Einrichtung).

## Implementation Decisions

- Entfernt werden: Einzelauftrags-Liste, Zurücksetzen einzeln/gesamt, Verwerfen; die Endpunkte
  `get-tse-signaturauftraege`, `tse-signaturauftrag-zuruecksetzen`, `tse-signaturauftraege-zuruecksetzen`,
  `tse-signaturauftrag-verwerfen`; der Status `verworfen` samt Spalten `verworfen_grund/von/am`.
- Bleiben: Queue-Zustand (`get-tse-signatur-queue`), Störungsprotokoll-Ansicht (`get-tse-stoerungen`),
  Dashboard-Warnungen, TSE-Status; Gate, Worker und Störungsprotokoll unverändert.
- Diagnose minimal: Queue-Zustand zählt `fehlgeschlagen` nur für die aktive Kassensitzung und liefert den
  letzten Fehlertext; keine Einzelauftrags-Ansicht.
- Ausfallbegriff: Endstatus sind nur noch `fehlgeschlagen` und `tse_nicht_konfiguriert`.
- Reparatur nach Bugfix: dokumentiertes SQL im Handbuch (Runbook), bewusst keine UI.
- Umsetzung auf dem Branch `feat/tse-signatur-outbox-phase1` (Pre-Release, Breaking erwünscht).

## Out of Scope

- Neue Admin-Funktionen jeder Art (Paginierung, Filter, Quittier-Buttons).
- Änderungen an Worker, Statusmaschine (bis auf den wegfallenden Status), Gate-Logik, Beleg- und Export-Pfad.
- Druckauftrags-Verwaltung (deren Status `verworfen` ist ein eigener, bleibender Begriff).
