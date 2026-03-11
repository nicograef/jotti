# Produktbeschreibung — jotti

Dieses Dokument definiert die Produktidentität von jotti: Positionierung, Zielgruppe, Abgrenzung, Kernbotschaften und Marketingtexte. Es dient als zentrale Referenz für alle externen Kommunikationskanäle (README, Website, Social Media, Pressemitteilungen).

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
10. [Marketingtexte](#10-marketingtexte)

---

## 1. Claim und Kurzbeschreibung

### 1.1 Claim

> **jotti — Das kostenlose Kassensystem für Vereinsfeste.**

### 1.2 Kurzbeschreibung (Einzeiler)

> jotti ist ein kostenloses, quelloffenes Mobile-Kassensystem (mPOS) für Vereine und gemeinnützige Organisationen — self-hosted, ohne Abo, ohne Hardware.

### 1.3 Kurzbeschreibung (3 Sätze)

> jotti ist ein kostenloses, quelloffenes Gastronomie-Kassensystem für Vereinsfeste, Weihnachtsmärkte, Konzerte und andere Non-Profit-Veranstaltungen. Servicekräfte nehmen Bestellungen direkt auf ihrem Smartphone auf, liefern aus, kassieren und stornieren — alles pro Tisch, alles im Browser. Kein Cloud-Abo, keine spezielle Hardware, kein Zahlungsgateway — einfach auf dem eigenen Server installieren und loslegen.

### 1.4 Elevator Pitch

> Euer Verein plant ein Sommerfest und braucht ein Kassensystem? Vergesst teure POS-Software mit monatlichen Gebühren. jotti ist ein kostenloses, Open-Source-Kassensystem, das auf jedem Smartphone läuft. Eure Servicekräfte öffnen einfach den Browser, nehmen Bestellungen auf, liefern aus und kassieren — pro Tisch, übersichtlich und schnell. Keine Installation auf dem Handy, keine spezielle Hardware, keine laufenden Kosten. Self-hosted per Docker, in Minuten einsatzbereit. Für Vereine und gemeinnützige Organisationen — kostenlos, für immer.

---

## 2. Positionierung

### 2.1 Positioning Statement

**Für** eingetragene Vereine, gemeinnützige Organisationen und Non-Profit-Veranstalter, **die** ein einfaches Kassensystem für ihre Gastronomie-Veranstaltungen brauchen, **ist jotti** ein kostenloses, quelloffenes Mobile-Point-of-Sale-System, **das** ohne Hardware-Investition, ohne Cloud-Abo und ohne technisches Vorwissen den kompletten Kassenbetrieb auf dem Smartphone ermöglicht. **Anders als** kommerzielle POS-Systeme wie Orderbird, Toast oder Zettle **erfordert jotti** keine laufenden Kosten, keine Kartenterminals und keine Anbieter-Abhängigkeit — es gehört dem Verein, läuft auf dem eigenen Server und ist speziell für den ehrenamtlichen Einsatz gebaut.

### 2.2 Marktkategorie

| Dimension       | Einordnung                                                                                      |
| --------------- | ----------------------------------------------------------------------------------------------- |
| Kategorie       | Mobile Point of Sale (mPOS) / Gastronomie-Kassensystem                                          |
| Segment         | Non-Profit / Vereinsgastronomie / Event-Catering                                                |
| Architektur     | Self-hosted, Source-Available, Mobile-first Web-App                                             |
| Preismodell     | Kostenlos für gemeinnützige Organisationen (AGPL-3.0 + Non-Commercial)                          |
| Wettbewerbsfeld | Vereinfacht: Stift & Papier → Excel → jotti. Kommerzielles Äquivalent: Orderbird, Toast, Zettle |

jotti positioniert sich bewusst **unterhalb** kommerzieller Kassensysteme: weniger Features, dafür null Kosten, null Komplexität und exakt der Funktionsumfang, den ein Vereinsfest braucht.

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

> „Ich organisiere seit Jahren unser Vereinsfest. Bisher hat jede Servicekraft mit Stift und Zettel gearbeitet. Am Ende saßen wir stundenlang zusammen, um die Abrechnung hinzubekommen. Ich will eine einfache Lösung, die nichts kostet und die jeder bedienen kann."

- Entscheider für Software-Anschaffungen
- Technisch grundlegend versiert (kann Docker nach Anleitung starten)
- Will Kontrolle über Produkte, Tische und Benutzer
- Braucht verlässliche Abrechnung

#### Maria, 23 — Servicekraft (Service)

> „Ich helfe beim Vereinsfest mit und bediene Tische. Ich will einfach auf meinem Handy sehen, was bestellt wurde, die Bestellung aufgeben, die Getränke bringen und am Tisch kassieren. Mehr brauche ich nicht."

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

---

## 5. Lösung

### 5.1 Was jotti bietet

jotti löst genau diese Probleme mit einem radikal einfachen Ansatz:

| Problem      | jotti-Lösung                                                                  |
| ------------ | ----------------------------------------------------------------------------- |
| Kosten       | Kostenlos und quelloffen — dauerhaft, keine versteckten Kosten                |
| Hardware     | Keine — läuft auf jedem Smartphone mit Browser (BYOD)                         |
| Komplexität  | Nur die Features, die ein Vereinsfest braucht — nicht mehr                    |
| Abhängigkeit | Self-hosted auf dem eigenen Server — volle Datenkontrolle                     |
| Abrechnung   | Echtzeit-Saldo pro Tisch, lückenlose Bestellhistorie, transparente Abrechnung |

### 5.2 So funktioniert's

```
1. Admin richtet ein          2. Team meldet sich an        3. Loslegen
┌──────────────────┐         ┌──────────────────┐         ┌──────────────────┐
│ • Produkte &     │         │ • Browser öffnen │         │ • Bestellungen   │
│   Varianten      │   ───►  │ • Einmalpasswort │   ───►  │   aufgeben       │
│   anlegen        │         │   eingeben       │         │ • Liefern        │
│ • Tische anlegen │         │ • Eigenes        │         │ • Kassieren      │
│ • Benutzer       │         │   Passwort       │         │ • Stornieren     │
│   erstellen      │         │   setzen         │         │ • Saldo prüfen   │
└──────────────────┘         └──────────────────┘         └──────────────────┘
```

---

## 6. Kernfeatures

### 6.1 Kassenbetrieb (Service-Bereich)

| Feature                    | Beschreibung                                                                             |
| -------------------------- | ---------------------------------------------------------------------------------------- |
| **Bestellungen aufgeben**  | Produkte und Varianten auswählen, Menge wählen, auf den Tisch buchen                     |
| **Lieferungen bestätigen** | Ausgelieferte Positionen als geliefert markieren                                         |
| **Zahlungen registrieren** | Am Tisch kassieren — Teilzahlung möglich                                                 |
| **Stornierungen**          | Falschbestellungen rückgängig machen (nur Serviceleitung/Admin)                          |
| **Tisch-Übersicht**        | Offener Saldo, bestellte/gelieferte/bezahlte Positionen, Bestellhistorie auf einen Blick |
| **Tischauswahl**           | Alle Tische mit Status und offenem Betrag — sofort den richtigen Tisch finden            |
| **Küchendisplay (KDS)**    | Eingehende Bestellungen in Echtzeit auf einem Bildschirm in Küche oder Ausgabe anzeigen  |
| **Bon-Druck**              | Bestell- und Küchenbons direkt an einen Bondrucker senden                                |

### 6.2 Verwaltung (Admin-Bereich)

| Feature                | Beschreibung                                                 |
| ---------------------- | ------------------------------------------------------------ |
| **Produktverwaltung**  | Produkte mit Varianten (Größe, Preis) anlegen und bearbeiten |
| **Tischverwaltung**    | Tische anlegen, benennen, aktivieren/deaktivieren            |
| **Benutzerverwaltung** | Accounts erstellen, Rollen zuweisen, Passwörter zurücksetzen |
| **Rollenmodell**       | Drei Rollen: Admin, Serviceleitung, Servicekraft             |

### 6.3 Abrechnung und Reporting

| Feature                         | Beschreibung                                                                        |
| ------------------------------- | ----------------------------------------------------------------------------------- |
| **Tagesabrechnung**             | Gesamtübersicht aller Umsätze, Zahlungen und offenen Beträge am Ende des Tages      |
| **Abrechnung pro Tisch**        | Detaillierte Aufstellung aller Bestellungen, Zahlungen und Stornierungen je Tisch   |
| **Abrechnung pro Servicekraft** | Umsatz und Transaktionen pro Servicekraft — für Transparenz und Nachvollziehbarkeit |
| **Reporting**                   | Auswertungen über Produktumsätze, meistverkaufte Varianten und Gesamteinnahmen      |

### 6.4 Sicherheit und Zuverlässigkeit

| Feature                              | Beschreibung                                                                 |
| ------------------------------------ | ---------------------------------------------------------------------------- |
| **Sicheres Onboarding**              | Einmalpasswort + eigenes Passwort setzen — kein unsicherer Account-Versand   |
| **JWT-Authentifizierung**            | Sichere, tokenbasierte Anmeldung                                             |
| **Rollenbasierte Zugriffskontrolle** | Jede Rolle sieht nur, was sie darf                                           |
| **Event-Sourcing**                   | Lückenlose, unveränderliche Bestellhistorie — kein Datenverlust, kein Betrug |
| **HTTPS**                            | Verschlüsselte Kommunikation per Let's Encrypt                               |

---

## 7. Abgrenzung

### 7.1 jotti vs. kommerzielle POS-Systeme

| Eigenschaft           | jotti                                            | Orderbird / Toast / Zettle           |
| --------------------- | ------------------------------------------------ | ------------------------------------ |
| **Preis**             | Kostenlos                                        | 30–100 €/Monat + Hardware            |
| **Hardware**          | Keine — eigenes Smartphone (BYOD)                | Dedizierte Terminals, iPad, Drucker  |
| **Hosting**           | Self-hosted (Docker)                             | Cloud-SaaS                           |
| **Zielgruppe**        | Vereine, gemeinnützige Organisationen            | Gastronomie-Betriebe, Restaurants    |
| **Zahlungsarten**     | Bargeld-Tracking                                 | Karte, NFC, Online-Payment           |
| **Fiskalkonformität** | GoBD-Grundsätze durch Event-Sourcing (keine TSE) | Zertifizierte TSE, GoBD, KassenSichV |
| **Offline-Modus**     | Server-basiert (lokales WLAN reicht)             | Cloud-abhängig oder Hybrid           |
| **Einrichtungszeit**  | Minuten (Docker Compose)                         | Tage bis Wochen                      |
| **Schulung**          | Keine — intuitive Mobile-UI                      | Oft erforderlich                     |
| **Vertragslaufzeit**  | Keine                                            | 12–24 Monate                         |
| **Datenhoheit**       | Volle Kontrolle (eigener Server)                 | Daten beim Anbieter                  |
| **Quellcode**         | Einsehbar (Source-Available)                     | Proprietär                           |
| **Küchendisplay**     | Integriertes KDS                                 | Meist kostenpflichtiges Add-on       |
| **Bon-Druck**         | Bondrucker-Anbindung                             | Standard (oft proprietär)            |
| **Abrechnung**        | Tagesabrechnung, pro Tisch & Servicekraft        | Umfangreiche Reporting-Suites        |

### 7.2 Was jotti bewusst NICHT ist

jotti ist kein Allzweck-Kassensystem. Folgende Features sind bewusst **nicht** enthalten:

- ❌ Kartenzahlung / Zahlungsgateway
- ❌ Zertifizierte TSE (kryptografische Technische Sicherheitseinrichtung)
- ❌ Reservierungssystem
- ❌ Inventory / Warenwirtschaft
- ❌ Lieferservice-Integration
- ❌ Multi-Standort-Verwaltung
- ❌ Kundenverwaltung / CRM
- ❌ Selbstbedienungs-Kiosk
- ❌ Trinkgeld-Tracking

Diese bewusste Reduktion ist ein Feature, kein Mangel. Jedes zusätzliche Feature erhöht Komplexität, Wartungsaufwand und Einarbeitungszeit — alles, was ein ehrenamtliches Team bei einem Vereinsfest nicht braucht.

### 7.3 Fiskalkonformität im Detail

jotti enthält keine zertifizierte TSE (Technische Sicherheitseinrichtung) und ist damit nicht für steuerlich buchungspflichtige Dauerbetriebe geeignet. Allerdings erfüllt jotti durch sein architektonisches Fundament viele der GoBD-Grundsätze:

| GoBD-Grundsatz                       | Umsetzung in jotti                                                                    |
| ------------------------------------ | ------------------------------------------------------------------------------------- |
| **Unveränderbarkeit**                | Event-Sourcing: alle Transaktionen sind append-only, nachträgliche Änderung unmöglich |
| **Nachvollziehbarkeit**              | Lückenloses Kassenjournal pro Tisch mit vollständiger Bestellhistorie                 |
| **Vollständigkeit**                  | Jede Bestellung, Zahlung, Lieferung und Stornierung wird als Event erfasst            |
| **Zeitgerechte Buchung**             | Events werden in Echtzeit mit Zeitstempel gespeichert                                 |
| **Ordnungsmäßigkeit**                | Strukturiertes Datenmodell, typisierte Events, klare Zuordnung zu Tisch und Benutzer  |
| **Kryptografische Verkettung (TSE)** | ❌ Nicht vorhanden — keine zertifizierte Sicherheitseinrichtung                       |

Für gemeinnützige Vereine, die keine Kassenpflicht nach KassenSichV haben, bietet jotti damit ein hohes Maß an Transparenz und Nachvollziehbarkeit — ohne den Overhead einer zertifizierten TSE.

### 7.4 Einsatzprofil

| Kriterium                | jotti geeignet für                             | jotti NICHT geeignet für                               |
| ------------------------ | ---------------------------------------------- | ------------------------------------------------------ |
| Betriebsart              | Temporäre Veranstaltungen (1–3 Tage)           | Dauerbetrieb (Restaurant, Café)                        |
| Organisation             | Vereine, gemeinnützige Orgs, NPOs              | Kommerzielle Gastro-Betriebe                           |
| Zahlungsart              | Bargeld                                        | Kartenzahlung, NFC, Online                             |
| Team                     | Ehrenamtliche Helfer (5–30 Personen)           | Professionelles Gastro-Personal                        |
| Buchungspflichten        | GoBD-Grundsätze weitgehend erfüllt (keine TSE) | Steuerlich buchungspflichtige Betriebe mit TSE-Pflicht |
| Veranstaltungsgröße      | Klein bis mittel (5–50 Tische)                 | Großveranstaltungen mit 100+ Tischen                   |
| Technische Infrastruktur | WLAN + ein Server (auch Raspberry Pi möglich)  | Kein Server oder kein WLAN verfügbar                   |

---

## 8. Alleinstellungsmerkmale (USPs)

### 8.1 Die fünf USPs von jotti

| #   | USP                          | Beschreibung                                                                                                                  |
| --- | ---------------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| 1   | **Kostenlos für Vereine**    | Keine Lizenzgebühren, kein Abo, keine versteckten Kosten. Für immer kostenlos für gemeinnützige Organisationen.               |
| 2   | **Kein Hardware-Kauf**       | Jedes Smartphone wird zur Kasse. Kein Terminal, kein iPad, kein Drucker — BYOD (Bring Your Own Device).                       |
| 3   | **In Minuten einsatzbereit** | Docker Compose starten, Produkte anlegen, Team einladen — fertig. Keine wochenlange Einrichtung.                              |
| 4   | **Lückenlose Transparenz**   | Event-Sourcing garantiert eine unveränderliche Bestellhistorie. Jede Bestellung, Zahlung und Stornierung ist nachvollziehbar. |
| 5   | **Volle Datenhoheit**        | Self-hosted auf dem eigenen Server. Keine Daten in fremden Clouds. Der Verein behält die volle Kontrolle.                     |

### 8.2 Der jotti-Vorteil in einem Satz

> Wo andere Vereine mit Stift und Papier jonglieren oder teure Abos bezahlen, bietet jotti ein kostenloses, mobiles Kassensystem, das in Minuten läuft und jede Bestellung lückenlos nachvollziehbar macht.

---

## 9. Lizenz und Kosten

### 9.1 Lizenzmodell

jotti ist lizenziert unter **AGPL-3.0-or-later mit Zusatzbedingungen** (Source-Available, Non-Commercial).

| Nutzungsfall                                                | Erlaubt?                          |
| ----------------------------------------------------------- | --------------------------------- |
| Eingetragene Vereine (e.V.) für eigene Veranstaltungen      | ✅ Kostenlos                      |
| Gemeinnützige Organisationen (gGmbH, gUG, Stiftungen, NGOs) | ✅ Kostenlos                      |
| Nicht-kommerzieller Fork unter gleichen Lizenzbedingungen   | ✅ Kostenlos                      |
| Kommerzielles SaaS oder gewerbliche Nutzung                 | ❌ Nicht ohne kommerzielle Lizenz |
| Proprietäre Abspaltung                                      | ❌ Nicht erlaubt                  |

### 9.2 Was „kostenlos" bedeutet

- **Keine Lizenzgebühren** — nicht jetzt, nicht später
- **Keine Nutzungslimits** — keine Beschränkung auf Tische, Produkte oder Benutzer
- **Kein Freemium** — der volle Funktionsumfang, keine versteckte Paywall
- **Keine Werbung** — keine Ads, kein Tracking, keine Datenmonetarisierung
- **Quellcode einsehbar** — Transparenz und Vertrauen durch offenen Code

### 9.3 Einzige Kosten

Die einzigen Kosten, die entstehen, sind Infrastrukturkosten für das Hosting — diese fallen unabhängig von jotti an:

| Hosting-Option              | Geschätzte Kosten       |
| --------------------------- | ----------------------- |
| Raspberry Pi (selbst)       | ~50 € einmalig          |
| VPS (z. B. Hetzner, Netcup) | ~3–5 €/Monat            |
| Vereinseigener Server       | 0 € (bereits vorhanden) |

---

## 10. Marketingtexte

### 10.1 Website-Hero (Startseite)

> ### Das kostenlose Kassensystem für euer Vereinsfest.
>
> jotti ist ein quelloffenes Mobile-Kassensystem für Vereine und gemeinnützige Organisationen. Bestellungen aufnehmen, ausliefern, kassieren — direkt auf dem Smartphone, pro Tisch, ohne spezielle Hardware.
>
> **Kostenlos. Self-hosted. Open Source.**

### 10.2 Feature-Teaser (Social Media / Flyer)

> 📱 **Euer Smartphone wird zur Kasse.**
> Bestellungen aufnehmen, Lieferungen bestätigen, am Tisch kassieren — alles im Browser, auf jedem Handy.
>
> 💰 **Kostenlos für Vereine.**
> Kein Abo, kein Kleingedrucktes. jotti ist quelloffen und für gemeinnützige Organisationen dauerhaft kostenlos.
>
> 🔒 **Eure Daten bleiben bei euch.**
> Self-hosted auf eurem eigenen Server. Keine Cloud, keine Abhängigkeit, volle Kontrolle.

### 10.3 GitHub-Beschreibung (Repository)

> A free, open-source mobile POS system for non-profit events. Built for volunteer-run festivals, charity markets, and community gatherings. Self-hosted, mobile-first, no special hardware required.

### 10.4 Einzeiler für verschiedene Kontexte

| Kontext         | Text                                                                                                          |
| --------------- | ------------------------------------------------------------------------------------------------------------- |
| GitHub Topics   | `pos`, `point-of-sale`, `mpos`, `gastro`, `non-profit`, `volunteer`, `festival`, `open-source`, `self-hosted` |
| GitHub About    | Free, open-source mobile POS for non-profit events. Self-hosted, mobile-first, no hardware needed.            |
| Technisch       | Mobile-first POS mit Event-Sourcing, Go-Backend, React-Frontend — self-hosted per Docker Compose.             |
| Nicht-technisch | Das kostenlose Kassensystem für Vereinsfeste — auf jedem Smartphone, ohne Abo, ohne Hardware.                 |
| Vereinsvorstand | Einfach Produkte und Tische anlegen, Team einladen, loslegen. Abrechnung auf Knopfdruck.                      |
| Servicekraft    | Browser öffnen, Tisch wählen, bestellen, liefern, kassieren. Fertig.                                          |

### 10.5 SEO-Keywords

**Primär:** Kassensystem Verein, Kassensystem Vereinsfest, POS Verein, kostenloses Kassensystem, Kassensystem ehrenamtlich, Gastronomie Kassensystem kostenlos

**Sekundär:** mPOS Non-Profit, Open-Source Kassensystem, Kassensystem Weihnachtsmarkt, Kassensystem Sommerfest, Mobile Kasse Verein, Kasse Smartphone, Self-hosted POS, Kassensystem ohne Abo

**Long-Tail:** kostenloses Kassensystem für Vereinsfeste, Open Source Gastronomie Kassensystem für Vereine, mobiles Kassensystem ohne Hardware für ehrenamtliche Veranstaltungen
