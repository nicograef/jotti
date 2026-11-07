---
title: Checkliste
description: 'Alle einmaligen und laufenden Pflichten für den jotti-Betrieb auf einen Blick, von der TSE-Einrichtung bis zur Datenarchivierung.'
---

**Einmalig, vor dem ersten Einsatz:**

- [ ] Nutzungsvereinbarung mit dem Autor abgeschlossen (siehe
      [Lizenzmodell](../lizenzmodell.md))
- [ ] fiskaly-Konto registriert, API-Key und API-Secret erstellt und sicher notiert
      (siehe [TSE einrichten](tse-einrichten.md))
- [ ] TSE in TEST auf einer Übungsinstallation (anderer Rechner) komplett
      durchgespielt; der Kassenrechner enthält keine Übungsverkäufe
- [ ] LIVE-TSS (die TSE im fiskaly-Konto) über den Assistenten eingerichtet
- [ ] Admin-PUK und Admin-PIN sicher und dauerhaft verwahrt
- [ ] Abschluss-Verbindungstest steht auf „Verbindung bestätigt"
- [ ] Betreiber-Stammdaten (Vereinsname, Adresse, Steuernummer) im Admin-Bereich
      gepflegt
- [ ] Produkte mit korrekten Steuersätzen angelegt
- [ ] Tische angelegt
- [ ] Benutzerkonten für die Helfer angelegt (Rollen „Service" und „Serviceleitung")
- [ ] Kasse innerhalb eines Monats nach Inbetriebnahme über ELSTER beim Finanzamt
      angemeldet (§ 146a Abs. 4 AO; siehe
      [Kasse beim Finanzamt anmelden](finanzamt-anmelden.md))
- [ ] Ggf. Antrag auf Befreiung von der Belegausgabe gestellt
- [ ] Backup- und Archivierungsroutine festgelegt (siehe
      [Datenaufbewahrung](datenaufbewahrung.md))
- [ ] Verfahrensdokumentation an den Verein angepasst (siehe
      [Muster-Verfahrensdokumentation](../verfahrensdokumentation.md))

**Laufend / regelmäßig (beide Betriebswege):**

- [ ] Nach jedem Veranstaltungstag: Tagesabschluss (Z-Bon) erstellt (siehe
      [Der Veranstaltungstag](veranstaltungstag.md))
- [ ] Nach jedem Veranstaltungstag: DSFinV-K-Export erzeugt, geprüft und an zwei
      getrennten Orten abgelegt
- [ ] Daten werden 10 Jahre archiviert
- [ ] Bei Stilllegung der Kasse: Abmeldung beim Finanzamt innerhalb eines Monats
- [ ] Vor Stilllegung der TSS oder Ende des fiskaly-Vertrags: TSE-Export (TAR-Datei)
      erstellt und archiviert (siehe [Datenaufbewahrung](datenaufbewahrung.md))

**Zusätzlich beim Server-Betrieb (Experten-Weg):**

- [ ] Tägliche Datenbank-Backups laufen (siehe
      [Backups](aktualisieren-backups.md#backups)) und werden regelmäßig auf einen
      anderen Rechner kopiert
