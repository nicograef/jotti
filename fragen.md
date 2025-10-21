# Offene Fragen zu Jotti

## Zielbild & Scope

- Was ist der MVP‑Scope? Nur Bestellen + Bezahlen, oder auch Lager/Inventar, Druck, Reports?
- Primär-Umgebung: stationäres PoS (Touchscreen), mobiles Gerät (Smartphone/Tablet), oder beides?
- Offline‑First notwendig (z. B. Zeltfest mit schwachem Internet) und wie soll Sync/Recovery funktionieren?
- Ein System pro Veranstaltung (Single‑Tenant je Event) oder multi‑tenant mit mehreren parallelen Veranstaltungen?

## Benutzer & Rollen

- Welche Rollen gibt es final (Service, Kasse/PoS, Admin, Bar/Küche, Supervisor)? Brauchen wir eine Rechte‑Matrix?
- Anmeldung: PIN, QR/Badge, Passwort, SSO? Session‑Timeouts und Sperren?
- Brauchen wir Schichtwechsel/Übergabe (Schichtstart/-ende, Kassensturz)?

## Produkte

- Attribute außer Name/Preis: Kategorie, Varianten (Größen), Modifikatoren/Optionen (z. B. „mit Soße“), Allergene, Tags?
- Steuern/Mehrwertsteuer: mehrere Sätze pro Produkt? Brutto/Netto-Preise, Rundung?
- Dynamische Preise (Happy Hour, Event-spezifisch), Rabatte/Aktionen, Gratisartikel?
- Ausverkauft: manuell gesetzt vs. Bestandsführung? Möchtet ihr echtes Inventar (Mengenabzug) oder nur „verfügbar/ausverkauft“?
- Darstellung/Gruppierung in der UI (Kacheln, Favoriten, Suchbar)?

## Kunden (z. B. Tische)

- Kundentypen: Tische, Stehtische, „Sammelkunde/Kasse“, externe Kunden?
- Workflows: Tische zusammenlegen/teilen, Positionen zwischen Kunden verschieben, umbenennen, schließen/wieder öffnen?
- Max. offene Kunden, Namenskonventionen („Tisch 1“, „Sonderposten“) – frei wählbar oder vordefinierte Karte?

## Bestellungen

- Zustände/Workflow: offen → in Zubereitung → serviert? Braucht’s Produktionswege (Bar/Küche) und Routing?
- Bearbeitung: nachträgliche Änderungen/Storno einzelner Positionen, Gründe/Belege?
- Notizen pro Position (z. B. „ohne Zwiebeln“), Pflichtauswahl bei bestimmten Produkten (Optionen)?
- Belegdruck: Küchen-/Bontickets, Re-Druck, Drucker pro Bereich?

## Bezahlungen

- Zahlungsarten: Bar, Karte (Terminal-Integration), Gutschein, Rechnung, Mischzahlungen, Trinkgeld?
- Teilzahlungen/Split bill: pro Produkt, pro Anteil, pro Person? Restbetrag-Handling, Rundung, Wechselgeld?
- Rabatte: pro Position vs. Gesamtrabatt, prozentual vs. absolut, Berechtigungen/Audit?
- Storno/Erstattung/Chargeback: wie dokumentieren, wer darf?
- Abschlussprozesse: Kunde „schließen“, Tagesabschluss/Z‑Bericht, Kassenbuch?

## Events & Datenmodell

- Event‑Quelle/Ziel: Nur internes Event‑Sourcing oder auch externes Messaging/Webhooks?
- Versionierung der Event‑Typen (v1, v2), Abwärtskompatibilität, Schemaevolution?
- Idempotenz: Müssen wir Wiederholungen erkennen (idempotency keys)?
- Reihenfolge/Partitionierung: Garantierte Ordering pro Kunde oder global?
- Retention/Audit: Aufbewahrungsfristen, DSGVO (PII im Subject/Payload?), Löschkonzepte/Anonymisierung.
- Naming-Konventionen: subject „customer:tisch17“ – ist das stabiler Schlüssel? Gibt es separate Customer-IDs?

## Reporting & Controlling

- Berichte: Umsatz je Produkt/Kategorie/Zeitraum/Steuersatz, offene Posten, Topseller, Stornos, Rabatte, Trinkgeld?
- Exportformate: CSV/Excel/DATEV, API, täglicher Abschlussreport?
- Realtime-Dashboards vs. Batch-Reports?

## Administration

- Produktpflege: Entwürfe, Bulk‑Import/‑Export, Preislisten je Event, Archivierung?
- Benutzerverwaltung: Rollen, Sperrungen, Berechtigungen, Audit-Logs?
- Konfiguration je Veranstaltung: Menüs, Zeitfenster, Standardsteuersatz, Druckerzuordnung.

## Hardware & Integrationen

- Drucker (Netzwerk, USB, Bluetooth), Bon‑Layout, Logo, QR‑Codes?
- Kartenterminals (z. B. ZVT, OPI, Adyen, Stripe Terminal, SumUp) – gewünschte Anbieter?
- Scanner (Barcodes/QR) für Produkte/Kunden?
- Kassenlade, Kundendisplay?

## Recht & Compliance (DE)

- KassenSichV/TSE‑Pflicht? GoBD‑Konformität, DSFinV‑K‑Export, Belegausgabepflicht (eBon/Papier)?
- Altersfreigaben (z. B. Alkohol), Jugendschutzprüfungen?
- Datenschutz: Rollenbasierter Zugriff, Protokollierung, Datenminimierung, AVV?

## Internationalisierung

- Sprachen (DE/EN), Währungen, Datums-/Zeitformate, Zeitzonen pro Event?
- Mehrere Steuersysteme/Länder geplant?

## Nicht-funktionale Anforderungen

- Nutzer- und Produktanzahlen (Skalierung): max. gleichzeitige Benutzer, Positionsbuchungen/min?
- Latenzanforderungen am PoS (<100 ms für UI‑Aktionen?), Verfügbarkeit, Lastspitzen (Ansturm zur Pause)?
- Sicherheitsanforderungen: RBAC, Audit, Rate‑Limiting, Backups/Restore, Disaster Recovery.

## UX/Flows

- Wichtigste Flows: „Tisch öffnen → Produkte buchen → Bon/Produktion → Bezahlen → Tisch schließen“.
- Tastatur-/Touch‑Optimierung, Schnellaktionen, Favoriten, Fehler-Undo.
- Barrierefreiheit: Kontraste, Schriftgrößen, Screenreader?

## Roadmap & Betrieb

- Hosting: Cloud vs. On‑Premise auf Kassen-Hardware; Mandantenfähigkeit?
- Environments: Demo/Training-Modus, Staging, Daten-Seeding.
- Lizenzierung/Preismodell, falls Produkt.

## Offene Event‑Typen (Erweiterungen)

- Kunde erstellt/geschlossen/zusammengelegt/geteilt.
- Bestellung geändert/storniert/serviert.
- Produkt ausverkauft/geändert, Modifikator gewählt.
- Bezahlung storniert/erstattet, Trinkgeld erfasst.
