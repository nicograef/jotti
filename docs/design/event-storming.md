# Event Storming — jotti

> **⚠️ HISTORISCHES DOKUMENT — NICHT ALS REFERENZ VERWENDEN**
>
> Dieses Dokument ist ein Artefakt der Entwurfsphase und spiegelt **nicht** den aktuellen Stand des Projekts wider. Es wird nicht aktuell gehalten. Coding Agents sollen dieses Dokument **ignorieren** und stattdessen das [Entwickler-Handbuch](handbuch.md) als verbindliche Architektur-Referenz verwenden.

Dieses Dokument protokolliert das Ergebnis einer simulierten Event-Storming-Session für jotti — ein kostenloses, quelloffenes Mobile-Kassensystem (mPOS) für Vereine und gemeinnützige Organisationen. Ziel der Session war es, die gesamte Fachdomäne eines Vereins-Gastronomie-Kassensystems gemeinsam zu erkunden, Domain Events zu identifizieren, Bounded Contexts abzugrenzen und eine gemeinsame Sprache (Ubiquitous Language) zu entwickeln.

Die Simulation folgt der Methode von Alberto Brandolini und durchläuft vier Phasen: Big Picture (chaotische Exploration), Clustering (Ordnung und Timeline), Process Modelling (Commands, Akteure, Policies) und Software Design (Aggregate, Read Models, Bounded Contexts).

> **Methode:** Event Storming nach Alberto Brandolini — vom Big Picture über Process Modelling bis zum Software Design Level.

> **Hinweis:** Alle genannten Personen sind fiktiv. Die Session simuliert einen Workshop, wie er in der Praxis für ein Vereinsprojekt stattfinden könnte. Die Domänenerkenntnisse spiegeln die Anforderungen von jotti wider.

---

## Inhaltsverzeichnis

1. [Setup & Teilnehmer](#1-setup--teilnehmer)
2. [Phase 1 — Big Picture: Chaotische Exploration](#2-phase-1--big-picture-chaotische-exploration)
3. [Phase 2 — Clustering und Pivot Points](#3-phase-2--clustering-und-pivot-points)
4. [Phase 3 — Process Modelling: Commands, Akteure, Policies](#4-phase-3--process-modelling-commands-akteure-policies)
5. [Phase 4 — Software Design: Aggregate und Read Models](#5-phase-4--software-design-aggregate-und-read-models)
6. [Bounded Contexts und Domain Map](#6-bounded-contexts-und-domain-map)
7. [Hotspots und offene Fragen](#7-hotspots-und-offene-fragen)
8. [Ergebnisse und nächste Schritte](#8-ergebnisse-und-nächste-schritte)

**Anhänge:**

- [Anhang A — Vollständige Event-Liste](#anhang-a--vollständige-event-liste)
- [Anhang B — Stickies-Legende](#anhang-b--stickies-legende)

---

## 1. Setup & Teilnehmer

### 1.1 Rahmenbedingungen

| Attribut     | Wert                                                               |
| ------------ | ------------------------------------------------------------------ |
| **Datum**    | Samstag, 11. März 2026, 10:00–15:00 Uhr                            |
| **Ort**      | Vereinsheim Sportverein Grüntal, Besprechungsraum                  |
| **Dauer**    | ca. 5 Stunden (inkl. Pausen)                                       |
| **Methode**  | Event Storming (Big Picture → Process Modelling → Software Design) |
| **Material** | Papierwand (3 m × 1,5 m), Haftnotizen in 6 Farben, Marker          |

### 1.2 Notation (Stickies-Legende)

| Farbe     | Symbol | Bedeutung                                                                |
| --------- | ------ | ------------------------------------------------------------------------ |
| 🟠 Orange | —      | **Domain Event** — etwas, das in der Domäne passiert ist (Vergangenheit) |
| 🔵 Blau   | —      | **Command** — Absicht, etwas zu tun (Imperativ)                          |
| 🟡 Gelb   | —      | **Aggregate** — transaktionale Grenze, schützt Geschäftsregeln           |
| 🟣 Lila   | —      | **Policy** — automatische Reaktion: „Wenn X, dann Y"                     |
| 🟢 Grün   | —      | **Read Model** — Lese-Sicht / Projektion für die Anzeige                 |
| ❤️ Rot    | —      | **Hotspot** — Unsicherheit, Diskussionsbedarf, offene Frage              |

### 1.3 Teilnehmer

| Kürzel   | Rolle                      | Person | Hintergrund                                                            |
| -------- | -------------------------- | ------ | ---------------------------------------------------------------------- |
| **FAC**  | Facilitator & Moderatorin  | Lisa   | Event-Storming-Erfahrung, hält Timeboxen ein, moderiert die Session    |
| **DEV1** | Senior Fullstack Developer | Anna   | Go + React, DDD-Erfahrung, kennt Event-Sourcing-Patterns               |
| **DEV2** | Senior Fullstack Developer | Tim    | Go + React, Systemarchitektur, CQRS-Erfahrung                          |
| **DOM**  | Domänenexperte             | Rudi   | Langjähriger Vereinsvorstand, 15 Jahre Festorganisation                |
| **SRV1** | Servicekraft               | Jonas  | Bedient seit 3 Jahren Tische beim Vereinsfest                          |
| **SRV2** | Servicekraft               | Maria  | Neu dabei, erstes Vereinsfest als Servicekraft                         |
| **SRL**  | Serviceleitung             | Felix  | serviceleitungkraft, darf stornieren, koordiniert das Service-Team     |
| **ADM**  | Administrator              | Thomas | Kümmert sich um Software und Hardware beim Verein                      |
| **KAS**  | Kassenwart                 | Eva    | Zuständig für Finanzen, Tagesabrechnung, Vereinsbuchhaltung und Export |
| **VER1** | Vereinsmitglied (Ausgabe)  | Klaus  | Aktiver Helfer, übernimmt oft die Getränkeausgabe und Küchenmitarbeit  |

---

## 2. Phase 1 — Big Picture: Chaotische Exploration

**Zeitrahmen:** 10:00–11:30 Uhr (90 Minuten)

**FAC (Lisa):** „Willkommen! Heute erkunden wir gemeinsam, was bei einer Vereinsveranstaltung passiert — aus Sicht aller Beteiligten. Schreibt auf orange Haftnotizen alle Ereignisse, die ihr für wichtig haltet. Ein Ereignis ist etwas, das **passiert ist** — in der Vergangenheitsform, auf Deutsch. Also nicht ‚Bestellung aufnehmen', sondern ‚Bestellung aufgegeben'. Beginnt ruhig chaotisch — kein Sortieren, einfach draufkleben. Ich stelle den Timer auf 20 Minuten. Danach besprechen wir, was wir haben, und ergänzen, was fehlt."

_(Alle nehmen sich Haftnotizen und Marker. Für die nächsten 20 Minuten wird geschrieben und geklebt — ohne Diskussion.)_

### 2.1 Erste Runde: Rohe Domain Events

**SRV1 (Jonas)** klebt auf:

- 🟠 Bestellung aufgegeben
- 🟠 Produkt ausgewählt
- 🟠 Menge geändert
- 🟠 Bestellung abgebrochen
- 🟠 Getränk ausgeliefert
- 🟠 Essen ausgeliefert
- 🟠 Tisch bezahlt

**SRV2 (Maria)** klebt auf:

- 🟠 Fehler bei Bestellung gemacht
- 🟠 Bestellung storniert
- 🟠 Tisch gewechselt
- 🟠 Servicekraft angemeldet
- 🟠 Am falschen Tisch bestellt

**DOM (Rudi)** klebt auf:

- 🟠 Veranstaltung eröffnet
- 🟠 Veranstaltung beendet
- 🟠 Tagesabrechnung erstellt
- 🟠 Einnahmen gezählt
- 🟠 Produkt vergriffen

**KAS (Eva)** klebt auf:

- 🟠 Zahlung registriert
- 🟠 Teilzahlung registriert
- 🟠 Tagesabrechnung abgeschlossen
- 🟠 Umsatz pro Servicekraft ermittelt
- 🟠 Kassenbon ausgedruckt
- 🟠 Abrechnung pro Tisch erstellt
- 🟠 Daten exportiert

**ADM (Thomas)** klebt auf:

- 🟠 Benutzer angelegt
- 🟠 Passwort zurückgesetzt
- 🟠 Produkt hinzugefügt
- 🟠 Produktpreis geändert
- 🟠 Tisch angelegt
- 🟠 Tisch deaktiviert
- 🟠 System gestartet

**SRL (Felix)** klebt auf:

- 🟠 Stornierung durchgeführt
- 🟠 Bestellung auf anderen Tisch umgebucht
- 🟠 Getränkeausgabe informiert
- 🟠 Essensausgabe informiert
- 🟠 Servicekraft eingewiesen

**VER1 (Klaus)** klebt auf:

- 🟠 Getränkebon empfangen
- 🟠 Getränk zubereitet
- 🟠 Bestellung als fertig markiert
- 🟠 Küchenbestellung auf Display angezeigt
- 🟠 Position als „in Zubereitung" markiert

**DEV1 (Anna)** klebt auf:

- 🟠 Event gespeichert
- 🟠 Snapshot erstellt
- 🟠 JWT ausgestellt
- 🟠 Rate-Limit überschritten

**DEV2 (Tim)** klebt auf:

- 🟠 Seite geladen
- 🟠 Tischübersicht aufgerufen
- 🟠 Saldo aktualisiert
- 🟠 Session abgelaufen

### 2.2 Diskussion: Kassenbetrieb

**FAC (Lisa):** „Der Timer ist abgelaufen. Schauen wir uns an, was an der Wand hängt. Ich sehe viele Events rund um Bestellungen, Zahlungen und Lieferungen. Gehen wir das gemeinsam durch — was gehört zum Kernprozess?"

**SRV1 (Jonas):** „Der Grundablauf ist klar: Ich gehe an den Tisch, nehme die Bestellung auf, bringe die Sachen, und am Ende wird bezahlt. Dafür brauche ich ‚Bestellung aufgegeben', ‚Produkte ausgeliefert' und ‚Zahlung registriert'." _(K-01, K-03, K-02)_

**SRV2 (Maria):** „Und wie sehe ich, was bestellt wurde? Ich brauche eine Übersicht pro Tisch — was ist bestellt, was ist geliefert, was ist noch offen."

**SRV1 (Jonas):** „Genau. Ich tippe auf den Tisch und sehe alles auf einen Blick."

_(Ergänzt:)_

- 🟠 Tisch ausgewählt _(K-05)_

**KAS (Eva):** „Teilzahlungen sind entscheidend. Wenn drei Leute am Tisch sitzen, zahlt oft jeder separat. Das muss gehen." _(K-02)_

**SRV1 (Jonas):** „Ja, und ich muss sehen, was noch offen ist — damit ich dem Gast sagen kann, wie viel er noch schuldig ist."

**DOM (Rudi):** „Und das Kassenjournal? Wir brauchen eine lückenlose Historie, was an jedem Tisch passiert ist — für den Fall, dass es Unstimmigkeiten gibt." _(K-06)_

**DEV1 (Anna):** „Das Kassenjournal ist die Grundidee hinter Event-Sourcing. Jede Aktion am Tisch — Bestellung, Lieferung, Zahlung, Stornierung — wird als unveränderliches Ereignis gespeichert. Daraus ergibt sich der aktuelle Zustand."

_(Ergänzte Events:)_

- 🟠 Tischkonto ausgeglichen _(K-02 — wenn Saldo = 0)_
- 🟠 Kassenjournal eingesehen _(K-06)_

**SRL (Felix):** „Stornierung ist wichtig. Wenn eine Servicekraft falsch bestellt, muss ich das rückgängig machen können. Aber nicht jeder darf das — nur ich als Serviceleitung oder Thomas als Admin." _(K-04)_

**SRV2 (Maria):** „Das finde ich gut. Ich würde Angst haben, versehentlich was zu stornieren."

**DOM (Rudi):** „Bei uns gab es mal den Fall, dass jemand aus Versehen den ganzen Tisch storniert hat. Nur Leute mit Erfahrung sollten das dürfen."

_(🟣 Policy: Stornierung nur für Serviceleitung und Admin — K-04)_

#### Komfort-Features

**FAC (Lisa):** „Ich sehe noch einige Events, die nicht zum Kernprozess gehören, aber den Betrieb erleichtern würden. Jonas, du hast ‚Tisch bezahlt' — gibt es noch mehr solche Komfort-Dinge?"

**SRV1 (Jonas):** „Was mir im Alltag auffällt: Manchmal bediene ich einen Tisch, an dem zwei Gruppen sitzen. Da wäre es super, wenn ich der Bestellung einen Namen geben könnte — ‚Familie Müller' oder ‚Gruppe links'." _(K-07)_

_(Ergänzt:)_

- 🟠 Bestellung benannt _(K-07)_

**SRV2 (Maria):** „Und was, wenn man den falschen Tisch erwischt hat? Beim letzten Fest habe ich auf den falschen Tisch bestellt."

**SRL (Felix):** „Dann müssen wir die Bestellung umbuchen — stornieren am einen Tisch, neu aufgeben am anderen. Das mache ich als Serviceleitung." _(K-08)_

_(Ergänzt:)_

- 🟠 Bestellung auf anderen Tisch umgebucht _(K-08)_

_(❤️ Hotspot: „Tischumbuchung — zwei Events atomar in einer Transaktion")_

**SRV1 (Jonas):** „Wenn wir 30 Tische haben und ich Tisch 27 suche, will ich nicht durch alle scrollen. Eine Schnellsuche wäre toll." _(K-10)_

_(Ergänzt:)_

- 🟠 Tisch über Schnellsuche gefunden _(K-10)_

**KAS (Eva):** „Und Rückgeld: Der Gast gibt mir 50 Euro, der Tisch kostet 23,50. Klar rechne ich das im Kopf, aber eine Anzeige wäre nett."

**DEV2 (Tim):** „Das ist reine Frontend-Logik — ein Eingabefeld ‚Gegeben', und wir zeigen die Differenz live an. Kein Backend-Aufruf nötig." _(K-09)_

_(Ergänzt:)_

- 🟠 Rückgeld berechnet _(K-09)_

### 2.3 Diskussion: Bons und Ausgabestationen

**VER1 (Klaus):** „Ich stehe an der Getränkeausgabe. Wenn kein Bon kommt, weiß ich nicht, was ich machen soll. Bisher rufen wir einfach — aber bei 15 Tischen ist das Chaos."

**DOM (Rudi):** „Die Getränkeausgabe braucht einen Bon, die Küche braucht einen Bon. Am besten automatisch nach jeder Bestellung."

**SRL (Felix):** „Genau. Wenn eine Bestellung aufgegeben wird, sollen Getränke automatisch an die Getränkeausgabe und Essen an die Küche gehen." _(K-11)_

**ADM (Thomas):** „Wir haben beim letzten Fest einen alten Thermodrucker gefunden. Der könnte funktionieren — aber ich müsste konfigurieren können, welcher Drucker welche Kategorie bekommt."

_(Ergänzte Events:)_

- 🟠 Küchenbon gedruckt _(K-11)_
- 🟠 Getränkebon gedruckt _(K-11)_

**VER1 (Klaus):** „Und was, wenn der Drucker mal ausfällt? Das passiert. Dann brauche ich eine andere Möglichkeit, die Bestellungen zu sehen."

**DEV1 (Anna):** „Dafür könnte ein Küchendisplay dienen — ein Bildschirm in der Küche oder an der Getränkeausgabe, der in Echtzeit die offenen Bestellungen anzeigt. Gruppiert nach Tisch, getrennt nach Kategorie." _(K-12)_

**VER1 (Klaus):** „Das wäre perfekt! Dann sehe ich immer, was noch offen ist — auch wenn ein Bon verloren geht."

_(Ergänzt:)_

- 🟠 Küchenbestellung auf Display angezeigt _(K-12)_

**DOM (Rudi):** „Und wenn Klaus die Getränke fertig hat, muss die Servicekraft das wissen — damit sie abholen kommt."

**VER1 (Klaus):** „Genau. Ich markiere eine Position als ‚in Zubereitung' und dann als ‚fertig'. Dann weiß Jonas, dass er abholen kann." _(K-13)_

**SRV1 (Jonas):** „Das wäre super. Dann muss ich nicht mehr in die Küche laufen und fragen, ob's schon fertig ist."

_(Ergänzt:)_

- 🟠 Position als „in Zubereitung" markiert _(K-13)_
- 🟠 Position als „fertig" markiert _(K-13)_

### 2.4 Diskussion: Stammdaten und Authentifizierung

**FAC (Lisa):** „Gehen wir zur Vorbereitung: Was passiert, bevor das Fest überhaupt losgeht?"

**ADM (Thomas):** „Ich richte das System ein: Produkte anlegen, Preise setzen, Tische erstellen, Benutzer mit Rollen anlegen." _(S-01, S-02, S-03)_

**DOM (Rudi):** „Das machen wir Wochen vorher. Beim Fest selbst kommen manchmal noch Änderungen — ein Tisch wird dazugestellt, ein Produkt ist ausverkauft."

**ADM (Thomas):** „Genau. Und ich muss beides können — vorher einrichten und während des Fests live anpassen."

_(Existierende Events zum Thema Stammdaten:)_

- 🟠 Produkt hinzugefügt _(S-01)_
- 🟠 Produktpreis geändert _(S-01)_
- 🟠 Produkt deaktiviert _(S-01)_
- 🟠 Tisch angelegt _(S-02)_
- 🟠 Tisch deaktiviert _(S-02)_
- 🟠 Benutzer angelegt _(S-03)_

_(Ergänzt:)_

- 🟠 Produktvariante hinzugefügt _(S-01)_
- 🟠 Tisch bearbeitet _(S-02)_
- 🟠 Benutzer bearbeitet _(S-03)_
- 🟠 Benutzer deaktiviert _(S-03)_

**DEV1 (Anna):** „Wichtig: Produkte und Tische werden nie wirklich gelöscht — nur deaktiviert. Soft-Delete, damit historische Bestellungen gültig bleiben."

**SRV2 (Maria):** „Wie melde ich mich eigentlich an? Thomas hat mir neulich ein Passwort gegeben." _(A-01)_

**ADM (Thomas):** „Genau. Ich lege dich als Benutzer an, und du bekommst ein Einmalpasswort. Beim ersten Login musst du dein eigenes Passwort setzen." _(A-02)_

**SRV2 (Maria):** „Das habe ich gemacht. Und am Ende des Abends melde ich mich wieder ab?" _(A-03)_

**FAC (Lisa):** „Genau — Login, Passwort setzen und Logout. Schreiben wir die Events dazu auf."

_(Ergänzte Auth-Events:)_

- 🟠 Benutzer angemeldet _(A-01)_
- 🟠 Passwort gesetzt _(A-02)_
- 🟠 Einmalpasswort ausgestellt _(A-02)_
- 🟠 Benutzer abgemeldet _(A-03)_

### 2.5 Diskussion: Abrechnung und Reporting

**FAC (Lisa):** „Eva, du hast schon einiges zur Abrechnung geschrieben. Was brauchst du am Ende des Tages?"

**KAS (Eva):** „Ich brauche einen Überblick über den Gesamtumsatz — was haben wir eingenommen, wie viel ist storniert, was ist noch offen. Das ist die Tagesabrechnung." _(R-01)_

_(Existierendes Event:)_

- 🟠 Tagesabrechnung erstellt _(R-01)_

**KAS (Eva):** „Und dann brauche ich das aufgeschlüsselt: pro Tisch — was wurde bestellt, was bezahlt, was storniert. Mit einem Gesamt-Saldo pro Tisch." _(R-03)_

_(Ergänzt:)_

- 🟠 Abrechnung pro Tisch erstellt _(R-03)_

**KAS (Eva):** „Außerdem: Wer hat wie viel Umsatz gemacht? Welche Servicekraft hat wie viele Bestellungen aufgenommen, wie viel kassiert? Für Transparenz und Nachvollziehbarkeit." _(R-04)_

_(Ergänzt:)_

- 🟠 Abrechnung pro Servicekraft erstellt _(R-04)_

**DOM (Rudi):** „Und welche Produkte haben sich gut verkauft? Das ist für die Planung des nächsten Fests wichtig." _(R-05)_

**KAS (Eva):** „Genau — ein Ranking der meistverkauften Varianten und Gesamteinnahmen pro Produkt."

_(Ergänzt:)_

- 🟠 Produktumsatz ausgewertet _(R-05)_

**KAS (Eva):** „Und am Ende muss ich die Daten exportieren können — als CSV, für unsere Vereinsbuchhaltung." _(R-02)_

_(Ergänzt:)_

- 🟠 Daten als CSV exportiert _(R-02)_

**DOM (Rudi):** „Was passiert am Ende des Tages? Wir brauchen einen formellen Abschluss — offene Tische prüfen, Abschlussbericht, und dann ist Schluss." _(R-06)_

_(Ergänzt:)_

- 🟠 Tagesabschluss durchgeführt _(R-06)_

**DOM (Rudi):** „Was, wenn noch Tische offen sind? Manchmal verschwinden Gäste, ohne zu zahlen."

**SRL (Felix):** „Dann müssen wir die Tische manuell abschließen — Eva oder Thomas schreiben das als Verlust ab."

_(❤️ Hotspot: „Tagesabschluss mit offenen Tischen — Verhalten unklar")_

### 2.6 Diskussion: Querschnittsthemen

**FAC (Lisa):** „Es gibt noch einige Themen, die quer durch alle Bereiche gehen. Wer möchte anfangen?"

**SRV1 (Jonas):** „Das Wichtigste für mich: Das muss auf dem Handy laufen. Wir benutzen unsere eigenen Smartphones — kein Tablet, kein PC. Alles muss auf einem kleinen Bildschirm funktionieren." _(Q-01)_

**SRV2 (Maria):** „Und ohne App-Download! Ich will einfach den Browser öffnen."

_(🟠 Mobile-first und Touch-optimiert — Q-01)_

**SRL (Felix):** „Und wenn drei Servicekräfte gleichzeitig am System arbeiten — das muss parallel funktionieren, ohne dass sich jemand in die Quere kommt." _(Q-02)_

**DEV1 (Anna):** „Mehrbenutzerfähigkeit ist zentral. Parallele Bestellungen an verschiedenen Tischen müssen unabhängig funktionieren. Und selbst am selben Tisch darf es keine Konflikte geben."

_(🟠 Mehrere Benutzer arbeiten gleichzeitig — Q-02)_

**DEV2 (Tim):** „Wichtig ist auch Validierung auf beiden Seiten. Das Frontend zeigt sofort Fehler an — aber das Backend prüft nochmal unabhängig. Wir vertrauen keinen Daten vom Client." _(Q-03)_

_(🟠 Doppelte Validierung — Frontend und Backend — Q-03)_

**DEV1 (Anna):** „Und Datenintegrität: Geldbeträge sind immer in Cent — keine Fließkommazahlen, keine Rundungsfehler. Das Kassenjournal ist append-only: Events werden nie geändert oder gelöscht." _(Q-04)_

**KAS (Eva):** „Als Kassenwart sage ich: genau richtig. Floats bei Geld sind eine Katastrophe."

_(🟠 Geldbeträge in Cent, Events unveränderlich — Q-04)_

**SRV1 (Jonas):** „Was ist bei Netzwerkausfall? Beim letzten Fest war das WLAN eine halbe Stunde weg. Dann geht gar nichts mehr." _(Q-05)_

**DOM (Rudi):** „Das ist schon zweimal passiert bei uns."

**DEV1 (Anna):** „Offline-Fähigkeit ist komplex — Service Worker, lokaler Speicher, Synchronisation, Konfliktauflösung. Das wäre ein erheblicher Zusatzaufwand."

**FAC (Lisa):** „Das klingt nach einem Hotspot — wichtig, aber nicht trivial."

_(❤️ Hotspot: „Offline-Fähigkeit — wünschenswert, aber aufwändig" — Q-05)_

**ADM (Thomas):** „Sicherheit: Die Verbindung muss verschlüsselt sein. HTTPS ist Pflicht." _(Q-06)_

**DEV2 (Tim):** „Das läuft über den Reverse Proxy mit Let's Encrypt — automatische Zertifikatserneuerung."

_(🟠 HTTPS / TLS — Q-06)_

**DEV1 (Anna):** „Und der Login-Endpunkt braucht Rate Limiting — sonst kann jemand Passwörter durchprobieren." _(Q-07)_

_(🟠 Rate Limiting auf Login — Q-07)_

**DEV2 (Tim):** „Außerdem Security Headers: CSP, X-Frame-Options, HSTS — das Übliche." _(Q-08)_

_(🟠 Security Headers — Q-08)_

### 2.7 Bewusste Abgrenzung

**FAC (Lisa):** „Bevor wir Phase 1 abschließen: Was gehört bewusst NICHT in unser System? Grenzen zu definieren ist genauso wichtig wie Features."

**DOM (Rudi):** „Kartenzahlung brauchen wir nicht. Auf unseren Festen wird bar bezahlt — das ist bei Vereinen üblich. Kartenterminals kosten Geld und machen Probleme."

**KAS (Eva):** „Und TSE oder KassenSichV — gemeinnützige Vereine sind davon in der Regel ausgenommen. Event-Sourcing gibt uns ohnehin eine lückenlose, unveränderliche Protokollierung."

**SRV1 (Jonas):** „Reservierungen? Bei uns setzt man sich, wo Platz ist."

**DOM (Rudi):** „Warenwirtschaft auch nicht. Wir kaufen für zwei Tage ein, das war's. Keine Lagerverwaltung nötig."

**ADM (Thomas):** „Multi-Standort brauchen wir nicht. Jedes Fest hat seine eigene Instanz."

**SRL (Felix):** „Kiosk-Modus oder Self-Order? Nein. Unsere Helfer bedienen persönlich — das gehört zum Vereinsleben."

**DEV1 (Anna):** „Lieferservice und CRM können wir streichen. Kein Verein braucht eine Kundendatenbank für sein Sommerfest."

**DOM (Rudi):** „Und Gäste sollen keine Benachrichtigungen bekommen. Die sitzen am Tisch und werden persönlich bedient."

**DEV2 (Tim):** „Trinkgeld-Tracking? Bei ehrenamtlichen Veranstaltungen unnötig."

**FAC (Lisa):** „Gut. Halten wir fest, was wir bewusst NICHT bauen: Kartenzahlung, TSE/KassenSichV, Reservierungen, Warenwirtschaft, Lieferservice, Multi-Standort, CRM, Kiosk-Modus, Gast-Benachrichtigungen, Trinkgeld-Tracking. Jedes Feature, das wir nicht bauen, ist weniger Komplexität für unser ehrenamtliches Team."

### 2.8 Technische vs. fachliche Events

**FAC (Lisa):** „Mir fällt auf, dass einige Events an der Wand eher technischer Natur sind — ‚JWT ausgestellt', ‚Event gespeichert', ‚Seite geladen'. Gehören die hierher?"

**DEV1 (Anna):** „Guter Punkt. Wir sollten Domain Events von technischen Events trennen. Domain Events sind fachliche Geschehnisse — ‚Bestellung aufgegeben', ‚Zahlung registriert'. Das sind Dinge, die Rudi und Eva verstehen."

**DEV2 (Tim):** „‚JWT ausgestellt' oder ‚Session abgelaufen' sind Infrastruktur-Events. Die sind wichtig fürs System, aber sie haben keine fachliche Bedeutung."

**DEV1 (Anna):** „Genau. Login und Logout sind aus Sicht der Domäne relevante Vorgänge — der Benutzer meldet sich an, der Benutzer meldet sich ab. Aber das ‚Wie' — JWT, Sessions, Rate Limiting — ist reine Technik. Auth-Events gehören zur Infrastruktur, nicht zu unserer Domänen-Event-Wand."

**DOM (Rudi):** „Also ‚Benutzer angemeldet' ist ein fachliches Event, aber ‚JWT ausgestellt' nicht?"

**DEV1 (Anna):** „Richtig. ‚Benutzer angemeldet' beschreibt, was aus Domänensicht passiert ist. ‚JWT ausgestellt' ist ein Implementierungsdetail — das vermischen wir nicht mit den Domänen-Events."

**FAC (Lisa):** „Dann legen wir die technischen Events zur Seite. Wir notieren sie, aber sie formen nicht unsere Domain Map."

_(Technische Events — separat gestellt:)_

- JWT ausgestellt · Session abgelaufen · Rate-Limit überschritten · Snapshot erstellt · Event gespeichert · Seite geladen · Saldo aktualisiert

---

## 3. Phase 2 — Clustering und Pivot Points

**Zeitrahmen:** 11:30–12:30 Uhr (60 Minuten, inkl. Kaffeepause 12:00–12:15)

### 3.1 Timeline erstellen

**FAC (Lisa):** „Super Arbeit in Phase 1! Jetzt bringen wir Ordnung ins Chaos. Verschiebt die Haftnotizen entlang einer Zeitachse — von links nach rechts: Was passiert zuerst, was danach? Es muss nicht perfekt sein, grob reicht. Und achtet darauf, welche Events zusammengehören — die schieben wir zu Gruppen zusammen."

_(Alle stehen auf und beginnen, die Haftnotizen an der Wand zu verschieben. Es wird diskutiert und sortiert.)_

**DOM (Rudi):** „Ganz links gehört die Vorbereitung — Produkte eingeben, Tische anlegen, Accounts erstellen. Das machen wir Tage oder Wochen vor dem Fest."

**ADM (Thomas):** „Genau. Das ist der Admin-Bereich. Wenn das Fest losgeht, ist das Setup abgeschlossen — wobei ich während des Betriebs manchmal noch Korrekturen mache."

**SRV1 (Jonas):** „In der Mitte kommt der eigentliche Betrieb — anmelden, Bestellungen aufnehmen, liefern, kassieren. Das ist der Kern."

**KAS (Eva):** „Und ganz rechts die Abrechnung. Wenn der letzte Gast gegangen ist, mache ich den Kassensturz."

**FAC (Lisa):** „Perfekt — ich sehe drei klare Phasen auf der Zeitachse."

_Nach 15 Minuten ergibt sich folgende grobe Zeitlinie:_

```
[VORBEREITUNG]  →  [BETRIEB]  →  [ABSCHLUSS]
Setup & Stamm-     Kassen-        Abrechnung &
daten              betrieb        Reporting
```

**DEV1 (Anna):** „Was mir auffällt: Die Vorbereitung und der Abschluss sind jeweils einmalig, aber der Betrieb in der Mitte besteht aus vielen parallelen Zyklen — jeder Tisch durchläuft seinen eigenen Ablauf unabhängig."

**DOM (Rudi):** „Genau. Tisch 5 kann schon abkassiert sein, während Tisch 12 gerade erst bestellt."

### 3.2 Pivot Points identifizieren

**FAC (Lisa):** „Jetzt suchen wir Pivot Points — das sind Schlüsselereignisse, nach denen sich der Ablauf grundlegend ändert. Markiert sie mit einem dickeren Strich auf der Timeline."

**DOM (Rudi):** „Das Offensichtlichste: ‚Veranstaltung eröffnet'. Vorher ist alles Vorbereitung, danach beginnt der Betrieb."

**SRV1 (Jonas):** „Für mich am Tisch ist es ‚Bestellung aufgegeben'. Ab da hat der Tisch einen Saldo, und alles Weitere — Lieferung, Zahlung, Stornierung — bezieht sich darauf."

**KAS (Eva):** „‚Tischkonto ausgeglichen' ist auch ein Wendepunkt. Wenn der Saldo null ist, ist der Tisch fertig — keine offenen Posten mehr."

**SRL (Felix):** „Und am Ende ‚Tagesabschluss durchgeführt'. Danach wird nicht mehr kassiert. Dann übernimmt Eva."

_Gemeinsam identifiziert:_

| Pivot Point                     | Beschreibung                                                      |
| ------------------------------- | ----------------------------------------------------------------- |
| **Veranstaltung eröffnet**      | System wechselt von Vorbereitung in aktiven Betrieb               |
| **Bestellung aufgegeben**       | Ab hier entsteht ein Saldo; das Kassenjournal des Tisches beginnt |
| **Tischkonto ausgeglichen**     | Tisch ist abgeschlossen; kein offener Saldo mehr                  |
| **Tagesabschluss durchgeführt** | Betrieb endet; Abrechnung und Auswertung beginnen                 |

**DEV1 (Anna):** „Interessant: ‚Bestellung aufgegeben' ist ein Pivot pro Tisch — nicht für das Gesamtsystem. Jeder Tisch hat seinen eigenen Lebenszyklus."

**DEV2 (Tim):** „Und ‚Tischkonto ausgeglichen' ist das logische Ende dieses Zyklus. Dazwischen können beliebig viele Bestellungen, Zahlungen und Lieferungen liegen."

### 3.3 Domänen-Cluster entstehen

**FAC (Lisa):** „Jetzt gruppieren wir die Events zu Clustern. Welche Events gehören fachlich zusammen — weil sie denselben Bereich betreffen oder von denselben Akteuren stammen?"

_(Die Teilnehmer schieben Events auf der Wand zusammen und diskutieren die Gruppierung.)_

#### Cluster A — Stammdaten & Auth

_(vor dem Fest und begleitend, Admin-Bereich)_

**ADM (Thomas):** „Alles, was ich als Admin einrichte, gehört zusammen: Produkte, Tische, Benutzer. Und Auth — Login, Passwort setzen, Logout — gehört auch hierher."

**DEV1 (Anna):** „Auth ist streng genommen Infrastruktur, nicht Fachdomäne. Aber für die Übersicht auf der Wand macht es Sinn, es beim Setup zu lassen."

- 🟠 Produkt hinzugefügt _(S-01)_
- 🟠 Produktvariante hinzugefügt _(S-01)_
- 🟠 Produktpreis geändert _(S-01)_
- 🟠 Produkt deaktiviert _(S-01)_
- 🟠 Tisch angelegt _(S-02)_
- 🟠 Tisch bearbeitet _(S-02)_
- 🟠 Tisch deaktiviert _(S-02)_
- 🟠 Benutzer angelegt _(S-03)_
- 🟠 Benutzer bearbeitet _(S-03)_
- 🟠 Benutzer deaktiviert _(S-03)_
- 🟠 Einmalpasswort ausgestellt _(A-02)_
- 🟠 Passwort gesetzt _(A-02)_
- 🟠 Benutzer angemeldet _(A-01)_
- 🟠 Benutzer abgemeldet _(A-03)_

#### Cluster B — Kassenbetrieb

_(während des Fests, Service-Bereich)_

**SRV1 (Jonas):** „Der Kern: Bestellen, Liefern, Bezahlen. Plus die Komfort-Sachen — Bestellung benennen, Schnellsuche, Rückgeld."

**SRV2 (Maria):** „Und die Tischübersicht — ohne die bin ich aufgeschmissen."

**SRL (Felix):** „Kassenjournal auch. Das gehört zum laufenden Betrieb — jederzeit nachschauen, was am Tisch passiert ist."

- 🟠 Tisch ausgewählt _(K-05)_
- 🟠 Tisch über Schnellsuche gefunden _(K-10)_
- 🟠 Bestellung aufgegeben _(K-01)_
- 🟠 Bestellung benannt _(K-07)_
- 🟠 Bestellung abgebrochen _(K-01)_
- 🟠 Produkte ausgeliefert _(K-03)_
- 🟠 Zahlung registriert _(K-02)_
- 🟠 Teilzahlung registriert _(K-02)_
- 🟠 Rückgeld berechnet _(K-09)_
- 🟠 Tischkonto ausgeglichen _(K-02)_
- 🟠 Kassenjournal eingesehen _(K-06)_

#### Cluster C — Stornierung & Umbuchung

_(Service-Bereich, erhöhte Berechtigung)_

**SRL (Felix):** „Stornierung und Umbuchung sind Korrektur-Vorgänge. Die stehen am Rand des normalen Betriebs — man hofft, dass man sie nicht braucht, aber wenn doch, braucht man sie sofort."

**DOM (Rudi):** „Und beides darf nicht jeder. Nur Felix oder Thomas."

- 🟠 Stornierung durchgeführt _(K-04)_
- 🟠 Bestellung auf anderen Tisch umgebucht ❤️ _(K-08)_

#### Cluster D — Ausgabestationen & Bons

_(Küche, Getränkeausgabe, begleitend zum Betrieb)_

**VER1 (Klaus):** „Alles, was bei mir an der Ausgabe ankommt: Bons, Display-Anzeige, Zubereitungsstatus. Das ist mein Arbeitsbereich."

**DEV1 (Anna):** „Das sind eigentlich Reaktionen auf die Bestellungen aus Cluster B — wenn eine Bestellung aufgegeben wird, muss die Info an die richtige Ausgabestation fließen."

- 🟠 Küchenbon gedruckt _(K-11)_
- 🟠 Getränkebon gedruckt _(K-11)_
- 🟠 Küchenbestellung auf Display angezeigt _(K-12)_
- 🟠 Position als „in Zubereitung" markiert _(K-13)_
- 🟠 Position als „fertig" markiert _(K-13)_

#### Cluster E — Abrechnung & Reporting

_(Tagesende und begleitend, Admin/Kassenwart)_

**KAS (Eva):** „Mein Bereich: Tagesabrechnung, Aufschlüsselungen, Export. Das brauche ich teils während des Betriebs zum Kontrollieren, vor allem aber am Ende."

**DOM (Rudi):** „Und der Tagesabschluss als formeller Schlussstrich."

- 🟠 Tagesabrechnung erstellt _(R-01)_
- 🟠 Abrechnung pro Tisch erstellt _(R-03)_
- 🟠 Abrechnung pro Servicekraft erstellt _(R-04)_
- 🟠 Produktumsatz ausgewertet _(R-05)_
- 🟠 Daten als CSV exportiert _(R-02)_
- 🟠 Tagesabschluss durchgeführt _(R-06)_

---

**FAC (Lisa):** „Fünf Cluster — Stammdaten, Kassenbetrieb, Korrekturen, Ausgabestationen und Abrechnung. Damit haben wir eine klare Landkarte unserer Domäne. In der nächsten Phase schauen wir uns jeden Cluster im Detail an: Wer löst was aus, welche Regeln gelten, und was passiert automatisch."

---

## 4. Phase 3 — Process Modelling: Commands, Akteure, Policies

**Zeitrahmen:** 12:30–14:00 Uhr (90 Minuten, inkl. Mittagspause 13:00–13:15)

**FAC (Lisa):** „Willkommen zurück! In Phase 1 haben wir Events gesammelt, in Phase 2 haben wir sie geordnet und geclustert. Jetzt gehen wir ins Detail: Für jeden Cluster modellieren wir die Prozesse. Wir fügen drei neue Elemente hinzu:"

- 🔵 **Command** (blau) — eine Absicht, etwas zu tun. „Bestellung aufgeben", „Zahlung registrieren". Commands stehen vor dem Event.
- 👤 **Akteur** — wer löst den Command aus? Servicekraft, Serviceleitung, Admin?
- 🟣 **Policy** (lila) — eine automatische Reaktion. „Wenn Bestellung aufgegeben, dann Bon drucken."

**FAC (Lisa):** „Die Notation funktioniert so: Links steht der Akteur mit seinem Command (blau), dann folgt das Aggregate — die transaktionale Grenze, die die Geschäftsregeln prüft — und rechts das resultierende Event (orange). Policies verbinden Events mit nachfolgenden Commands. Gehen wir Cluster für Cluster durch."

### 4.1 Kassenbetrieb — Bestellung bis Zahlung

#### Bestellung aufnehmen (K-01, K-07)

**FAC (Lisa):** „Fangen wir mit dem Kernprozess an: Eine Bestellung wird aufgegeben. Jonas, beschreib mal deinen typischen Ablauf."

**SRV1 (Jonas):** „Ich gehe zum Tisch, tippe drauf, wähle Produkte aus dem Katalog — nach Kategorie gegliedert: Essen, Getränke, Sonstiges. Ich stelle die Menge ein und gebe die Bestellung ab."

**FAC (Lisa):** „Was ist der Command?"

**DEV1 (Anna):** „Der Command ist ‚Bestellung aufgeben'. Der Akteur ist die Servicekraft — oder die Serviceleitung, oder der Admin. Alle drei dürfen bestellen."

```
👤 Servicekraft / Serviceleitung / Admin
    🔵 Bestellung aufgeben
        → [Aggregate prüft: mindestens eine Position? Preise gültig?]
            → 🟠 Bestellung aufgegeben
```

**SRV2 (Maria):** „Was sind die Regeln? Was kann schiefgehen?"

**DEV2 (Tim):** „Das Aggregate prüft: Mindestens eine Position muss enthalten sein. Jede Position ist eine Produktvariante plus Menge. Die Preise werden zum Zeitpunkt der Bestellung festgehalten — als Cent-Beträge, keine Floats."

**DEV1 (Anna):** „Und das Event ‚Bestellung aufgegeben' enthält alle Informationen — Produkt, Variante, Menge, Preis. Ein sogenanntes Fat Event. Selbst wenn sich der Produktpreis später ändert, bleibt die historische Bestellung korrekt."

**DOM (Rudi):** „Und der optionale Kommentar? Manchmal schreibt die Servicekraft ‚ohne Zwiebeln' dazu."

**DEV2 (Tim):** „Der Kommentar ist optional, maximal 100 Zeichen, und wird im Event mit gespeichert."

**SRV2 (Maria):** „Und wenn ich mich vertippe und die Bestellung abbrechen will, bevor ich sie abschicke?"

**SRV1 (Jonas):** „Das ist Frontend — du verwirfst einfach deine Eingabe. Da entsteht kein Event."

**DEV2 (Tim):** „Genau. Solange der Command nicht abgeschickt wird, passiert nichts im Backend. Kein Event, kein Eintrag im Journal."

#### Lieferung bestätigen (K-03)

**FAC (Lisa):** „Die Bestellung ist aufgegeben. Was passiert als Nächstes?"

**SRV1 (Jonas):** „Ich gehe zur Ausgabe, hole die Sachen und bringe sie zum Tisch. Danach markiere ich die Positionen als geliefert."

**FAC (Lisa):** „Das ist ein eigener Command?"

**DEV1 (Anna):** „Ja — ‚Lieferung bestätigen'. Der Akteur ist wieder Servicekraft, Serviceleitung oder Admin."

```
👤 Servicekraft / Serviceleitung / Admin
    🔵 Lieferung bestätigen
        Eingabe: Auswahl ungelieferter Positionen, opt. Kommentar
        → [Aggregate: min. 1 ungelieferte Position ausgewählt]
            → 🟠 Produkte geliefert
```

**SRV2 (Maria):** „Muss ich alles auf einmal liefern?"

**SRV1 (Jonas):** „Nein. Wenn die Getränke schneller da sind als das Essen, liefere ich die Getränke und markiere nur die. Das Essen kommt später."

**DEV2 (Tim):** „Jede Lieferung erzeugt ein eigenes Event. Es kann also mehrere ‚Produkte geliefert'-Events pro Tisch geben."

**DOM (Rudi):** „Und der Zustand ‚alles geliefert' — ergibt der sich dann automatisch?"

**DEV1 (Anna):** „Genau. Der Tischzustand wird aus allen Events berechnet: Wenn für jede bestellte Position ein Liefer-Event existiert, sind alle Positionen geliefert. Es gibt kein separates ‚Tisch fertig'-Event."

#### Zahlung registrieren (K-02, K-09)

**FAC (Lisa):** „Kommen wir zum Bezahlen."

**KAS (Eva):** „Das ist der kritischste Moment. Der Gast will zahlen, und die Servicekraft muss schnell und fehlerfrei kassieren."

**SRV1 (Jonas):** „Ich tippe auf den Tisch, sehe die offenen Positionen, wähle aus, was der Gast bezahlen will, und registriere die Zahlung."

**KAS (Eva):** „Und Teilzahlung muss gehen! Drei Leute am Tisch, jeder zahlt seinen Teil."

**DEV1 (Anna):** „Der Command ist ‚Zahlung registrieren'. Die Servicekraft wählt unbezahlte Positionen aus — eine, mehrere oder alle. Die Summe ergibt sich aus den gewählten Positionen."

```
👤 Servicekraft / Serviceleitung / Admin
    🔵 Zahlung registrieren
        Eingabe: Auswahl unbezahlter Positionen, opt. Kommentar
        → [Aggregate: min. 1 unbezahlte Position, Betrag in Cent]
            → 🟠 Zahlung registriert (mit Betrag, Positionen)
```

**SRV2 (Maria):** „Und wenn der Gast einen 50-Euro-Schein gibt?"

**SRV1 (Jonas):** „Dann muss ich Rückgeld rausgeben. Wäre praktisch, wenn das System mir anzeigt, wie viel."

**DEV2 (Tim):** „Das ist die Rückgeldberechnung. Ein Eingabefeld ‚Gegeben' — die Servicekraft tippt den erhaltenen Betrag ein, und das Frontend zeigt die Differenz live an. Das ist reine UI-Logik, kein Backend-Aufruf." _(K-09)_

**DEV1 (Anna):** „Wichtig: Der ‚Gegeben'-Betrag wird **nicht** im Event gespeichert. Das Event enthält nur den Zahlungsbetrag — die Summe der gewählten Positionen. Die Rückgeldanzeige ist eine Hilfe für die Servicekraft, mehr nicht."

**KAS (Eva):** „Und wann ist der Tisch fertig?"

**DEV2 (Tim):** „Wenn der Saldo null ist — also wenn die Summe aller Zahlungen die Summe aller Bestellungen abzüglich Stornierungen abdeckt. Dann ist ‚Tischkonto ausgeglichen' erreicht. Das ist kein eigenes Event, sondern ein abgeleiteter Zustand."

**DOM (Rudi):** „Gibt es eine Situation, wo mehr gezahlt wird als bestellt?"

**DEV1 (Anna):** „Nein. Die Zahlung bezieht sich immer auf konkrete Positionen — man kann nicht mehr bezahlen als bestellt. Der Betrag ergibt sich aus der Auswahl."

#### Stornierung (K-04)

**FAC (Lisa):** „Jetzt wird es interessant — die Stornierung. Hier gelten besondere Regeln."

**SRL (Felix):** „Wenn eine Servicekraft was Falsches bestellt — zum Beispiel drei Bier statt drei Cola — dann muss ich das korrigieren. Ich als Serviceleitung oder Thomas als Admin stornieren die falschen Positionen."

**SRV2 (Maria):** „Und ich als einfache Servicekraft kann das nicht?"

**SRL (Felix):** „Nein. Das ist Absicht. Beim letzten Fest hat jemand versehentlich einen ganzen Tisch storniert. Seitdem machen das nur Leute mit Erfahrung."

**DEV1 (Anna):** „Das ist eine klare Policy: Stornierung nur für Serviceleitung und Admin."

```
👤 Serviceleitung / Admin
    🔵 Positionen stornieren
        Eingabe: Auswahl von Positionen, opt. Kommentar
        → [Aggregate: min. 1 Position ausgewählt, Akteur hat Berechtigung]
            → 🟠 Positionen storniert (mit Positionen, Betrag)
```

🟣 **Policy:** Stornierung nur durch Rollen `serviceleitung` und `admin`. Rolle `service` hat keinen Zugriff.

**DOM (Rudi):** „Kann man auch schon bezahlte Positionen stornieren?"

**DEV2 (Tim):** „Das ist eine gute Frage. Fachlich: Wenn ein Gast ein Essen reklamiert, das er schon bezahlt hat, müsste man es stornieren und den Betrag gutschreiben."

**SRL (Felix):** „In der Praxis kommt das selten vor — aber es muss möglich sein."

**DEV1 (Anna):** „Die Stornierung reduziert den Saldo des Tisches. Wenn der Gast bereits bezahlt hat, entsteht ein negativer Teilbetrag — der könnte mit der nächsten Bestellung verrechnet oder manuell ausgeglichen werden."

**KAS (Eva):** „Für die Abrechnung ist vor allem wichtig, dass die Stornierung im Kassenjournal sichtbar ist — wer hat wann was storniert und warum."

**DEV2 (Tim):** „Das Event enthält die stornierten Positionen, den Stornobetrag, den Akteur und den optionalen Kommentar. Alles unveränderlich im Journal."

### 4.2 Stammdaten-Verwaltung

#### Produktverwaltung (S-01)

**FAC (Lisa):** „Wechseln wir zu den Stammdaten. Thomas, wie läuft die Produktpflege?"

**ADM (Thomas):** „Ich lege Produkte an — zum Beispiel ‚Bratwurst' in der Kategorie Essen. Dazu kommen Varianten: ‚Bratwurst im Brötchen' für 3,50 €, ‚Bratwurst ohne Brötchen' für 2,50 €. Jede Variante hat ihren eigenen Preis."

**DEV1 (Anna):** „Das ist klassisches CRUD — kein Event-Sourcing nötig. Produkte sind Stammdaten, keine Transaktionen."

```
👤 Admin
    🔵 Produkt anlegen
        Eingabe: Name, Kategorie (Essen / Getränke / Sonstiges)
        → 🟠 Produkt hinzugefügt

    🔵 Produktvariante anlegen
        Eingabe: Produktvariante mit Name und Preis (in Cent)
        → 🟠 Produktvariante hinzugefügt

    🔵 Produkt bearbeiten
        Eingabe: Name, Kategorie
        → 🟠 Produkt bearbeitet

    🔵 Produktpreis ändern
        Eingabe: Neuer Preis (in Cent) für Variante
        → 🟠 Produktpreis geändert

    🔵 Produkt deaktivieren
        → 🟠 Produkt deaktiviert (Soft-Delete)
```

**DOM (Rudi):** „Und wenn ich ein Produkt lösche, das schon bestellt wurde?"

**DEV1 (Anna):** „Deshalb Soft-Delete. Das Produkt wird auf ‚deaktiviert' gesetzt und verschwindet aus dem Katalog für die Servicekräfte. Aber alle historischen Bestellungen bleiben gültig, weil die Events die Produktdaten zum Zeitpunkt der Bestellung enthalten."

**DEV2 (Tim):** „Das ist der Vorteil von Fat Events — die Bestellung weiß selbst, was bestellt wurde und zu welchem Preis. Egal was danach mit dem Produktkatalog passiert."

**ADM (Thomas):** „Und Varianten kann ich auch einzeln deaktivieren? Wenn zum Beispiel die großen Brezeln aus sind?"

**DEV1 (Anna):** „Ja. Varianten haben eigene Status — aktiv oder inaktiv. Inaktive Varianten erscheinen nicht mehr im Service-Katalog."

#### Tischverwaltung (S-02)

**FAC (Lisa):** „Wie sieht es mit den Tischen aus?"

**ADM (Thomas):** „Ich lege Tische an — ‚Tisch 1' bis ‚Tisch 30', oder auch ‚Stehtisch Eingang'. Jeder Tisch hat einen Namen. Während des Fests kann ich auch noch Tische dazunehmen oder deaktivieren."

**DOM (Rudi):** „Letztes Jahr mussten wir spontan fünf Tische dazustellen — das muss auch im laufenden Betrieb gehen."

```
👤 Admin
    🔵 Tisch anlegen
        Eingabe: Name
        → 🟠 Tisch angelegt

    🔵 Tisch bearbeiten
        Eingabe: Neuer Name
        → 🟠 Tisch bearbeitet

    🔵 Tisch deaktivieren
        → 🟠 Tisch deaktiviert (Soft-Delete)
```

**DEV2 (Tim):** „Auch hier Soft-Delete. Ein deaktivierter Tisch verschwindet aus der Tischübersicht, aber seine Events bleiben im System."

**SRV1 (Jonas):** „Kann ein Tisch deaktiviert werden, an dem noch offene Bestellungen hängen?"

**DEV1 (Anna):** „Gute Frage. Fachlich sollte man erst alle offenen Posten klären, bevor ein Tisch deaktiviert wird. Im System wäre das eine Warnung oder eine Sperre — das klären wir im Detail-Design."

#### Benutzerverwaltung (S-03)

**FAC (Lisa):** „Und die Benutzerverwaltung — Thomas, wie legst du die Helfer an?"

**ADM (Thomas):** „Ich erstelle für jeden Helfer einen Account: Name, Benutzername und Rolle. Es gibt drei Rollen — Admin, Serviceleitung und Servicekraft. Jeder bekommt ein Einmalpasswort, das er beim ersten Login durch ein eigenes Passwort ersetzt."

**SRV2 (Maria):** „Genau, so hab ich das gemacht. Thomas hat mir das Einmalpasswort gegeben, ich hab mich angemeldet und mein eigenes Passwort gesetzt."

```
👤 Admin
    🔵 Benutzer anlegen
        Eingabe: Name, Benutzername, Rolle (admin / serviceleitung / service)
        → 🟠 Benutzer angelegt
        → 🟠 Einmalpasswort ausgestellt

    🔵 Benutzer bearbeiten
        Eingabe: Name, Benutzername, Rolle
        → 🟠 Benutzer bearbeitet

    🔵 Benutzer deaktivieren
        → 🟠 Benutzer deaktiviert (Soft-Delete)

    🔵 Passwort zurücksetzen
        → 🟠 Einmalpasswort ausgestellt
```

**SRL (Felix):** „Und wenn jemand sein Passwort vergisst?"

**ADM (Thomas):** „Dann setze ich es zurück. Der Benutzer bekommt ein neues Einmalpasswort und muss beim nächsten Login wieder ein eigenes setzen."

**DEV1 (Anna):** „Die Passwörter werden mit Argon2id gehasht — das ist der aktuelle Standard für sichere Passwort-Hashes. Aber das ist ein Implementierungsdetail, kein fachliches Event."

**DEV2 (Tim):** „Und deaktivierte Benutzer können sich nicht mehr anmelden. Auch hier: Soft-Delete — der Account bleibt im System, ist aber gesperrt."

### 4.3 Abrechnung und Reporting

#### Tagesabrechnung / Umsatzübersicht (R-01)

**FAC (Lisa):** „Eva, du bist die Kassenwart-Expertin. Was brauchst du für die Abrechnung?"

**KAS (Eva):** „Am Ende des Tages — oder auch zwischendurch — will ich einen Gesamtüberblick: Wie viel haben wir eingenommen? Wie viel ist storniert? Wie viel ist noch offen? Das ist die Tagesabrechnung."

**DOM (Rudi):** „Früher saßen wir bis Mitternacht an den Zetteln. Mit einem System sollte das auf einen Klick gehen."

**DEV1 (Anna):** „Die Tagesabrechnung ist ein Read Model — eine Auswertung, die aus den vorhandenen Events berechnet wird. Kein neues Event nötig."

```
👤 Admin
    🔵 Tagesabrechnung anzeigen
        → 🟢 Read Model: Gesamtumsatz, Summe Stornierungen,
           offene Beträge, Umsatz pro Servicekraft (Übersicht)
```

**KAS (Eva):** „Das muss jederzeit abrufbar sein — nicht erst beim Tagesabschluss. Wenn ich zwischendurch schauen will, wie es läuft, muss das gehen."

**DEV2 (Tim):** „Genau. Die Daten liegen in den Events — wir aggregieren sie bei Bedarf. Kein Schreibvorgang, nur Lesen."

#### Datenexport (R-02)

**KAS (Eva):** „Und ich muss die Daten exportieren können. Für die Vereinsbuchhaltung brauche ich eine CSV-Datei — Umsätze, Bestellungen, Artikeldaten."

**ADM (Thomas):** „Das klingt nach einem Download-Button."

**DEV2 (Tim):** „Genau. Der Admin klickt auf ‚Exportieren', das Backend sammelt die Daten, formatiert sie als CSV und liefert sie zum Download."

```
👤 Admin
    🔵 Daten exportieren
        → Backend aggregiert Events, generiert CSV
        → 🟠 Daten als CSV exportiert
```

**DEV1 (Anna):** „Hier könnte man diskutieren, ob ‚Daten exportiert' ein fachliches Event ist. Es verändert keinen Domänenzustand — es ist eher eine Aktion, die Daten bereitstellt."

**KAS (Eva):** „Mir ist wichtig, dass es funktioniert. Ob dahinter ein Event steht oder nicht, ist mir egal."

**FAC (Lisa):** „Halten wir fest: Der Export ist ein Command, der Daten bereitstellt. Ob wir ein Event dafür speichern, klären wir im Design."

#### Abrechnung pro Tisch (R-03)

**KAS (Eva):** „Neben dem Gesamtüberblick brauche ich eine Detailansicht pro Tisch: Was wurde bestellt, was geliefert, was bezahlt, was storniert. Chronologisch, mit einem Saldo am Ende."

**SRL (Felix):** „Das ist quasi das Kassenjournal, aber aufbereitet für die Abrechnung?"

**DEV1 (Anna):** „Genau. Das Kassenjournal zeigt die rohen Events. Die Tisch-Abrechnung ist ein aufbereitetes Read Model: Sie fasst die Events zusammen und zeigt Bestellt, Bezahlt, Offen und Storniert als Saldo."

```
👤 Admin
    🔵 Abrechnung pro Tisch anzeigen
        Eingabe: Tisch-ID
        → 🟢 Read Model: Bestellungen, Zahlungen, Lieferungen,
           Stornierungen des Tisches, Gesamt-Saldo
```

**KAS (Eva):** „Wenn ein Tisch am Ende noch einen offenen Betrag hat, muss ich das sofort sehen."

**DEV2 (Tim):** „Der Saldo wird aus den Events berechnet: Summe Bestellungen minus Summe Zahlungen minus Summe Stornierungen. Alles in Cent, keine Rundungsfehler."

#### Abrechnung pro Servicekraft (R-04)

**KAS (Eva):** „Ich brauche auch eine Aufschlüsselung pro Servicekraft: Wer hat wie viel Umsatz gemacht, wie viele Bestellungen aufgenommen, wie viel storniert."

**DOM (Rudi):** „Das ist nicht für die Kontrolle gedacht, sondern für Transparenz. Wenn am Ende die Kasse nicht stimmt, will ich nachvollziehen können, wer was gemacht hat."

**SRL (Felix):** „Und für mich als Serviceleitung ist es nützlich zu sehen, wie die Arbeit verteilt war."

**DEV1 (Anna):** „Jedes Event im Kassenjournal enthält den Akteur — wer hat den Command ausgelöst. Das ist bereits in den Events gespeichert. Die Servicekraft-Abrechnung gruppiert einfach nach Akteur."

```
👤 Admin
    🔵 Abrechnung pro Servicekraft anzeigen
        → 🟢 Read Model: Umsatz, Anzahl Bestellungen,
           Stornierungen pro Servicekraft
```

**KAS (Eva):** „Nur der Admin sieht das, richtig?"

**DEV2 (Tim):** „Richtig. Personenbezogene Auswertungen sind nur für den Admin zugänglich."

#### Produktumsatz-Reporting (R-05)

**DOM (Rudi):** „Was sich gut verkauft hat, ist für die Planung des nächsten Fests Gold wert. Wie viele Bratwürste, wie viel Bier — und was ist liegengeblieben."

**KAS (Eva):** „Genau. Verkaufte Mengen pro Produkt und Variante, ein Ranking der Bestseller und Gesamteinnahmen pro Produkt."

**DEV1 (Anna):** „Auch das ist ein Read Model. Die Bestellungs-Events enthalten Produkt, Variante, Menge und Preis. Wir aggregieren über alle Tische."

```
👤 Admin
    🔵 Produktumsatz anzeigen
        → 🟢 Read Model: Verkaufte Menge pro Produkt/Variante,
           Ranking meistverkaufte Varianten,
           Gesamteinnahmen pro Produkt
```

**DOM (Rudi):** „Müssen Stornierungen berücksichtigt werden?"

**DEV2 (Tim):** „Ja. Stornierte Positionen werden abgezogen — sonst stimmt die Bilanz nicht. Wenn 100 Bratwürste bestellt und 5 storniert wurden, zeigt das Reporting 95 verkaufte."

**KAS (Eva):** „Auch das nur für den Admin sichtbar?"

**DEV1 (Anna):** „Ja, alle Reporting-Ansichten sind Admin-only."

#### Tagesabschluss (R-06)

**FAC (Lisa):** „Zum Schluss des Reporting-Blocks: der Tagesabschluss. Rudi, wie stellt ihr euch das vor?"

**DOM (Rudi):** „Am Ende des Abends sage ich: ‚Das Fest ist vorbei.' Dann muss jemand den formellen Schlussstrich ziehen — offene Tische prüfen, Abschlussbericht, und dann ist Schluss."

**KAS (Eva):** „Ich brauche eine Zusammenfassung — im Grunde die Tagesabrechnung, aber als formeller Abschluss. Und wenn noch Tische offen sind, will ich das vorher wissen."

**SRL (Felix):** „Was passiert mit Tischen, die noch einen offenen Saldo haben? Manchmal haut eine Gruppe ab, ohne zu bezahlen."

**DOM (Rudi):** „Das ist leider Realität. Die müssen wir als Verlust abschreiben."

**DEV1 (Anna):** „Der Tagesabschluss prüft zuerst alle Tische. Wenn Tische mit Saldo > 0 existieren, wird eine Warnung angezeigt. Der Admin kann entscheiden: Trotzdem abschließen, oder erst die Tische klären."

```
👤 Admin
    🔵 Tagesabschluss einleiten
        → [Prüfung: Offene Tische mit Saldo > 0?]
            → Warnung, falls offene Tische vorhanden
        → 🟠 Tagesabschluss durchgeführt (Abschlussbericht generiert)
```

**DEV2 (Tim):** „Und die Frage: Was passiert danach? Soll das System für das nächste Fest zurückgesetzt werden? Events löschen widerspricht Event-Sourcing — die sind append-only."

**DEV1 (Anna):** „Archivieren wäre eine Option — alte Events in ein Archiv verschieben und frisch anfangen. Aber das ist komplex."

**DOM (Rudi):** „Warum nicht einfach die ganze Datenbank zurücksetzen? Wir brauchen die Daten vom letzten Fest nicht im nächsten."

**DEV2 (Tim):** „Das wäre der einfachste Weg — nach dem Export eine frische Instanz aufsetzen. Kein Archiv, keine Migration."

_(❤️ Hotspot: „Tagesabschluss mit offenen Tischen und System-Reset — Verhalten offen")_

**FAC (Lisa):** „Das klingt nach einem Hotspot — wir parken das für später."

### 4.4 Bondruck und Ausgabestationen

#### Bon-Druck (K-11)

**FAC (Lisa):** „Jetzt zum Cluster D: Bons und Ausgabestationen. Klaus, das ist dein Bereich."

**VER1 (Klaus):** „Wenn eine Bestellung aufgegeben wird, muss an der richtigen Stelle ein Bon ankommen. Getränke bei mir an der Theke, Essen in der Küche."

**SRL (Felix):** „Das muss automatisch passieren. Die Servicekraft soll nach der Bestellung nicht noch manuell entscheiden, wo der Bon hingeht."

**DEV1 (Anna):** „Das ist eine Policy: Wenn eine Bestellung aufgegeben wird, wird automatisch ein Bon pro Kategorie gedruckt. Getränkepositionen zum Getränkedrucker, Essenspositionen zum Küchendrucker."

```
🟠 Bestellung aufgegeben
    → 🟣 Policy: Automatischer Bon-Druck nach Kategorie
        → 🔵 Küchenbon drucken (für Essenspositionen)
            → 🟠 Küchenbon gedruckt
        → 🔵 Getränkebon drucken (für Getränkepositionen)
            → 🟠 Getränkebon gedruckt
```

**ADM (Thomas):** „Ich muss konfigurieren können, welcher Drucker welche Kategorie bedient. Wir haben zwei Drucker — einen in der Küche und einen an der Bar."

**DEV2 (Tim):** „Drucker-Konfiguration ist Admin-Sache: Drucker anlegen, Kategorie zuordnen. Das gehört zu den Stammdaten."

**VER1 (Klaus):** „Der Bon muss alle Infos enthalten: Tischnummer, Servicekraft, was bestellt wurde mit Mengen, Uhrzeit und eventuell den Kommentar."

**SRV1 (Jonas):** „Und wenn ein Bon verloren geht oder der Drucker klemmt? Dann brauche ich einen Nachdruck."

**DEV1 (Anna):** „Nachdruck einzelner Positionen — das sollte möglich sein. Der Bon ist ja nur eine Darstellung der Bestelldaten, kein eigener Datenzustand."

_(❤️ Hotspot: „Drucker-Integration — Hardware-Anbindung, Fehlerbehandlung bei Druckerproblemen")_

#### Küchendisplay / KDS (K-12)

**VER1 (Klaus):** „Das Küchendisplay ist für mich das Wichtigste. Wenn ich an der Getränkeausgabe stehe, will ich auf einem Bildschirm sehen, welche Getränkebestellungen offen sind — gruppiert nach Tisch. Und die Küche braucht dasselbe für Essen."

**DEV1 (Anna):** „Das KDS ist ein Read Model — es zeigt in Echtzeit die offenen Bestellungen nach Kategorie an. Keine Eingabe, nur Anzeige."

```
🟠 Bestellung aufgegeben
    → 🟣 Policy: Offene Positionen nach Kategorie auf Display anzeigen
        → 🟢 Read Model: KDS-Ansicht (offene Positionen,
           gruppiert nach Tisch, gefiltert nach Kategorie)
```

**SRV2 (Maria):** „Sieht die Getränkeausgabe dann nur Getränke und die Küche nur Essen?"

**DEV2 (Tim):** „Genau. Das Display filtert nach Kategorie. Jede Ausgabestation sieht nur ihre relevanten Positionen."

**VER1 (Klaus):** „Und wenn ein Bon verloren geht, kann ich trotzdem nachschauen, was bestellt wurde. Das ist die Rückfallebene."

**DEV1 (Anna):** „Die Echtzeit-Aktualisierung ist eine technische Herausforderung. Polling, WebSockets oder Server-Sent Events — da müssen wir im Design eine Entscheidung treffen."

_(❤️ Hotspot: „KDS-Architektur — Echtzeit-Datenfluss zur Küche/Ausgabe")_

**DOM (Rudi):** „Wie viele Displays brauchen wir? Einen pro Ausgabestation?"

**ADM (Thomas):** „Idealerweise ja — ein Tablet oder Monitor an der Bar, eines in der Küche. Aber die URL ist dieselbe, nur die Kategorie-Filter unterscheiden sich."

**DEV2 (Tim):** „Genau. Das KDS ist eine eigene Ansicht im Browser, die per URL-Parameter oder Einstellung auf eine Kategorie gefiltert wird. Kein eigener Server, keine eigene App."

#### Zubereitungsstatus (K-13)

**VER1 (Klaus):** „Jetzt kommt mein Wunsch: Wenn ich ein Getränk zubereite, will ich es als ‚in Zubereitung' markieren. Und wenn es fertig ist, als ‚fertig'. Dann weiß die Servicekraft, dass sie es abholen kann."

**SRV1 (Jonas):** „Das wäre genial. Dann muss ich nicht mehr in die Küche laufen und fragen. Ich sehe auf meinem Handy, welche Positionen fertig sind."

**DEV1 (Anna):** „Das ist ein aktiver Prozess — Klaus interagiert mit dem System. Der Command ist ‚Zubereitungsstatus ändern'."

```
👤 Servicekraft / Serviceleitung / Admin / Ausgabe-Mitarbeiter
    🔵 Position als „in Zubereitung" markieren
        → 🟠 Position als „in Zubereitung" markiert

    🔵 Position als „fertig" markieren
        → 🟠 Position als „fertig" markiert
```

**DEV2 (Tim):** „Die Frage ist: Sind das Domain Events im Kassenjournal, oder eher UI-State?"

**DEV1 (Anna):** „Gute Frage. Der Zubereitungsstatus ist kurzlebig — er ist nur während des laufenden Betriebs relevant. Für die Abrechnung ist er egal. Er hat keinen Einfluss auf den Saldo."

**VER1 (Klaus):** „Für mich an der Ausgabe ist er aber sehr relevant. Wenn ich wissen will, was ich noch machen muss, schaue ich auf mein Display."

**DEV2 (Tim):** „Wir könnten es als eigene Events modellieren — dann ist es persistent und nachvollziehbar. Oder als transienten UI-State — dann geht die Info bei einem Seitenrefresh verloren."

**DEV1 (Anna):** „Ich tendiere zu Events. Die Daten sind klein, und es hat einen fachlichen Mehrwert: Die Servicekraft sieht den Status, und nachträglich könnte man analysieren, wie lange die Zubereitung gedauert hat."

**FAC (Lisa):** „Das parken wir als offene Frage — Events oder UI-State?"

_(❤️ Hotspot: „Zubereitungsstatus — eigene Events oder transienter UI-State?")_

**SRV2 (Maria):** „Wer darf den Status ändern? Nur die Ausgabe-Leute?"

**VER1 (Klaus):** „Eigentlich alle, die an der Ausgabe stehen — und im Notfall auch die Servicekraft selbst."

**DEV1 (Anna):** „Am einfachsten: Jeder angemeldete Benutzer kann den Status ändern. Die Berechtigung ist hier nicht sicherheitskritisch — es ist ein Hilfsmittel, kein finanzieller Vorgang."

**FAC (Lisa):** „Damit haben wir alle vier Cluster durchgearbeitet. Gut gemacht! Machen wir 15 Minuten Pause, dann geht es weiter mit Phase 4 — dem Software Design."

---

## 5. Phase 4 — Software Design: Aggregate und Read Models

**Zeitrahmen:** 13:30–14:45 Uhr (75 Minuten)

**FAC (Lisa):** „Willkommen zurück! In den letzten drei Phasen haben wir Events entdeckt, sie gruppiert und Prozesse mit Commands, Akteuren und Policies modelliert. Jetzt geht es ans Software Design. Wir beantworten drei Fragen: **Erstens**, welche Aggregate brauchen wir — also welche transaktionalen Grenzen schützen unsere Geschäftsregeln? **Zweitens**, welche Event-Typen benennen wir — und wie benennen wir sie? **Drittens**, welche Read Models brauchen die verschiedenen Akteure — Servicekräfte, Admin, Kassenwart, Ausgabe-Mitarbeiter?"

**DEV1 (Anna):** „Gelbe Stickies für Aggregate, grüne für Read Models. Wir gehen die Cluster durch und schauen, wo transaktionale Grenzen liegen."

**DEV2 (Tim):** „Und wir sprechen über eine wichtige Architekturentscheidung: Wo setzen wir Event-Sourcing ein — und wo reicht klassisches CRUD?"

**FAC (Lisa):** „Genau. Los geht's."

### 5.1 Aggregate und Persistenzstrategie

#### Aggregate identifizieren

**FAC (Lisa):** „Fangen wir mit der zentralen Frage an: Welche Aggregate haben wir? Anna, kannst du kurz erklären, was ein Aggregat ist — für alle, die nicht aus der Softwareentwicklung kommen?"

**DEV1 (Anna):** „Gerne. Ein Aggregat ist eine Gruppe von zusammengehörigen Daten, die immer gemeinsam konsistent sein müssen. Es hat eine klare Grenze: Alles innerhalb des Aggregats wird in einer Transaktion verarbeitet. Und es schützt Geschäftsregeln — sogenannte Invarianten."

**SRV1 (Jonas):** „Könnt ihr ein Beispiel geben?"

**DEV2 (Tim):** „Klar. Nehmen wir den Tisch. Wenn eine Bestellung aufgegeben wird, eine Zahlung registriert wird oder eine Stornierung erfolgt — all das verändert den Zustand **eines** Tisches. Der Saldo muss stimmen, die Positionen müssen konsistent sein. Das ist ein natürliches Aggregat."

**FAC (Lisa):** „Gehen wir die Cluster der Reihe nach durch. Fangen wir mit dem Kassenbetrieb an — Cluster B."

**DEV1 (Anna):** _(klebt ein gelbes Sticky)_ „Das zentrale Aggregat im Kassenbetrieb ist der **Tisch**. Alle Operationen aus Cluster B und C — Bestellung, Lieferung, Zahlung, Stornierung, Umbuchung — verändern den Zustand eines bestimmten Tisches."

**DOM (Rudi):** „Der Tisch ist wirklich das Zentrum von allem. Alles dreht sich darum: Was wurde bestellt, was geliefert, was bezahlt, was storniert — und was ist noch offen."

**DEV2 (Tim):** „Die Invarianten des Tisch-Aggregats sind: Der Saldo darf nicht negativ werden durch eine Zahlung, die höher ist als der offene Betrag. Es darf nur storniert werden, was vorher bestellt wurde. Lieferungen beziehen sich nur auf bestellte Positionen."

**SRL (Felix):** „Und Stornierung nur durch Serviceleitung oder Admin — das ist auch eine Regel."

**DEV1 (Anna):** „Genau. Die Berechtigung ist eine Invariante, die das Aggregat prüfen kann — oder alternativ die Schicht davor. Aber die fachliche Regel gehört zum Tisch."

```
🟡 Aggregat: TISCH
    Invarianten:
    - Saldo = Summe(Bestellungen) − Summe(Zahlungen) − Summe(Stornierungen)
    - Nur bestellte Positionen können geliefert werden
    - Nur bestellte, nicht-stornierte Positionen können storniert werden
    - Nur bestellte, nicht-bezahlte Positionen können bezahlt werden
    - Stornierung nur durch serviceleitung oder admin

    Events:
    🟠 Bestellung aufgegeben (K-01, K-07)
    🟠 Produkte geliefert (K-03)
    🟠 Zahlung registriert (K-02)
    🟠 Produkte storniert (K-04)
```

**FAC (Lisa):** „Gut. Was ist mit den Stammdaten — Cluster A?"

**ADM (Thomas):** „Produkte, Tische und Benutzer — die verwalte ich als Admin."

**DEV1 (Anna):** „Jedes davon ist ein eigenes Aggregat. Das **Produkt** enthält die Varianten, der **Tisch** ist hier die Stammdaten-Entität — Name und Status —, und der **Benutzer** enthält Zugangsdaten und Rolle."

**DEV2 (Tim):** „Moment — wir haben ‚Tisch' jetzt zweimal. Einmal als Stammdaten-Entität: Name, aktiv/inaktiv. Und einmal als Kassenbetrieb-Aggregat: Bestellungen, Zahlungen, Saldo."

**DOM (Rudi):** „Stimmt. Der Tisch, den ich anlege — ‚Tisch 1, Tisch 2' — ist etwas anderes als der Tisch, an dem bestellt und bezahlt wird."

**DEV1 (Anna):** „Genau. Im Kassenbetrieb ist der Tisch ein Event-Stream — eine Ansammlung von Vorgängen. In den Stammdaten ist er eine einfache Entität mit Name und Status. Die teilen sich nur die ID."

**FAC (Lisa):** „Guter Punkt. Wir sollten das klar trennen."

```
🟡 Aggregat: PRODUKT (Stammdaten)
    Invarianten:
    - Name und Kategorie dürfen nicht leer sein
    - Variante braucht Name und Preis (> 0, in Cent)
    - Soft-Delete: Status wird auf „deleted" gesetzt, nicht physisch gelöscht

🟡 Aggregat: TISCH (Stammdaten)
    Invarianten:
    - Name darf nicht leer sein
    - Soft-Delete: Status wird auf „deleted" gesetzt

🟡 Aggregat: BENUTZER (Stammdaten)
    Invarianten:
    - Benutzername muss eindeutig sein
    - Rolle muss gültig sein (admin, serviceleitung, service)
    - Passwort wird gehasht gespeichert
    - Soft-Delete: Status wird auf „deleted" gesetzt
```

**KAS (Eva):** „Und was ist mit den Abrechnungen? Ist das ein Aggregat?"

**DEV2 (Tim):** „Nein. Die Abrechnung ist kein eigener Datenzustand — sie ist eine **Auswertung** der bestehenden Daten. Ein Read Model, das aus den Tisch-Events berechnet wird. Dazu kommen wir gleich."

**VER1 (Klaus):** „Und das Küchendisplay? Das ist auch kein Aggregat?"

**DEV1 (Anna):** „Richtig. Das KDS zeigt Daten an, die aus den Bestellungen kommen. Es ist eine Lese-Ansicht — ein Read Model. Es hat keinen eigenen Zustand, den es konsistent halten muss."

**DEV2 (Tim):** „Wobei — der Zubereitungsstatus ist eine offene Frage. Wenn wir den als Events modellieren, brauchen wir dafür vielleicht ein eigenes kleines Aggregat oder erweitern das Tisch-Aggregat."

**DEV1 (Anna):** „Stimmt. Ob der Zubereitungsstatus zum Tisch-Aggregat gehört oder eigenständig ist, parken wir — das hängt davon ab, ob wir ihn als Domain Event oder als transienten State behandeln."

_(❤️ Hotspot: „Zubereitungsstatus — Teil des Tisch-Aggregats oder eigenständig?")_

#### Event-Sourcing vs. CRUD

**FAC (Lisa):** „Tim hat vorhin Event-Sourcing und CRUD erwähnt. Können wir das jetzt klären? Wo setzen wir welches Pattern ein?"

**DEV2 (Tim):** „Event-Sourcing bedeutet: Wir speichern nicht den aktuellen Zustand, sondern die komplette Geschichte — alle Events, die passiert sind. Den aktuellen Zustand berechnen wir aus den Events."

**SRV2 (Maria):** „Warum nicht einfach den aktuellen Stand speichern?"

**DEV1 (Anna):** „Gute Frage. Beim Tisch ist die Geschichte wichtig. Thomas will wissen: Wann hat wer was bestellt, wann bezahlt, wann storniert. Das Kassenjournal **ist** die Liste der Events (K-06). Wenn wir nur den aktuellen Saldo speichern würden, wäre die gesamte Historie verloren."

**KAS (Eva):** „Aus Buchhaltungssicht ist das unverzichtbar. Ich muss jeden einzelnen Vorgang nachvollziehen können — für den Verein und eventuell fürs Finanzamt."

**DEV2 (Tim):** „Genau. Für den Tisch-Kassenbetrieb ist Event-Sourcing perfekt: Unveränderliche Events, lückenlose Historie, der Zustand wird berechnet. Keine Updates, keine Deletes."

**DOM (Rudi):** „Und für Produkte? Brauche ich da auch die ganze Geschichte?"

**DEV1 (Anna):** „Eher nicht. Wenn du den Preis einer Bratwurst von 3,50 € auf 4,00 € änderst, interessiert dich nur der aktuelle Preis. Die Preishistorie ist irrelevant — die alten Preise stecken ja schon in den Bestellungs-Events der Tische."

**DEV2 (Tim):** „Für Stammdaten — Produkte, Tische, Benutzer — reicht klassisches CRUD. Anlegen, Lesen, Aktualisieren, Löschen. Einfacher, weniger Overhead."

**ADM (Thomas):** „Das macht Sinn. Ich will ein Produkt anlegen und fertig. Nicht eine Event-Geschichte aufbauen."

**DEV1 (Anna):** _(schreibt auf ein Sticky)_ „Also die Aufteilung:"

```
Persistenzstrategie:

EVENT-SOURCING (append-only):
    🟡 Tisch (Kassenbetrieb) — Bestellungen, Lieferungen, Zahlungen, Stornierungen
    → Jede Operation wird als unveränderliches Event gespeichert
    → Der aktuelle Tischzustand wird aus dem Event-Stream berechnet
    → Das Kassenjournal (K-06) IST der Event-Stream

CRUD (klassisch):
    🟡 Produkt — Anlegen, Bearbeiten, Deaktivieren (Soft-Delete)
    🟡 Tisch (Stammdaten) — Anlegen, Umbenennen, Deaktivieren (Soft-Delete)
    🟡 Benutzer — Anlegen, Bearbeiten, Deaktivieren (Soft-Delete)
```

**FAC (Lisa):** „Schöne klare Trennung. Event-Sourcing dort, wo die Historie fachlich relevant ist — beim Kassenbetrieb. CRUD dort, wo wir nur den aktuellen Stand brauchen — bei Stammdaten."

**DEV2 (Tim):** „Ein wichtiger Punkt noch: Die Bestellungs-Events im Tisch-Aggregat speichern die Produktdaten zum Zeitpunkt der Bestellung mit — den Produktnamen, den Variantennamen, den Preis. So sind die Events autark und unabhängig davon, ob sich das Produkt später ändert."

**DEV1 (Anna):** „Genau — sogenannte Fat Events. Jedes Event enthält alle Informationen, die es braucht, um sich selbst zu erklären. Das ist wichtig, weil wir sonst bei einem Event-Replay gar nicht mehr wüssten, was damals bestellt wurde."

**KAS (Eva):** „Also wenn ich eine Bratwurst für 3,50 € ins Abrechnungsdokument sehe, ist das der Preis zum Zeitpunkt der Bestellung — auch wenn der Preis danach geändert wurde?"

**DEV2 (Tim):** „Exakt."

### 5.2 Event-Typen und Namenskonventionen

#### Event-Typen benennen

**FAC (Lisa):** „Jetzt benennen wir die Event-Typen konkret. Bevor wir loslegen: Welche Konventionen wollen wir?"

**DEV1 (Anna):** „Vorschlag: Alle Event-Typen auf Deutsch — wir sind in der Fachdomäne eines deutschen Vereins. Die Events sollen für alle Beteiligten lesbar sein, nicht nur für Entwickler."

**DOM (Rudi):** „Deutsch ist gut. Dann versteht auch der Vorstand, was im System passiert."

**DEV2 (Tim):** „Und in der **Vergangenheitsform** — weil Events etwas beschreiben, das bereits passiert ist. ‚Bestellung aufgegeben', nicht ‚Bestellung aufgeben'."

**DEV1 (Anna):** „Genau. Und als **zusammengesetzter Name** — erst das Substantiv, dann das Partizip. Also ‚BestellungAufgegeben', nicht ‚AufgegebeneBestellung' oder ‚OrderPlaced'."

**SRV1 (Jonas):** „PascalCase — alles zusammen, Großbuchstaben am Anfang jeder Einheit?"

**DEV2 (Tim):** „Genau. Kein Leerzeichen, kein Unterstrich, kein Bindestrich. Das ist im Code direkt als Typname verwendbar."

**FAC (Lisa):** „Dann gehen wir die Events der Reihe nach durch. Fangen wir beim Tisch-Aggregat an."

**DEV1 (Anna):** _(notiert auf gelben und orangenen Stickies)_ „Die Events des Tisch-Aggregats aus Phase 3:"

**SRL (Felix):** „‚Bestellung aufgegeben' — das ist klar. Da steckt alles drin: Tisch, Positionen, Mengen, Preise, Kommentar."

**DEV2 (Tim):** „‚**BestellungAufgegeben**'. Enthält: Tisch-ID, Positionen mit Produkt, Variante, Menge, Einzelpreis; optionaler Kommentar, Servicekraft, Zeitstempel."

**SRV1 (Jonas):** „Dann die Lieferung: ‚**ProdukteGeliefert**'. Weil es immer mehrere Positionen auf einmal geliefert werden können."

**DEV1 (Anna):** „Enthält: Tisch-ID, die gelieferten Positionen mit Referenz auf die Bestellung, optionaler Kommentar, Servicekraft, Zeitstempel."

**KAS (Eva):** „Die Zahlung: ‚**ZahlungRegistriert**'. Mit den bezahlten Positionen und dem Betrag."

**DEV2 (Tim):** „Enthält: Tisch-ID, die bezahlten Positionen, Gesamtbetrag in Cent, optionaler Kommentar, Servicekraft, Zeitstempel."

**SRL (Felix):** „Stornierung: ‚**ProdukteStorniert**'. Plural, weil mehrere Positionen storniert werden können."

**DEV1 (Anna):** „Enthält: Tisch-ID, stornierte Positionen mit Referenz auf die Bestellung, Stornobetrag in Cent, optionaler Kommentar, durchführende Person, Zeitstempel."

**FAC (Lisa):** „Gibt es weitere Events im Tisch-Aggregat?"

**DEV1 (Anna):** „Wenn wir den Zubereitungsstatus als Events modellieren, kämen noch zwei dazu: ‚**PositionInZubereitungGesetzt**' und ‚**PositionFertigGemeldet**'. Aber das ist noch offen."

**DEV2 (Tim):** „Und bei der Umbuchung (K-08): Da entstehen zwei Events — eine Stornierung am Quell-Tisch und eine Bestellung am Ziel-Tisch. Kein eigener Event-Typ, sondern eine Kombination bestehender Typen in einer atomaren Transaktion."

```
Event-Typen des Tisch-Aggregats:

🟠 BestellungAufgegeben (K-01, K-07)
    Tisch-ID, Positionen [{Produkt, Variante, Menge, Einzelpreis}],
    opt. Kommentar, Servicekraft, Zeitstempel

🟠 ProdukteGeliefert (K-03)
    Tisch-ID, gelieferte Positionen [{Referenz auf Bestellposition}],
    opt. Kommentar, Servicekraft, Zeitstempel

🟠 ZahlungRegistriert (K-02)
    Tisch-ID, bezahlte Positionen [{Referenz auf Bestellposition}],
    Betrag (Cent), opt. Kommentar, Servicekraft, Zeitstempel

🟠 ProdukteStorniert (K-04)
    Tisch-ID, stornierte Positionen [{Referenz auf Bestellposition}],
    Stornobetrag (Cent), opt. Kommentar, Person, Zeitstempel
```

**KAS (Eva):** „Und was ist mit dem Tagesabschluss? Ist ‚TagesabschlussDurchgeführt' auch ein Event?"

**DEV1 (Anna):** „Guter Punkt. Der Tagesabschluss (R-06) betrifft nicht einen einzelnen Tisch, sondern die gesamte Veranstaltung. Das ist kein Tisch-Event."

**DEV2 (Tim):** „Es wäre ein **systemweites Event** — wenn wir es als Event modellieren wollen. Aber eigentlich ist der Tagesabschluss eine Aktion, die einen Bericht generiert und optional das System zurücksetzt. Kein klassisches Domain Event."

**FAC (Lisa):** „Parken wir das beim Hotspot ‚Tagesabschluss'."

#### Fachliche vs. technische Events

**FAC (Lisa):** „In Phase 1 hatten wir auch Events rund um Login, Logout und Passwort setzen. Wie ordnen wir die ein?"

**DEV1 (Anna):** „Das ist eine wichtige Unterscheidung. Wir haben **fachliche Events** — die beschreiben, was in der Geschäftsdomäne passiert. Bestellung, Zahlung, Lieferung, Stornierung. Und wir haben **technische Events** — die betreffen die Infrastruktur. Login, Logout, Session-Verwaltung."

**DEV2 (Tim):** „Fachliche Events gehören in unsere Event-Streams und Aggregate. Technische Events gehören zur Infrastruktur — sie haben nichts im Kassenjournal verloren."

**DOM (Rudi):** „Also wenn sich Jonas morgens einloggt, ist das kein Event, das neben den Bestellungen steht?"

**DEV1 (Anna):** „Genau. Das Login ist wichtig für die Sicherheit und die Authentifizierung — aber es ist kein Geschäftsvorfall. Es hat keinen Einfluss auf den Saldo eines Tisches."

**SRV2 (Maria):** „Und wenn ich mein Passwort setze?"

**DEV2 (Tim):** „Das ist Infrastruktur — Benutzerverwaltung und Authentifizierung. Technisch notwendig, aber kein Domain Event. Das wird klassisch in der Datenbank gespeichert, nicht im Event-Stream."

**ADM (Thomas):** „Macht Sinn. Die Auth-Sachen laufen separat."

**DEV1 (Anna):** _(schreibt auf ein Sticky)_

```
Fachliche Domain Events (im Event-Stream):
    🟠 BestellungAufgegeben
    🟠 ProdukteGeliefert
    🟠 ZahlungRegistriert
    🟠 ProdukteStorniert
    (optional: PositionInZubereitungGesetzt, PositionFertigGemeldet)

Technische / Infrastruktur-Vorgänge (NICHT im Event-Stream):
    ⚙️ Benutzer eingeloggt (A-01)
    ⚙️ Passwort gesetzt (A-02)
    ⚙️ Benutzer ausgeloggt (A-03)
    ⚙️ Produkt angelegt / geändert / deaktiviert (S-01)
    ⚙️ Tisch angelegt / umbenannt / deaktiviert (S-02)
    ⚙️ Benutzer angelegt / geändert / deaktiviert (S-03)
```

**FAC (Lisa):** „Klare Trennung. Die fachlichen Events leben im Kassenjournal, die technischen Vorgänge werden klassisch als CRUD-Operationen gespeichert."

**KAS (Eva):** „Für die Abrechnung sind nur die fachlichen Events relevant — das leuchtet ein."

### 5.3 Read Models und Projektionen

**FAC (Lisa):** „Jetzt kommen die grünen Stickies — die Read Models. Welche Ansichten brauchen unsere verschiedenen Akteure?"

**DEV1 (Anna):** „Ein Read Model ist eine speziell aufbereitete Lese-Ansicht der Daten. Es wird aus den Events oder den Stammdaten zusammengebaut und enthält genau die Informationen, die ein bestimmter Akteur in einer bestimmten Situation braucht."

#### Ansichten für den Service-Betrieb

**FAC (Lisa):** „Fangen wir bei den Servicekräften an. Jonas, Maria, Felix — was braucht ihr auf dem Bildschirm?"

**SRV1 (Jonas):** „Als erstes: Die Tischübersicht (K-05). Alle Tische auf einen Blick — wie Karten. Ich will sofort sehen, welcher Tisch offen ist und wie hoch der Saldo ist."

**DEV2 (Tim):** _(klebt ein grünes Sticky)_ „Read Model: **Tischübersicht**."

```
🟢 Read Model: TISCHÜBERSICHT (K-05)
    Quelle: Tisch-Events + Tisch-Stammdaten
    Inhalt pro Tisch:
    - Tischname
    - Aktueller Saldo (offen)
    - Anzahl unbezahlte Positionen
    - Anzahl ungelieferte Positionen
    Anzeige: Karten-Layout, alle aktiven Tische
    Akteure: Servicekraft, Serviceleitung, Admin
```

**SRV2 (Maria):** „Und wenn ich auf einen Tisch tippe, will ich die Details sehen: Was wurde bestellt, was ist geliefert, was bezahlt, was noch offen."

**DEV1 (Anna):** „Read Model: **Tischdetails**."

```
🟢 Read Model: TISCHDETAILS (K-05)
    Quelle: Tisch-Events
    Inhalt:
    - Alle Positionen mit Status (bestellt / geliefert / bezahlt / storniert)
    - Gruppiert nach Bestellung
    - Aktueller Saldo
    - Unbezahlte Positionen (für Zahlung)
    - Ungelieferte Positionen (für Lieferung)
    Anzeige: Tabs — Übersicht, Bestellen, Liefern, Bezahlen, Stornieren, Historie
    Akteure: Servicekraft, Serviceleitung, Admin
```

**SRL (Felix):** „Ich brauche auch die Schnellsuche (K-10). Wenn 30 Tische da sind, will ich nicht scrollen, sondern direkt die Nummer eingeben."

**DEV2 (Tim):** „Das ist kein eigenes Read Model, sondern ein Filter auf der Tischübersicht. Aber notieren wir es als Anforderung an die Ansicht."

**KAS (Eva):** „Und das Kassenjournal (K-06)? Das hatten wir in Phase 3 als die vollständige Historie eines Tisches."

**DEV1 (Anna):** „Read Model: **Kassenjournal**. Das zeigt alle Events eines Tisches in chronologischer Reihenfolge — im Grunde der Event-Stream in menschenlesbarer Form."

```
🟢 Read Model: KASSENJOURNAL (K-06)
    Quelle: Tisch-Events (Event-Stream)
    Inhalt:
    - Chronologische Liste aller Vorgänge am Tisch
    - Pro Eintrag: Zeitstempel, Typ (Bestellung/Lieferung/Zahlung/Stornierung),
      Positionen, Betrag, Servicekraft, Kommentar
    Anzeige: Timeline / Liste im Tisch-Detail (Tab „Historie")
    Akteure: Servicekraft, Serviceleitung, Admin
```

#### Reporting-Ansichten (R-01 bis R-06)

**FAC (Lisa):** „Jetzt zu den Reporting-Ansichten. Eva, das ist dein Bereich."

**KAS (Eva):** „Genau. Ich brauche am Ende des Tages — oder auch zwischendurch — einen Überblick über die Finanzen."

**DEV1 (Anna):** „Alle Reporting-Ansichten sind Read Models, die aus den Tisch-Events über alle Tische hinweg aggregiert werden."

**KAS (Eva):** „Die Tagesabrechnung (R-01) zeigt den Gesamtumsatz, den Umsatz pro Servicekraft und alle Stornierungen."

```
🟢 Read Model: TAGESABRECHNUNG (R-01)
    Quelle: Tisch-Events (tischübergreifend)
    Inhalt:
    - Gesamtumsatz (Summe aller Zahlungen)
    - Umsatz pro Servicekraft (Übersichtswerte)
    - Übersicht aller Stornierungen (Zeitpunkt, Tisch, Positionen, Betrag)
    - Offene Beträge (noch nicht bezahlte Positionen)
    Akteure: Admin
```

**DOM (Rudi):** „Die Abrechnung pro Tisch (R-03) brauche ich, um nachzuvollziehen, was an einem bestimmten Tisch passiert ist."

**DEV2 (Tim):** „Das ist im Grunde das Kassenjournal eines einzelnen Tisches, aber mit einer zusammenfassenden Ansicht für den Admin — Saldo, Bestellsumme, Zahlungssumme, Stornierungen."

```
🟢 Read Model: ABRECHNUNG PRO TISCH (R-03)
    Quelle: Tisch-Events (einzelner Tisch)
    Inhalt:
    - Alle Bestellungen, Zahlungen, Lieferungen, Stornierungen (chronologisch)
    - Gesamt-Saldo (bestellt, bezahlt, offen, storniert)
    Akteure: Admin
```

**KAS (Eva):** „Und die Abrechnung pro Servicekraft (R-04). Da will ich sehen: Wie viel Umsatz hat Jonas gemacht, wie viel Maria? Und wer hat storniert?"

**DEV1 (Anna):** „Das aggregiert die Tisch-Events nach der Servicekraft, die die Aktion durchgeführt hat."

```
🟢 Read Model: ABRECHNUNG PRO SERVICEKRAFT (R-04)
    Quelle: Tisch-Events (tischübergreifend, gruppiert nach Servicekraft)
    Inhalt:
    - Umsatz pro Servicekraft (Summe registrierter Zahlungen)
    - Anzahl aufgegebener Bestellungen
    - Anzahl und Betrag der Stornierungen
    Akteure: Admin
```

**DOM (Rudi):** „Und das Produktumsatz-Reporting (R-05) — was hat sich gut verkauft?"

**DEV2 (Tim):** „Das aggregiert die Bestellungs-Events nach Produkt und Variante. Stornierte Positionen werden abgezogen."

```
🟢 Read Model: PRODUKTUMSATZ (R-05)
    Quelle: Tisch-Events (tischübergreifend, gruppiert nach Produkt/Variante)
    Inhalt:
    - Verkaufte Menge pro Produkt und Variante (abzüglich Stornierungen)
    - Ranking der meistverkauften Varianten
    - Gesamteinnahmen pro Produkt/Variante
    Akteure: Admin
```

**KAS (Eva):** „Und der Datenexport (R-02)? Ich brauche die Daten als CSV für unsere Vereinsbuchhaltung."

**DEV1 (Anna):** „Der Export ist kein eigenständiges Read Model — er ist eine alternative Darstellung der bestehenden Daten. Er nimmt die Daten aus den anderen Read Models und gibt sie als CSV aus."

**DEV2 (Tim):** „Genau. Umsätze, Bestellungen, Artikeldaten — als CSV-Download. Die Daten sind dieselben, nur das Format ist anders."

**KAS (Eva):** „Und der Tagesabschluss (R-06)?"

**DEV1 (Anna):** „Der Tagesabschluss ist weniger ein Read Model als ein Prozess: Offene Tische prüfen, Abschlussbericht generieren, optional das System zurücksetzen. Der Bericht selbst ist im Grunde die Tagesabrechnung (R-01) — aber mit einem formellen Abschluss-Charakter."

#### Küchen- und Ausgabe-Ansichten (K-12, K-13)

**FAC (Lisa):** „Und jetzt die Ansichten für die Ausgabestationen. Klaus, was brauchst du?"

**VER1 (Klaus):** „Wie gesagt: Ich stehe an der Getränkeausgabe und will auf einem Bildschirm sehen, was offen ist. Nur Getränke, gruppiert nach Tisch. Und am besten in Echtzeit — wenn Jonas eine Cola bestellt, soll die bei mir erscheinen."

**DEV2 (Tim):** „Read Model: **KDS-Ansicht** — gefiltert nach Kategorie."

```
🟢 Read Model: KDS-ANSICHT (K-12)
    Quelle: Tisch-Events (BestellungAufgegeben), gefiltert nach Kategorie
    Inhalt:
    - Offene (ungelieferte) Positionen einer Kategorie (Essen ODER Getränke)
    - Gruppiert nach Tisch
    - Pro Position: Produkt, Variante, Menge, Zeitpunkt der Bestellung
    Anzeige: Echtzeit-Updates, große Schrift (für Monitore in Küche/Ausgabe)
    Akteure: Ausgabe-Mitarbeiter (VER1), Servicekraft (Einsicht)
```

**VER1 (Klaus):** „Und wenn wir den Zubereitungsstatus dazunehmen (K-13): Ich markiere eine Position als ‚in Zubereitung', und wenn sie fertig ist, als ‚fertig'. Dann sieht Jonas auf seinem Handy, dass er die Cola abholen kann."

**DEV1 (Anna):** „Dann brauchen wir eine erweiterte Ansicht — die KDS-Ansicht plus Zubereitungsstatus."

```
🟢 Read Model: ZUBEREITUNGSSTATUS (K-13)
    Quelle: Tisch-Events (BestellungAufgegeben) + Zubereitungsstatus-Events
    Inhalt:
    - Offene Positionen mit Status: „offen" → „in Zubereitung" → „fertig"
    - Gruppiert nach Tisch
    - Zeitpunkt des letzten Statuswechsels
    Anzeige: Farbcodierung nach Status auf KDS und Servicekraft-Ansicht
    Akteure: Ausgabe-Mitarbeiter (Status ändern), Servicekraft (Status einsehen)
```

**SRV1 (Jonas):** „Wo sehe ich den Zubereitungsstatus als Servicekraft? Auf dem Tisch-Detail?"

**DEV2 (Tim):** „Das wäre sinnvoll — im Tisch-Detail die offenen Positionen mit ihrem Zubereitungsstatus anzeigen. ‚in Zubereitung' und ‚fertig' als visueller Indikator."

**SRV2 (Maria):** „Cool. Dann muss ich nicht mehr in die Küche laufen und fragen."

#### Wie Read Models aus Events aufgebaut werden

**FAC (Lisa):** „Letzte Frage zum Software Design: Wie werden diese Read Models technisch aus den Events aufgebaut?"

**DEV1 (Anna):** „Das Grundprinzip ist die **Projektion**. Wir nehmen den Stream aller Events eines Tisches — oder aller Tische — und berechnen daraus den gewünschten Zustand. Dazu iterieren wir über die Events und aktualisieren das Read Model schrittweise."

**DEV2 (Tim):** „Ein Beispiel: Für den Saldo eines Tisches starten wir bei 0. Dann kommt eine ‚BestellungAufgegeben' — wir addieren den Bestellwert. Dann eine ‚ZahlungRegistriert' — wir subtrahieren den Zahlungsbetrag. Dann ‚ProdukteStorniert' — wir subtrahieren den Stornobetrag. Am Ende haben wir den aktuellen Saldo."

**SRV1 (Jonas):** „Und das passiert bei jedem Aufruf? Klingt langsam, wenn ein Tisch 50 Events hat."

**DEV1 (Anna):** „Guter Einwand. Bei kurzen Event-Streams ist das kein Problem. Aber für Performance-Optimierung gibt es **Snapshots**: Wir speichern zwischendurch den berechneten Zustand als Momentaufnahme ab. Beim nächsten Aufruf starten wir beim letzten Snapshot und wenden nur die neuen Events an."

**DEV2 (Tim):** „Wichtig: Snapshots sind rein technisch — sie haben keine fachliche Bedeutung. Sie beschleunigen nur das Laden. Die Wahrheit bleibt immer der Event-Stream."

**KAS (Eva):** „Und die tischübergreifenden Auswertungen? Die Tagesabrechnung geht ja über alle Tische."

**DEV1 (Anna):** „Dafür projizieren wir über alle Tisch-Streams hinweg. Jedes Zahlungs-Event trägt die Servicekraft-ID und die Produkt-IDs — daraus können wir Umsatz pro Servicekraft, Umsatz pro Produkt und Gesamtumsatz berechnen."

**DEV2 (Tim):** „Im Grunde sind die Reporting-Ansichten Aggregationen über die gesamte Event-Datenbank — gefiltert nach Zeitraum, gruppiert nach der jeweiligen Dimension."

**DOM (Rudi):** „Klingt solide. Die Events sind einmal gespeichert, und je nachdem, wer fragt, bekommt er eine andere Sicht auf dieselben Daten."

**FAC (Lisa):** „Genau. Das ist das Kern-Prinzip: Eine einzige Datenquelle — die Events — und viele verschiedene Lese-Ansichten. Damit haben wir Phase 4 abgeschlossen. Weiter geht's mit den Bounded Contexts!"

---

## 6. Bounded Contexts und Domain Map

**Zeitrahmen:** 14:45–15:00 Uhr (15 Minuten)

**FAC (Lisa):** „Letzte Etappe! Wir haben Aggregate, Events, Read Models — jetzt ziehen wir die Grenzen. Wir identifizieren **Bounded Contexts**: Bereiche unserer Fachdomäne, die jeweils eigene Sprache und eigene Regeln haben. Danach ordnen wir sie ein — was ist unser Kern, was unterstützt, was ist rein technisch? Und schließlich zeichnen wir eine **Context Map**, die zeigt, wie die Bereiche miteinander kommunizieren."

**DEV1 (Anna):** „Ein Bounded Context ist ein klar abgegrenzter Bereich, in dem bestimmte Begriffe eine genau definierte Bedeutung haben. ‚Tisch' kann in einem Kontext ‚der Event-Stream mit Bestellungen und Zahlungen' bedeuten und in einem anderen einfach ‚eine Stammdaten-Entität mit Name und Status'. Und das ist okay — solange wir wissen, in welchem Kontext wir uns bewegen."

**DEV2 (Tim):** „Genau. Und die Bounded Contexts helfen uns zu entscheiden, wie wir das System schneiden — welche Teile unabhängig voneinander entwickelt und verändert werden können."

**FAC (Lisa):** „Los geht's. Schauen wir uns unsere Cluster und Aggregate an und suchen nach natürlichen Grenzen."

### 6.1 Context-Identifikation

**FAC (Lisa):** „Welche fachlichen Bereiche haben wir identifiziert, die jeweils eigene Sprache und eigene Regeln haben?"

**DEV1 (Anna):** „Am offensichtlichsten ist der **Kassenbetrieb**. Da dreht sich alles um den Tisch als Event-Stream: Bestellungen aufgeben, Lieferungen bestätigen, Zahlungen registrieren, Stornierungen durchführen. Die Sprache dort ist: Bestellung, Position, Lieferung, Zahlung, Stornierung, Saldo, Kassenjournal."

**SRV1 (Jonas):** „Das ist genau das, was ich jeden Tag mache — am Tisch arbeiten."

**DOM (Rudi):** „Und der Kassenbetrieb hat seine eigenen Regeln: Saldo muss stimmen, nur bestellte Positionen können geliefert oder bezahlt werden, Stornierung nur mit Berechtigung."

**DEV2 (Tim):** „Zweiter Bereich: **Stammdaten**. Da geht es um die Verwaltung von Produkten, Tischen und Benutzern. Ganz andere Sprache: Produkt, Variante, Kategorie, Preis, Rolle, aktiv/deaktiviert. Und ganz andere Mechanik — klassisches CRUD, kein Event-Sourcing."

**ADM (Thomas):** „Das ist mein Bereich. Ich lege Produkte an, verwalte Tische, erstelle Benutzer-Accounts. Das hat nichts mit dem Kassenbetrieb zu tun — ich mache das vor der Veranstaltung."

**FAC (Lisa):** „Guter Punkt. Die Stammdaten leben zeitlich vor dem Kassenbetrieb — sie sind die Vorbereitung."

**VER1 (Klaus):** „Und was ist mit den Ausgabestationen? KDS, Bons, Zubereitungsstatus — das ist doch auch ein eigener Bereich?"

**DEV1 (Anna):** „Gute Frage. Die Ausgabestationen haben eine eigene Perspektive: Sie lesen Bestelldaten aus dem Kassenbetrieb und stellen sie für Küche und Getränkeausgabe dar. Der Bondruck und das KDS konsumieren Kassenbetrieb-Events, haben aber eigene Darstellungs- und Steuerungslogik."

**DEV2 (Tim):** „Und wenn wir den Zubereitungsstatus dazunehmen — ‚in Zubereitung', ‚fertig' —, dann hat dieser Bereich sogar eigene Schreiboperationen, nicht nur Lese-Ansichten."

**DOM (Rudi):** „Die Sprache in der Küche ist auch anders: Da redet keiner von ‚Saldo' oder ‚Zahlung'. Die reden von ‚Bestellungen', ‚Positionen' und ‚fertig'."

**DEV1 (Anna):** „Genau. Das ist ein Indiz für einen eigenen Bounded Context: Gleiche Daten, aber andere Perspektive und andere Sprache. Ich schlage vor, wir nennen ihn **Ausgabe**."

**KAS (Eva):** „Und die Abrechnungen? Tagesabrechnung, Umsatz pro Servicekraft, Produktumsatz, Datenexport — das ist auch ein eigener Bereich, oder?"

**DEV2 (Tim):** „Absolut. Die Abrechnung konsumiert die Events aus dem Kassenbetrieb, aggregiert sie aber auf eine ganz andere Art: tischübergreifend, nach Servicekraft, nach Produkt, nach Zeitraum. Die Sprache dort ist: Umsatz, Gesamtumsatz, Stornoquote, Export, Tagesabschluss."

**DEV1 (Anna):** „Die Abrechnung liest nur — sie schreibt keine Events. Sie projiziert die Event-Daten in Auswertungsformen. Das ist klassischerweise ein Read-Side-Context."

**FAC (Lisa):** „Also vier fachliche Bereiche. Was ist mit Auth — Login, Logout, Passwort?"

**DEV1 (Anna):** „Auth ist kein fachlicher Bounded Context. Das ist **technische Infrastruktur** — es hat keine eigene Fachsprache, keine eigenen Geschäftsregeln. Es stellt sicher, dass nur berechtigte Benutzer zugreifen können, aber es versteht nichts von Bestellungen, Produkten oder Abrechnungen."

**ADM (Thomas):** „Stimmt. Der Login ist Mittel zum Zweck — damit Jonas auch wirklich Jonas ist und nicht jemand anders."

**DEV2 (Tim):** „Auth liefert dem Rest des Systems die Identität und die Rolle des Benutzers — über ein Token. Es ist ein Dienst, den alle Bereiche nutzen, aber er gehört zu keiner Fachdomäne."

**FAC (Lisa):** „Also fünf Bereiche insgesamt — vier fachliche und einer technisch:"

```
Identifizierte Bereiche:

1. KASSENBETRIEB — Tisch-basierte Vorgänge: Bestellen, Liefern, Bezahlen, Stornieren
   Sprache: Bestellung, Position, Lieferung, Zahlung, Stornierung, Saldo, Kassenjournal
   Persistenz: Event-Sourcing

2. STAMMDATEN — Verwaltung von Produkten, Tischen, Benutzern
   Sprache: Produkt, Variante, Kategorie, Preis, Rolle, aktiv/deaktiviert
   Persistenz: CRUD

3. AUSGABE — Bondruck, Küchendisplay, Zubereitungsstatus
   Sprache: Bon, Kategorie (Essen/Getränke), Zubereitungsstatus, offen/in Zubereitung/fertig
   Persistenz: Event-getrieben (konsumiert Kassenbetrieb-Events)

4. ABRECHNUNG — Tagesabrechnung, Umsatzauswertungen, Datenexport
   Sprache: Umsatz, Gesamtumsatz, Stornoquote, Export, Tagesabschluss
   Persistenz: Read-only (Projektionen über Kassenbetrieb-Events)

5. AUTH — Login, Logout, Passwort, Session-Verwaltung
   Sprache: Benutzer, Passwort, Token, Rolle
   Persistenz: Infrastruktur (kein fachlicher Context)
```

### 6.2 Context-Klassifikation

**FAC (Lisa):** „Jetzt ordnen wir diese Bereiche ein. In DDD unterscheiden wir drei Typen: **Core Domain** — das, was uns einzigartig macht und den größten Aufwand verdient. **Supporting Sub-Domain** — notwendig für den Betrieb, aber kein Alleinstellungsmerkmal. Und **Generic Sub-Domain** — Standardfunktionalität, die fast jedes System braucht."

**DEV1 (Anna):** „Der Kassenbetrieb ist eindeutig unsere **Core Domain**. Das ist das Herzstück von jotti — die Tischbestellungen per Smartphone, die Event-basierte Abrechnung, das Kassenjournal. Das ist der Grund, warum Vereine jotti nutzen."

**DOM (Rudi):** „Genau. Das unterscheidet uns von einer Excel-Tabelle oder einem Papierblock."

**DEV2 (Tim):** „Die Stammdaten sind **Supporting**. Wir brauchen sie — Produkte, Tische, Benutzer müssen verwaltet werden. Aber das ist keine Raketenwissenschaft. Jedes CRUD-Framework kann das."

**ADM (Thomas):** „Von meiner Seite reicht da eine solide Verwaltungsoberfläche. Nichts Ausgefallenes."

**KAS (Eva):** „Und die Abrechnung?"

**DEV1 (Anna):** „Auch **Supporting**. Wichtig für den Verein und die Buchhaltung, aber es ist im Wesentlichen eine Auswertung der Kassenbetrieb-Daten. Die Logik ist: Events lesen, aggregieren, anzeigen."

**KAS (Eva):** „Für mich fühlt sich das ziemlich zentral an — aber ich verstehe das Argument. Die Innovation steckt im Kassenbetrieb, nicht in der Auswertung."

**VER1 (Klaus):** „Die Ausgabestationen?"

**DEV2 (Tim):** „**Supporting**. Das KDS und der Zubereitungsstatus sind für den Betrieb nützlich, aber sie konsumieren nur Kassenbetrieb-Daten und stellen sie für die Küche dar. Der Bondruck ist zusätzlich von externen Druckern abhängig."

**DEV1 (Anna):** „Wobei der Bondruck speziell eher **Generic** ist — Drucker-Integration ist Commodity, das machen tausende Systeme. Aber im Moment klammern wir das vielleicht nicht zu fein auf. Die gesamte Ausgabe als Supporting reicht."

**FAC (Lisa):** „Und Auth?"

**DEV2 (Tim):** „**Generic**. Standard JWT-Authentifizierung. Login, Token-Erstellung, Passwort-Hashing. Das ist technisches Handwerk, keine Fachdomäne."

**SRL (Felix):** „Solange es funktioniert und sicher ist, interessiert mich die Technik nicht."

**FAC (Lisa):** _(notiert auf einem Sticky)_

```
Context-Klassifikation:

| Context        | Typ                    | Begründung                                              |
|----------------|------------------------|---------------------------------------------------------|
| Kassenbetrieb  | Core Domain            | Alleinstellungsmerkmal: Vereins-mPOS per Smartphone     |
| Stammdaten     | Supporting Sub-Domain  | Notwendig, aber generisch — Produkt/Tisch/Benutzer-CRUD |
| Ausgabe        | Supporting Sub-Domain  | KDS und Zubereitungsstatus für den Betrieb nützlich     |
| Abrechnung     | Supporting Sub-Domain  | Auswertung der Kassenbetrieb-Events, wichtig für Verein |
| Auth           | Generic Sub-Domain     | Standard-Authentifizierung, keine Fachdomäne            |
```

### 6.3 Context Map

**FAC (Lisa):** „Letzte Frage: Wie kommunizieren diese Bereiche miteinander? Wer liefert Daten, wer konsumiert sie?"

**DEV1 (Anna):** „Fangen wir mit den klaren Abhängigkeiten an. Der **Kassenbetrieb** braucht Daten aus den **Stammdaten** — Produktkatalog, Tischliste, Benutzerdaten. Aber er schreibt nie in die Stammdaten zurück. Das ist eine Einbahnstraße: Stammdaten sind Upstream, Kassenbetrieb ist Downstream."

**DEV2 (Tim):** „Im Kassenbetrieb kopieren wir die relevanten Stammdaten in die Events — Produktname, Variantenname, Preis. So sind die Events autark. Wenn sich ein Produktpreis später ändert, bleiben die alten Bestellungen korrekt."

**DOM (Rudi):** „Das ist wie bei einer Quittung: Was draufsteht, gilt — egal was danach passiert."

**DEV1 (Anna):** „Genau. Das nennt man eine **Anti-Corruption Layer** — der Kassenbetrieb schützt sich vor Änderungen in den Stammdaten, indem er die Daten zum Zeitpunkt der Bestellung einfriert."

**DEV2 (Tim):** „Nächste Beziehung: **Kassenbetrieb → Ausgabe**. Die Ausgabestationen konsumieren die Bestellungs-Events — jede neue Bestellung erscheint auf dem KDS. Der Kassenbetrieb publiziert seine Events, die Ausgabe reagiert darauf."

**VER1 (Klaus):** „Also wenn Jonas eine Bestellung aufgibt, bekomme ich die automatisch auf meinem Display?"

**DEV1 (Anna):** „Genau. Das ist eine **Published Language** — der Kassenbetrieb definiert die Event-Struktur, und die Ausgabe versteht sie."

**KAS (Eva):** „Und die Abrechnung? Die liest auch die Kassenbetrieb-Events?"

**DEV2 (Tim):** „Ja. Die **Abrechnung** ist ebenfalls Downstream des Kassenbetrieb. Sie projiziert die Events in Auswertungsformen — Tagesabrechnung, Umsatz pro Servicekraft, Produktumsatz. Auch hier: reine Leserichtung, kein Rückkanal."

**DEV1 (Anna):** „Und **Auth** steht quer: Es liefert allen anderen Bereichen die Identität und Berechtigung des Benutzers. Kassenbetrieb, Stammdaten, Ausgabe, Abrechnung — alle prüfen über Auth, ob der Benutzer zugreifen darf."

**ADM (Thomas):** „Auth ist also sowas wie ein Dienst, den alle nutzen?"

**DEV2 (Tim):** „Genau. In der Context-Map-Terminologie ist das ein **Open Host Service** — Auth stellt eine standardisierte Schnittstelle bereit (Token mit Benutzer-ID und Rolle), die alle Contexts konsumieren."

**SRV2 (Maria):** „Und die Stammdaten brauchen den Kassenbetrieb gar nicht?"

**DEV1 (Anna):** „Richtig. Die Stammdaten sind komplett unabhängig vom Kassenbetrieb. Du kannst Produkte anlegen, ohne dass eine einzige Bestellung existiert. Die Abhängigkeit geht nur in eine Richtung."

**FAC (Lisa):** „Fassen wir die Beziehungen zusammen:"

```
Context-Map — Beziehungen:

| Upstream       | Downstream     | Beziehungstyp                       | Beschreibung                                             |
|----------------|----------------|--------------------------------------|----------------------------------------------------------|
| Stammdaten     | Kassenbetrieb  | Customer/Supplier + ACL              | Kassenbetrieb liest Produkte/Tische, friert Daten in Events ein |
| Kassenbetrieb  | Ausgabe        | Published Language (Event-driven)    | Bestellungs-Events triggern KDS-Anzeige und Bon-Druck    |
| Kassenbetrieb  | Abrechnung     | Published Language (Event-driven)    | Tisch-Events werden zu Auswertungen projiziert           |
| Auth           | Kassenbetrieb  | Open Host Service                    | Token mit Benutzer-ID und Rolle                          |
| Auth           | Stammdaten     | Open Host Service                    | Token mit Benutzer-ID und Rolle                          |
| Auth           | Ausgabe        | Open Host Service                    | Token mit Benutzer-ID und Rolle                          |
| Auth           | Abrechnung     | Open Host Service                    | Token mit Benutzer-ID und Rolle                          |
```

**FAC (Lisa):** „Und jetzt das Ganze als Diagramm an der Wand."

**DEV2 (Tim):** _(zeichnet auf ein großes Blatt Papier)_

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                           jotti — Context Map                                │
│                                                                              │
│  ┌──────────────────────────────────────────┐                                │
│  │  KASSENBETRIEB (Core Domain)             │                                │
│  │                                          │                                │
│  │  Sprache: Bestellung, Position,          │                                │
│  │  Lieferung, Zahlung, Stornierung,        │                                │
│  │  Saldo, Kassenjournal                    │                                │
│  │                                          │                                │
│  │  Aggregat: Tisch (Event-Sourcing)        │                                │
│  │  Events: BestellungAufgegeben,           │                                │
│  │  ProdukteGeliefert, ZahlungRegistriert,  │                                │
│  │  ProdukteStorniert                       │                                │
│  └──────┬──────────────┬────────────────────┘                                │
│         │              │                                                     │
│         │ Published    │ Published                                            │
│         │ Language     │ Language                                             │
│         ▼              ▼                                                     │
│  ┌──────────────┐  ┌───────────────────┐                                     │
│  │  AUSGABE     │  │  ABRECHNUNG       │                                     │
│  │  (Supporting)│  │  (Supporting)     │                                     │
│  │              │  │                   │                                     │
│  │  KDS, Bons,  │  │  Tagesabrechnung, │                                     │
│  │  Zuberei-    │  │  Umsatzberichte,  │                                     │
│  │  tungsstatus │  │  Datenexport      │                                     │
│  └──────────────┘  └───────────────────┘                                     │
│                                                                              │
│  ┌──────────────────────────────────────────┐                                │
│  │  STAMMDATEN (Supporting Sub-Domain)      │                                │
│  │                                          │                                │
│  │  Sprache: Produkt, Variante, Kategorie,  │                                │
│  │  Preis, Tisch (Stamm), Benutzer, Rolle   │                                │
│  │                                          │                                │
│  │  Aggregate: Produkt, Tisch, Benutzer     │                                │
│  │  Persistenz: CRUD                        │                                │
│  └──────────────┬───────────────────────────┘                                │
│                 │ Customer/Supplier + ACL                                     │
│                 ▼                                                             │
│         KASSENBETRIEB (liest Produkte/Tische, friert Daten in Events ein)    │
│                                                                              │
│  ┌──────────────────────────────────────────┐                                │
│  │  AUTH (Generic — Infrastruktur)          │                                │
│  │                                          │                                │
│  │  Login, Logout, Passwort, Token          │                                │
│  │  → Open Host Service für alle Contexts   │                                │
│  └──────────────────────────────────────────┘                                │
└──────────────────────────────────────────────────────────────────────────────┘
```

**SRV1 (Jonas):** „Also der Kassenbetrieb ist in der Mitte — alles dreht sich darum?"

**DEV1 (Anna):** „Genau. Er ist die Core Domain und die Datenquelle für Ausgabe und Abrechnung. Die Stammdaten liefern ihm die Grunddaten, Auth schützt den Zugang."

**DOM (Rudi):** „Und die Grenzen sind klar: Jeder Bereich hat seine eigene Sprache, seine eigenen Regeln, seine eigene Art, Daten zu speichern."

**KAS (Eva):** „Mir gefällt, dass die Abrechnung ein eigener Bereich ist. So kann man die Auswertungen weiterentwickeln, ohne den Kassenbetrieb anzufassen."

**DEV2 (Tim):** „Das ist genau der Punkt. Die Bounded Contexts ermöglichen unabhängige Weiterentwicklung. Wenn wir ein neues Reporting-Feature brauchen, ändern wir nur den Abrechnung-Context — der Kassenbetrieb bleibt unberührt."

**FAC (Lisa):** „Wunderbar. Damit haben wir unsere Domain Map fertig — fünf Bereiche, klare Grenzen, klare Kommunikationswege. Weiter geht's mit den offenen Fragen!"

---

## 7. Hotspots und offene Fragen

**FAC (Lisa):** „In den letzten Stunden haben wir an vielen Stellen rote Kärtchen geklebt — ❤️ Hotspots, wo Unsicherheit oder Diskussionsbedarf besteht. Jetzt nehmen wir uns die Zeit, alle Hotspots systematisch durchzugehen, zusammenzufassen und offene Fragen zu formulieren."

**DEV1 (Anna):** „Ich sehe acht rote Kärtchen an der Wand."

**FAC (Lisa):** „Genau. Gehen wir sie der Reihe nach durch."

### 7.1 Tischumbuchung (K-08)

**FAC (Lisa):** „Der erste Hotspot kam in Phase 1 auf: Tischumbuchung — eine Bestellung nachträglich auf einen anderen Tisch umbuchen."

**SRV1 (Jonas):** „Das passiert mir ständig. Gast setzt sich um, oder ich tippe den falschen Tisch ein. Dann muss ich stornieren und neu bestellen. Nervig."

**DEV1 (Anna):** „Die Herausforderung ist die Atomarität: Am Quell-Tisch muss eine Stornierung entstehen, am Ziel-Tisch eine neue Bestellung — und beides muss in einer einzigen Transaktion passieren. Wenn nur eins von beiden klappt, haben wir inkonsistente Daten."

**DEV2 (Tim):** „Das heißt: zwei Events auf zwei verschiedenen Aggregaten in einer Transaktion. Das ist bei Event-Sourcing nicht trivial — normalerweise ändert ein Command nur ein Aggregat."

**SRL (Felix):** „Wer darf das überhaupt? Ich würde sagen, nur Serviceleitung und Admin. Die normale Servicekraft soll das nicht können — zu viel Fehlerpotenzial."

**DOM (Rudi):** „Einverstanden. Umbuchung ist ein Korrekturvorgang, kein Standardprozess."

**KAS (Eva):** „Für die Abrechnung muss klar nachvollziehbar sein, dass es eine Umbuchung war — nicht einfach eine Stornierung plus eine neue Bestellung."

**DEV1 (Anna):** „Wir könnten einen eigenen Event-Typ dafür definieren — dann ist die Semantik klar. Oder wir erzeugen explizit markierte Stornierung und Neubestellung mit einem Umbuchungs-Verweis."

**Offene Fragen:**

- Wie wird Atomarität über zwei Aggregate hinweg sichergestellt?
- Eigener Event-Typ oder Kombination aus Stornierung + Neubestellung mit Verweis?
- Wird der Kommentar „umgebucht von/auf Tisch X" automatisch gesetzt?
- Nur Serviceleitung und Admin, oder auch Servicekraft?

**Priorität:** Nice-to-have (K-08)

---

### 7.2 Bondruck und Drucker-Integration (K-11)

**FAC (Lisa):** „Dritter Hotspot: Bondruck und Drucker-Integration — die Hardware-Seite."

**VER1 (Klaus):** „Ohne Bon geht bei uns nichts. Wenn eine Bestellung reinkommt, muss der Bon automatisch aus dem Drucker kommen — an der Bar für Getränke, in der Küche für Essen."

**ADM (Thomas):** „Und ich muss konfigurieren können, welcher Drucker welche Kategorie bekommt. Und was ist, wenn der Drucker offline ist?"

**DEV2 (Tim):** „Die erste Frage ist das Druckerprotokoll. ESC/POS ist der Standard bei Bondruckern. Aber der Browser hat keinen direkten Zugriff auf USB-Drucker — jedenfalls nicht ohne Weiteres."

**DEV1 (Anna):** „Es gibt verschiedene Ansätze: WebUSB, ein lokaler Print-Agent, oder ein netzwerkfähiger Drucker, den das Backend direkt anspricht. Jeder Ansatz hat eigene Komplexität."

**SRV1 (Jonas):** „Und wenn der Drucker klemmt oder das Papier leer ist? Die Bestellung muss trotzdem durchgehen."

**DEV2 (Tim):** „Unbedingt. Der Bon-Druck darf den Bestellvorgang nie blockieren. Die Policy ist: asynchron drucken, Fire-and-Forget mit Retry. Wenn der Drucker nicht erreichbar ist, wird die Bestellung trotzdem gespeichert."

**DOM (Rudi):** „Und der Nachdruck? Wenn ein Bon verloren geht, muss man ihn einzeln nachdrucken können."

**DEV1 (Anna):** „Nachdruck ist eine reine Darstellung — die Bestelldaten sind im System, der Bon wird daraus neu generiert. Kein neues Event nötig."

**Offene Fragen:**

- Welches Druckerprotokoll? ESC/POS, WebUSB, lokaler Print-Agent, Netzwerkdrucker?
- Wie wird die Drucker-Konfiguration verwaltet? (Admin-Oberfläche, Zuordnung Drucker ↔ Kategorie)
- Fehlerbehandlung: Was passiert bei Druckerausfall? (Retry-Logik, Fehlermeldung an Servicekraft)
- Nachdruck einzelner Positionen oder ganzer Bons?

**Priorität:** Should-have (K-11)

---

### 7.3 KDS-Architektur (K-12)

**FAC (Lisa):** „Vierter Hotspot: Das Küchendisplay — wie kommen die Bestelldaten in Echtzeit zur Küche und zur Getränkeausgabe?"

**VER1 (Klaus):** „Für mich ist die Kernfrage: Wenn Jonas eine Bestellung aufgibt, wie schnell sehe ich die auf meinem Display? Sofort? Nach ein paar Sekunden? Muss ich die Seite neu laden?"

**DEV1 (Anna):** „Das ist eine Architekturfrage. Es gibt drei gängige Ansätze: Polling — das Frontend fragt alle paar Sekunden nach neuen Daten. WebSockets — eine permanente Verbindung, über die der Server neue Events pusht. Oder Server-Sent Events — ähnlich wie WebSockets, aber einfacher, nur in eine Richtung."

**DEV2 (Tim):** „Polling ist am einfachsten zu implementieren, aber ineffizient — viele leere Anfragen. WebSockets sind bidirektional, was wir gar nicht brauchen, da das KDS nur liest. Server-Sent Events wären der beste Kompromiss: Server schickt Updates, Client empfängt."

**SRV2 (Maria):** „Was passiert bei einem Netzwerkausfall am Display? Gehen dann Bestellungen verloren?"

**DEV1 (Anna):** „Die Bestellungen sind im Backend sicher gespeichert. Das Display ist nur eine Ansicht. Wenn die Verbindung wiederkommt, wird die aktuelle Liste neu geladen."

**ADM (Thomas):** „Brauchen wir dafür neue Infrastruktur? Einen WebSocket-Server oder so?"

**DEV2 (Tim):** „Das hängt vom Ansatz ab. Server-Sent Events laufen über HTTP — das geht mit unserem bestehenden Backend. WebSockets bräuchten eine eigene Handling-Schicht."

**Offene Fragen:**

- Echtzeit-Mechanismus: Polling, WebSockets oder Server-Sent Events?
- Wie wird der initiale Zustand geladen, wenn das Display geöffnet wird?
- Reconnect-Verhalten bei Verbindungsabbruch?
- Performance bei vielen gleichzeitigen Bestellungen?

**Priorität:** Should-have (K-12)

---

### 7.4 Zubereitungsstatus-Modellierung (K-13)

**FAC (Lisa):** „Fünfter Hotspot: Der Zubereitungsstatus — wie wird er technisch modelliert?"

**DEV2 (Tim):** „Wir hatten in Phase 3 und Phase 4 die Diskussion: Sind das fachliche Domain Events, die im Kassenjournal landen? Oder ist das transienter UI-State, der bei einem Seitenrefresh verloren geht?"

**VER1 (Klaus):** „Für mich an der Ausgabe ist der Status essenziell. Ich markiere ‚in Zubereitung' und später ‚fertig'. Wenn das bei einem Refresh weg ist, bringt es nichts."

**DEV1 (Anna):** „Argumente für Events: Persistenz, Nachvollziehbarkeit, Analysemöglichkeit — wie lange dauert die Zubereitung im Schnitt? Argumente dagegen: Es sind sehr kurzlebige Informationen, die den Event-Stream aufblähen."

**SRV1 (Jonas):** „Mir ist die Nachvollziehbarkeit egal — ich will nur sehen, ob meine Bestellung fertig ist. Aber wenn der Status nach einem Refresh weg ist, bringt das ganze Feature nichts."

**DEV2 (Tim):** „Dann noch die Frage der Aggregat-Zuordnung: Gehört der Zubereitungsstatus zum Tisch-Aggregat? Oder ist er eigenständig — ein eigener kleiner Zustand pro Ausgabestation?"

**DEV1 (Anna):** „Beim Tisch-Aggregat muss er dann alle Invarianten mittragen. Das Event-Replay wird komplexer. Eigenständig wäre sauberer, aber wir hätten ein weiteres Aggregat."

**KAS (Eva):** „Hat der Zubereitungsstatus Einfluss auf die Abrechnung?"

**DEV2 (Tim):** „Nein. Der Saldo ändert sich nicht. Es ist rein operativ — für den Betriebsablauf."

**Offene Fragen:**

- Domain Events oder transienter UI-State?
- Falls Events: Gehören sie zum Tisch-Aggregat oder zu einem eigenen Aggregat?
- Wie wird der Status beim Seitenrefresh wiederhergestellt? (nur relevant bei Event-Lösung)
- Echtzeit-Benachrichtigung an Servicekräfte, wenn Status sich ändert? (gleiche Infrastruktur wie KDS)

**Priorität:** Nice-to-have (K-13)

---

### 7.5 Offline-Fähigkeit (Q-05)

**FAC (Lisa):** „Sechster Hotspot: Offline-Fähigkeit — was passiert, wenn während des Festes das Netz ausfällt?"

**SRV1 (Jonas):** „Das ist das Worst-Case-Szenario. Mitten im Service kein Netz — dann steht alles still."

**DOM (Rudi):** „Uns ist das schon zweimal passiert. Einmal hat der Router einfach aufgehört. Dann ging gar nichts mehr."

**DEV1 (Anna):** „Echte Offline-Fähigkeit ist technisch sehr aufwändig: PWA mit Service Worker, lokaler Event-Store in IndexedDB, bidirektionale Synchronisation, Konfliktauflösung bei Wiederverbindung. Das ist ein eigenes Projekt."

**DEV2 (Tim):** „Die Kernfrage ist: Wie löst man Konflikte, wenn zwei Servicekräfte offline am selben Tisch gearbeitet haben? Wessen Bestellung gewinnt? Was passiert mit doppelten Zahlungen?"

**ADM (Thomas):** „Gibt es eine einfachere Lösung? Ohne das volle Programm?"

**DEV1 (Anna):** „Ein pragmatischer Kompromiss: Optimistisches UI — die Eingabe wird sofort angezeigt, und bei Netzwerkproblemen kommt eine Fehlermeldung. Dazu ein sichtbarer Netzwerk-Status-Indikator. Keine echte Offline-Zwischenspeicherung, aber der Benutzer merkt sofort, wenn etwas nicht klappt."

**SRV2 (Maria):** „Und was machen wir dann bei einem Ausfall? Zurück zu Stift und Papier?"

**DOM (Rudi):** „Leider ja — als Rückfallebene. Aber mit einem stabilen lokalen Netz (eigener Router, kein Internet nötig) reduziert sich das Risiko erheblich."

**Offene Fragen:**

- Echte Offline-Fähigkeit (PWA, IndexedDB, Sync) oder nur optimistisches UI mit Fehlermeldung?
- Falls offline: Wie werden Konflikte bei Wiederverbindung aufgelöst?
- Wie wird der Netzwerkstatus dem Benutzer angezeigt?
- Empfehlung für stabile lokale Netzwerk-Infrastruktur (eigener Access Point)?

**Priorität:** Nice-to-have (Q-05)

---

### 7.6 Tagesabschluss mit offenen Tischen (R-06)

**FAC (Lisa):** „Siebter Hotspot: Der Tagesabschluss — was passiert, wenn am Ende des Abends noch Tische einen offenen Saldo haben?"

**KAS (Eva):** „Das passiert jedes Mal. Gäste verschwinden, ohne zu zahlen. Oder jemand hat vergessen, eine Zahlung einzutragen. Am Ende stimmt die Kasse nicht."

**DOM (Rudi):** „Bisher schreiben wir das als Verlust ab. Oder der Kassenwart legt den Differenzbetrag privat drauf. Beides ist unbefriedigend."

**SRL (Felix):** „Es muss eine Möglichkeit geben, den offenen Saldo manuell auszugleichen — mit Begründung."

**DEV2 (Tim):** „Ein Command ‚Tisch manuell schließen' — Admin only. Erzeugt eine automatische Stornierung oder ein Differenz-Event für den offenen Betrag. Mit Pflicht-Kommentar: Warum wird manuell geschlossen?"

**DEV1 (Anna):** „Und dann die zweite Frage: Soll das System nach dem Tagesabschluss zurückgesetzt werden? Also alle Tische wieder auf Saldo 0, bereit für die nächste Veranstaltung?"

**KAS (Eva):** „Ja, unbedingt. Beim nächsten Fest will ich mit sauberen Tischen starten."

**DEV2 (Tim):** „Aber das widerspricht dem Append-only-Prinzip unseres Event-Sourcings. Events löschen geht nicht. Wir könnten die Events archivieren — in einen separaten Speicher verschieben — oder einfach pro Veranstaltung einen neuen Kontext aufmachen."

**DEV1 (Anna):** „Am saubersten wäre: Eine Veranstaltung ist ein logischer Rahmen. Der Tagesabschluss markiert das Ende einer Veranstaltung. Die Events bleiben erhalten, werden aber beim nächsten Fest nicht mehr berücksichtigt."

**ADM (Thomas):** „Muss ich dann vor jedem Fest eine neue ‚Veranstaltung' anlegen?"

**DEV2 (Tim):** „Das wäre eine Option. Oder der Tagesabschluss setzt einen Marker, und alle neuen Bestellungen gehören automatisch zur nächsten Veranstaltung."

**Offene Fragen:**

- Wie werden offene Saldi beim Tagesabschluss behandelt? (Manuelle Stornierung, Differenz-Event, Verlust-Buchung)
- Wie wird das System für die nächste Veranstaltung zurückgesetzt, ohne Events zu löschen?
- Braucht es ein Konzept „Veranstaltung" als logischen Rahmen?
- Voraussetzungen für den Tagesabschluss: Müssen alle Tische ausgeglichen sein?

**Priorität:** Nice-to-have (R-06)

---

### 7.7 Reporting-Aggregation (R-01 bis R-05)

**FAC (Lisa):** „Achter und letzter Hotspot: Wie werden die Reporting-Auswertungen aus den Rohdaten berechnet?"

**KAS (Eva):** „Die Frage ist: Wenn ich die Tagesabrechnung aufrufe, wird die dann in Echtzeit aus allen Events berechnet? Oder laufen die Berechnungen vorab?"

**DEV1 (Anna):** „Es gibt zwei Ansätze: On-the-fly-Berechnung — bei jedem Abruf werden alle relevanten Events durchlaufen und aggregiert. Oder vorberechnete Projektionen — die Auswertungen werden bei jedem neuen Event aktualisiert und stehen sofort bereit."

**DEV2 (Tim):** „On-the-fly ist einfacher zu implementieren, kann aber bei vielen Events und Tischen langsam werden. Vorberechnete Projektionen sind schneller beim Abruf, aber aufwändiger im Aufbau."

**DOM (Rudi):** „Wie viele Events haben wir an einem Festabend? Größenordnung?"

**DEV1 (Anna):** „Überschlag: 30 Tische, je 10 Bestellungen, dazu Zahlungen, Lieferungen, Stornierungen — vielleicht 500 bis 2000 Events am Abend. Das ist überschaubar."

**KAS (Eva):** „Dann reicht wahrscheinlich on-the-fly? Das sind ja keine Millionen Datensätze."

**DEV2 (Tim):** „Für die Größenordnung eines Vereinsfestes: ja, on-the-fly sollte performant genug sein. PostgreSQL kann 2000 Zeilen in Millisekunden aggregieren."

**DEV1 (Anna):** „Trotzdem sollten wir das im Auge behalten. Wenn wir merken, dass es langsam wird, können wir immer noch auf Projektionen wechseln — die Event-Sourcing-Architektur erlaubt das jederzeit."

**SRV2 (Maria):** „Sehe ich als Servicekraft auch Auswertungen?"

**KAS (Eva):** „Die Tagesabrechnung und Detailberichte sind nur für den Admin. Servicekräfte sehen nur ihre eigenen Tische."

**Offene Fragen:**

- On-the-fly-Aggregation oder vorberechnete Projektionen?
- Reicht SQL-Aggregation über die Event-Tabelle, oder brauchen wir separate Reporting-Tabellen?
- Wie wird der Zeitraum für Auswertungen definiert? (pro Tag, pro Veranstaltung, frei wählbar)
- Performance-Schwellenwerte: Ab wann lohnt sich der Wechsel zu Projektionen?

**Priorität:** Should-have (R-01 bis R-05)

---

**FAC (Lisa):** „Damit haben wir alle acht Hotspots dokumentiert. Das sind die offenen Fragen, die bei der Implementierung geklärt werden müssen — entweder durch Prototypen, technische Spikes oder Priorisierungsentscheidungen."

**DEV1 (Anna):** „Wichtig: Kein Hotspot blockiert den Start. Die Must-have-Features — Bestellung, Zahlung, Lieferung, Stornierung, Stammdaten — sind alle klar. Die Hotspots betreffen weiterführende Features oder Architekturentscheidungen."

**DEV2 (Tim):** „Genau. Wir können mit dem Kern beginnen und die Hotspots iterativ angehen — priorisiert nach Should-have und Nice-to-have."

**DOM (Rudi):** „Das beruhigt mich. Die Grundfunktionen fürs nächste Vereinsfest stehen — alles andere kann wachsen."

**FAC (Lisa):** „Gut gesagt. Dann gehen wir jetzt zu den Ergebnissen und nächsten Schritten."

---

## 8. Ergebnisse und nächste Schritte

**FAC (Lisa):** „Wir haben die letzten fünf Stunden intensiv gearbeitet — von den ersten chaotischen Events über Cluster, Prozesse und Software Design bis zu den Bounded Contexts und Hotspots. Jetzt ist es Zeit, unsere Ergebnisse zusammenzufassen. Was sind die fünf wichtigsten Erkenntnisse, die wir heute gewonnen haben?"

### 8.1 Gemeinsames Verständnis — Fünf Kernaussagen

**FAC (Lisa):** „Lasst uns die fünf Kernaussagen formulieren, auf die wir uns heute geeinigt haben. Das sind die Punkte, die jeder im Team kennen und mittragen sollte."

**DEV1 (Anna):** „Erstens: **Der Tisch ist das zentrale Aggregat im Kassenbetrieb.** Alle Operationen — Bestellen, Liefern, Bezahlen, Stornieren — verändern den Zustand eines Tisches. Der Saldo berechnet sich ausschließlich aus dem Event Stream. Kein manuelles Saldo-Tracking, keine redundante Buchführung."

**DOM (Rudi):** „Das spiegelt auch unsere Realität wider. Beim Fest denkt jeder in Tischen: ‚Tisch 7 hat noch nicht bezahlt', ‚Tisch 12 braucht die Getränke'. Der Tisch ist die natürliche Einheit."

**DEV2 (Tim):** „Zweitens: **Event-Sourcing für den Kassenbetrieb, CRUD für Stammdaten.** Nicht alles braucht Events. Produkte, Tische und Benutzer werden klassisch angelegt und geändert — mit Soft-Deletes statt physischem Löschen. Nur der Kassenbetrieb — wo eine lückenlose, unveränderliche Historie geschäftskritisch ist — nutzt Event-Sourcing."

**KAS (Eva):** „Das leuchtet mir ein. Für meine Abrechnung brauche ich die komplette Kette: Wer hat was bestellt, wann geliefert, wann bezahlt. Da darf nichts fehlen und nichts nachträglich verändert werden."

**SRL (Felix):** „Drittens: **Stornierung braucht ein klares Berechtigungsmodell.** Nur Serviceleitung und Admin dürfen stornieren. Normale Servicekräfte nicht. Das schützt vor versehentlichen oder missbräuchlichen Stornierungen — und das ist bei ehrenamtlichen Teams mit unterschiedlicher Erfahrung enorm wichtig."

**SRV2 (Maria):** „Finde ich gut. Ich bin neu dabei und hätte bei meinem ersten Fest Angst, versehentlich etwas Falsches zu stornieren. So ist klar: Wenn ich einen Fehler gemacht habe, gehe ich zu Felix."

**DEV1 (Anna):** „Viertens: **Fat Events sichern die historische Treue.** Jedes Event im Kassenjournal enthält alle relevanten Daten zum Zeitpunkt der Aktion — Produktname, Variantenname, Einzelpreis. Wenn der Admin danach einen Preis ändert, bleiben alte Bestellungen mit dem ursprünglichen Preis im Journal. Die Events sind ihre eigene Quelle der Wahrheit."

**ADM (Thomas):** „Das heißt, ich kann Preise während des Fests anpassen, und es gibt kein Durcheinander bei der Abrechnung?"

**DEV1 (Anna):** „Genau. Das Anti-Corruption Layer zwischen Stammdaten und Kassenbetrieb sorgt dafür: Die Events frieren den Zustand zum Bestellzeitpunkt ein."

**FAC (Lisa):** „Und die fünfte Kernaussage?"

**DEV2 (Tim):** „Fünftens: **Fünf Bounded Contexts mit klarer Verantwortung.** Kassenbetrieb als Core Domain — das ist unser Alleinstellungsmerkmal. Stammdaten und Ausgabe als Supporting Sub-Domains, die den Kassenbetrieb unterstützen. Abrechnung als weitere Supporting Sub-Domain, die Events zu Auswertungen projiziert. Und Auth als generische Infrastruktur, die nur Tokens bereitstellt."

**DOM (Rudi):** „Was mir dabei wichtig ist: Jeder Context hat seine eigene Sprache. ‚Tisch' im Kassenbetrieb ist etwas anderes als ‚Tisch' in der Stammdatenverwaltung. Im Kassenbetrieb ist ein Tisch ein lebendiges Aggregat mit Saldo und Geschichte. In den Stammdaten ist er ein Eintrag mit Name und Status."

**FAC (Lisa):** „Perfekt. Fünf Kernaussagen, die wir alle mittragen:

> 1. **Der Tisch ist das zentrale Aggregat** — alle Kassenbetrieb-Operationen verändern den Zustand eines Tisches, der Saldo berechnet sich aus dem Event Stream.
> 2. **Event-Sourcing für den Kassenbetrieb, CRUD für Stammdaten** — die Persistenzstrategie folgt den fachlichen Anforderungen, nicht einem technischen Dogma.
> 3. **Stornierung erfordert erhöhte Berechtigung** — nur Serviceleitung und Admin dürfen stornieren, um Missbrauch und Fehler zu verhindern.
> 4. **Fat Events sichern die historische Treue** — jedes Event enthält alle relevanten Daten zum Zeitpunkt der Aktion, unabhängig von späteren Stammdaten-Änderungen.
> 5. **Fünf Bounded Contexts mit klarer Verantwortung** — Kassenbetrieb (Core), Stammdaten (Supporting), Ausgabe (Supporting), Abrechnung (Supporting), Auth (Generic)."

### 8.2 Priorisierte Erkenntnisse

**FAC (Lisa):** „Jetzt priorisieren wir unsere Erkenntnisse. Welche Anforderungen sind Must-have — also unverzichtbar für den ersten Einsatz beim Vereinsfest? Welche sind Should-have — wichtig, aber nicht blockierend? Und welche sind Nice-to-have — Komfort, der wachsen kann?"

**DOM (Rudi):** „Ohne Bestellen, Liefern, Bezahlen und Stornieren geht gar nichts. Das ist unser Kerngeschäft."

**KAS (Eva):** „Und ohne Tagesabrechnung stehe ich am Ende des Abends im Dunkeln. Zumindest eine Grundversion davon brauche ich."

**ADM (Thomas):** „Produkte, Tische und Benutzer muss ich anlegen können, sonst kann niemand arbeiten. Und Login muss funktionieren."

**DEV2 (Tim):** „Und Mehrbenutzerfähigkeit, Validierung, Datenintegrität — das sind technische Grundvoraussetzungen, keine optionalen Features."

**FAC (Lisa):** „Gut. Dann ordnen wir alle Anforderungen in die Prioritäten ein:"

| Priorität        | ID   | Anforderung                             | Erkenntnis aus der Session                                                |
| ---------------- | ---- | --------------------------------------- | ------------------------------------------------------------------------- |
| **Must-have**    | K-01 | Bestellung aufgeben                     | Zentrales Command, erzeugt `BestellungAufgegeben`-Event im Tisch-Aggregat |
| **Must-have**    | K-02 | Zahlung registrieren                    | Teilzahlung möglich, reduziert Saldo, Positionen einzeln auswählbar       |
| **Must-have**    | K-03 | Lieferung bestätigen                    | Nachverfolgung offener Positionen, wichtig für Servicekräfte              |
| **Must-have**    | K-04 | Stornierung                             | Nur `serviceleitung` / `admin`, mit Pflichtbegründung                     |
| **Must-have**    | K-05 | Tischübersicht und Navigation           | Startseite mit Tischkarten, Tap-Navigation zum Detail                     |
| **Must-have**    | K-06 | Kassenjournal (Historie)                | Unveränderlicher Event Stream pro Tisch, Quelle der Wahrheit              |
| **Must-have**    | S-01 | Produktverwaltung                       | CRUD mit Soft-Delete, Varianten mit Cent-Preisen                          |
| **Must-have**    | S-02 | Tischverwaltung                         | CRUD mit Soft-Delete, nur aktive Tische im Service sichtbar               |
| **Must-have**    | S-03 | Benutzerverwaltung                      | Drei Rollen (`admin`, `serviceleitung`, `service`), Soft-Delete           |
| **Must-have**    | A-01 | Login                                   | JWT-basiert, Argon2id, generische Fehlermeldung                           |
| **Must-have**    | A-02 | Passwort setzen                         | Einmalpasswort bei Erstanmeldung, dann eigenes Passwort                   |
| **Must-have**    | A-03 | Logout                                  | Token entfernen, Weiterleitung auf Login                                  |
| **Must-have**    | Q-01 | Usability und Mobile-first              | BYOD, Touch-optimiert, Drawer-Konzept für Tischoperationen                |
| **Must-have**    | Q-02 | Mehrbenutzerfähigkeit                   | Parallele Zugriffe auf verschiedene und gleiche Tische                    |
| **Must-have**    | Q-03 | Validierung                             | Frontend (Zod) + Backend (zog), doppelte Absicherung                      |
| **Must-have**    | Q-04 | Datenintegrität                         | Transaktionen, append-only Events, Cent-Beträge, Soft-Deletes             |
| **Must-have**    | Q-06 | HTTPS / TLS                             | Let's Encrypt, nginx Reverse Proxy                                        |
| **Should-have**  | K-11 | Bondruck                                | Automatisch nach Bestellung, getrennt nach Kategorie                      |
| **Should-have**  | K-12 | Küchendisplay (KDS)                     | Echtzeit-Anzeige nach Kategorie, gruppiert nach Tisch                     |
| **Should-have**  | Q-07 | Rate Limiting                           | Login-Endpunkt absichern gegen Brute-Force                                |
| **Should-have**  | Q-08 | Security Headers                        | CSP, HSTS, X-Frame-Options, X-Content-Type-Options                        |
| **Should-have**  | R-01 | Tagesabrechnung                         | Gesamtumsatz, Stornierungen, jederzeit abrufbar                           |
| **Should-have**  | R-03 | Abrechnung pro Tisch                    | Detaillierte Aufstellung pro Tisch, chronologisch                         |
| **Should-have**  | R-04 | Abrechnung pro Servicekraft             | Umsatz und Stornierungen pro Person, Transparenz                          |
| **Should-have**  | R-05 | Produktumsatz-Reporting                 | Verkaufte Mengen, Ranking, Einnahmen pro Produkt                          |
| **Nice-to-have** | K-08 | Bestellungen umbuchen                   | Hotspot: 2-Aggregat-Transaktion, Atomarität klären                        |
| **Nice-to-have** | K-09 | Rückgeldberechnung                      | Reiner Anzeigeaspekt, clientseitige Berechnung                            |
| **Nice-to-have** | K-10 | Tisch-Schnellsuche                      | Suchfeld/Nummernpad für schnelle Navigation                               |
| **Nice-to-have** | K-13 | Ausgabestationen mit Zubereitungsstatus | Hotspot: Domain Events oder UI-State? Aggregat-Zuordnung klären           |
| **Nice-to-have** | Q-05 | Offline-Fähigkeit                       | Hotspot: PWA, lokale Persistierung, Synchronisation                       |
| **Nice-to-have** | R-02 | Datenexport                             | CSV-Export für Vereinsbuchhaltung                                         |
| **Nice-to-have** | R-06 | Tagesabschluss                          | Hotspot: Offene Tische, Archivierung, Veranstaltungs-Konzept              |

### 8.3 Ubiquitous Language — Terminologie aus der Session

**FAC (Lisa):** „In den letzten Stunden haben wir eine gemeinsame Sprache entwickelt — eine Ubiquitous Language, die im Team einheitlich verwendet werden soll. Lasst uns die wichtigsten Begriffe festhalten, wie sie in der Session entstanden sind."

**DEV1 (Anna):** „Alle Begriffe der Fachdomäne sind deutsch. Das war eine bewusste Entscheidung: Die Domäne ist deutsch, die Benutzer sind deutsch, die Fachbegriffe sollen das widerspiegeln. Infrastruktur-Begriffe wie Token oder Login bleiben englisch."

**FAC (Lisa):** „Genau. Hier die Terminologie, wie sie in unserer Session gewachsen ist:"

#### Kassenbetrieb (Core Domain)

| Begriff           | Bedeutung                                                                                                                              |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| **Tisch**         | Zentrales Aggregat im Kassenbetrieb. Trägt einen Event Stream, aus dem sich der aktuelle Zustand (Saldo, offene Positionen) berechnet. |
| **Bestellung**    | Ein Vorgang, bei dem eine Servicekraft Positionen für einen Tisch aufgibt. Erzeugt ein `BestellungAufgegeben`-Event.                   |
| **Position**      | Ein einzelner Posten innerhalb einer Bestellung: Produktvariante + Menge + Einzelpreis.                                                |
| **Lieferung**     | Die Bestätigung, dass bestellte Positionen dem Gast übergeben wurden. Erzeugt ein `ProdukteGeliefert`-Event.                           |
| **Zahlung**       | Die Registrierung einer Barzahlung. Kann sich auf einzelne Positionen beziehen (Teilzahlung). Erzeugt ein `ZahlungRegistriert`-Event.  |
| **Stornierung**   | Die nachträgliche Aufhebung bestellter Positionen. Nur durch Serviceleitung oder Admin. Erzeugt ein `ProdukteStorniert`-Event.         |
| **Saldo**         | Der offene Betrag eines Tisches: Summe der Bestellungen minus Summe der Zahlungen minus Summe der Stornierungen. Immer in Cent.        |
| **Kassenjournal** | Der vollständige, unveränderliche Event Stream eines Tisches. Enthält alle Operationen in chronologischer Reihenfolge.                 |
| **Kommentar**     | Optionale Freitextnotiz zu einer Bestellung, Zahlung, Lieferung oder Stornierung (max. 100 Zeichen).                                   |

#### Stammdaten (Supporting Sub-Domain)

| Begriff         | Bedeutung                                                                                                                     |
| --------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| **Produkt**     | Ein Artikel im Produktkatalog (z. B. „Bratwurst", „Radler"). Gehört zu einer Kategorie.                                       |
| **Variante**    | Eine konkrete Ausprägung eines Produkts mit eigenem Namen und Preis in Cent (z. B. „Halbe 0,5 l" für 3,50 €).                 |
| **Kategorie**   | Gruppierung von Produkten: Essen, Getränke oder Sonstiges. Bestimmt die Zuordnung zu Ausgabestationen.                        |
| **Preis**       | Immer ganzzahlig in Cent. Niemals Fließkommazahlen. 3,50 € = 350 Cent.                                                        |
| **Soft-Delete** | Logisches Löschen durch Status-Änderung auf „deleted". Der Datensatz bleibt erhalten, ist aber im aktiven Betrieb unsichtbar. |

#### Authentifizierung (Generic Sub-Domain)

| Begriff            | Bedeutung                                                                                  |
| ------------------ | ------------------------------------------------------------------------------------------ |
| **Rolle**          | Berechtigungsstufe eines Benutzers: `admin`, `serviceleitung` oder `service`.              |
| **Einmalpasswort** | Vom Admin generiertes 6-stelliges Passwort für die Erstanmeldung oder nach Passwort-Reset. |
| **Token**          | JWT mit Benutzer-ID und Rolle, 12 Stunden gültig. Wird bei jedem API-Aufruf mitgesendet.   |

#### Ausgabe (Supporting Sub-Domain)

| Begriff                 | Bedeutung                                                                                                       |
| ----------------------- | --------------------------------------------------------------------------------------------------------------- |
| **Bon**                 | Gedruckter Beleg mit Tisch, Servicekraft, Positionen, Mengen, Zeitstempel und optionalem Kommentar.             |
| **Küchendisplay (KDS)** | Echtzeit-Anzeige offener Bestellungen an der Ausgabestation, gruppiert nach Tisch und gefiltert nach Kategorie. |
| **Zubereitungsstatus**  | Status einer Position an der Ausgabestation: offen → in Zubereitung → fertig.                                   |
| **Ausgabestation**      | Physischer Ort (Küche, Getränketheke), an dem Positionen zubereitet und ausgegeben werden.                      |

#### Abrechnung (Supporting Sub-Domain)

| Begriff             | Bedeutung                                                                                                                              |
| ------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| **Tagesabrechnung** | Übersicht über Gesamtumsatz, Stornierungen und Umsatz pro Servicekraft — jederzeit vom Admin abrufbar.                                 |
| **Umsatz**          | Summe aller registrierten Zahlungen in einem bestimmten Zeitraum. In Cent.                                                             |
| **Stornoquote**     | Verhältnis von Stornierungsbetrag zu Bestellsumme. Indikator für Fehler oder Unregelmäßigkeiten.                                       |
| **Tagesabschluss**  | Administrativer Vorgang zum Ende einer Veranstaltung: offene Tische prüfen, Abschlussbericht generieren, optional System zurücksetzen. |
| **Export**          | CSV-Download von Umsätzen, Bestellungen und Artikeldaten für die Vereinsbuchhaltung.                                                   |

#### Übergreifende Prinzipien

| Prinzip                         | Bedeutung                                                                                                                                              |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Event-Sourcing**              | Persistenzmuster für den Kassenbetrieb: Zustand wird nicht direkt gespeichert, sondern aus unveränderlichen Events berechnet.                          |
| **Fat Event**                   | Event, das alle relevanten Daten zum Zeitpunkt der Aktion enthält (inkl. Produktname, Preis) — unabhängig von späteren Stammdaten-Änderungen.          |
| **Anti-Corruption Layer (ACL)** | Schutzmechanismus zwischen Bounded Contexts: Der Kassenbetrieb friert Stammdaten in Events ein und ist damit unabhängig von nachträglichen Änderungen. |
| **Append-only**                 | Grundprinzip des Event Streams: Events werden nur hinzugefügt, nie geändert oder gelöscht.                                                             |
| **BYOD**                        | Bring Your Own Device — Servicekräfte nutzen ihre eigenen Smartphones.                                                                                 |

### 8.4 Feedback der Teilnehmer

**FAC (Lisa):** „Zum Abschluss möchte ich eine kurze Runde machen. Was nehmt ihr aus der heutigen Session mit? Was hat euch überrascht, was bestätigt?"

**DOM (Rudi):** „Ich war skeptisch, ob ein paar bunte Zettel an der Wand uns wirklich weiterbringen. Aber als die Events an der Timeline hingen und wir die Cluster gesehen haben — da wurde die Struktur plötzlich sichtbar. Der Tisch als zentrales Element, die drei Phasen Vorbereitung-Betrieb-Abschluss — das ist genau, wie ein Vereinsfest abläuft."

**SRV1 (Jonas):** „Für mich war die Prozess-Modellierung am hilfreichsten. Zu sehen, welche Schritte zwischen ‚Gast bestellt' und ‚Geld in der Kasse' liegen, hat mir gezeigt, wie viel eigentlich passiert, das wir bisher im Kopf erledigen. Und dass Stornierung eine eigene Berechtigung braucht — macht absolut Sinn."

**SRV2 (Maria):** „Als Neuling hatte ich Sorge, dass ich den technischen Diskussionen nicht folgen kann. Aber die Events in normaler Sprache — ‚Bestellung aufgegeben', ‚Zahlung registriert' — das versteht jeder. Gut, dass wir Deutsch gewählt haben."

**DEV1 (Anna):** „Mich hat überrascht, wie natürlich die Trennung zwischen Event-Sourcing und CRUD aus den Anforderungen heraus entstanden ist. Wir haben nicht mit einer Architekturentscheidung angefangen, sondern mit den fachlichen Abläufen — und dann kam das Ergebnis von allein."

**DEV2 (Tim):** „Die Context Map war für mich der wichtigste Moment. Fünf Bounded Contexts, klare Kommunikationswege, Published Language für das Event-driven Zusammenspiel. Und dass Auth generisch ist — das spart uns Komplexität in der Domäne."

**SRL (Felix):** „Mir war wichtig, dass Stornierung nicht leichtfertig gehandhabt wird. Die Diskussion über Berechtigungen hat gezeigt, dass alle verstehen, warum das eingeschränkt sein muss. Das ist kein Misstrauen — das ist Schutz für die Ehrenamtlichen."

**ADM (Thomas):** „Ich nehme mit, dass das System für mich als Admin beherrschbar bleibt. Produkte, Tische, Benutzer — das ist klar und überschaubar. Und dass ich Preise während des Fests ändern kann, ohne alte Bestellungen zu verfälschen, ist beruhigend."

**KAS (Eva):** „Die Reporting-Diskussion war für mich zentral. Dass wir Tagesabrechnung, Abrechnung pro Tisch, pro Servicekraft und Produktumsatz als eigene Anforderungen herausgearbeitet haben — das deckt genau das ab, was ich nach dem Fest für den Kassenbericht brauche. Und der CSV-Export für die Buchhaltung ist das Sahnehäubchen."

**VER1 (Klaus):** „Endlich wurde auch an die Leute hinter der Theke gedacht! Das Küchendisplay und der Zubereitungsstatus — das sind Funktionen, die den Abend so viel entspannter machen. Keine verlorenen Bons mehr, keine ‚Wer hatte nochmal die Pommes für Tisch 8?'-Momente."

**FAC (Lisa):** „Danke an alle. Wir haben heute in fünf Stunden die gesamte Domäne durchgearbeitet — von rohen Events über Prozesse und Aggregate bis zu Bounded Contexts. Wir haben eine gemeinsame Sprache, eine klare Priorisierung und acht konkrete Hotspots, die wir iterativ angehen können. Das ist eine solide Grundlage für die Entwicklung."

---

## Anhang A — Vollständige Event-Liste

Diese Liste fasst alle in der Session identifizierten Domain Events, Commands, Akteure und Policies zusammen — gruppiert nach Bounded Context und geordnet nach dem typischen Ablauf einer Veranstaltung.

### Kassenbetrieb (Core Domain) — Tisch-Aggregat, Event-Sourcing

| #   | Domain Event                       | Command                                 | Akteur                                | Anforderung | Policy / Nachfolge                                       |
| --- | ---------------------------------- | --------------------------------------- | ------------------------------------- | ----------- | -------------------------------------------------------- |
| 1   | 🟠 **BestellungAufgegeben**        | Bestellung aufgeben                     | Servicekraft / Serviceleitung / Admin | K-01, K-07  | Bon-Druck nach Kategorie (K-11); KDS aktualisiert (K-12) |
| 2   | 🟠 **ProdukteGeliefert**           | Lieferung bestätigen                    | Servicekraft / Serviceleitung / Admin | K-03        | Ungelieferte Positionen aktualisiert                     |
| 3   | 🟠 **ZahlungRegistriert**          | Zahlung registrieren                    | Servicekraft / Serviceleitung / Admin | K-02        | Saldo aktualisiert; Tischkonto ggf. ausgeglichen         |
| 4   | 🟠 **ProdukteStorniert**           | Positionen stornieren                   | Serviceleitung / Admin                | K-04        | 🟣 Nur `serviceleitung` / `admin`; Saldo aktualisiert    |
| 5   | 🟠 PositionInZubereitungGesetzt ❤️ | Position als „in Zubereitung" markieren | Alle angemeldeten Benutzer            | K-13        | Hotspot: Domain Event oder UI-State?                     |
| 6   | 🟠 PositionFertigGemeldet ❤️       | Position als „fertig" markieren         | Alle angemeldeten Benutzer            | K-13        | Hotspot: Domain Event oder UI-State?                     |

**Invarianten des Tisch-Aggregats:**

- Saldo = Σ(Bestellungen) − Σ(Zahlungen) − Σ(Stornierungen)
- Nur bestellte Positionen können geliefert, bezahlt oder storniert werden
- Stornierung nur durch Rollen `serviceleitung` und `admin`
- Alle Beträge in Cent (int), niemals Floats
- Events sind append-only — nie ändern oder löschen

**Abgeleitete Zustände (keine eigenen Events):**

- Tischkonto ausgeglichen — entsteht wenn Saldo = 0
- Alle Positionen geliefert — entsteht wenn für jede bestellte Position ein Liefer-Event existiert

**Nicht formalisiert (Hotspot):**

- Bestellung umgebucht (K-08) — 2-Aggregat-Transaktion: Storno am Quell-Tisch + Bestellung am Ziel-Tisch

### Stammdaten (Supporting Sub-Domain) — CRUD, kein Event-Sourcing

| #   | Vorgang                     | Command                                  | Akteur | Anforderung | Anmerkung                                             |
| --- | --------------------------- | ---------------------------------------- | ------ | ----------- | ----------------------------------------------------- |
| 8   | Produkt hinzugefügt         | Produkt anlegen                          | Admin  | S-01        | Kategorie: Essen / Getränke / Sonstiges               |
| 9   | Produktvariante hinzugefügt | Produktvariante anlegen                  | Admin  | S-01        | Eigener Name und Preis in Cent                        |
| 10  | Produkt bearbeitet          | Produkt bearbeiten                       | Admin  | S-01        | Name, Kategorie                                       |
| 11  | Produktpreis geändert       | Produktpreis ändern                      | Admin  | S-01        | Nur für zukünftige Bestellungen                       |
| 12  | Produkt deaktiviert         | Produkt deaktivieren                     | Admin  | S-01        | Soft-Delete — historische Bestellungen bleiben gültig |
| 13  | Tisch angelegt              | Tisch anlegen                            | Admin  | S-02        | Auch im laufenden Betrieb möglich                     |
| 14  | Tisch bearbeitet            | Tisch bearbeiten                         | Admin  | S-02        | Name ändern                                           |
| 15  | Tisch deaktiviert           | Tisch deaktivieren                       | Admin  | S-02        | Soft-Delete — Events bleiben im System                |
| 16  | Benutzer angelegt           | Benutzer anlegen                         | Admin  | S-03        | Name, Benutzername, Rolle                             |
| 17  | Einmalpasswort ausgestellt  | Benutzer anlegen / Passwort zurücksetzen | Admin  | S-03        | 6-stellig, für Erstanmeldung oder Reset               |
| 18  | Benutzer bearbeitet         | Benutzer bearbeiten                      | Admin  | S-03        | Name, Benutzername, Rolle                             |
| 19  | Benutzer deaktiviert        | Benutzer deaktivieren                    | Admin  | S-03        | Soft-Delete — Account gesperrt                        |

### Ausgabe (Supporting Sub-Domain) — Event-getrieben, konsumiert Kassenbetrieb-Events

| #   | Vorgang                  | Auslöser                             | Akteur | Anforderung | Anmerkung                                              |
| --- | ------------------------ | ------------------------------------ | ------ | ----------- | ------------------------------------------------------ |
| 20  | Küchenbon gedruckt       | 🟣 Policy: nach BestellungAufgegeben | System | K-11        | Essenspositionen → Küchendrucker                       |
| 21  | Getränkebon gedruckt     | 🟣 Policy: nach BestellungAufgegeben | System | K-11        | Getränkepositionen → Getränkedrucker                   |
| 22  | KDS-Ansicht aktualisiert | 🟣 Policy: nach BestellungAufgegeben | System | K-12        | Offene Positionen nach Kategorie, gruppiert nach Tisch |

### Abrechnung (Supporting Sub-Domain) — Read-only, Projektionen über Kassenbetrieb-Events

| #   | Read Model                  | Command                              | Akteur | Anforderung | Quelle                                               |
| --- | --------------------------- | ------------------------------------ | ------ | ----------- | ---------------------------------------------------- |
| 23  | Tagesabrechnung             | Tagesabrechnung anzeigen             | Admin  | R-01        | Alle Tisch-Events (tischübergreifend)                |
| 24  | Abrechnung pro Tisch        | Abrechnung pro Tisch anzeigen        | Admin  | R-03        | Tisch-Events (einzelner Tisch)                       |
| 25  | Abrechnung pro Servicekraft | Abrechnung pro Servicekraft anzeigen | Admin  | R-04        | Tisch-Events (gruppiert nach Akteur)                 |
| 26  | Produktumsatz               | Produktumsatz anzeigen               | Admin  | R-05        | BestellungAufgegeben-Events (gruppiert nach Produkt) |
| 27  | CSV-Export                  | Daten exportieren                    | Admin  | R-02        | Alle Reporting-Daten                                 |
| 28  | Tagesabschluss ❤️           | Tagesabschluss einleiten             | Admin  | R-06        | Hotspot: Offene Tische, Archivierung                 |

### Auth (Generic Sub-Domain) — Infrastruktur, keine fachlichen Domain Events

| #   | Vorgang             | Command         | Akteur        | Anforderung | Anmerkung                            |
| --- | ------------------- | --------------- | ------------- | ----------- | ------------------------------------ |
| 29  | Benutzer eingeloggt | Anmelden        | Alle Benutzer | A-01        | JWT ausgestellt (12 h gültig)        |
| 30  | Passwort gesetzt    | Passwort setzen | Alle Benutzer | A-02        | Ersetzt Einmalpasswort durch eigenes |
| 31  | Benutzer ausgeloggt | Abmelden        | Alle Benutzer | A-03        | Token invalidiert                    |

### Zuordnung: Weitere Anforderungen ohne eigene Events

| Anforderung | Beschreibung                  | Abdeckung in der Session                                                 |
| ----------- | ----------------------------- | ------------------------------------------------------------------------ |
| K-05        | Tischübersicht und Navigation | 🟢 Read Model: Tischübersicht (Projektion aus Tisch-Events + Stammdaten) |
| K-06        | Kassenjournal (Historie)      | 🟢 Read Model: Kassenjournal (Event-Stream in menschenlesbarer Form)     |
| K-08        | Bestellungen umbuchen         | ❤️ Hotspot: 2-Aggregat-Transaktion, noch nicht formalisiert              |
| K-09        | Rückgeldberechnung            | Reine UI-Logik (Frontend), kein Backend-Event                            |
| K-10        | Tisch-Schnellsuche            | UI-Filter auf Tischübersicht, kein Event                                 |
| Q-01        | Mobile-first / BYOD           | Architektur-Prinzip — kein Event                                         |
| Q-02        | Mehrbenutzerfähigkeit         | Architektur-Prinzip — parallele Zugriffe auf Tisch-Aggregat              |
| Q-03        | Validierung (Zod + zog)       | Querschnittliche Regel — Frontend + Backend                              |
| Q-04        | Datenintegrität               | Transaktionen, append-only Events, Cent-Beträge, Soft-Deletes            |
| Q-05        | Offline-Fähigkeit             | ❤️ Hotspot: PWA, lokale Persistierung, Synchronisation                   |
| Q-06        | HTTPS / TLS                   | Infrastruktur: Let's Encrypt, nginx Reverse Proxy                        |
| Q-07        | Rate Limiting                 | Infrastruktur: Login-Endpunkt absichern                                  |
| Q-08        | Security Headers              | Infrastruktur: CSP, HSTS, X-Frame-Options                                |

---

## Anhang B — Stickies-Legende

| Farbe     | Symbol | Bedeutung                                                                | Verwendung in der Session                                                                                            |
| --------- | ------ | ------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------- |
| 🟠 Orange | —      | **Domain Event** — etwas, das in der Domäne passiert ist (Vergangenheit) | Zentral: Kassenjournal, Event-Sourcing, Bestellungen, Zahlungen, Stornierungen                                       |
| 🔵 Blau   | —      | **Command** — Absicht, etwas zu tun (Imperativ)                          | Auslöser für Geschäftslogik: „Bestellung aufgeben", „Zahlung registrieren"                                           |
| 🟡 Gelb   | —      | **Aggregate** — transaktionale Grenze, schützt Geschäftsregeln           | Tisch-Aggregat (Kassenbetrieb), Produkt/Tisch/Benutzer (Stammdaten)                                                  |
| 🟣 Lila   | —      | **Policy** — automatische Reaktion: „Wenn X, dann Y"                     | Bon-Druck nach Bestellung, KDS-Aktualisierung, Stornierungsberechtigung                                              |
| 🟢 Grün   | —      | **Read Model** — Lese-Sicht / Projektion für die Anzeige                 | Tischübersicht, Kassenjournal, KDS-Ansicht, Tagesabrechnung, Reporting                                               |
| ❤️ Rot    | —      | **Hotspot** — Unsicherheit, Diskussionsbedarf, offene Frage              | 7 Hotspots: Umbuchung, Bondruck, KDS-Architektur, Zubereitungsstatus, Offline, Tagesabschluss, Reporting-Aggregation |

---

_Dieses Dokument basiert auf der Event-Storming-Session vom 11. März 2026. Alle genannten Personen sind fiktiv. Die Domänenerkenntnisse spiegeln die Anforderungen von jotti wider._
