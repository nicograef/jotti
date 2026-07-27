# PRD: Storno-Zuordnung in der Abrechnung pro Servicekraft

## Problem Statement

Storniert ein Admin oder eine Serviceleitung stellvertretend für eine
Servicekraft, stimmt die Abrechnung dieser Servicekraft nicht: Am Ende des
Betriebstags gibt sie weniger Bargeld ab, als der Bericht ausweist, und die
Differenz sieht aus wie ein Fehlbetrag.

Der Befund im Code ist etwas anders, als die Meldung vermutet — **der Betrag
wird nicht beim Admin abgezogen, er wird überhaupt nirgends abgezogen**:

- **„Umsatz pro Servicekraft"** (Tagesbericht) und die Team-Liste im
  Live-Dashboard zeigen ausschließlich die Summe der `zahlung-kassiert`-Events
  je Akteur. Eine Warenrücknahme mindert diese Zahl bei **keiner** Servicekraft
  — auch dann nicht, wenn die Servicekraft selbst storniert hat. Die Zahl ist
  also kein Abrechnungs-Saldo, sondern eine Brutto-Kassiert-Summe.
- **Der „N Storno"-Marker** neben einer Servicekraft (und die aggregierte Zeile
  „felix 1 · sophie 1") gruppiert die Storno-Events über den **Akteur** des
  Storno-Events. Storniert die Serviceleitung stellvertretend, erscheint der
  Marker bei ihr, und die betroffene Servicekraft sieht unauffällig aus. Genau
  das ist die gemeldete Beobachtung.
- **Der Storno-Betrag pro Servicekraft** (`stornierungenCents`) wird vom Backend
  bereits geliefert, aber im Frontend nirgends angezeigt.

Für den Kassenwart heißt das: Es gibt heute keine Zahl, gegen die er das
Bargeld einer Servicekraft prüfen kann. Er muss die Storno-Detailliste von Hand
durchgehen und selbst zuordnen, wessen Kasse die Bar-Rückgabe belastet hat.

## Solution

Die Aufschlüsselung pro Servicekraft wird von einer Brutto-Kassiert-Summe zu
einer echten Bargeld-Abrechnung des Tischservice:

```
Kassiert   = Tischzahlungen
− Rücknahmen = Warenrücknahmen
= Abzugeben
```

Entscheidend ist, **wem** eine Rücknahme angerechnet wird. Jedes kassenwirksame
Storno-Event trägt bereits einen präzisen Rückverweis auf den Vorgang, dessen
Bargeld es zurückgibt — dieser Verweis, nicht der Akteur, bestimmt die
Zuordnung:

| Vorgang                                        | Verweis     | Zugeordnet auf                                        | Wirkung                     |
| ---------------------------------------------- | ----------- | ----------------------------------------------------- | --------------------------- |
| Warenrücknahme (bereits bezahlte Positionen)   | `zahlungId` | die Servicekraft, die **diese Zahlung kassiert** hat  | mindert „Abzugeben"         |
| Geldneutrale Korrektur (unbezahlte Positionen) | Positions-IDs | die Servicekraft, die **die Positionen bestellt** hat | nur Kontroll-Marker, kein Betrag |
| Direktverkauf-Storno                            | `verkaufId` | die Servicekraft, die **den Verkauf getätigt** hat    | nur Detailzeile, keine Aggregation |

Wer den Storno ausgelöst hat, ist für die Abrechnung damit tatsächlich
nebensächlich. Er bleibt in der Storno-Detailliste sichtbar (Kontroll-Signal
und Nachvollziehbarkeit), aber als Nebeninformation: Die Zeile nennt zuerst die
betroffene Servicekraft und danach — nur wenn abweichend — wer stellvertretend
storniert hat.

**Direktverkauf bleibt bewusst außerhalb dieser Abrechnung.** Er läuft über eine
eigene, vom Tischservice getrennte Kasse; ihn in dieselbe Zahl zu mischen würde
zwei Geldbeutel zu einem verrechnen. Die Servicekraft-Abrechnung bleibt daher
reine Tischservice-Abrechnung, so wie die Aggregation es heute schon abgrenzt.
Direktverkauf-Stornos bekommen in der Detailliste dieselbe Zuordnungsregel,
fließen aber in keine Servicekraft-Summe ein.

Für Admin und Kassenwart entsteht so eine Zahl, gegen die das abgegebene
Tischservice-Bargeld direkt geprüft werden kann. Die Servicekraft selbst
erfährt auf ihrem Dashboard von einer Rücknahme nur dann, wenn ihr eine
zugeordnet ist — dann aber mit Erklärung, statt sie bei der Abgabe zu
überraschen.

## User Stories

1. Als Kassenwart möchte ich im Tagesbericht pro Servicekraft sehen, wie viel
   Bargeld sie aus dem Tischservice abzugeben hat, damit ich die Abgabe ohne
   Nachrechnen prüfen kann.
2. Als Kassenwart möchte ich, dass eine Bar-Rückgabe die Kasse der Servicekraft
   belastet, die das Geld ursprünglich kassiert hat, damit ihr Saldo auch dann
   stimmt, wenn Admin oder Serviceleitung stellvertretend storniert haben.
3. Als Admin möchte ich in der Storno-Liste zuerst die betroffene Servicekraft
   sehen und den stellvertretenden Akteur nur als Zusatz, damit ich einen Storno
   sofort der richtigen Abrechnung zuordnen kann.
4. Als Admin möchte ich auch eine geldneutrale Korrektur bei der Servicekraft
   sehen, deren Bestellung korrigiert wurde, damit Häufungen bei einer Person
   auffallen — auch wenn kein Geld fließt.
5. Als Admin möchte ich, dass ein Direktverkauf-Storno den ursprünglichen
   Verkäufer nennt, damit die Storno-Liste durchgängig zeigt, wen ein Storno
   betrifft, statt wer ihn ausgelöst hat.
6. Als Serviceleitung möchte ich für eine Servicekraft stornieren können, ohne
   dass mein eigener Abrechnungs-Saldo dadurch verfälscht wird.
7. Als Servicekraft möchte ich auf meinem Dashboard erfahren, wenn für mich eine
   Rücknahme gebucht wurde — und was ich dadurch abzugeben habe —, damit ich
   beim Abrechnen nicht auf eine Differenz laufe, die ich mir nicht erklären
   kann.

## Implementation Decisions

### Fiskalische Ebene bleibt unberührt

Die Zuordnung ist **ausschließlich eine Reporting-Sicht**. Weder Events noch
Kassenjournal ändern sich:

- Kein neues Event, keine neue Event-Version, keine Änderung an
  Event-JSON-Contracts. Die Freeze-Disziplin bleibt vollständig gewahrt.
- `kassenjournal.user_id` / `user_name` bleiben der Akteur. DSFinV-K
  `BEDIENER_ID` / `BEDIENER_NAME` sind daran gebunden (erfassende Person nach
  DSFinV-K Tz. 3.2.1) und dürfen sich nicht ändern. Der Export wird nicht
  angefasst.
- Auch die Tisch-Historie und der Kassenbeleg zeigen weiterhin den Akteur.

Die Zuordnung wird zur Lesezeit aus den vorhandenen Verweisen abgeleitet —
dasselbe Muster wie das bestehende `barRueckgabe`-Feld, das ebenfalls aus dem
Event-Typ abgeleitet und nie gespeichert wird.

### Zuordnungs-Auflösung (Backend, Repository-Schicht)

Die Storno-Query löst je Storno-Art einen anderen Verweis auf, jeweils per
Self-Join innerhalb derselben Kassensitzung:

- **Warenrücknahme** → das `zahlung-kassiert:v1`-Event mit derselben
  `zahlungId`; dessen `user_id` / `user_name` ist die belastete Servicekraft.
  Einwertig, weil eine Zahlung genau einen Akteur hat.
- **Direktverkauf-Storno** → das `direktverkauf-getaetigt:v1`-Event mit
  derselben `verkaufId`. Ebenfalls einwertig.
- **Geldneutrale Korrektur** → für jede Positions-ID im Storno das
  `bestellung-aufgenommen:v1`-Event, dessen Positions-Array diese ID enthält;
  dessen Akteur ist der Besteller. **Mehrwertig:** Umfasst eine Korrektur
  Positionen aus Bestellungen verschiedener Servicekräfte, ist jede von ihnen
  betroffen.

`zahlungId` und `verkaufId` sind serverseitig erzeugte UUIDs, die Auflösung ist
deterministisch. Eine Positions-ID ist ebenfalls serverseitig erzeugt und stammt
immer aus genau einem `bestellung-aufgenommen`-Event. Lässt sich ein Verweis
nicht auflösen, fällt die Zeile auf den Akteur zurück, statt ohne Zuordnung zu
bleiben.

**Grenzfall Umbuchung (offene Entscheidung, hier bewusst nicht getroffen).** Die
ursprüngliche Annahme, eine Umbuchung verschiebe die Position ohne Änderung
ihrer ID, trifft nicht zu: `kasse.NewBestellungUmgebuchtEvents()` vergibt auf
dem Zieltisch **frische** Positions-IDs, die in keinem
`bestellung-aufgenommen`-Event vorkommen. Die Korrektur einer umgebuchten
Position findet über ihre Positions-ID daher keinen Besteller und fällt nach der
Regel oben auf den Akteur zurück. Zwei Auswege stehen offen — die Zuordnung auf
den **Umbucher** (konsistent zum restlichen System: `tagBesteller()` stempelt bei
der Umbuchung den Umbucher als Besteller auf die Zieltisch-Positionen, worauf
offene Arbeit, Erledigt-Sicht und „eigene zuerst" bereits aufbauen) oder eine
Abstammungs-Auflösung über die Umbuchungs-Event-Paare auf den ursprünglichen
Besteller. Beides ist eine Produktentscheidung und nicht Teil dieser Umsetzung.

Indizes auf `(data->>'zahlungId')` bzw. auf die Positions-IDs werden bewusst
**nicht** angelegt: Das Datenvolumen einer Kassensitzung liegt im niedrigen
vierstelligen Bereich, und die Reporting-Queries laufen on-demand für genau eine
Sitzung. Falls Messungen später etwas anderes zeigen, ist ein Index eine rein
additive Migration.

### Aufschlüsselung pro Servicekraft (Backend)

Die bestehende „Umsatz pro Servicekraft"-Aggregation wird zur
Abrechnungs-Aufschlüsselung. Pro Servicekraft werden geführt:

- **Kassiert:** Summe der Tischzahlungen, nach Akteur (unverändert).
- **Rücknahme:** Summe der Warenrücknahmen, nach Kassierer-Zuordnung.
- **Abzugeben:** Kassiert − Rücknahme.
- **Anzahl Stornos:** Kontroll-Marker über beide Tisch-Storno-Arten —
  Warenrücknahmen (Kassierer-Zuordnung) plus geldneutrale Korrekturen
  (Besteller-Zuordnung).

Direktverkäufe und Direktverkauf-Stornos gehen in keine dieser Zahlen ein.

**Invariante: „Abzugeben" ist nie negativ.** Eine Warenrücknahme kann nur
Positionen zurücknehmen, die in der referenzierten Zahlung enthalten und noch
nicht zurückgenommen sind (die FIFO-Aufteilung führt darüber Buch). Pro Zahlung
gilt daher Σ Rücknahmen ≤ Zahlbetrag, und weil beide Seiten demselben Kassierer
zugeordnet werden, gilt es auch pro Person. Das ist eine prüfbare Eigenschaft
der Zuordnungsregel — unter der bisherigen Akteurs-Zuordnung galt sie nicht.

**Bewusste Inkonsistenz beim Marker:** Betrifft eine Korrektur mehrere
Besteller, zählt sie bei jedem von ihnen. Die Summe aller Marker kann dadurch
größer sein als die Storno-Anzahl in der Kopfkennzahl. Das ist als
Kontroll-Signal richtig — beide Personen sind betroffen — darf im Frontend aber
nicht so dargestellt werden, als wäre der Marker eine Aufteilung der
Kopfkennzahl.

Eine Servicekraft erscheint in der Liste, sobald sie kassiert hat, einen Storno
zugeordnet bekommt oder (im Live-Dashboard) offene Arbeit hat. Die bestehende
Sortierung nach Umsatz absteigend gilt künftig für „Abzugeben". Die
Zusammenführung mit der offenen eigenen Arbeit im Live-Dashboard bleibt
unverändert.

### Storno-Detailzeile (Contract)

Die Storno-Detailzeile trennt die zwei Rollen sauber, statt eine einzelne
`user`-Angabe doppeldeutig zu belegen:

- **betroffene Servicekräfte** — die zugeordneten Personen. Bei Warenrücknahme
  und Direktverkauf-Storno genau eine, bei einer geldneutralen Korrektur eine
  oder mehrere.
- **Akteur** — wer den Storno ausgelöst hat; immer gesetzt.

Beides jeweils mit stabiler Benutzer-ID, eingefrorenem Username und live
aufgelöstem Klarnamen, konsistent zur heutigen Darstellung. HTTP-Response und
Zod-Schemas des Frontends ziehen mit; eine API-Versionierung ist nicht nötig,
weil Frontend und Backend gemeinsam ausgeliefert werden.

### Fachbegriffe

`docs/language.md` erhält zwei Ergänzungen, analog zum dort bereits definierten
**Besteller (bestellende Servicekraft)**, der für die Korrektur-Zuordnung
wiederverwendet wird:

- **Kassierer (kassierende Servicekraft):** die Servicekraft, die eine Zahlung
  kassiert hat — Reporting- und Anzeigekonzept, aus dem Event-Umschlag
  abgeleitet.
- **Storno-Zuordnung:** die Regel, dass ein Storno der Servicekraft zugeordnet
  wird, deren Vorgang er rückgängig macht (Kassierer bei der Warenrücknahme,
  Besteller bei der Korrektur, Verkäufer beim Direktverkauf-Storno) — nicht dem
  Akteur.

Die Reporting-Feldnamen (`UmsatzServicekraft`, `Breakdowns`) werden an die neue
Bedeutung angepasst und in `docs/language.md` nachgezogen.

### Darstellung (Frontend)

- **Tagesbericht, Abschnitt „Abrechnung pro Servicekraft"** (bisher „Umsatz pro
  Servicekraft"): pro Zeile „Abzugeben" als Hauptzahl, darunter Kassiert und
  Rücknahme als Nebenzeile, damit der Abzug nachvollziehbar bleibt. Die
  Abschnitts-Unterzeile weist aus, dass Direktverkäufe nicht enthalten sind. Der
  Storno-Marker bleibt als Kontroll-Signal, jetzt bei der betroffenen
  Servicekraft.
- **Live-Dashboard, Team-Liste:** dieselbe Hauptzahl; die Rücknahme wird nur
  eingeblendet, wenn sie ungleich null ist, damit die mobile Zeile schlank
  bleibt.
- **Storno-Detailzeile:** nennt die betroffene Servicekraft (bei mehreren:
  alle). Weicht der Akteur davon ab, folgt „storniert von <Akteur>" als
  gedämpfter Zusatz.
- **Eigene Übersicht der Servicekraft** (Service-Dashboard): Die zwei Kacheln
  (Bestellungen, Kassiert) bleiben unverändert — es ist die Betriebs-Seite
  während des Abends, keine Abrechnung, und Stornos sind selten. Ist dieser
  Servicekraft eine Rücknahme zugeordnet, erscheint darunter eine Hinweiszeile,
  die den Abzug erklärt und den abzugebenden Betrag nennt, damit sie bei der
  Abgabe nicht auf eine unerklärte Differenz läuft. Ohne Rücknahme ist die Seite
  unverändert. Geldneutrale Korrekturen bleiben hier außen vor: Sie ändern
  nichts an dem, was sie abzugeben hat.

### Was sich nicht ändert

Gesamtkennzahlen (Kassierter Umsatz, Storniert, Steuersatz-Aufschlüsselung,
Produktstatistik) bleiben unverändert: Sie aggregieren über alle Events und sind
von der Zuordnung je Person unabhängig. Die geldneutrale Korrektur zählt
weiterhin in „Storniert" und erscheint weiterhin in der Storno-Detailliste.

## Testing Decisions

Getestet wird ausschließlich beobachtbares Verhalten: Welche Beträge und welche
Servicekraft eine Query bzw. eine Ansicht ausgibt — nicht, über welchen Join
oder welche Zwischenfunktion das zustande kommt.

**Repository-Integrationstests** (Prior Art: `reporting_repo/repo_test.go`,
`summen_abschluss_test.go` — echte DB, Events schreiben, Query lesen). Kernfälle:

- Servicekraft kassiert, Serviceleitung nimmt stellvertretend Ware zurück →
  Rücknahme liegt bei der Servicekraft, die Serviceleitung bleibt unbelastet.
- Servicekraft nimmt selbst zurück → unverändert bei ihr; die Zuordnung ist
  keine Sonderregel für stellvertretende Stornos.
- Storno über mehrere Zahlungen verschiedener Kassierer (die FIFO-Aufteilung
  erzeugt je Zahlung ein Event) → jede Rücknahme landet bei ihrem eigenen
  Kassierer.
- Bezahlt Servicekraft A, was B bestellt hat, und wird zurückgenommen → Rücknahme
  bei A. Der Grenzfall, der die Regel „Kassierer, nicht Besteller" festschreibt.
- Geldneutrale Korrektur durch einen Dritten → Marker beim Besteller, kein
  Betrag; „Abzugeben" bleibt unverändert.
- Korrektur über Positionen zweier Besteller → Marker bei beiden.
- Korrektur einer umgebuchten Position → dokumentiert den Rückfall auf den
  Akteur, solange der Grenzfall Umbuchung offen ist (siehe „Grenzfall
  Umbuchung").
- Direktverkauf-Storno durch einen anderen Benutzer → Detailzeile nennt den
  ursprünglichen Verkäufer; keine Servicekraft-Summe verändert sich.
- „Abzugeben" ist nie negativ — inklusive des Falls, dass eine Zahlung
  vollständig zurückgenommen wird (Ergebnis: null, nicht negativ).

**Anwendungsschicht-Tests** (Prior Art: `application/query_test.go`,
`query_export_konsistenz_test.go`): Die Summe aller „Abzugeben" plus der offenen
Salden bleibt konsistent zum kassierten Tischservice-Umsatz der Sitzung — der
Test, der verhindert, dass Aufschlüsselung und Kopfkennzahl auseinanderlaufen.
Direktverkäufe sind auf beiden Seiten der Gleichung ausgenommen.

**Frontend-Tests** (Prior Art: `ReportingResults.test.tsx`,
`LiveReportingSection.test.tsx`, React Testing Library gegen gerenderten Text):
Hauptzahl und Nebenzeile pro Servicekraft, Storno-Marker bei der betroffenen
Person, Storno-Zeile mit und ohne abweichenden Akteur, Storno-Zeile mit mehreren
Betroffenen, sowie die eigene Übersicht mit und ohne zugeordnete Rücknahme
(ohne: unverändert zwei Kacheln, keine Hinweiszeile).

Der bestehende Event-JSON-Contract-Test bleibt unverändert grün — das ist der
Beleg, dass die Änderung rein lesend ist.

## Out of Scope

- **Änderungen an Events, Kassenjournal, TSE-Signatur oder DSFinV-K-Export.**
  Der Akteur bleibt dort überall der Erfasser.
- **Migration oder Umdeutung bestehender Daten.** Die Zuordnung wird zur
  Lesezeit berechnet und gilt rückwirkend auch für abgeschlossene
  Kassensitzungen, ohne dass eine Zeile angefasst wird.
- **Eine Abrechnung pro Direktverkäufer.** Der Direktverkauf hat eine eigene
  Kasse und verdient, wenn überhaupt, eine eigene Aufschlüsselung (Verkauft −
  Storniert) statt einer Vermischung mit dem Tischservice. Eigenes PRD, falls
  der Bedarf belegt ist.
- **Ein eigener Abrechnungs-Workflow** (Geld abgegeben, quittiert, Differenz je
  Person erfasst). Die Abgabe bleibt ein Vorgang außerhalb des Systems; der
  Bericht liefert nur die Sollzahl.
- **Zuordnung von Trinkgeld, Geldtransit oder Kassensturz-Differenz auf
  Personen.** Diese wirken auf die Kassensitzung, nicht auf eine Servicekraft.
- **Indizes auf `zahlungId` oder Positions-IDs.** Erst bei belegtem
  Performance-Bedarf.

## Further Notes

- **Anforderungen:** R-04 („Abrechnung pro Servicekraft") trägt den Titel
  bereits, liefert aber bisher nur den Brutto-Umsatz. Die Beschreibung in
  `docs/anforderungen.md` wird auf den Abrechnungs-Saldo nachgezogen; R-06
  („Eigene Übersicht") ebenfalls.
- **Handbuch:** Der Reporting-Abschnitt in `docs/handbuch.md` beschreibt die
  Endpunkt-Rückgaben; die Zuordnungsregel gehört dort als Read-Model-Regel
  hinterlegt.
- **Rückwirkung:** Weil die Zuordnung zur Lesezeit entsteht, zeigen auch bereits
  abgeschlossene Kassensitzungen sofort die korrigierte Aufschlüsselung. Ein
  gedruckter Altbericht kann daher von einem neu gedruckten abweichen — für
  Berichte, die keine fiskalische Ausfertigung sind, ist das unkritisch und der
  neue Stand ist der richtige.
- **Kein neuer Bedienschritt.** Die Änderung fügt weder Feld noch Klick noch
  Status hinzu; sie korrigiert nur, in welcher Zeile eine bereits erfasste Zahl
  landet.
