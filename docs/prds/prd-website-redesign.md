# PRD: Website-Redesign (jotti.rocks)

## Problem Statement

Die Marketing-Website ist der erste Kontaktpunkt für Vereinsverantwortliche, die ein
Kassensystem für ihr Fest suchen. Die heutige Landing erfüllt ihren Zweck, bleibt aber
hinter dem Produkt zurück: Die Screenshots sind nach den Admin- und Service-Redesigns
veraltet, die Marke (Spektralfarben) taucht nur als Haarlinie auf, es gibt keine
Preis-Aussage als eigene Sektion, keine FAQ, keinen Dark-Mode-Schalter und keine
dedizierte Seite für die Nutzungsanfrage. Die Anfrage selbst versteckt sich hinter
einem unauffälligen E-Mail-Link. Impressum und Datenschutzerklärung fehlen ganz,
obwohl die Seite geschäftsmäßig betrieben wird und aktiv zur Kontaktaufnahme auffordert.

Für den Neuauftritt liegt ein High-Fidelity-Design-Handoff vor (interaktiver
HTML-Prototyp mit Tokens, Copy, Assets und Referenz-Screenshots). Er ist die
gestalterische Vorlage, aber nicht 1:1 übernehmbar: Er lädt Google Fonts (verstößt
gegen die Produktiv-CSP und die Selbst-Hosting-Linie der Seite), zeigt statt echter
App-Screenshots durchgehend UI-Nachbauten, enthält ein Anfrage-Formular ohne Versand,
verspricht im Download-Bereich ein natives Windows-Programm, das es noch nicht gibt,
und lässt Inhalte der heutigen Seite (Service-Angebot, Tracking-Freiheit,
Leitfaden-Link in der Navigation, Rechtsseiten) weg.

## Solution

Die Marketing-Website wird nach dem Design-Handoff neu aufgebaut, innerhalb der
bestehenden Astro/Starlight-Codebasis und der bestehenden strikten CSP. Der Handoff
gibt Layout, Tokens, Typografie, Interaktionen und Copy vor; die folgenden bewussten
Abweichungen wurden mit dem Betreiber entschieden:

- **Echte Screenshots statt Nachbauten:** UI-Nachbauten nur dort, wo Interaktion sie
  erfordert (die Live-Demo). Hero und Screenshot-Sektion zeigen echte, per Playwright
  neu aufgenommene App-Screenshots in Hell und Dunkel.
- **Kombinierte Screenshot-Sektion:** Das "Ein Design. Jedes Gerät."-Statement des
  Handoffs (Desktop-Admin plus zwei Phones) wird Aufmacher einer Sektion, die darunter
  eine kompakte Galerie weiterer App-Ansichten zeigt (Nachfolger der heutigen
  "Einblicke"-Galerie).
- **Anfrage ohne Backend:** Das Formular auf der neuen Seite "Für Vereine" erzeugt
  beim Absenden einen vorbefüllten E-Mail-Entwurf (mailto) aus den Feldwerten, mit
  ehrlichem Hinweis-State und sichtbarer E-Mail-Adresse als Fallback.
- **Ehrliche Download-Copy:** Der Download-Bereich beschreibt das reale Release
  (ZIP mit Starter, benötigt Docker Desktop, der Leitfaden führt durch) statt des
  im Handoff versprochenen nativen ".exe in 5 Minuten".
- **Erhaltene Inhalte:** Das Service-Angebot (bezahlte Unterstützung auf Anfrage)
  wird FAQ-Item und auf der "Für Vereine"-Seite erwähnt; "keine Werbung, kein
  Tracking" wird Punkt in der Preis-Karte; die Navigation behält einen prominenten
  Leitfaden-Link zur Doku.
- **Rechtsseiten:** Impressum und Datenschutzerklärung kommen als statische Seiten
  hinzu und werden im Footer verlinkt.
- **Doku zieht mit:** Das Starlight-Theme der Doku übernimmt die neuen Fonts und
  Farbtokens, damit Landing und Doku eine Marke bleiben.

Der Spektral-Einsatz des Handoffs (animierter Textverlauf im H1, sechs Akzentfarben,
weiche Blobs, Scroll-Reveal) wird vollständig übernommen und löst die bisherige
Festlegung "Spektrum nur dekorativ" ab. Hell- und Dunkel-Theme sind per Schalter
wechselbar und gelten konsistent über Landing und Doku.

## User Stories

1. Als Vereinsverantwortliche will ich im Hero auf einen Blick sehen, was jotti ist
   und dass es für Vereinsfeste gemacht und kostenlos ist, damit ich sofort weiß, ob
   sich Weiterlesen lohnt.
2. Als Vereinsverantwortliche will ich die sechs Funktionsbereiche in einem
   interaktiven Explorer antippen und je Bereich Details sehen, damit ich den
   Funktionsumfang ohne Doku-Lektüre verstehe.
3. Als Vereinsverantwortliche will ich in einer interaktiven Live-Demo selbst
   Produkte antippen und kassieren (oder der Auto-Demo zusehen), damit ich ein
   Gefühl für die Bedienung am Handy bekomme.
4. Als Vereinsverantwortliche will ich den Ablauf vom Einrichten bis zum Z-Bon in
   vier Schritten sehen, damit ich den Aufwand für unser Fest einschätzen kann.
5. Als Vereinsverantwortliche will ich ehrlich lesen, wofür jotti geeignet und
   nicht geeignet ist, damit ich keine Fehlentscheidung treffe.
6. Als Vereinsverantwortliche will ich echte App-Screenshots auf Desktop und Phone,
   in Hell und Dunkel, plus eine Galerie weiterer Ansichten sehen, damit ich dem
   gezeigten Produkt vertrauen kann.
7. Als Vereinsverantwortliche will ich verstehen, dass jotti dauerhaft kostenlos,
   werbefrei und trackingfrei ist und welche laufenden Fremdkosten (Cloud-TSE,
   optional Server) anfallen, damit ich das Budget realistisch plane.
8. Als Vereinsverantwortliche will ich die fiskalischen Bausteine (TSE, Belegausgabe,
   GoBD-Journal, DSFinV-K, Rollen, Onboarding) auf einen Blick sehen, damit ich der
   Rechtskonformität vertraue.
9. Als Vereinsverantwortliche will ich häufige Fragen (Kosten, TSE, Hardware, PWA,
   Installation, Zielgruppe, Quellcode, Unterstützung beim Einrichten) aufklappen
   können, damit meine Restzweifel beantwortet werden.
10. Als Vereinsverantwortliche will ich auf der Seite "Für Vereine" ein Formular
    ausfüllen, das einen vorbefüllten E-Mail-Entwurf öffnet (mit sichtbarer
    E-Mail-Adresse als Fallback), damit die Nutzungsanfrage ohne Hürde rausgeht.
11. Als technikaffiner Helfer will ich im Download-Bereich ehrlich lesen, was das
    Windows-Release ist und voraussetzt, und direkt zu Leitfaden und Quellcode
    kommen, damit ich realistisch loslegen kann.
12. Als Besucher will ich zwischen hellem und dunklem Design umschalten und meine
    Wahl gespeichert wissen, damit die Seite meiner Umgebung entspricht.
13. Als Doku-Leser will ich beim Wechsel zwischen Landing und Leitfaden dieselbe
    Schrift, Farbwelt und Theme-Wahl vorfinden, damit jotti.rocks wie aus einem
    Guss wirkt.
14. Als Besucher am Smartphone will ich eine funktionierende mobile Navigation und
    einspaltige Layouts, damit die Seite auf kleinen Screens vollständig nutzbar ist.
15. Als Besucher mit reduzierter Bewegungspräferenz will ich, dass alle Animationen
    (Verlaufs-Sheen, Blobs, Scroll-Reveal, Auto-Demo) neutralisiert sind, damit die
    Seite für mich angenehm bleibt.
16. Als Besucher will ich Impressum und Datenschutzerklärung im Footer finden,
    damit klar ist, wer die Seite betreibt und was mit meinen Daten passiert.
17. Als Betreiber will ich die Website-Screenshots mit einem Skript reproduzierbar
    gegen eine geseedete Demo-Instanz neu erzeugen, damit die Seite nach
    UI-Änderungen ohne Handarbeit aktuell bleibt.

## Implementation Decisions

- **Seitenstruktur:** Vier Routen: Startseite (Landing), "Für Vereine" (Anfrage),
  Impressum, Datenschutz. Die Doku bleibt unverändert unter ihrem Pfad. Der
  clientseitige View-Wechsel des Prototyps wird durch echte Routen ersetzt.
- **Sektionsfolge der Landing** (nach Handoff, mit den beschlossenen Abweichungen):
  Header/Nav, Hero, Funktionen (Explorer), Live-Demo, Ablauf, Für wen,
  Screenshots ("Jedes Gerät" plus Galerie), Preis, Sicherheit und Compliance, FAQ,
  Download-CTA, Footer.
- **Rendering-Ansatz:** Die Seite bleibt statisch generiert (Astro). Interaktive
  Komponenten (Theme-Toggle, mobile Navigation, Feature-Explorer, Live-Demo,
  FAQ-Accordion, Anfrage-Formular) werden React-Islands über die offizielle
  Astro-React-Integration; alle übrigen Sektionen bleiben statische
  Astro-Komponenten ohne Hydration. React und die Integration sind als neue
  Abhängigkeiten des Website-Pakets genehmigt.
- **CSP-Konformität als harte Anforderung:** Die Produktiv-CSP (unter anderem
  script-src 'self', font-src 'self') bleibt unverändert; das gesamte Frontend
  inklusive Island-Hydration und Theme-Initialisierung muss ohne Inline-Skripte
  und ohne externe Hosts funktionieren. Die Verifikation gegen die Produktiv-CSP
  ist Abnahmekriterium.
- **Fonts self-hosted:** Space Grotesk (Headings, Wortmarke) und Inter (Body, UI)
  ersetzen Montserrat, als selbst gehostete woff2-Dateien (bevorzugt Variable
  Fonts, Gewichte gemäß Handoff) mit Preload wie bisher. Keine Google-Fonts-Einbindung.
- **Token-Set:** Die Handoff-Tokens (Light- und Dark-Palette, Spektralfarben,
  Radien, Schatten, Typo-Skala) ersetzen die bisherigen Markenwerte im gemeinsamen
  Token-Set von Landing und Doku. Das Tailwind-Theme wird auf diese Tokens gemappt;
  handgeschriebenes CSS nur für Keyframes und Spezialfälle (Spektral-Verläufe,
  Scroll-Reveal).
- **Theme-Verhalten:** Hell/Dunkel über ein Attribut am Wurzelelement, initialisiert
  vor dem ersten Paint über ein externes Head-Skript (kein Aufblitzen des falschen
  Themes), Fallback auf die Systempräferenz. Persistenz nutzt denselben
  Speicher-Mechanismus wie der Theme-Schalter der Doku, damit die Wahl über Landing
  und Doku konsistent ist; der im Handoff genannte eigene localStorage-Key entfällt
  zugunsten dieses Mechanismus.
- **Spektral-Einsatz:** Wie im Handoff, inklusive animiertem H1-Verlauf, sechs
  Akzentfarben für Features, Ablauf-Schritte und Compliance-Karten, Spektral-Streifen
  auf Karten, weichen Blobs und Scroll-Reveal. prefers-reduced-motion neutralisiert
  alle Animationen. Die frühere Festlegung "Spektrum nur dekorativ, keine UI-Fläche"
  ist damit abgelöst; das Wort "Regenbogen" darf nirgends auf der Seite vorkommen.
- **Bildstrategie:** UI-Nachbau ausschließlich für die interaktive Live-Demo. Der
  Hero zeigt einen echten Screenshot der Bestellansicht in einem Telefon-Rahmen,
  theme-abhängig als helle oder dunkle Variante. Die Screenshot-Sektion kombiniert
  das "Jedes Gerät"-Statement (Desktop-Admin, Phone hell, Phone dunkel, echte
  Screenshots in Geräte-Rahmen) mit einer kompakten Galerie weiterer Ansichten,
  angelehnt an die heutige Auswahl (Bestellen, Zahlung, Stornierung, Direktverkauf,
  Tischübersicht, Produkte, Benutzer, Geldtransit, Auswertung). Alle Screenshots
  werden per Playwright-Skript gegen eine geseedete Demo-Instanz neu aufgenommen;
  das Skript ist Teil des Lieferumfangs, die veralteten Bilddateien werden ersetzt.
- **Live-Demo:** Verhalten wie im Handoff: Auto-Demo startet beim Scrollen in den
  Viewport, stoppt dauerhaft bei manueller Interaktion, Mengen-Stepper mit live
  mitwachsender Summe, Kassieren-Overlay mit Erfolgsanimation, Reset-Button. Die
  Demo-Logik (Warenkorb, Summenbildung, Skriptablauf) wird als reines, UI-freies
  Modul geschnitten, das die React-Insel nur rendert.
- **Anfrage-Formular:** Felder wie im Handoff (Verein/Organisation,
  Ansprechpartner:in, E-Mail, Rechtsform, Nachricht). Absenden validiert
  clientseitig und öffnet einen vorbefüllten E-Mail-Entwurf an die
  Betreiber-Adresse; der Bestätigungs-State sagt ehrlich, dass der Entwurf geöffnet
  wurde und noch abzusenden ist. Die E-Mail-Adresse steht zusätzlich sichtbar auf
  der Seite als Fallback. Der Aufbau der mailto-URL wird als reine Funktion
  geschnitten. Kein serverseitiger Versand, kein Spam-Schutz nötig.
- **Navigation:** Anker-Links wie im Handoff (Funktionen, Ablauf, Für wen,
  Sicherheit, FAQ) plus Leitfaden-Link zur Doku; CTA "Für Vereine" führt auf die
  Anfrage-Seite. Mobile Navigation als Burger-Menü. Der heutige Beta-Banner im
  Header entfällt; die Beta-Kommunikation läuft über das Hero-Badge und die
  Hinweis-Pill im Download-Bereich.
- **Download-CTA:** Layout wie im Handoff, Copy an die Realität angepasst:
  Windows-Release als ZIP mit Starter, benötigt Docker Desktop, der Leitfaden
  führt durch die Einrichtung; dazu Links auf Leitfaden und Quellcode. Kein
  ".exe"- und kein "5 Minuten"-Versprechen, solange das native Windows-Paket
  nicht existiert.
- **Erhaltene Inhalte:** FAQ um ein Item "Gibt es Unterstützung beim Einrichten?"
  (bezahlter Service auf Anfrage) ergänzt, das Angebot wird auch auf der "Für
  Vereine"-Seite erwähnt. Die Preis-Karte erhält den Punkt "Keine Werbung, kein
  Tracking".
- **Footer:** Spalten wie im Handoff (Produkt, Ressourcen, Rechtliches), ergänzt um
  Impressum und Datenschutz unter Rechtliches. Ressourcen-Links führen auf die
  Doku-Seiten und GitHub.
- **Rechtsseiten:** Impressum (Anbieterkennzeichnung des Betreibers) und
  Datenschutzerklärung (trackingfreie statische Seite ohne Cookies, funktionale
  localStorage-Theme-Speicherung, Verarbeitung von E-Mail-Anfragen). Kein
  Cookie-Banner, da keine einwilligungspflichtigen Techniken eingesetzt werden.
  Die Textinhalte liefert der Betreiber.
- **Logos und Icons:** Ausschließlich die Original-Logo-Assets aus dem Handoff,
  keine Nachbauten. UI-Icons aus Lucide, wie im Handoff zugeordnet; die
  Icon-Bedeutungen bleiben erhalten (Bestellung = Beleg, Zahlung = Geldbörse und
  ausdrücklich kein Kartenterminal, Direktverkauf = Einkaufstasche).
- **Meta und SEO:** Canonical-, OpenGraph- und Twitter-Meta wie heute je Route;
  das OG-Bild wird auf das neue Design aktualisiert. Doku-Theme (Starlight)
  übernimmt Fonts und Farbtokens aus dem gemeinsamen Token-Set, sonst keine
  Doku-Änderungen.
- **Barrierefreiheit:** FAQ-Accordion und Feature-Explorer mit korrekter
  Tastaturbedienung und ARIA-Semantik; Theme-Toggle mit zugänglichem Label;
  Farbkontraste der Token-Palette werden bei der Umsetzung geprüft.

## Testing Decisions

- Gute Tests prüfen externes Verhalten (Eingabe zu Ergebnis), keine
  Implementierungsdetails wie interne States oder DOM-Strukturen.
- **Unit-Tests (vitest):** für die zwei rein geschnittenen Logik-Module: Demo-Logik
  (Mengen ändern, Summenbildung, Ablauf des Auto-Demo-Skripts, Stopp bei manueller
  Interaktion, Reset) und mailto-Builder (Feldwerte zu korrekt encodierter
  mailto-URL, Pflichtfeld-Validierung). Prior Art: die bestehenden
  Link-Rewriter-Tests des Website-Pakets.
- **Build als Gate:** Der bestehende Check des Website-Pakets (Typprüfung plus
  Build) muss grün bleiben; die bestehenden Tests laufen weiter.
- **Abnahme:** Sichtprüfung durch den Betreiber anhand der neuen Seite in Hell und
  Dunkel, Desktop und Mobil; Verifikation, dass die Seite unter der Produktiv-CSP
  ohne Konsolen-Verstöße läuft. Keine automatisierten E2E- oder visuellen
  Regressionstests für die Marketing-Seite.

## Out of Scope

- Serverseitiger Formularversand, Form-Backend oder Spam-Schutz.
- Das native Windows-Programm ohne Docker (eigene PRD) und jede Copy, die es
  voraussetzt.
- Änderungen an App-Frontend, Backend, Reverse-Proxy, CSP oder Infrastruktur.
- Neue oder umgeschriebene Doku-Inhalte (die Doku erhält nur das neue Theme).
- Blog, Newsletter, Analytics, Cookie-Banner, Mehrsprachigkeit.
- Automatisierte visuelle Regressionstests der Website.
- Änderungen am Design der App selbst (die App bleibt beim bestehenden grünen
  Design; die Spektral-Erweiterung gilt nur für die Website).

## Further Notes

- Design-Quelle ist das Handoff-Bundle unter `docs/prds/design_handoff_jotti_website/`
  (HTML-Prototyp, README mit Tokens und Copy, Original-Assets, Referenz-Screenshots).
  Bei Detailfragen zu Abständen, Farben oder Verhalten gilt der Prototyp beziehungsweise
  dessen README; bei Widersprüchen zu dieser PRD gilt die PRD. Nach abgeschlossener
  Umsetzung wird das Bundle wie bei früheren Handoffs aus dem Repo entfernt (die
  Git-Historie bewahrt es).
- Die exakte Copy des Prototyps ist Vorlage; die in dieser PRD beschlossenen
  Copy-Abweichungen (Download-Bereich, Formular-Bestätigung, FAQ-Ergänzung,
  Preis-Karte) gehen vor.
- Die Ablösung der Festlegung "Spektrum nur dekorativ" wird bei der Umsetzung als
  kurzer ADR festgehalten, da sie eine dokumentierte Branding-Entscheidung umkehrt.
- Der veraltete Screenshot-Ordner auf Repo-Ebene ist zur sofortigen Löschung
  freigegeben (in dieser Session an der Sandbox-Berechtigung gescheitert); die
  Screenshot-Dateien im Website-Paket werden im Zuge der Umsetzung durch neu
  aufgenommene ersetzt.
- Browser-Unterstützung: Der Prototyp nutzt moderne CSS-Features (unter anderem
  color-mix und Scroll-getriebene Animationen). Scroll-Reveal wird als Progressive
  Enhancement umgesetzt: Ohne Unterstützung bleiben Sektionen schlicht sichtbar,
  Inhalte dürfen nie hinter fehlender Animations-Unterstützung verschwinden.
