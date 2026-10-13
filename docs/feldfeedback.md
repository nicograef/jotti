# Feldfeedback

Wünsche, Fehlerberichte und Fragen aus den Mails der Vereine, anonymisiert und gezählt. Keine Namen, Orte, Termine oder Zitate: Die Zuordnung zu Vereinen und die Original-Mails liegen nur im gitignorierten Ordner `.mails/`.

„Vereine" zählt die Vereine, die einen Punkt unabhängig voneinander genannt haben.

## Lage

- 31 Organisationen haben angefragt oder die Nutzungsvereinbarung geschlossen, 3 davon aus Österreich; 1 hat zurückgezogen.
- 3 Vereine berichten vom Einsatz am Fest, alle positiv. Gemeldete Störungen: ein still hängender Bondrucker und WLAN mit mehreren Access Points.
- Als Wechselgrund nennen die Vereine teure Miet-Kassensysteme und eine defekte Registrierkasse.

## Umgesetzt

| Wunsch                                              | Vereine | Umsetzung                               |
| --------------------------------------------------- | ------: | --------------------------------------- |
| Direktverkauf ohne Tisch                            |       1 | K-24                                    |
| Rückgeldberechnung                                  |       1 | K-10                                    |
| Ausgabe-Schritt streichen                           |       2 | D01                                     |
| Zweispaltige Theken-Ansicht ab Tablet               |       1 | D08                                     |
| Kategorie-Auswahl über der Produktliste             |       1 | D10 (Kategorie-Pills)                   |
| Lange Produktnamen am Handy umbrechen               |       1 | Q-01                                    |
| Reihenfolge von Produkten und Varianten festlegen   |       2 | S-01                                    |
| Einzelbon pro Stück                                 |       1 | K-25 (Bonmodus am Abholbon)             |
| Erster Login dokumentiert                           |       1 | Leitfaden „Installation"                |
| Liste bewährter Bondrucker                          |       1 | Leitfaden „Installation"                |
| Bedienung ohne lokales WLAN, Handys über Mobilfunk  |       1 | Leitfaden „Self-Hosting"                |
| Anfrageformular, wenn sich kein Mailprogramm öffnet |       3 | Kopier-Fallback im Formular der Website |
| Neuinstallation scheitert (Migration, pnpm 11)      |       1 | behoben                                 |
| README nannte nicht vorhandene Funktionen           |       1 | behoben                                 |

## Abgelehnt

| Wunsch                                      | Vereine | Entscheidung                                                   |
| ------------------------------------------- | ------: | -------------------------------------------------------------- |
| Vereinslogo auf dem Bon                     |       1 | Nicht-Ziel in [anforderungen.md](anforderungen.md)             |
| Beleg per E-Mail statt Bondrucker           |       1 | Nicht-Ziel in [anforderungen.md](anforderungen.md)             |
| Helferdeckel                                |       1 | Nicht-Ziel in [anforderungen.md](anforderungen.md)             |
| Hardware-TSE gegen die laufenden TSE-Kosten |       1 | [D09](decisions.md)                                            |
| Produktebene über der Variantenliste        |       1 | [D10](decisions.md)                                            |
| Österreichische RKSV                        |       3 | [D16](decisions.md)                                            |
| Mehrtägiger Betrieb ganz ohne Internet      |       2 | TSE braucht Internet; F-14 deckt Ausfälle, keinen Dauerzustand |

## Offen

| Wunsch oder Fehler                                         | Vereine | Stand                    |
| ---------------------------------------------------------- | ------: | ------------------------ |
| Gläserpfand: Rückgabe und Verrechnung                      |       2 | [backlog.md](backlog.md) |
| Höchstmenge je Produkt mit Verkaufsstopp                   |       1 | [backlog.md](backlog.md) |
| Laufende Bewirtung im Vereinsheim                          |       1 | [backlog.md](backlog.md) |
| Tische mit eigenen offenen Bestellungen automatisch zeigen |       1 | [backlog.md](backlog.md) |
| Letzter Admin hat sein Passwort vergessen                  |       1 | [backlog.md](backlog.md) |
| Bondrucker hängt still, Relay meldet nichts                |       1 | [backlog.md](backlog.md) |
| WLAN mit mehreren Access Points verliert die Adresse       |       1 | [backlog.md](backlog.md) |
| Installation ohne Docker und ohne Server-Kenntnisse        |       1 | [backlog.md](backlog.md) |

## Wiederkehrende Fragen

| Frage                                            | Vereine | Antwort steht in                 |
| ------------------------------------------------ | ------: | -------------------------------- |
| Braucht jotti am Fest Internet?                  |       3 | Leitfaden „Häufige Fragen"       |
| Wo bekommen wir Zugangsdaten oder eine Freigabe? |       2 | Website „Für Vereine"            |
| Welche Bondrucker passen?                        |       1 | Leitfaden „Installation"         |
| Wie läuft ein mehrtägiges Fest?                  |       1 | noch in keinem Leitfaden-Kapitel |
| Was passiert bei Stromausfall?                   |       1 | Leitfaden „Häufige Fragen"       |
