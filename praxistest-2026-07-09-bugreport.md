# Bugreport Praxistest 2026-07-09 — PWA/iOS-Probleme im Service-UI

Analyse des Bildmaterials aus dem Onsite-Test (lokales Notebook im Vereins-WLAN,
`…lokal.jotti.rocks`, mehrere iPhones + mind. ein Android, Service/Serviceleitung; teils
installierte PWA „Zum Home-Bildschirm", teils Safari-Tab). Ergänzt die Notizen in
`praxistest-2026-07-09.md`.

**Fassung 2** — überarbeitet nach einem adversarialen Multi-Experten-Audit (Code-Mechanik
gegen den Quellcode, Bibliotheks-/Plattform-Behauptungen gegen vaul-1.1.2-Quellcode,
GitHub-Issues, WebKit-Bugzilla und MDN, Bildinterpretationen gegen alle 1068 Videoframes).
Korrigiert gegenüber Fassung 1: der vaul-Lock-Mechanismus in Bug 1, die
Scroll-Offset-Behauptung, die Flicker-Frequenz, ein übersehener dritter Videozustand sowie
mehrere Detailpräzisierungen (siehe „Audit-Vermerk" am Ende).

## Beweismaterial — was wo zu sehen ist

Vier einzigartige Screenshots (zwei davon doppelt übermittelt, per MD5 identisch) und ein
Screenrecording (17,99 s, 60 fps, 1068 Frames). Zwei Geräte: **Gerät A** (Akku ~91 %,
„Vodafone WiFi Calling") lieferte das Video und die hellen Screenshots aus der installierten
PWA; **Gerät B** (Akku 37→33 %) die dunklen Screenshots aus dem Safari-Tab. Die Uhrzeiten
beider Geräte sind nicht abgleichbar, die minutengenaue Reihenfolge zwischen den Geräten also
nicht belastbar.

| Material | Zeit | Gerät/Modus | Seite/Komponente | Befund |
|---|---|---|---|---|
| Screenshot dark, kurze Liste | 18:29 | B, Safari-Tab (Toolbar mit URL-Kapsel unten), Dark Mode | `ZahlungDrawer` („Zahlung für Tisch 01", 3 Positionen, 6,80 €), Trinkgeld 10,00 / Erhalten 20,00 eingegeben, Rückgeld 10,00 € / Trinkgeld 3,20 € angezeigt | „Kassieren"-Button außerhalb des sichtbaren Bereichs; direkt unter dem Kommentarfeld ist eine 2 px hohe grüne Kante messbar (RGB ≈ #10b981 = Dark-Mode-Primary, `index.css`) — die Oberkante des Kassieren-Buttons. Keine Tastatur offen: Der Footer ist schon **ohne** Tastatur unerreichbar. Drawer-Inhalt nicht scrollbar |
| Screenshot dark, lange Liste | 19:00 | B, Safari-Tab, Dark Mode | `ZahlungDrawer` („Zahlung für Tisch 01", 9+ Positionen, 65,10 €), Eingaben leer | Positionsliste intern gescrollt (oberste Zeile halbiert; sichtbare Positionen summieren nur 40,10 € von 65,10 €), Gesamt/Eingaben/Kommentar sichtbar, Footer mit „Kassieren"/„Abbrechen" komplett außerhalb des Vollbild-Screenshots |
| Screenshot light (2× identisch) | 18:55 | A, PWA (kein Browser-Chrome) | `BestellungDrawer` („Bestellung für Tisch 08", 6 Positionen, 45,10 €) | Drawer vollständig und korrekt, Buttons sichtbar. Achtung: zeigt exakt denselben Warenkorb wie das Flicker-Video — mutmaßlich eine Momentaufnahme der „guten" Phase **derselben** Sitzung, als unabhängiger „Sollzustand" nur eingeschränkt tauglich |
| Screenshot light (2× identisch) | 19:01 | A, PWA | Bestellen-Tab, Produktliste (Süßgetränke/Essen) | `StickyActionBar` („0 · Bestellung überprüfen · 0,00 €") **und** Tab-Leiste (Bestellen/Kassieren/Historie) stehen mitten im Viewport (~45 % der Höhe über ihrem Soll-Platz); Produktkarten laufen ober- und unterhalb weiter. Kein Screenshot-Beschnitt (Vollbild-Format wie 18:55), keine Scroll-Animation (derselbe Zustand existiert im Video 1,25 s statisch). Die Halbtransparenz der Bar ist der reguläre `disabled`-Zustand bei 0 Positionen (`opacity-50`), kein weiterer Defekt |
| Screenrecording | 18:57 | A, PWA | `BestellungDrawer` Tisch 08, 6 Positionen, 45,10 € | Bistabiler Flicker + Phase mit verschwundenem Drawer + fehlgeleitete Touch-Interpretation; Details unten |

### Frame-Analyse des Videos (1068 Frames @ 60 fps)

Drei Renderzustände wechseln über ~16 s:

- **Zustand A (korrekt, ~715 Frames):** Bottom-Sheet unten, Hintergrund abgedunkelt/geblurrt,
  Hintergrundseite steht am Seitenanfang (Tisch-Header sichtbar).
- **Zustand B (~18 Episoden, je 33 ms–1,4 s):** Kein Overlay; der Drawer-Inhalt ist um einen
  **konstanten Betrag (~40 % der Viewporthöhe, ≈ 324 px)** nach oben versetzt (Handle und
  Header oberhalb des Viewports abgeschnitten, Liste beginnt direkt unter der Statusleiste).
  Unterhalb der Sheet-Unterkante ist die Seite ungedimmt sichtbar — und zwar der **Anfang**
  der Produktliste (kollabierte Gruppen „Mineralwasser"/„Süßgetränke"): Die Seite dahinter
  steht wie in A am Seitenanfang. Die Wechsel A↔B erfolgen in einem einzigen Frame, ohne
  Zwischenpositionen, an pixelidentische Positionen — keine Drag-Geste, keine Animation. In
  keinem der 1068 Frames ist eine Tastatur (oder deren Animation) sichtbar.
- **Zustand C (Frames 761–835, 1,25 s):** Der Drawer verschwindet **vollständig**; sichtbar
  ist die Seite mit mitten im Viewport klebender StickyActionBar („6 · Bestellung überprüfen ·
  45,10 €") und Tab-Leiste — exakt das Fehlbild der 19:01-Screenshots, hier also **während**
  der Sitzung. Bemerkenswert: Oberhalb der Leisten ist die echte Scrollposition (Essen/Falafel)
  zu sehen, unterhalb der Listen-Anfang (Getränke) — **zwei Scroll-Offsets im selben Frame**,
  der direkteste visuelle Beleg für einen Compositing-/Repaint-Fehler. Danach (836–858) öffnet
  der Drawer mit regulärer Animation wieder — mutmaßlich ein erfolgreicher Tap auf die
  versetzte Bar.
- **Fehlgeleitete Touch-Interpretation (Frames ~250, 287–355):** Die iOS-Textlupe erscheint
  über dem Button-/Kommentarbereich, gefolgt von ~1,1 s blaugrauem Textauswahl-Highlight über
  dem gesamten Drawer-Inhalt — ein Long-Press wurde als **Textauswahl** statt als Klick
  interpretiert. Direkter Bildbeleg dafür, dass Touch-Eingaben den Button nicht als
  Klickziel erreichen.

Flicker-Frequenz: im Mittel ~1 Zustandswechselpaar pro Sekunde, in Schüben bis ~4 Wechsel in
0,5 s. In keinem Frame erscheint ein Spinner oder Toast; bis zum Aufnahmeende (Control Center)
bleibt die Bestellung (45,10 €) unabgeschickt.

## Bug 1 — vaul-Drawer bricht in der iOS-Standalone-PWA: Flicker, versetzte `fixed`-Elemente, wirkungslose Taps

**Notizen:** „Bei großer Bestellung flickert Hintergrund und es passiert nichts … kleine
Bestellungen (1–3 Positionen) klappen problemlos", „iOS Flicker bei installierter PWA. Im
Browser (Chrome auf iOS) scheint es zu funktionieren", `praxistest-2026-07-09.md` Zeilen 3 + 5.

**Betroffen:** alle Drawer (`BestellungDrawer`, `ZahlungDrawer`, Historie-/Storno-Drawer) —
gemeinsame Basis `frontend/src/components/ui/drawer.tsx` auf **vaul 1.1.2** (aktuellste
existierende Version, Stand 2026-07); als Folgeschaden `StickyActionBar` und die fixierte
Tab-Leiste in `TablePage.tsx`.

**Mechanismus — am vaul-Quellcode (Tag v1.1.2) verifiziert:**

1. vaul kennt zwei Scroll-Sperren. Der Body-Lock via `position: fixed` + negativem `top`
   (`use-position-fixed.ts`) wird im Standalone-Modus **bewusst übersprungen**:
   `window.matchMedia('(display-mode: standalone)').matches` → `setPositionFixed()` entfällt
   (Guard seit v0.9.1, Changelog: „Fix position fixed causing layout shifts on standalone
   sites (pwa)"). In der installierten PWA ist also **nur** die zweite Sperre aktiv:
   `usePreventScroll` → `preventScrollMobileSafari()` (auf allen iOS-Browsern), die beim
   Öffnen `window.scrollTo(0, 0)` ausführt und per Scroll-Listener auf `(0, 0)` festhält —
   die im Quellcode vorgesehene kompensierende negative Body-Margin ist **auskommentiert**.
   Der Sprung der Seite an den Seitenanfang ist damit unkompensiert.
2. Das erklärt die Modus-Abhängigkeit: Im Safari-**Tab** läuft zusätzlich der
   `position:fixed`-Body-Lock, der die Scrollposition konserviert und den `scrollTo(0,0)`
   neutralisiert (das Dokument ist nicht mehr scrollbar) — der Drawer funktioniert. In der
   **Standalone-PWA** fehlt genau dieser Lock, das Dokument springt real an den Anfang, und
   WebKit gerät beim Neuzeichnen der `position: fixed`-Ebenen (Drawer, Overlay,
   StickyActionBar, Tab-Leiste) in den beobachteten fehlerhaften Zustand.
3. Die Fehlbilder im Material: Drawer-Inhalt um konstant ~40 % Viewporthöhe nach oben
   versetzt bei unsichtbarem Overlay (Zustand B); fixierte Leisten mitten im Viewport, mit
   zwei verschiedenen Scroll-Offsets im selben Frame (Zustand C und Screenshots 19:01);
   Touch-Eingaben erreichen den Button nicht — ein Long-Press wird als Textauswahl
   interpretiert (Lupe + Auswahl-Highlight im Video), Taps verpuffen wirkungslos. Dass das
   Overlay in iOS-Standalone Pointer-Events nicht zuverlässig isoliert, deckt sich mit
   shadcn-ui#8507.
4. **Einordnung der Belegqualität:** Der unkompensierte `scrollTo(0,0)`-Lock (Punkt 1) und
   die Fehlbilder (Punkt 3) sind quellcode- bzw. bildverifiziert. Die genaue
   WebKit-interne Ursache des Repaint-/Hit-Test-Versagens ist dagegen **Hypothese**: WebKit
   Bugzilla dokumentiert `position:fixed`-Instabilitäten bei Viewport-Änderungen
   ([Bug 297779](https://bugs.webkit.org/show_bug.cgi?id=297779),
   [Bug 254926](https://bugs.webkit.org/show_bug.cgi?id=254926)), allerdings mit kleineren
   Versätzen (~10–24 px) — kein Primärbeleg deckt exakt das hier beobachtete Ausmaß. Der
   gemessene Versatz (≈ 324 px) entspricht auch **nicht** dem Scroll-Offset der Seite
   (≈ 880 px); die in Fassung 1 behauptete Entsprechung ist widerlegt.

**Warum nur große Bestellungen:** Für eine große Bestellung scrollt man weit durch die
Produktliste; der Drawer öffnet „below the fold" → der unkompensierte `scrollTo(0,0)` springt
real und WebKit muss die fixed-Ebenen über einen großen Scroll-Sprung hinweg neu zeichnen.
Bei 1–3 Positionen steht die Seite nahe am Anfang, `scrollTo(0,0)` ist ein No-Op, nichts
springt. Exakt diese Abhängigkeit beschreibt das offene Upstream-Issue
[vaul#505 „Drawer is unusable on a standalone installed PWA"](https://github.com/emilkowalski/vaul/issues/505)
(u. a. „body scrolls back to top", Interaktionen schließen/verfehlen den Drawer,
funktioniert am Seitenanfang).

**Warum der Browser-Tab funktioniert:** Siehe Mechanismus Punkt 2. Chrome auf iOS lief beim
Test faktisch ebenfalls auf WebKit (die DMA-Öffnung für alternative Engines ist bislang ohne
ausgelieferten Blink-Chrome); vaul behandelt ohnehin alle iOS-Browser wie Safari
(`isSafari()`-UA-Check trifft auch CriOS).

**Upstream-Referenzen:** Tragfähig sind
[vaul#505](https://github.com/emilkowalski/vaul/issues/505) (offen, kein Fix) und
[shadcn-ui#8507](https://github.com/shadcn-ui/ui/issues/8507) (Pointer-Durchlässigkeit in
iOS-PWA; dokumentierter Workaround: Radix-basiertes `Sheet` statt vaul-Drawer). Verwandt,
aber schwächer: [vaul#620](https://github.com/emilkowalski/vaul/issues/620) (offen,
Input-Sprünge nur in PWA), [vaul#365](https://github.com/emilkowalski/vaul/issues/365)
(geschlossen; erwähnt Page-Flash + Scroll-Verschiebung beim Öffnen),
[vaul#216](https://github.com/emilkowalski/vaul/issues/216) (geschlossen, alt, v0.8.0),
[vaul#536](https://github.com/emilkowalski/vaul/issues/536) (Edge/iOS-16-spezifisch).
**Ein vaul-Update ist kein Fix-Pfad:** 1.1.2 (2024-12) ist die letzte veröffentlichte
Version; npm `latest` = 1.1.2, keine neueren Releases.

**Abgrenzung:** Der eingebaute Workaround `data-vaul-no-drag` (Kommentar in `drawer.tsx`)
adressiert ein anderes iOS-Problem (Tap wird als Drag interpretiert) und greift hier nicht.

## Bug 2 — Drawer-Inhalt nicht scrollbar: Primär-Button gerät aus dem Viewport

**Notizen:** „Eingabe des Trinkgelds und des gegebenen Geldes → Kassenzahlung nicht
abschließbar", „Wenn die Liste zu lang wird, kann ich nicht mehr abkassieren."
Reproduziert **im normalen Safari-Tab** (beide dunklen Screenshots), unabhängig von Bug 1.

**Betroffen:** primär `ZahlungDrawer.tsx`, strukturell identisch `BestellungDrawer.tsx`.

**Mechanismus (code-verifiziert):** `DrawerContent` (`drawer.tsx`) ist `fixed bottom-0 …
flex flex-col` mit `max-h-[80vh]` und `overflow: visible` — **es gibt keinen Scrollcontainer
um den Gesamtinhalt**. Einziger scrollbarer Bereich ist die Positionsliste im `Receipt`
(`overflow-y-auto max-h-[40dvh]`); sie hängt aber in einem gewöhnlichen Block-Wrapper
(`<div className="mx-auto w-full max-w-sm">`, via `display: contents`-Zwischenschicht das
eigentliche Flex-Item) und ist damit vom Höhenbudget des `DrawerContent` entkoppelt — ihr
internes Scrollen schafft dem Footer keinen Platz. Übersteigt Header + Liste (bis 40 dvh) +
Gesamt + Eingabezeilen + Kommentar + Footer die 80 vh, bleibt die Box an `bottom: 0` auf
80 vh gekappt und der Überschuss fließt als sichtbarer Overflow **unterhalb** der
Viewport-Unterkante heraus: Der Footer mit „Kassieren" wird gerendert, liegt aber off-screen
und ist nicht erscrollbar. (Dass der Wrapper als Flex-Item mit `min-height: auto` nicht
schrumpft, ist dabei nachrangig — auch ein schrumpfender Wrapper ohne Scrollcontainer würde
den Inhalt unten herausquellen lassen.)

Zwei Auslöser:

- **(a) Lange Positionsliste** (Screenshot 19:00): Die Liste kappt zwar bei 40 dvh, aber der
  Rest sprengt die verfügbare Höhe trotzdem.
- **(b) Trinkgeld/Erhalten eingeben** (Screenshot 18:29): Die Zeilen „Rückgeld", „Trinkgeld"
  und der zweizeilige Hinweistext erscheinen per bedingtem Rendering **erst nach der Eingabe**
  (≈ 95–110 px) und schieben den Footer aus dem Bild — selbst bei nur 3 Positionen.

Verschärfend: `max-h-[80vh]` nutzt `vh`, das per Spezifikation dem **großen** Viewport
entspricht (`vh` ≡ `lvh`, [MDN](https://developer.mozilla.org/en-US/docs/Web/CSS/length)).
Im Safari-Tab mit eingeblendeter Bottom-Toolbar ist der real sichtbare Bereich kleiner.
Die iOS-Tastatur (öffnet erst bei Fokus — kein Autofokus im Code) verkleinert den sichtbaren
Bereich zusätzlich; die Screenshots belegen aber, dass der Footer bereits **ohne** offene
Tastatur unerreichbar ist.

## Bug 3 (sekundär) — kein wahrnehmbares Lade-/Pending-Feedback

**Notiz:** „Kein Loading State."

`useActionSubmit` liefert `loading` und die Submit-Buttons zeigen einen `Spinner` + `disabled`
— ein Loading-State **existiert** also. Er wurde im Test nie sichtbar, weil in Bug 1 der Tap
den Button gar nicht erreicht: kein Request, kein Spinner, kein Toast (im gesamten Video
bestätigt). Aus Nutzersicht ist „flickert und nichts passiert" damit nicht von einem
hängenden Request unterscheidbar. Kein eigener Defekt, aber ein Diagnose-/UX-Manko: ein
deutlicherer Pending-Zustand (z. B. Drawer blockieren) würde Fehlerbilder wie dieses sofort
unterscheidbar machen.

## Zuordnung Notizen → Bugs

| Testnotiz | Bug |
|---|---|
| Große Bestellung: Hintergrund flickert, nichts passiert; kleine Bestellungen ok | **Bug 1** (Auslöser = Öffnen „below the fold" nach langem Scrollen), Wahrnehmung „kein Loading" = **Bug 3** |
| Trinkgeld + Erhalten eingegeben → Zahlung nicht abschließbar | **Bug 2 (b)** |
| Viele Bestellungen / Liste zu lang → kann nicht abkassieren | **Bug 2 (a)** |
| iOS-Flicker nur in installierter PWA, Browser ok | **Bug 1** |
| „Generell Probleme bei installierter PWA auf iOS **und Android**" | iOS: Bug 1; für Android liegt kein Bildmaterial vor — vermutlich Bug 2 (geräteunabhängig) und/oder eigenes Fehlerbild, im nächsten Test gezielt reproduzieren |

## Empfohlene Fix-Richtungen (nicht umgesetzt, zur Diskussion)

1. **Bug 2 (Quick Win, framework-unabhängig):** Drawer-Layout umbauen — Mittelteil (Liste +
   Eingaben + Kommentar) in einen `overflow-y-auto`-Bereich mit `min-h-0` legen, Footer
   außerhalb des Scrollbereichs immer sichtbar halten; `max-h` auf `dvh` statt `vh` umstellen
   und `env(safe-area-inset-bottom)` einrechnen. Behebt beide Auslöser (a) und (b) auch im
   Browser-Tab.
2. **Bug 1 — primäre Option:** Für die kritischen Bestätigungs-Drawer auf ein Radix-basiertes
   `Sheet`/`Dialog` ohne vauls iOS-Scroll-Sperren ausweichen (in shadcn-ui#8507 als
   funktionierender Workaround dokumentiert). Ein vaul-Update existiert nicht (1.1.2 ist
   latest), vaul#505 ist unverändert offen.
   **Sekundär, falls vaul bleiben soll:** `disablePreventScroll={false}` testen (Achtung,
   invertierte Benennung: der Default `true` lässt `usePreventScroll` **aktiv**; `false`
   schaltet den `scrollTo(0,0)`-Lock ab). `repositionInputs={false}` wirkt über eine interne
   Kopplung ähnlich, verliert aber die Tastatur-Repositionierung. `noBodyStyles` ist in
   Standalone wirkungslos (der Body-Lock läuft dort ohnehin nicht). Jede Variante auf einem
   echten iPhone in der installierten PWA mit weit gescrollter Produktliste verifizieren.
3. **Bug 3:** Pending-Zustand im Drawer deutlicher machen (Inhalt blockieren/dimmen), damit
   „Tap kam nicht an" und „Request läuft" unterscheidbar sind.

## Audit-Vermerk (Fassung 1 → 2)

Drei unabhängige adversariale Prüfungen (Code-Mechanik, Upstream-/Plattform-Quellen,
forensische Bildanalyse aller 1068 Frames). Alle Endaussagen zu Bug 2 und Bug 3 sowie die
Kernbefunde zu Bug 1 (Flicker + versetzte fixed-Elemente nur in der Standalone-PWA, Taps
wirkungslos, Auslöser-Abhängigkeit vom Scrollzustand) wurden bestätigt. Korrigiert wurden:

- **Bug-1-Mechanismus:** Fassung 1 machte vauls `position:fixed`-Body-Lock verantwortlich —
  am Quellcode widerlegt: Der ist in Standalone seit vaul v0.9.1 deaktiviert; aktiv ist der
  unkompensierte `scrollTo(0,0)`-Lock aus `usePreventScroll`. Die Fix-Empfehlung wurde
  entsprechend umgestellt (Radix-Ausweichen primär; vaul-Props mit korrigierten Namen).
- **Scroll-Offset-Behauptung:** Verschiebung in Zustand B (≈ 324 px) ≠ Scroll-Offset
  (≈ 880 px); Hintergrund in B steht am Listen-Anfang, nicht „weit gescrollt". Korrigiert.
- **Flicker-Frequenz:** statt „mehrmals pro Sekunde" real Ø ~1 Wechsel/s mit Bursts.
- **Ergänzt:** Zustand C (Drawer 1,25 s komplett weg, Leisten mitten im Viewport — das
  19:01-Fehlbild schon während der Sitzung; zwei Scroll-Offsets im selben Frame),
  Textlupe/Auswahl-Highlight als direkter Hit-Test-Beleg, Duplikat-Screenshots,
  Zwei-Geräte-Zuordnung, Pixel-Verifikation der grünen Buttonkante (18:29), Einordnung der
  WebKit-Belege (Bugzilla 297779/254926) als Hypothesen-Stütze, Issue-Status (#216/#365
  geschlossen, #536 nur schwach relevant).
- **Bug-2-Begründung präzisiert:** tragende Ursache ist der fehlende Scrollcontainer +
  `overflow: visible` am gekappten fixed-Container (nicht `min-height: auto`).
