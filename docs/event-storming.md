# Event Storming — jotti

Dieses Dokument protokolliert das Ergebnis einer simulierten Event-Storming-Session für jotti. Ziel war es, die gesamte Fachdomäne eines Vereins-Gastronomie-POS-Systems gemeinsam zu erkunden, Domain Events zu identifizieren, Bounded Contexts abzugrenzen und eine gemeinsame Sprache zu entwickeln.

> **Methode:** Event Storming nach Alberto Brandolini — vom Big Picture über Process Modelling bis zum Software Design Level.

---

## Inhaltsverzeichnis

1. [Setup & Teilnehmer](#1-setup--teilnehmer)
2. [Phase 1 — Big Picture: Chaotische Exploration](#2-phase-1--big-picture-chaotische-exploration)
3. [Phase 2 — Clustering und Pivot Points](#3-phase-2--clustering-und-pivot-points)
4. [Phase 3 — Process Modelling: Commands, Akteure, Policies](#4-phase-3--process-modelling-commands-akteure-policies)
5. [Phase 4 — Software Design: Aggregate und Read Models](#5-phase-4--software-design-aggregate-und-read-models)
6. [Bounded Contexts und Domain Map](#6-bounded-contexts-und-domain-map)
7. [Hotspots und offene Fragen](#7-hotspots-und offene-fragen)
8. [Ergebnisse und nächste Schritte](#8-ergebnisse-und-nächste-schritte)

---

## 1. Setup & Teilnehmer

### 1.1 Rahmenbedingungen

| Attribut      | Wert                                                          |
| ------------- | ------------------------------------------------------------- |
| **Datum**     | Samstag, 9. März 2026, 10:00–17:00 Uhr                       |
| **Ort**       | Vereinsheim Sportverein Grüntal, Besprechungsraum             |
| **Dauer**     | ca. 6 Stunden (inkl. Pausen)                                  |
| **Methode**   | Event Storming (Big Picture → Process Modelling → Software Design) |
| **Material**  | Papierwand (3 m × 1,5 m), Haftnotizen in 6 Farben, Marker    |
| **Notation**  | Domain Events: 🟠 orange · Commands: 🔵 blau · Akteure: 🟡 gelb · Read Models: 🟢 grün · Policies: 🟣 lila · External Systems: 🌸 rosa · Hotspots: ❤️ rot |

### 1.2 Teilnehmer

| Kürzel | Rolle                     | Hintergrund                                                              |
| ------ | ------------------------- | ------------------------------------------------------------------------ |
| **FAC** | Facilitator (Lisa)        | Event-Storming-Erfahrung, neutral, moderiert die Session                 |
| **SCR** | Scrum Master (Markus)     | Agile Coach, hält Timeboxen ein, schützt den Prozess                     |
| **DEV1** | Developer — Backend (Anna) | Go-Entwicklerin, kennt den Event-Sourcing-Stack                         |
| **DEV2** | Developer — Frontend (Tim) | React-Entwickler, zuständig für die Service-Oberfläche                  |
| **ARC** | Architekt (Stefan)        | Systemarchitektur, DDD-Erfahrung, kennt Aggregate-Grenzen               |
| **DOM** | Domänenexperte (Rudi)     | Langjähriger Vereinsvorstand, 15 Jahre Festorganisation                  |
| **POS** | POS-Experte (Petra)       | Betreibt Café, Erfahrung mit kommerziellen Kassensystemen (Toast, Square) |
| **VER1** | Vereinsmitglied (Klaus)   | Aktiver Helfer, übernimmt oft die Getränkeausgabe                        |
| **VER2** | Vereinsmitglied (Sandra)  | Kassiererin und Schriftführerin, erstellt die Jahresabrechnung           |
| **SRV1** | Servicekraft (Jonas)      | Bedient seit 3 Jahren Tische beim Vereinsfest                            |
| **SRV2** | Servicekraft (Maria)      | Neu dabei, erstes Vereinsfest als Servicekraft                           |
| **ADM** | Administrator (Thomas)    | Kümmert sich um Software und Hardware beim Verein                        |
| **KAS** | Kassenwart (Eva)          | Zuständig für Finanzen, will am Ende Tagesabrechnung sehen               |
| **SRL** | Serviceleitung (Felix)    | Senior-Servicekraft, darf stornieren, koordiniert das Service-Team       |

---

## 2. Phase 1 — Big Picture: Chaotische Exploration

**Zeitrahmen:** 10:00–11:30 Uhr (90 Minuten)

**FAC:** „Willkommen! Heute erkunden wir gemeinsam, was bei einer Vereinsveranstaltung passiert — aus Sicht aller Beteiligten. Schreibt auf orange Haftnotizen alle Ereignisse, die ihr für wichtig haltet. Ein Ereignis ist etwas, das **passiert ist** — in der Vergangenheitsform, auf Deutsch. Also nicht 'Bestellung aufnehmen', sondern 'Bestellung aufgegeben'. Beginnt ruhig chaotisch — kein Sortieren, einfach draufkleben."

**SCR:** „Ich stelle den Timer auf 20 Minuten. Danach besprechen wir kurz, was ihr geschrieben habt."

*(Alle beginnen, Haftnotizen zu schreiben und an die Wand zu kleben.)*

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

- 🟠 Tisch zugewiesen
- 🟠 Fehler bei Bestellung gemacht
- 🟠 Bestellung storniert
- 🟠 Tisch gewechselt
- 🟠 Servicekraft angemeldet

**DOM (Rudi)** klebt auf:

- 🟠 Veranstaltung eröffnet
- 🟠 Veranstaltung beendet
- 🟠 Kassensystem gestartet
- 🟠 Produkt vergriffen
- 🟠 Tagesabrechnung erstellt
- 🟠 Einnahmen gezählt

**KAS (Eva)** klebt auf:

- 🟠 Zahlung registriert
- 🟠 Tagesabrechnung abgeschlossen
- 🟠 Umsatz pro Servicekraft ermittelt
- 🟠 Differenz zwischen Kasse und System festgestellt
- 🟠 Kassenbon ausgedruckt

**ADM (Thomas)** klebt auf:

- 🟠 Benutzer angelegt
- 🟠 Passwort zurückgesetzt
- 🟠 Produkt hinzugefügt
- 🟠 Produktpreis geändert
- 🟠 Tisch angelegt
- 🟠 Tisch deaktiviert
- 🟠 System deployed

**SRL (Felix)** klebt auf:

- 🟠 Stornierung genehmigt
- 🟠 Bestellung auf anderen Tisch umgebucht
- 🟠 Getränkeausgabe informiert
- 🟠 Essensausgabe informiert
- 🟠 Servicekraft eingewiesen

**VER1 (Klaus)** klebt auf:

- 🟠 Getränkebon empfangen
- 🟠 Getränk zubereitet
- 🟠 Bestellung als fertig markiert
- 🟠 Ausgabestation informiert

**VER2 (Sandra)** klebt auf:

- 🟠 Abrechnung exportiert
- 🟠 Jahresbericht aktualisiert
- 🟠 Steuerbericht erstellt

**POS (Petra)** klebt auf:

- 🟠 Küchenbon gedruckt
- 🟠 Getränkebon gedruckt
- 🟠 Drucker offline gegangen
- 🟠 Netzwerk ausgefallen
- 🟠 Teilzahlung registriert

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

### 2.2 Diskussion: Was fällt auf?

**FAC:** „Toll, das ist schon eine Menge! Ich sehe einige interessante Cluster entstehen. Bevor wir sortieren: Wer hat etwas Wichtiges vergessen oder sieht etwas Fehlendes?"

**DOM (Rudi):** „Bei uns ist die Situation oft, dass Leute zwischen den Tischen wechseln. Ich vermisse 'Gast wechselt Tisch' oder dass jemand für mehrere Gruppen zahlt."

**SRV1 (Jonas):** „Genau! Und wenn jemand nur teilweise bezahlt — zum Beispiel, weil drei Leute getrennt zahlen — das passiert ständig."

**KAS (Eva):** „Das Wichtigste für mich fehlt noch: Wann wird die Kasse 'abgeschlossen'? Wir brauchen eine klare Trennung zwischen laufendem Betrieb und Tagesabschluss."

**POS (Petra):** „In meinem Café trennen wir das strikt: während des Betriebs kann jeder Tisch seinen Status haben, aber beim Tagesabschluss wird alles auf null gesetzt. Für den Verein braucht ihr das wahrscheinlich auch."

*(Weitere Haftnotizen werden ergänzt:)*

- 🟠 Teilzahlung registriert
- 🟠 Tischkonto ausgeglichen
- 🟠 Gast hat getrennt gezahlt
- 🟠 Tagesabschluss durchgeführt
- 🟠 Kasse auf null gesetzt
- 🟠 Kassenstand erfasst

**SRL (Felix):** „Wir haben auch das Thema Freibon vergessen — manchmal haben Gäste besondere Wünsche, die nicht im System sind. Und dann macht jemand einfach einen Preis auf."

**ARC (Stefan):** „Das ist ein Hotspot — ich klebe da mal ein rotes Kärtchen dran. Der Freibon wirft Fragen zur Datenintegrität auf."

*(❤️ Hotspot: „Freibon — freie Preiseingabe ohne Produkt-Zuordnung")*

**DEV1 (Anna):** „Die System-Events wie 'JWT ausgestellt' und 'Event gespeichert' sind eher technische Interna — die gehören vielleicht nicht auf die Domain-Wand?"

**FAC:** „Guter Hinweis, Anna. Wir unterscheiden Domain Events von System Events. Lass uns die technischen Events zur Seite legen — wir notieren sie, aber sie formen nicht unsere Domain Map."

*(Technische Events werden an den Rand gestellt:)*

- Technische Events (separat): JWT ausgestellt · Session abgelaufen · Rate-Limit überschritten · Snapshot erstellt · Event gespeichert

---

## 3. Phase 2 — Clustering und Pivot Points

**Zeitrahmen:** 11:30–13:00 Uhr (90 Minuten, inkl. Mittagspause 12:00–12:30)

### 3.1 Timeline erstellen

**FAC:** „Jetzt bringen wir die Events in eine zeitliche Reihenfolge. Von links nach rechts: Was passiert zuerst, was danach? Nicht perfekt — ungefähr."

*(Alle beginnen gemeinsam, die Haftnotizen zu verschieben. Diskussionen entstehen.)*

**DOM (Rudi):** „Zuerst muss jemand das System aufsetzen: Produkte eingeben, Tische anlegen, Accounts erstellen."

**ADM (Thomas):** „Richtig. Das ist der Admin-Bereich. Der läuft schon Wochen vor dem Fest."

**SRV1 (Jonas):** „Dann beginnt das Fest. Wir melden uns an, sehen die Tischübersicht und fangen an zu bedienen."

**SRL (Felix):** „Und am Ende: Abrechnung, alles abschließen."

*Nach 20 Minuten ergibt sich folgende grobe Zeitlinie:*

```
[VORBEREITUNG]  →  [BETRIEB]  →  [ABSCHLUSS]
Setup           →  Kassenbetrieb  →  Tagesabrechnung
```

### 3.2 Pivot Points identifizieren

**FAC:** „Pivot Points sind Schlüsselereignisse, die einen Prozesswechsel einleiten — der Flow ändert sich danach grundlegend. Markiert sie mit einem dickeren Strich."

*Gemeinsam identifiziert:*

| Pivot Point                        | Beschreibung                                                                    |
| ---------------------------------- | ------------------------------------------------------------------------------- |
| **Veranstaltung eröffnet**         | System wechselt von Vorbereitung in aktiven Betrieb                             |
| **Bestellung aufgegeben**          | Ab hier wird getrackt; Saldo entsteht                                           |
| **Tischkonto ausgeglichen**        | Tisch ist abgeschlossen; kein Saldo mehr offen                                  |
| **Tagesabschluss durchgeführt**    | Betrieb endet; Abrechnung beginnt                                               |

**ARC (Stefan):** „Interessant: 'Bestellung aufgegeben' ist ein Pivot Point, weil ab da das Kassenjournal für diesen Tisch lebt. Alle nachfolgenden Events — Zahlung, Lieferung, Stornierung — beziehen sich auf diesen Tisch."

**DOM (Rudi):** „Bei uns läuft ein Tisch manchmal den ganzen Abend. Der ist nicht nach einer Bestellung fertig."

**KAS (Eva):** „Genau. Ein Tisch kann viele Bestellungen haben, mehrere Zahlungen — und erst wenn der Saldo null ist, ist er 'fertig'."

### 3.3 Domänen-Cluster entstehen

Nach dem Timeline-Walkthrough kristallisieren sich vier Cluster heraus:

#### Cluster A — Stammdaten-Verwaltung

*(vor dem Fest, Admin-Bereich)*

- 🟠 Benutzer angelegt
- 🟠 Benutzer bearbeitet
- 🟠 Passwort zurückgesetzt
- 🟠 Einmalpasswort ausgestellt
- 🟠 Passwort selbst gesetzt
- 🟠 Produkt hinzugefügt
- 🟠 Produktvariante hinzugefügt
- 🟠 Produktpreis geändert
- 🟠 Produkt deaktiviert
- 🟠 Produkt aktiviert
- 🟠 Tisch angelegt
- 🟠 Tisch bearbeitet
- 🟠 Tisch deaktiviert

#### Cluster B — Kassenbetrieb

*(während des Fests, Service-Bereich)*

- 🟠 Servicekraft angemeldet
- 🟠 Tisch ausgewählt
- 🟠 Produkt ausgewählt
- 🟠 Menge geändert
- 🟠 Position aus Bestellung entfernt
- 🟠 Kommentar hinzugefügt
- 🟠 Bestellung aufgegeben
- 🟠 Bestellung abgebrochen
- 🟠 Produkte ausgeliefert
- 🟠 Teilzahlung registriert
- 🟠 Zahlung registriert
- 🟠 Tischkonto ausgeglichen

#### Cluster C — Stornierung & Umbuchung

*(Service-Bereich, erhöhte Berechtigung)*

- 🟠 Stornierung beantragt
- 🟠 Stornierung genehmigt
- 🟠 Positionen storniert
- 🟠 Bestellung auf anderen Tisch umgebucht ❤️

#### Cluster D — Ausgabestationen & Bons

*(externe Systeme, zukünftige Features)*

- 🟠 Küchenbon gedruckt
- 🟠 Getränkebon gedruckt
- 🟠 Getränkeausgabe informiert
- 🟠 Essensausgabe informiert
- 🟠 Drucker offline gegangen ❤️
- 🟠 Bestellung als fertig markiert

#### Cluster E — Abrechnung

*(Tagesende, Admin/Kassenwart)*

- 🟠 Umsatz pro Servicekraft ermittelt
- 🟠 Tagesabrechnung abgeschlossen
- 🟠 Abrechnung exportiert
- 🟠 Kassenstand erfasst
- 🟠 Differenz festgestellt

---

## 4. Phase 3 — Process Modelling: Commands, Akteure, Policies

**Zeitrahmen:** 12:30–15:00 Uhr (150 Minuten)

**FAC:** „Jetzt wird es detaillierter. Zu jedem Domain Event fragen wir: Wer hat das ausgelöst? Mit welchem Befehl? Und gibt es Regeln (Policies), die das steuern?"

**Notation:**

- 🔵 **Command** (blau): Was jemand bewusst tut — Imperativ: „Bestellung aufgeben"
- 🟡 **Akteur** (gelb): Wer den Command auslöst — Person oder System
- 🟣 **Policy** (lila): Automatische Reaktion auf ein Event: „Wenn X, dann Y"
- 🌸 **External System** (rosa): Drittsystem außerhalb unseres Bounded Context

### 4.1 Kassenbetrieb — Bestellung bis Zahlung

#### Bestellung aufnehmen

```
🟡 Servicekraft
     ↓
🔵 Tisch auswählen
     ↓
🟠 Tisch ausgewählt
     ↓
🔵 Produkte auswählen und Mengen festlegen
     ↓
🟠 Bestellung aufgegeben
     ↓
🟣 Policy: Bestellung persistieren als Event "tisch.bestellung-aufgegeben:v1"
🟣 Policy: Saldo des Tisches aktualisieren
```

**SRV1 (Jonas):** „Ich wähle den Tisch aus der Liste. Dann sehe ich die Produkte und klicke auf + und −. Wenn alles stimmt, bestätige ich."

**SRV2 (Maria):** „Was ist, wenn ich den falschen Tisch gewählt habe?"

**SRL (Felix):** „Dann muss ich die Bestellung umbuchen — aber das ist Serviceleitung. Du als normale Servicekraft kannst das nicht."

**DOM (Rudi):** *(lacht)* „Bei uns ist das früher häufig passiert. Dann wurde es einfach manuell korrigiert."

**ARC (Stefan):** „Umbuchung auf anderen Tisch ist ein eigener Flow — Event: 'Bestellung umgebucht'. Kleben wir das als eigene Sequenz daran. Das ist ein ❤️ Hotspot, weil wir Atomizität brauchen: Storno am Quell-Tisch und neue Bestellung am Ziel-Tisch."

*(❤️ Hotspot: „Tischumbuchung — atomar, zwei Events in einer Transaktion")*

#### Lieferung bestätigen

```
🟡 Servicekraft
     ↓
🔵 Positionen als geliefert markieren
     ↓
🟠 Produkte ausgeliefert
     ↓
🟣 Policy: Event "tisch.produkte-geliefert:v1" persistieren
🟣 Policy: Ungelieferte Positionen des Tisches aktualisieren
```

**SRV1 (Jonas):** „Das mache ich, wenn ich das Tablett gebracht habe. Ich tippe kurz auf 'Geliefert' — das hilft mir, den Überblick zu behalten."

**POS (Petra):** „In professionellen Systemen gibt es da oft einen KDS (Kitchen Display System). Die Küche markiert die Bestellung als fertig, dann weiß die Servicekraft, dass sie abholen kann."

**VER1 (Klaus):** „Bei uns macht das keiner so aufwendig. Wir rufen einfach. Aber ein System wäre schön."

*(❤️ Hotspot: „KDS / Ausgabestation — zukünftiges Feature, noch nicht in Scope")*

#### Zahlung registrieren

```
🟡 Servicekraft
     ↓
🔵 Zahlbetrag eingeben (Gesamtbetrag oder Teilbetrag)
     ↓
🟠 Zahlung registriert
     ↓
🟣 Policy: Event "tisch.zahlung-registriert:v1" persistieren
🟣 Policy: Saldo = Bestellsumme − Zahlungen − Stornierungen neu berechnen
🟣 Policy: Wenn Saldo = 0 → Tischkonto ausgeglichen
```

**KAS (Eva):** „Hier ist es wichtig: Teilzahlungen müssen möglich sein. Wenn drei Freunde am Tisch sitzen, zahlt oft jeder separat."

**SRV1 (Jonas):** „Ja, genau. Und ich muss sehen können, was noch offen ist."

**ARC (Stefan):** „Die Saldo-Berechnung ist eine reine Projektion — Summe aller Bestellungen minus Zahlungen minus Stornierungen. Das ist State-Rekonstruktion aus dem Event Stream."

**DEV1 (Anna):** „Das haben wir schon so implementiert. `GetSaldoFromEvents` in `domain/table/events.go`."

#### Stornierung

```
🟡 Serviceleitung oder Admin
     ↓
🔵 Positionen zur Stornierung auswählen
     ↓
🔵 Stornierung bestätigen
     ↓
🟠 Positionen storniert
     ↓
🟣 Policy: Event "tisch.produkte-storniert:v1" persistieren
🟣 Policy: Saldo aktualisieren (Stornierungsbetrag abziehen)
```

**SRL (Felix):** „Ich als Serviceleitung darf stornieren. Eine normale Servicekraft nicht — das ist wichtig!"

**SRV2 (Maria):** „Das ist gut. Ich würde sonst vielleicht versehentlich etwas falsch stornieren."

**DOM (Rudi):** „Bei uns gab es mal einen Fall, wo jemand aus Versehen den ganzen Tisch storniert hat. Das war ein Chaos. Deshalb: Stornierung nur mit erhöhten Rechten."

**ARC (Stefan):** „Hier brauchen wir eine Policy: Wenn der Akteur nicht `senior_service` oder `admin` ist, wird der Command abgelehnt. Das ist Authorization-Logik auf der Middleware-Ebene."

*(🟣 Policy: Stornierung nur für Serviceleitung und Admin)*

### 4.2 Stammdaten-Verwaltung

#### Produkt verwalten

```
🟡 Admin
     ↓
🔵 Produkt erstellen / bearbeiten / deaktivieren
     ↓
🟠 Produkt hinzugefügt / geändert / deaktiviert
     ↓
🟣 Policy: Produktkatalog aktualisieren (CRUD, kein Event-Sourcing)
```

**ADM (Thomas):** „Das ist klassisches CRUD. Ich lege Produkte an, setze Preise, deaktiviere Dinge, die ausverkauft sind."

**DOM (Rudi):** „Preisänderungen während des Fests — geht das? Was passiert mit Bestellungen, die schon aufgegeben wurden?"

**ARC (Stefan):** „Gute Frage! Da die Bestellung als Event gespeichert wird, enthält das Event den Preis zum Zeitpunkt der Bestellung. Änderungen am Produkt danach beeinflussen alte Events nicht. Das ist ein Vorteil von Event-Sourcing."

**POS (Petra):** „Das ist genau richtig. In jedem guten POS-System wird der Preis im Beleg eingefroren, nicht referenziert."

*(🟣 Policy: Event-Daten sind immutable — Preisänderungen wirken nur auf zukünftige Bestellungen)*

#### Tisch verwalten

```
🟡 Admin
     ↓
🔵 Tisch erstellen / bearbeiten / deaktivieren
     ↓
🟠 Tisch angelegt / geändert / deaktiviert
     ↓
🟣 Policy: Tischliste aktualisieren (CRUD)
🟣 Policy: Deaktivierter Tisch erscheint nicht in der Tischauswahl
```

**SRV1 (Jonas):** „Manchmal werden Tische mitten im Abend dazugestellt oder weggeräumt. Kann Thomas das live ändern?"

**ADM (Thomas):** „Ja, das sollte möglich sein. Und der neue Tisch erscheint dann sofort für alle."

**ARC (Stefan):** „Deaktivierung eines Tisches ist ein Soft-Delete — der Tisch existiert weiterhin im System (für Historik), ist aber nicht mehr wählbar."

#### Benutzer verwalten

```
🟡 Admin
     ↓
🔵 Benutzer anlegen (mit Einmalpasswort)
     ↓
🟠 Benutzer angelegt
     ↓
🟣 Policy: Einmalpasswort generieren
🟣 Policy: Benutzer muss Passwort beim ersten Login ändern

🟡 Servicekraft (beim ersten Login)
     ↓
🔵 Passwort selbst setzen
     ↓
🟠 Passwort gesetzt
     ↓
🟣 Policy: JWT ausstellen
🟣 Policy: Einmalpasswort als "verbraucht" markieren
```

**SRV2 (Maria):** „Das hab ich so gemacht! Thomas hat mir ein Passwort gegeben, beim ersten Einloggen musste ich es ändern. Das fand ich gut."

**ADM (Thomas):** „Genau. Ich sehe nie die echten Passwörter — nur der Benutzer kennt sein Passwort."

**ARC (Stefan):** „Passwörter werden mit Argon2id gehasht — sicherstes Verfahren heute. Das Einmalpasswort ist nur zum Bootstrapping."

### 4.3 Abrechnung und Tagesabschluss

#### Umsatzübersicht

```
🟡 Kassenwart / Admin
     ↓
🔵 Umsatzbericht abrufen
     ↓
🟠 Umsatz pro Servicekraft ermittelt
     ↓
🟣 Policy: Events aggregieren → Umsatz je userID berechnen
```

**KAS (Eva):** „Ich brauche nach dem Fest ganz schnell eine Übersicht: Wer hat wie viel Umsatz gemacht? Wie viel haben wir insgesamt eingenommen? Gibt es Stornierungen, die auffällig sind?"

**ARC (Stefan):** „Das ist eine CQRS-Projektion — wir lesen den Event-Stream und aggregieren. Das ist noch nicht implementiert."

*(❤️ Hotspot: „Tagesabrechnung / Umsatz pro Servicekraft — Anforderung #26, noch offen")*

**KAS (Eva):** „Und ich will das als CSV exportieren können. Für die Vereinsbuchhaltung."

*(❤️ Hotspot: „CSV-Export — Anforderung #40, noch offen")*

**VER2 (Sandra):** „Für unsere Jahresabrechnung brauchen wir auch die Daten. Ein einfacher Export reicht völlig."

#### Tagesabschluss

```
🟡 Kassenwart / Admin
     ↓
🔵 Tagesabschluss einleiten
     ↓
🟠 Tagesabschluss durchgeführt
     ↓
🟣 Policy: Alle offenen Tische prüfen
🟣 Policy: Abschlussbericht generieren
🟣 Policy: System für nächste Veranstaltung zurücksetzen
```

**DOM (Rudi):** „Was passiert, wenn Tische noch offen sind am Ende? Das kommt vor — Gäste verschwinden manchmal."

**KAS (Eva):** „Die müssen wir dann manuell abschließen oder als uneinbringlich markieren."

**ARC (Stefan):** „Das ist ein offenes Design-Problem. Wir könnten einen 'Tisch schließen trotz offenem Saldo'-Command einführen. Oder der Kassenwart macht eine manuelle Stornierung auf null."

*(❤️ Hotspot: „Tagesabschluss mit offenen Tischen — Verhalten unklar")*

### 4.4 Bon-Druck und Ausgabestationen

**POS (Petra):** „Ich vermisse das Thema Bons komplett. In meinem Café geht ohne Bon-Druck gar nichts."

**DOM (Rudi):** „Bei uns ist das auch wichtig. Die Getränkeausgabe braucht einen Bon, die Küche braucht einen Bon."

**VER1 (Klaus):** „Ich stehe an der Getränkeausgabe. Wenn kein Bon kommt, weiß ich nicht, was ich machen soll."

**FAC:** „Dann fügen wir das als eigenen Prozess ein."

```
🟠 Bestellung aufgegeben
     ↓
🟣 Policy: Wenn Getränke in Bestellung → Getränkebon an Getränkedrucker
🟣 Policy: Wenn Essen in Bestellung → Küchenbon an Küchendrucker
     ↓
🟠 Getränkebon gedruckt / Küchenbon gedruckt
     ↓
🌸 External System: Thermaldrucker (Getränkeausgabe)
🌸 External System: Thermaldrucker (Küche)
```

**DEV1 (Anna):** „Bon-Druck ist noch nicht implementiert. Das braucht Drucker-Integration."

**ADM (Thomas):** „Wir haben beim letzten Fest einen alten Thermodrucker gefunden. Der könnte funktionieren."

**ARC (Stefan):** „Das ist ein zukünftiges Feature. Der Bon-Druck läuft als Side-Effect einer Policy — wenn ein Bestell-Event kommt, schickt das System einen Druckjob raus. Das ist asynchron."

*(❤️ Hotspot: „Bon-Druck — Anforderungen #27–#32, noch alle offen")*

---

## 5. Phase 4 — Software Design: Aggregate und Read Models

**Zeitrahmen:** 15:15–16:30 Uhr (75 Minuten, nach Kaffeepause)

**FAC:** „Jetzt übersetzen wir die Domain Map in Software-Konzepte. Anna und Stefan, ihr seid die Experten hier. Was sind unsere Aggregate?"

**ARC (Stefan):** „Ein Aggregat ist die transaktionale Grenze — alles, was immer konsistent zusammen geändert wird. Ich sehe drei klare Aggregate."

### 5.1 Aggregat: Tisch

**Aggregatwurzel:** `Tisch` (identifiziert durch `tisch:<id>`)

**Invarianten (Geschäftsregeln, die immer gelten müssen):**

| # | Invariante                                                           |
| - | -------------------------------------------------------------------- |
| 1 | Saldo ≥ 0 (keine Überzahlung)                                        |
| 2 | Stornierung nur für Positionen, die bestellt, aber nicht bezahlt sind |
| 3 | Lieferung nur für Positionen, die bestellt wurden                    |
| 4 | Zahlung darf Bestellsumme nicht übersteigen                          |

**Domain Events:**

| Event-Typ                            | Ausgelöst durch                     | Beschreibung                              |
| ------------------------------------ | ----------------------------------- | ----------------------------------------- |
| `tisch.bestellung-aufgegeben:v1`     | `BestellungAufgeben` (Command)      | Neue Bestellung mit Positionen und Preis  |
| `tisch.zahlung-registriert:v1`       | `ZahlungRegistrieren` (Command)     | Zahlungseingang für offene Positionen     |
| `tisch.produkte-geliefert:v1`        | `ProdukteAusliefern` (Command)      | Positionen als ausgeliefert markiert      |
| `tisch.produkte-storniert:v1`        | `ProdukteStornieren` (Command)      | Positionen storniert (erhöhte Rechte)     |
| `tisch.snapshot:v1`                  | System (Policy, periodisch)         | Performance-Optimierung, kein Domain Event |

**State-Rekonstruktion:**

```
Saldo               = Σ(Bestellungen) − Σ(Zahlungen) − Σ(Stornierungen)
UnbezahltePositionen = bestellt − bezahlt − storniert
UngeliefertePositionen = bestellt − geliefert − storniert
```

**SRV1 (Jonas):** „Ich sehe das Tisch-Aggregat wie ein Konto. Alles, was darauf läuft, ist eine Buchung."

**ARC (Stefan):** „Genau das ist es. Das Kassenjournal im DDD-Sinne."

### 5.2 Aggregat: Produkt

**Aggregatwurzel:** `Produkt` (identifiziert durch `product_id`)

**Struktur:**

```
Produkt
  ├── id
  ├── name
  ├── category (food | beverage | other)
  ├── status (active | inactive | deleted)
  └── Varianten[]
        ├── id
        ├── name
        ├── priceCents (int, NIEMALS float)
        └── status (active | inactive | deleted)
```

**Commands und Events:**

| Command                   | Event                          |
| ------------------------- | ------------------------------ |
| ProduktErstellen          | ProduktHinzugefügt             |
| ProduktBearbeiten         | ProduktGeändert                |
| ProduktDeaktivieren       | ProduktDeaktiviert             |
| VarianteHinzufügen        | VarianteHinzugefügt            |
| VarianteBearbeiten        | VarianteGeändert               |
| VarianteDeaktivieren      | VarianteDeaktiviert            |

**Persistenz:** CRUD (kein Event-Sourcing für Stammdaten)

**DOM (Rudi):** „Preise in Cent — das habt ihr mir erklärt. Kein Komma, keine Rundungsfehler."

**KAS (Eva):** „Als Kassenwart sage ich: absolut richtig. Floats bei Geld sind eine Katastrophe."

### 5.3 Aggregat: Benutzer

**Aggregatwurzel:** `Benutzer` (identifiziert durch `user_id`)

**Rollen-Hierarchie:**

```
admin
  ↑ Vollzugriff auf alles
senior_service
  ↑ Kassenbetrieb + Stornierung
service
  ↑ Kassenbetrieb (Bestellen, Liefern, Bezahlen)
```

**Commands und Events:**

| Command                   | Event                          |
| ------------------------- | ------------------------------ |
| BenutzerAnlegen           | BenutzerAngelegt               |
| BenutzerBearbeiten        | BenutzerGeändert               |
| PasswortSetzen            | PasswortGesetzt                |
| PasswortZurücksetzen      | PasswortZurückgesetzt          |

**Persistenz:** CRUD

### 5.4 Read Models (Projektionen)

Read Models sind optimierte Lese-Sichten, die aus dem Event-Stream berechnet werden. Sie dienen der Anzeigelogik im Frontend.

| Read Model                          | Konsument               | Inhalt                                                                      | Status     |
| ----------------------------------- | ----------------------- | --------------------------------------------------------------------------- | ---------- |
| **TischÜbersicht**                  | Servicekraft            | Alle aktiven Tische mit Saldo und offenem Status                            | ✅ implementiert |
| **TischDetail**                     | Servicekraft            | Ein Tisch mit Saldo, offene Positionen, Historie                            | ✅ implementiert |
| **ProduktKatalog**                  | Servicekraft            | Alle aktiven Produkte mit Varianten, nach Kategorie gruppiert               | ✅ implementiert |
| **UmsatzProServicekraft**           | Kassenwart / Admin      | Umsatz je Servicekraft für einen Zeitraum                                   | ❌ offen   |
| **TagesGesamtumsatz**               | Kassenwart / Admin      | Gesamtumsatz der Veranstaltung                                              | ❌ offen   |
| **OffeneTische**                    | Kassenwart              | Tische mit Saldo > 0 am Ende des Tages                                      | ❌ offen   |
| **StornierungsÜbersicht**           | Admin / Serviceleitung  | Alle Stornierungen mit Grund und Akteur                                     | ❌ offen   |
| **EigeneBeststellungen**            | Servicekraft            | Übersicht der eigenen Bestellungen/Tische mit Status                        | ❌ offen   |

**KAS (Eva):** „Die Read Models für die Abrechnung fehlen alle noch. Das ist mein wichtigstes Anliegen."

**DEV1 (Anna):** „Die lassen sich gut als Projektionen über den Event-Stream implementieren. CQRS-Pattern."

**ARC (Stefan):** „Genau. Commands schreiben Events. Read Models lesen Events. Getrennte Pfade."

---

## 6. Bounded Contexts und Domain Map

**Zeitrahmen:** 16:30–17:00 Uhr (30 Minuten)

### 6.1 Bounded Contexts

Nach der Session kristallisieren sich zwei Haupt-Bounded-Contexts heraus:

```
┌──────────────────────────────────────────────────────────────────────────┐
│                         jotti — Domain Map                               │
│                                                                          │
│  ┌────────────────────────────────────┐                                  │
│  │  Kassenbetrieb (service)           │                                  │
│  │                                    │                                  │
│  │  Ubiquitous Language:              │                                  │
│  │  Tisch, Bestellung, Position,      │                                  │
│  │  Lieferung, Zahlung, Stornierung,  │                                  │
│  │  Kassenjournal, Saldo              │                                  │
│  │                                    │                                  │
│  │  Persistenz: Event-Sourcing        │                                  │
│  │  Aggregate: Tisch                  │                                  │
│  └──────────────┬─────────────────────┘                                  │
│                 │ verwendet Stammdaten (readonly)                         │
│  ┌──────────────▼─────────────────────┐                                  │
│  │  Stammdaten (admin)                │                                  │
│  │                                    │                                  │
│  │  Ubiquitous Language:              │                                  │
│  │  Produkt, Variante, Kategorie,     │                                  │
│  │  Tisch (Stammdaten), Benutzer,     │                                  │
│  │  Admin, Servicekraft               │                                  │
│  │                                    │                                  │
│  │  Persistenz: CRUD (PostgreSQL)     │                                  │
│  │  Aggregates: Produkt, Benutzer     │                                  │
│  └────────────────────────────────────┘                                  │
│                                                                          │
│  ┌────────────────────────────────────┐  (zukünftig)                     │
│  │  Ausgabestationen (external)       │                                  │
│  │                                    │                                  │
│  │  Küche, Getränkeausgabe            │                                  │
│  │  Bon-Druck, KDS                    │                                  │
│  │                                    │                                  │
│  │  Persistenz: Extern (Drucker API)  │                                  │
│  └────────────────────────────────────┘                                  │
│                                                                          │
│  [Auth — technische Infrastruktur, kein Bounded Context]                 │
└──────────────────────────────────────────────────────────────────────────┘
```

### 6.2 Context Map: Beziehungen

| Upstream            | Downstream          | Beziehungstyp                          | Beschreibung                                              |
| ------------------- | ------------------- | -------------------------------------- | --------------------------------------------------------- |
| Stammdaten          | Kassenbetrieb       | Customer/Supplier + Anti-Corruption-Layer | Kassenbetrieb liest Produkte/Tische, schreibt nie dorthin |
| Auth (Infrastruktur)| Kassenbetrieb       | Open Host Service                      | JWT-Auth stellt Claims bereit (userID, role)              |
| Auth (Infrastruktur)| Stammdaten          | Open Host Service                      | Gleicher JWT-Mechanismus                                  |
| Kassenbetrieb       | Ausgabestationen    | Published Language (Event-driven)      | Events triggern Bon-Druck (zukünftig)                     |

**ARC (Stefan):** „Der Kassenbetrieb hängt von den Stammdaten ab — Produkte und Tische kommen von dort. Aber der Kassenbetrieb schreibt niemals in die Stammdaten zurück. Das ist eine Einbahnstraße."

**DEV2 (Tim):** „Im Frontend merkt man das auch: Die Produkt-API gibt den Katalog zurück, und die Bestellungs-API nimmt Varianten-IDs entgegen. Die Preise werden vom Backend beim Event-Eintrag eingefroren."

### 6.3 Sub-Domain-Klassifikation

| Sub-Domain          | Typ                 | Begründung                                                          |
| ------------------- | ------------------- | ------------------------------------------------------------------- |
| **Kassenbetrieb**   | **Core Domain**     | Das ist das Alleinstellungsmerkmal von jotti — Vereins-POS-System   |
| **Stammdaten**      | Supporting Sub-Domain | Notwendig, aber generisch — könnte fast jedes System lösen        |
| **Auth**            | Generic Sub-Domain  | Standard JWT-Auth — keine eigene Entwicklung wert                   |
| **Bon-Druck**       | Generic Sub-Domain  | Drucker-Integration ist Commodity                                   |
| **Abrechnung**      | Supporting Sub-Domain | Notwendig für Verein, aber kein Wettbewerbsmerkmal                 |

---

## 7. Hotspots und offene Fragen

Während der Session wurden folgende ❤️ Hotspots identifiziert — Bereiche, wo Unsicherheit oder Diskussionsbedarf besteht:

### 7.1 Tischumbuchung (Hotspot #1)

**Problem:** Bestellung wurde auf falschen Tisch gebucht. Servicekraft möchte auf anderen Tisch umbuchen.

**Diskussion:**

- Erfordert atomare Transaktion: Storno-Event am Quell-Tisch + neues Bestell-Event am Ziel-Tisch
- Wer darf das? Nur Serviceleitung/Admin?
- Historik: Die Stornierung erscheint im Kassenjournal des Quell-Tisches

**Vorschlag ARC (Stefan):** Neuer Event-Typ `tisch.bestellung-umgebucht:v1`. Beide Events in einer PostgreSQL-Transaktion.

**Offene Fragen:**
- Gehört Umbuchung in Scope von `service` oder `senior_service`?
- Wird der Kommentar „umgebucht auf Tisch X" automatisch gesetzt?

**Status:** ❌ Anforderung #25, noch offen

---

### 7.2 Freibon / freie Preiseingabe (Hotspot #2)

**Problem:** Sonderpositionen, die nicht im Produktkatalog stehen, müssen erfasst werden können.

**Diskussion:**

- POS-Experte Petra: „In jedem kommerziellen System gibt es das. Nötig für Sonderposten."
- Admin Thomas: „Könnte missbraucht werden — jemand gibt einfach einen falschen Preis ein."
- Kassenwart Eva: „Für mich ist das okay, wenn es in der Abrechnung klar markiert ist."

**Vorschlag DEV1 (Anna):** `variant_id = null` + freies `name`- und `priceCents`-Feld in `BestellungAufgegebenEvent.data.positionen`.

**Offene Fragen:**
- Rechteprüfung: Darf jede Servicekraft Freipreise setzen?
- Ausweisung in der Abrechnung?
- Verhindert das Inkonsistenzen im Produktkatalog?

**Status:** ❌ Anforderung #31, noch offen

---

### 7.3 Bon-Druck und Drucker-Integration (Hotspot #3)

**Problem:** Küche und Getränkeausgabe brauchen Bons. Druckerfehler dürfen den Kassenbetrieb nicht blockieren.

**Diskussion:**

- Vereinsmitglied Klaus: „Ohne Bon weiß ich nicht, was ich machen soll."
- Domänenexperte Rudi: „Früher haben wir einfach gerufen. Aber bei 15 Tischen geht das nicht."
- POS-Experte Petra: „Bon-Druck muss asynchron sein. Wenn der Drucker offline ist, geht die Bestellung trotzdem durch."

**Vorschlag ARC (Stefan):** Event-Policy: Wenn `tisch.bestellung-aufgegeben:v1` gefeuert wird, schickt ein asynchroner Handler einen Druckjob. Fire-and-forget mit Retry.

**Offene Fragen:**
- Welche Drucker-Protokolle? ESC/POS? WebUSB?
- Fallback, wenn Drucker offline?
- Getränke und Essen auf denselben Bon oder separate Bons?

**Status:** ❌ Anforderungen #27–#32, alle offen

---

### 7.4 Offline-Fähigkeit (Hotspot #4)

**Problem:** Bei Netzwerkausfall (häufig auf Vereinsfesten) soll die Bestellaufnahme weiterlaufen.

**Diskussion:**

- Servicekraft Jonas: „Das ist das Schlimmste — mitten im Abend kein Netz. Dann geht gar nichts mehr."
- Domänenexperte Rudi: „Bei uns ist das schon zweimal passiert."
- Developer Anna: „Offline-Fähigkeit ist sehr komplex — PWA mit IndexedDB, Sync-Mechanismus, Konfliktauflösung."

**Einschätzung:**

Offline-Fähigkeit mit Datenkonsistenz ist ein erheblicher Mehraufwand (Service Worker, lokaler Event-Store, bidirektionale Synchronisation). Für jotti v1 kein Scope.

**Kompromiss-Vorschlag:**
- Optimistisches UI (sofortige Anzeige bei Aktion, Fehler-Toast bei Netzwerkproblem)
- Klar sichtbarer Netzwerk-Status-Indikator in der App
- Dokumentation: jotti ist nicht offline-fähig

**Status:** ❌ Anforderung #33, noch offen — wahrscheinlich zukünftiger Scope

---

### 7.5 Tagesabschluss mit offenen Tischen (Hotspot #5)

**Problem:** Was passiert, wenn am Ende des Abends noch Tische einen offenen Saldo haben?

**Diskussion:**

- Kassenwart Eva: „Das passiert. Leute verschwinden manchmal, ohne zu zahlen."
- Domänenexperte Rudi: „Wir schreiben das als Verlust ab."
- Serviceleitung Felix: „Oder der Kassenwart macht eine manuelle Stornierung auf null."

**Vorschlag:**

Neuer Command `TischSchließen` (nur Admin), der eine automatische Stornierung der offenen Positionen erzeugt. Kommentar-Pflichtfeld: Grund für manuelle Schließung.

**Status:** ❌ Kein Anforderungseintrag — neu identifiziert in dieser Session

---

### 7.6 Tagesabrechnung und CSV-Export (Hotspot #6)

**Problem:** Kassenwart und Vereinsbuchhaltung brauchen exportierbare Umsatzzahlen.

**Diskussion:**

- Kassenwart Eva: „Excel-Export wäre toll. CSV reicht auch."
- Vereinsmitglied Sandra: „Für die Jahresabrechnung brauchen wir die Daten aus dem System."
- Scrum Master Markus: „Das ist ein klares Must-have für den Abschluss einer Veranstaltung."

**Vorschlag DEV1 (Anna):** Server-Side CSV-Generierung: Endpunkt `POST /admin/export-tagesabrechnung` gibt CSV zurück. Frontend-Download via Blob URL.

**Status:** ❌ Anforderungen #26 (Umsatz pro Bediener) und #40 (CSV-Export), noch offen

---

### 7.7 Rückgeldberechnung (Hotspot #7)

**Problem:** Servicekraft kennt den Zahlbetrag des Gastes, will das Rückgeld sehen.

**Diskussion:**

- Servicekraft Jonas: „Der Gast gibt mir einen 50-Euro-Schein, der Tisch ist 23,50 Euro. Ich rechne kurz im Kopf. Aber ein Anzeige wäre besser."
- POS-Experte Petra: „Standard-Feature in jedem Kassensystem."

**Vorschlag DEV2 (Tim):** Rein Frontend-Logik: Im ZahlungsDrawer ein optionales Feld „Gegeben" — Differenz wird live angezeigt. Keine Backend-Änderung nötig.

**Status:** ❌ Anforderung #37, noch offen

---

## 8. Ergebnisse und nächste Schritte

### 8.1 Gemeinsames Verständnis

Nach der Session teilen alle Beteiligten ein gemeinsames mentales Modell der Domäne:

1. **Tisch ist die zentrale Einheit.** Alles dreht sich um den Tisch — nicht um die Servicekraft, nicht um die Bestellung.
2. **Das Kassenjournal ist unveränderlich.** Was einmal gebucht wurde, wird nicht gelöscht, sondern korrigiert.
3. **Rollen steuern Berechtigungen.** Stornierung ist erhöhtes Recht — nicht jeder darf alles.
4. **Preise werden eingefroren.** Eine Bestellung enthält den Preis zum Zeitpunkt der Bestellung — Produktpreisänderungen wirken nicht rückwirkend.
5. **Stammdaten sind CRUD, Kassenbetrieb ist Event-Sourcing.** Bewusste Entscheidung für das richtige Werkzeug je Kontext.

### 8.2 Priorisierte Erkenntnisse

| Priorität | Erkenntnis / Feature                              | Anforderung | Nächster Schritt                               |
| --------- | ------------------------------------------------- | ----------- | ---------------------------------------------- |
| 🔴 Hoch   | Umsatz pro Servicekraft (Tagesabrechnung)         | #26         | Backend-Projektion + Frontend-Seite            |
| 🔴 Hoch   | Tisch-Schnellsuche                                | #23         | Frontend-Filter auf Tischübersicht             |
| 🟠 Mittel | Rückgeldberechnung                                | #37         | Frontend-only, ZahlungsDrawer                  |
| 🟠 Mittel | Übersicht eigene Bestellungen                     | #24         | Backend-Projektion + Frontend-Seite            |
| 🟠 Mittel | CSV-Export                                        | #40         | Backend-Export-Endpunkt                        |
| 🟡 Niedrig | Tischumbuchung                                   | #25         | Neuer Event-Typ, atomar                        |
| 🟡 Niedrig | Freibon / freie Preiseingabe                     | #31         | Neues Positions-Modell ohne Produkt-Zuordnung  |
| 🟡 Niedrig | Bon-Druck                                        | #27–#32     | Drucker-Integration, asynchrone Policy         |
| ⚪ Später  | Offline-Fähigkeit                                | #33         | PWA + Service Worker — erheblicher Aufwand     |
| ⚪ Später  | Tagesabschluss mit offenen Tischen               | —           | Neuer Command, Policy für manuellen Abschluss  |

### 8.3 Ubiquitous Language — Ergänzungen aus der Session

Folgende Begriffe wurden in der Session geprägt und sollten in [docs/language.md](language.md) ergänzt werden:

| Begriff               | Beschreibung                                                                    |
| --------------------- | ------------------------------------------------------------------------------- |
| **Kassenjournal**     | Explizit als Event Stream definiert — der unveränderliche Tisch-Event-Stream    |
| **Tischkonto**        | Mentales Modell: Ein Tisch ist wie ein Konto — mit Buchungen und Saldo          |
| **Pivot Point**       | Schlüsselereignis, das einen Prozesswechsel einleitet                           |
| **Freibon**           | Freitext-Position ohne Produkt-Zuordnung (noch nicht implementiert)             |
| **Tagesabschluss**    | Formeller Abschluss aller Tische nach der Veranstaltung                         |
| **Kassenstand**       | Physisch gezähltes Bargeld in der Kasse am Tagesende                            |
| **Ausgabestation**    | Küche oder Getränkeausgabe — empfängt Bons, markiert Bestellungen als fertig    |
| **Küchenbon**         | Bon für die Küche — enthält nur Essen-Positionen                                |
| **Getränkebon**       | Bon für die Getränkeausgabe — enthält nur Getränke-Positionen                   |

### 8.4 Feedback der Teilnehmer

**SRV1 (Jonas):** „Ich habe jetzt viel besser verstanden, warum das System so aufgebaut ist. Das mit den Events — macht Sinn! Wenn ich etwas falsches buche, bleibt es im System, aber wir stornieren es richtig."

**KAS (Eva):** „Ich freue mich auf die Tagesabrechnung. Das ist für mich das Wichtigste. Und bitte CSV-Export nicht vergessen!"

**DOM (Rudi):** „Sehr aufschlussreich. Ich hatte nie so über das System nachgedacht — jetzt verstehe ich, warum das nicht einfach eine Excel-Tabelle ist."

**POS (Petra):** „Der Bon-Druck fehlt mir noch. Aber das Grundkonzept ist solide. Für einen Verein genau das Richtige."

**VER1 (Klaus):** „Also ich an der Getränkeausgabe brauche wirklich einen Bon. Sonst weiß ich nicht, was ich machen soll. Bitte bald!"

**ARC (Stefan):** „Ich bin beeindruckt, wie klar sich die Aggregate-Grenzen herausgestellt haben. Das Tisch-Aggregat mit Event-Sourcing und die Stammdaten mit CRUD — das war die richtige Entscheidung."

**DEV1 (Anna):** „Event Storming hat mir geholfen, die Domain-Events im Code nochmal zu validieren. Unsere vier Event-Typen decken den Kern genau ab."

**SCR (Markus):** „Gutes Ergebnis für einen Tag! Die Hotspots sind klar, die Priorisierung stimmt. Nächster Sprint: Tischschnellsuche, Rückgeld, Umsatz-Projektion."

---

## Anhang A — Vollständige Event-Liste (geordnet)

### Kassenbetrieb (Bounded Context: service)

| # | Domain Event                         | Command                        | Akteur                      | Policy / Nachfolge-Event             |
| - | ------------------------------------ | ------------------------------ | --------------------------- | ------------------------------------ |
| 1 | Servicekraft angemeldet              | Anmelden                       | Servicekraft                | JWT ausgestellt                      |
| 2 | Tisch ausgewählt                     | Tisch auswählen                | Servicekraft                | TischDetail geladen                  |
| 3 | Produkt ausgewählt                   | Produkt zur Bestellung hinzufügen | Servicekraft             | —                                    |
| 4 | Menge geändert                       | Menge erhöhen / verringern     | Servicekraft                | —                                    |
| 5 | Position aus Bestellung entfernt     | Position entfernen             | Servicekraft                | —                                    |
| 6 | Kommentar hinzugefügt                | Kommentar eingeben             | Servicekraft                | —                                    |
| 7 | Bestellung abgebrochen               | Bestellung verwerfen           | Servicekraft                | Lokaler State geleert                |
| 8 | **Bestellung aufgegeben** 🔑         | BestellungAufgeben             | Servicekraft                | Event persistiert · Saldo aktualisiert · (Bon gedruckt) |
| 9 | Produkte ausgeliefert                | ProdukteAusliefern             | Servicekraft                | Event persistiert · Ungelieferte Positionen aktualisiert |
| 10 | Zahlung registriert                 | ZahlungRegistrieren            | Servicekraft                | Event persistiert · Saldo aktualisiert |
| 11 | Tischkonto ausgeglichen             | — (Policy)                     | System                      | Ausgelöst wenn Saldo = 0             |
| 12 | Positionen storniert                | ProdukteStornieren             | Serviceleitung / Admin      | Event persistiert · Saldo aktualisiert |
| 13 | Bestellung umgebucht ❌              | BestellungUmbuchen             | Serviceleitung / Admin      | Storno alt + Bestellung neu (atomar) |

### Stammdaten (Bounded Context: admin)

| # | Domain Event                         | Command                        | Akteur  | Policy / Nachfolge-Event         |
| - | ------------------------------------ | ------------------------------ | ------- | -------------------------------- |
| 14 | Benutzer angelegt                   | BenutzerAnlegen                | Admin   | Einmalpasswort generiert         |
| 15 | Passwort gesetzt                    | PasswortSetzen                 | Benutzer | JWT ausgestellt                 |
| 16 | Passwort zurückgesetzt              | PasswortZurücksetzen           | Admin   | Neues Einmalpasswort generiert   |
| 17 | Produkt hinzugefügt                 | ProduktErstellen               | Admin   | Produktkatalog aktualisiert      |
| 18 | Produktvariante hinzugefügt         | VarianteHinzufügen             | Admin   | Produktkatalog aktualisiert      |
| 19 | Produktpreis geändert               | VarianteBearbeiten             | Admin   | Nur für zukünftige Bestellungen  |
| 20 | Produkt deaktiviert                 | ProduktDeaktivieren            | Admin   | Produkt aus Katalog entfernt     |
| 21 | Tisch angelegt                      | TischAnlegen                   | Admin   | Tisch in Übersicht verfügbar     |
| 22 | Tisch deaktiviert                   | TischDeaktivieren              | Admin   | Tisch aus Übersicht entfernt     |

### Abrechnung (Querschnitt)

| # | Domain Event                         | Command                        | Akteur        | Policy / Nachfolge-Event              |
| - | ------------------------------------ | ------------------------------ | ------------- | ------------------------------------- |
| 23 | Umsatz pro Servicekraft ermittelt ❌ | UmsatzAbrufen                  | Kassenwart    | Events aggregieren nach userID        |
| 24 | Abrechnung exportiert ❌             | AbrechnungExportieren          | Kassenwart    | CSV-Datei generieren                  |
| 25 | Tagesabschluss durchgeführt ❌       | TagesabschlussEinleiten        | Admin/Kassenwart | Abschlussbericht generieren         |

---

## Anhang B — Stickies-Legende

| Farbe    | Symbol | Bedeutung                                                              |
| -------- | ------ | ---------------------------------------------------------------------- |
| 🟠 Orange | —     | **Domain Event** — etwas, das in der Domäne passiert ist (Vergangenheit) |
| 🔵 Blau  | —      | **Command** — Absicht, etwas zu tun (Imperativ)                        |
| 🟡 Gelb  | —      | **Akteur** — Person oder System, das den Command auslöst               |
| 🟢 Grün  | —      | **Read Model** — Lese-Sicht / Projektion für die Anzeige               |
| 🟣 Lila  | —      | **Policy** — automatische Reaktion: „Wenn X, dann Y"                   |
| 🌸 Rosa  | —      | **External System** — System außerhalb des Bounded Context             |
| ❤️ Rot   | —      | **Hotspot** — Unsicherheit, Diskussionsbedarf, offene Frage            |

---

*Dieses Dokument basiert auf dem Event-Storming-Workshop vom 9. März 2026. Alle genannten Personen sind fiktiv; die Domain-Erkenntnisse spiegeln den realen Ist- und Soll-Zustand von jotti wider.*
