# Feldfeedback

Wünsche, Fehlerberichte und Fragen aus den Mails und GitHub-Issues der Vereine, anonymisiert und gezählt. Keine Namen, Orte, Termine oder Zitate: Die Zuordnung zu Vereinen liegt nur im gitignorierten Register `.crm/`.

`make crm` erzeugt diese Datei aus dem Register; Änderungen von Hand gehen beim nächsten Lauf verloren. „Vereine" zählt die Vereine und Einzelpersonen, die einen Punkt unabhängig voneinander genannt haben.

## Lage

- 28 Organisationen haben die Nutzungsvereinbarung geschlossen, 1 davon zurückgezogen; 27 Vereinbarungen gelten.
- Die geltenden Annahmesätze nennen die TERMS-Fassungen vom 3. Juni 2026 (2), 14. Juli 2026 (14) und 7. September 2026 (11).
- 7 weitere Organisationen und 2 Einzelpersonen haben angefragt, ohne eine Vereinbarung zu schließen. 1 Absender hat seine Mail bestritten.
- Vereine aus dem Ausland: 3 aus Österreich und 2 aus der Schweiz.
- 3 Vereine berichten vom Einsatz am Fest.
- Alle Einsatzberichte sind positiv.
- Einer der berichtenden Vereine lief zwei Festtage mit über 1600 Tischbestellungen stabil.
- Über GitHub kamen 2 Anfragen: ein Verein aus der Schweiz und ein Nutzer ohne Vereinsangabe.
- Ein weiterer Verein richtete jotti samt Bondrucker in 2 Stunden ein.
- Gemeldete Störungen: ein still hängender Bondrucker, WLAN mit mehreren Access Points und unlesbare Umlaute auf einem Epson-Drucker.
- Als Wechselgrund nennen 2 Vereine teure Miet-Kassensysteme und 1 Verein eine defekte alte Kasse.

## Umgesetzt

| Wunsch                                                  | Vereine | Umsetzung                                               |
| ------------------------------------------------------- | ------: | ------------------------------------------------------- |
| Erster Login dokumentiert                               |       2 | Leitfaden „Installation" und README                     |
| Reihenfolge von Produkten und Varianten festlegen       |       2 | S-01                                                    |
| Anfrageformular ohne Mailprogramm oder ohne Annahmesatz |       1 | Kopier-Fallback und Annahmesatz im Formular der Website |
| Ausgabe-Schritt streichen                               |       1 | D01                                                     |
| Bedienung ohne lokales WLAN, Handys über Mobilfunk      |       1 | Leitfaden „Self-Hosting"                                |
| Direktverkauf ohne Tisch                                |       1 | K-24                                                    |
| Einzelbon pro Stück                                     |       1 | K-25 (Bonmodus am Abholbon)                             |
| Kategorie-Auswahl über der Produktliste                 |       1 | D10 (Kategorie-Pills)                                   |
| Lange Produktnamen am Handy umbrechen                   |       1 | Q-01                                                    |
| Liste bewährter Bondrucker                              |       1 | Leitfaden „Installation"                                |
| Neuinstallation scheitert (Migration, pnpm 11)          |       1 | behoben                                                 |
| README nannte nicht vorhandene Funktionen               |       1 | behoben                                                 |
| Rückgeldberechnung                                      |       1 | K-10                                                    |

## Abgelehnt

| Wunsch                                      | Vereine | Entscheidung                                                                       |
| ------------------------------------------- | ------: | ---------------------------------------------------------------------------------- |
| Mehrtägiger Betrieb ganz ohne Internet      |       2 | Nicht-Ziel in [anforderungen.md](anforderungen.md)                                 |
| Beleg per E-Mail statt Bondrucker           |       1 | Nicht-Ziel in [anforderungen.md](anforderungen.md)                                 |
| Hardware-TSE gegen die laufenden TSE-Kosten |       1 | [D09](decisions.md)                                                                |
| Helferdeckel                                |       1 | Nicht-Ziel in [anforderungen.md](anforderungen.md)                                 |
| Höchstmenge je Produkt mit Verkaufsstopp    |       1 | Nicht-Ziel in [anforderungen.md](anforderungen.md)                                 |
| Laufende Bewirtung im Vereinsheim           |       1 | Nicht-Ziel in [anforderungen.md](anforderungen.md)                                 |
| Nutzung in Österreich                       |       1 | [D16](decisions.md): erlaubt, aber nur mit deutschem Fiskalrecht und deutscher TSE |
| Produktebene über der Variantenliste        |       1 | [D10](decisions.md)                                                                |
| Vereinslogo auf dem Bon                     |       1 | Nicht-Ziel in [anforderungen.md](anforderungen.md)                                 |

## Offen

| Wunsch oder Fehler                                         | Vereine | Stand                                                                                                                                             |
| ---------------------------------------------------------- | ------: | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| Gläserpfand: Rückgabe und Verrechnung                      |       3 | [backlog.md](backlog.md), [PRD](prds/prd-pfandrueckgabe.md)                                                                                       |
| Ausgabestationen unabhängig von Produktkategorien          |       1 | Piepton entfernt, „Kein Bon" im Direktverkauf; Modell im [PRD](prds/prd-stationsmodell.md), [#146](https://github.com/nicograef/jotti/issues/146) |
| Bestellkommentar erscheint auf allen Arbeitsbons           |       1 | [PRD](prds/prd-kommentar-je-position.md), [#145](https://github.com/nicograef/jotti/issues/145)                                                   |
| Bondrucker hängt still, Relay meldet nichts                |       1 | [backlog.md](backlog.md)                                                                                                                          |
| Installation ohne Docker oder gehosteter Betrieb           |       1 | [backlog.md](backlog.md), [PRD](prds/prd-windows-einzeldatei.md)                                                                                  |
| Letzter Admin hat sein Passwort vergessen                  |       1 | [backlog.md](backlog.md)                                                                                                                          |
| Rabatte und Gutschriften                                   |       1 | außerhalb von [PRD](prds/prd-pfandrueckgabe.md), [backlog.md](backlog.md)                                                                         |
| Tische mit eigenen offenen Bestellungen automatisch zeigen |       1 | [backlog.md](backlog.md)                                                                                                                          |
| Umlaute auf Epson-Bondruckern unlesbar                     |       1 | PC437 umgesetzt, Druckprobe offen; [#144](https://github.com/nicograef/jotti/issues/144)                                                          |
| WLAN mit mehreren Access Points verliert die Adresse       |       1 | [backlog.md](backlog.md)                                                                                                                          |

## Wiederkehrende Fragen

| Frage                                                        | Vereine | Antwort steht in                                                      |
| ------------------------------------------------------------ | ------: | --------------------------------------------------------------------- |
| Braucht jotti am Fest Internet?                              |       3 | Leitfaden „Häufige Fragen"                                            |
| Wo bekommen wir Zugangsdaten, Freigabe oder Lizenz?          |       3 | Leitfaden „Häufige Fragen", Website „Für Vereine"                     |
| Dürfen Vereine aus Österreich oder der Schweiz jotti nutzen? |       2 | Leitfaden „Häufige Fragen", [D16](decisions.md)                       |
| Nehmen wir die aktuelle Version oder warten wir auf 1.0?     |       2 | Leitfaden „Häufige Fragen"                                            |
| Darf ein Partnerverein unsere Installation mitnutzen?        |       1 | Leitfaden „Häufige Fragen"                                            |
| Gibt es eine Demo oder ein Testsystem?                       |       1 | Leitfaden „Häufige Fragen"                                            |
| Läuft jotti bei schwachem Netz stabil?                       |       1 | Leitfaden „Häufige Fragen"                                            |
| Was passiert bei Stromausfall?                               |       1 | Leitfaden „Häufige Fragen"                                            |
| Welche Bondrucker passen?                                    |       1 | Leitfaden „Installation"                                              |
| Welche Rechtsform gibt eine Ortsgruppe eines Verbands an?    |       1 | neue TERMS-Fassung in Prüfung, [backlog.md](backlog.md)               |
| Wie aktualisieren wir jotti?                                 |       1 | Leitfaden „Aktualisieren"                                             |
| Wie läuft ein mehrtägiges Fest?                              |       1 | Leitfaden „Der Veranstaltungstag"                                     |
| Wie verhält sich jotti unter hoher Last?                     |       1 | Leitfaden „Häufige Fragen"; Messung offen in [backlog.md](backlog.md) |
