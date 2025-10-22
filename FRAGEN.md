# Offene Fragen zu Jotti

Im Folgenden sind nur noch die offenen und für das MVP relevanten Fragen gelistet, basierend auf der aktualisierten Projektbeschreibung in `README.md`.

## Produkte & Preise

- Kategorien: Sind die Kategorien fix auf „Getränk“ und „Essen“ begrenzt oder erweiterbar/konfigurierbar?
- Mehrwertsteuer: Je Produkt konfigurierbar (7% oder 19%)? Brutto-/Netto-Preisführung und Rundungsregeln (Kassenrundung, Bankers Rounding)?
- Preissnapshot: Wird Preis inkl. MwSt. und Steuersatz beim Bestell-/Bezahlzeitpunkt im Event mitgespeichert (für unveränderliche Auswertungen)?
- Dynamische Preise/Rabatte: Gehört das im MVP ausdrücklich nicht dazu? Falls doch, minimaler Umfang (z. B. prozentualer Gesamtrabatt je Tisch)?

## Ausverkauf & Bestand

- Ausverkauft-Status: rein manuell per Toggle im Admin-Bereich oder gibt es eine einfache Bestandszahl (ohne komplettes Inventar) im MVP?
- Verhalten bei Ausverkauf während Aufnahme: Wie soll die UI reagieren, wenn ein Produkt zwischen Auswahl und Bestätigung ausverkauft wird?

## Tische & Workflows

- Tisch-Namen: Namensregeln, Eindeutigkeit und maximale Anzahl von Tischen? Archivieren/Deaktivieren statt Löschen?
- Zusammenlegen/Teilen von Tischen und Positionsverschiebung: bewusst außerhalb des MVP, korrekt?

## Bestellungen

- Positionsnotizen/Optionen: Brauchen wir im MVP Notizen (z. B. „ohne Zwiebeln“) oder Pflichtauswahl-Optionen – oder erst später?
- Stornierung: Gilt die Idee „Storno innerhalb 1 Minute“ als MVP-Regel? Wie protokollieren/auditen wir Stornos (Event-Typ, Grund, wer/ wann)?

## Bezahlungen

- Teilzahlungen: Die Bezahl-API erlaubt Auswahl einzelner offener Positionen. Gibt es weitere Regeln (z. B. Mindestposition, Sperren parallel geöffneter Bezahlvorgänge)?
- Idempotenz bei Doppelregistrierungen: Wie erkennen/verhindern wir versehentliche doppelte Bezahl-Events (z. B. Retry, Doppelklick)?

## Events & Datenmodell

- Event-Schema: Finale Payload-Felder für `bestellung-aufgegeben:v1` und `bezahlung-registriert:v1` (Produkt-ID, Menge, Preis/MwSt-Snapshot, Benutzer-ID)?
- Ordering: Garantierte Reihenfolge pro Tisch (Partitionierung nach `tisch:<id>`), und wie wird sie in Postgres/Log sichergestellt?
- Idempotenzschlüssel: Sollen Commands einen `idempotencyKey` tragen, der im Event-Log geprüft wird?
- Versionierung/Evolution: Wie gehen wir mit späteren Schemaänderungen um (z. B. `v2` Events, Migrationsstrategie)?

## Reporting & Export

- Tagesabschluss: Exakte Definition (Zeitzone, Zeitraumgrenzen), Rundung, und welche Kennzahlen müssen enthalten sein?
- CSV-Export: Zielformat (Trennzeichen, Dezimaltrennzeichen, Encoding), Feldliste, und ob ein Import ins Buchhaltungstool vorgesehen ist.

## Auth & Sessions

- JWT (12h): Gibt es Token-Refresh oder Soft-Logout? Wie invalidieren wir Tokens bei Passwortwechsel/Logout?
- Passwort-Policy: Mindestlänge, Komplexitätsregeln, Lockout-Strategie bei Fehlversuchen.
- Admin-Bootstrap: Wie wird der erste Admin angelegt (Seed/CLI/Setup-Flow)?

## Nicht-funktionale & Betrieb

- Zielgröße: Erwartete Anzahl gleichzeitiger Nutzer, Produkte, Buchungen pro Minute (für Dimensionierung und Tests).
- Backups/Restore: Frequenz, Aufbewahrungsdauer und Wiederherstellungstest (On-Premise Leitplanken).
- Deploy/Update: Wie erfolgt das On-Prem-Update (Docker-Compose Pull/Restart, Migrationsfenster)?

## Datenschutz & Audit

- Retention: Aufbewahrungsfristen für Events/Logs; Anonymisierung/Löschung von Benutzerdaten (DSGVO).
- Audit-Zugriff: Wer kann Audit-Logs einsehen, und brauchen wir unveränderliche Exporte?

## Frontend & UX

- Mobile-UI: Pflichtfeatures im MVP (Kachel-Layout, Suche, Favoriten, schneller Mengen-Modifier)?
- Barrierefreiheit: Mindestanforderungen (Kontraste, Schriftgrößen) für den MVP.
