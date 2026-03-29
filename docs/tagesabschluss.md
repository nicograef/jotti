# Tagesabschluss

Der **Tagesabschluss** (Z-Bon) ist die aggregierende Zusammenfassung aller abgeschlossenen Kassenvorgänge (Vorgangstyp „Beleg") für einen Festtag oder eine Schicht. Er stellt die Verbindung zwischen den Kasseneinzelbewegungen und der Tagessumme für die Finanzbuchhaltung her. Rechtsgrundlagen: AO, GoBD, KassenSichV, DSFinV-K.

## §1 Rechtliche Anforderungen

- **Nur abgeschlossene Belege:** In den Tagesabschluss fließen ausschließlich Vorgänge vom Typ „Beleg". Offene Tisch-Bestellungen (Vorgangstyp „AVBestellung") bleiben unberücksichtigt, bis sie bezahlt werden — erst dann entsteht ein Beleg, der in den Abschluss des jeweiligen Tages einfließt.
- **Z-Bon vs. X-Bon:** Der Z-Bon setzt den internen Kassenumsatzspeicher auf null zurück. Ein Zwischenbericht (X-Bon) ersetzt ihn rechtlich nicht.
- **Fortlaufende Nummerierung (`Z_NR`):** Aufsteigend, lückenlos, nicht zurücksetzbar. Dient als Primärschlüssel und Foreign Key in allen DSFinV-K-Export-Tabellen.
- **Stammdaten-Snapshot:** Zu jedem Abschluss müssen die aktuell gültigen Stammdaten (Steuersätze, TSE-Zertifikate, Kassen-IDs) eingefroren werden. **Regel:** Vor jeder Stammdaten-Änderung (z.B. Preisanpassung) muss zwingend erst ein Kassenabschluss durchgeführt werden.
- **Aufbewahrung:** Z-Bons müssen unveränderbar für 10 Jahre archiviert werden.
- **Kassensturzfähigkeit:** Der Kassenbestand muss jederzeit überprüfbar sein (Soll-Ist-Abgleich). Differenzen müssen als eigener Geschäftsvorfall (`DifferenzSollIst`) gebucht werden.

## §2 Kassensturz (Soll-Ist-Abgleich)

- **Soll-Bestand:** Anfangsbestand + Bareinnahmen − Barausgaben (vom System errechnet).
- **Ist-Bestand:** Manuell gezählter Bargeldbestand, eingegeben über ein digitales Zählprotokoll.
- **Invariante:** Der Kassenbestand darf niemals negativ sein.
- **Hinweis für Betriebsprüfer:** Eine Kasse, die „immer auf den Cent genau stimmt", gilt als manipulationsverdächtig. Die Finanzverwaltung sucht mit IDEA-Software gezielt nach fehlenden Differenzen.

### Differenzbuchung

Ergibt der Abgleich eine Abweichung, darf der Soll-Bestand **nicht** stillschweigend per `UPDATE` überschrieben werden. Stattdessen:

1. Einen **Eigenbeleg** generieren (neuer, automatisierter Datensatz).
2. Diesen über die **TSE** kryptografisch signieren lassen.
3. Im DSFinV-K Export als `GV_TYP = DifferenzSollIst` deklarieren (in `lines.csv` und `businesscases.csv`). Fehlbetrag = negativer Wert, Überschuss = positiver Wert.
4. **Steuerliche Behandlung:** Umsatzsteuerlich neutral exportieren (`NichtSteuerbar` / `UmsatzsteuerNichtErmittelbar`). Die steuerliche Qualifikation obliegt nachgelagerten Systemen (z.B. DATEV).

## §3 Geldtransit

Bargeldentnahme während des Betriebs (z.B. Abschöpfung in den Tresor) ist eine reine Vermögensumschichtung.

- **Belegpflicht:** Eigenbeleg erzeugen und über TSE signieren lassen. „Keine Buchung ohne Beleg."
- **DSFinV-K:** `GV_TYP = Geldtransit`. Bei Überführung in den privaten Bereich stattdessen `Privatentnahme`.
- **Steuer:** Nicht steuerbar (USt-Schlüssel ID 5).
- **Kassensturzfähigkeit:** Durch die Erfassung sinkt der Soll-Bestand passend zum verbleibenden physischen Ist-Bestand.

## §4 Trinkgelder

Es muss architektonisch strikt zwischen Unternehmer- und Arbeitnehmer-Trinkgeld unterschieden werden. Trinkgelder dürfen **nicht** als regulärer Artikel in der Artikeldatenbank angelegt werden — nur eine dedizierte Trinkgeld-Funktion mit korrektem `GV_TYP` ist finanzamtkonform.

### A. Unternehmer-Trinkgeld (`TrinkgeldAG`)

Trinkgeld für den Vereinsvorstand/Inhaber ist steuerlich Teil des regulären Umsatzes und **umsatzsteuerpflichtig**. Die spätere physische Entnahme wird über `Geldtransit` oder `Privatentnahme` gebucht.

### B. Arbeitnehmer-Trinkgeld (`TrinkgeldAN`)

Trinkgeld für Helfer/Bedienungen ist für den Verein ein durchlaufender Posten ohne lohn- oder umsatzsteuerliche Konsequenzen.

- **Problem:** Wird Trinkgeld in die gemeinsame Kassenlade gelegt, **muss** es erfasst werden, sonst ist die Kassensturzfähigkeit gefährdet (Ist ≠ Soll).
- **Lösung:** Einzahlung und Auszahlung (Entnahme durch Kellner am Schichtende) beide über `GV_TYP = TrinkgeldAN` buchen.

## §5 DSFinV-K Export (technisch)

### Kassenabschlussmodul

| CSV-Datei | Inhalt |
|---|---|
| `businesscases.csv` (Z_GV_Typ) | Tagesumsätze aggregiert nach Geschäftsvorfalltyp und USt-Schlüssel |
| `payment.csv` (Z_Zahlart) | Summen nach Zahlungsarten (Bar, EC etc.) |
| `cash_per_currency.csv` | Kassenbestand nach Währungen |
| `cashpointclosing.csv` (Stamm_Abschluss) | Stammdaten-Snapshot (Hardware, Zertifikate, Konfiguration) |

### Primärschlüssel

Alle Export-Zeilen müssen die Schlüssel `Z_KASSE_ID`, `Z_ERSTELLUNG` und `Z_NR` mitführen, damit die Daten in der Prüfsoftware relational verknüpft werden können.

### TSE-Absicherung

Der Z-Bon wird als eigene TSE-Transaktion mit `processType = SonstigerVorgang-V1` signiert.

## §6 Betreiber-Anleitung (Ablauf Tagesabschluss)

1. **Bargeld zählen** — physische Bestandsaufnahme mit Zählprotokoll.
2. **Tagesabschluss auslösen** — Z-Bon im System erstellen (kein X-Bon).
3. **Abgleich** — gezähltes Bargeld (Ist) mit Z-Bon-Betrag (Soll) vergleichen.
4. **Differenzen buchen** — Abweichungen im System erfassen.
5. **Archivieren** — Z-Bon und Zählprotokoll für 10 Jahre sichern.
