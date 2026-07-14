# PRD: Spektral-Redesign des App-Frontends

## Problem Statement

Die Website wurde neu gestaltet und setzt den Spektral-Verlauf als tragendes
Markenelement ein (ADR 05). Das App-Frontend, in dem Servicekräfte und Admins
tatsächlich arbeiten, wirkt dagegen markenfern: Die Wortmarke ist schlichter
schwarzer Text, Ladezustände sind graues Standard-Pulsieren, Überschriften nutzen
dieselbe Schrift wie der Fließtext, und es gibt keinerlei sichtbare Verbindung zur
Spektral-Sprache der Website. Wer von jotti.rocks in die App wechselt, erlebt einen
Stilbruch.

Zugleich gibt die App wenig Interaktionsfeedback: Der einzige Erfolgsmechanismus
nach Bestellen, Kassieren und Direktverkauf ist ein kleiner Toast am Bildschirmrand,
der im Trubel eines Vereinsfests leicht übersehen wird. Statuswechsel, Listenaufbau
und Zahlenänderungen passieren sprunghaft und ohne Übergang.

Für den Neuauftritt liegt ein High-Fidelity-Design-Handoff vor
(`docs/prds/design_handoff_spektral_redesign/`: Delta-Spezifikation, interaktiver
HTML-Prototyp, Screenshots Hell/Dunkel). Er ist als Restyling-Layer angelegt, nicht
als Layout-Umbau, und wurde gegen den echten Quellcode rekonstruiert. Die
Verifikation gegen die Codebasis bestätigt ihn weitgehend; drei Punkte wurden mit
dem Betreiber entschieden (siehe Implementation Decisions): Der vorgesehene
Fortschrittsbalken hätte keinen Konsumenten, der Direktverkauf fehlt im Handoff,
und die Doku nutzt bewusst kein Space Grotesk.

## Solution

Die Spektral-Markensprache der Website wird dezent auf das App-Frontend übertragen:
Login, Service-Bereich und Verwaltung. Es ist ein reiner Restyling-Layer — bestehende
Layouts, Komponentenstruktur, Bedienabläufe und Barrierefreiheits-Entscheidungen
bleiben unangetastet, Grün bleibt die Primärfarbe, die Status-Semantik
(Grün/Rot/Amber) ändert sich nicht.

Das Spektrum erscheint an genau vier Stellen und nirgendwo sonst:

1. **Wortmarke:** „jotti" mit Spektralverlauf als Text-Füllung (Login, Admin-Sidebar,
   mobiler Admin-Kopf), in der Heading-Schrift Space Grotesk.
2. **Hintergrund-Glows:** weiche, stark geblurte Farbkreise hinter der Login-Karte
   (mit langsamer Drift) und als Ellipsen-Paar hinter den Admin-Seitenköpfen, mit
   eigener Farbstimmung je Seite.
3. **Ladezustände:** alle Skeleton-Platzhalter shimmern mit leichter Spektral-Tönung
   statt grau zu pulsieren.
4. **Hairline-Akzente:** eine feine Spektral-Linie an der Oberkante der
   Hero-Kennzahlkarte der Übersicht und ein schmaler Spektral-Marker am aktiven
   Sidebar-Navigationseintrag.

Dazu kommt ein verbindliches Motion-Inventar aus subtilen UX-Animationen, die
Bediensicherheit erhöhen statt zu dekorieren: Press-Feedback auf Buttons und
Steppern, sanfte Übergänge bei Tab-Wechsel, Listeneintritt und Statuswechseln, ein
Erfolgs-Pop als unübersehbare Buchungsbestätigung im Service (ersetzt dort die
Erfolgs-Toasts, inklusive Direktverkauf, mit Tap-to-dismiss), animiert zählende
Beträge (Tisch-Saldo, Übersicht-Kennzahlen) und ein Live-Puls am
Kasse-offen-Punkt. Bei reduzierter Bewegungspräferenz stehen sämtliche Animationen
still.

Die Typografie wird mit der Website vereinheitlicht: Inter bleibt die Schrift für
UI und Fließtext, Space Grotesk kommt neu für Wortmarke und Überschriften hinzu.

## User Stories

1. Als Servicekraft will ich nach Bestellen, Kassieren und Direktverkauf eine
   unübersehbare, kurz eingeblendete Erfolgsbestätigung, damit ich im Trubel sicher
   weiß, dass die Buchung durch ist.
2. Als Servicekraft will ich die Erfolgsbestätigung antippen können, um sie sofort
   zu schließen, damit mich die Einblendung nicht ausbremst.
3. Als Servicekraft will ich, dass der Tisch-Saldo bei Änderungen animiert zum neuen
   Wert zählt, damit ich die Wirkung meiner Buchung nachvollziehen kann.
4. Als Servicekraft will ich beim Drücken von Buttons und Steppern ein spürbares
   Press-Feedback, damit ich auch bei hektischer Bedienung merke, dass ein Tipp
   angekommen ist.
5. Als Servicekraft will ich sanfte Übergänge bei Tab-Wechsel, Listeneintritt und
   Statuswechseln, damit Zustandsänderungen nachvollziehbar sind statt zu springen.
6. Als Servicekraft will ich beim Laden spektral schimmernde Platzhalter, damit die
   App auch im Wartezustand als jotti erkennbar ist.
7. Als Servicekraft will ich beim Anmelden eine spektrale Wortmarke über weichen
   Farb-Glows sehen, damit der Einstieg in die App die Marke der Website
   wiedererkennen lässt.
8. Als Admin will ich Wortmarke, einen dezenten Seitenkopf-Glow mit eigener
   Farbstimmung je Seite und eine feine Spektral-Kante auf der wichtigsten
   Kennzahlkarte, damit die Verwaltung erkennbar zur Marke gehört, ohne bunter zu
   arbeiten.
9. Als Admin will ich am pulsierenden Punkt erkennen, dass die Kasse offen ist bzw.
   die Übersicht live aktualisiert, damit ich den Betriebszustand auf einen Blick
   erfasse.
10. Als Admin will ich den aktiven Navigationseintrag zusätzlich an einem
    Spektral-Marker erkennen, damit die Orientierung in der Sidebar schneller ist.
11. Als Admin will ich, dass sich Kennzahlen der Übersicht bei Aktualisierung
    zählend statt sprunghaft ändern, damit Änderungen auffallen.
12. Als Admin will ich Kassenberichte weiterhin ohne Deko-Elemente drucken, damit
    die Ausdrucke sachlich und lesbar bleiben.
13. Als Helferin mit Bewegungsempfindlichkeit will ich, dass bei reduzierter
    Bewegungspräferenz alle Animationen stillstehen, damit mich die App nicht
    belastet.
14. Als Betreiber will ich eine einheitliche Typografie und Wortmarke über Website
    und App, damit beide als ein Produkt wahrgenommen werden.

## Implementation Decisions

- **Restyling-Layer, kein Umbau:** Bestehende Komponentenstruktur, Layouts, Semantik
  und WCAG-Entscheidungen (AA-Kontraste, Disabled-Token, Warn-Amber) bleiben
  unangetastet. Emerald-Grün bleibt Primärfarbe.
- **Vier-Stellen-Regel (verbindlich):** Spektrum ausschließlich in Wortmarke,
  Hintergrund-Glows, Ladezuständen und Hairline-Akzenten. Buttons, Badges,
  Status-Semantik, Fließtext, Icons und Diagramme bleiben unverändert.
- **Token-Übernahme:** Die acht Spektral-Farbtöne und der Verlauf werden wertidentisch
  von der Website übernommen (je Hell- und Dunkel-Variante), mit Quellverweis als
  Single-Source-Kommentar. Die App wendet die Dunkel-Werte über ihren bestehenden
  Dark-Mode-Mechanismus an; der unterscheidet sich vom Mechanismus der Website —
  maßgeblich ist Wertgleichheit, nicht Selektorgleichheit.
- **Zentrales Animations-Inventar:** Keyframes und wiederverwendbare
  Animations-Utilities werden einmal zentral definiert und an den Einsatzstellen
  konsumiert, nicht ad hoc pro Stelle. Eine zentrale Regel schaltet bei
  `prefers-reduced-motion: reduce` alle Animationen und Transitions ab.
- **Motion-Inventar (verbindlich):**

  | Animation | Dauer | Easing | Einsatz |
  |---|---|---|---|
  | Button-Press | 100 ms | linear | alle Buttons (leichtes Absenken/Skalieren), Stepper stärker |
  | Hover/Fokus-Transitions | 150 ms | ease | Hintergrund, Rahmen, Fokus-Ring |
  | Tab-/Detail-Wechsel (fadeUp) | 250 ms | ease | Tab-Inhalte, Drawer-Inhalte |
  | Statuswechsel (pop) | 350 ms | ease | Badges, Häkchen |
  | Listen-Eintritt (fadeUp) | 450 ms, Stagger 60 ms | sanfte Kurve | Karten-/Zeilenlisten, nur beim ersten Aufbau |
  | Erfolgs-Pop | 450 ms | Spring-Kurve | Buchungsbestätigung, Auto-Dismiss ~1,4 s |
  | Zahlen zählen | 700 ms | ease-out-cubic | Tisch-Saldo, Übersicht-Kennzahlen |
  | Skeleton-Shimmer | 1,6 s Loop | linear | alle Skeletons |
  | Live-Puls | 2,4 s Loop | ease-out | Kasse-offen-Punkt, Live-Anzeige |
  | Glow-Drift | 14–22 s Loop | ease-in-out | nur Login-Glows |

- **Wortmarke als Komponente:** wiederverwendbar, Spektralverlauf als Text-Füllung
  in Space Grotesk; bleibt echter, selektierbarer Text (kein Bild). Einsatzorte:
  Login-Karte, Admin-Sidebar, mobiler Admin-Kopf.
- **Glows sind reine Dekoration:** für Screenreader unsichtbar, klickdurchlässig,
  vom Druck ausgenommen. Login erhält drei Farbkreise mit langsamer Drift; jede
  Admin-Seite ein Ellipsen-Paar mit festem Farbpaar (Übersicht teal+violett,
  Kassentag orange+teal, Produkte grün+blau, Benutzer blau+rot, übrige
  teal+violett). Kontrast-Leitplanke: Deckkraft so begrenzt, dass Text darüber AA
  erfüllt.
- **Skeletons global auf Spektral-Shimmer:** ersetzt das graue Pulsieren an allen
  Verwendungsstellen einheitlich.
- **Erfolgs-Pop ersetzt Erfolgs-Toasts im Service:** bei Bestellen, Kassieren und
  Direktverkauf (Letzterer ergänzt gegenüber dem Handoff, gleicher Flow). Overlay
  mit Häkchen-Kreis in Primärgrün, Auto-Dismiss nach etwa 1,4 s plus Tap-to-dismiss
  (Ergänzung gegenüber dem Handoff, damit schnelle Helfer nicht warten). Sichtbare
  Statuswechsel (Badge, Listen) folgen erst nach dem Schließen. Fehler-Toasts und
  alle Toasts der Verwaltung bleiben unverändert; wo der Pop läuft, feuert kein
  Erfolgs-Toast mehr (kein Doppel-Feedback).
- **Zahlen-Zählen als wiederverwendbarer Hook:** animiert nur bei Wertänderung, mit
  tabellarischen Ziffern gegen Layout-Shift; ohne Animationsmöglichkeit (reduzierte
  Bewegung, Testumgebung) erscheint sofort der Endwert.
- **Listen-Eintritt nur beim ersten Aufbau:** Daten-Refetches animieren nie erneut.
- **Status-Punkt mit optionaler Puls-Variante:** nur am Kasse-offen-Punkt der
  Sidebar und an der Live-Anzeige der Übersicht; die übrigen Status-Punkte bleiben
  statisch.
- **Hairlines:** 2 px Spektral-Linie oben auf der Hero-Kennzahlkarte der Übersicht,
  3 px Spektral-Marker links am aktiven Sidebar-Eintrag, beide gedämpft.
- **Typografie:** Space Grotesk (variabel, SIL OFL, self-hosted) wird als
  Heading-Schrift über den bereits vorhandenen Heading-Font-Alias eingeführt;
  Überschriften und Kartentitel greifen sie damit ohne Einzelumbauten ab. Inter
  bleibt UI- und Fließtextschrift. Das Font-Paket ist die einzige neue Dependency.
  Die Doku bleibt bewusst vollständig bei Inter (bestehende, dokumentierte
  Entscheidung im Doku-Theme) — die Vereinheitlichung betrifft App und Website.
- **Kein Fortschrittsbalken:** Der im Handoff vorgesehene Spektral-Fortschrittsbalken
  entfällt bewusst — im Frontend existiert heute kein Fortschrittsindikator
  (TSE-Einrichtung und Export nutzen Spinner), die Komponente wäre toter Code.
  Falls später echter Fortschritt mit Prozentwert entsteht, wird sie mit ihrem
  Konsumenten gebaut.
- **Neues ADR bei Umsetzung:** Die dezente Spektral-Erweiterung der App löst die
  Festlegung „das Produkt-Frontend behält sein bestehendes, zurückhaltendes Design"
  aus ADR 05 ab (dort ausdrücklich auf die Website beschränkt). Analog zum
  Website-Präzedenzfall wird das als eigenes ADR festgehalten, inklusive der
  Vier-Stellen-Regel als neuer Leitplanke.
- **Benennung:** Der Markenbegriff ist „Spektral"/„Spektrum"; das Wort „Regenbogen"
  kommt wie auf der Website weder im Quelltext noch in Benutzer-sichtbaren Texten
  vor.

## Testing Decisions

- Gute Tests prüfen externes Verhalten: sichtbare Texte, Reihenfolge von Callbacks,
  Endzustände — nicht CSS-Klassen, Keyframe-Namen oder Animationsdauern.
- **Erfolgs-Pop:** erscheint nach erfolgreicher Buchung, schließt nach Ablauf der
  Anzeigedauer und vorzeitig bei Tap, und der nachgelagerte Statuswechsel wird erst
  nach dem Schließen ausgelöst (mit künstlichen Timern getestet).
- **Zähl-Hook:** endet exakt am Zielwert; liefert sofort den Endwert, wenn keine
  Animationsumgebung verfügbar ist.
- **Wortmarke:** „jotti" bleibt als Text im DOM auffindbar; bestehende Tests, die
  auf den Text matchen, laufen unverändert.
- **Status-Punkt:** Beschriftung und Zugänglichkeit bleiben mit Puls-Variante
  unverändert.
- Bestehende Vitest-Suites (Testing-Library-Komponententests, das vorherrschende
  Muster im Frontend) müssen unverändert grün bleiben: Dekorative Elemente sind für
  Accessibility-Queries unsichtbar, alle Texte und Bedienelemente bleiben erhalten.
  Snapshot-Tests gibt es nicht und werden nicht eingeführt.
- Visuelle Treue (Farben, Dauern, Easing, Hell/Dunkel) und das Verhalten bei
  reduzierter Bewegung sind Gegenstand der manuellen Abnahme gegen den interaktiven
  Prototyp, nicht von Unit-Tests.

## Out of Scope

- Layout- oder Strukturumbauten, neue Features, geänderte Bedienabläufe (einzige
  bewusste Verhaltensänderung: Erfolgs-Pop statt Erfolgs-Toast im Service).
- Änderungen an Farb- und Status-Semantik (Primärgrün, Destruktiv-Rot, Warn-Amber)
  oder an bestehenden Kontrast-/Barrierefreiheits-Entscheidungen.
- Website und Doku: beide bereits umgesetzt bzw. bewusst bei Inter; keine
  Doku-Typografie-Angleichung.
- Spektral-Fortschrittsbalken (kein Konsument, siehe Implementation Decisions).
- Spektrum an weiteren Stellen als den vier definierten (keine spektralen Buttons,
  Badges, Diagramme, Flächen).
- Fehler- und Warn-Feedback sowie alle übrigen Toasts.
- Beleg-/Bondruck und die Druck-Layouts der Kassenberichte (dort ändert sich nichts;
  neue Deko-Elemente sind lediglich vom Druck ausgenommen).

## Further Notes

- **Referenz:** Der Handoff-Ordner (`docs/prds/design_handoff_spektral_redesign/`)
  enthält die pixelgenaue Delta-Spezifikation, den interaktiven Prototyp (Hell/Dunkel,
  ein Screen interaktiv) und eine Phasenplan-Vorlage. Für Maße, Farben und Kurven
  gilt der Handoff; bei inhaltlichem Widerspruch gilt diese PRD. Nach der Umsetzung
  wird das Bundle gelöscht (Präzedenzfall Website-Redesign).
- **Abnahme:** Sichtvergleich aller Screens gegen den Prototyp in Hell und Dunkel,
  Reduced-Motion-Gegenprobe (DevTools-Emulation) und Druck-Stichprobe der
  Kassenberichte.
- **Klärungen mit dem Betreiber (2026-07-14):** voller Handoff-Umfang in einer PRD;
  Erfolgs-Pop ersetzt Erfolgs-Toasts im Service inklusive Direktverkauf, mit
  Tap-to-dismiss; Fortschrittsbalken entfällt; Glow-Drift am Login wird
  aufgenommen.
