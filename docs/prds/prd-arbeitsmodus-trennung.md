# PRD: Arbeitsmodus-Trennung (Tischservice / Direktverkauf)

## Problem Statement

jotti kennt zwei Verkaufsabläufe: **Tischservice** (Servicekraft nimmt am Tisch Bestellungen
auf, bestätigt die Ausgabe und kassiert) und **Direktverkauf** (Barverkauf an der Theke:
bestellen, zahlen und ausgeben in einem Schritt, der Gast erhält einen Abholbon). Auf einem
Vereinsfest sind das getrennte Arbeitsplätze: Ein Helfer arbeitet entweder stundenlang als
Servicekraft an den Tischen oder stundenlang an der Theke — gewechselt wird selten bis nie.
Manche Vereine nutzen sogar nur einen der beiden Abläufe.

Die heutige Service-Oberfläche bildet das nicht ab:

- Die Tischauswahl-Seite — die Startseite aller Service-Nutzer — beginnt mit einem
  bildschirmbreiten, primären „Direktverkauf"-Button. Für Servicekräfte an den Tischen (die
  Mehrheit) ist das prominenteste Element ihrer Startseite nutzlos und verdrängt ihren
  eigentlichen Arbeitsbereich („Meine Tische") nach unten.
- Helfer an der Theke müssen bei jedem Öffnen von jotti erst über die Tischauswahl-Seite in
  den Direktverkauf navigieren — bei einem Arbeitsplatz, der den ganzen Abend derselbe
  bleibt. Nach Browser-Neustart oder versehentlichem Zurück-Tippen landen sie wieder bei den
  Tischen.
- Der Direktverkauf präsentiert sich als Unterseite der Tischauswahl (Rücklink „Meine
  Tische" im Kopfbereich), obwohl er fachlich ein gleichrangiger Arbeitsbereich ist.

## Solution

Der Service-Bereich erhält zwei gleichrangige **Arbeitsmodi**: **Tischservice** und
**Direktverkauf**. Ein Helfer arbeitet immer in genau einem Modus; die Oberfläche zeigt nur
diesen Arbeitsbereich.

- Beim Öffnen des Service-Bereichs landet der Helfer direkt in seinem **zuletzt genutzten
  Modus** (gemerkt pro Gerät — die Helfer nutzen ihre eigenen Smartphones). Wer jotti zum
  ersten Mal öffnet, startet im Tischservice.
- Ein Helfer an der Theke wechselt **einmal zu Beginn** über das Benutzermenü in den
  Direktverkauf — und landet von da an bei jedem Öffnen, nach jedem Neustart, direkt dort.
- Der prominente Direktverkauf-Button auf der Tischauswahl-Seite **entfällt ersatzlos**. Der
  Moduswechsel lebt ausschließlich im Benutzermenü (wo bereits Verwaltung, Theme und Logout
  liegen).
- Der Kopfbereich des Service-Bereichs zeigt den aktiven Modus an („Meine Tische" bzw.
  „Direktverkauf"). Der Direktverkauf ist keine Unterseite mehr: Der Rücklink „Meine Tische"
  im Direktverkauf entfällt; nur die Tischdetail-Seite behält ihren Rücklink zur
  Tischauswahl.

Für Vereine, die nur Tischservice machen, ist der Direktverkauf damit aus dem Weg — ein
unbenutzter Menüeintrag stört nicht. Für Vereine, die nur Direktverkauf machen, genügt ein
einmaliger Wechsel pro Gerät.

## User Stories

### Servicekraft am Tisch

1. Als Servicekraft im Tischservice möchte ich nach dem Login direkt auf „Meine Tische"
   landen, damit meine Startseite meinen Arbeitsbereich zeigt und nicht den eines anderen
   Arbeitsplatzes.
2. Als Servicekraft im Tischservice möchte ich keinen prominenten Direktverkauf-Button über
   meinen Tischen sehen, damit die wichtigste Fläche meiner Startseite meinen Tischen
   gehört und ich nichts versehentlich antippe.
3. Als Servicekraft möchte ich von der Tischdetail-Seite weiterhin mit einem Tipp zurück zur
   Tischauswahl kommen, damit sich an meinem gewohnten Arbeitsfluss nichts ändert.

### Helfer an der Theke (Direktverkauf)

4. Als Helfer an der Theke möchte ich zu Schichtbeginn einmal über das Benutzermenü in den
   Direktverkauf wechseln, damit ich danach durchgehend in meinem Arbeitsbereich bin.
5. Als Helfer an der Theke möchte ich bei jedem Öffnen von jotti — auch nach
   Browser-Neustart oder erneutem Login — direkt im Direktverkauf landen, damit ich nicht
   vor jedem Verkauf durch die Tischauswahl navigieren muss.
6. Als Helfer an der Theke möchte ich im Direktverkauf keinen permanenten Rücklink „Meine
   Tische" sehen, damit ich nicht versehentlich aus meinem Arbeitsbereich herausnavigiere.
7. Als Helfer an der Theke möchte ich im Kopfbereich erkennen, dass ich im Direktverkauf
   arbeite, damit ich mich jederzeit orientieren kann.

### Helfer, die wechseln

8. Als Helfer, der spontan die Station wechselt (z. B. von der Theke an die Tische), möchte
   ich jederzeit über das Benutzermenü in den anderen Modus wechseln können, damit der
   Stationswechsel keinen Admin und keine Neuanmeldung braucht.
9. Als Helfer möchte ich, dass jotti sich nach einem Wechsel den neuen Modus merkt, damit
   der nächste Aufruf wieder direkt am richtigen Arbeitsplatz startet.
10. Als Helfer möchte ich auch über einen direkten Link oder ein Lesezeichen in einen Modus
    springen können, und jotti merkt sich diesen als zuletzt genutzten, damit sich die App
    konsistent verhält, egal wie ich navigiere.

### Serviceleitung und Vereins-Admin

11. Als Serviceleitung oder Admin im Service-Bereich möchte ich denselben Moduswechsel im
    Benutzermenü haben wie die Servicekräfte, damit es nur ein Bedienkonzept gibt.
12. Als Vereins-Admin eines reinen Tischservice-Vereins möchte ich, dass der Direktverkauf
    meine Helfer nicht ablenkt, damit ich keine Sonderschulung für ein ungenutztes Feature
    brauche.
13. Als Vereins-Admin eines reinen Direktverkauf-Vereins möchte ich, dass meine Helfer nach
    einmaligem Wechsel pro Gerät dauerhaft im Direktverkauf arbeiten, damit der Tischservice
    ihnen nicht im Weg steht.

### Entwickler

14. Als Entwickler möchte ich das Arbeitsmodus-Konzept in genau einem Frontend-Modul
    gekapselt haben (Modus-Typ, Persistenz, Lesen/Schreiben), damit Routing, Layout und Menü
    dieselbe Quelle nutzen und der Modus testbar ist.
15. Als Entwickler möchte ich, dass die Arbeitsmodus-Trennung ohne Backend-Änderung
    auskommt, damit der Umbau ein reiner Frontend-Schnitt bleibt.

## Implementation Decisions

### Arbeitsmodus-Modul

- Ein neues, kleines Frontend-Modul kapselt den Arbeitsmodus: zwei Werte (Tischservice,
  Direktverkauf), Persistenz pro Gerät über localStorage, Lese- und Schreibzugriff über eine
  schmale Schnittstelle (Hook bzw. Funktionen) — analog zur bestehenden Theme-Präferenz.
- Standardwert ohne gespeicherte Präferenz ist Tischservice.
- Die Präferenz ist eine **Geräte-Einstellung**: Sie überlebt Logout und Login (BYOD — die
  Geräte sind persönlich) und wird nicht pro Benutzer im Backend gespeichert.
- Der gespeicherte Modus wird beim **Besuch einer Modus-Route** aktualisiert (nicht nur beim
  Wechsel über das Menü) — damit zählen auch Deep-Links und Lesezeichen als „zuletzt
  genutzt". Die Tischdetail-Seite gehört zum Modus Tischservice.

### Routing

- Der Einstieg in den Service-Bereich leitet auf die Route des zuletzt genutzten Modus
  weiter (statt heute fix auf die Tischauswahl).
- Die bestehenden Routen für Tischauswahl, Tischdetail und Direktverkauf bleiben bestehen;
  es gibt keine neuen Routen.

### Service-Layout und Navigation

- Der Kopfbereich des Service-Layouts zeigt den aktiven Modus als Titel: „Meine Tische" im
  Tischservice, „Direktverkauf" im Direktverkauf. Die separate Seitenüberschrift auf der
  Direktverkauf-Seite entfällt dafür (keine doppelte Benennung).
- Der Rücklink im Kopfbereich erscheint nur noch auf der Tischdetail-Seite (zurück zur
  Tischauswahl). Im Direktverkauf gibt es keinen Rücklink — er ist keine Unterseite mehr.
- Das Benutzermenü erhält innerhalb des Service-Bereichs einen Wechsel-Eintrag, der immer
  den **jeweils anderen** Modus anbietet („Direktverkauf" bzw. „Meine Tische"). Er steht
  allen Rollen mit Service-Zugang zur Verfügung (Servicekraft, Serviceleitung, Admin) und
  setzt die Geräte-Präferenz beim Wechsel.
- Der bildschirmbreite Direktverkauf-Button auf der Tischauswahl-Seite wird ersatzlos
  entfernt.

### Kein Backend-Anteil

- Keine Schema-, API- oder Event-Änderungen. Der Direktverkauf selbst (Aggregat, Endpunkte,
  Bondruck) bleibt unverändert; es ändert sich ausschließlich, wie die Service-Oberfläche
  ihn erreicht.

## Testing Decisions

- **Was ein guter Test ist:** Tests prüfen von außen beobachtbares Verhalten — wohin
  geleitet wird, welcher Modus nach welcher Aktion gespeichert ist — nicht
  Implementierungsdetails wie localStorage-Schlüssel oder interne Strukturen.
- **Arbeitsmodus-Modul** (Prior Art: bestehende Hook-Tests): ohne Präferenz gilt
  Tischservice; nach Setzen eines Modus liefert das Modul diesen zurück; der Wert überlebt
  ein Neu-Initialisieren (Persistenz).
- **Service-Einstieg** (Prior Art: bestehende Routing-Tests): der Einstieg in den
  Service-Bereich leitet je nach gespeichertem Modus auf die Tischauswahl bzw. den
  Direktverkauf weiter.
- **Nicht automatisiert getestet:** Benutzermenü-Eintrag, Kopfbereich-Titel und das
  Entfernen des Buttons (visuell verifiziert, mobile Viewports).

## Out of Scope

- **Org-Einstellung zum Abschalten des Direktverkaufs:** bewusst verworfen — als reiner
  Menüeintrag kostet ein ungenutzter Modus nichts, und jede Admin-Einstellung erhöht die
  Komplexität für ehrenamtliche Teams. Kann nachgerüstet werden, wenn echte Vereine danach
  fragen.
- **Modus-Zuweisung pro Benutzer durch den Admin** (Backend-Feld): bewusst verworfen —
  Helfer wechseln Stationen spontan; eine Geräte-Präferenz genügt.
- **Funktionale Änderungen am Direktverkauf** (Verkaufsablauf, Stornierung, Abholbon,
  Druckverhalten).
- **Der Admin-Bereich** bleibt unberührt.

## Further Notes

- **Arbeitsmodus** ist ein neuer Begriff der Oberfläche (kein Domänenbegriff des Backends).
  Bei der Umsetzung in `docs/language.md` aufnehmen, damit die Benennung verbindlich wird.
  Wichtig: Es gibt bewusst keine `Verkaufsstelle`-Entität (siehe `docs/language.md`,
  Direktverkauf) — der Modus ist eine reine Frontend-Sicht, kein Stammdatum.
- Sollte sich später herausstellen, dass Vereine eine zentrale Steuerung brauchen (Modus pro
  Benutzer oder Org-Toggle), ist das Arbeitsmodus-Modul die einzige Stelle, an der die
  Bezugsquelle der Präferenz ausgetauscht werden müsste.
