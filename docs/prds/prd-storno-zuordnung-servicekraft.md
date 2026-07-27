# PRD: Storno-Zuordnung in der Abrechnung pro Servicekraft

> Die vier Klärungsfragen zu diesem PRD wurden nicht beantwortet. Das Dokument
> entscheidet daher nach den im Chat empfohlenen Varianten; jede so getroffene
> Entscheidung ist unter [Annahmen](#annahmen) als **Annahme** markiert und kann
> ohne Umbau des Rests umgestoßen werden.

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
- **Direktverkäufe** fehlen in der Servicekraft-Aufschlüsselung vollständig,
  obwohl das Bargeld in derselben Tasche landet. Ihre Stornos tauchen dagegen in
  der Storno-Liste auf.

Für den Kassenwart heißt das: Es gibt heute keine Zahl, gegen die er das
Bargeld einer Servicekraft prüfen kann. Er muss die Storno-Detailliste von Hand
durchgehen und selbst zuordnen, wessen Kasse die Bar-Rückgabe belastet hat.

## Solution

Die Aufschlüsselung pro Servicekraft wird von einer Brutto-Kassiert-Summe zu
einer echten Bargeld-Abrechnung:

```
Kassiert   = Tischzahlungen + Direktverkäufe
− Rückgabe = Warenrücknahmen + Direktverkauf-Stornos
= Abzugeben
```

Entscheidend ist, **wem** eine Rückgabe angerechnet wird. Jedes kassenwirksame
Storno-Event trägt bereits einen präzisen Rückverweis auf den Vorgang, dessen
Bargeld es zurückgibt — dieser Verweis, nicht der Akteur, bestimmt die
Zuordnung:

| Vorgang                                        | Verweis     | Angerechnet auf                            |
| ---------------------------------------------- | ----------- | ------------------------------------------ |
| Warenrücknahme (bereits bezahlte Positionen)   | `zahlungId` | die Servicekraft, die diese Zahlung kassiert hat |
| Direktverkauf-Storno                            | `verkaufId` | die Servicekraft, die diesen Direktverkauf getätigt hat |
| Geldneutrale Korrektur (unbezahlte Positionen) | —           | niemanden — es fließt kein Bargeld         |

Wer den Storno ausgelöst hat, ist für die Abrechnung damit tatsächlich
nebensächlich. Er bleibt in der Storno-Detailliste sichtbar (Kontroll-Signal
und Nachvollziehbarkeit), aber als Nebeninformation: Die Zeile nennt zuerst die
Servicekraft, deren Kasse betroffen ist, und danach — nur wenn abweichend — wer
stellvertretend storniert hat.

Für Admin und Kassenwart entsteht so eine Zahl, gegen die das abgegebene
Bargeld direkt geprüft werden kann; für die Servicekraft zeigt die eigene
Übersicht denselben Betrag, den sie abzugeben hat.

## User Stories

1. Als Kassenwart möchte ich im Tagesbericht pro Servicekraft sehen, wie viel
   Bargeld sie abzugeben hat, damit ich die Abgabe ohne Nachrechnen prüfen kann.
2. Als Kassenwart möchte ich, dass eine Bar-Rückgabe die Kasse der Servicekraft
   belastet, die das Geld ursprünglich kassiert hat, damit ihr Saldo auch dann
   stimmt, wenn Admin oder Serviceleitung stellvertretend storniert haben.
3. Als Kassenwart möchte ich, dass Direktverkäufe und deren Stornos in derselben
   Servicekraft-Zahl stecken wie die Tischzahlungen, damit die Zahl das gesamte
   Bargeld dieser Person abbildet.
4. Als Admin möchte ich in der Storno-Liste zuerst die betroffene Servicekraft
   sehen und den stellvertretenden Akteur nur als Zusatz, damit ich einen Storno
   sofort der richtigen Abrechnung zuordnen kann.
5. Als Admin möchte ich Stornos weiterhin als Kontroll-Signal je Servicekraft
   erkennen, damit auffällige Häufungen sichtbar bleiben.
6. Als Serviceleitung möchte ich für eine Servicekraft stornieren können, ohne
   dass mein eigener Abrechnungs-Saldo dadurch verfälscht wird.
7. Als Servicekraft möchte ich auf meinem Dashboard den Betrag sehen, den ich
   tatsächlich abzugeben habe, damit ich beim Abrechnen nicht auf eine
   Differenz laufe, die ich mir nicht erklären kann.

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

Die Zuordnung wird zur Lesezeit aus den vorhandenen Event-Verweisen abgeleitet —
dasselbe Muster wie das bereits bestehende `barRueckgabe`-Feld, das ebenfalls
aus dem Event-Typ abgeleitet und nie gespeichert wird.

### Zuordnungs-Auflösung (Backend, Repository-Schicht)

Die Storno-Query löst den Verweis per Self-Join auf das Kassenjournal auf:

- `stornierung-erteilt:v1` → das `zahlung-kassiert:v1`-Event mit derselben
  `zahlungId`; dessen `user_id` / `user_name` ist die belastete Servicekraft.
- `direktverkauf-storniert:v1` → das `direktverkauf-getaetigt:v1`-Event mit
  derselben `verkaufId`.
- `bestellung-korrigiert:v1` → kein Verweis, keine Zuordnung; die Zeile trägt
  nur den Akteur.

Beide Verweise sind serverseitig erzeugte UUIDs und je Ziel-Event eindeutig,
die Auflösung ist damit deterministisch und einwertig. Der Join bleibt innerhalb
derselben Kassensitzung.

Ein Index auf `(data->>'zahlungId')` für `zahlung-kassiert:v1` wird bewusst
**nicht** angelegt: Das Datenvolumen einer Kassensitzung liegt im niedrigen
vierstelligen Bereich, und die Reporting-Queries laufen on-demand für genau eine
Sitzung. Falls Messungen später etwas anderes zeigen, ist der Index eine rein
additive Migration.

### Aufschlüsselung pro Servicekraft (Backend)

Die bestehende „Umsatz pro Servicekraft"-Aggregation wird zur
Abrechnungs-Aufschlüsselung erweitert. Pro Servicekraft werden geführt:

- **Kassiert:** Summe der Tischzahlungen **und** Direktverkäufe, je nach Akteur
  des jeweiligen Events (Direktverkauf ist neu enthalten).
- **Rückgabe:** Summe der Warenrücknahmen und Direktverkauf-Stornos, zugeordnet
  nach der obigen Verweis-Auflösung.
- **Abzugeben:** Kassiert − Rückgabe. Kann rechnerisch negativ werden (eine
  Servicekraft nimmt Ware zurück, die eine andere kassiert hat, ohne selbst zu
  kassieren); das wird nicht abgeschnitten, sondern als negativer Betrag
  ausgewiesen — er ist die Auszahlung, die diese Person aus der Kasse erhalten
  hat.
- **Anzahl Stornos:** Kontroll-Signal, gezählt über dieselbe Zuordnung.

Eine Servicekraft erscheint in der Liste, sobald sie kassiert hat, eine
Rückgabe zugeordnet bekommt oder (im Live-Dashboard) offene Arbeit hat. Die
bestehende Sortierung nach Umsatz absteigend gilt künftig für „Abzugeben".

Die Zusammenführung mit der offenen eigenen Arbeit im Live-Dashboard bleibt
unverändert bestehen.

### Storno-Detailzeile (Contract)

Die Storno-Detailzeile trennt die zwei Rollen sauber, statt eine einzelne
`user`-Angabe doppeldeutig zu belegen:

- **betroffene Servicekraft** — die belastete Person; leer bei geldneutralen
  Korrekturen.
- **Akteur** — wer den Storno ausgelöst hat; immer gesetzt.

Beides jeweils mit stabiler Benutzer-ID, eingefrorenem Username und live
aufgelöstem Klarnamen, konsistent zur heutigen Darstellung. Die HTTP-Response
und die Zod-Schemas des Frontends ziehen mit; eine API-Versionierung ist nicht
nötig, weil Frontend und Backend gemeinsam ausgeliefert werden.

### Fachbegriffe

`docs/language.md` erhält zwei Ergänzungen, analog zum dort bereits definierten
**Besteller (bestellende Servicekraft)**:

- **Kassierer (kassierende Servicekraft):** die Servicekraft, die eine Zahlung
  kassiert oder einen Direktverkauf getätigt hat — Reporting- und
  Anzeigekonzept, aus dem Event-Umschlag abgeleitet.
- **Storno-Zuordnung:** die Regel, dass ein kassenwirksamer Storno dem
  Kassierer des referenzierten Vorgangs angerechnet wird, nicht dem Akteur.

Die Reporting-Feldnamen (`UmsatzServicekraft`, `Breakdowns`) werden an die neue
Bedeutung angepasst und in `docs/language.md` nachgezogen.

### Darstellung (Frontend)

- **Tagesbericht, Abschnitt „Abrechnung pro Servicekraft"** (bisher „Umsatz pro
  Servicekraft"): pro Zeile „Abzugeben" als Hauptzahl, darunter Kassiert und
  Rückgabe als Nebenzeile, damit der Abzug nachvollziehbar bleibt. Der
  Storno-Marker bleibt als Kontroll-Signal, jetzt bei der betroffenen
  Servicekraft.
- **Live-Dashboard, Team-Liste:** dieselbe Hauptzahl; die Rückgabe wird nur
  eingeblendet, wenn sie ungleich null ist, damit die mobile Zeile schlank
  bleibt.
- **Storno-Detailzeile:** zeigt die betroffene Servicekraft; weicht der Akteur
  ab, folgt „storniert von <Akteur>" als gedämpfter Zusatz. Bei geldneutralen
  Korrekturen steht wie bisher nur der Akteur.
- **Eigene Übersicht der Servicekraft** (Service-Dashboard): die Kachel
  „Kassiert" wird um denselben Rückgabe-Abzug ergänzt und zeigt „Abzugeben", auf
  derselben Zuordnungsregel. Ohne das würden Servicekraft und Kassenwart auf
  zwei verschiedene Zahlen schauen. Sichtbar erweiterter Scope gegenüber der
  ursprünglichen Meldung — bewusst enthalten, siehe [Annahmen](#annahmen).

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

- Servicekraft kassiert, Serviceleitung storniert stellvertretend → Rückgabe
  liegt bei der Servicekraft, die Serviceleitung bleibt unbelastet.
- Servicekraft storniert selbst → unverändert bei ihr; die Zuordnung ist keine
  Sonderregel für stellvertretende Stornos.
- Storno über mehrere Zahlungen verschiedener Kassierer (FIFO-Aufteilung erzeugt
  je Zahlung ein Event) → jede Rückgabe landet bei ihrem eigenen Kassierer.
- Direktverkauf-Storno durch einen anderen Benutzer → Rückgabe beim
  ursprünglichen Verkäufer; Direktverkaufs-Umsatz und -Storno derselben Person
  heben sich auf.
- Geldneutrale Korrektur → verändert keine Servicekraft-Summe, erscheint aber in
  der Detailliste und in den Gesamtkennzahlen.
- Bezahlt Servicekraft A, was B bestellt hat, und wird storniert → Rückgabe bei
  A. Grenzfall, der die gewählte Regel (Kassierer, nicht Besteller) festschreibt.
- Reine Rückgabe ohne eigenes Kassieren → negatives „Abzugeben", nicht
  abgeschnitten.

**Anwendungsschicht-Tests** (Prior Art: `application/query_test.go`,
`query_export_konsistenz_test.go`): Die Summe aller „Abzugeben" plus der offenen
Salden bleibt konsistent zum kassierten Gesamtumsatz der Sitzung — der Test, der
verhindert, dass die Aufschlüsselung und die Kopfkennzahl auseinanderlaufen.

**Frontend-Tests** (Prior Art: `ReportingResults.test.tsx`,
`LiveReportingSection.test.tsx`, React Testing Library gegen gerenderten Text):
Hauptzahl und Nebenzeile pro Servicekraft, Storno-Marker bei der betroffenen
Person, Storno-Zeile mit und ohne abweichenden Akteur, „Abzugeben" in der
eigenen Übersicht.

Der bestehende Event-JSON-Contract-Test bleibt unverändert grün — das ist der
Beleg, dass die Änderung rein lesend ist.

## Out of Scope

- **Änderungen an Events, Kassenjournal, TSE-Signatur oder DSFinV-K-Export.**
  Der Akteur bleibt dort überall der Erfasser.
- **Migration oder Umdeutung bestehender Daten.** Die Zuordnung wird zur
  Lesezeit berechnet und gilt rückwirkend auch für abgeschlossene
  Kassensitzungen, ohne dass eine Zeile angefasst wird.
- **Zuordnung geldneutraler Korrekturen auf den Besteller.** Erfordert eine
  Auflösung Position → Bestellung über JSONB und ist mehrdeutig, sobald eine
  Korrektur Positionen aus Bestellungen verschiedener Servicekräfte umfasst —
  für einen Vorgang, der kein Bargeld bewegt.
- **Ein eigener Abrechnungs-Workflow** (Geld abgegeben, quittiert, Differenz je
  Person erfasst). Die Abgabe bleibt ein Vorgang außerhalb des Systems; der
  Bericht liefert nur die Sollzahl.
- **Zuordnung von Trinkgeld, Geldtransit oder Kassensturz-Differenz auf
  Personen.** Diese wirken auf die Kassensitzung, nicht auf eine Servicekraft.
- **Index auf `zahlungId`.** Erst bei belegtem Performance-Bedarf.

## Annahmen

Die folgenden Punkte wurden mangels Antwort nach der jeweils empfohlenen
Variante entschieden.

> **Annahme 1 — Zuordnung über den Kassierer, nicht den Besteller.** Die Meldung
> nennt den „Bestell-User". Für den Bargeld-Saldo ist das falsch: Zurückgegeben
> wird Geld, das der **Kassierer** eingenommen hat, und genau darauf zeigt die
> `zahlungId` im Storno-Event. Fallen Besteller und Kassierer auseinander,
> stimmt nur die Kassierer-Zuordnung — und nur sie ist eindeutig, wenn ein
> Storno Positionen aus mehreren Bestellungen umfasst. In der Praxis ist es
> meist dieselbe Person. Sollte stattdessen wirklich der Besteller gemeint sein,
> ändert sich die Auflösungsregel; die übrige Struktur bleibt.

> **Annahme 2 — „Abzugeben" wird die Hauptzahl.** Kassiert und Rückgabe bleiben
> als Nebenzeile sichtbar. Nur mit einer Netto-Zahl „stimmt die Abrechnung"
> ohne Nachrechnen.

> **Annahme 3 — Direktverkäufe kommen in die Servicekraft-Zahl.** Sie sind heute
> bewusst ausgeklammert („Tischservice-Umsatz"). Für einen Bargeld-Saldo müssen
> sie hinein, sonst zieht die Zahl einen Direktverkauf-Storno ab, dessen Umsatz
> sie nie enthalten hat. Der Abschnittstitel wechselt deshalb von „Umsatz" zu
> „Abrechnung pro Servicekraft".

> **Annahme 4 — geldneutrale Korrekturen bleiben aus der
> Servicekraft-Aufschlüsselung heraus.** Hier weicht das PRD von der im Chat
> empfohlenen Variante ab: Bei genauerer Betrachtung erfordert eine Zuordnung
> auf den Besteller eine JSONB-Auflösung Position → Bestellung und ist
> mehrdeutig, sobald mehrere Besteller betroffen sind — für einen Vorgang, der
> kein Bargeld bewegt. Der Verzicht macht die Servicekraft-Zahlen eindeutig
> kassenwirksam. Preis: Häufen sich Korrekturen bei einer Person, fällt das nur
> noch in der Storno-Detailliste auf, nicht mehr am Marker.

> **Annahme 5 — die eigene Übersicht der Servicekraft zieht mit.** Ohne das
> zeigen Service-Dashboard und Admin-Bericht zwei verschiedene Zahlen für
> dieselbe Abgabe. Ist der Scope zu weit, lässt sich dieser Punkt isoliert
> streichen.

## Further Notes

- **Anforderungen:** R-04 („Abrechnung pro Servicekraft") trägt den Titel
  bereits, liefert aber bisher nur den Brutto-Umsatz. Die Beschreibung in
  `docs/anforderungen.md` wird auf den Abrechnungs-Saldo nachgezogen; R-06
  („Eigene Übersicht") ebenfalls.
- **Handbuch:** Der Reporting-Abschnitt in `docs/handbuch.md` beschreibt die
  Endpunkt-Rückgaben; die neue Zuordnungsregel gehört dort als Read-Model-Regel
  hinterlegt.
- **Rückwirkung:** Weil die Zuordnung zur Lesezeit entsteht, zeigen auch bereits
  abgeschlossene Kassensitzungen sofort die korrigierte Aufschlüsselung. Ein
  gedruckter Altbericht kann daher von einem neu gedruckten abweichen — für
  Berichte, die keine fiskalische Ausfertigung sind, ist das unkritisch und der
  neue Stand ist der richtige.
- **Kein neuer Bedienschritt.** Die Änderung fügt weder Feld noch Klick noch
  Status hinzu; sie korrigiert nur, in welcher Zeile eine bereits erfasste Zahl
  landet.
