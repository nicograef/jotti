# PRD: Geldwirksames Storno und Wegfall der Auszahlung

> Ersetzt das aktuelle Storno-/Auszahlungs-Modell am Tisch. Hintergrund:
> Die `auszahlung-geleistet` existiert heute nur, um einen negativen
> Tisch-Saldo auszugleichen, der entsteht, wenn bereits bezahlte Positionen
> nachträglich storniert werden. Sie ist ein Saldo-Pflaster ohne fachliche
> Eigenständigkeit, fiskalisch als USt-neutrale Auszahlung (DSFinV-K GV_TYP
> `Auszahlung`) fehlklassifiziert, und entkoppelt vom auslösenden Storno.
> Der Direktverkauf macht es bereits richtig: sein Storno ist sofort
> kassenwirksam, ohne separate Auszahlung. Diese PRD bringt den Tisch-Pfad
> auf dasselbe, fiskalisch korrekte Modell.

## Problem Statement

Storniere ich am Tisch eine Position, die der Gast schon bezahlt hat, muss
ich das Geld zurückgeben. Heute passiert das in zwei entkoppelten Schritten:
Die Stornierung dreht den Saldo ins Minus, danach buche ich eine separate
„Auszahlung" mit frei eingetragenem Betrag und Pflichtkommentar, die den Saldo
wieder auf null zieht. Das ist umständlich, fehleranfällig (der Betrag wird
von Hand getippt, statt aus dem Storno zu folgen) und fachlich verwirrend:
„Storno" und „Auszahlung" sind für Helfer, Küche und Rechner ein einziger
Vorgang, nämlich die Rückgabe.

Fiskalisch ist es falsch. Eine Rückerstattung bezahlter Speisen ist eine
Warenrücknahme: ein negativer Umsatz am ursprünglichen Steuersatz mit Bezug
auf den Ursprungsbeleg (DSFinV-K Tz. 4.2.5 „Vorgänge mit Negativpositionen“; die
Referenz auf den Ursprungsbeleg regelt Tz. 4.2.2). jotti bucht
sie stattdessen als USt-neutrale Auszahlung ohne Steuersatz und ohne Referenz.
Zusätzlich widersprechen sich die beiden fiskalischen Darstellungen desselben
Stornos: Die TSE-signierten `processData` enthalten eine `-X:Bar`-Zeile
(Bar-Rückgabe), der DSFinV-K-Export behandelt das Storno aber als geldneutral
und bucht die Rückgabe erst über die Auszahlung. Im TSE-Journal wird die
Bar-Rückgabe damit doppelt geführt, im Export einfach. Das untergräbt die
Kassensturzfähigkeit, die der AEAO zu § 146a (Nr. 2.2.3.6.1) gerade über diese
Daten sichern will.

Drittens ist der Tisch-Saldo ein einziger Netto-Skalar, der Bestelltes,
Gezahltes, Storniertes und Ausgezahltes vermischt. Sitzen mehrere Gästegruppen
an einem Tisch, kann eine Gruppe eine Rückgabe brauchen, während eine andere
noch offen ist; die Rückgabe verrechnet sich dann gegen die offene Rechnung
der anderen Gruppe, was fachlich und fiskalisch unzusammenhängende Vorgänge
vermengt.

Schließlich verschiebt die Umbuchung (z. B. versehentlich falscher Tisch
gewählt) Positionen heute über eine interne Stornierung; in der Historie
erscheint dann eine „Stornierung" für etwas, das eigentlich ein Verschieben
ist.

## Solution

Storno wird zu einem einzigen, in sich abgeschlossenen Vorgang. Als
Servicekraft wähle ich die zu stornierenden Positionen und löse „Stornieren"
aus. Das System erkennt selbst, ob die Positionen bereits bezahlt sind:

- **Unbezahlte Positionen** werden geldneutral storniert (reine
  Auftragskorrektur, kein Geld, kein Umsatz).
- **Bezahlte Positionen** werden als Warenrücknahme storniert: der Umsatz wird
  am ursprünglichen Steuersatz negativ gebucht, die Bar-Rückgabe ist Teil
  desselben Belegs, und der Beleg referenziert die ursprüngliche Zahlung. Es
  gibt keine separate Auszahlung mehr.

In der Oberfläche heißt das durchgehend „Storno"/„Stornieren"; die fiskalische
Unterscheidung passiert unsichtbar im Hintergrund. Den kassenwirksamen Storno
(Bar-Rückgabe) löst nur die Serviceleitung aus. Auf Wunsch wird ein
Stornobeleg gedruckt, wie beim regulären Kassenbeleg und beim
Direktverkauf-Storno.

Der Tisch-Saldo bedeutet künftig nur noch „offener Betrag" und ist nie
negativ: Eine Warenrücknahme ist ein abgeschlossener Beleg, kein
Negativ-Zustand, der geplombt werden müsste. Die Umbuchung wird ein eigener,
geldneutraler Vorgang, der in Historie und Export als „Umbuchung" erscheint,
nicht als „Storno".

## User Stories

**Servicekraft (Helfer)**

1. Als Servicekraft möchte ich eine noch nicht bezahlte Position stornieren, damit ein versehentlich aufgenommener Artikel wieder von der offenen Rechnung des Tisches verschwindet.
2. Als Servicekraft möchte ich beim Stornieren denselben Begriff „Storno" sehen, den wir im Team mündlich verwenden, damit die App meiner Arbeitsweise entspricht.
3. Als Servicekraft möchte ich, dass der offene Betrag eines Tisches nie negativ wird, damit ich immer sofort sehe, was der Tisch noch schuldet.
4. Als Servicekraft möchte ich eine Bestellung auf einen anderen Tisch umbuchen, wenn ich den falschen Tisch gewählt habe, ohne dass das wie eine Stornierung aussieht.
5. Als Servicekraft möchte ich bei einer Umbuchung nur noch nicht bezahlte Positionen verschieben können, damit bereits abgerechnete Positionen unangetastet bleiben.
6. Als Servicekraft möchte ich, dass ein versehentlich falsch storniertes Verschieben in der Historie korrekt als „Umbuchung" erscheint, damit die Serviceleitung den Vorgang richtig einordnet.

**Serviceleitung**

7. Als Serviceleitung möchte ich eine bereits bezahlte Position stornieren und dem Gast bar zurückgeben, in einem Schritt, ohne danach noch eine Auszahlung tippen zu müssen.
8. Als Serviceleitung möchte ich, dass der Rückgabebetrag automatisch aus den stornierten, bezahlten Positionen folgt, damit ich mich nicht vertippen kann.
9. Als Serviceleitung möchte ich, dass nur ich (oder ein Admin) eine kassenwirksame Bar-Rückgabe auslösen kann, damit Geldabflüsse rollengeschützt bleiben.
10. Als Serviceleitung möchte ich einen Storno stornieren können, der teils bezahlte und teils unbezahlte Positionen umfasst, ohne die beiden Fälle selbst trennen zu müssen.
11. Als Serviceleitung möchte ich bei jedem kassenwirksamen Storno einen Pflichtkommentar hinterlegen, damit der Grund der Rückgabe für die Betriebsprüfung dokumentiert ist.
12. Als Serviceleitung möchte ich auf Anforderung einen Stornobeleg drucken, damit der Gast einen Beleg über die Rückgabe erhält und die Belegausgabepflicht erfüllt ist.
13. Als Serviceleitung möchte ich nicht mehr nach einer separaten „Auszahlung" suchen müssen, weil die Rückgabe Teil des Stornos ist.

**Küche / Theke**

14. Als Küchenkraft möchte ich, dass eine geldneutrale Korrektur einer noch nicht ausgegebenen Position die zugehörige Arbeit von der Ausstehend-Liste nimmt, damit ich nichts zubereite, was storniert wurde.

**Gast**

15. Als Gast möchte ich bei einer Rückgabe bar mein Geld zurückbekommen und auf Wunsch einen Stornobeleg erhalten.
16. Als Gast einer von mehreren Gruppen am selben Tisch möchte ich, dass meine Rückgabe nicht mit der offenen Rechnung einer anderen Gruppe verrechnet wird.

**Rechner / Buchhaltung / Betriebsprüfer**

17. Als Buchhalter möchte ich, dass eine Rückerstattung bezahlter Ware als negativer Umsatz am ursprünglichen Steuersatz erscheint, nicht als USt-neutrale Auszahlung, damit die Umsatzsteuer korrekt gemindert wird.
18. Als Betriebsprüfer möchte ich, dass jeder Storno-Beleg über `REF_BON_ID` auf die ursprüngliche Zahlung verweist, damit der Vorgang lückenlos nachvollziehbar ist.
19. Als Betriebsprüfer möchte ich, dass die TSE-signierten `processData` und der DSFinV-K-Export dieselbe Bar-Bewegung zeigen, damit die Kassensturzfähigkeit gewahrt ist und nichts doppelt gebucht wird.
20. Als Buchhalter möchte ich, dass eine geldneutrale Auftragskorrektur (Storno unbezahlter Positionen) als geldneutraler Vorgang ohne Umsatz und ohne Bar-Bewegung gebucht wird.
21. Als Buchhalter möchte ich, dass eine Umbuchung als geldneutraler Vorgang erkennbar ist und nicht als Umsatz oder Bar-Bewegung erscheint.
22. Als Betriebsprüfer möchte ich, dass jeder kassenwirksame Storno als `Kassenbeleg-V1` über die TSE abgesichert ist (AEAO Nr. 2.2.3.6.1), inklusive Stornobeleg.

**Admin / Reporting**

23. Als Admin möchte ich im Reporting Stornoquote und Stornierungsbeträge sehen, ohne eine separate Auszahlungs-Kennzahl, weil es keine Auszahlungen mehr gibt.
24. Als Admin möchte ich, dass der Tagesabschluss (Z-Bon) die Warenrücknahmen als negative Umsätze und nicht als Auszahlungen ausweist.
25. Als Admin möchte ich, dass der Kassenbestand am Tagesende stimmt, weil kassenwirksame Stornos den Bargeldbestand mindern und geldneutrale Vorgänge ihn nicht berühren.

**Entwickler**

26. Als Entwickler möchte ich, dass „eine Storno-Aktion in der UI" auf genau eine oder mehrere klar getypte Domain-Events abgebildet wird, jedes mit genau einer TSE-Transaktion, damit die Invariante 1 Event = 1 TSE-Transaktion erhalten bleibt.
27. Als Entwickler möchte ich, dass der Tisch-Saldo ein abgeleiteter, nie negativer Wert ist (Summe unbezahlter Positionen), damit kein Pflaster-Mechanismus nötig ist.
28. Als Entwickler möchte ich keinen Event-Typ `auszahlung-geleistet` mehr in Domain, Projektion, Mapper, Reporting und Frontend pflegen müssen.
29. Als Entwickler möchte ich, dass der Direktverkauf-Storno und der Tisch-Storno demselben Modell folgen, damit die beiden Pfade konsistent sind.

## Implementation Decisions

**Event-Modell (Domain `kasse`).** Eine „Storno"-Aktion in der UI bildet auf
ein oder mehrere Domain-Events ab; das Command-Routing wählt nach Bezahlstatus
der referenzierten Positionen. Bezahlt und unbezahlt gemischte Anforderungen
werden serverseitig in je ein Event pro Teil aufgeteilt; betrifft der bezahlte
Teil mehrere Teilzahlungen, entsteht ein `stornierung-erteilt`-Event je
referenzierter Zahlung (Zuordnung der stornierten Mengen FIFO, älteste
begleichende Zahlung zuerst). Es gilt durchgehend eine TSE-Transaktion pro Event
und ein Storno-Beleg je referenzierter Zahlung:

- `stornierung-erteilt:v1` wird neu definiert als **kassenwirksame
  Warenrücknahme bereits bezahlter Positionen**. Trägt die Positionen (mit
  Steuersatz), den Gesamtbetrag, eine Referenz auf genau eine ursprüngliche
  Zahlung und einen Pflichtkommentar. Wird als `Kassenbeleg-V1` mit negativem
  Bruttoumsatz je Steuersatz und einer negativen Bar-Zahlung signiert.
- `bestellung-korrigiert:v1` (neu) ist die **geldneutrale Stornierung
  unbezahlter Positionen**. Trägt die Positionen und einen Kommentar, wird als
  `Bestellung-V1` ohne Zahlungszeile und ohne Umsatz signiert.
- `bestellung-umgebucht:v1` (neu) ist die **eigenständige, geldneutrale
  Umbuchung** unbezahlter Positionen zwischen zwei Tischen. Ersetzt die
  bisherige Modellierung als Stornierung + Bestellung. Quell- und Zielstrom
  erhalten verknüpfte Einträge; beide sind geldneutral.
- `auszahlung-geleistet:v1` entfällt vollständig, samt zugehörigem
  `Auszahlung`-Struct, Historien-Eintrag, Command, Reporting-Kennzahl und UI.

**Benennung / Ubiquitous Language.** Der stärkste Begriff `stornierung-erteilt`
bleibt bei der Geld-raus-Aktion, die Helfer und Rechner emphatisch „Storno"
nennen. Die geldneutralen Vorgänge heißen `bestellung-korrigiert` und
`bestellung-umgebucht`. In der Oberfläche erscheint für die Stornofälle
durchgehend „Storno"/„Stornieren"; die fiskalische Aufteilung ist nicht
nutzersichtbar. Die Umbuchung erscheint in UI, Historie und Export als
„Umbuchung". `docs/language.md` wird entsprechend angepasst (Auszahlung
gestrichen, Stornierung neu gefasst, Umbuchung ergänzt).

**Tisch-Saldo (Projektion `TischSession`).** `SaldoCents` bedeutet künftig den
offenen Betrag, abgeleitet als Summe der unbezahlten Positionen, und ist nie
negativ. Eine kassenwirksame Warenrücknahme entfernt die betroffenen
Positionen aus den aktiven Listen und erzeugt einen abgeschlossenen
Rückgabe-Beleg; sie verändert den offenen Betrag nicht und kann ihn nicht ins
Minus drehen. Die geldneutrale Korrektur reduziert den offenen Betrag und die
Ausstehend-Liste.

**Steuer und Referenz.** Da Positionen ihren Steuersatz tragen, folgt die
negative Umsatzaufteilung der Warenrücknahme direkt aus den stornierten
Positionen (Regel-, ermäßigter, befreiter und Kombi-Satz wie beim Verkauf).
Die `processData` der geldneutralen Vorgänge enthalten keine `:Bar`-Zeile
(behebt die bisherige Doppelbuchung). Der DSFinV-K-Export setzt für den
kassenwirksamen Storno `REF_BON_ID` auf die **Zahlung** (nicht die Bestellung).
Weil Zahlungen mengenweise erfolgen, ordnet das Command die stornierten Mengen
ihren begleichenden Zahlungen zu (FIFO) und legt je betroffener Zahlung ein
eigenes `stornierung-erteilt`-Event mit genau einer Zahlungsreferenz an; der
Mapper übernimmt diese Referenz unverändert. So verweist jeder Storno-Beleg auf
genau einen Ursprungs-Zahlungsbon.

**Belegausgabe.** Der kassenwirksame Tisch-Storno erzeugt auf Anforderung einen
Stornobeleg (negativer Betrag, Referenz auf den Ursprungsbeleg), über denselben
Druck-Endpunkt wie der reguläre Kassenbeleg, analog zum bestehenden
Direktverkauf-Storno-Beleg.

**Eigenbeleg-`processData` (Feld 5).** Der gemeinsame Eigenbeleg-Builder
(Geldtransit, Kassendifferenz; die Auszahlung entfällt) schreibt heute alle fünf
Bruttofelder auf `0.00` und füllt nur die Zahlung, sodass Bruttosumme und
Zahlung nicht balancieren. Korrekt trägt ein USt-neutraler Bargeldfluss seinen
Betrag im 0-%-Feld (Feld 5), wie die DSFinV-K-Beispiele in Anhang I (Geldtransit,
Privatentnahme). Der Builder wird entsprechend korrigiert; der Export bucht den
Betrag ohnehin schon unter `UST_SCHLUESSEL` 5, erst danach stimmen signierte
`processData` und Export überein und der Eigenbeleg ist kassensturzfähig.

**Berechtigung.** Der kassenwirksame Storno (`stornierung-erteilt`) bleibt der
Rolle `serviceleitung` (und Admin) vorbehalten, wie heute Stornierung und
Auszahlung. Die geldneutrale Korrektur unbezahlter Positionen
(`bestellung-korrigiert`) und die Umbuchung können regulären Servicekräften
offenstehen; finale Rollenzuordnung wird bei der Umsetzung an der bestehenden
Rollen- und Endpunkt-Aufteilung (Serviceleitung vs. Service) ausgerichtet.

**DSFinV-K-Mapper.** Der Mapper entfällt für `auszahlung-geleistet`. Der
kassenwirksame Storno wird ein negativer Beleg (`BON_TYP` Beleg, GV_TYP `Umsatz`,
Zahlart Bar, Referenz auf die Zahlung). Er ist eine Warenrücknahme als negative
Belegdarstellung (DSFinV-K Tz. 4.2.5 und Hinweis zu `AVBelegstorno`), kein
Vorgangs-Storno: `BON_STORNO` bleibt `0`. Das Storno-Kennzeichen ist der
vollständigen Aufhebung eines ganzen Belegs vorbehalten; jotti bucht eine
Teilmenge negativ zurück und verkettet sie allein über die Referenz auf die
Zahlung. Der bestehende `direktverkauf-storniert` wird auf dasselbe Modell
umgestellt (er setzt heute `BON_STORNO = 1`). `bestellung-korrigiert` und
`bestellung-umgebucht` werden geldneutrale Vorgänge (`AVBestellung`, keine
Zahlart, keine Bargeldwirkung). Der GV_TYP `Auszahlung` wird im Mapper nicht
mehr verwendet.

**Tagesabschluss.** Das `tagesabschluss-erstellt`-Event verliert das Feld
`AuszahlungenCents`; Warenrücknahmen fließen als negative Umsätze in die
Umsatz-/Stornokennzahlen. (Berührt den separat geplanten Kassenabschluss-Umbau,
siehe Further Notes.)

**Aktive Entwicklungsphase.** Gemäß AGENTS.md ist jotti pre-release: Events
werden direkt geändert, alte Events nicht migriert, kein Dual-Read. DB-Schema-
Änderungen erfolgen direkt in der bestehenden Initial-Migration. Breaking
Changes sind erwünscht.

## Testing Decisions

Tests prüfen ausschließlich beobachtbares Verhalten der Module über ihre
öffentliche Schnittstelle, keine internen Implementierungsdetails. Sie sind
DB-frei und table-driven, wo möglich. Prior Art: die bestehenden
`processData`-Tests (unit-Build-Tag), die Projektionstests der Tisch-Session
und die DSFinV-K-Mapper-Tests.

Getestet werden alle vier Deep Modules:

- **processData-Builder.** Geldneutrale Vorgänge (`Bestellung-V1`) erzeugen
  keine `:Bar`-Zeile; die Warenrücknahme erzeugt einen `Kassenbeleg-V1` mit
  korrekt negierten Bruttobeträgen je Steuersatz und einer negativen Bar-Zahlung,
  die der Summe entspricht. Steuersatz-Aufteilung inklusive Kombi-Positionen. Der
  Eigenbeleg (Geldtransit, Kassendifferenz) trägt seinen Betrag im 0-%-Feld, und
  die Bruttosumme gleicht die Bar-Zahlung aus.
- **TischSession-Projektion.** Über beliebige Event-Folgen gilt: der offene
  Betrag bleibt größer oder gleich null; eine Warenrücknahme bezahlter
  Positionen lässt den offenen Betrag unverändert und entfernt die Positionen;
  eine geldneutrale Korrektur reduziert offenen Betrag und Ausstehend-Liste;
  eine Umbuchung verschiebt nur unbezahlte Positionen.
- **Command-Routing.** Eine Storno-Anforderung erzeugt nach Bezahlstatus das
  korrekte Event; gemischte Anforderungen werden in einen geldneutralen und
  einen kassenwirksamen Teil aufgeteilt; betreffen die bezahlten Positionen
  mehrere Teilzahlungen, entsteht ein kassenwirksames Event je Zahlung (FIFO),
  jedes mit genau einer Zahlungsreferenz; der kassenwirksame Storno ist auf die
  Serviceleitung beschränkt; der Rückgabebetrag folgt aus den bezahlten
  Positionen und ist nicht frei wählbar.
- **DSFinV-K-Mapper.** Der kassenwirksame Storno wird ein negativer Beleg ohne
  `BON_STORNO`-Kennzeichen, mit Zahlart Bar und Referenz auf die Zahlung;
  Korrektur und Umbuchung sind geldneutral; kein Beleg trägt mehr den GV_TYP
  `Auszahlung`; der summierte Bargeldbestand stimmt mit den tatsächlichen
  Bar-Bewegungen überein (keine Doppelbuchung).

## Out of Scope

- **Kartenzahlung / unbare Rückgaben.** jotti ist bewusst bar-only; alle
  Rückgaben sind Bar.
- **Per-Gast-/Rechnungs-Aggregat.** Es wird kein eigenes „Gastgruppe"- oder
  „Rechnung"-Aggregat eingeführt. Die fiskalische Einheit ist die Zahlung
  (ein Bon); eine Warenrücknahme referenziert die jeweilige Zahlung. Die
  saubere Trennung mehrerer Gruppen ergibt sich daraus, ohne neues Aggregat.
- **Teil-Mengen-Rückgabe über die bestehende Positionslogik hinaus.** Die
  vorhandene Mengen-Auflösung (`PositionRef`, Restmengen) wird wiederverwendet,
  nicht neu entworfen.
- **Geldtransit und Kassendifferenz.** Bleiben als Vorgänge unverändert; nur
  der gemeinsame Eigenbeleg-`processData`-Builder wird mitkorrigiert (Feld 5,
  siehe Implementation Decisions). Ein darüber hinausgehender Umbau dieser
  Vorgänge ist nicht Teil dieser PRD.

## Further Notes

**Bezug zu bestehenden Vorhaben.** Der Wegfall von `AuszahlungenCents` im
Tagesabschluss berührt das separat geplante Kassenabschluss-Vereinfachungs-PRD
(Bereinigung der 0-Beträge im `tagesabschluss-erstellt`-Event). Beide Umbauten
sollten beim Schema des Tagesabschluss-Events aufeinander abgestimmt werden.

**Konsistenz-Gewinn.** Nach dem Umbau folgen Tisch-Storno und
Direktverkauf-Storno demselben Modell (sofort kassenwirksam, REF auf den
Ursprungsbeleg, Stornobeleg auf Anforderung), und es gibt nur noch eine
Geld-raus-Aktion am Tisch statt der entkoppelten Storno-plus-Auszahlung-Folge.
