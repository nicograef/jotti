# Produktbeschreibung — jotti

Produktidentität von jotti: Positionierung, Zielgruppe, Abgrenzung, Kernfeatures, USPs und Lizenz. Kanonische Referenz für Entwickler und Agenten.

---

## Inhaltsverzeichnis

1. [Claim und Kurzbeschreibung](#1-claim-und-kurzbeschreibung)
2. [Positionierung](#2-positionierung)
3. [Zielgruppe](#3-zielgruppe)
4. [Problemstellung](#4-problemstellung)
5. [Lösung](#5-lösung)
6. [Kernfeatures](#6-kernfeatures)
7. [Abgrenzung](#7-abgrenzung)
8. [Alleinstellungsmerkmale (USPs)](#8-alleinstellungsmerkmale-usps)
9. [Lizenz und Kosten](#9-lizenz-und-kosten)

---

## 1. Claim und Kurzbeschreibung

> **jotti — Das kostenlose Kassensystem für Vereinsfeste.**

jotti ist ein kostenloses, quelloffenes Gastronomie-Kassensystem für Vereinsfeste, Weihnachtsmärkte, Konzerte und andere Non-Profit-Veranstaltungen. Servicekräfte nehmen Bestellungen direkt auf ihrem Smartphone auf, bestätigen die Ausgabe, kassieren und stornieren — alles pro Tisch, alles im Browser. Fiskalkonform mit TSE-Anbindung und DSFinV-K-Export, kein Cloud-Abo, keine spezielle Hardware — einfach auf dem eigenen Server installieren und loslegen.

---

## 2. Positionierung

### 2.1 Positioning Statement

**Für** eingetragene Vereine, gemeinnützige Organisationen und Non-Profit-Veranstalter, **die** ein einfaches und fiskalkonformes Kassensystem für ihre Gastronomie-Veranstaltungen brauchen, **ist jotti** ein kostenloses, quelloffenes Mobile-Point-of-Sale-System, **das** ohne Hardware-Investition, ohne Cloud-Abo und ohne technisches Vorwissen den kompletten Kassenbetrieb auf dem Smartphone ermöglicht — inklusive TSE-Anbindung, DSFinV-K-Export und Belegausgabe nach KassenSichV. **Anders als** kommerzielle POS-Systeme wie Orderbird, Toast oder Zettle **erfordert jotti** keine laufenden Kosten, keine Kartenterminals und keine Anbieter-Abhängigkeit — es gehört dem Verein, läuft auf dem eigenen Server und ist speziell für den ehrenamtlichen Einsatz gebaut.

### 2.2 Marktkategorie

| Dimension       | Einordnung                                                                                                          |
| --------------- | ------------------------------------------------------------------------------------------------------------------- |
| Kategorie       | Mobile Point of Sale (mPOS) / Gastronomie-Kassensystem                                                              |
| Segment         | Non-Profit / Vereinsgastronomie / Event-Catering                                                                    |
| Architektur     | Self-hosted, Source-Available, Mobile-first Web-App                                                                 |
| Preismodell     | Kostenlos für gemeinnützige Organisationen (proprietäre Source-Available-Lizenz, Nutzungsvereinbarung erforderlich) |
| Wettbewerbsfeld | Vereinfacht: Stift & Papier → Excel → jotti. Kommerzielles Äquivalent: Orderbird, Toast, Zettle                     |

jotti positioniert sich bewusst **unterhalb** kommerzieller Kassensysteme: weniger Features, dafür null Kosten, null Komplexität und exakt der Funktionsumfang, den ein Vereinsfest braucht — bei voller Fiskalkonformität.

---

## 3. Zielgruppe

### 3.1 Primäre Zielgruppe

| Segment                          | Beispiele                                                                    |
| -------------------------------- | ---------------------------------------------------------------------------- |
| **Eingetragene Vereine (e.V.)**  | Sportvereine, Musikvereine, Schützenvereine, Karnevalsvereine, Fördervereine |
| **Gemeinnützige Organisationen** | gGmbH, gUG, Stiftungen, kirchliche Träger                                    |
| **Non-Profit-Veranstalter**      | Bürgerinitiativen, Nachbarschaftsfeste, Schulfördervereine                   |

### 3.2 Typische Veranstaltungen

- Vereinsfeste und Sommerfeste
- Weihnachtsmärkte und Adventsmärkte
- Maihocks und Straßenfeste
- Konzerte und Kulturveranstaltungen
- Sportevents und Turniere mit Bewirtung
- Fasching, Karneval, Kirmes
- Jubiläen und Festbankette

### 3.3 Personas

#### Thomas, 52 — Vereinsvorstand (Admin)

> „Ich organisiere seit Jahren unser Vereinsfest. Bisher hat jede Servicekraft mit Stift und Zettel gearbeitet. Am Ende saßen wir stundenlang zusammen, um die Abrechnung hinzubekommen. Ich will eine einfache Lösung, die nichts kostet, die jeder bedienen kann und die bei einer Betriebsprüfung standhält."

- Entscheider für Software-Anschaffungen
- Technisch grundlegend versiert (kann Docker nach Anleitung starten)
- Will Kontrolle über Produkte, Tische und Benutzer
- Braucht verlässliche Abrechnung und Fiskalkonformität

#### Maria, 23 — Servicekraft (Service)

> „Ich helfe beim Vereinsfest mit und bediene Tische. Ich will einfach auf meinem Handy sehen, was bestellt wurde, die Bestellung aufnehmen, die Getränke bringen und am Tisch kassieren. Mehr brauche ich nicht."

- Nutzt ihr eigenes Smartphone (BYOD)
- Erwartet intuitive Bedienung ohne Schulung
- Arbeitet ehrenamtlich, hat wenig Geduld für komplizierte Software
- Mobile-first: alles im Browser

#### Felix, 34 — Serviceleitung (Senior Service)

> „Als Serviceleiter muss ich den Überblick behalten. Wenn etwas falsch bestellt wurde, muss ich stornieren können. Ich will schnell sehen, welcher Tisch was bestellt hat und wo noch was offen ist."

- Koordiniert das Service-Team
- Braucht Stornierungsberechtigung
- Will Tisch-Übersicht auf einen Blick
- Erfahrener Helfer, kennt den Ablauf

---

## 4. Problemstellung

### 4.1 Status quo bei Vereinsveranstaltungen

Die meisten Vereine in Deutschland bewirtschaften ihre Veranstaltungen mit einem von drei Ansätzen:

| Ansatz                | Probleme                                                                                                            |
| --------------------- | ------------------------------------------------------------------------------------------------------------------- |
| **Stift & Papier**    | Fehleranfällig, keine Echtzeit-Übersicht, aufwändige Abrechnung, unleserliche Handschrift, Bonblöcke gehen verloren |
| **Excel-Tabelle**     | Nicht Echtzeit-fähig, kein Mehrbenutzerbetrieb, keine Tisch-Zuordnung, manuelle Fehler                              |
| **Kommerzielles POS** | Zu teuer (Abo + Hardware), zu komplex, überdimensioniert für 2–3 Veranstaltungen pro Jahr                           |

### 4.2 Kernprobleme

1. **Kosten**: Kommerzielle Kassensysteme kosten 30–100 €/Monat — für Vereine, die 2–3 Mal im Jahr ein Fest veranstalten, nicht vertretbar.
2. **Hardware**: Dedizierte Terminals, Bondrucker, Kartenterminals — unnötige Investitionen für temporäre Veranstaltungen.
3. **Komplexität**: Funktionen wie Kartenzahlung, Reservierungssysteme, Küchendrucker, Inventory — alles, was ein Vereinsfest nicht braucht.
4. **Abhängigkeit**: Cloud-Abo, Anbieterbindung, kein Zugriff auf die eigenen Daten.
5. **Abrechnung**: Mit Stift und Papier fehlt die Transparenz. Am Ende des Abends stimmt die Kasse nicht, und niemand weiß warum.
6. **Fiskalkonformität**: Mit Stift und Papier oder Excel lässt sich die KassenSichV nicht erfüllen. Kommerzielle Systeme lösen das — aber zu einem Preis, der für Vereine unverhältnismäßig ist.

---

## 5. Lösung

### 5.1 Was jotti bietet

jotti löst genau diese Probleme mit einem radikal einfachen Ansatz:

| Problem           | jotti-Lösung                                                                       |
| ----------------- | ---------------------------------------------------------------------------------- |
| Kosten            | Kostenlos und quelloffen — dauerhaft, keine versteckten Kosten                     |
| Hardware          | Keine — läuft auf jedem Smartphone mit Browser (BYOD)                              |
| Komplexität       | Nur die Features, die ein Vereinsfest braucht — nicht mehr                         |
| Abhängigkeit      | Self-hosted auf dem eigenen Server — volle Datenkontrolle                          |
| Abrechnung        | Echtzeit-Saldo pro Tisch, lückenlose Bestellhistorie, transparente Abrechnung      |
| Fiskalkonformität | TSE-Anbindung, DSFinV-K-Export, Belegausgabe, Tagesabschluss — KassenSichV-konform |

---

## 6. Kernfeatures

| Feature                              | Beschreibung                                                                            |
| ------------------------------------ | --------------------------------------------------------------------------------------- |
| **Bestellungen aufnehmen**           | Produkte und Varianten auf den Tisch buchen — Menge, Steuersatz, Kommentar              |
| **Ausgabe bestätigen**               | Positionen als ausgegeben markieren                                                     |
| **Zahlung kassieren**                | Bargeld kassieren — Teilzahlung und Rückgeldberechnung                                  |
| **Stornierungen**                    | Falschbestellungen rückgängig machen (Serviceleitung/Admin, Pflichtkommentar)           |
| **Auszahlung leisten**               | Negativen Saldo ausgleichen nach Stornierung bereits kassierter Positionen              |
| **Bestellungen umbuchen**            | Bestellung atomisch auf einen anderen Tisch verschieben                                 |
| **Tisch-Übersicht**                  | Saldo, Bestellungen, Ausgaben, Zahlungen und Historie auf einen Blick                   |
| **Tisch-Favoriten**                  | Eigene Tische als Rich Cards auf dem Dashboard markieren                                |
| **Tisch-Schnellsuche**               | Tische per Name oder Nummer filtern                                                     |
| **Küchendisplay (KDS)**              | Eingehende Bestellungen in Echtzeit im Küchenbereich anzeigen                           |
| **Ausgabestationen**                 | Zubereitungsstatus verwalten — Servicekräfte sehen, wann Positionen abholbereit sind    |
| **Bon-Druck**                        | Bestell- und Küchenbons automatisch an Bondrucker senden (pro Kategorie konfigurierbar) |
| **Produktverwaltung**                | Produkte mit Varianten anlegen und bearbeiten (Name, Preis, Steuersatz)                 |
| **Tischverwaltung**                  | Tische anlegen, benennen, aktivieren/deaktivieren                                       |
| **Benutzerverwaltung**               | Accounts erstellen, Rollen zuweisen, Passwörter zurücksetzen                            |
| **Betreiber-Stammdaten**             | Vereinsname, Adresse und Steuernummer — für Belege, Z-Bons und DSFinV-K                 |
| **Rollenmodell**                     | Drei Rollen: Admin, Serviceleitung, Servicekraft                                        |
| **Abrechnungskreis**                 | Fortlaufend nummerierte Kassensitzungen eröffnen und schließen                          |
| **Anfangsbestand**                   | Wechselgeld zu Beginn erfassen                                                          |
| **Kassenbestand**                    | Soll-Kassenbestand jederzeit abrufen — aufgeschlüsselt nach Komponenten                 |
| **Kassenbewegungen**                 | Geldtransit, Privatentnahmen und Privateinlagen buchen                                  |
| **Kassensturz**                      | Ist-Bestand eingeben, Differenz berechnen, Abweichung automatisch buchen                |
| **Tagesabschluss (Z-Bon)**           | Formaler Abschluss mit fortlaufender Nummer, Umsatzaggregation und Stammdaten-Snapshot  |
| **Tagesabrechnung**                  | Gesamtübersicht aller Umsätze und Zahlungen nach Steuersatz                             |
| **Abrechnung pro Tisch**             | Aufstellung aller Bestellungen, Zahlungen und Stornierungen je Tisch                    |
| **Abrechnung pro Servicekraft**      | Umsatz und Transaktionen pro Servicekraft                                               |
| **Produktumsatz-Reporting**          | Verkaufte Mengen pro Variante, Ranking und Gesamteinnahmen                              |
| **Datenexport (CSV)**                | Umsätze und Bestellungen als CSV für die Vereinsbuchhaltung                             |
| **DSFinV-K-Export**                  | Maschinenlesbarer Export nach DSFinV-K v2.4 als ZIP-Archiv                              |
| **Sicheres Onboarding**              | Einmalpasswort + eigenes Passwort — kein unsicherer Account-Versand                     |
| **JWT-Authentifizierung**            | Tokenbasierte Anmeldung                                                                 |
| **Rollenbasierte Zugriffskontrolle** | Jede Rolle sieht nur, was sie darf                                                      |
| **Event-Sourcing**                   | Unveränderliche Bestellhistorie — GoBD-konform durch Append-only-Architektur            |
| **Kryptografische Hash-Chain**       | SHA-256-Verkettung aller Events — nachträgliche Manipulation ist nachweisbar            |
| **TSE-Anbindung (Cloud-TSE)**        | Integrierte Cloud-TSE-Schnittstelle mit fiskaly-Adapter — jeder Vorgang wird signiert   |
| **Belegausgabe**                     | Gesetzeskonforme Belege mit TSE-Signatur, QR-Code, Steuersatz und Betreiberadresse      |
| **Seriennummer der Kasse**           | Automatisch generierte UUID — für ELSTER-Meldung und DSFinV-K                           |
| **HTTPS**                            | Verschlüsselte Kommunikation per Let's Encrypt                                          |

---

## 7. Abgrenzung

### 7.1 jotti vs. kommerzielle POS-Systeme

| Eigenschaft           | jotti                                                                    | Orderbird / Toast / Zettle           |
| --------------------- | ------------------------------------------------------------------------ | ------------------------------------ |
| **Preis**             | Kostenlos                                                                | 30–100 €/Monat + Hardware            |
| **Hardware**          | Keine — eigenes Smartphone (BYOD)                                        | Dedizierte Terminals, iPad, Drucker  |
| **Hosting**           | Self-hosted (Docker)                                                     | Cloud-SaaS                           |
| **Zielgruppe**        | Vereine, gemeinnützige Organisationen                                    | Gastronomie-Betriebe, Restaurants    |
| **Zahlungsarten**     | Bargeld                                                                  | Karte, NFC, Online-Payment           |
| **Fiskalkonformität** | KassenSichV-konform: TSE (Cloud-TSE via fiskaly), DSFinV-K, Belegausgabe | Zertifizierte TSE, GoBD, KassenSichV |
| **Offline-Modus**     | Server-basiert (lokales WLAN reicht)                                     | Cloud-abhängig oder Hybrid           |
| **Einrichtungszeit**  | Minuten (Docker Compose)                                                 | Tage bis Wochen                      |
| **Schulung**          | Keine — intuitive Mobile-UI                                              | Oft erforderlich                     |
| **Vertragslaufzeit**  | Keine                                                                    | 12–24 Monate                         |
| **Datenhoheit**       | Volle Kontrolle (eigener Server)                                         | Daten beim Anbieter                  |
| **Quellcode**         | Einsehbar (Source-Available)                                             | Proprietär                           |
| **Küchendisplay**     | Integriertes KDS mit Zubereitungsstatus                                  | Meist kostenpflichtiges Add-on       |
| **Bon-Druck**         | Bondrucker-Anbindung (pro Kategorie konfigurierbar)                      | Standard (oft proprietär)            |
| **Kassenführung**     | Abrechnungskreis, Kassensturz, Z-Bon, Kassenbewegungen                   | Vergleichbare Funktionen             |
| **Abrechnung**        | Tagesabrechnung, pro Tisch, pro Servicekraft, Produktumsätze, CSV-Export | Umfangreiche Reporting-Suites        |

### 7.2 Was jotti bewusst NICHT ist

jotti ist kein Allzweck-Kassensystem. Folgende Features sind bewusst **nicht** enthalten:

- ❌ Kartenzahlung / Zahlungsgateway
- ❌ Reservierungssystem
- ❌ Inventory / Warenwirtschaft
- ❌ Lieferservice-Integration
- ❌ Multi-Standort-Verwaltung
- ❌ Kundenverwaltung / CRM
- ❌ Selbstbedienungs-Kiosk (Gäste bestellen/zahlen selbst) — der **personalbediente Direktverkauf an der Theke** ist hingegen enthalten

Diese bewusste Reduktion ist ein Feature, kein Mangel. Jedes zusätzliche Feature erhöht Komplexität, Wartungsaufwand und Einarbeitungszeit — alles, was ein ehrenamtliches Team bei einem Vereinsfest nicht braucht.

### 7.3 Fiskalkonformität

jotti ist ein **elektronisches Aufzeichnungssystem** im Sinne von § 1 KassenSichV und erfüllt die TSE-Pflicht nach § 146a AO — unabhängig von der Rechtsform des Betreibers (e.V., gGmbH, Stiftung) oder dem temporären Charakter einer Veranstaltung.

| Anforderung           | Umsetzung in jotti                                                            |
| --------------------- | ----------------------------------------------------------------------------- |
| Unveränderbarkeit     | Event-Sourcing (Append-Only) + SHA-256-Hash-Chain — Manipulation nachweisbar  |
| Kassenjournal         | Lückenlose, chronologische Transaktionshistorie                               |
| TSE-Signatur          | Integrierte Cloud-TSE-Schnittstelle (fiskaly) — jeder Vorgang wird signiert   |
| Belegausgabe          | Belege mit Pflichtfeldern nach § 6 KassenSichV inkl. TSE-Signatur und QR-Code |
| Steuersätze           | 19 % (Standard), 7 % (ermäßigt), 0 % — konfigurierbar pro Produktvariante     |
| Abrechnungskreis      | Fortlaufend nummerierte Kassensitzungen mit Tagesabschluss (Z-Bon)            |
| DSFinV-K-Export       | Vollständiger Export als ZIP (CSV + index.xml) nach DSFinV-K v2.4             |
| Seriennummer / ELSTER | UUID beim ersten Start; dokumentierte Anleitung für die ELSTER-Meldung        |

**Architekturprinzip:** Die Smartphones der Servicekräfte sind reine Eingabegeräte — TSE-Anbindung, Protokollierung und DSFinV-K-Persistenz laufen zentral im Backend; bei Verbindungsverlust blockiert die Webapp sofort (kein Offline-Kassieren). Da jotti self-hosted läuft, schließen Betreiber den Cloud-TSE-Vertrag selbst ab und hinterlegen die API-Schlüssel per `.env` (Bring Your Own TSE); ohne TSE-Konfiguration startet jotti nur im Entwicklungsmodus. Die rechtlichen Grundlagen sowie die Betreiberpflichten (ELSTER-Meldung, Datensicherung, 10-jährige GoBD-konforme Aufbewahrung) stehen in [docs/compliance.md](compliance.md).

### 7.4 Einsatzprofil

| Kriterium                | jotti geeignet für                                | jotti NICHT geeignet für             |
| ------------------------ | ------------------------------------------------- | ------------------------------------ |
| Betriebsart              | Temporäre Veranstaltungen (1–3 Tage)              | Dauerbetrieb (Restaurant, Café)      |
| Organisation             | Vereine, gemeinnützige Orgs, NPOs                 | Kommerzielle Gastro-Betriebe         |
| Zahlungsart              | Bargeld                                           | Kartenzahlung, NFC, Online           |
| Team                     | Ehrenamtliche Helfer (5–30 Personen)              | Professionelles Gastro-Personal      |
| Fiskalkonformität        | KassenSichV-konform (TSE, DSFinV-K, Belegausgabe) | —                                    |
| Veranstaltungsgröße      | Klein bis mittel (5–50 Tische)                    | Großveranstaltungen mit 100+ Tischen |
| Technische Infrastruktur | WLAN + ein Server (auch Raspberry Pi möglich)     | Kein Server oder kein WLAN verfügbar |

---

## 8. Alleinstellungsmerkmale (USPs)

Verdichtet auf sechs Punkte, die jotti von kommerziellen Kassensystemen und von Stift-und-Papier abheben:

| USP                          | Kurz                                                                 |
| ---------------------------- | -------------------------------------------------------------------- |
| **Kostenlos für Vereine**    | Keine Lizenzgebühren, kein Abo, keine versteckten Kosten             |
| **Kein Hardware-Kauf**       | Jedes Smartphone wird zur Kasse — BYOD statt Terminal, iPad, Drucker |
| **In Minuten einsatzbereit** | Docker Compose starten, Produkte anlegen, Team einladen              |
| **Lückenlose Transparenz**   | Event-Sourcing — unveränderliche, nachvollziehbare Bestellhistorie   |
| **Volle Datenhoheit**        | Self-hosted auf dem eigenen Server, keine fremde Cloud               |
| **Fiskalkonform ab Werk**    | TSE, DSFinV-K-Export, Belegausgabe und Tagesabschluss inklusive      |

Wie jotti die typischen Vereinsprobleme löst, zeigt [§5 Lösung](#5-lösung); der direkte Vergleich mit kommerziellen Systemen steht in [§7.1](#71-jotti-vs-kommerzielle-pos-systeme).

---

## 9. Lizenz und Kosten

jotti steht unter einer **proprietären Source-Available-Lizenz**: Der Quellcode ist öffentlich einsehbar, Nutzungsrechte werden aber nicht automatisch gewährt. Eingetragene Vereine und gemeinnützige Organisationen (gGmbH, gUG, Stiftungen, NGOs) dürfen jotti **kostenlos** nutzen — nach Abschluss einer schriftlichen Nutzungsvereinbarung mit dem Autor (Nico Gräf). Quellcode lesen (Lernen, Evaluation, Sicherheitsaudit) und Pull Requests unter CLA sind ohne Vereinbarung erlaubt; Forks, Weitergabe sowie kommerzielle oder gewerbliche Nutzung erfordern eine separate Lizenz. Lizenzmodell, erlaubte Nutzungsfälle und die vollständigen Bedingungen: [lizenz-und-nutzung.md](lizenz-und-nutzung.md) und [nutzungsbedingungen.md](nutzungsbedingungen.md).

„Kostenlos" heißt: keine Lizenzgebühren, keine Nutzungslimits, kein Freemium, keine Werbung, kein Tracking — voller Funktionsumfang. Es entstehen nur Infrastrukturkosten, die unabhängig von jotti anfallen:

| Kostenart                   | Geschätzte Kosten                              |
| --------------------------- | ---------------------------------------------- |
| VPS (z. B. Hetzner, Netcup) | ~3–5 €/Monat                                   |
| Raspberry Pi (selbst)       | ~50 € einmalig                                 |
| Vereinseigener Server       | 0 € (bereits vorhanden)                        |
| Cloud-TSE (z. B. fiskaly)   | Abhängig vom Anbieter (BYOT — eigener Vertrag) |
