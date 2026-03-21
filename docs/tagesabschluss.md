# Tagesabschluss

Der **Tagesabschluss** (oft als **Z-Bon** oder Z-Abschlag bezeichnet) ist die aggregierende Zusammenfassung aller abgeschlossenen Kassenvorgänge (vom Vorgangstyp "Beleg") für einen bestimmten Zeitraum – bei einem mehrtägigen Fest also für den jeweiligen Festtag oder die jeweilige Schicht.

Für dein Systemdesign bedeutet das, dass der Tagesabschluss nicht nur ein Stück Papier ist, sondern eine zentrale buchhalterische und relationale Funktion erfüllt. Er stellt die Verbindung von den detaillierten Einzelbuchungen (den einzelnen Bier- und Bratwurstverkäufen) zur täglichen Gesamtsumme für die Finanzbuchhaltung des Vereins her.

Hier ist die genaue rechtliche und technische Bedeutung für dein Festzelt-Kassensystem:

**1. Die Pflicht zum Kassensturz (Soll-Ist-Abgleich)**
Kasseneinnahmen und -ausgaben müssen zwingend täglich festgehalten werden. Der Tagesabschluss dient der **Kassensturzfähigkeit**: Der Verein muss am Ende des Tages nachprüfen können, ob der vom System errechnete Bargeldbestand (Soll) mit dem tatsächlich in der Kasse oder Geldkatze gezählten Geld (Ist) übereinstimmt.
Tritt hierbei eine Kassendifferenz (Fehlbetrag oder Überschuss) auf, musst du diese in der Kassensoftware zwingend als eigenen Geschäftsvorfall vom Typ **"DifferenzSollIst"** erfassen und protokollieren.

**2. Interner Speicher-Reset (Z-Bon vs. X-Bon)**
Mit der Erstellung des Tagesabschlusses (Z-Bon) wird der interne Kassenumsatzspeicher deines Systems für die laufenden Summen wieder auf 0 zurückgesetzt, um den nächsten Festtag sauber beginnen zu können.
Ein reiner Zwischenbericht (sogenannter X-Bon), bei dem das Personal nur den aktuellen Stand abfragt, ohne den Speicher zu nullen, ersetzt den Z-Bon rechtlich nicht. Die Z-Bons müssen fortlaufend nummeriert und über 10 Jahre lückenlos archiviert werden, da bei fehlenden Z-Bons sofort die Vollständigkeit der Einnahmen angezweifelt wird.

**3. Umgang mit offenen Tischen bei Tagesende**
Da die DSFinV-K und die Kassenlogik auf Tagesabschlüssen basieren, stellt sich im Festzelt oft die Frage nach Tischen, die kurz vor Feierabend noch offen sind.
In den Kassenabschluss fließen **ausschließlich** Geschäftsvorfälle ein, die bereits steuerlich abgeschlossen sind (Vorgangstyp "Beleg"). Noch unbezahlte Bestellungen am Tisch haben den Typ "AVBestellung" (ein sogenannter "Anderer Vorgang", der noch nicht zu verbuchen ist). Diese verbleiben unangetastet im System. Erst wenn der Tisch am nächsten Tag bezahlt wird, entsteht ein "Beleg", der dann in den Kassenabschluss des neuen Tages einfließt.

**4. Architektonische Umsetzung im DSFinV-K Export**
Für dich als Entwickler ist der Kassenabschluss der Dreh- und Angelpunkt deines Datenexports. In der DSFinV-K bildet das **Kassenabschlussmodul** eine eigene Struktur.

- **Die `Z_NR` (Der Primärschlüssel):** Jeder Abschluss erhält eine fortlaufende, nicht zurücksetzbare Kassenabschlussnummer (Feld `Z_NR`). Diese Nummer wird als Foreign Key in fast allen anderen Export-Tabellen benötigt, um Vorgänge einem bestimmten Abschluss zuzuordnen.
- **Aggregierte CSV-Dateien:** Das Modul generiert Dateien wie die `businesscases.csv` (Z_GV_TYP), welche die zu verbuchenden Summen nach Steuersätzen und Geschäftsvorfällen gruppiert, sowie die `payment.csv` (Z_Zahlart), welche die Summen nach Zahlungsarten (z.B. Bar, EC-Karte) trennt.
- **Stammdaten-Snapshot (Wichtige Regel):** Um eine komplexe Historisierung von Stammdaten zu vermeiden, verlangt der Gesetzgeber, dass das System zu **jedem** Kassenabschluss einen Snapshot der aktuell gültigen Stammdaten (wie Steuersätze, Kassen-IDs, TSE-Zertifikate) in Dateien wie `cashpointclosing.csv` mitspeichert. **Achtung:** Wenn der Verein mitten am Tag Stammdaten ändert (z.B. den Preis für das Bier anpasst), muss deine Software zwingend _vor_ der Änderung automatisch einen Kassenabschluss durchführen.

---

Der **Tagesabschluss** (auch Z-Bon oder Z-Abschlag genannt) ist die aggregierende Zusammenfassung einer Kasse über alle Einzelbewegungen mit dem Vorgangstyp „Beleg“ für einen bestimmten Zeitraum, beispielsweise für einen Festtag oder eine Kassenschicht. Mit der Erstellung des Tagesabschlusses wird der interne Kassenumsatzspeicher für die laufenden Summen wieder auf null zurückgesetzt.

**Der Gesamtkontext: Einordnung in GoBD, KassenSichV und DSFinV-K**
Der Tagesabschluss erfüllt eine zentrale buchhalterische Funktion: Er stellt die Verbindung zwischen jedem einzeln bonierten Bier (den Kasseneinzelbewegungen) und der aufsummierten Tagessumme für die Finanzbuchhaltung her.
Nach der Abgabenordnung (AO) und den GoBD müssen Kasseneinnahmen zwingend täglich festgehalten und kassensturzfähig sein. In der standardisierten Exportschnittstelle des Finanzamts (DSFinV-K) bildet der Tagesabschluss ein eigenes, hochrelevantes Modul (Kassenabschlussmodul), das Prüfern einen strukturierten Überblick über alle Geschäftsvorfälle der Kasse liefert.

**Detaillierte rechtliche Anforderungen**

- **Ausschließliche Erfassung abgeschlossener Belege:** In den Tagesabschluss fließen nur Vorgänge ein, die als „Beleg“ deklariert sind. Noch unbezahlte, offene Tisch-Bestellungen im Festzelt (Vorgangstyp „AVBestellung“) bleiben unberücksichtigt, bis sie tatsächlich bezahlt werden.
- **Fortlaufende Nummerierung:** Z-Bons müssen eine aufsteigende, lückenlose und nicht zurücksetzbare Kassenabschlussnummer (`Z_NR`) erhalten.
- **Sicherung der Stammdaten:** Um Redundanzen und komplexe Historisierungen zu vermeiden, fordert das Gesetz, dass zu jedem Kassenabschluss ein „Snapshot“ der aktuell gültigen Stammdaten (wie Steuersätze, TSE-Zertifikate, Kassen-IDs) fest gespeichert wird.
- **Kassensturzfähigkeit & Differenzen:** Der Kassenbestand muss jederzeit überprüfbar sein. Tritt beim Abgleich eine Differenz zwischen errechnetem und realem Bestand auf, muss diese zwingend als eigener Geschäftsvorfall („DifferenzSollIst“) gebucht werden.
- **Aufbewahrungsfrist:** Alle Tagesabschlüsse (Z-Bons) müssen unveränderbar für 10 Jahre archiviert werden.

**Anleitung: Was du als Verantwortlicher des Vereins tun musst**

1.  **Bargeld zählen:** Führe am Ende des Festtages oder nach Schichtende eine physische Bestandsaufnahme (Kassensturz) durch, indem du das gesamte Geld in der Kasse zählst. Es wird dringend empfohlen, hierfür ein Zählprotokoll zu erstellen.
2.  **Tagesabschluss im System auslösen:** Führe in deiner Kassensoftware den offiziellen Tagesabschluss (Z-Bon) durch. Nutze dafür **keinen** Zwischenbericht (X-Bon), da dieser den Speicher nicht zurücksetzt und gesetzlich nicht ausreicht.
3.  **Abgleich durchführen:** Vergleiche das von dir gezählte Bargeld (Ist-Bestand) mit dem auf dem Z-Bon ausgewiesenen Betrag (Soll-Bestand).
4.  **Fehlbeträge oder Überschüsse buchen:** Gibt es eine Abweichung, musst du diese Differenz im Kassensystem buchen und aufklären. **Achtung:** Der Kassenbestand darf niemals negativ sein.
5.  **Dokumente sichern:** Speichere oder drucke den Z-Bon ab und archiviere ihn zusammen mit dem Zählprotokoll sicher für 10 Jahre.
6.  **Regel bei Stammdaten-Änderungen:** Wenn du während des Festes Preise ändern möchtest (z.B. Bier wird am zweiten Tag teurer), musst du zwingend _vor_ dieser Änderung einen Kassenabschluss durchführen.

**Technische Anforderungen an das Kassensystem (für Softwareentwickler)**

- **TSE-Absicherung des Abschlusses:** Der Tagesabschluss (Z-Bon) muss als eigene Transaktion an die TSE gesendet und kryptografisch abgesichert werden. Der `processType` hierfür lautet in der Regel `SonstigerVorgang-V1`, da es sich nicht um einen klassischen Verkaufsbeleg handelt.
- **Kassenabschlussmodul im DSFinV-K Export:** Die Software muss die Abschlussdaten in genormten CSV-Dateien generieren:
  - **`businesscases.csv` (Z_GV_Typ):** Hier müssen die Tagesumsätze exakt nach Geschäftsvorfalltyp (z.B. "Umsatz" oder "Rabatt") und nach Umsatzsteuerschlüssel (z.B. 19% oder 7%) aggregiert ausgewiesen werden.
  - **`payment.csv` (Z_Zahlart):** Die Summen müssen nach den eingesetzten Zahlarten (z.B. Bar, EC-Karte) aufgeschlüsselt werden.
  - **`cash_per_currency.csv`:** Ausweisung des Kassenbestands nach Währungen.
- **Stammdaten-Snapshot erstellen:** Die Software muss parallel zum Abschluss die Datei `cashpointclosing.csv` (Stamm_Abschluss) sowie weitere Stammdaten-Dateien erzeugen, um den Konfigurationszustand der Kasse (Hardware, Zertifikate) zum exakten Zeitpunkt des Abschlusses "einzufrieren".
- **Verknüpfung der Primärschlüssel:** Alle Zeilen im Export müssen die Schlüssel `Z_KASSE_ID`, `Z_ERSTELLUNG` (Zeitstempel des Abschlusses) und `Z_NR` (die Nummer des Z-Bons) mitführen, damit die Finanzämter die Daten in ihrer Prüfsoftware relational verknüpfen können.

---

Für die programmtechnische Umsetzung des Kassensturzes und der daraus resultierenden Differenzen musst du dich strikt an die Vorgaben der GoBD und der DSFinV-K (Digitale Schnittstelle der Finanzverwaltung) halten. Die Finanzverwaltung sucht bei Betriebsprüfungen mithilfe der IDEA-Software gezielt nach Auffälligkeiten wie fehlenden Differenzen, da eine Kasse, die "immer auf den Cent genau stimmt", als hochgradig manipulationsverdächtig gilt.

Hier ist die konkrete architektonische Umsetzung für deine Kassensoftware:

**1. Die Logik des Kassensturzes (Soll-Ist-Abgleich)**
Deine Software muss eine Funktion für den Kassensturz (meist vor dem Tagesabschluss) bereitstellen.

- **Soll-Bestand:** Das System errechnet den theoretischen Bargeldbestand (Anfangsbestand + Bareinnahmen - Barausgaben).
- **Ist-Bestand:** Die Servicekraft gibt den tatsächlich in der Kassenschublade oder Geldkatze gezählten Bargeldbestand in eine Eingabemaske ein (idealerweise über ein digitales Zählprotokoll in der Software).
- **Die goldene Regel:** Ein Kassenbestand darf **niemals negativ** sein. Deine Software muss eine Validierung einbauen, die eine Eingabe blockiert, falls der Ist-Bestand unter 0,00 Euro fallen würde.

**2. Erzeugung eines "Eigenbelegs" (Strikes elektronisches Radierverbot)**
Ergibt die Subtraktion von Soll- und Ist-Bestand eine Differenz, darfst du den Soll-Bestand in der Datenbank **unter keinen Umständen** einfach stillschweigend mit einem `UPDATE` überschreiben.

- Du musst für diese Abweichung zwingend einen neuen, automatisierten Datensatz (einen sogenannten Eigenbeleg) generieren.
- Dieser Beleg muss wie ein ganz normaler Vorgang (Kassenbeleg) über die **TSE (Technische Sicherheitseinrichtung)** kryptografisch signiert werden, da er die Vermögenszusammensetzung der Kasse verändert.

**3. Das Mapping im DSFinV-K Export (`DifferenzSollIst`)**
Wenn du den DSFinV-K Export generierst, verlangt das Finanzamt, dass Kassenfehlbeträge oder -überschüsse eindeutig als solche deklariert werden.

- **Das Feld `GV_TYP` (Geschäftsvorfalltyp):** In den Dateien `lines.csv` (Bonpos) und `businesscases.csv` (Z_GV_TYP) musst du diesem generierten Datensatz zwingend den exakten Geschäftsvorfalltyp **`DifferenzSollIst`** zuweisen. Andere Bezeichnungen sind hier nicht zulässig.
- **Vorzeichen:** Das Feld fasst sowohl Kassenfehlbeträge (es ist weniger Geld in der Kasse als erwartet) als auch positive Differenzen (Überschüsse) zusammen. Ein Fehlbetrag wird dabei als negativer Wert exportiert, ein Überschuss als positiver Wert.

**4. Die steuerliche Behandlung (UST_SCHLUESSEL)**
Oft fragen sich Entwickler, mit welchem Steuersatz eine Kassendifferenz gebucht werden muss. Die DSFinV-K liefert hier eine klare Entwarnung für die Kassensoftware:

- Die ertrag- und umsatzsteuerliche Qualifikation (also ob z. B. ein Mitarbeiter für den Fehlbetrag haftet oder ob es sich um einen steuerpflichtigen Überschuss handelt) muss **nicht** zwingend von der Kasse selbst entschieden werden.
- Das Gesetz besagt, dass diese steuerlichen Konsequenzen bezogen auf den Sachverhalt im Nachhinein zu prüfen und in "nachgelagerten Systemen" (also der Finanzbuchhaltungssoftware des Steuerberaters wie z.B. DATEV) zu dokumentieren sind. Du kannst diese Buchung also als umsatzsteuerlich neutral (z.B. mit dem Schlüssel für "Nicht Steuerbar" oder "UmsatzsteuerNichtErmittelbar") in den Export übergeben.

Zusammenfassend: Sobald die Servicekraft beim Zählen einen abweichenden Betrag eingibt, generiert deine Software einen internen TSE-Beleg, flaggt diesen im Export-Datensatz als `DifferenzSollIst` und bereinigt so den internen Kassenumsatzspeicher für den anstehenden Tagesabschluss sauber auf den echten Zählwert.

---

Wenn während des laufenden Festzelt-Betriebs Bargeld aus der Kasse entnommen wird (z. B. um es in den Tresor oder zur Bank zu bringen), handelt es sich rechtlich um eine reine Vermögensumschichtung. In deiner Kassensoftware und dem DSFinV-K Export wird dies über den Geschäftsvorfalltyp **`Geldtransit`** abgebildet.

- **TSE-Signatur & Belegpflicht:** Auch wenn kein Verkauf an einen Kunden stattfindet, verändert die Abschöpfung den internen Bargeldbestand der Kasse. Du musst dafür in der Software zwingend einen "Eigenbeleg" erzeugen, diesen über die TSE signieren lassen und aufbewahren. Es gilt der strikte Grundsatz: "Keine Buchung ohne Beleg".
- **DSFinV-K Mapping:** In der Exportdatei `businesscases.csv` (Z_GV_TYP) erhält dieser Vorgang zwingend den Typ `Geldtransit`. Dient die Entnahme hingegen zur Überführung in den privaten Bereich des Betreibers, ist stattdessen der Typ `Privatentnahme` zu wählen.
- **Steuerliche Behandlung:** Ein Geldtransit hat keinerlei umsatzsteuerliche Relevanz. Er wird im Export daher dem USt-Schlüssel für "Nicht Steuerbar" (z. B. ID 5) zugeordnet.
- **Sicherung der Kassensturzfähigkeit:** Durch die Erfassung des Transits sinkt der im System geführte Soll-Bestand, sodass er jederzeit exakt mit dem nach der Abschöpfung verbleibenden physischen Ist-Bestand (z. B. dem Wechselgeld) in der Kassenschublade übereinstimmt.

**2. Korrekte Handhabung von Trinkgeldern**
Die Programmierung von Trinkgeldern ist technisch komplexer, da du architektonisch strikt zwischen **Trinkgeld für den Unternehmer (Arbeitgeber)** und **Trinkgeld für das Personal (Arbeitnehmer)** unterscheiden musst.

**A. Unternehmer-Trinkgeld (`TrinkgeldAG`)**
Wenn der Vereinsvorstand oder Inhaber selbst hinter der Theke steht und Trinkgeld erhält, ist dieses steuerlich Teil des regulären Umsatzes und umsatzsteuerpflichtig.

- In der Software und im DSFinV-K Export buchst du diesen Zufluss mit dem Geschäftsvorfalltyp **`TrinkgeldAG`**.
- Das System muss dieses Trinkgeld gemäß der jeweiligen Umsatzsteuer verarbeiten. Die spätere physische Entnahme dieses Geldes aus der Kasse buchst du dann über `Geldtransit` oder `Privatentnahme`.

**B. Arbeitnehmer-Trinkgeld (`TrinkgeldAN`)**
Trinkgeld für freiwillige Helfer oder angestellte Bedienungen ist für den Unternehmer (den Verein) ein durchlaufender Posten ohne lohn- oder umsatzsteuerliche Konsequenzen.

- **Das Problem mit der Kassenschublade:** Gesetzlich müsste Personal-Trinkgeld gar nicht zwingend über die Kasse erfasst werden. Wirft der Kellner sein Trinkgeld im Festzelt-Trubel aber in die _gemeinsame_ physische Kassenlade, **muss** deine Kassensoftware dies zwingend aufzeichnen. Tust du das nicht, stimmt beim Tagesabschluss der gezählte Ist-Bestand nicht mehr mit dem Soll-Bestand überein und die Kassensturzfähigkeit ist gefährdet.
- **Die technische Lösung:** Du musst eine dedizierte Trinkgeld-Funktion in deine Kassen-App einbauen. Wird hier ein Betrag eingegeben, buchst du ihn mit dem DSFinV-K Typ **`TrinkgeldAN`**.
- **Ein- und Auszahlung:** Da dieses Geld dem Verein nicht gehört, wird es treuhänderisch verwaltet. Nimmt der Kellner sein gesammeltes Trinkgeld am Schichtende aus der Lade mit nach Hause, buchst du diese Entnahme in der Software **ebenfalls** über den Typ `TrinkgeldAN`, um den Kassenbestand wieder auszugleichen (dieser Typ darf laut DSFinV-K sowohl für Ein- als auch Auszahlungen des Personals genutzt werden).
- **Architektonische Warnung:** Du darfst für Trinkgelder nicht einfach einen regulären "Freien Artikel" namens "Trinkgeld" in der Artikeldatenbank anlegen. Die Erfassung funktioniert nur dann finanzamtkonform, wenn die dafür tief im Code verankerte Trinkgeld-Funktion mit dem korrekten `GV_TYP` genutzt wird und keine separate Artikelposition manuell hinzugebucht wird.
