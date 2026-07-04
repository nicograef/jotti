# PRD: Regenbogen-Branding für Logo und Website

Entstanden aus einer Branding-Diskussion (2026-07-04), direkt nach der Umfärbung aller Bild-Assets auf das
Markengrün. Kernentscheidungen: volles Spektrum im Logo, Website nur dekorativ bunt, Landing straffen,
Werte implizit. Die App/das Frontend sind bewusst ausgeklammert.

## Problem Statement

jotti steht für Gemeinnützigkeit, Hilfsbereitschaft und eine offene Gesellschaft. Der aktuelle Auftritt
transportiert das nur bedingt: Logo und Website sind monochrom auf das Markengrün aufgebaut und wirken
eher technisch-nüchtern als einladend. Gleichzeitig ist die Landing lang (zehn Sektionen), textreich und
stellt Technik- und Compliance-Themen prominent dar, obwohl die Zielnutzer nicht-technische Vereinshelfer
sind. Wer nicht vom Fach ist, muss sich den Nutzen von jotti erarbeiten, statt ihn gezeigt zu bekommen.

Die Buntheit soll dabei keine politische Botschaft senden. Ein Regenbogen wird von vielen als
Pride-Statement gelesen; die Gestaltung muss diese Lesart bewusst vermeiden, ohne die Farbigkeit
aufzugeben.

## Solution

Das gefaltete J bekommt einen kontinuierlichen Spektral-Verlauf (Rot bis Violett) entlang seiner
Faltgeometrie. Die Form, die Schattierung und die neutralen Bestandteile (Schriftzug, Hintergründe)
bleiben unverändert, nur der Farbton wandert über das Spektrum. Es entstehen keine Streifen und keine
Flaggen-Anordnung; das Markengrün liegt naturgemäß in der Mitte des Verlaufs. Die Spektral-Varianten
ersetzen die grünen Master und werden das kanonische Logo; die Website nutzt sie überall (Header,
Favicons, Doku). Die grünen Icon-Kopien der App bleiben unangetastet, bis die App in einem separaten
Vorhaben nachzieht.

Die Website bleibt in der Bedienung grün: Buttons, CTAs, Links und alle semantischen UI-Tokens ändern sich
nicht. Farbigkeit kommt ausschließlich dekorativ dazu, etwa als mehrfarbiger Hero-Hintergrund und als
feine Spektral-Haarlinie an wenigen, wiederkehrenden Stellen. Inhaltlich wird die Landing gestrafft:
Nutzen zuerst, einfachere Sprache, Technik und Compliance kompakt weiter hinten. Es kommen keine neuen
Inhalte dazu, und die Werte werden nicht ausformuliert, sondern transportieren sich über Farben, Ton und
die bestehende Positionierung.

## User Stories

1. Als Vereinsmensch, der jotti zum ersten Mal sieht, möchte ich ein freundliches, buntes Logo, damit
   jotti einladend und menschlich wirkt statt technisch-nüchtern.
2. Als wiederkehrender Besucher möchte ich das gefaltete J sofort wiedererkennen, damit der Farbwechsel
   die Markenidentität nicht bricht.
3. Als Besucher möchte ich die Buntheit als Ausdruck von Offenheit und Vielfalt wahrnehmen, ohne dass mir
   eine politische Botschaft begegnet, damit ich mich unabhängig von meiner Haltung willkommen fühle.
4. Als Betreiber möchte ich den Regenbogen als fließenden Verlauf ohne Flaggen-Anmutung (keine Streifen,
   keine Sechs-Farben-Anordnung), damit die Subtilität gestalterisch abgesichert ist und nicht vom
   Betrachter abhängt.
5. Als Besucher im Dark Mode möchte ich Logo und dekorative Farbelemente in einer darauf abgestimmten
   Ausprägung sehen, damit nichts grell wirkt oder im Kontrast absäuft.
6. Als Browser-Nutzer möchte ich das Favicon auch in 16 Pixeln als buntes J erkennen, damit der Tab
   auffindbar bleibt.
7. Als nicht-technischer Vereinshelfer möchte ich auf der Startseite in wenigen Sätzen verstehen, was
   jotti ist und was es mir bringt, damit ich ohne Fachwissen entscheiden kann, ob ich weiterlese.
8. Als Vereinsvorstand möchte ich den Nutzen (einfache Kasse fürs Fest, faire Bedingungen) vor Technik und
   Compliance präsentiert bekommen, damit die Entscheidung nicht von Fachbegriffen ausgebremst wird.
9. Als Kassenwart oder Prüfer möchte ich Compliance- und Technikinformationen weiterhin auf der Landing
   finden, nur kompakter und weiter hinten, damit die inhaltliche Tiefe erhalten bleibt.
10. Als mobiler Besucher möchte ich kürzere Sektionen und weniger Scroll-Last, damit die Seite auf dem
    Handy erfassbar bleibt.
11. Als Besucher möchte ich Call-to-Actions weiterhin eindeutig im Markengrün erkennen, damit die
    dekorative Farbe nicht mit der Bedienung konkurriert.
12. Als Besucher der Doku möchte ich dieselbe Marke sehen wie auf der Landing, damit Landing und Doku als
    ein Auftritt wahrgenommen werden.
13. Als Betreiber möchte ich die Werte implizit transportieren, damit keine Werte-Sektion und kein
    Statement-Text entsteht.
14. Als Betreiber möchte ich die Spektral-Assets reproduzierbar aus den Mastern generieren, damit künftige
    Anpassungen (andere Verlaufsachse, andere Tonwerte) kein manuelles Neuzeichnen erfordern.
15. Als Betreiber möchte ich, dass App und PWA vorerst unverändert grün bleiben, damit der
    Website-Rollout die Kassen-Geräte nicht berührt.
16. Als Betreiber möchte ich alle neuen Varianten und Landing-Zustände vor dem Ersetzen visuell abnehmen,
    damit nichts Ungeprüftes live geht.
17. Als Verein, der Abrechnungen mit dem hellen Full-Logo druckt, möchte ich dass das Logo auch gedruckt
    funktioniert, damit Dokumente seriös bleiben.
18. Als Betreiber möchte ich, dass die Asset-Doku den neuen Stand beschreibt, damit sie nicht erneut
    veraltet.

## Implementation Decisions

- Spektral-Generator: ein versioniertes Skript färbt die bestehenden Master per ortsabhängiger
  Hue-Zuordnung im OKLCH-Farbraum um. Der Farbton folgt der Verlaufsachse des J, die Helligkeit und
  Schattierung stammen unverändert aus dem Master, Chroma wird ans sRGB-Gamut geklemmt. Neutrale Töne
  (Schriftzug, Hintergründe) und der Alpha-Kanal bleiben unangetastet.
- Verlauf: kontinuierlich von Rot nach Violett über die Faltgeometrie, ohne diskrete Farbstreifen. Das
  Markengrün liegt in der Mitte des Spektrums und bleibt so Teil des Logos.
- Kanonische Master: der komplette Varianten-Satz (Full-Logo hell/dunkel/transparent, quadratische Icons
  hell/dunkel, Favicons in drei Größen, freigestelltes Symbol) wird spektral neu erzeugt und ersetzt die
  grünen Master. Die Website übernimmt alle neuen Assets; die grünen Kopien im App-Frontend bleiben bis zu
  einem separaten App-Follow-up bestehen.
- Dekorative Tokens: das Marken-Token-Set erhält Spektrum-Verlaufstoken mit Light- und Dark-Ausprägung.
  Eingesetzt werden sie nur dekorativ (Hero-Hintergrund, feine Spektral-Haarlinie an wenigen Stellen). Die
  semantischen UI-Tokens der Landing und des Doku-Themes bleiben unverändert grün.
- Landing-Straffung: Zielreihenfolge ist Nutzen zuerst (Hero, Versprechen, Features, Einblicke, So
  funktioniert's, Service), danach Technik und Compliance kompakt, zum Schluss der CTA. Jede Sektion
  bekommt ein Textbudget; Sprache wird vereinfacht (kurze Sätze, Fachbegriffe erklärt oder gestrichen).
  Keine neuen Inhalte, keine neuen Seiten.
- Werte bleiben implizit: kein Werte-Text, keine Flaggen- oder Kampagnen-Symbolik, keine saisonalen
  Aktionen.
- Kein neues Tooling im Website-Build; der Generator läuft als eigenständiges Skript außerhalb des Builds.

## Testing Decisions

- Gute Tests prüfen beobachtbares Verhalten: messbare Eigenschaften der erzeugten Bilder und der
  gebauten Seite, nicht die interne Rechenweise des Generators.
- Generator-Checks (automatisiert im Skript): der Hue-Verlauf deckt das Spektrum ab, Neutraltöne bleiben
  wertstabil, der Alpha-Kanal bleibt unverändert, der Varianten-Satz ist vollständig, Dateigrößen bleiben
  im Rahmen der bisherigen Assets.
- Visuelle Abnahme: alle Logo-Varianten sowie die Landing-Zustände (Light/Dark, Mobil/Desktop) werden vor
  dem Ersetzen der Master zur Freigabe vorgelegt.
- Prior Art: OKLCH-Stichprobenmessung aus der Grün-Umfärbung (Juli 2026) für die Generator-Checks;
  Headless-Firefox-Screenshots aus der Design-System-Migration für die visuelle Abnahme.
- Der Build bekommt bewusst keinen Link- oder Anker-Validator (bestehende Entscheidung, bleibt).

## Out of Scope

- App und Frontend: Theme, PWA-Icons, Manifest und alles auf den Kassen-Geräten. Die App zieht in einem
  eigenen Vorhaben nach.
- Neue Website-Inhalte: FAQ, Über-Seite, Werte-Sektion oder zusätzliche Landing-Sektionen.
- Druckmaterialien und Flyer über das bestehende Logo hinaus.
- Die Neuerstellung des Dokumentationsbilds mit allen Logo-Varianten und Farbangaben; es bleibt vorerst
  auf dem alten Stand und wird in der Asset-Doku entsprechend gekennzeichnet.
- Politische Statements oder Kampagnen jeder Art.

## Further Notes

- Das Spannungsfeld Regenbogen versus unpolitisch ist bewusst entschieden: die Farbigkeit konzentriert
  sich im Logo, der Rest des Auftritts bleibt ruhig und grün. Leitplanken sind der fließende Verlauf statt
  Streifen, der Verzicht auf die Sechs-Farben-Anordnung, das Grün als weiterhin dominante Produkt- und
  Bedienfarbe sowie der Verzicht auf jeden Werte-Text.
- Übergangszustand: bis zum App-Follow-up existieren zwei Logo-Stände (Website spektral, App grün). Das
  ist akzeptiert und wird in der Asset-Doku vermerkt.
- Die Asset-Doku wird im Zuge der Umsetzung aktualisiert (Farbbeschreibungen, Einsatzregeln, Hinweis auf
  den Übergangszustand).
