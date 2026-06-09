# PRD: Kassenbeleg (F-03) — Steueraufschlüsselung & modulare TSE-Bereitschaft

> **Zuschnitt:** Gap-PRD. Die Basis von F-03 (Beleg auf Anforderung, Pflicht-Basisfelder, Drucker-Fehlermeldung) ist bereits implementiert und getestet. Dieses PRD spezifiziert ausschließlich die **Restarbeiten**: die steuerliche Aufschlüsselung auf dem Beleg (F-07) und den **modularen Einhängepunkt** für die spätere TSE-Integration (F-02), sodass der Beleg heute vollständig **ohne TSE** funktioniert.

## Problem Statement

Servicekräfte können bereits auf Anforderung einen Kassenbeleg für eine Tischzahlung oder einen Direktverkauf drucken. Der gedruckte Beleg enthält Vereinsdaten, Kassen-Seriennummer, Datum, Positionen mit Einzelpreis × Menge, Gesamtbetrag und Zahlungsart — aber **keine steuerliche Aufschlüsselung**. Damit erfüllt er die Pflichtangaben nach § 14 UStG / § 6 KassenSichV noch nicht: Es fehlen das Steuerkennzeichen je Position und die Steuermatrix (Netto, Steuer, Brutto je Steuersatz) im Belegfuß.

Gleichzeitig soll die spätere TSE-Integration den Beleg nur **modular ergänzen**, ohne dass das System heute eine TSE benötigt. Ein Verein im Test- oder Demobetrieb — oder einer, der (noch) keine TSE einsetzt — muss einen vollständigen, druckbaren Beleg erhalten. Es braucht also einen klar definierten Einhängepunkt, der bei fehlenden TSE-Daten einfach nichts druckt.

## Solution

Der bestehende Kassenbeleg wird um die gesetzlich vorgeschriebene Steueraufschlüsselung erweitert: Jede Position erhält ein **Steuerkennzeichen**, und im **Belegfuß** erscheint eine **Steuermatrix** mit Netto-, Steuer- und Bruttobetrag je Steuersatz. `kombi`-Positionen (70/30-Pauschalierung) fließen anteilig in die 7-%- und 19-%-Zeilen ein. Die Berechnung nutzt die bereits vorhandene, getestete Steuer-Aufteilungslogik.

Der Beleg-Formatter erhält zusätzlich einen **optionalen TSE-Abschnitt**, der nur gedruckt wird, wenn TSE-Daten vorliegen. Heute liegen keine vor — der Beleg druckt exakt wie bisher, nur ergänzt um die Steueraufschlüsselung. Die spätere TSE-Integration (F-02) befüllt diesen Abschnitt, ohne den Beleg-Code erneut umbauen zu müssen.

Zuletzt wird die veraltete Statusangabe in der Anforderungsdokumentation korrigiert, da F-03 bereits weitgehend umgesetzt ist.

## User Stories

**Beleg auf Anforderung (Baseline — bereits umgesetzt, hier zur Vollständigkeit & Regressionsschutz)**

1. Als Servicekraft möchte ich nach einer Tischzahlung auf Anforderung einen Kassenbeleg drucken, damit ein Gast bei Bedarf einen Beleg erhält.
2. Als Servicekraft möchte ich nach einem Direktverkauf auf Anforderung einen Kassenbeleg drucken, damit ein Theken-Gast einen Beleg erhält.
3. Als Servicekraft möchte ich denselben Beleg bei Bedarf erneut drucken (Nachdruck), ohne den Kassiervorgang fachlich zu wiederholen, damit ein verlorener Bon ersetzt werden kann.
4. Als Servicekraft möchte ich eine klare Fehlermeldung erhalten, wenn kein Kassenbeleg-Drucker konfiguriert ist, damit ich weiß, dass die Admin-Einstellungen zu prüfen sind.
5. Als Admin möchte ich Vereinsname und Adresse pflegen, damit sie auf dem Beleg erscheinen.
6. Als Admin möchte ich die Kassenbeleg-Drucker-IP konfigurieren, damit Belege am richtigen Drucker ausgegeben werden.
7. Als Servicekraft möchte ich, dass die bestehenden Beleg-Basisangaben (Vereinsdaten, Kassen-Seriennummer, Datum/Uhrzeit, Positionen mit Preis, Gesamtbetrag, Zahlungsart „bar", Bon-Nummer) unverändert erhalten bleiben, damit keine Regression entsteht.

**Steueraufschlüsselung (F-07 auf dem Beleg)**

8. Als Gast möchte ich pro Position ein Steuerkennzeichen sehen, damit ich erkenne, mit welchem Steuersatz die Position besteuert wurde.
9. Als Gast möchte ich im Belegfuß eine Steuermatrix mit Netto-, Steuer- und Bruttobetrag je Steuersatz sehen, damit die Steueraufschlüsselung nachvollziehbar ist.
10. Als Finanzamt/Prüfer möchte ich, dass der Beleg keine unaufgeteilte Gesamtsumme ohne Steueraufschlüsselung ausweist, damit der Beleg § 14 UStG / § 6 KassenSichV genügt.
11. Als Gast möchte ich, dass eine `kombi`-Position (z. B. Speise-/Getränke-Kombi) korrekt im Verhältnis 70/30 auf den ermäßigten und den Regelsatz aufgeteilt in der Steuermatrix erscheint, damit Mischkalkulationen korrekt ausgewiesen sind.
12. Als Gast möchte ich, dass steuerbefreite Positionen (0 %) korrekt mit eigenem Kennzeichen und eigener Matrixzeile ausgewiesen werden, damit auch Befreiungen transparent sind.
13. Als Betreiber möchte ich, dass die Beträge der Steuermatrix in Summe exakt dem Gesamtbetrag entsprechen (centgenaue Rundung), damit der Beleg in sich stimmig ist.

**Modulare TSE-Bereitschaft (Vorbereitung F-02)**

14. Als Verein im Test- oder Demobetrieb (ohne TSE) möchte ich, dass der Beleg vollständig und druckbar ist, damit ich jotti ohne TSE nutzen und vorführen kann.
15. Als Entwickler möchte ich, dass der Beleg-Formatter einen optionalen TSE-Abschnitt besitzt, der nur bei vorhandenen TSE-Daten gedruckt wird, damit der Beleg ohne TSE wie bisher funktioniert und TSE später modular andockt.
16. Als Entwickler möchte ich, dass die Datenstruktur des Belegs eine klar definierte, optionale Stelle für TSE-Felder vorsieht (Integrationsvertrag), damit die TSE-Integration (F-02) diese Stelle nur noch befüllen muss, ohne den Beleg-Code umzubauen.

**Dokumentation**

17. Als Mitwirkender möchte ich, dass der Umsetzungsstatus von F-03 in der Anforderungsdokumentation der Realität entspricht, damit niemand bereits gebaute Funktionalität erneut plant.

## Implementation Decisions

### Ausgangslage (Baseline — bereits vorhanden, bleibt unverändert)

- Ein POST-Endpunkt im Service-Bereich erzeugt den Kassenbeleg **auf Anforderung** und nimmt entweder eine Direktverkauf-Referenz **oder** die Kombination aus Tisch- und Zahlungsreferenz entgegen.
- Der zugehörige Application-Command liest das Quell-Ereignis (Zahlung bzw. Direktverkauf) aus dem Kassenjournal, lädt Betreiber-Stammdaten, Kassen-Seriennummer und Drucker-Konfiguration, formatiert den Beleg als ESC/POS-Payload und reiht **genau einen** Druckauftrag (Bon-Art „Kassenbeleg") in die Druck-Outbox ein. Erneuter Aufruf bewirkt einen idempotenten Nachdruck.
- Fehlt die Drucker-IP, liefert der Command einen klar abbildbaren Fehler; das Frontend zeigt eine verständliche Meldung. Beide Service-Flows (Tischzahlung, Direktverkauf) besitzen bereits den Auslöse-Button.
- Diese Bestandteile werden **nicht** neu gebaut. Bestehende Tests bleiben grün (Regressionsschutz).

### Änderung 1 — Steuersatz bis zum Beleg durchreichen (reine Lese-Seite)

- Die Domain-Ereignisse `zahlung-kassiert:v1` und `direktverkauf-getaetigt:v1` **persistieren den Steuersatz je Position bereits** (Teil der unveränderlichen Positionsdaten). Lediglich die Lese-/Decode-Strukturen der Application-Schicht für den Belegdruck lassen das Feld derzeit fallen.
- Entscheidung: Der Steuersatz wird in den Decode-Strukturen ergänzt und in das Positionsmodell des Belegs übernommen.
- **Kein** Eventformat-Wechsel, **keine** Schemaänderung, **keine** Migration und **keine** Write-Seiten-Änderung — die Daten liegen bereits im Journal. Es ist ausschließlich eine Lese-Seiten-Korrektur.

### Änderung 2 — Steuermatrix-Berechnung (neues, isoliert testbares Modul)

- Ein neues, **pures** Modul berechnet aus der Positionsliste eines Belegs die Steuermatrix für den Belegfuß: Eingabe sind die Positionen (Bruttobetrag = Einzelpreis × Menge plus Steuersatz), Ausgabe ist eine nach Steuersatz aggregierte Liste von Zeilen mit Netto-, Steuer- und Bruttobetrag.
- Die eigentliche Aufteilung Brutto → Netto/Steuer pro Satz nutzt die **bestehende, getestete** Steuer-Aufteilungslogik der Domäne (inklusive der 70/30-Aufteilung für `kombi`). Diese Logik wird **wiederverwendet, nicht dupliziert**.
- `kombi`-Positionen erzeugen über die Aufteilungslogik zwei Teilbeträge (ermäßigt/Regel), die in die jeweiligen Matrixzeilen einfließen. Die Matrix führt also nie eine eigene „kombi"-Zeile, sondern nur die effektiven Sätze (19 %, 7 %, 0 %).
- Das Modul ist bewusst schmal in der Schnittstelle und tief in der Logik (Aggregation + Rundungsverhalten gekapselt), damit es in Isolation getestet werden kann und sich selten ändert.

### Änderung 3 — Beleg-Formatter um Steuerausweis erweitern

- **Pro Position** wird ein Steuerkennzeichen ausgegeben (Regelsatz, ermäßigt, befreit als unterscheidbare Kennungen, z. B. Buchstabencodes gemäß steuerrecht.md §6 / compliance.md §5.2).
- Im **Belegfuß** wird die in Änderung 2 berechnete Steuermatrix gedruckt (je Satz: Netto, Steuer, Brutto). Der Gesamtbetrag bleibt erhalten, erscheint aber nie mehr ohne zugehörige Steueraufschlüsselung.
- Der bestehende Basisinhalt (Kopf, Positionszeilen mit Preis, Gesamt, Zahlungsart, Schnittbefehl) bleibt unverändert; der Steuerausweis wird additiv ergänzt.

### Änderung 4 — Optionaler TSE-Abschnitt als modularer Einhängepunkt

- Die Datenstruktur des Belegs (`KassenbelegData`-Äquivalent) erhält eine **optionale** TSE-Sektion (präsenzbasiert: vorhanden / nicht vorhanden).
- Der Formatter druckt den TSE-Block **nur**, wenn die Sektion vorhanden ist. Ist sie leer (heutiger Zustand und jeder Betrieb ohne TSE), bleibt der Beleg unverändert druckbar.
- **Integrationsvertrag mit F-02:** Die spätere TSE-Integration ergänzt die TSE-Felder als optionale Felder an den betroffenen Kasse-Ereignissen (gemäß handbuch.md §3.13) und befüllt aus ihnen die optionale Beleg-Sektion. Dieses PRD legt ausschließlich die **Beleg-seitige** optionale Stelle und das Rendering an; es werden **keine** ungenutzten TSE-Felder an Ereignissen angelegt (kein Dead Code vor der eigentlichen Integration).

### Änderung 5 — Dokumentation angleichen

- Der Status von F-03 in der Anforderungsdokumentation wird auf die Realität gebracht (Basis umgesetzt; offen verbleiben Steueraufschlüsselung und TSE-Felder). Die entsprechenden Akzeptanzkriterien werden abgehakt bzw. präzisiert. Vorhandene Quellcode-Hinweise (Formatter-TODOs) werden konsistent gehalten.

### Annahmen

> **Annahme (kombi-Darstellung pro Position):** Bei `kombi`-Positionen wird auf der Positionszeile ein gemeinsamer Positionstext mit Verweis auf die Steuermatrix im Belegfuß gedruckt (statt zwei Teilzeilen pro Position). compliance.md §5.2 lässt beide Varianten zu; die kombinierte Darstellung ist auf dem schmalen Thermobon (48 Zeichen) lesbarer, die exakte Aufteilung steht autoritativ in der Matrix.

> **Annahme (On-Screen-Belegvorschau bleibt unverändert):** Die im Frontend vorhandene Bildschirm-Belegvorschau beim Kassieren ist **kein** fiskalischer Beleg und wird in diesem PRD nicht um einen Steuerausweis erweitert. Maßgeblich ist der gedruckte Kassenbeleg.

## Testing Decisions

**Was einen guten Test ausmacht:** Getestet wird **externes Verhalten**, nicht die interne Umsetzung. Für den Formatter heißt das: Prüfen, dass der gerenderte Belegtext die geforderten Angaben enthält (Steuerkennzeichen, Matrixzeilen je Satz, Beträge, Gesamt) — **nicht** die exakte Byte-Anordnung der ESC/POS-Steuerzeichen. Für die Steuerlogik: Prüfen der berechneten Beträge anhand repräsentativer Eingaben.

**Zu testende Module:**

- **Steuermatrix-Modul (Pflicht):** Pure Unit-Tests über repräsentative Positionsmengen — einzelner Satz, mehrere Sätze gemischt, `kombi` mit korrekter 70/30-Aufteilung, befreit (0 %), Rundungs-/Summenkonsistenz (Summe der Matrix = Gesamtbetrag). Vorbild: die bestehenden Tests der Steuer-Aufteilungslogik der Domäne.
- **Formatter-Erweiterung (Pflicht):** Inhaltsprüfungen, dass Steuerkennzeichen je Position und die Steuermatrix im Belegfuß erscheinen und der Beleg weiterhin mit dem Schnittbefehl endet. Vorbild: die bestehenden „Pflichtfelder"- und „CutPaper"-Tests des Kassenbelegs.
- **Optionaler TSE-Block (Pflicht):** Zwei Verhaltensfälle — ohne TSE-Daten erscheint **kein** TSE-Block (Beleg unverändert); mit TSE-Daten erscheinen die TSE-Pflichtfelder.
- **Lese-Seiten-Durchreichung (Soll):** Ein Test, der end-to-end belegt, dass der Steuersatz aus einem Zahlungs- bzw. Direktverkauf-Ereignis im gerenderten Beleg ankommt (korrekte Matrix für bekannte Positionen). Vorbild: die bestehenden Kassenbeleg-Command-Tests.

**Regressionsschutz:** Alle bestehenden Beleg-, Command- und Handler-Tests bleiben unverändert grün.

## Out of Scope

- **Tatsächliche TSE-Integration (F-02):** Anbindung an eine Cloud-TSE, `TSEClient`-Interface, Signatur/Transaktionsnummern, Befüllung der TSE-Felder. Dieses PRD bereitet nur den Beleg-seitigen Einhängepunkt vor.
- **„Erste Bestellung"-Klarschrift-Zeitstempel beim Durchbedienen** (compliance.md §5.3): hängt an TSE-`Bestellung-V1`-Transaktionen und gehört zu F-02.
- **QR-Code auf dem Beleg** (TSE-/DSFinV-K-Anhang I, F-09): erst mit TSE relevant.
- **DSFinV-K-Export** (F-04) inkl. `vat.csv`/`transactions_vat.csv`.
- **Digitaler eBeleg** (F-09) und Beleg-Archivierung.
- **Gutschein-Sonderfälle** (Mehrzweck-Gutschein, 0 % bei Ausgabe).
- **Bildschirm-Belegvorschau** im Frontend (kein fiskalischer Beleg).
- **Belegausgabe-Befreiungs-Workflow** am Fest (organisatorisch/rechtlich, kein Code).

## Further Notes

- Die On-Demand-Druckweise entspricht bereits der Belegausgabe-Befreiung für Vereinsfeste (compliance.md §5.1): kein Automatikdruck nach jeder Zahlung, aber jederzeit erstellbar.
- Nach diesem PRD sind alle **nicht-TSE-abhängigen** Akzeptanzkriterien von F-03 erfüllt; offen bleibt ausschließlich der TSE-Block, der mit F-02 befüllt wird.
- Die Steuer-Aufteilungslogik der Domäne ist die einzige Quelle der Wahrheit für Brutto→Netto/Steuer und die `kombi`-Pauschalierung; der Beleg rechnet nicht selbst.
- Referenzen: compliance.md §5.2 (Pflichtangaben), steuerrecht.md §6 (Belegausweis), handbuch.md §3.13 (TSE-Datenmodell) und §4.6 (Bondruck: Arbeitsbon vs. Kassenbeleg).
