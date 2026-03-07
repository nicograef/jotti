# POS — Point of Sale: Einordnung von jotti

## Was ist ein POS-System?

Ein **Point-of-Sale-System (POS)** kombiniert Hardware und Software, um Verkaufsvorgänge zu erfassen, Zahlungen entgegenzunehmen und den Geschäftsbetrieb abzurechnen — die digitale Weiterentwicklung der klassischen Registrierkasse.

Kernfunktionen eines POS-Systems:

- **Kassierfunktion**: Transaktionserfassung, Preisberechnung, Zahlungsabwicklung
- **Produktkatalog**: Artikel, Preise, Varianten
- **Benutzer- und Rollenverwaltung**: Wer darf was?
- **Reporting**: Tagesumsätze, Bedienerstatistiken, Abschlussberichte

POS-Systeme sind breit aufgestellt — vom einfachen Einzelhandels-Terminal bis zum vollvernetzten Restaurantsystem mit Küchenmonitor, Bondrucker und Inventarverwaltung.

## POS in der Gastronomie

Gastronomische POS-Systeme haben gegenüber Einzelhandels-Kassensystemen spezifische Anforderungen:

| Merkmal | Beschreibung |
|---|---|
| **Tischbasierter Workflow** | Bestellungen werden pro Tisch geführt, nicht pro Kassenvorgang |
| **Offener Saldo** | Mehrere Bestellrunden, eine Endabrechnung pro Tisch |
| **Rollen** | Servicekraft, Küche/Ausgabe, Kassierer, Admin |
| **Bon-Druck** | Küchenbons, Quittungen, Tagesabschluss |
| **Geschwindigkeit** | Bestellaufnahme in Sekunden, auch unter Stress |

Verbreitete kommerzielle Gastro-POS-Systeme: **Lightspeed Restaurant**, **Orderbird**, **Toast**, **Square for Restaurants**.

## jotti: Positionierung

jotti ist ein **leichtgewichtiges Gastronomie-Kassensystem für Vereine und Non-Profit-Organisationen** bei Veranstaltungen.

### Zielgruppe

- Fußball-, Dorf-, Musik- und Sportvereine
- Bürger- und Kulturinitiativen
- Jede Organisation, die Vereinsfeste, Weihnachtsmärkte, Maihocks, Konzerte oder Sommerfeste ausrichtet

### Was jotti von kommerziellen Systemen unterscheidet

Kommerzielle Gastro-POS-Systeme sind für dauerhaften Gastronomiebetrieb ausgelegt: monatliche Abo-Kosten, Hardware-Bindung, komplexe Einrichtung, Fiskalgesetzgebung. Für eine Vereinsveranstaltung ist das zu viel.

jotti dagegen:

- **Kostenlos und Open Source** — keine laufenden Kosten
- **Self-hosted** auf eigener VM — kein Cloud-Abo, vollständige Datenkontrolle
- **Mobile-first** — Servicekräfte nutzen ihre eigenen Smartphones, keine proprietäre Hardware
- **Geringer Einrichtungsaufwand** — wenige Schritte bis zum ersten Einsatz
- **Qualitätsfokus** — professionelle Architektur, aber nur das notwendige Featureset

> Grundsatz: **Höchste Qualität statt umfangreichem Featureset.**

### Umfang und bewusste Grenzen

| Funktion | Status |
|---|---|
| Bestellaufnahme per Smartphone | ✅ |
| Mehrere Servicekräfte gleichzeitig | ✅ |
| Produkte mit Varianten und Kategorien | ✅ |
| Tischübersicht mit offenem Saldo | ✅ |
| Lieferungen bestätigen | ✅ |
| Bezahlungen (inkl. Teilzahlung) | ✅ |
| Stornierungen (nur Serviceleitung/Admin) | ✅ |
| Rollenverwaltung | ✅ |
| Lückenloses Kassenjournal (Event Sourcing) | ✅ |
| Bon-Druck | ❌ geplant |
| Elektronische Zahlungsabwicklung (Karte/NFC) | ❌ nicht geplant |
| Inventarverwaltung | ❌ nicht geplant |
| Multi-Tenant (mehrere Vereine) | ❌ Single-Tenant |
| TSE / GoBD-Zertifizierung | ❌ nicht relevant (Non-Profit) |

### Technische Positionierung

jotti ist ein **cloudnatives, mobile-first Web-POS**:

- Keine native App, kein App-Store — läuft im Smartphone-Browser
- Single-Tenant: eine Instanz pro Verein/Veranstaltung, deployed via Docker Compose
- Backend: Go (REST-API, POST-only)
- Datenbank: PostgreSQL mit Event-Sourcing für den Kassenbetrieb
- Frontend: React PWA (Progressive Web App), responsive für Smartphones

Das Event-Sourcing-Modell für Tischoperationen bietet einen **lückenlosen Kassenjournal**: Jede Bestellung, Zahlung, Lieferung und Stornierung ist als unveränderliches Event gespeichert — vergleichbar mit dem Kassenprotokoll einer herkömmlichen Registrierkasse, aber vollständig nachvollziehbar und ohne Manipulationsmöglichkeit.

## Ähnliche Produkte

| Produkt | Typ | Besonderheit |
|---|---|---|
| [Orderbird](https://www.orderbird.com) | Kommerziell, SaaS | iPad-POS für Gastronomie, DE |
| [Lightspeed Restaurant](https://www.lightspeedhq.com) | Kommerziell, SaaS | Vollständiges Restaurantsystem |
| [Toast POS](https://pos.toasttab.com) | Kommerziell, SaaS | Gastronomie (v.a. USA) |
| [Square for Restaurants](https://squareup.com) | Kommerziell, SaaS | Flexibles Gastro-POS |
| [UniCenta](https://unicenta.com) | Open Source | Allg. Einzelhandel/Gastronomie |
| [Floreant POS](http://floreant.org) | Open Source | Restaurant-fokussiert |

jotti unterscheidet sich von allen genannten durch seinen **spezifischen Fokus auf Vereinsveranstaltungen** und die **Freiheit von Hardware-Bindung und laufenden Kosten**.
