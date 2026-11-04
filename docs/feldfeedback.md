# Feldfeedback

Wünsche, Fehlerberichte und Fragen aus den Mails und GitHub-Issues der Vereine, anonymisiert und gezählt. Keine Namen, Orte, Termine oder Zitate: Die Zuordnung zu Vereinen und die Original-Mails liegen nur im gitignorierten Ordner `.mails/`.

„Vereine" zählt die Vereine und Einzelpersonen, die einen Punkt unabhängig voneinander genannt haben.

## Lage

- 28 Organisationen haben die Nutzungsvereinbarung geschlossen, 1 hat zurückgezogen; 27 Vereinbarungen gelten.
- Die Annahmesätze nennen die TERMS-Fassungen vom 3. Juni (2), 14. Juli (15) und 7. September 2026 (11).
- 6 weitere Organisationen und 1 Privatperson haben nur angefragt. Über GitHub kamen 2 Anfragen: ein Verein aus der Schweiz und ein unbekannter Nutzer.
- 3 Organisationen kommen aus Österreich, eine davon hat zurückgezogen. 2 kommen aus der Schweiz.
- 3 Vereine berichten vom Einsatz am Fest, alle positiv. Einer lief zwei Festtage mit über 1600 Tischbestellungen stabil.
- Ein weiterer Verein richtete jotti samt Bondrucker in 2 Stunden ein.
- Gemeldete Störungen: ein still hängender Bondrucker, WLAN mit mehreren Access Points und unlesbare Umlaute auf einem Epson-Drucker.
- Als Wechselgrund nennen 2 Vereine teure Miet-Kassensysteme und 1 Verein eine defekte alte Kasse.

## Umgesetzt

| Wunsch                                                  | Vereine | Umsetzung                                               |
| ------------------------------------------------------- | ------: | ------------------------------------------------------- |
| Direktverkauf ohne Tisch                                |       1 | K-24                                                    |
| Rückgeldberechnung                                      |       1 | K-10                                                    |
| Ausgabe-Schritt streichen                               |       1 | D01                                                     |
| Kategorie-Auswahl über der Produktliste                 |       1 | D10 (Kategorie-Pills)                                   |
| Lange Produktnamen am Handy umbrechen                   |       1 | Q-01                                                    |
| Reihenfolge von Produkten und Varianten festlegen       |       2 | S-01                                                    |
| Einzelbon pro Stück                                     |       1 | K-25 (Bonmodus am Abholbon)                             |
| Erster Login dokumentiert                               |       2 | Leitfaden „Installation" und README                     |
| Liste bewährter Bondrucker                              |       1 | Leitfaden „Installation"                                |
| Bedienung ohne lokales WLAN, Handys über Mobilfunk      |       1 | Leitfaden „Self-Hosting"                                |
| Anfrageformular ohne Mailprogramm oder ohne Annahmesatz |       2 | Kopier-Fallback und Annahmesatz im Formular der Website |
| Neuinstallation scheitert (Migration, pnpm 11)          |       1 | behoben                                                 |
| README nannte nicht vorhandene Funktionen               |       1 | behoben                                                 |

Die zweispaltige Theken-Ansicht ab Tablet ([D08](decisions.md)) stammt aus dem eigenen Frontend eines Vereins, nicht aus einem Wunsch.

## Abgelehnt

| Wunsch                                      | Vereine | Entscheidung                                                                       |
| ------------------------------------------- | ------: | ---------------------------------------------------------------------------------- |
| Vereinslogo auf dem Bon                     |       1 | Nicht-Ziel in [anforderungen.md](anforderungen.md)                                 |
| Beleg per E-Mail statt Bondrucker           |       1 | Nicht-Ziel in [anforderungen.md](anforderungen.md)                                 |
| Helferdeckel                                |       1 | Nicht-Ziel in [anforderungen.md](anforderungen.md)                                 |
| Hardware-TSE gegen die laufenden TSE-Kosten |       1 | [D09](decisions.md)                                                                |
| Produktebene über der Variantenliste        |       1 | [D10](decisions.md)                                                                |
| Nutzung in Österreich                       |       2 | [D16](decisions.md): erlaubt, aber nur mit deutschem Fiskalrecht und deutscher TSE |
| Mehrtägiger Betrieb ganz ohne Internet      |       2 | TSE braucht Internet; F-14 deckt Ausfälle, keinen Dauerzustand                     |

## Offen

| Wunsch oder Fehler                                         | Vereine | Stand                                                                           |
| ---------------------------------------------------------- | ------: | ------------------------------------------------------------------------------- |
| Gläserpfand: Rückgabe und Verrechnung                      |       3 | [backlog.md](backlog.md)                                                        |
| Rabatte und Gutschriften                                   |       1 | [backlog.md](backlog.md)                                                        |
| Höchstmenge je Produkt mit Verkaufsstopp                   |       1 | [backlog.md](backlog.md)                                                        |
| Laufende Bewirtung im Vereinsheim                          |       1 | [backlog.md](backlog.md)                                                        |
| Tische mit eigenen offenen Bestellungen automatisch zeigen |       1 | [backlog.md](backlog.md)                                                        |
| Letzter Admin hat sein Passwort vergessen                  |       1 | [backlog.md](backlog.md)                                                        |
| Bondrucker hängt still, Relay meldet nichts                |       1 | [backlog.md](backlog.md)                                                        |
| WLAN mit mehreren Access Points verliert die Adresse       |       1 | [backlog.md](backlog.md)                                                        |
| Installation ohne Docker oder gehosteter Betrieb           |       1 | [backlog.md](backlog.md)                                                        |
| Umlaute auf Epson-Bondruckern unlesbar                     |       1 | [backlog.md](backlog.md), [#144](https://github.com/nicograef/jotti/issues/144) |
| Bestellkommentar erscheint auf allen Arbeitsbons           |       1 | [backlog.md](backlog.md), [#145](https://github.com/nicograef/jotti/issues/145) |
| Ausgabestationen unabhängig von Produktkategorien          |       1 | [backlog.md](backlog.md), [#146](https://github.com/nicograef/jotti/issues/146) |

## Wiederkehrende Fragen

| Frage                                                        | Vereine | Antwort steht in                                                      |
| ------------------------------------------------------------ | ------: | --------------------------------------------------------------------- |
| Braucht jotti am Fest Internet?                              |       3 | Leitfaden „Häufige Fragen"                                            |
| Wo bekommen wir Zugangsdaten, Freigabe oder Lizenz?          |       3 | Leitfaden „Häufige Fragen", Website „Für Vereine"                     |
| Dürfen Vereine aus Österreich oder der Schweiz jotti nutzen? |       2 | Leitfaden „Häufige Fragen", [D16](decisions.md)                       |
| Nehmen wir die aktuelle Version oder warten wir auf 1.0?     |       2 | Leitfaden „Häufige Fragen"                                            |
| Läuft jotti bei schwachem Netz stabil?                       |       1 | Leitfaden „Häufige Fragen"                                            |
| Wie verhält sich jotti unter hoher Last?                     |       1 | Leitfaden „Häufige Fragen"; Messung offen in [backlog.md](backlog.md) |
| Was passiert bei Stromausfall?                               |       1 | Leitfaden „Häufige Fragen"                                            |
| Gibt es eine Demo oder ein Testsystem?                       |       1 | Leitfaden „Häufige Fragen"                                            |
| Welche Bondrucker passen?                                    |       1 | Leitfaden „Installation"                                              |
| Wie aktualisieren wir jotti?                                 |       1 | Leitfaden „Aktualisieren"                                             |
| Wie läuft ein mehrtägiges Fest?                              |       1 | Leitfaden „Der Veranstaltungstag"                                     |
| Welche Rechtsform gibt eine Ortsgruppe eines Verbands an?    |       1 | offen in [backlog.md](backlog.md)                                     |
| Darf ein Partnerverein unsere Installation mitnutzen?        |       1 | offen in [backlog.md](backlog.md)                                     |
