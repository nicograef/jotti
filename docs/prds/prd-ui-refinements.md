# PRD: UI-Verfeinerungen (Dashboard, Produktstatistik, Autofokus, Tischsuche)

## Problem Statement

Mehrere zusammenhängende UI-Reibungspunkte erschweren die Bedienung — besonders
mobil (BYOD), unter Stress, durch ehrenamtliche Helfer:

1. **Admin-Dashboard-Layout.** Auf der Übersicht steht die Produktstatistik
   („Verkäufe pro Produkt") als lange, ungedeckelte Liste ohne eigenen Scroll —
   sie wächst mit der Produktzahl unbegrenzt die Seite hinunter. Die wichtigeren
   Stornierungen stehen darunter und rutschen dadurch aus dem Blick.
2. **Verwirrende Produktzahlen.** In der Produktstatistik ruhen „Ausgegeben"
   (bestellt − Korrekturen) und „Umsatz" (kassiert − Storno) auf
   unterschiedlichen Grundlagen. Sie gehen nicht ineinander auf, was verwirrt:
   die Euro-Spalte ist nicht der Wert der gezählten Portionen.
3. **Automatische Tastatur.** Dialoge und Formulare fokussieren beim Öffnen
   automatisch das erste Eingabefeld, wodurch die Mobil-Tastatur aufspringt. Im
   Bestell-Drawer ist das erste Feld sogar das *optionale* Kommentarfeld; die
   Tastatur verdeckt die primäre Aktion („Bestellung aufnehmen" / „Kassieren"),
   der Nutzer muss die Tastatur erst schließen, um abzusenden.
4. **Tischsuche sucht das Falsche.** Die Suche auf der Service-Hauptseite filtert
   nur die Favoriten („Meine Tische") — eine kleine Menge, in der man selten
   sucht. Um einen beliebigen Tisch zu finden, muss man erst den „Alle Tische"-
   Drawer öffnen, der eine *zweite*, getrennte Suche über alle aktiven Tische
   hat. Zwei Suchfelder, zwei Datenbestände, verwirrend.

Begleitend traten beim Audit weitere Reibungspunkte derselben Klasse zutage
(primäre Aktion hinter der Tastatur, zu kleine Touch-Ziele,
Formatierungs- und Begriffs-Inkonsistenzen), die mitbehoben werden.

## Solution

1. **Dashboard umsortieren und scrollbar machen.** Auf der Live-Übersicht rückt
   der Stornierungen-Block über die Produktstatistik. Die geteilte
   Produktstatistik-Liste bekommt eine Höhenbegrenzung mit eigenem vertikalem
   Scroll — auf der Übersicht *und* auf „Berichte & Export". (Auf „Berichte &
   Export" stehen die Stornierungen bereits über der Produktstatistik; dort ist
   nur der Scroll nötig.)
2. **Umsatz auf Bestellbasis umstellen.** „Umsatz" in der Produktstatistik wird
   auf dieselbe Ereignis-Grundlage wie „Ausgegeben" gestellt: der Euro-Wert
   (zu Bestellzeit-Preisen) genau der Portionen, die „Ausgegeben" zählt (bestellt
   − Korrekturen, inkl. Direktverkauf). Damit ist die Produkttabelle in sich
   stimmig (Umsatz = Wert der ausgegebenen Portionen) und geht im Normalfall
   ohne nachträgliche Stornos mit dem kassierten Gesamtumsatz auf. Der
   Beschreibungstext wird auf eine kurze Aussage reduziert (z. B. „Zahlen
   basieren auf Bestellungen"). Gilt auf Übersicht und „Berichte & Export".
3. **Keine automatische Tastatur.** Der Autofokus beim Öffnen wird zentral in
   den geteilten Overlay-Primitiven (Drawer/Dialog/Sheet) unterdrückt — kein
   Dialog und kein Formular öffnet mehr beim Erscheinen die Tastatur. Der Nutzer
   tippt selbst das Feld an, das er befüllen will; die primäre Aktion bleibt
   sichtbar.
4. **Eine Suche über alle Tische.** Die Suche auf der Service-Hauptseite sucht
   direkt über alle aktiven Tische; ein Treffer öffnet den Tisch direkt. Der
   „Alle Tische"-Drawer bleibt zum Durchblättern und Favorisieren erhalten,
   verliert aber sein nun redundantes zweites Suchfeld.

Begleitend (derselbe Problemkreis):

5. **Admin-Dialoge scrollbar mit fixierter Aktionsleiste.** `Dialog`/
   `AlertDialog` erhalten dieselbe Behandlung wie der Drawer (Höhen-Cap +
   interner Scroll + gepinnte Fußleiste), damit die Absende-Schaltfläche nie
   hinter der Tastatur verschwindet — u. a. Zählhilfe und Kassenabschluss.
6. **Zählhilfe-Touch-Ziele.** Die zu kleinen (32 px) Zähl-Eingaben werden auf
   ein mobiltaugliches Touch-Maß vergrößert.
7. **Währungsformat konsistent.** Admin-seitige Beträge werden über die
   vorhandene `formatEuro`-Hilfe (geschütztes Leerzeichen) gerendert, damit das
   „€" auf schmalen Displays nicht in die nächste Zeile umbricht.
8. **Begriffe vereinheitlichen.** Für die Servicekraft-Rolle wird durchgängig
   „Servicekraft" statt „Bediener" verwendet; in der Helfer-Verwaltung wird die
   Person durchgängig als „Helfer" benannt (Feld „Benutzername" als Zugangs-
   Credential bleibt), statt „Helfer"/„Benutzer" im selben Fluss zu mischen.

## User Stories

1. Als Admin möchte ich auf der Übersicht die Stornierungen über der
   Produktstatistik sehen, damit die wichtigeren Vorgänge nicht unter einer
   langen Produktliste verschwinden.
2. Als Admin möchte ich, dass die Produktstatistik-Liste einen eigenen Scroll
   hat, damit sie bei vielen Produkten nicht die ganze Seite hinunterwächst —
   auf der Übersicht und auf „Berichte & Export".
3. Als Admin möchte ich, dass „Ausgegeben" und „Umsatz" je Produkt auf
   derselben Grundlage beruhen (Umsatz = Wert der ausgegebenen Portionen),
   damit die Zahlen nachvollziehbar zusammenpassen.
4. Als Admin möchte ich einen kurzen, klaren Hinweis („Zahlen basieren auf
   Bestellungen"), damit ich die Grundlage der Zahlen sofort verstehe.
5. Als Servicekraft möchte ich, dass beim Öffnen eines Drawers/Dialogs nicht
   automatisch die Tastatur aufspringt, damit die primäre Aktion sichtbar
   bleibt und ich nicht erst die Tastatur schließen muss.
6. Als Admin möchte ich, dass beim Öffnen eines Formulardialogs nicht
   automatisch die Tastatur aufspringt, aus demselben Grund.
7. Als Servicekraft möchte ich die Suche auf der Hauptseite über alle aktiven
   Tische nutzen und einen Treffer direkt öffnen, damit ich jeden Tisch finde,
   ohne erst einen Drawer zu öffnen.
8. Als Servicekraft möchte ich weiterhin den „Alle Tische"-Drawer zum
   Durchblättern und Favorisieren nutzen — ohne verwirrendes zweites Suchfeld.
9. Als Admin möchte ich, dass in Formulardialogen (inkl. Zählhilfe,
   Kassenabschluss) die Absende-Schaltfläche erreichbar bleibt (Scroll,
   fixierte Fußleiste), damit ich die Aktion auch bei geöffneter Tastatur oder
   viel Inhalt auslösen kann.
10. Als Admin möchte ich in der Zählhilfe ausreichend große Eingabefelder,
    damit ich mich beim Zählen unter Stress nicht vertippe.
11. Als Nutzer möchte ich Beträge ohne umgebrochenes „€" sehen, damit Salden
    und Summen sauber lesbar sind.
12. Als Nutzer möchte ich innerhalb eines Ablaufs konsistente Begriffe
    („Servicekraft", „Helfer"), damit die Oberfläche nicht verwirrt.

## Implementation Decisions

**Dashboard-Layout (Übersicht):**
- In der Live-Reporting-Sektion werden die beiden Geschwister-Blöcke
  Produktstatistik und Stornierungen vertauscht (Stornierungen zuerst). Rein
  präsentational, kein geteilter State.

**Produktstatistik-Scroll (geteilt):**
- Die geteilte Produktstatistik-Komponente erhält einen Höhen-Cap mit
  `overflow-y-auto`, sodass die Liste innerhalb ihres Rahmens scrollt statt die
  Seite zu verlängern. Da die Komponente auf Übersicht und „Berichte & Export"
  geteilt wird, wirkt die Änderung auf beiden Seiten.

**Umsatz auf Bestellbasis (Backend, Single Source of Truth):**
- Die Reporting-Query für die Produktstatistik wird angepasst: der Umsatz-Term
  je Variante verwendet **dieselbe Ereignismenge und Gewichtung wie der
  Ausgegeben-Term** (`bestellung-aufgenommen` +, `bestellung-korrigiert` −,
  `direktverkauf-getaetigt` +), jeweils `einzelpreisCents × menge`. Damit ist
  der Umsatz je Variante der Euro-Wert genau der in „Ausgegeben" gezählten
  Portionen (zu den zum Bestellzeitpunkt eingefrorenen Preisen).
- `zahlung-kassiert`, `stornierung-erteilt` und `direktverkauf-storniert`
  fließen nicht länger in den **Produkt**-Umsatz ein.
- Bewusste Divergenz: Die KPI „Kassierter Umsatz", die Umsätze je Servicekraft
  und die fiskalischen Größen (Z-Bon / DSFinV-K) bleiben **kassenbasiert** und
  unverändert. Bei nachträglichen (kassenwirksamen) Stornos weicht der
  Produkt-Umsatz vom kassierten Umsatz um genau diese Storno-Beträge ab; ohne
  solche Stornos stimmen sie überein.
- Reiner Read-Query-/Read-Model-Eingriff: keine Migration, keine Änderung an
  Event-Formaten oder persistierten Daten (Freeze-Disziplin gewahrt). Nach der
  Query-Änderung `make sqlc` ausführen.

**Beschreibungstext:**
- Der erklärende Absatz über der Produkttabelle (geteilt) wird auf eine kurze
  Aussage reduziert (z. B. „Zahlen basieren auf Bestellungen"). Die
  Spaltenüberschriften „Ausgegeben"/„Umsatz" bleiben.

**Autofokus zentral unterdrücken (Frontend-Primitive):**
- In den geteilten Overlay-Content-Komponenten (Drawer, Dialog, Sheet) wird das
  Standard-Verhalten „ersten fokussierbaren Nachfahren fokussieren" beim Öffnen
  unterbunden (Radix `onOpenAutoFocus` verhindern), ohne den Fokus-Trap / die
  Tastaturbedienung (Tab, Escape, Fokusrückgabe beim Schließen) zu brechen.
- `AlertDialog` (reine Bestätigungen) fokussiert weiterhin eine Schaltfläche und
  bleibt unverändert.

**Tischsuche vereinheitlichen (Service):**
- Die Hauptseiten-Suche filtert über **alle aktiven Tische** (die bereits
  vorhandene „aktive Tische mit Favoriten"-Datenquelle wird genutzt), nicht mehr
  nur über Favoriten. Bei leerem Suchfeld zeigt die Hauptseite weiterhin die
  Favoriten („Meine Tische"); sobald gesucht wird, erscheinen Treffer aus allen
  aktiven Tischen, und ein Treffer öffnet den Tisch direkt.
- Das zweite Suchfeld im „Alle Tische"-Drawer entfällt; der Drawer bleibt zum
  Durchblättern (sortiert) und Favorisieren erhalten.
- Der Platzhaltertext der Hauptseiten-Suche wird passend gehalten (Suche über
  Tischname, der die Nummer enthält).

**Admin-Dialoge scrollbar mit fixierter Fußleiste (Frontend-Primitive):**
- `Dialog`/`AlertDialog`-Content bekommen einen Höhen-Cap (analog Drawer,
  `max-h`), einen intern scrollenden Inhaltsbereich und eine gepinnte Fußleiste,
  sodass die Aktionsschaltflächen bei geöffneter Tastatur oder viel Inhalt
  erreichbar bleiben. Wirkt auf alle Admin-Formulardialoge (u. a. Produkt-,
  Benutzer-, Tisch-Dialoge, Geldtransit, Zählhilfe, Kassenabschluss).

**Zählhilfe-Touch-Ziele:**
- Die Zähl-Eingaben werden auf ein mobiltaugliches Touch-Maß vergrößert
  (mindestens ~44 px Höhe), Layout weiterhin kompakt.

**Währungsformat:**
- Admin-seitige Beträge (u. a. Kassenabschluss, Tischauswahl-Saldo) werden über
  die vorhandene `formatEuro`-Hilfe gerendert statt „`{formatCents(x)} €`" mit
  normalem Leerzeichen.

**Begriffe:**
- Reporting-Oberfläche: „Bediener" → „Servicekraft" (kanonisch laut
  `docs/language.md`).
- Helfer-Verwaltung: Person durchgängig „Helfer" (Trigger, Dialogtitel,
  Schaltfläche, Toast); das Zugangsfeld bleibt „Benutzername". Falls das Team
  stattdessen die Entität „Benutzer" bevorzugt, gilt: eine Wahl, konsistent
  angewandt. `docs/language.md` bleibt Quelle der Wahrheit und wird bei
  UI-Begriffsänderung mitgezogen.

## Testing Decisions

Gute Tests prüfen **externes Verhalten**, nicht Implementierungsdetails.

- **Produkt-Umsatz (Backend, höchster Testwert).** Query-/Repository-Test für
  die Produktstatistik: Nach Bestellung + Korrektur + Kassierung ist der
  Produkt-Umsatz je Variante gleich dem Euro-Wert der ausgegebenen Portionen
  (bestellt − Korrektur, inkl. Direktverkauf). Insbesondere: eine nachträgliche
  (kassenwirksame) **Stornierung reduziert den Produkt-Umsatz nicht** (sie
  reduziert weiterhin die separate kassenbasierte KPI). Prior art: bestehende
  Reporting-Query-/Repository-Tests im Reporting-Kontext.
- **Tischsuche (Frontend).** Test der Service-Hauptseite: Eingabe eines
  Suchbegriffs, der auf einen **nicht favorisierten** aktiven Tisch passt, zeigt
  diesen als Treffer und ermöglicht das Öffnen. Prior art: `TablePage.test.tsx`,
  `TischAuswahlDrawer.test.tsx`.
- **Autofokus (Frontend).** Test, dass beim Öffnen eines Drawers/Dialogs **kein
  Eingabefeld** den Fokus erhält (kein Auto-Öffnen der Tastatur), am Beispiel
  des Bestell-Drawers (optionales Kommentarfeld). Externes, beobachtbares
  Verhalten (`document.activeElement`).
- **Nicht gesondert getestet** (rein präsentational / geringer Testwert, visuell
  zu verifizieren): Block-Reihenfolge auf der Übersicht, Scroll-Cap der
  Produktliste, Admin-Dialog-Scroll/Fußleiste, Touch-Ziel-Größen,
  `formatEuro`-Umstellung, Begriffs-Strings.

## Out of Scope

- Änderung der KPI „Kassierter Umsatz", der Umsätze je Servicekraft oder der
  fiskalischen Auswertungen (Z-Bon, DSFinV-K) — diese bleiben kassenbasiert.
- Änderung der Bedeutung von „Ausgegeben" (die Ereignismenge des
  Ausgegeben-Terms bleibt unverändert; der Umsatz zieht mit).
- Änderungen an Event-Formaten, DB-Schema oder persistierten Daten.
- Ein neues „zuletzt bearbeitete Tische"-Konzept getrennt von Favoriten —
  „Meine Tische" bleibt gleichbedeutend mit Favoriten.
- Vollständige, repo-weite Begriffs-Migration über die genannten Flächen hinaus
  (Reporting-Rolle, Helfer-Verwaltung); weitergehende Sprachpflege wäre ein
  eigenes Vorhaben.
- Verhaltensänderung des `AlertDialog`-Autofokus (Bestätigungsdialoge bleiben
  auf einer Schaltfläche fokussiert).

## Further Notes

- **Vertikale Slices für den Plan** (weitgehend unabhängig, einzeln
  auslieferbar): (1) Dashboard-Reihenfolge + Produktliste-Scroll; (2)
  Produkt-Umsatz auf Bestellbasis + Beschreibungstext; (3) Autofokus zentral +
  Admin-Dialog-Scroll/Fußleiste (gemeinsamer Problemkreis „primäre Aktion hinter
  Tastatur"); (4) Tischsuche über alle Tische + Drawer-Suchfeld entfernen; (5)
  kleinere Politur: Zählhilfe-Touch-Ziele, `formatEuro`, Begriffe.
- Slice 3 verbindet die vom Nutzer genannte Autofokus-Reibung mit dem beim Audit
  gefundenen Admin-Dialog-Scroll-Problem, weil beide dieselbe Ursache-Klasse
  („primäre Aktion nicht erreichbar") betreffen und sich in denselben geteilten
  Primitiven lösen lassen.
- Nach Backend-Query-Änderung `make sqlc`; nach Frontend-Änderungen `make lint`.
