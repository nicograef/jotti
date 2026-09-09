# Changelog

Dieses Changelog fasst die für Anwenderinnen und Anwender wichtigen Änderungen an jotti
zusammen: von Hand gepflegt, auf Deutsch und bewusst verständlich gehalten. Ab Version
1.0.0 wird es manuell fortgeschrieben.

Das Format orientiert sich an [Keep a Changelog](https://keepachangelog.com/de/1.1.0/),
die Versionierung an [Semantic Versioning](https://semver.org/lang/de/). Unabhängig davon
werden die technischen Release-Notes je Version automatisch aus den Commits erzeugt; sie
erscheinen bei den GitHub-Releases.

## [1.0.0]

Erste stabile Version von jotti, dem kostenlosen Kassensystem für Vereinsfeste.

### Kassenbetrieb

- Bestellungen pro Tisch aufnehmen, mit Produkten, Varianten, Steuersätzen und Kommentaren.
- Zahlungen kassieren, inklusive Teilzahlungen und Rückgeldberechnung.
- Bestellungen stornieren mit Pflichtkommentar; vorbehalten für Admin und Serviceleitung.
- Bestellungen auf einen anderen Tisch umbuchen.
- Tisch-Übersicht mit offenem Saldo, Positionen und Bestellhistorie.
- „Meine Tische": Favoriten als große Tischkarten auf dem Dashboard, Schnellsuche nach Name oder Nummer.
- Direktverkauf ohne Tisch: bestellen und kassieren in einem Schritt, mit Historie und Storno.
- Schlägt ein Laden fehl, erscheint ein Fehlerhinweis mit „Erneut versuchen" statt leerer Nullwerte.

### Küche

- Automatischer Bon-Druck von Bestell- und Küchenbons an zugeordnete Bondrucker, pro Warenkategorie konfigurierbar.
- Abholbons beim Direktverkauf in drei Modi: ein Bon pro Position, ein Sammelbon pro Bestellung oder ein Bon je Stück.

### Kassenführung

- Fortlaufend nummerierte Kassensitzungen eröffnen und schließen.
- Anfangsbestand (Wechselgeld) zu Veranstaltungsbeginn erfassen.
- Soll-Kassenbestand jederzeit einsehen, aufgeschlüsselt nach Komponenten.
- Einlagen und Entnahmen (Geldtransit) buchen.
- Kassensturz: Ist-Bestand eingeben, Differenz berechnen und die Abweichung automatisch verbuchen.
- Die Kassensturz-Differenz erscheint überall als Ist minus Soll.
- Tagesabschluss (Z-Bon) mit fortlaufender, nie zurücksetzbarer Nummer.
- Bleibt ein Kassenabschluss unterwegs stehen, zeigt die Kassentag-Seite „Abschluss unterbrochen" und bietet den erneuten Abschluss an.
- Wurde seit dem Kassensturz Geld bewegt, bricht der erneute Abschluss ab und verweist an den Administrator.

### Abrechnung und Reporting

- Tagesabrechnung über alle Umsätze, Zahlungen und offenen Beträge, nach Steuersatz aufgeschlüsselt.
- Abrechnung je Tisch und je Servicekraft.
- Produktumsatz-Reporting: meistverkaufte Varianten, Mengen und Einnahmen je Produkt.
- DSFinV-K-Export als ZIP-Archiv (Version 2.4) für die Finanzverwaltung.

### Verwaltung, Sicherheit und Compliance

- Admin-Bereich für Produkte (mit Varianten und Steuersätzen), Tische, Benutzer und Betreiber-Stammdaten.
- Anzeigereihenfolge der Produkte und Varianten im Admin-Bereich festlegen; die Bestellliste im Service folgt ihr.
- Rollenmodell mit den Rollen Admin, Serviceleitung und Service.
- Sicheres Onboarding per Einmalpasswort, Passwort-Hashing mit Argon2id und Anmeldung über JWT.
- Die Anmeldung nimmt ein Passwort mit umgebenden Leerzeichen genauso an wie das Setzen des Passworts.
- Event-Sourcing für eine lückenlose, unveränderliche Bestellhistorie (GoBD-konform durch ein Append-only-Kassenjournal).
- Anbindung einer BSI-zertifizierten Cloud-TSE von fiskaly, die jeden Vorgang signiert.
- Die TSE-Einrichtung bleibt gesperrt, solange eine Kassensitzung aktiv ist.
- Fehlgeschlagene TSE-Signaturen meldet die Finanzamt-Seite vor dem Signatur-Rückstand.
- Gesetzeskonforme Belegausgabe mit TSE-Signatur, QR-Code, Steuersatz und Betreiberadresse.
- Alle Zeitstempel auf Kassenbelegen und Bons in deutscher Ortszeit.
- Betreiber-Stammdaten in den amtlichen DSFinV-K-Feldlängen, nach Zeichen gezählt: Umlaute zählen einfach.
- Das ELSTER-Meldedatum folgt dem deutschen Kalendertag.

### Betrieb

- Windows: `jotti-restore.cmd` und `jotti-repair.cmd` enden nach der Datenbankarbeit; gestartet wird jotti danach mit `jotti-start.exe`.
- Server-Stack unter eigener Domain: `JOTTI_DOMAIN` ist Pflicht, ohne die Variable bricht jeder Befehl des Stacks ab.
