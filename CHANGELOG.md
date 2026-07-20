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
- Direktverkauf ohne Tisch: bestellen, kassieren und ausgeben in einem Schritt, mit Historie und Storno.

### Küche

- Automatischer Bon-Druck von Bestell- und Küchenbons an zugeordnete Bondrucker, pro Warenkategorie konfigurierbar.

### Kassenführung

- Fortlaufend nummerierte Kassensitzungen eröffnen und schließen.
- Anfangsbestand (Wechselgeld) zu Veranstaltungsbeginn erfassen.
- Soll-Kassenbestand jederzeit einsehen, aufgeschlüsselt nach Komponenten.
- Einlagen und Entnahmen (Geldtransit) buchen.
- Kassensturz: Ist-Bestand eingeben, Differenz berechnen und die Abweichung automatisch verbuchen.
- Tagesabschluss (Z-Bon) mit fortlaufender Nummer und Stammdaten-Snapshot.

### Abrechnung und Reporting

- Tagesabrechnung über alle Umsätze, Zahlungen und offenen Beträge, nach Steuersatz aufgeschlüsselt.
- Abrechnung je Tisch und je Servicekraft.
- Produktumsatz-Reporting: meistverkaufte Varianten, Mengen und Einnahmen je Produkt.
- DSFinV-K-Export als ZIP-Archiv (Version 2.4) für die Finanzverwaltung.

### Verwaltung, Sicherheit und Compliance

- Admin-Bereich für Produkte (mit Varianten und Steuersätzen), Tische, Benutzer und Betreiber-Stammdaten.
- Rollenmodell mit den Rollen Admin, Serviceleitung und Service.
- Sicheres Onboarding per Einmalpasswort, Passwort-Hashing mit Argon2id und Anmeldung über JWT.
- Event-Sourcing für eine lückenlose, unveränderliche Bestellhistorie (GoBD-konform durch ein Append-only-Kassenjournal).
- Anbindung einer BSI-zertifizierten Cloud-TSE von fiskaly, die jeden Vorgang signiert.
- Gesetzeskonforme Belegausgabe mit TSE-Signatur, QR-Code, Steuersatz und Betreiberadresse.
