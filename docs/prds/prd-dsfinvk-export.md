# PRD: DSFinV-K-Export (F-04)

> Herkunft: anforderungen.md F-04 (Fiskalkonformität, Phase 2). Schließt die zwei offenen Kriterien von F-06 (Abrechnungskreis) mit ab.
> Quellen: DSFinV-K-Spezifikation v2.4 (Volltext geprüft), BZSt-Bekanntmachungen zu v2.5 (aktuell verbindlich, Stand 2026) und 3.0-Diskussionsentwurf (unverbindlich); GoBD-Anlage „Ergänzende Informationen zur Datenträgerüberlassung" (index.xml/DTD). compliance.md §6.
> Klärungsentscheidungen (Session 2026-06-16): Zielversion v2.5 mit konfigurierbarem Versionsstring; ein Export = genau eine Kassensitzung; TSE-Stammdaten beim TSE-Setup persistieren; F-06 hier mit abschließen; Modulschnitt wie unten; Tests für Mapper, CSV-Serializer und index.xml-Generator.

## Problem Statement

Bei einer Kassen-Nachschau oder Betriebsprüfung verlangt die Finanzverwaltung die strukturierten Kassendaten in einem genormten, maschinell auswertbaren Format (DSFinV-K), das die Prüfsoftware IDEA einlesen kann. jotti zeichnet alle Geschäftsvorfälle bereits revisionssicher im Kassenjournal auf und signiert sie über die TSE, kann sie aber nicht in diesem Format herausgeben. Ohne den Export ist ein Verein im Prüfungsfall nicht auskunftsfähig, obwohl alle Daten vorliegen. Der Verein steht dann vor einer manuellen, fehleranfälligen Aufbereitung, die im Festzeltbetrieb niemand leisten kann, und riskiert Schätzungen oder ein Bußgeld nach § 379 AO.

Erschwerend kommen zwei Lücken hinzu, die heute niemandem auffallen, weil es noch keinen Export gibt:

- Die `tse.csv` (Stamm_TSE) verlangt Signaturalgorithmus, Public Key und Zertifikat der TSE. jotti speichert davon nur die TSE-Seriennummer abfragbar; Algorithmus und Public Key liegen aktuell nur als Bestandteil des QR-Code-Strings vor, das Zertifikat gar nicht.
- F-06 ist konzeptionell geklärt (jede Tisch-Session ist ein Abrechnungskreis), aber zwei seiner Akzeptanzkriterien sind erst prüfbar, wenn der Export existiert und den `ABRECHNUNGSKREIS` korrekt ausweist.

## Solution

jotti erhält einen Admin-Endpunkt, der für eine gewählte Kassensitzung ein vollständiges DSFinV-K-Archiv erzeugt: ein ZIP mit den vorgeschriebenen CSV-Dateien (offizielle englische Kleinschreib-Dateinamen, Semikolon-getrennt), einer beschreibenden `index.xml` und der zugehörigen `gdpdu-01-09-2004.dtd`. Der Admin wählt im Reporting-Bereich eine Kassensitzung (Standard: die aktuelle) und lädt das Archiv herunter. Inhaltlich entspricht ein Archiv genau einem Kassenabschluss (Z-Bon): `Z_NR` ist die `kassensitzung_nr`.

Der Export liest ausschließlich vorhandene Daten: das append-only Kassenjournal, die Betreiber-Stammdaten, die Kassenidentität (Seriennummer), die Steuersätze und die TSE-Signaturen (aus Event-Payload und Nachsignier-Seitentabelle vereint). Ergänzend wird beim TSE-Setup eine kleine TSE-Stammdaten-Tabelle (Algorithmus, Public Key, Zertifikat) befüllt, damit die `tse.csv` vollständig ist und der Export offline-fähig bleibt.

Der Kern ist eine reine Transformation ohne Seiteneffekte: Events plus Stammdaten ergeben typisierte Tabellenzeilen, diese werden DSFinV-K-konform serialisiert und mit Beschreibungsdateien zu einem ZIP gepackt. Dadurch ist der fiskalisch heikle Teil (Mapping, Steueraufteilung, Storno-Verkettung, Abrechnungskreis) isoliert und über Golden-File-Tests nachweisbar korrekt.

## User Stories

### Admin: Export auslösen

1. Als Admin möchte ich für eine gewählte Kassensitzung ein DSFinV-K-Archiv herunterladen, damit ich es bei einer Prüfung vorlegen kann.
2. Als Admin möchte ich standardmäßig die aktuelle (offene) Kassensitzung vorausgewählt bekommen, damit ich ohne Suchen exportieren kann.
3. Als Admin möchte ich auch eine bereits abgeschlossene Kassensitzung exportieren können, damit ich vergangene Betriebstage nachreichen kann.
4. Als Admin möchte ich, dass nur ich (Rolle `admin`) den Export auslösen kann, damit fiskalische Rohdaten nicht für Servicekräfte zugänglich sind.
5. Als Admin möchte ich eine aussagekräftige Fehlermeldung erhalten, wenn die gewählte Kassensitzung nicht existiert oder leer ist, statt eines defekten Archivs.
6. Als Admin möchte ich, dass das ZIP einen klaren, sprechenden Dateinamen trägt (Seriennummer plus Kassensitzung plus Zeitstempel), damit ich Archive eindeutig zuordnen kann.

### Betriebsprüfer: Daten einlesen

7. Als Betriebsprüfer möchte ich das Archiv ohne Anpassung in IDEA importieren können, weil `index.xml` und Dateistruktur dem Standard entsprechen.
8. Als Betriebsprüfer möchte ich die offiziellen englischen Dateinamen (`transactions.csv`, `lines.csv` usw.) vorfinden, damit mein Prüftool die Tabellen erkennt.
9. Als Betriebsprüfer möchte ich pro Bon Brutto-, Netto- und Steuerbeträge nach Steuerschlüsseln aufgeschlüsselt sehen, damit ich die Umsatzsteuer nachvollziehen kann.
10. Als Betriebsprüfer möchte ich jede TSE-Transaktion mit Transaktionsnummer, Signaturzähler und Signatur vorfinden, damit ich die Unveränderbarkeit prüfen kann.
11. Als Betriebsprüfer möchte ich Stornierungen als eigene Negativ-Datensätze mit Referenz auf den Ursprungsbon vorfinden, damit das Radierverbot eingehalten ist.
12. Als Betriebsprüfer möchte ich den Betreiber (Name, Anschrift, Steuernummer) als Stammdaten korrekt ausgewiesen sehen.
13. Als Betriebsprüfer möchte ich jeden Bon einem `ABRECHNUNGSKREIS` zugeordnet sehen (Tisch bzw. Theke), damit ich Bestellungen und Zahlungen eines Tisches als Einheit nachvollziehen kann.
14. Als Betriebsprüfer möchte ich den Kassenabschluss (Z-Bon) mit aggregierten Geschäftsvorfall- und Zahlart-Summen vorfinden, damit ich die Tagessumme gegen die Einzelbons abgleichen kann.

### Verein / Betreiber: Vertrauen und Betrieb

15. Als Vereins-Admin möchte ich den Export jederzeit selbst auslösen können, ohne Entwickler oder Steuerberater, damit ich im Prüfungsfall sofort auskunftsfähig bin.
16. Als Vereins-Admin möchte ich, dass der Export auch funktioniert, wenn gerade kein Internet verfügbar ist, weil alle nötigen Daten lokal vorliegen.
17. Als Vereins-Admin möchte ich, dass das Archiv selbsterklärend und ohne proprietäre Software lesbar ist (CSV plus XML), damit es Teil meiner 10-Jahres-Aufbewahrung sein kann.

### System / Datenintegrität

18. Als System möchte ich signierte Vorgänge sowohl aus dem Event-Payload als auch aus der Nachsignier-Seitentabelle einsammeln, damit auch nach einem TSE-Ausfall nachsignierte Vorgänge im Export erscheinen.
19. Als System möchte ich beträge konsistent in Cent aus dem Journal übernehmen und erst bei der Serialisierung als Dezimalzahl mit Punkt darstellen, damit keine Rundungsfehler entstehen.
20. Als System möchte ich `kombi`-Positionen beim Export in ihre Steueranteile (70 % zu 7 %, 30 % zu 19 %) entfalten, damit die USt-Aufschlüsselung stimmt.
21. Als System möchte ich Direktverkäufe als eigene Belegvorgänge abbilden (positiver Verkauf und Storno je eigener Beleg mit Referenz), damit der theke-seitige Barverkauf konform ist.
22. Als System möchte ich nicht zutreffende Tabellen (Terminals, Agenturen) weglassen und nicht in der `index.xml` deklarieren, damit das Archiv nur tatsächlich vorhandene Daten beschreibt.
23. Als System möchte ich offene Bestellungen (Forderungsentstehung) und ihre spätere Zahlung (Forderungsauflösung) korrekt als getrennte Geschäftsvorfälle abbilden, damit der gastronomische Tisch-Ablauf konform ist.

### TSE-Stammdaten (Voraussetzung)

24. Als System möchte ich auf jedem Einrichtungspfad (Neuanlage wie Übernahme einer vorhandenen TSS) Signaturalgorithmus, Public Key und Zertifikat der TSE persistieren, damit die `tse.csv` auch bei per Übernahme onboardeten Instanzen vollständig ist.
25. Als Vereins-Admin möchte ich, dass diese Stammdaten ohne Zusatzschritt im Rahmen der TSE-Einrichtung gespeichert werden, damit der Export später ohne fiskaly-Verbindung funktioniert.

## Implementation Decisions

### Module (Deep-Module-Schnitt, wie abgestimmt)

- **DSFinV-K-Mapper** (rein, kein I/O): Kernmodul. Eingabe sind die Events einer Kassensitzung plus ein Stammdaten-Snapshot (Betreiber, Seriennummer, TSE-Stammdaten, Steuersätze, Tischnamen). Ausgabe sind typisierte Zeilen-Kollektionen für alle erzeugten Tabellen. Hier liegen GV-Typ/Beleg-Typ-Mapping, Steueraufteilung inklusive `kombi`-Entfaltung, Storno als Negativ-Bon mit Referenz, Ableitung des `ABRECHNUNGSKREIS` aus dem Subject sowie die Vereinigung von Event-`TSEData` und Nachsignier-Seitentabelle.
- **CSV-Serializer** (rein, generisch): typisierte Zeilen zu DSFinV-K-konformen Bytes. Verantwortet Trennzeichen, Zeilenende, Header, Spaltenreihenfolge, Zahlen- und Textdarstellung. Für alle Tabellen wiederverwendbar.
- **index.xml-/DTD-Generator** (rein): erzeugt den GDPdU-Descriptor (deklariert nur die vorhandenen Tabellen mit Spalten, Typen und Trennzeichen) und liefert die statische `gdpdu-01-09-2004.dtd`.
- **ZIP-Packer** (dünn): bündelt serialisierte CSVs, `index.xml` und `.dtd` zu einem ZIP-Stream.
- **Export-Orchestrator** (App-Service): lädt Events und Stammdaten, ruft Mapper, Serializer, index-Generator und ZIP-Packer, liefert das Archiv.
- **Admin-Handler**: `GET /admin/export/dsfinvk?kassensitzung=N`, Rolle `admin`, streamt das ZIP mit passendem Dateinamen und Content-Type. Ohne Parameter gilt die aktuelle Kassensitzung.
- **TSE-Stammdaten-Persistenz**: neue Singleton-Tabelle (Signaturalgorithmus, Public Key, Zertifikat, Format-/Zeitformat-Angaben), befüllt im Zuge des TSE-Setups. Speist `tse.csv`. Der Hook hängt am gemeinsamen „Konfiguration speichern"-Schritt, den alle Einrichtungspfade durchlaufen (Neuanlage `RichteTSEEin` ebenso wie Übernahme `UebernimmTSE` und PUK-Reset), nicht am Anlage-Lebenszyklus. Sonst bekämen per Übernahme onboardete Instanzen ein unvollständiges `tse.csv`. Die Stammdaten werden heute im Setup nicht gelesen (`TSSInfo` trägt nur `ID` und `State`); die Persistenz erfordert eine zusätzliche fiskaly-Leseoperation auf der TSS-Ressource und eine Erweiterung des `SetupClient`-Interface.

### Format- und Strukturregeln (autoritativ verifiziert)

- Zielversion **v2.5** (aktuell verbindlich). Der Versionsstring (in `index.xml` bzw. `cashregister.csv`) wird konfigurierbar gehalten, da die Tabellenstruktur seit v2.0 stabil ist und v3.0 als Entwurf am Horizont steht.
- ZIP enthält CSV-Dateien, **`index.xml`** und zwingend die **`gdpdu-01-09-2004.dtd`**.
- CSV-Regeln: Header-Zeile zwingend; Semikolon als Trennzeichen; CRLF; Punkt als Dezimaltrennzeichen; keine Tausendertrennzeichen; mindestens eine Stelle vor dem Punkt; Beträge mit zwei Dezimalstellen (technisch bis fünf zulässig); Spaltenreihenfolge exakt nach Spezifikation; UTF-8. Trennzeichen und Feldtypen werden zusätzlich in `index.xml` deklariert.
- Systemspezifische Zusatzfelder, falls nötig, nur am Zeilenende und in `index.xml` definiert.

### Pflicht-Dateiumfang (vollständige v2.x-Liste, 20 Tabellen)

Offizieller Dateiname / logische Bezeichnung / jotti-Behandlung:

Stammdatenmodul

- `cashpointclosing.csv` / Stamm_Abschluss / Z-Bon-Metadaten der Kassensitzung
- `location.csv` / Stamm_Orte / Betriebsstätte aus Betreiber-Stammdaten
- `cashregister.csv` / Stamm_Kassen / Seriennummer, Software-Typ und -Version
- `slaves.csv` / Stamm_Terminals / nicht zutreffend, weggelassen
- `pa.csv` / Stamm_Agenturen / nicht zutreffend (kein Agenturgeschäft), weggelassen
- `vat.csv` / Stamm_USt / verwendete Steuersätze (19, 7, 0)
- `tse.csv` / Stamm_TSE / Seriennummer, Algorithmus, Public Key, Zertifikat

Kassenabschlussmodul

- `businesscases.csv` / Z_GV_TYP / aggregierte Beträge je Geschäftsvorfalltyp nach Steuersätzen
- `payment.csv` / Z_Zahlart / aggregierte Zahlart-Summen
- `cash_per_currency.csv` / Z_WAEHRUNGEN / Bargeldbestand je Währung (EUR)

Einzelaufzeichnungsmodul

- `transactions.csv` / Bonkopf / ein Datensatz je Bon
- `allocation_groups.csv` / Bonkopf_AbrKreis / Zuordnung Bon zu `ABRECHNUNGSKREIS`
- `transactions_vat.csv` / Bonkopf_USt / USt-Aufschlüsselung je Bon
- `datapayment.csv` / Bonkopf_Zahlarten / Zahlarten je Bon
- `lines.csv` / Bonpos / Artikelzeilen
- `lines_vat.csv` / Bonpos_USt / USt-Aufschlüsselung je Zeile
- `itemamounts.csv` / Bonpos_Preisfindung / nur befüllt bei Preisfindung, sonst header-only oder weggelassen
- `subitems.csv` / Bonpos_Zusatzinfo / nur befüllt bei Zusatzinfos, sonst header-only oder weggelassen
- `references.csv` / Bon_Referenzen / Storno- und sonstige Referenzen
- `transactions_tse.csv` / TSE_Transaktionen / TSE-Transaktionsnummer, Signaturzähler, Signatur

Schlüsselfelder `Z_KASSE_ID`, `Z_ERSTELLUNG`, `Z_NR`, `BON_ID` in den führenden Spalten gemäß Spezifikation.

### Event-zu-DSFinV-K-Mapping (im Mapper)

- `bestellung-aufgenommen` ist eine Forderungsentstehung (processType `Bestellung-V1`), kein sofortiger Umsatz.
- `zahlung-kassiert` löst die Forderung auf und realisiert den Umsatz (processType `Kassenbeleg-V1`), aufgeschlüsselt nach Steuerschlüsseln.
- `direktverkauf-getaetigt` ist ein eigener Beleg (`Kassenbeleg-V1`); `direktverkauf-storniert` ein negativer Beleg mit Referenz.
- `stornierung-erteilt` erzeugt einen Negativ-Datensatz mit `BON_STORNO` und `REF_BON_ID` auf den Ursprung, eigene TSE-Signatur, gleicher `ABRECHNUNGSKREIS`.
- `geldtransit-gebucht` wird Geschäftsvorfalltyp Geldtransit; `differenz-soll-ist-gebucht` (Kassensturz) wird Differenz Soll/Ist; `auszahlung-geleistet` und `anfangsbestand` werden gemäß DSFinV-K-Geschäftsvorfalltypen abgebildet.
- Trinkgeld und Rückgeld (K-10) sind rein clientseitig und nicht fiskalisch persistiert; sie erscheinen nicht als eigener Geschäftsvorfall.
- Die vollständige Belegtyp- und Geschäftsvorfalltyp-Zuordnung folgt DSFinV-K Anhang C und E und ist Teil des Mapper-Moduls; Golden-Tests sichern sie ab.

### Abrechnungskreis (F-06-Abschluss)

- Der `ABRECHNUNGSKREIS` wird aus dem Subject abgeleitet (Tisch-Session zu Tischname, z. B. `Tisch 42`).
- Jede TSE-Transaktion ist über ihren Bon einem `ABRECHNUNGSKREIS` zugeordnet; im Export in `allocation_groups.csv` ausgewiesen.
- Direktverkäufe tragen keinen `ABRECHNUNGSKREIS` (Feld optional). Annahme, dokumentiert.

### Schema-Änderung

- Neue Singleton-Tabelle für TSE-Stammdaten (Algorithmus, Public Key, Zertifikat, Log-Time-Format). Befüllt am gemeinsamen Speicher-Schritt aller Einrichtungspfade. Keine Änderung am Kassenjournal.

### API-Kontrakt

- `GET /admin/export/dsfinvk` mit optionalem Query-Parameter `kassensitzung` (Default: aktuelle). Rolle `admin`. Antwort: `200` mit `application/zip` und `Content-Disposition`-Dateiname; `404`, wenn die Kassensitzung nicht existiert; `403` ohne Admin-Rolle.

## Testing Decisions

Ein guter Test prüft beobachtbares externes Verhalten (die erzeugten Zeilen und Bytes), nicht die innere Mechanik. Golden-File-Tests passen hier besonders, weil die erwartete Ausgabe ein stabiler, fachlich prüfbarer Vertrag ist.

Getestet werden (wie abgestimmt):

- **DSFinV-K-Mapper**: Golden-Szenarien je relevantem Ablauf, mindestens einfacher Verkauf, Storno mit Referenz, `kombi`-Steueraufteilung 70/30, Direktverkauf und Direktverkauf-Storno, sowie ein über die Nachsignier-Seitentabelle nachsignierter Vorgang nach TSE-Ausfall. Geprüft werden die erzeugten Zeilen je Tabelle (Beträge, Steuerschlüssel, `BON_STORNO`, `REF_BON_ID`, `ABRECHNUNGSKREIS`).
- **CSV-Serializer**: Formatregeln (Semikolon, CRLF, Dezimalpunkt, keine Tausendertrenner, Spaltenreihenfolge, Header, Escaping von Sonderzeichen und Trennzeichen im Inhalt).
- **index.xml-/DTD-Generator**: korrekter Descriptor für genau die vorhandenen Tabellen samt Spalten und Typen; weggelassene Tabellen (Terminals, Agenturen) werden nicht deklariert; eingebettete DTD.

Nicht als eigener automatisierter Test vorgesehen (bewusste Entscheidung): Orchestrator und Handler werden über die bestehende manuelle bzw. spätere Integrationsprüfung abgedeckt, nicht über einen dedizierten Integrationstest in diesem Scope.

Prior Art im Repo: die Aggregations-Tests im Reporting-Modul, der Event-JSON-Contract-Test, der TSE-Roundtrip-Test und die Steuermatrix-Tests (erwartete Aufteilungen als Golden-Werte) zeigen den Stil.

## Out of Scope

- Validierung gegen die Original-IDEA-Software oder ein externes Prüftool als Teil der CI. Manuell wünschenswert, aber nicht Bestandteil dieses PRD.
- DSFinV-K v3.0 (noch Diskussionsentwurf, unverbindlich). Nur durch konfigurierbaren Versionsstring vorbereitet.
- 10-Jahres-Archiv-Bundle (F-10) und eBeleg (F-09). Der DSFinV-K-Export ist ein Baustein für F-10, F-10 selbst bleibt separat.
- GoBD-Integritäts-Selbsttest (F-08), eigenes Item.
- Agenturgeschäft und Slave-/Terminal-Kassen (`pa.csv`, `slaves.csv`). Für jottis Betriebsmodell gegenstandslos.
- Programmatische Übermittlung an die Finanzverwaltung. Der Export ist ein Download für Nachschau und Prüfung.

## Further Notes

- Die aus der Quellen-Evaluation abgeleiteten Korrekturen sind bereits in die Bestandsdokumente eingepflegt (2026-06-16): anforderungen.md F-04 und compliance.md §§2.5, 6.1 auf v2.5; Pflicht-DTD `gdpdu-01-09-2004.dtd` in §6.2 und §6.7 ergänzt; Dateiliste in §6.3 um `slaves.csv`, `pa.csv`, `itemamounts.csv`, `subitems.csv` vervollständigt (`slaves`/`pa` als für jotti gegenstandslos markiert); Versionsbezüge in produktbeschreibung.md und language.md angeglichen.
- Die im Spec-Fließtext vereinzelt auftauchenden Schreibweisen `Bonkopf.csv` / `Bonpos.csv` sind Spec-interne Inkonsistenzen; kanonisch sind laut Inhaltsverzeichnis `transactions.csv` und `lines.csv`. jottis bisherige Festlegung (englische Dateinamen) ist korrekt.
- v2.4 brachte gegenüber v2.3 keine inhaltlichen Änderungen (nur AEAO-redaktionell); die Tabellenstruktur ist seit v2.0 stabil. Damit ist jottis Datenmodell vorwärtskompatibel, und der Versionswechsel betrifft im Wesentlichen den deklarierten Versionsstring.
- Die TSE-Stammdaten-Persistenz (User Stories 24, 25) ist eine Voraussetzung des Exports und berührt das TSE-Setup. Sie überschneidet sich direkt mit dem gerade umgesetzten [plan-tse-setup-recovery.md](../plans/plan-tse-setup-recovery.md), der dieselben Dateien ändert (`setup.go`, `domain/tse`-`SetupClient`, `fiskaly_setup.go`) und neue Einrichtungspfade einführt (PIN-freie Übernahme F8, PUK-Reset), die den Anlage-Lebenszyklus überspringen. Reihenfolge: die Persistenz nach bzw. mit diesem Plan umsetzen und an den gemeinsamen Speicher-Schritt hängen. Der Plan deklariert „keine Schema-Änderung" nur für seinen eigenen Scope; die neue Stammdaten-Tabelle steht dazu nicht im Widerspruch.
- Die harte LIVE-Sperre des Recovery-Plans (genau eine LIVE-TSS pro Kasse) sichert ab, dass `tse.csv` im Produktivfall genau eine Stamm_TSE-Zeile hat. TEST kann mehrere TSS sammeln, ist aber steuerlich ungültig und nicht prüfungsrelevant.
