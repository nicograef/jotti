# PRD: Ein einziger Leitfaden für Vereine (Doku-Konsolidierung)

## Problem Statement

Als Vereinsvorstand (Thomas) will ich jotti rechtssicher aufsetzen und betreiben,
finde die nötigen Informationen aber über vier getrennte Markdown-Leitfäden und
zwei zusätzliche Website-Seiten verteilt. Ich muss zwischen ihnen hin- und
herspringen, lese dieselben Dinge mehrfach in leicht abweichender Form und stoße
auf Widersprüche: Die README sagt, die TSE-Anbindung sei „in Entwicklung", der
Betreiber-Leitfaden sagt, sie sei vorhanden. An einer Stelle heißt der Weg
„Standard-Weg", auf der Website „Weg A". Mal ist von nginx die Rede, mal von
Caddy. Und fiskaly wird mal als bloßes Beispiel („z. B. fiskaly"), mal mit
genannten Alternativen (D-Trust) dargestellt, obwohl jotti ausschließlich mit
fiskaly funktioniert. Die Texte sind außerdem länger als nötig, mit großen Tabellen, einem
Glossar und KI-typischen Formulierungen, obwohl ich als ehrenamtlicher Helfer nur
schnell wissen will, was ich konkret tun muss.

Für den Maintainer entsteht das Spiegelproblem doppelt: Jede Änderung muss an bis
zu sechs Stellen nachgezogen werden, weshalb sie regelmäßig auseinanderlaufen.

## Solution

Es gibt genau einen kanonischen Leitfaden für Vereine als eine Markdown-Datei im
Repository. Er führt den einfachen Standardfall zuerst und vollständig durch und
schiebt Expertenwege und Sonderfälle ans Ende. Er ist in einfacher Sprache
geschrieben, kompakt, ohne große Tabellen, ohne Glossar, ohne KI-Slop, und er ist
inhaltlich korrekt und auf dem aktuellen Stand.

Alles Umliegende wird zu dünnen Verweisen auf diesen einen Leitfaden oder mit ihm
faktisch in Deckung gebracht:

- Die öffentliche Website bekommt eine einzige Vereins-Leitfaden-Seite, die den
  kanonischen Markdown-Leitfaden kompakt spiegelt; die heutigen zwei Seiten
  entfallen.
- README, Website-Startseite und Produktbeschreibung nennen denselben, korrekten
  Funktionsstatus und denselben Reverse Proxy.
- Hartcodierte Doku-Links im Code (Status-Seite, Admin-Bereich) zeigen auf den
  neuen Leitfaden.

Der Leitfaden wird damit zum tiefen Modul der Doku: eine Datei, eine URL, eine
Quelle der Wahrheit. Website, README und Code-Links sind nur noch Zeiger darauf.

## User Stories

1. Als Vorstand will ich einen einzigen Leitfaden öffnen und darin den ganzen Weg
   von „nichts" bis „läuft beim Fest" finden, ohne zwischen Dokumenten zu
   springen.
2. Als Vorstand will ich oben in 60 Sekunden verstehen, was jotti ist, welche drei
   Pflichten ich habe und was es kostet.
3. Als nicht-technischer Helfer will ich zuerst den einfachen Standardweg
   (Einzelgerät im WLAN, Doppelklick-Start, Handys per QR-Code verbinden), ohne
   von Server-, Domain- oder Kommandozeilen-Themen abgelenkt zu werden.
4. Als Vorstand will ich die TSE-Einrichtung als knappe Schrittfolge (fiskaly-Konto
   und API-Key, geführter Assistent, PUK/PIN sicher verwahren, von TEST auf LIVE),
   ohne in jeden Sonderfall eingeführt zu werden, bevor ich den Normalfall kenne.
5. Als Vorstand will ich meine gesetzlichen Pflichten knapp erklärt bekommen
   (ELSTER-Meldung mit der Seriennummer aus dem Admin-Bereich, Belege und
   Steuersätze, zehn Jahre aufbewahren mit Backups), ohne Paragrafentabelle.
6. Als Vorstand will ich am Ende eine kurze Checkliste (einmalig vor dem Fest,
   laufend), die ich abhaken kann.
7. Als technisch versierter Betreiber will ich den Experten-Weg (eigener Server,
   Domain, HTTPS, Update, Backup, Härten) weiter unten im selben Dokument finden,
   klar als optional erkennbar.
8. Als Betreiber will ich TSE-Sonderfälle (PIN verloren oder gesperrt per PUK
   zurücksetzen, vorhandene TSS übernehmen, manuelle Konfiguration, Test-Limit)
   weiter unten finden, ohne dass sie den Normalfall überladen.
9. Als Betreiber, dessen grüne Adresse im WLAN nicht lädt, will ich die
   DNS-Rebind-Hilfe als klar abgegrenzten Fehlersuche-Abschnitt finden, inklusive
   des Fallbacks, mit dem der Verkauf trotzdem sofort weiterläuft.
10. Als Vorstand will ich, dass jede Aussage stimmt: TSE-Anbindung und
    Belegausgabe sind vorhanden, DSFinV-K-Export ist in Entwicklung, eine
    Hash-Chain gibt es nicht.
11. Als Vorstand will ich keine widersprüchlichen Begriffe lesen (ein Name pro
    Weg, ein Reverse Proxy), egal ob in Leitfaden, README oder auf der Website.
12. Als Leser will ich einfache, kurze Sätze ohne KI-Floskeln, ohne übermäßiges
    Fettdruck-Markup und ohne Gedankenstrich-Manier.
13. Als Besucher der Website will ich genau eine Vereins-Leitfaden-Seite finden,
    die dasselbe sagt wie der Repo-Leitfaden, nur fürs Web aufbereitet, ohne
    Glossar.
14. Als Besucher der Startseite will ich einen Funktionsstatus sehen, der mit dem
    Leitfaden und der README übereinstimmt.
15. Als Entwickler, der das Repo öffnet, will ich eine README, die jotti korrekt
    beschreibt (Status, Tech-Stack, ein Verweis auf den einen Leitfaden) und mich
    nicht mit veralteten „in Entwicklung"-Markierungen in die Irre führt.
16. Als Admin im Bereich „Finanzamt" will ich, dass der dort verlinkte Leitfaden
    auf das aktuelle Dokument zeigt und nicht ins Leere oder auf eine alte Datei.
17. Als Helfer, der die Status-Seite der Kasse öffnet, will ich, dass der Link zur
    DNS-Rebind-Hilfe auf den richtigen Abschnitt des neuen Leitfadens führt.
18. Als Maintainer will ich nur noch eine Markdown-Datei pflegen müssen; alle
    anderen Stellen verweisen darauf oder spiegeln sie bewusst und nachvollziehbar.
19. Als Maintainer will ich, dass kein interner Link nach der Umstellung tot ist
    (weder in Docs noch im Code noch auf der Website).
20. Als beitragender Entwickler oder Agent will ich, dass die kanonischen
    Referenzdokumente (AGENTS.md, Produktbeschreibung) denselben Status und
    dieselbe Infrastruktur nennen wie der Leitfaden.
21. Als Vorstand, der den Leitfaden quer liest, will ich, dass er die
    verbindlichen Fachbegriffe verwendet (Kassensitzung, Tagesabschluss/Z-Bon,
    Servicekraft, Serviceleitung, Admin, Direktverkauf), damit er zur App passt.
22. Als Maintainer will ich den gesamten Ordner `docs/betrieb/` entfernt wissen,
    damit niemand versehentlich die veraltete Fassung liest oder verlinkt.
23. Als Vorstand will ich klar lesen, dass ich für die TSE ein fiskaly-Konto
    brauche, und nicht durch „z. B. fiskaly" oder genannte Alternativ-Anbieter den
    Eindruck bekommen, ich hätte die Wahl oder müsste Anbieter vergleichen.

## Implementation Decisions

### Der kanonische Leitfaden (das tiefe Modul)

- Der kanonische Leitfaden liegt unter `docs/leitfaden.md`. Er ersetzt die vier
  heutigen Dateien (`leitfaden-betreiber.md`, `leitfaden-hosting.md`,
  `leitfaden-tse-einrichtung.md`, `dns-rebind-schutz.md`); der gesamte Ordner
  `docs/betrieb/` wird anschließend gelöscht. Alle Referenzen (siehe unten) zeigen
  danach auf `docs/leitfaden.md`.
- Aufbau nach dem Prinzip „einfacher Standardfall zuerst, Expertenkram unten"
  (progressive disclosure), grob:
  1. In 60 Sekunden: Was ist jotti, welche drei Pflichten, was kostet es.
  2. Schnellstart Standardweg: Einzelgerät im WLAN, Doppelklick-Start, Handys per
     QR-Code, grünes Schloss als Normalfall, Fallback-Adresse als Auffanglösung.
  3. TSE einrichten: fiskaly-Konto und API-Key, geführter Assistent, PUK/PIN
     verwahren, TEST→LIVE inklusive grober Kostenangabe.
  4. Pflichten erfüllen: ELSTER-Meldung mit der Seriennummer aus dem
     Admin-Bereich, Belege und Steuersätze, zehn Jahre aufbewahren mit Backups.
  5. Checkliste: einmalig vor dem Fest, laufend.
  6. Ab hier optional/Experten: eigener Server (VPS) mit Domain, HTTPS, Update,
     Backup, Härten.
  7. TSE-Sonderfälle: PIN per PUK zurücksetzen, TSS übernehmen, manuelle
     Konfiguration, Test-Limit, Wiederaufnahme nach Abbruch.
  8. Fehlersuche: grüne Adresse lädt nicht (DNS-Rebind), Router-Hinweise,
     Fallback.
  9. Häufige Fragen.
- Die deduplizierte Fassung übernimmt den sachlichen Inhalt der Altdokumente,
  kürzt aber: doppelte Erklärungen verschmelzen, seltene Detailfälle werden knapp.

### Sprache und Stil

- Einfache Sprache, kurze Sätze, aktive Formulierung, Vereins-Ansprache („ihr").
- Keine großen Tabellen (kleine, wirklich nötige Vergleiche dürfen als knappe
  Liste bleiben), kein Glossar, keine unnötigen Details.
- Kein KI-Slop: keine Gedankenstrich-Manier, kein liberales Fettdruck-Markup,
  keine Floskeln. Run-in-Labels (kurzer fetter Satzanfang) sind erlaubt.
  Verifikation über einen Wortstrom-Diff gegen den Altbestand.
- Verbindliche Fachbegriffe gemäß der Terminologie-Referenz verwenden, damit der
  Leitfaden zur App-Sprache passt.

### Korrektheits-Baseline (verbindlicher Faktenstand)

Alle überarbeiteten Texte nennen denselben Stand:

- TSE-Anbindung: vorhanden. Belegausgabe: vorhanden.
- DSFinV-K-Export: in Entwicklung, Zielformat v2.5.
- Hash-Chain: existiert nicht und ist nicht geplant; jede Erwähnung entfällt.
- TSE-Anbieter: fiskaly ist der einzige unterstützte Anbieter (nur dieser Adapter
  ist implementiert). fiskaly wird nicht mehr als Beispiel oder Empfehlung
  präsentiert, sondern als der Anbieter. Kanonische Formulierung: „Cloud-TSE von
  fiskaly". Beispiel-Wording („z. B. fiskaly") und Alternativ-Anbieter (etwa
  D-Trust) entfallen. Der Verein schließt weiterhin selbst einen fiskaly-Vertrag
  ab und hinterlegt seine eigenen API-Schlüssel (Bring Your Own fiskaly-Konto).
- Reverse Proxy für den Betrieb (lokal und Produktion): Caddy. nginx wird nur im
  `jotti.rocks`-Demo-Stack verwendet und nicht mehr als Betriebs-Proxy genannt.
- Ein Name pro Hosting-Weg, durchgängig identisch in Leitfaden und Website.

### Website-Spiegel

- Die zwei heutigen Seiten (`leitfaden-fuer-vereine`, `jotti-selbst-betreiben`)
  werden zu einer einzigen Vereins-Leitfaden-Seite zusammengeführt, die den
  Markdown-Leitfaden kompakt spiegelt. Das Glossar entfällt. Die Wegbenennung wird
  an den Leitfaden angeglichen.
- Die Startseite (`index.html`) wird faktisch angeglichen (Status, Reverse Proxy,
  keine Hash-Chain).
- Navigation, mobile Navigation, Footer, `sitemap.xml` und interne Links werden an
  die neue Seitenstruktur angepasst; die wegfallenden URLs werden sauber behandelt
  (Weiterleitung oder Entfernen, je nachdem was die Hosting-Konfiguration erlaubt).
- Markdown bleibt kanonisch; die Website ist die bewusste, von Hand gepflegte
  Spiegelung. Es wird keine Build-Pipeline Markdown→HTML eingeführt.

### README und kanonische Referenzen

- README wird neu geschrieben, bleibt aber entwickler- und GitHub-orientiert:
  korrekter Funktionsstatus, korrekter Tech-Stack (Caddy), Schnellstart, ein
  Verweis auf den einen Leitfaden, Lizenz. Veraltete „in Entwicklung"-Markierungen
  und die Hash-Chain werden entfernt; DSFinV-K-Version auf v2.5 korrigiert.
- `produktbeschreibung.md` wird faktisch angeglichen (Status, Reverse Proxy),
  bleibt ansonsten die interne Produktidentität.
- `AGENTS.md` und enthaltene Tech-Tabellen: nginx→Caddy korrigieren (Betriebspfad).

### Code- und Querverweis-Repointing

- Status-Seite (Go): die hartcodierte DNS-Rebind-URL zeigt auf den
  Fehlersuche-Abschnitt von `docs/leitfaden.md` (URL plus Anker).
- Frontend Admin „Finanzamt": der Eintrag „Betreiber-Leitfaden" zeigt auf
  `docs/leitfaden.md`; der Compliance-Link bleibt.
- `compliance.md`: die Verweise auf Betreiber- und TSE-Leitfaden zeigen auf
  `docs/leitfaden.md` (Datei plus Anker). Zusätzlich wird an dieser Stelle die
  Anbieter-Aussage „fiskaly oder D-Trust" auf „fiskaly" korrigiert (einzeilige
  Faktenkorrektur, siehe Out of Scope). compliance.md wird ansonsten nicht
  überarbeitet.
- Repo-weiter Link-Sweep: keine verbleibenden Verweise auf `docs/betrieb/` oder
  die vier alten Dateinamen in Docs, Code oder Website.
- Der gesamte Ordner `docs/betrieb/` wird gelöscht (kein paralleler Altbestand).

## Testing Decisions

Dies ist überwiegend eine Doku- und Content-Aufgabe; „Tests" heißt hier
Konsistenz- und Verknüpfungsprüfung statt Unit-Tests. Ein guter Test prüft
beobachtbares Außenverhalten, nicht Formulierungen:

- Link-Integrität: Nach der Umstellung existiert kein toter interner Link und kein
  Verweis mehr auf `docs/betrieb/` oder die vier alten Dateinamen (Docs, Code,
  Website). Prüfung per Repo-weitem Suchlauf; Teil der Definition of Done.
- Faktenabgleich: TSE/Beleg = vorhanden, DSFinV-K = in Entwicklung, keine
  Hash-Chain, Caddy als Betriebs-Proxy. Gegen die Anforderungs-Statusliste
  abgeglichen.
- Bestehende automatisierte Tests, die geänderte Werte festschreiben, werden
  mitgezogen: Falls der Test der Status-Seite die DNS-Rebind-URL als Konstante
  prüft, wird er auf die neue URL aktualisiert (Prior Art: der vorhandene
  Status-Seiten-Test im Reverse-Proxy-Modul).
- Stil-Gegenprobe: Wortstrom-Diff des neuen Leitfadens gegen den Altbestand, um
  KI-Slop (Gedankenstrich-Manier, liberales Bold, Floskeln) sichtbar zu machen.
- Keine neuen Unit-Tests für reine Inhalte. Die Frontend-Verlinkungssektion wird
  nicht eigens betestet, wenn nur eine URL-Konstante wechselt.

## Out of Scope

- `compliance.md` und `steuerrecht.md` werden nicht überarbeitet; sie bleiben die
  juristisch/steuerliche Tiefen-Referenz hinter dem Leitfaden. Einzige Ausnahme:
  in `compliance.md` werden die Links auf den neuen Leitfaden umgebogen und die
  eine falsche Anbieter-Aussage („fiskaly oder D-Trust") auf „fiskaly" korrigiert.
  Das ist eine punktuelle Faktenkorrektur, keine Überarbeitung des Dokuments. Die
  allgemeine Erklärung der TSE-Bauformen (Hardware-TSE-Hersteller wie Swissbit,
  Epson) bleibt unangetastet, weil sie nicht behauptet, jotti unterstütze sie.
- Keine eigene Anleitung für Servicekräfte/Helfer. Der Leitfaden bleibt auf
  Admin/Betreiber fokussiert; die App gilt als selbsterklärend.
- Keine Build-Pipeline, die die Website aus Markdown generiert.
- Kein Redesign der Website (Layout, Styling); nur Inhalt, Struktur und
  Konsistenz.
- Keine inhaltliche Erweiterung über den Altbestand hinaus (z. B. keine neuen
  Betriebsthemen); Ziel ist Konsolidierung, Korrektur und Kürzung.
- `handbuch.md` (Entwickler-Architektur) und `anforderungen.md` bleiben unberührt.

## Further Notes

- Der bewusst gewählte Bruch mit „eine Quelle für Repo und Web automatisch": Die
  Website wird von Hand gespiegelt. Das ist Absicht (kein Generator), erzeugt aber
  eine Pflegepflicht; deshalb ist die Markdown-Datei eindeutig als kanonisch
  markiert und die Website verweist sichtbar auf sie.
- Aktive Pre-Release-Phase: Es gibt keine produktiven Instanzen; alte URLs dürfen
  ohne Migrationsrücksicht entfernt werden. Für die öffentliche Website ist
  dennoch eine schlichte Weiterleitung der zwei alten Pfade wünschenswert, falls
  bereits verlinkt/indiziert.
- Reihenfolge der Umsetzung: erst `docs/leitfaden.md` schreiben (er ist die
  Wahrheit), dann README/Produktbeschreibung/AGENTS angleichen, dann Website
  spiegeln, zuletzt Code-Links und Querverweise umbiegen und den Ordner
  `docs/betrieb/` löschen. So gibt es zu keinem Zeitpunkt einen Link ins Leere.
- Die in der Analyse gefundenen konkreten Widersprüche dienen als Mindest-Prüfliste
  für die Korrektheit: README-Status vs. Anforderungs-Status, nginx vs. Caddy,
  DSFinV-K v2.4 vs. v2.5, Hash-Chain-Reste, „Weg A/B" vs. „Standard/Experte",
  fiskaly als Beispiel/„z. B." vs. fiskaly als der Anbieter.
