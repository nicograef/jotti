# PRD: Zusätzliche Drucker für einzelne Produkte

## Problem Statement

Bei einem Fest gibt es oft mehrere Ausgabebereiche, und jeder braucht seinen eigenen Bondrucker. jotti leitet einen Arbeitsbon allein über die Produktkategorie an einen Drucker. Es gibt drei feste Kategorien, also höchstens drei Drucker für Arbeitsbons aus Tischbestellungen.

Ein Verein stieß im Echtbetrieb an diese Grenze. Sein Weinfest hatte 1603 Tischbestellungen, 970 Direktverkäufe und drei räumlich getrennte Ausgabebereiche. Der Bericht nennt drei Fälle außerhalb des Kategoriemodells. Die ersten zwei traten am Fest auf, der dritte folgt aus dem Schema:

1. Ein Gericht stand unter „Getränke“. Es wurde im Backhaus zubereitet, und dort stand der Getränkedrucker.
2. Der Bereich für Kaffee und Kuchen bekam keinen eigenen Drucker. Dort führte das Team eine Strichliste.
3. Zwei Essensausgaben mit je eigenem Drucker lassen sich nicht abbilden.

Die falsche Kategorie kostet an drei Stellen. Die Servicekraft sucht das Gericht unter dem Getränke-Reiter. Die Produktstatistik zählt es als Getränk. Der DSFinV-K-Export meldet die Kategorie als Warengruppe.

Die meisten Vereine brauchen keine weiteren Drucker. Ein Drucker je Kategorie samt Abholbon und Kassenbeleg deckt ihr Fest ab. Diese Einrichtung muss genau so einfach bleiben, wie sie ist.

Zwei weitere Punkte des Berichts sind gebaut. Der Piepton ist entfallen. Der Abholbon kennt den Bonmodus „Kein Bon“. Mit ihm druckt der Direktverkauf an der Zapfanlage nichts. Dieses PRD behandelt nur die Bindung des Druckers an die Kategorie.

## Solution

Die Seite „Bondrucker“ bleibt, wie sie ist. Jede Kategorie hat ihren Drucker, dazu kommen Abholbon und Kassenbeleg. Wer nichts weiter braucht, sieht nur eine Karte mehr.

Diese Karte „Zusätzlicher Drucker“ steht unter den Druckstationen. Dort legt der Admin bei Bedarf einen benannten Drucker an: Name, Drucker-IP und Bonmodus. Dem Drucker ordnet er einzelne Produkte zu.

Ein Produkt mit Zuordnung druckt an seinem zusätzlichen Drucker. Jedes andere Produkt druckt wie heute am Drucker seiner Kategorie. Die Kategorie bleibt so fachlich richtig.

| Fall aus dem Bericht             | Einrichtung                                                                                                                           |
| -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| 1. Gericht am Getränkedrucker    | Das Gericht wandert nach „Essen“. Ein zusätzlicher Drucker „Backhaus“ trägt die IP des Druckers dort. Ihm ist das Gericht zugeordnet. |
| 2. Kaffee und Kuchen mit Drucker | Ein zusätzlicher Drucker „Kaffee & Kuchen“ bekommt die Produkte Kaffee und Kuchen. Ihre Kategorien bleiben Getränke und Essen.        |
| 3. Zwei Essensausgaben           | Der Essen-Drucker bedient die erste Ausgabe. Ein zusätzlicher Drucker „Ausgabe 2“ bekommt die Gerichte der zweiten.                   |

Ein zusätzlicher Drucker ohne IP druckt nicht. Seine Produkte erzeugen dann keinen Arbeitsbon. So kann ein Bereich ohne Drucker bewusst stumm bleiben.

Tragen ein Kategoriedrucker und ein zusätzlicher Drucker dieselbe IP, kommen ihre Bons getrennt. Im Fall 1 druckt eine Bestellung mit Wein und Gericht zwei Bons im Backhaus. Das gilt heute schon für zwei Kategorien an einem Drucker.

Wird ein zusätzlicher Drucker gelöscht, drucken seine Produkte wieder am Drucker ihrer Kategorie. Nach dem Update ist nichts einzustellen, und nichts ändert sich.

## User Stories

1. Als Admin eines gewöhnlichen Fests möchte ich die Seite „Bondrucker“ wie gewohnt vorfinden, damit ich nichts Neues lernen muss.
2. Als Admin möchte ich einen zusätzlichen Drucker mit eigenem Namen anlegen, damit ein weiterer Ausgabebereich seine Bons bekommt.
3. Als Admin möchte ich einzelne Produkte einem zusätzlichen Drucker zuordnen, damit ihre Kategorie fachlich richtig bleibt.
4. Als Admin möchte ich am zusätzlichen Drucker sehen, welche Produkte dort drucken, damit ich den Aufbau vor Ort prüfe.
5. Als Admin möchte ich an jeden zusätzlichen Drucker einen Testbon schicken, damit ich Drucker und Netzwerk vor dem Fest prüfe.
6. Als Admin möchte ich einen zusätzlichen Drucker ohne IP anlegen, damit ich den Aufbau vor dem Fest vorbereite.
7. Als Admin möchte ich einen zusätzlichen Drucker löschen, damit seine Produkte wieder am Drucker ihrer Kategorie drucken.
8. Als Helfer an einer Ausgabe möchte ich nur die Bons meines Bereichs bekommen, damit ich nichts aussortieren muss.
9. Als Servicekraft möchte ich jedes Gericht unter „Essen“ finden, damit ich unter Stress nicht suchen muss.
10. Als Vorstand möchte ich eine Produktstatistik mit richtigen Kategorien, damit die Auswertung des Fests stimmt.
11. Als Betreiber möchte ich nach dem Update nichts umstellen, damit das nächste Fest wie gewohnt druckt.

## Implementation Decisions

### Abwägung

| Variante                                | Umfang                                                                                        | Aufwand | Risiko                                                             | Umkehrbarkeit                                                |
| --------------------------------------- | --------------------------------------------------------------------------------------------- | ------- | ------------------------------------------------------------------ | ------------------------------------------------------------ |
| Z: Zusatzdrucker (gewählt)              | Druckstationen bleiben; optionale benannte Drucker, denen einzelne Produkte zugeordnet sind   | gering  | gering, ohne Zusatzdrucker läuft der heutige Druckpfad unverändert | hoch, Zusatzdrucker löschen stellt das heutige Verhalten her |
| A: Stationsmodell                       | Stationen ersetzen die Kategoriezeilen; Kategorien und einzelne Produkte zeigen auf Stationen | mittel  | mittel, jede Installation druckt über umkopierte Daten             | gering, Schema und Datenübernahme nur per Backup zurück      |
| C: Produkt wählt einen Kategoriedrucker | Ein Produkt druckt am Drucker einer anderen Kategorie                                         | gering  | gering                                                             | hoch, eine optionale Zuordnung                               |
| 0: nicht bauen                          | Piepton-Wegfall und „Kein Bon“ genügen                                                        | keiner  | Kategorie bleibt Steuerungsmittel                                  | –                                                            |

| Fall                             | Z   | A   | C                                         |
| -------------------------------- | --- | --- | ----------------------------------------- |
| 1. Gericht am Getränkedrucker    | ja  | ja  | ja                                        |
| 2. Kaffee und Kuchen mit Drucker | ja  | ja  | nur auf dem Sonstiges-Drucker, falls frei |
| 3. Zwei Essensausgaben           | ja  | ja  | nur auf dem Sonstiges-Drucker, falls frei |

Gewählt ist Variante Z. Sie deckt alle drei Fälle ab und lässt den gewöhnlichen Fall unberührt.

- Variante A deckt dieselben Fälle ab. Sie baut dafür die Einrichtung jedes Vereins um, auch ohne Bedarf. Die Seite bekommt drei Abschnitte, und eine Datenübernahme berührt jede Installation.
- Variante C bleibt bei drei Druckern für Arbeitsbons. Die Fälle 2 und 3 gehen nur einzeln über den Sonstiges-Drucker. Testbon und Seite nennen dann „Sonstiges“, obwohl dort Kaffee druckt.

Die Zuordnung zeigt nur auf Zusatzdrucker, nie auf einen Kategoriedrucker. Ein einziger Zieltyp hält Schema und Oberfläche klein. Teilt sich ein Bereich den Drucker einer Kategorie, trägt der Zusatzdrucker dieselbe IP.

Die Zuordnung hängt am Produkt, nicht an der Variante. Der meldende Verein hätte 19 Produkte statt 107 Varianten zuordnen müssen. Braucht eine Variante einen anderen Ort, wird sie ein eigenes Produkt. Der Bestell-Bildschirm zeigt ohnehin eine flache Variantenliste (D10).

### Die fünf Zeilen der Druckstationen

Die Tabelle der Druckstationen hat genau fünf Zeilen, eine je fester Station:

| Zeile       | Zweck                                                                                                                   | In Variante Z |
| ----------- | ----------------------------------------------------------------------------------------------------------------------- | ------------- |
| Essen       | IP und Bonmodus der Arbeitsbons für Essen, aus Tischbestellungen und Direktverkauf ohne Abholbon                        | unverändert   |
| Getränke    | dasselbe für Getränke                                                                                                   | unverändert   |
| Sonstiges   | dasselbe für Sonstiges                                                                                                  | unverändert   |
| Abholbon    | IP und Bonmodus der Abholbons; „Kein Bon“ schaltet den Direktverkauf stumm, leere IP leitet ihn an die Kategoriedrucker | unverändert   |
| Kassenbeleg | IP des fiskalischen Kassenbeleg-Druckers, ohne Bonmodus                                                                 | unverändert   |

Variante Z braucht alle fünf Zeilen weiter. Sie bleiben der Standard für jedes Produkt ohne Zuordnung. Die Frage nach dem Löschen stellt sich damit nicht.

Erst Variante A machte die drei Kategoriezeilen überflüssig. Ihr Löschen hätte vier Folgen:

- Es wäre eine nicht-additive Migration. Das bräuchte eine ausdrückliche Ausnahme des Owners von der Migrationsregel.
- Postgres kann einen Enum-Wert nicht entfernen. Der Stationstyp behielte die Werte, ein engerer CHECK müsste sie sperren.
- Ein Rückweg auf das alte Release ginge nur per Backup-Restore. Das alte Release schreibt die Kategoriezeilen, der engere CHECK lehnte das ab. Alles seit dem Update Erfasste ginge mit dem Restore verloren.
- Behielte A die Zeilen, stünden tote IPs in der Datenbank. Wer dort nachsieht, hielte sie für wirksam.

### Modell

- Zusatzdrucker: Name, Drucker-IP (IPv4 oder leer), Bonmodus „Pro Position“ oder „Pro Bestellung“.
- Der Name ist Pflicht. Er erscheint in der Verwaltung, auf dem Testbon und an einem fehlgeschlagenen Testbon.
- Der Name muss nicht eindeutig sein. Die Verwaltung zeigt alle Zusatzdrucker auf einer Seite, ein Fehlauftrag nennt auch die IP.
- Ein neuer Zusatzdrucker startet ohne IP und mit „Pro Position“, wie die Druckstationen im Auslieferungszustand.
- Zuordnung: Ein Produkt hat höchstens einen Zusatzdrucker. Ein Zusatzdrucker hat beliebig viele Produkte.
- Ordnet der Admin ein zugeordnetes Produkt einem anderen Zusatzdrucker zu, wandert die Zuordnung dorthin.
- Auflösung: Zusatzdrucker des Produkts vor Druckstation der Kategorie. Die Regel lebt an genau einer Stelle im Druck-Modul.
- Löschen ist endgültig und entfernt die Zuordnungen in derselben Transaktion. Kein Event und kein Druckauftrag verweist auf einen Zusatzdrucker. Ein Soft Delete hätte also keine Historie zu bewahren.
- Die Zuordnung gehört dem Druck-Modul, nicht dem Produkt-Aggregat. Produktdialog und Produkt-Endpunkte bleiben unverändert.

### Schema

Eine additive Migration mit der nächsten freien Nummer, in einer Transaktion:

- Eine Tabelle für Zusatzdrucker mit Name, Drucker-IP und Bonmodus. Der Bonmodus ist TEXT mit CHECK.
- Eine Zuordnungstabelle mit dem Produkt als Primärschlüssel und einem Fremdschlüssel auf den Zusatzdrucker. Löschen kaskadiert auf die Zuordnungen.
- Die Tabelle der Druckstationen bleibt unberührt, samt ihrem CHECK.
- Keine Datenübernahme. Beide Tabellen starten leer, jede Installation druckt danach wie vorher.

Kein Event ändert sich, und es entsteht kein Event-Typ. Das Kassenjournal bleibt unberührt. Die Druckkonfiguration bleibt CRUD wie die heutige Druckstationen-Einstellung.

### Arbeitsbon-Policy

- Die Policy bleibt eine reine Funktion. Sie bekommt zusätzlich je Variante der Bestellung den Zusatzdrucker ihres Produkts.
- Der Aufrufer lädt diese Zuordnung mit einer Abfrage. Die Produkte der Bestellung kennt er schon aus der Anreicherung.
- Die Policy gruppiert Positionen nach Drucker: Zusatzdrucker, sonst Druckstation der Kategorie. Je Drucker gilt dessen Bonmodus.
- Eine Position, deren Drucker keine IP hat, erzeugt keinen Druckauftrag.
- Ohne Zuordnung gruppiert die Policy wie heute nach Kategorie. Die bestehenden Tests pinnen dieses Verhalten.
- Die Auflösung nutzt die Zuordnung im Moment der Bestellung. Ein Arbeitsbon entsteht genau dann und nie ein zweites Mal.
- Direktverkauf: Erst greift „Kein Bon“, dann der Abholbon für alle Positionen. Sonst druckt jede Position an ihrem aufgelösten Drucker.
- Testbon: je Zusatzdrucker mit seinem Namen. Die Referenz des Testbons trägt den Namen, die Liste fehlgeschlagener Aufträge zeigt ihn.
- Arbeitsbons behalten ihre heutige Referenz auf das Event.
- Das Print-Relay ändert sich nicht. Es sieht weiter nur Ziel-IP und Payload.

### API und Verträge

- Alle Endpunkte sind POST und prüfen mit zog im Backend und Zod im Frontend.
- Zusatzdrucker auflisten, anlegen, ändern und löschen.
- Die Liste trägt je Zusatzdrucker seine Produkte mit Namen, fertig vom Backend zusammengestellt. Gelöschte Produkte fehlen darin.
- Produkt einem Zusatzdrucker zuordnen und Zuordnung entfernen.
- Testbon je Zusatzdrucker drucken.
- Die Endpunkte der Druckstationen bleiben unverändert.
- Backend und Frontend ändern sich im selben Release, die API trägt keine Version.

### Admin-UX

Der bestehende Teil der Seite „Bondrucker“ bleibt unverändert. Neu sind allein die Karten der Zusatzdrucker darunter.

- Unter den Druckstationen steht eine gestrichelte Karte „Zusätzlicher Drucker“, wie heute die Karte nicht konfigurierter Stationen.
- Ein Satz erklärt den Zweck: „Für Produkte, die an einem anderen Ort zubereitet werden.“ Darunter steht der Knopf „Drucker hinzufügen“.
- „Drucker hinzufügen“ fragt nach dem Namen und legt den Zusatzdrucker an.
- Je Zusatzdrucker gibt es eine Karte wie bei den Druckstationen: Name, Drucker-IP, Bonmodus, Testbon, Löschen.
- Unter dem Bonmodus listet die Karte ihre Produkte, je mit Entfernen. „Produkt hinzufügen“ öffnet eine Auswahl aller nicht gelöschten Produkte.
- Die Auswahl zeigt bei jedem Produkt seine Kategorie und einen bestehenden Zusatzdrucker.
- Löschen fragt nach: „Die Produkte drucken danach wieder am Drucker ihrer Kategorie.“
- Name und Drucker-IP speichern beim Verlassen des Felds, der Bonmodus beim Antippen. So arbeitet die Seite heute.

Die Zuordnung steht auf der Druckerseite, nicht im Produktdialog. Die ganze Druckeinrichtung geschieht so vor Ort auf einer Seite. Der Produktdialog bleibt für jeden Verein gleich.

## Testing Decisions

Tests prüfen nur äußeres Verhalten. Erweitert werden die bestehenden Testdateien.

- Arbeitsbon-Policy: neue Testfälle in der bestehenden Testdatei, je Fall eine Funktion wie heute. Die Fälle: Zusatzdrucker schlägt Kategorie, Zusatzdrucker ohne IP, Sammelbon je Drucker, Direktverkauf mit und ohne Abholbon.
- Die bestehenden Policy-Tests laufen ohne Zuordnung und behalten ihre Erwartungen. Sie belegen, dass der gewöhnliche Fall gleich druckt.
- Domäne: ein Tabellentest wie der der Druckstation prüft Name, IP und Bonmodus des Zusatzdruckers.
- Repository: Integrationstest gegen echtes Postgres wie im Druckstationen-Repository. Er prüft die Auflösung je Variante, das Umhängen einer Zuordnung und ihr Entfernen beim Löschen.
- HTTP: keine. Die Feldabbildung deckt der Integrationstest ab, die Fehlercodes erzwingt der Vertragstest der Fehlerabbildung.
- Frontend: Der Seitentest prüft die gestrichelte Karte ohne Zusatzdrucker. Den Testbon-Namen in der Fehlerliste pinnt der bestehende Referenz-Test.
- E2E: Die bestehende Spezifikation „Admin verwaltet Druckstationen“ legt einen Zusatzdrucker an, ordnet ein Produkt zu und schickt einen Testbon.
- Migration: kein eigener Test. Sie legt nur leere Tabellen an, den Upgrade-Pfad deckt der CI-Job ab.
- Events: keine, abgedeckt durch den Vertragstest der Event-JSON-Formate.
- Produktstatistik und DSFinV-K: keine, abgedeckt durch ihre bestehenden Tests. Das Modell berührt beide nicht.

## Out of Scope

- Zuordnung je Variante. Eine Variante mit eigenem Ort wird ein eigenes Produkt.
- Zuordnung eines Produkts zum Drucker einer anderen Kategorie. Ein Zusatzdrucker mit derselben IP leistet das.
- Zusammenfassen der Bons zweier Drucker mit derselben IP.
- Ersatz der Kategoriedrucker durch frei benannte Stationen (Variante A).
- Löschen oder Umbau der Zeilen in der Tabelle der Druckstationen.
- Ein Hinweis auf den Zusatzdrucker in Produktliste oder Produktdialog.
- Eine Position an zwei Druckern, etwa als Kopie für die Theke.
- Mehrere Kassenbeleg- oder Abholbon-Drucker.
- Frei definierbare Kategorien. Es bleiben die drei festen Werte.
- Umleiten offener Druckaufträge, wenn ein Drucker ausfällt oder seine IP wechselt.
- Eine Auswertung je Drucker.
- Kategorien vergangener Kassensitzungen umschreiben. Die Positionen im Kassenjournal bleiben, wie sie sind.
- Kommentar je Position (#145). Dafür gibt es ein eigenes PRD.

## Further Notes

**Bauentscheidung.** Den Bedarf meldet ein Verein (#146). Die gebauten Punkte lösen den stummen Direktverkauf und den Piepton. Offen sind die drei Fälle oben.

- Für den Bau spricht: Die zweckentfremdete Kategorie verfälscht Produktstatistik und DSFinV-K-Warengruppe. Variante Z kostet den gewöhnlichen Verein nichts.
- Gegen den Bau spricht: Ein dritter Ausgabeort geht heute schon über „Sonstiges“, solange die Kategorie frei ist. Kein zweiter Verein hat ihn gemeldet.

**Angebot des Vereins.** Der meldende Verein bot in #146 an, die Umsetzung selbst zu schreiben. Sein Vorschlag entspricht Variante A.

**Auswirkung auf die Produktstatistik.** Sie gruppiert nach der Kategorie, die jede Position im Kassenjournal trägt. Variante Z ändert an ihr keinen Code.

Der Gewinn ist mittelbar: Ein Verein kann jedes Produkt in seine richtige Kategorie legen. Ab der nächsten Bestellung zählt es dort, vergangene Kassensitzungen bleiben unverändert.

Wer ein Produkt während einer laufenden Kassensitzung umsortiert, sieht es in zwei Kategorie-Abschnitten.

**Folgeänderungen bei Umsetzung.**

- Der Handbuch-Abschnitt zum Bondruck nennt die Auflösungsregel und den Zusatzdrucker im Direktverkauf-Routing.
- Das Handbuch rät, Kategorien vor dem Eröffnen der Kassensitzung zu ordnen.
- Die Anforderung K-25 nennt zusätzlich die Zuordnung einzelner Produkte zu Zusatzdruckern.
- Der Abschnitt „Bondruck & Infrastruktur“ der Sprachtabelle bekommt den Eintrag unten. Der Testbon-Eintrag nennt auch den Zusatzdrucker.
- Der Backlog-Eintrag zu #146 entfällt.

**Vorgeschlagener Eintrag für die Sprachtabelle.** Er kommt mit der Umsetzung, denn die Sprachtabelle führt nur gebaute Begriffe.

| Begriff       | Bedeutung                                                                                                                                                                                                                                 | Code-Mapping                                                                              |
| ------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- |
| Zusatzdrucker | Optionaler, benannter Bondrucker für die Arbeitsbons einzeln zugeordneter Produkte. Ein Produkt ohne Zuordnung druckt an der Druckstation seiner Kategorie. UI: „Zusätzlicher Drucker“. Nicht „Station“, „Extra-Drucker“ oder „Override“. | Druck-Modul · DB: Tabelle der Zusatzdrucker und Zuordnungstabelle Produkt → Zusatzdrucker |

**Offene Entscheidungen für den Owner:**

1. Bauen oder nicht. Empfohlen: Variante Z bauen. Sie löst alle drei Fälle und lässt den gewöhnlichen Fall unberührt. Die Alternative ist Variante 0.
2. Wer baut. Empfohlen: dem Verein in #146 Variante Z vorschlagen und sein Angebot annehmen. Die Review bleibt beim Owner.
