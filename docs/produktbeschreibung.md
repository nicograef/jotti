---
title: Produktbeschreibung
description: "Produktidentität von jotti: Positionierung, Zielgruppe, Abgrenzung, Kernfeatures und Lizenz."
---

Produktidentität von jotti: Positionierung, Zielgruppe, Abgrenzung, Kernfeatures, USPs und Lizenz. Kanonische Referenz für Entwickler und Agenten.

---

## 1. Claim und Kurzbeschreibung

> jotti: Das kostenlose Kassensystem für Vereinsfeste.

jotti ist ein kostenloses, quelloffenes Gastronomie-Kassensystem für Vereinsfeste, Weihnachtsmärkte, Konzerte und andere Non-Profit-Veranstaltungen. Servicekräfte nehmen Bestellungen direkt auf ihrem Smartphone auf, kassieren und stornieren, alles pro Tisch, alles im Browser. jotti bringt die fiskalischen Bausteine mit: eine BSI-zertifizierte Cloud-TSE, Belegausgabe nach § 146a AO, ein append-only Kassenjournal (GoBD) und den DSFinV-K-Export (v2.4). Den konformen Betrieb (TSE-Vertrag, Kassenmeldung, Aufbewahrung) verantwortet der Betreiber. Kein Cloud-Abo, keine spezielle Hardware, einfach auf dem eigenen Server installieren und loslegen.

---

## 2. Positionierung

### 2.1 Positioning Statement

Für eingetragene Vereine, gemeinnützige Organisationen und Non-Profit-Veranstalter, die ein einfaches Kassensystem mit fiskalischen Bausteinen für ihre Gastronomie-Veranstaltungen brauchen, ist jotti ein kostenloses, quelloffenes Mobile-Point-of-Sale-System, das ohne Hardware-Investition, ohne Cloud-Abo und ohne technisches Vorwissen den kompletten Kassenbetrieb auf dem Smartphone ermöglicht, inklusive TSE-Anbindung, Belegausgabe nach KassenSichV und DSFinV-K-Export (v2.4). Anders als kommerzielle POS-Systeme wie Orderbird, Toast oder Zettle erfordert jotti keine laufenden Kosten, keine Kartenterminals und keine Anbieter-Abhängigkeit; es gehört dem Verein, läuft auf dem eigenen Server und ist speziell für den ehrenamtlichen Einsatz gebaut.

### 2.2 Marktkategorie

| Dimension       | Einordnung                                                                                                          |
| --------------- | ------------------------------------------------------------------------------------------------------------------- |
| Kategorie       | Mobile Point of Sale (mPOS) / Gastronomie-Kassensystem                                                              |
| Segment         | Non-Profit / Vereinsgastronomie / Event-Catering                                                                    |
| Architektur     | Self-hosted, Source-Available, Mobile-first Web-App                                                                 |
| Preismodell     | Kostenlos für gemeinnützige Organisationen (proprietäre Source-Available-Lizenz, Nutzungsvereinbarung erforderlich) |
| Wettbewerbsfeld | Vereinfacht: Stift & Papier → Excel → jotti. Kommerzielles Äquivalent: Orderbird, Toast, Zettle                     |

jotti positioniert sich bewusst unterhalb kommerzieller Kassensysteme: weniger Features, dafür null Kosten, null Komplexität und exakt der Funktionsumfang, den ein Vereinsfest braucht.

---

## 3. Zielgruppe

### 3.1 Primäre Zielgruppe

| Segment                      | Beispiele                                                                    |
| ---------------------------- | ---------------------------------------------------------------------------- |
| Eingetragene Vereine (e.V.)  | Sportvereine, Musikvereine, Schützenvereine, Karnevalsvereine, Fördervereine |
| Gemeinnützige Organisationen | gGmbH, gUG, Stiftungen, kirchliche Träger                                    |
| Non-Profit-Veranstalter      | Bürgerinitiativen, Nachbarschaftsfeste, Schulfördervereine                   |

### 3.2 Typische Veranstaltungen

Vereins- und Sommerfeste, Weihnachts- und Adventsmärkte, Maihocks und Straßenfeste, Konzerte und Kulturveranstaltungen, Sportevents mit Bewirtung, Fasching/Karneval/Kirmes sowie Jubiläen und Festbankette.

### 3.3 Personas

- **Thomas, 52, Vereinsvorstand (Admin):** Entscheider für die Software-Anschaffung, mit technischen Grundkenntnissen (startet Docker nach Anleitung). Will eine kostenlose, von jedem bedienbare Lösung mit verlässlicher Abrechnung, die einer Betriebsprüfung standhält.
- **Maria, 23, Servicekraft (Service):** Nutzt ihr eigenes Smartphone (BYOD), arbeitet ehrenamtlich, erwartet intuitive Bedienung ohne Schulung. Will am Tisch bestellen und kassieren, mehr nicht.
- **Felix, 34, Serviceleitung (Senior Service):** Koordiniert das Service-Team, braucht Stornierungsberechtigung und die Tisch-Übersicht auf einen Blick.

---

## 4. Problem & Lösung

Die meisten Vereine bewirtschaften ihre Veranstaltungen mit Stift & Papier (fehleranfällig, keine Echtzeit-Übersicht, aufwändige Abrechnung), einer Excel-Tabelle (kein Mehrbenutzerbetrieb, keine Tisch-Zuordnung) oder einem kommerziellen POS (zu teuer, zu komplex, überdimensioniert für 2–3 Veranstaltungen im Jahr). jotti löst die daraus folgenden Kernprobleme mit einem radikal einfachen, auf die KassenSichV ausgelegten Ansatz:

| Problem                                                         | jotti-Lösung                                                                  |
| --------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| Kosten (kommerzielle Systeme: 30–100 €/Monat + Hardware)        | Kostenlos und quelloffen, dauerhaft, keine versteckten Kosten                 |
| Hardware (dedizierte Terminals, Bon- und Kartendrucker)         | Keine, läuft auf jedem Smartphone mit Browser (BYOD)                          |
| Komplexität (Reservierung, Inventory, Kartenzahlung)            | Nur die Funktionen, die ein Vereinsfest braucht, nicht mehr                   |
| Abhängigkeit (Cloud-Abo, Anbieterbindung, kein Datenzugriff)    | Self-hosted auf dem eigenen Server, volle Datenkontrolle                      |
| Abrechnung (am Ende stimmt die Kasse nicht, niemand weiß warum) | Echtzeit-Saldo pro Tisch, lückenlose Bestellhistorie, transparente Abrechnung |
| Fiskalkonformität (mit Papier/Excel nicht erfüllbar)            | TSE-Anbindung, Belegausgabe und Tagesabschluss; DSFinV-K-Export (v2.4)        |

---

## 5. Kernfeatures

Mit _(in Entwicklung)_ markierte Features sind geplant, aber noch nicht umgesetzt; Status pro Anforderung: [anforderungen.md](anforderungen.md).

| Feature                                    | Beschreibung                                                                                                                    |
| ------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------- |
| Bestellung                                 | Produkte und Varianten auf den Tisch buchen (Menge, Steuersatz, Kommentar)                                                      |
| Zahlung kassieren                          | Bargeld kassieren: Teilzahlung und Rückgeldberechnung                                                                           |
| Stornierung                                | Falschbestellungen rückgängig machen (Serviceleitung/Admin): bezahlte Positionen als Warenrücknahme mit Bar-Rückgabe, unbezahlte geldneutral                |
| Umbuchen & Direktverkauf                   | Bestellungen atomar zwischen Tischen verschieben; Barverkauf an der Theke in einem Schritt                                      |
| Tisch-Übersicht                            | Saldo, Bestellungen, Zahlungen und Historie auf einen Blick; Favoriten als Rich Cards, Schnellsuche per Name/Nummer             |
| Bon-Druck                                  | Bestell- und Küchenbons automatisch an Bondrucker senden (pro Kategorie konfigurierbar)                                         |
| Stammdaten-Verwaltung                      | Produkte (mit Varianten), Tische und Benutzer anlegen, bearbeiten und Rollen zuweisen                                           |
| Betreiber-Stammdaten                       | Vereinsname, Adresse und Steuernummer (für Belege, Z-Bons und DSFinV-K)                                                         |
| Rollenmodell                               | Drei Rollen: Admin, Serviceleitung, Servicekraft                                                                                |
| Kassensitzung                              | Fortlaufend nummerierte Abrechnungskreise eröffnen und schließen                                                                |
| Anfangsbestand & Kassenbestand             | Wechselgeld erfassen; Soll-Bestand jederzeit nach Komponenten aufgeschlüsselt abrufen                                           |
| Kassenbewegungen                           | Geldtransit (Einlage und Entnahme) buchen                                                                                       |
| Kassensturz                                | Ist-Bestand eingeben, Differenz berechnen, Abweichung automatisch buchen                                                        |
| Tagesabschluss (Z-Bon)                     | Formaler Abschluss mit fortlaufender Nummer, Umsatzaggregation und Stammdaten-Snapshot                                          |
| Abrechnung                                 | Tagesabrechnung nach Steuersatz, pro Tisch und pro Servicekraft                                                                 |
| Produktumsatz-Reporting _(in Entwicklung)_ | Verkaufte Mengen, Ranking und Gesamteinnahmen pro Variante                                                                      |
| DSFinV-K-Export                            | Maschinenlesbarer Export nach DSFinV-K v2.4 als ZIP-Archiv                                                                      |
| TSE-Anbindung (Cloud-TSE)                  | Integrierte Cloud-TSE-Schnittstelle mit fiskaly-Adapter; jeder Vorgang wird signiert                                            |
| Belegausgabe                               | Gesetzeskonforme Belege mit TSE-Signatur, QR-Code, Steuersatz und Betreiberadresse                                              |
| Kassen-Seriennummer                        | Automatisch generierte UUID (für ELSTER-Meldung und DSFinV-K)                                                                   |
| Event-Sourcing                             | Unveränderliche Bestellhistorie, GoBD-konform durch Append-only-Architektur                                                     |
| Sicherheit & Zugriff                       | Sicheres Onboarding (Einmalpasswort), Argon2id, JWT-Auth, rollenbasierte Zugriffskontrolle, HTTPS (Let's Encrypt)               |

---

## 6. Abgrenzung

### 6.1 jotti vs. kommerzielle POS-Systeme

| Eigenschaft       | jotti                                                            | Orderbird / Toast / Zettle           |
| ----------------- | ---------------------------------------------------------------- | ------------------------------------ |
| Preis             | Kostenlos                                                        | 30–100 €/Monat + Hardware            |
| Hardware          | Keine, eigenes Smartphone (BYOD)                                 | Dedizierte Terminals, iPad, Drucker  |
| Hosting           | Self-hosted (Docker)                                             | Cloud-SaaS                           |
| Zielgruppe        | Vereine, gemeinnützige Organisationen                            | Gastronomie-Betriebe, Restaurants    |
| Zahlungsarten     | Bargeld                                                          | Karte, NFC, Online-Payment           |
| Fiskalkonformität | TSE (Cloud-TSE/fiskaly) und Belegausgabe; DSFinV-K-Export (v2.4) | Zertifizierte TSE, GoBD, KassenSichV |
| Vertragslaufzeit  | Keine                                                            | 12–24 Monate                         |
| Datenhoheit       | Volle Kontrolle (eigener Server)                                 | Daten beim Anbieter                  |
| Quellcode         | Einsehbar (Source-Available)                                     | Proprietär                           |

### 6.2 Was jotti bewusst NICHT ist

jotti ist kein Allzweck-Kassensystem. Folgende Features sind bewusst nicht enthalten: Kartenzahlung/Zahlungsgateway, Reservierungssystem, Inventory/Warenwirtschaft, Lieferservice-Integration, Multi-Standort-Verwaltung, Kundenverwaltung/CRM und Selbstbedienungs-Kiosk (Gäste bestellen/zahlen selbst); der personalbediente Direktverkauf an der Theke ist hingegen enthalten.

Diese bewusste Reduktion ist ein Feature, kein Mangel: Jedes zusätzliche Feature erhöht Komplexität, Wartungsaufwand und Einarbeitungszeit, alles, was ein ehrenamtliches Team bei einem Vereinsfest nicht braucht.

### 6.3 Fiskalkonformität

jotti ist ein elektronisches Aufzeichnungssystem im Sinne von § 1 KassenSichV und erfüllt die TSE-Pflicht nach § 146a AO, unabhängig von der Rechtsform des Betreibers (e.V., gGmbH, Stiftung) oder dem temporären Charakter einer Veranstaltung.

| Anforderung           | Umsetzung in jotti                                                            |
| --------------------- | ----------------------------------------------------------------------------- |
| Unveränderbarkeit     | Event-Sourcing (Append-Only); jeder Geschäftsvorfall zusätzlich TSE-signiert  |
| Kassenjournal         | Lückenlose, chronologische Transaktionshistorie                               |
| TSE-Signatur          | Integrierte Cloud-TSE-Schnittstelle (fiskaly); jeder Vorgang wird signiert    |
| Belegausgabe          | Belege mit Pflichtfeldern nach § 6 KassenSichV inkl. TSE-Signatur und QR-Code |
| Steuersätze           | 19 % (Standard), 7 % (ermäßigt), 0 %, konfigurierbar pro Produktvariante      |
| Abrechnungskreis      | Fortlaufend nummerierte Kassensitzungen mit Tagesabschluss (Z-Bon)            |
| DSFinV-K-Export       | Vollständiger Export als ZIP (CSV + index.xml) nach DSFinV-K v2.4             |
| Seriennummer / ELSTER | UUID beim ersten Start; ELSTER-Meldeanleitung _(in Entwicklung)_              |

**Architekturprinzip:** Die Smartphones der Servicekräfte sind reine Eingabegeräte: TSE-Anbindung, Protokollierung und DSFinV-K-Persistenz laufen zentral im Backend; jeder Vorgang ist ein synchroner Backend-Request, ohne Verbindung ist kein Kassieren möglich (keine Offline-Erfassung). Da jotti self-hosted läuft, schließen Betreiber den Cloud-TSE-Vertrag selbst ab und hinterlegen die API-Schlüssel (Bring Your Own TSE); ohne TSE-Konfiguration bleiben Buchungen unsigniert, was nur für Test und Entwicklung gedacht und für echte Verkäufe nicht konform ist (§ 379 AO). Rechtliche Grundlagen und Betreiberpflichten (ELSTER-Meldung, Datensicherung, 10-jährige GoBD-konforme Aufbewahrung): [compliance.md](compliance.md).

### 6.4 Einsatzprofil

| Kriterium                | jotti geeignet für                            | jotti NICHT geeignet für             |
| ------------------------ | --------------------------------------------- | ------------------------------------ |
| Betriebsart              | Temporäre Veranstaltungen (1–3 Tage)          | Dauerbetrieb (Restaurant, Café)      |
| Organisation             | Vereine, gemeinnützige Orgs, NPOs             | Kommerzielle Gastro-Betriebe         |
| Zahlungsart              | Bargeld                                       | Kartenzahlung, NFC, Online           |
| Team                     | Ehrenamtliche Helfer (5–30 Personen)          | Professionelles Gastro-Personal      |
| Veranstaltungsgröße      | Klein bis mittel (5–50 Tische)                | Großveranstaltungen mit 100+ Tischen |
| Technische Infrastruktur | WLAN + ein Server (auch Raspberry Pi möglich) | Kein Server oder kein WLAN verfügbar |

---

## 7. Alleinstellungsmerkmale (USPs)

Sechs Punkte, die jotti von kommerziellen Kassensystemen und von Stift-und-Papier abheben:

| USP                      | Kurz                                                                    |
| ------------------------ | ----------------------------------------------------------------------- |
| Kostenlos für Vereine    | Keine Lizenzgebühren, kein Abo, keine versteckten Kosten                |
| Kein Hardware-Kauf       | Jedes Smartphone wird zur Kasse, BYOD statt Terminal, iPad, Drucker     |
| In Minuten einsatzbereit | Docker Compose starten, Produkte anlegen, Team einladen                 |
| Lückenlose Transparenz   | Event-Sourcing: unveränderliche, nachvollziehbare Bestellhistorie       |
| Volle Datenhoheit        | Self-hosted auf dem eigenen Server, keine fremde Cloud                  |
| Fiskalische Bausteine    | TSE, Belegausgabe und Tagesabschluss integriert; DSFinV-K-Export (v2.4) |

Wie jotti die typischen Vereinsprobleme löst, zeigt [§4 Problem & Lösung](#4-problem--lösung); der direkte Vergleich mit kommerziellen Systemen steht in [§6.1](#61-jotti-vs-kommerzielle-pos-systeme).

---

## 8. Lizenz und Kosten

jotti steht unter einer proprietären Source-Available-Lizenz: Der Quellcode ist öffentlich einsehbar, Nutzungsrechte werden aber nicht automatisch gewährt. Eingetragene Vereine und gemeinnützige Organisationen (gGmbH, gUG, Stiftungen, NGOs) dürfen jotti kostenlos nutzen, nach Abschluss einer schriftlichen Nutzungsvereinbarung mit dem Autor (Nico Gräf). Quellcode lesen (Lernen, Evaluation, Sicherheitsaudit) und Pull Requests unter CLA sind ohne Vereinbarung erlaubt; Forks, Weitergabe sowie kommerzielle oder gewerbliche Nutzung erfordern eine separate Lizenz. Lizenzmodell, erlaubte Nutzungsfälle und die vollständigen Bedingungen: [lizenzmodell.md](lizenzmodell.md) und [TERMS.md](../TERMS.md).

„Kostenlos" heißt: keine Lizenzgebühren, keine Nutzungslimits, kein Freemium, keine Werbung, kein Tracking, voller Funktionsumfang. Es entstehen nur Infrastrukturkosten, die unabhängig von jotti anfallen:

| Kostenart                   | Geschätzte Kosten                       |
| --------------------------- | --------------------------------------- |
| VPS (z. B. Hetzner, Netcup) | ~3–5 €/Monat                            |
| Raspberry Pi (selbst)       | ~50 € einmalig                          |
| Vereinseigener Server       | 0 € (bereits vorhanden)                 |
| Cloud-TSE von fiskaly       | Abhängig vom Anbieter (eigener Vertrag) |

Optionaler Service: Die Software bleibt für gemeinnützige Organisationen kostenlos. Unabhängig davon bietet der Autor auf Anfrage entgeltliche Unterstützung an (Einrichtung, Hosting-Hilfe, Tech-Support, Schulung); die Konditionen werden individuell vereinbart. Diese Leistungen sind freiwillig und keine Voraussetzung für die Nutzung.
